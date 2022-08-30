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
			now := time.Now()
			pms, err := boiler.PlayerMechRepairBays(
				boiler.PlayerMechRepairBayWhere.Status.EQ(boiler.RepairBayStatusREPAIRING),
				boiler.PlayerMechRepairBayWhere.NextRepairTime.LTE(null.TimeFrom(now)),
				qm.Load(boiler.PlayerMechRepairBayRels.RepairCase),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to load repairing cases.")
				continue
			}

			wg := sync.WaitGroup{}
			for _, pm := range pms {
				wg.Add(1)
				go func(playerMechRepairBay *boiler.PlayerMechRepairBay) {
					defer wg.Done()
					if playerMechRepairBay.R == nil || playerMechRepairBay.R.RepairCase == nil {
						// todo: log warning
						return
					}

					rc := playerMechRepairBay.R.RepairCase

					// load system repair offer
					systemOffer, err := rc.RepairOffers(
						boiler.RepairOfferWhere.OfferedByID.IsNull(),
					).One(gamedb.StdConn)
					if err != nil && !errors.Is(err, sql.ErrNoRows) {
						// todo: log error
						return
					}

					if systemOffer == nil {
						// todo: log warning
						return
					}

					if systemOffer.ClosedAt.Valid || rc.CompletedAt.Valid {
						// todo: swap bay
						return
					}

					// do repair

					// check repair complete

					// return if complete

					// otherwise swap bay

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
