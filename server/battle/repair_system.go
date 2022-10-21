package battle

import (
	"database/sql"
	"fmt"
	"golang.org/x/exp/slices"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func (am *ArenaManager) SendRepairFunc(fn func() error) error {
	am.RepairFuncMx.Lock()
	defer am.RepairFuncMx.Unlock()
	return fn()
}

func (am *ArenaManager) RepairOfferCleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	repairBayTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			am.expiredRepairOfferChecker()
		case <-repairBayTicker.C:
			am.repairBayCompleteChecker()
		}
	}
}

// CloseRepairOffers close the given repair offer
// IMPORTANT: this function should NOT be used outside of "SendRepairFunc" !!!
func (am *ArenaManager) CloseRepairOffers(repairOfferIDs []string, offerCloseReason string, agentCloseReason string) error {
	now := time.Now()

	// load repair offers
	ros, err := boiler.RepairOffers(
		boiler.RepairOfferWhere.ID.IN(repairOfferIDs),
		boiler.RepairOfferWhere.ClosedAt.IsNull(),
		qm.Load(boiler.RepairOfferRels.RepairCase),
		qm.Load(
			boiler.RepairOfferRels.RepairAgents,
			boiler.RepairAgentWhere.FinishedAt.IsNull(),
		),
		qm.Load(boiler.RepairOfferRels.OfferedBy),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get repair offers")
		return terror.Error(err, "Failed to get repair offers")
	}

	if len(ros) == 0 {
		return nil
	}

	// NOTE: this need to be at the outside of db transaction scope,
	// so it can stop repair blocks from being inserted through db trigger
	_, err = ros.UpdateAll(gamedb.StdConn, boiler.M{
		boiler.RepairOfferColumns.ClosedAt:       null.TimeFrom(now),
		boiler.RepairOfferColumns.FinishedReason: null.StringFrom(offerCloseReason),
	})
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to close expired repair offer.")
		return terror.Error(err, "Failed to close expired repair offer.")
	}

	for _, ro := range ros {
		if ro.R == nil {
			continue
		}

		err = func(ro *boiler.RepairOffer) error {
			// for broadcast
			ro.ClosedAt = null.TimeFrom(now)
			ro.FinishedReason = null.StringFrom(offerCloseReason)

			tx, err := gamedb.StdConn.Begin()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to begin db transaction.")
				return terror.Error(err, "Failed to begin db transaction.")
			}

			defer tx.Rollback()

			// broadcast close offer
			rc := ro.R.RepairCase
			sro := &server.RepairOffer{
				RepairOffer:          ro,
				BlocksRequiredRepair: rc.BlocksRequiredRepair,
				BlocksRepaired:       rc.BlocksRepaired,
				SupsWorthPerBlock:    ro.OfferedSupsAmount.Div(decimal.NewFromInt(int64(ro.BlocksTotal))),
				WorkingAgentCount:    0,
			}

			if ro.R.OfferedBy != nil {
				sro.JobOwner = server.PublicPlayerFromBoiler(ro.R.OfferedBy)

				ws.PublishMessage(fmt.Sprintf("/secure/repair_offer/%s", ro.ID), server.HubKeyRepairOfferSubscribe, sro)
				ws.PublishMessage("/secure/repair_offer/update", server.HubKeyRepairOfferUpdateSubscribe, []*server.RepairOffer{sro})
				ws.PublishMessage(fmt.Sprintf("/secure/mech/%s/active_repair_offer", rc.MechID), server.HubKeyMechActiveRepairOffer, sro)
			}

			if ro.R.RepairAgents != nil && len(ro.R.RepairAgents) > 0 {
				_, err = ro.R.RepairAgents.UpdateAll(tx, boiler.M{
					boiler.RepairAgentColumns.FinishedAt:     null.TimeFrom(now),
					boiler.RepairAgentColumns.FinishedReason: null.StringFrom(agentCloseReason),
				})
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to close expired repair agent.")
					return terror.Error(err, "Failed to close expired repair agent.")
				}
			}

			// refund process
			if !ro.OfferedByID.Valid || ro.OfferedSupsAmount.Equal(decimal.Zero) {
				return nil
			}

			totalRefundBlocks := ro.BlocksTotal
			totalRepairedBlocks, err := ro.RepairBlocks().Count(tx)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("repair offer", ro).Msg("Failed to get total repair blocks.")
				return nil
			}

			totalRefundBlocks = totalRefundBlocks - int(totalRepairedBlocks)

			if totalRefundBlocks > 0 {
				amount := ro.OfferedSupsAmount.Div(decimal.NewFromInt(int64(ro.BlocksTotal))).Mul(decimal.NewFromInt(int64(totalRefundBlocks)))

				if amount.GreaterThan(decimal.Zero) {
					// refund reward
					refundTxID, err := am.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
						FromUserID:           uuid.Must(uuid.FromString(server.RepairCenterUserID)),
						ToUserID:             uuid.Must(uuid.FromString(ro.OfferedByID.String)),
						Amount:               amount.StringFixed(0),
						TransactionReference: server.TransactionReference(fmt.Sprintf("refund_unclaimed_repair_offer_reward|%s|%d", ro.ID, time.Now().UnixNano())),
						Group:                string(server.TransactionGroupSupremacy),
						SubGroup:             string(server.TransactionGroupRepair),
						Description:          "refund unclaimed repair offer reward.",
					})
					if err != nil {
						gamelog.L.Error().
							Str("player_id", ro.OfferedByID.String).
							Str("repair offer id", ro.ID).
							Str("amount", amount.StringFixed(0)).
							Err(err).Msg("Failed to refund unclaimed repair offer reward.")
						return terror.Error(err, "Failed to refund unclaimed repair offer reward.")
					}

					ro.RefundTXID = null.StringFrom(refundTxID)
					_, err = ro.Update(tx, boil.Whitelist(boiler.RepairOfferColumns.RefundTXID))
					if err != nil {
						gamelog.L.Error().
							Interface("repair offer", ro).
							Err(err).Msg("Failed to update repair offer refund transaction id")
					}
				}
			}
			err = tx.Commit()
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
				return terror.Error(err, "Failed to commit db transaction.")
			}

			return nil
		}(ro)
		if err != nil {
			return err
		}
	}

	return nil
}

