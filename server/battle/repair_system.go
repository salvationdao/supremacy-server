package battle

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"
)

type RepairSystem struct {
	passport *xsyn_rpcclient.XsynXrpcClient
}

func New(pp *xsyn_rpcclient.XsynXrpcClient) *RepairSystem {
	s := &RepairSystem{
		passport: pp,
	}

	// start repair case cleaner
	go s.start()

	return s
}

// start a routine that cleans up any ended repair cases
func (rs *RepairSystem) start() {
	for {
		time.Sleep(5 * time.Second)

		// wrap the process in a function to enable db transaction and defer rollback
		func() {
			tx, err := gamedb.StdConn.Begin()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to begin db transaction")
				return
			}

			defer tx.Rollback()

			q := `
				UPDATE
					mech_repair_cases
				SET
					ended_at = expected_end_at
				WHERE
					expected_end_at NOTNULL AND expected_end_at <= NOW() AND ended_at ISNULL;
			`
			_, err = tx.Exec(q)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to update ended at in mech repair cases table")
				return
			}

			endedCases, err := boiler.MechRepairCases(
				boiler.MechRepairCaseWhere.EndedAt.IsNotNull(),
			).All(tx)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to load ended mech repair cases")
				return
			}

			// log and remove cases, if there is any ended cases
			if len(endedCases) > 0 {
				// log ended repair cases
				for _, ec := range endedCases {
					mrl := &boiler.MechRepairLog{
						MechID:    ec.MechID,
						Type:      boiler.MechRepairLogTypeREPAIR_ENDED,
						CreatedAt: ec.EndedAt.Time,
					}
					err = mrl.Insert(tx, boil.Infer())
					if err != nil {
						gamelog.L.Error().Interface("mech repair log", mrl).Err(err).Msg("Failed to insert mech repair log.")
						return
					}
				}

				// delete all the mech repair cases
				_, err = endedCases.DeleteAll(tx)
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to delete ended repair cases")
					return
				}

				mechIDs := []string{}
				for _, ec := range endedCases {
					mechIDs = append(mechIDs, ec.MechID)
				}

				// broadcast nil to all the mech
				go func(mechIDs []string) {
					cis, err := boiler.CollectionItems(
						boiler.CollectionItemWhere.ItemID.IN(mechIDs),
						qm.Load(boiler.CollectionItemRels.Owner),
					).All(gamedb.StdConn)
					if err != nil {
						gamelog.L.Error().Err(err).Strs("mech ids", mechIDs).Msg("Failed to load collection items")
						return
					}

					for _, ci := range cis {
						if ci.R != nil && ci.R.Owner != nil && ci.R.Owner.FactionID.Valid {
							ws.PublishMessage(fmt.Sprintf("/faction/%s/mech/%s/repair_status", ci.R.Owner.FactionID.String, ci.ItemID), server.WarMachineRepairStatusSubscribe, nil)
						}
					}
				}(mechIDs)
			}

			err = tx.Commit()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to commit db transaction")
				return
			}
		}()
	}
}

// RegisterMechRepairCase register a mech repair case
func (rs *RepairSystem) RegisterMechRepairCase(mechID string, maxHealth, remainHealth uint32) error {
	// no need to repair
	if maxHealth == remainHealth {
		return nil
	}

	now := time.Now()

	mh := decimal.NewFromInt(int64(maxHealth))
	rh := decimal.NewFromInt(int64(remainHealth))
	damagedPortion := mh.Sub(rh).Div(mh)

	drm := db.GetDecimalWithDefault(db.KeyMechRepairDefaultDurationMinutes, decimal.NewFromInt(360))

	// calculate portion of the remaining health
	durationMinutes := decimal.NewFromInt(1).Add(drm.Mul(damagedPortion)).IntPart()

	mrc := &boiler.MechRepairCase{
		MechID:              mechID,
		Fee:                 db.GetDecimalWithDefault(db.KeyMechStandardRepairFee, decimal.New(5, 18)),
		FastRepairFee:       db.GetDecimalWithDefault(db.KeyMechFastRepairFee, decimal.New(30, 18)),
		RepairPeriodMinutes: int(durationMinutes),
		MaxHealth:           mh,
		RemainHealth:        rh,
	}

	err := mrc.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Interface("mech repair case", mrc).Err(err).Msg("Failed to insert mech repair case.")
		return terror.Error(err, "Failed to register mech repair case")
	}

	// mech repair log
	mrl := &boiler.MechRepairLog{
		MechID:    mechID,
		Type:      boiler.MechRepairLogTypeREGISTER_REPAIR,
		CreatedAt: now,
	}

	err = mrl.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to log mech repair event.")
	}

	return nil
}

