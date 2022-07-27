package battle

import (
	"fmt"
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
	"time"
)

func RepairOfferCleaner() {
	for {
		time.Sleep(10 * time.Second)

		// expire repair offer
		expiredOfferIDs, err := db.CloseExpiredRepairOffers()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to close expired repair offers.")
			continue
		}

		for _, offerID := range expiredOfferIDs {
			// broadcast close message
			ro, err := db.RepairOfferDetail(offerID)
			if err != nil {
				gamelog.L.Error().Err(err).Str("repair offer id", offerID).Msg("Failed to load repair offers.")
				continue
			}

			ws.PublishMessage(fmt.Sprintf("/public/repair_offer/%s", ro.ID), server.HubKeyRepairOfferSubscribe, ro)
		}

		// expire repair agent
		_, err = boiler.RepairAgents(
			boiler.RepairAgentWhere.FinishedAt.IsNull(),
			qm.Where(
				fmt.Sprintf(
					"EXISTS (SELECT 1 FROM %s WHERE %s = %s AND %s NOTNULL)",
					boiler.TableNames.RepairOffers,
					qm.Rels(boiler.TableNames.RepairOffers, boiler.RepairOfferColumns.ID),
					qm.Rels(boiler.TableNames.RepairAgents, boiler.RepairAgentColumns.RepairOfferID),
					qm.Rels(boiler.TableNames.RepairOffers, boiler.RepairOfferColumns.ClosedAt),
				),
			),
		).UpdateAll(gamedb.StdConn,
			boiler.M{
				boiler.RepairAgentColumns.FinishedAt:     null.TimeFrom(time.Now()),
				boiler.RepairAgentColumns.FinishedReason: null.StringFrom(boiler.RepairAgentFinishReasonEXPIRED),
			},
		)
	}
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
		IsSelf:            true,
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

	ws.PublishMessage(fmt.Sprintf("/public/mech/%s/repair_case", rc.MechID), server.HubKeyMechRepairCase, &server.MechRepairStatus{
		RepairCase:    rc,
		BlocksDefault: model.RepairBlocks,
	})

	return nil
}