// expiredRepairOfferChecker close any expired repair offers
func (am *ArenaManager) expiredRepairOfferChecker() {
	am.RepairFuncMx.Lock()
	defer am.RepairFuncMx.Unlock()

	// expire repair offer
	ros, err := boiler.RepairOffers(
		boiler.RepairOfferWhere.ExpiresAt.LTE(time.Now()),
		boiler.RepairOfferWhere.ClosedAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get repair offer")
		return
	}

	if len(ros) == 0 {
		return
	}

	roIDs := []string{}
	for _, ro := range ros {
		roIDs = append(roIDs, ro.ID)
	}

	err = am.CloseRepairOffers(roIDs, boiler.RepairFinishReasonEXPIRED, boiler.RepairAgentFinishReasonEXPIRED)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to close expired repair offers.")
	}
}

// repairBayCompleteChecker check if there are any repair slot complete
func (am *ArenaManager) repairBayCompleteChecker() {
	am.RepairFuncMx.Lock()
	defer am.RepairFuncMx.Unlock()

	now := time.Now()
	pms, err := boiler.PlayerMechRepairSlots(
		boiler.PlayerMechRepairSlotWhere.Status.EQ(boiler.RepairSlotStatusREPAIRING),
		boiler.PlayerMechRepairSlotWhere.NextRepairTime.LTE(null.TimeFrom(now)),
		qm.Load(boiler.PlayerMechRepairSlotRels.RepairCase),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load repairing cases.")
		return
	}

	// skip, if there is no completed repair slot
	if pms == nil {
		return
	}

	nextRepairDurationSeconds := db.GetIntWithDefault(db.KeyAutoRepairDurationSeconds, 600)
	wg := sync.WaitGroup{}
	for _, pm := range pms {
		wg.Add(1)
		go func(playerMechRepairSlot *boiler.PlayerMechRepairSlot) {
			defer wg.Done()

			// mark current repairing slot to "DONE" and mark next slot to "REPAIRING"
			swapSlot := func(pmr *boiler.PlayerMechRepairSlot) {
				tx, err := gamedb.StdConn.Begin()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to start db transaction.")
					return
				}

				defer tx.Rollback()

				nexSlot, err := boiler.PlayerMechRepairSlots(
					boiler.PlayerMechRepairSlotWhere.PlayerID.EQ(pmr.PlayerID),
					boiler.PlayerMechRepairSlotWhere.Status.EQ(boiler.RepairSlotStatusPENDING),
					qm.OrderBy(boiler.PlayerMechRepairSlotColumns.SlotNumber),
				).One(tx)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					gamelog.L.Error().Str("player id", pmr.PlayerID).Err(err).Msg("Failed to load player mech repair bays.")
					return
				}

				// set next slot status to "REPAIRING"
				if nexSlot != nil {
					nexSlot.Status = boiler.RepairSlotStatusREPAIRING
					nexSlot.NextRepairTime = null.TimeFrom(now.Add(time.Duration(nextRepairDurationSeconds) * time.Second))
					_, err = nexSlot.Update(tx, boil.Whitelist(
						boiler.PlayerMechRepairSlotColumns.Status,
						boiler.PlayerMechRepairSlotColumns.NextRepairTime,
					))
					if err != nil {
						gamelog.L.Error().Err(err).Interface("player repair slot", pmr).Msg("Failed to complete current repair slot.")
						return
					}
				}

				// decrement slot number
				err = db.DecrementRepairSlotNumber(tx, pmr.PlayerID, pmr.SlotNumber)
				if err != nil {
					gamelog.L.Error().Err(err).Str("player id", pmr.PlayerID).Msg("Failed to decrement slot number.")
					return
				}

				// update current bay status to "DONE"
				pmr.Status = boiler.RepairSlotStatusDONE
				pmr.SlotNumber = 0
				pmr.NextRepairTime = null.TimeFromPtr(nil)
				_, err = pmr.Update(tx,
					boil.Whitelist(
						boiler.PlayerMechRepairSlotColumns.Status,
						boiler.PlayerMechRepairSlotColumns.SlotNumber,
						boiler.PlayerMechRepairSlotColumns.NextRepairTime,
					),
				)
				if err != nil {
					gamelog.L.Error().Err(err).Interface("player repair slot", pmr).Msg("Failed to complete current repair slot.")
					return
				}

				err = tx.Commit()
				if err != nil {
					gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
					return
				}

				// broadcast current repair bay
				go BroadcastRepairBay(pmr.PlayerID)
			}

			// if no repair case, swap repair slot
			if playerMechRepairSlot.R == nil || playerMechRepairSlot.R.RepairCase == nil {
				gamelog.L.Warn().Interface("repair slot", playerMechRepairSlot).Msg("The mech is missing repair case.")
				swapSlot(playerMechRepairSlot)
				return
			}

			rc := playerMechRepairSlot.R.RepairCase

			// load system repair offer
			systemOffer, err := rc.RepairOffers(
				boiler.RepairOfferWhere.OfferedByID.IsNull(),
			).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Error().Err(err).Interface("repair case", rc).Msg("Failed to load system repair offer of the repair case.")
				return
			}

			// if no system offer, swap repair case
			if systemOffer == nil {
				gamelog.L.Warn().Interface("repair case", rc).Msg("The mech is missing system repair offer.")
				swapSlot(playerMechRepairSlot)
				return
			}

			// if system offer or repair case are already closed
			if systemOffer.ClosedAt.Valid || rc.CompletedAt.Valid {
				swapSlot(playerMechRepairSlot)
				return
			}

			// generate an agent from repair center
			ra := &boiler.RepairAgent{
				RepairCaseID:   rc.ID,
				RepairOfferID:  systemOffer.ID,
				PlayerID:       server.RepairCenterUserID,
				RequiredStacks: 0,
			}

			err = ra.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Interface("repair agent", ra).Msg("Failed to generate repair agent from repair center user")
				return
			}

			deleteAgent := func(agentID string) {
				_, err = boiler.RepairAgents(
					boiler.RepairAgentWhere.ID.EQ(agentID),
				).DeleteAll(gamedb.StdConn, true)
				if err != nil {
					gamelog.L.Error().Err(err).Interface("repair agent id", agentID).Msg("Failed to delete repair agent")
					return
				}
			}

			// do repair
			rb := boiler.RepairBlock{
				RepairAgentID: ra.ID,
				RepairCaseID:  rc.ID,
				RepairOfferID: systemOffer.ID,
			}

			err = rb.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				// clean up repair agent
				deleteAgent(ra.ID)

				// swap repair slot, if unable to write block
				if err.Error() == "unable to write block" {
					swapSlot(playerMechRepairSlot)
					return
				}

				gamelog.L.Error().Err(err).Interface("repair block", rb).Msg("Failed to insert repair block by repair center.")
				return
			}

			// check repair complete
			err = rc.Reload(gamedb.StdConn)
			if err != nil {
				// might be cleaned up already
				gamelog.L.Warn().Err(err).Msg("Failed to reload repair case.")
				swapSlot(playerMechRepairSlot)
				return
			}

			// repair broadcast repair details
			ws.PublishMessage(fmt.Sprintf("/secure/mech/%s/repair_case", rc.MechID), server.HubKeyMechRepairCase, rc)

			// if not complete
			if rc.BlocksRequiredRepair > rc.BlocksRepaired {
				// set next repair time
				playerMechRepairSlot.NextRepairTime = null.TimeFrom(now.Add(time.Duration(nextRepairDurationSeconds) * time.Second))
				_, err = playerMechRepairSlot.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerMechRepairSlotColumns.NextRepairTime))
				if err != nil {
					gamelog.L.Error().Err(err).Interface("repair slot", playerMechRepairSlot).Msg("Failed to update next repair time of the repair slot.")
				}

				ws.PublishMessage(fmt.Sprintf("/secure/mech/%s/repair_case", rc.MechID), server.HubKeyMechRepairCase, rc)

				// broadcast mech status
				go BroadcastMechQueueStatus(playerMechRepairSlot.PlayerID, rc.MechID)

				// broadcast current repair bay
				go BroadcastRepairBay(playerMechRepairSlot.PlayerID)

				return
			}

			ws.PublishMessage(fmt.Sprintf("/secure/mech/%s/repair_case", rc.MechID), server.HubKeyMechRepairCase, nil)

			// otherwise swap bay
			swapSlot(playerMechRepairSlot)

			// close offers
			rc.CompletedAt = null.TimeFrom(now)
			_, err = rc.Update(gamedb.StdConn, boil.Whitelist(boiler.RepairCaseColumns.CompletedAt))
			if err != nil {
				gamelog.L.Error().Err(err).Interface("repair case", rc).Msg("Failed to complete repair case.")
			}

			// close related offers
			ros, err := rc.RepairOffers(
				boiler.RepairOfferWhere.ClosedAt.IsNull(),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("repair case", rc).Msg("Failed to load repair offers from the repair case.")
			}

			if ros == nil || len(ros) == 0 {
				return
			}

			roIDs := []string{}
			for _, ro := range ros {
				roIDs = append(roIDs, ro.ID)
			}
			err = am.CloseRepairOffers(roIDs, boiler.RepairAgentFinishReasonSUCCEEDED, boiler.RepairAgentFinishReasonEXPIRED)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("repair offers", ros).Msg("Failed to close repair offer.")
				return
			}
			// broadcast mech status
			go BroadcastMechQueueStatus(playerMechRepairSlot.PlayerID, rc.MechID)
		}(pm)
	}

	wg.Wait()
}

