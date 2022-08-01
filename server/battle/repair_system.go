package battle

import (
	"fmt"
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
	"time"
)

type RepairOfferClose struct {
	OfferIDs          []string
	OfferClosedReason string
	AgentClosedReason string
}

func (arena *Arena) RepairOfferCleaner() {
	expiryCheckChan := make(chan bool)
	go func(expiryCheckChan chan bool) {
		for {
			time.Sleep(5 * time.Second)
			expiryCheckChan <- true
		}
	}(expiryCheckChan)

	for {
		select {
		case <-expiryCheckChan:
			now := time.Now()
			// expire repair offer
			ros, err := boiler.RepairOffers(
				boiler.RepairOfferWhere.ExpiresAt.LTE(now),
				boiler.RepairOfferWhere.ClosedAt.IsNull(),
				qm.Load(boiler.RepairOfferRels.RepairCase),
				qm.Load(boiler.RepairOfferRels.RepairBlocks),
				qm.Load(
					boiler.RepairOfferRels.RepairAgents,
					boiler.RepairAgentWhere.FinishedAt.IsNull(),
				),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to get repair offer")
				continue
			}

			if len(ros) == 0 {
				continue
			}

			err = arena.closeRepairOffers(ros, boiler.RepairFinishReasonEXPIRED, boiler.RepairAgentFinishReasonEXPIRED)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to close expired repair offers.")
				continue
			}

		case roc := <-arena.RepairOfferCloseChan:
			ros, err := boiler.RepairOffers(
				boiler.RepairOfferWhere.ID.IN(roc.OfferIDs),
				boiler.RepairOfferWhere.ClosedAt.IsNull(), // double check it is not closed yet
				qm.Load(boiler.RepairOfferRels.RepairCase),
				qm.Load(boiler.RepairOfferRels.RepairBlocks),
				qm.Load(
					boiler.RepairOfferRels.RepairAgents,
					boiler.RepairAgentWhere.FinishedAt.IsNull(),
				),
			).All(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to get repair offers")
				continue
			}

			if len(ros) == 0 {
				continue
			}

			err = arena.closeRepairOffers(ros, roc.OfferClosedReason, roc.AgentClosedReason)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to close repair offers.")
				continue
			}
		}
	}
}

func (arena *Arena) closeRepairOffers(ros boiler.RepairOfferSlice, offerCloseReason string, agentCloseReason string) error {
	now := time.Now()
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to begin db transaction.")
		return terror.Error(err, "Failed to begin db transaction.")
	}

	defer tx.Rollback()

	for _, ro := range ros {
		ro.ClosedAt = null.TimeFrom(now)
		ro.FinishedReason = null.StringFrom(offerCloseReason)
		_, err := ro.Update(tx, boil.Whitelist(
			boiler.RepairOfferColumns.ClosedAt,
			boiler.RepairOfferColumns.FinishedReason,
		))
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to close expired repair offer.")
			return terror.Error(err, "Failed to close expired repair offer.")
		}

		// broadcast close offer
		rc := ro.R.RepairCase
		sro := &server.RepairOffer{
			RepairOffer:          ro,
			BlocksRequiredRepair: rc.BlocksRequiredRepair,
			BlocksRepaired:       rc.BlocksRepaired,
			SupsWorthPerBlock:    ro.OfferedSupsAmount.Div(decimal.NewFromInt(int64(ro.BlocksTotal))),
			WorkingAgentCount:    0,
		}

		ws.PublishMessage(fmt.Sprintf("/public/repair_offer/%s", ro.ID), server.HubKeyRepairOfferSubscribe, sro)
		ws.PublishMessage("/public/repair_offer/update", server.HubKeyRepairOfferUpdateSubscribe, []*server.RepairOffer{sro})
		if ro.OfferedByID.Valid {
			ws.PublishMessage(fmt.Sprintf("/public/mech/%s/active_repair_offer", rc.MechID), server.HubKeyMechActiveRepairOffer, sro)
		}

		if ro.R == nil {
			continue
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
			continue
		}

		totalRefundBlocks := ro.BlocksTotal
		if ro.R.RepairBlocks != nil {
			totalRefundBlocks = totalRefundBlocks - len(ro.R.RepairBlocks)
		}

		if totalRefundBlocks > 0 {
			amount := ro.OfferedSupsAmount.Div(decimal.NewFromInt(int64(ro.BlocksTotal))).Mul(decimal.NewFromInt(int64(totalRefundBlocks)))

			if amount.Equal(decimal.Zero) {
				continue
			}

			// refund reward
			_, err = arena.RPCClient.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           uuid.Must(uuid.FromString(server.RepairCenterUserID)),
				ToUserID:             uuid.Must(uuid.FromString(ro.OfferedByID.String)),
				Amount:               amount.StringFixed(0),
				TransactionReference: server.TransactionReference(fmt.Sprintf("refund_unclaimed_repair_offer_reward|%s|%d", ro.ID, time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupRepair),
				Description:          "refund unclaimed repair offer reward.",
				NotSafe:              true,
			})
			if err != nil {
				gamelog.L.Error().
					Str("player_id", ro.OfferedByID.String).
					Str("repair offer id", ro.ID).
					Str("amount", amount.StringFixed(0)).
					Err(err).Msg("Failed to refund unclaimed repair offer reward.")
				return terror.Error(err, "Failed to refund unclaimed repair offer reward.")
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to commit db transaction.")
		return terror.Error(err, "Failed to commit db transaction.")
	}

	return nil
}

// RegisterMechRepairCase insert mech repair case and track repair stack
func RegisterMechRepairCase(mechID string, modelID string, maxHealth uint32, remainHealth uint32) error {
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
	model, err := boiler.FindMechModel(gamedb.StdConn, modelID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech model id", modelID).Msg("Failed to load mech model for repair block detail.")
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

	return nil
}
