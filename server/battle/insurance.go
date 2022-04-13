package battle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"time"
)

const (
	RepairModeStandard = "STANDARD"
	RepairModeFast     = "FAST"
)

func (btl *Battle) processWarMachineRepair(payload *BattleEndPayload) {
	// get war machines that required repair
	requireRepairedWarMachinIDs := []string{}
	for _, wm := range btl.WarMachines {
		isWin := false
		for _, wwm := range payload.WinningWarMachines {
			if wm.Hash == wwm.Hash {
				isWin = true
				break
			}
		}
		if !isWin {
			requireRepairedWarMachinIDs = append(requireRepairedWarMachinIDs, wm.ID)
		}
	}

	if len(requireRepairedWarMachinIDs) == 0 {
		gamelog.L.Warn().Str("battle id", btl.ID).Msg("There is no war machine needs repair, which shouldn't happen!!!")
		return
	}

	mechs, err := boiler.Mechs(
		qm.Select(boiler.MechColumns.ID, boiler.MechColumns.IsInsured),
		boiler.MechWhere.ID.IN(requireRepairedWarMachinIDs),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Str("battle id", btl.ID).Interface("mech id list", requireRepairedWarMachinIDs).Msg("Failed to get mechs from db")
		return
	}

	now := time.Now()
	for _, mech := range mechs {
		repairFee := btl.arena.InsurancePrice(mech.ID)

		ar := boiler.AssetRepair{
			MechID:        mech.ID,
			RepairMode:    RepairModeFast,
			CompleteUntil: now.Add(30 * time.Minute),
			FullRepairFee: repairFee,
		}

		// if mech is not insured
		if !mech.IsInsured {
			ar.RepairMode = RepairModeStandard
			ar.CompleteUntil = now.Add(24 * time.Hour)              // change repair time to 24 hours
			ar.FullRepairFee = repairFee.Mul(decimal.NewFromInt(3)) // three time insurance fee
		}

		err := ar.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Str("mech id", mech.ID).Err(err).Msg("Failed to insert asset repair")
		}
	}
}

// InsurancePrice handle price calculation
func (arena *Arena) InsurancePrice(mechID string) decimal.Decimal {
	// get insurance price from mech

	// else get current global insurance price

	return decimal.New(10, 18)
}

type AssetRepairPayFeeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"payload"`
}

const HubKeyAssetRepairPayFee hub.HubCommandKey = hub.HubCommandKey("ASSET:REPAIR:PAY:FEE")

func (arena *Arena) AssetRepairPayFeeHandler(ctx context.Context, hubc *hub.Client, payload []byte, userFactionID uuid.UUID, reply hub.ReplyFunc) error {
	req := &AssetRepairPayFeeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	playerID := uuid.FromStringOrNil(hubc.Identifier())
	if playerID.IsNil() {
		return terror.Error(terror.ErrForbidden, "You are not login")
	}

	// get mech
	mech, err := boiler.FindMech(gamedb.StdConn, req.Payload.MechID)
	if err != nil {
		return terror.Error(err, "Failed to get mech from db")
	}

	if mech.OwnerID != hubc.Identifier() {
		return terror.Error(terror.ErrForbidden, "You are not the owner of the mech")
	}

	now := time.Now()

	// check repair center
	ar, err := boiler.AssetRepairs(
		boiler.AssetRepairWhere.MechID.EQ(mech.ID),
		boiler.AssetRepairWhere.CompleteUntil.GT(now),
	).One(gamedb.StdConn)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return terror.Error(err, "Your mech is not in the repair center")
		}

		return terror.Error(err, "Failed to get asset repair record from db")
	}

	// calculate pay fee
	fullDurationSecond := decimal.NewFromFloat(ar.CompleteUntil.Sub(ar.CreatedAt).Seconds())
	alreadyPassedSeconds := decimal.NewFromFloat(now.Sub(ar.CreatedAt).Seconds())
	remainSeconds := fullDurationSecond.Sub(alreadyPassedSeconds)
	ratio := remainSeconds.Div(fullDurationSecond)

	fee := ar.FullRepairFee.Mul(ratio)

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, "Failed to start db transaction")
	}

	defer tx.Rollback()

	ar.CompleteUntil = now
	_, err = ar.Update(tx, boil.Whitelist(boiler.AssetRepairColumns.CompleteUntil))
	if err != nil {
		return terror.Error(err, "Failed to update asset repair")
	}

	// get syndicate account
	factionAccountID, ok := server.FactionUsers[userFactionID.String()]
	if !ok {
		gamelog.L.Error().
			Str("player id", playerID.String()).
			Str("faction ID", userFactionID.String()).
			Err(fmt.Errorf("failed to get hard coded syndicate player id")).
			Msg("unable to get hard coded syndicate player ID from faction ID")
		return terror.Error(err, "Failed to load syndicate id")
	}

	// pay sups
	_, err = arena.RPCClient.SpendSupMessage(rpcclient.SpendSupsReq{
		FromUserID:           playerID,
		ToUserID:             uuid.FromStringOrNil(factionAccountID),
		Amount:               fee.StringFixed(18),
		TransactionReference: server.TransactionReference(fmt.Sprintf("pay_asset_repair_fee|%s|%d", ar.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupBattle),
		Description:          "pay asset repair fee: " + ar.ID,
		NotSafe:              true,
	})
	if err != nil {
		gamelog.L.Error().Str("asset repair id", ar.ID).Err(err).Msg("Failed to pay asset repair fee")
		return terror.Error(err, "Failed to pay asset repair fee")
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed to commit db transaction")
	}

	reply(true)

	return nil
}

type AssetRepairStatusRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		MechID string `json:"mech_id"`
	} `json:"payload"`
}

type AssetRepairStatusResponse struct {
	TotalRequiredSeconds int             `json:"total_required_seconds"`
	RemainSeconds        int             `json:"remain_seconds"`
	FullRepairFee        decimal.Decimal `json:"full_repair_fee"`
}

const HubKeyAssetRepairStatus hub.HubCommandKey = hub.HubCommandKey("ASSET:REPAIR:STATUS")

func (arena *Arena) AssetRepairStatusHandler(ctx context.Context, hubc *hub.Client, payload []byte, userFactionID uuid.UUID, reply hub.ReplyFunc) error {
	req := &AssetRepairStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	playerID := uuid.FromStringOrNil(hubc.Identifier())
	if playerID.IsNil() {
		return terror.Error(terror.ErrForbidden, "You are not login")
	}

	// get mech
	mech, err := boiler.FindMech(gamedb.StdConn, req.Payload.MechID)
	if err != nil {
		return terror.Error(err, "Failed to get mech from db")
	}

	if mech.OwnerID != hubc.Identifier() {
		return terror.Error(terror.ErrForbidden, "You are not the owner of the mech")
	}

	now := time.Now()

	// check repair center
	ar, err := boiler.AssetRepairs(
		boiler.AssetRepairWhere.MechID.EQ(mech.ID),
		boiler.AssetRepairWhere.CompleteUntil.GT(now),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to get asset repair record from db")
	}

	if ar == nil {
		reply(nil)
		return nil
	}

	// reply asset repair status
	reply(&AssetRepairStatusResponse{
		TotalRequiredSeconds: int(ar.CompleteUntil.Sub(ar.CreatedAt).Seconds()),
		RemainSeconds:        int(ar.CompleteUntil.Sub(now).Seconds()),
		FullRepairFee:        ar.FullRepairFee,
	})

	return nil
}