func (rs *RepairSystem) IsStillRepairing(mechID string) (bool, error) {
	mrc, err := boiler.MechRepairCases(
		boiler.MechRepairCaseWhere.MechID.EQ(mechID),
		boiler.MechRepairCaseWhere.EndedAt.IsNotNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, terror.Error(err, "Failed to check mech repair status")
	}

	return mrc != nil, nil
}

func (rs *RepairSystem) StartStandardRepair(userID string, mechID string) error {
	now := time.Now()
	mrc, err := boiler.MechRepairCases(
		boiler.MechRepairCaseWhere.MechID.EQ(mechID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("mech id", mechID).Msg("Failed to log mech repair case.")
		return terror.Error(err, "Failed to load mech repair case.")
	}

	if mrc == nil {
		return terror.Error(fmt.Errorf("no mech to repair"), "The mech is not in the repair center.")
	}

	if mrc.EndedAt.Valid {
		return terror.Error(fmt.Errorf("no mech to repair"), "The mech is already repaired.")
	}

	// cna only start standard repair when current status is pending
	if mrc.Status != boiler.MechRepairStatusPENDING {
		return terror.Error(fmt.Errorf("repair process has already started"), "The repair process has already started.")
	}

	// start the process
	mrc.StartedAt = null.TimeFrom(now)
	mrc.ExpectedEndAt = null.TimeFrom(now.Add(time.Duration(mrc.RepairPeriodMinutes) * time.Minute))
	mrc.Status = boiler.MechRepairStatusSTANDARD_REPAIR
	_, err = mrc.Update(gamedb.StdConn, boil.Whitelist(
		boiler.MechRepairCaseColumns.StartedAt,
		boiler.MechRepairCaseColumns.ExpectedEndAt,
		boiler.MechRepairCaseColumns.Status,
	))
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to update mech repair process.")
		return terror.Error(err, "Failed to start repair process.")
	}

	// mech repair log
	mrl := &boiler.MechRepairLog{
		MechID:           mechID,
		Type:             boiler.MechRepairLogTypeSTART_STANDARD_REPAIR,
		InvolvedPlayerID: null.StringFrom(userID),
		CreatedAt:        now,
	}
	err = mrl.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to log mech repair event.")
	}

	BroadcastMechRepairStatus(mrc)

	return nil
}