// RegisterMechRepairCase insert mech repair case and track repair stack
func RegisterMechRepairCase(mechID string, blueprintID string, maxHealth uint32, remainHealth uint32) error {
	if remainHealth == maxHealth {
		return nil
	}

	damagedPortion := decimal.NewFromInt(1)
	if remainHealth != 0 {
		mh := decimal.NewFromInt(int64(maxHealth))
		rh := decimal.NewFromInt(int64(remainHealth))
		damagedPortion = mh.Sub(rh).Div(mh)
	}

	// get mech model
	model, err := boiler.FindBlueprintMech(gamedb.StdConn, blueprintID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech model id", blueprintID).Msg("Failed to load mech model for repair block detail.")
		return terror.Error(err, "Failed to load mech model")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction.")
		return terror.Error(err, "Failed to begin db transaction.")
	}

	defer tx.Rollback()

	// set block total
	blocksTotal := decimal.NewFromInt(int64(model.RepairBlocks)).Mul(damagedPortion).Ceil().IntPart()

	rc := &boiler.RepairCase{
		MechID:               mechID,
		BlocksRequiredRepair: int(blocksTotal),
	}

	err = rc.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("repair case", rc).Msg("Failed to insert repair case.")
		return terror.Error(err, "Failed to insert repair case.")
	}

	// insert self repair offer
	ro := boiler.RepairOffer{
		RepairCaseID:      rc.ID,
		BlocksTotal:       int(blocksTotal),
		OfferedSupsAmount: decimal.Zero,
		ExpiresAt:         time.Now().AddDate(10, 0, 0),
	}
	err = ro.Insert(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert repair offer.")
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed to commit db transaction.")
	}

	ws.PublishMessage(fmt.Sprintf("/secure/mech/%s/repair_case", rc.MechID), server.HubKeyMechRepairCase, rc)

	return nil
}

