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
	"sync"
	"time"
)

func (am *ArenaManager) SendRepairFunc(fn func() error) error {
	var err error

	wg := sync.WaitGroup{}
	wg.Add(1)

	am.RepairOfferFuncChan <- func() {
		err = fn()
		wg.Done()
	}

	wg.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (am *ArenaManager) RepairOfferCleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	repairBayTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			// expire repair offer
			ros, err := boiler.RepairOffers(
				boiler.RepairOfferWhere.ExpiresAt.LTE(now),
				boiler.RepairOfferWhere.ClosedAt.IsNull(),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to get repair offer")
				continue
			}

			if len(ros) == 0 {
				continue
			}

			roIDs := []string{}
			for _, ro := range ros {
				roIDs = append(roIDs, ro.ID)
			}

			err = am.CloseRepairOffers(roIDs, boiler.RepairFinishReasonEXPIRED, boiler.RepairAgentFinishReasonEXPIRED)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to close expired repair offers.")
				continue
			}

		case <-repairBayTicker.C:
			nextRepairDurationSeconds := db.GetIntWithDefault(db.KeyAutoRepairDurationSeconds, 600)

			now := time.Now()
			pms, err := boiler.PlayerMechRepairSlots(
				boiler.PlayerMechRepairSlotWhere.Status.EQ(boiler.RepairSlotStatusREPAIRING),
				boiler.PlayerMechRepairSlotWhere.NextRepairTime.LTE(null.TimeFrom(now)),
				qm.Load(boiler.PlayerMechRepairSlotRels.RepairCase),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to load repairing cases.")
				continue
			}

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

					// if not complete
					if rc.BlocksRequiredRepair > rc.BlocksRepaired {
						// set next repair time
						playerMechRepairSlot.NextRepairTime = null.TimeFrom(now.Add(time.Duration(nextRepairDurationSeconds) * time.Second))
						_, err = playerMechRepairSlot.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerMechRepairSlotColumns.NextRepairTime))
						if err != nil {
							gamelog.L.Error().Err(err).Interface("repair slot", playerMechRepairSlot).Msg("Failed to update next repair time of the repair slot.")
						}
						return
					}

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

				}(pm)
			}

			wg.Wait()

		case fn := <-am.RepairOfferFuncChan:
			fn()
		}
	}
}

// CloseRepairOffers close the given repair offer
// REMINDER: this function is protected by channel
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