// StartFastRepair start a fast repair process or speed up an existing standard repair process
func (rs *RepairSystem) StartFastRepair(userID string, mechID string) error {
	now := time.Now()

	mrc, err := boiler.MechRepairCases(
		boiler.MechRepairCaseWhere.MechID.EQ(mechID),
		boiler.MechRepairCaseWhere.EndedAt.IsNotNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to load mech repair case.")
	}

	if mrc == nil {
		return terror.Error(fmt.Errorf("no mech is repairing"), "The mech is not in the repair center.")
	}

	if mrc.EndedAt.Valid {
		return terror.Error(fmt.Errorf("repair process has already ended"), "The repair process has already ended")
	}

	// cna only start standard repair when current status is pending
	if mrc.Status == boiler.MechRepairStatusFAST_REPAIR {
		return terror.Error(fmt.Errorf("already in fast repair"), "The mech is already in fast repair mode.")
	}

	if mrc.FastRepairTXID.Valid {
		return terror.Error(fmt.Errorf("process already speed up"), "Repair process is already speed up.")
	}

	damagedPortion := mrc.MaxHealth.Sub(mrc.RemainHealth).Div(mrc.MaxHealth)
	drm := db.GetDecimalWithDefault(db.KeyMechFastRepairDurationMinutes, decimal.NewFromInt(30))
	durationMinutes := decimal.NewFromInt(1).Add(drm.Mul(damagedPortion)).IntPart()

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Msg("Failed to begin db transaction.")
		return terror.Error(err, "Failed to pay instant repair fee.")
	}

	defer tx.Rollback()

	fastRepairFee := mrc.FastRepairFee
	logType := boiler.MechRepairLogTypeSTART_FAST_REPAIR
	if mrc.StartedAt.Valid {
		// deduct fee if repair is already started
		fastRepairFee = mrc.FastRepairFee.Sub(mrc.Fee)

		// change log to speed up
		logType = boiler.MechRepairLogTypeSPEED_UP
	}

	if !mrc.StartedAt.Valid {
		mrc.StartedAt = null.TimeFrom(now)
	}

	mrc.Status = boiler.MechRepairStatusFAST_REPAIR
	mrc.RepairPeriodMinutes = int(durationMinutes)
	mrc.ExpectedEndAt = null.TimeFrom(mrc.StartedAt.Time.Add(time.Duration(durationMinutes) * time.Minute))
	mrc.FastRepairTXID = null.StringFrom("SPEED_UP_TX") // will be replaced after sups spend is success

	// check repair is ended
	if mrc.ExpectedEndAt.Time.After(now) {
		mrc.EndedAt = null.TimeFrom(now)
	}

	_, err = mrc.Update(tx, boil.Whitelist(
		boiler.MechRepairCaseColumns.StartedAt,
		boiler.MechRepairCaseColumns.Status,
		boiler.MechRepairCaseColumns.RepairPeriodMinutes,
		boiler.MechRepairCaseColumns.ExpectedEndAt,
		boiler.MechRepairCaseColumns.FastRepairTXID,
		boiler.MechRepairCaseColumns.EndedAt,
	))
	if err != nil {
		gamelog.L.Error().Interface("mech repair case", mrc).Err(err).Msg("Failed to update the ended_at column of mech repair case.")
		return terror.Error(err, "Failed to pay instant repair fee")
	}

	// pay instant repair fee
	txID, err := rs.passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           uuid.FromStringOrNil(userID),
		ToUserID:             uuid.FromStringOrNil(server.XsynTreasuryUserID.String()),
		Amount:               fastRepairFee.StringFixed(0),
		TransactionReference: server.TransactionReference(fmt.Sprintf("pay_mech_repair_fee|%s|%d", mrc.MechID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupSupremacy),
		SubGroup:             string(server.TransactionGroupBattle),
		Description:          "Paying mech repair fee " + mechID + ".",
		NotSafe:              true,
	})
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Str("asset repair id", mrc.MechID).Err(err).Msg("Failed to pay asset repair fee")
		return terror.Error(err, "Failed to pay asset repair fee")
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed to pay instant repair fee.")
	}

	mrc.FastRepairTXID = null.StringFrom(txID)
	_, err = mrc.Update(tx, boil.Whitelist(boiler.MechRepairCaseColumns.FastRepairFee))
	if err != nil {
		gamelog.L.Error().Interface("mech repair case", mrc).Err(err).Msg("Failed to update the instant repair txid of mech repair case.")
	}

	// mech repair log
	mrl := &boiler.MechRepairLog{
		MechID:           mechID,
		Type:             logType,
		InvolvedPlayerID: null.StringFrom(userID),
		CreatedAt:        now,
	}
	err = mrl.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to log mech repair event.")
	}

	BroadcastMechRepairStatus(mrc)

	return nil
}

type MechRepairStatus struct {
	RepairStatus  string   `json:"repair_status"`
	RemainSeconds null.Int `json:"remain_seconds"`
}

func BroadcastMechRepairStatus(mrc *boiler.MechRepairCase) {
	if mrc == nil {
		return
	}

	// get faction
	ci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(mrc.MechID),
		qm.Load(boiler.CollectionItemRels.Owner),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("item id (mech)", mrc.MechID).Err(err).Msg("Failed to load collection item.")
		return
	}

	mrs := MechRepairStatus{
		RepairStatus: mrc.Status,
	}

	if mrc.ExpectedEndAt.Valid && mrc.StartedAt.Valid {
		// add a second delay
		mrs.RemainSeconds = null.IntFrom(1 + int(mrc.ExpectedEndAt.Time.Sub(mrc.StartedAt.Time).Seconds()))
	}

	if ci.R != nil && ci.R.Owner != nil && ci.R.Owner.FactionID.Valid {
		ws.PublishMessage(fmt.Sprintf("/faction/%s/mech/%s/repair_status", ci.R.Owner.FactionID.String, mrc.MechID), server.WarMachineRepairStatusSubscribe, mrs)
	}
}
