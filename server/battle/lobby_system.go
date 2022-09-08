package battle

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
)

func (am *ArenaManager) SendBattleQueueFunc(fn func() error) error {
	am.BattleQueueFuncMx.Lock()
	defer am.BattleQueueFuncMx.Unlock()
	return fn()
}

func (am *ArenaManager) BroadcastBattleLobbyUpdate(battleLobbyID string) {
	bl, err := boiler.BattleLobbies(
		boiler.BattleLobbyWhere.ID.EQ(battleLobbyID),
		qm.Load(
			boiler.BattleLobbyRels.BattleLobbiesMechs,
			boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
			boiler.BattleLobbiesMechWhere.DeletedAt.IsNull(),
		),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("battle lobby id", battleLobbyID).Msg("Failed to query battle lobby")
		return
	}

	if bl == nil {
		return
	}

	ws.PublishMessage("/secure/battle_lobbies", server.HubKeyBattleLobbyListUpdate, server.BattleLobbiesFromBoiler([]*boiler.BattleLobby{bl}))
}

// CheckMechCanQueue return error if any mechs is not available to queue
func (am *ArenaManager) CheckMechCanQueue(playerID string, mechIDs []string) ([]*boiler.RepairCase, error) {
	mcis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(mechIDs),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Strs("mech ids", mechIDs).Err(err).Msg("unable to retrieve mech collection item from hash")
		return nil, err
	}

	if len(mcis) != len(mechIDs) {
		return nil, terror.Error(fmt.Errorf("contain non-mech assest"), "The list contains non-mech asset.")
	}

	for _, mci := range mcis {
		if mci.XsynLocked {
			err := fmt.Errorf("mech is locked to xsyn locked")
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is xsyn locked")
			return nil, err
		}

		if mci.LockedToMarketplace {
			err := fmt.Errorf("mech is listed in marketplace")
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is listed in marketplace")
			return nil, err
		}

		battleReady, err := db.MechBattleReady(mci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load battle ready status")
			return nil, err
		}

		if !battleReady {
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Msg("war machine is not available for queuing")
			return nil, fmt.Errorf("mech is cannot be used")
		}

		if mci.OwnerID != playerID {
			return nil, terror.Error(fmt.Errorf("does not own the mech"), "This mech is not owned by you")
		}
	}

	// check mech is still in repair
	rcs, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.IN(mechIDs),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
		qm.Load(boiler.RepairCaseRels.RepairOffers, boiler.RepairOfferWhere.ClosedAt.IsNull()),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Strs("mech ids", mechIDs).Msg("Failed to get repair case")
		return nil, terror.Error(err, "Failed to queue mech.")
	}

	if rcs != nil && len(rcs) > 0 {
		canDeployRatio := db.GetDecimalWithDefault(db.KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))

		for _, rc := range rcs {
			totalBlocks := db.TotalRepairBlocks(rc.MechID)

			// broadcast current mech stat if repair is above can deploy ratio
			if decimal.NewFromInt(int64(rc.BlocksRequiredRepair - rc.BlocksRepaired)).Div(decimal.NewFromInt(int64(totalBlocks))).GreaterThan(canDeployRatio) {
				// if mech has more than half of the block to repair
				return nil, terror.Error(fmt.Errorf("mech is not fully recovered"), "One of your mechs is still under repair.")
			}
		}
	}

	return rcs, nil
}

// CheckMechAlreadyInQueue return error if mech is already in queue
func (am *ArenaManager) CheckMechAlreadyInQueue(mechIDs []string) error {
	// check any mechs is already queued
	blm, err := boiler.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.MechID.IN(mechIDs),
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
		qm.Where(
			fmt.Sprintf(
				"EXISTS ( SELECT 1 FROM %s WHERE %s = %s AND %s ISNULL)",
				boiler.TableNames.BattleLobbies,
				boiler.BattleLobbyTableColumns.ID,
				boiler.BattleLobbiesMechTableColumns.BattleLobbyID,
				boiler.BattleLobbyTableColumns.FinishedAt,
			),
		),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Strs("mech id list", mechIDs).Msg("Failed to check mech queue.")
		return terror.Error(err, "Failed to check mech queue.")
	}

	// return error, if any mech is in queue
	if blm != nil {
		return terror.Error(fmt.Errorf("mech already in queue"), "Your mech is already in queue.")
	}

	// check battle queue
	bqs, err := boiler.BattleQueues(
		boiler.BattleQueueWhere.MechID.IN(mechIDs),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Strs("mech id list", mechIDs).Msg("Failed to check mech queue.")
		return terror.Error(err, "Failed to check mech queue.")
	}

	if bqs != nil {
		return terror.Error(fmt.Errorf("mech already in queue"), "Your mech is already in queue.")
	}

	return nil
}