func BroadcastRepairBay(playerID string) {
	l := gamelog.L.With().Str("player id", playerID).Str("func name", "BroadcastRepairBay").Logger()

	resp := []*boiler.PlayerMechRepairSlot{}
	pms, err := boiler.PlayerMechRepairSlots(
		boiler.PlayerMechRepairSlotWhere.PlayerID.EQ(playerID),
		boiler.PlayerMechRepairSlotWhere.Status.NEQ(boiler.RepairSlotStatusDONE),
		qm.OrderBy(boiler.PlayerMechRepairSlotColumns.SlotNumber),
	).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("Failed to load player mech .")
		return
	}

	if pms != nil {
		resp = pms
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/repair_bay", playerID), server.HubKeyMechRepairSlots, resp)
}

// PauseRepairCases pause the repair cases of the mechs and close all the related repair offers
func (am *ArenaManager) PauseRepairCases(mechIDs []string) error {
	am.RepairFuncMx.Lock()
	defer am.RepairFuncMx.Unlock()

	repairCases, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.IN(mechIDs),
		boiler.RepairCaseWhere.CompletedAt.IsNull(), // not completed
		boiler.RepairCaseWhere.PausedAt.IsNull(),    // not paused
		qm.Load(boiler.RepairCaseRels.RepairOffers, boiler.RepairOfferWhere.ClosedAt.IsNull()),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Strs("mech id list", mechIDs).Msg("Failed to load repair cases of the mechs")
		return terror.Error(err, "Failed to load reqpir case")
	}

	if repairCases == nil {
		return nil
	}

	_, err = repairCases.UpdateAll(gamedb.StdConn, boiler.M{boiler.RepairCaseColumns.PausedAt: null.TimeFrom(time.Now())})
	if err != nil {
		gamelog.L.Error().Err(err).Interface("repair cases", repairCases).Msg("Failed to pause repair cases")
		return terror.Error(err, "Failed to pause repair case")
	}

	repairOfferIDs := []string{}
	for _, rc := range repairCases {
		if rc.R == nil || rc.R.RepairOffers == nil {
			continue
		}
		for _, ro := range rc.R.RepairOffers {
			repairOfferIDs = append(repairOfferIDs, ro.ID)
		}
	}

	err = am.CloseRepairOffers(repairOfferIDs, boiler.RepairFinishReasonSTOPPED, boiler.RepairAgentFinishReasonEXPIRED)
	if err != nil {
		return err
	}

	return nil
}

// RestartRepairCases unclose all the repair cases and related system offers
func (am *ArenaManager) RestartRepairCases(mechIDs []string) error {
	am.RepairFuncMx.Lock()
	defer am.RepairFuncMx.Unlock()

	repairCases, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.IN(mechIDs),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),                                               // not completed
		boiler.RepairCaseWhere.PausedAt.IsNotNull(),                                               // is paused
		qm.Load(boiler.RepairCaseRels.RepairOffers, boiler.RepairOfferWhere.OfferedByID.IsNull()), // get system offer
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Strs("mech id list", mechIDs).Msg("Failed to load repair cases of the mechs")
		return terror.Error(err, "Failed to load reqpir case")
	}

	if repairCases == nil {
		return nil
	}

	// restart repair cases
	_, err = repairCases.UpdateAll(gamedb.StdConn, boiler.M{boiler.RepairCaseColumns.PausedAt: null.TimeFromPtr(nil)})
	if err != nil {
		gamelog.L.Error().Err(err).Interface("repair cases", repairCases).Msg("Failed to pause repair cases")
		return terror.Error(err, "Failed to pause repair case")
	}

	// collect all the system repair offers id
	repairOfferIDs := []string{}
	for _, rc := range repairCases {
		if rc.R == nil || rc.R.RepairOffers == nil {
			continue
		}
		for _, ro := range rc.R.RepairOffers {
			repairOfferIDs = append(repairOfferIDs, ro.ID)
		}
	}

	// unclose all the related system offers
	_, err = boiler.RepairOffers(
		boiler.RepairOfferWhere.ID.IN(repairOfferIDs),
	).UpdateAll(gamedb.StdConn,
		boiler.M{
			boiler.RepairOfferColumns.ClosedAt:       null.TimeFromPtr(nil),
			boiler.RepairOfferColumns.FinishedReason: null.StringFromPtr(nil),
		},
	)

	return nil
}

func (am *ArenaManager) RepairGameBlockProcesser(repairAgentID string, repairGameBlockLogID string, stackedBlockDimension *server.RepairGameBlockDimension, isFailed bool) (*server.RepairGameBlock, error) {
	l := gamelog.L.With().Str("func", "RepairGameBlockProcesser").Str("repair agent id", repairAgentID).Str("repair game block log id", repairGameBlockLogID).Logger()

	bombReduceBlockCount := db.GetIntWithDefault(db.KeyDeductBlockCountFromBomb, 3)
	requiredScore := db.GetIntWithDefault(db.KeyRequiredRepairStacks, 50)

	keys := []string{
		boiler.RepairGameBlockTriggerKeyM,
		boiler.RepairGameBlockTriggerKeyN,
		boiler.RepairGameBlockTriggerKeySPACEBAR,
	}

	// pre-load repair game block to shorten the process time
	repairGameBlocks, err := boiler.RepairGameBlocks(
		boiler.RepairGameBlockWhere.Type.NEQ(boiler.RepairGameBlockTypeEND),
	).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("Failed to laod repair block type.")
		return nil, terror.Error(err, "Failed to load repair block")
	}

	am.RepairGameBlockMx.Lock()
	defer am.RepairGameBlockMx.Unlock()

	// get the latest block
	repairGameBlockLogs, err := boiler.RepairGameBlockLogs(
		boiler.RepairGameBlockLogWhere.RepairAgentID.EQ(repairAgentID),
		qm.OrderBy(boiler.RepairGameBlockLogColumns.CreatedAt+" DESC"),
	).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("Failed to load latest repair game log.")
		return nil, terror.Error(err, "Failed to varify block.")
	}

	// this should never happen, but just in case
	if repairGameBlockLogs == nil {
		l.Error().Err(err).Msg("Can't find any repair game block log.")
		return nil, terror.Error(fmt.Errorf("empty repair game block log."))
	}

	lastBlock := repairGameBlockLogs[0]

	// skip, if the block is not the latest
	if lastBlock.ID != repairGameBlockLogID {
		return nil, terror.Error(fmt.Errorf("invalid repair game record"), "This is not the latest block.")
	}

	// check, if the block size grow
	if lastBlock.Width.LessThan(stackedBlockDimension.Width.RoundFloor(5)) || lastBlock.Depth.LessThan(stackedBlockDimension.Depth.RoundFloor(5)) {
		if lastBlock.RepairGameBlockType != boiler.RepairGameBlockTypeBOMB || isFailed {
			return nil, terror.Error(fmt.Errorf("cheat detected"), "The block grow bigger!")
		}
	}

	// update the latest block
	lastBlock.IsFailed = isFailed
	lastBlock.StackedAt = null.TimeFrom(time.Now())
	lastBlock.StackedWidth = decimal.NewNullDecimal(stackedBlockDimension.Width.Round(5))
	lastBlock.StackedDepth = decimal.NewNullDecimal(stackedBlockDimension.Depth.Round(5))

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		l.Error().Err(err).Msg("Failed to start db transaction.")
		return nil, terror.Error(err, "Failed to validate block.")
	}

	defer tx.Rollback()

	_, err = lastBlock.Update(tx, boil.Whitelist(
		boiler.RepairGameBlockLogColumns.IsFailed,
		boiler.RepairGameBlockLogColumns.StackedAt,
		boiler.RepairGameBlockLogColumns.StackedWidth,
		boiler.RepairGameBlockLogColumns.StackedDepth,
	))
	if err != nil {
		l.Error().Err(err).Interface("new repair game block", lastBlock).Msg("Failed to record repair game block.")
		return nil, terror.Error(err, "Failed to reocrd current block.")
	}

	var result *server.RepairGameBlock
	// if the block does not failed and the block is not a bomb
	if !isFailed || lastBlock.RepairGameBlockType == boiler.RepairGameBlockTypeBOMB {
		// generate next block
		var nextRepairBlock *boiler.RepairGameBlockLog

		// calculate score
		totalScore := 0
		for _, record := range repairGameBlockLogs {
			if record.RepairGameBlockType == boiler.RepairGameBlockTypeEND {
				return nil, terror.Error(fmt.Errorf("repair agent is already ended"), "Repair agent is closed.")
			}

			if !record.StackedAt.Valid || record.IsFailed {
				continue
			}

			if record.RepairGameBlockType == boiler.RepairGameBlockTypeBOMB {
				totalScore -= bombReduceBlockCount
				continue
			}

			totalScore += 1
		}

		if totalScore < 0 {
			totalScore = 0
		}

		if totalScore >= requiredScore {

			// generate the end block
			nextRepairBlock = &boiler.RepairGameBlockLog{
				RepairAgentID:       repairAgentID,
				RepairGameBlockType: boiler.RepairGameBlockTypeEND,
				SizeMultiplier:      decimal.NewFromInt(1),
				SpeedMultiplier:     decimal.NewFromInt(1),
				TriggerKey:          boiler.RepairGameBlockTriggerKeySPACEBAR,
				Width:               decimal.NewFromInt(10),
				Depth:               decimal.NewFromInt(10),
			}

			err = nextRepairBlock.Insert(tx, boil.Infer())
			if err != nil {
				l.Error().Err(err).Interface("end block", nextRepairBlock).Msg("Failed to generate the end block.")
				return nil, terror.Error(err, "Failed to generate the end block.")
			}

		} else {
			// otherwise, generate next block

			// remove bomb from the options, if the score is lower than what bomb will deduct
			if totalScore < bombReduceBlockCount {
				index := slices.IndexFunc(repairGameBlocks, func(rgl *boiler.RepairGameBlock) bool { return rgl.Type == boiler.RepairGameBlockTypeBOMB })
				repairGameBlocks = slices.Delete(repairGameBlocks, index, index+1)
			}

			// random select a block
			var pool []*boiler.RepairGameBlock
			for _, rgb := range repairGameBlocks {
				for i := decimal.Zero; i.LessThan(rgb.Probability.Mul(decimal.NewFromInt(100))); i = i.Add(decimal.NewFromInt(1)) {
					pool = append(pool, rgb)
				}
			}

			// shuffle the slice of opted in supporters
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

			block := pool[0]
			sizeMultiDiff := block.MaxSizeMultiplier.Sub(block.MinSizeMultiplier).Mul(decimal.NewFromFloat(rand.Float64()))
			sizeMultiplier := block.MaxSizeMultiplier.Sub(sizeMultiDiff)
			speedMultiDiff := block.MaxSpeedMultiplier.Sub(block.MinSpeedMultiplier).Mul(decimal.NewFromFloat(rand.Float64()))
			speedMultiplier := block.MinSpeedMultiplier.Add(speedMultiDiff).Mul(block.MaxSpeedMultiplier)

			nextRepairBlock = &boiler.RepairGameBlockLog{
				RepairAgentID:       repairAgentID,
				RepairGameBlockType: block.Type,
				SizeMultiplier:      sizeMultiplier,
				SpeedMultiplier:     speedMultiplier,
				TriggerKey:          keys[0],
				Width:               stackedBlockDimension.Width.Mul(sizeMultiplier).Round(5),
				Depth:               stackedBlockDimension.Depth.Mul(sizeMultiplier).Round(5),
			}

			err = nextRepairBlock.Insert(tx, boil.Infer())
			if err != nil {
				l.Error().Err(err).Interface("end block", nextRepairBlock).Msg("Failed to generate next block.")
				return nil, terror.Error(err, "Failed to generate next block.")
			}
		}

		result = &server.RepairGameBlock{
			ID:              nextRepairBlock.ID,
			Type:            nextRepairBlock.RepairGameBlockType,
			Key:             nextRepairBlock.TriggerKey,
			SpeedMultiplier: nextRepairBlock.SpeedMultiplier,
			TotalScore:      totalScore,
			Dimension: server.RepairGameBlockDimension{
				Width: nextRepairBlock.Width,
				Depth: nextRepairBlock.Depth,
			},
		}
	}

	// generate new block
	err = tx.Commit()
	if err != nil {
		l.Error().Err(err).Msg("Failed to commit db transaction.")
		return nil, terror.Error(err, "Faield to validate repair block.")
	}

	return result, nil
}
