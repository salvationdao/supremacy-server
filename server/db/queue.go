package db

import (
	"database/sql"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/exp/slices"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/friendsofgo/errors"
)

const FACTION_MECH_LIMIT = 3

func GetCollectionItemStatus(collectionItem boiler.CollectionItem) (*server.MechArenaInfo, error) {
	l := gamelog.L.With().Str("func", "GetCollectionItemStatus").Interface("collectionItem", collectionItem).Logger()

	// Check in marketplace
	now := time.Now()
	inMarketplace, err := collectionItem.ItemSales(
		boiler.ItemSaleWhere.EndAt.GT(now),
		boiler.ItemSaleWhere.SoldAt.IsNull(),
		boiler.ItemSaleWhere.DeletedAt.IsNull(),
	).Exists(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("failed to check in marketplace")
		return nil, err
	}

	if inMarketplace {
		return &server.MechArenaInfo{
			Status:    server.MechArenaStatusMarket,
			CanDeploy: false,
		}, nil
	}

	mechID := collectionItem.ItemID

	// Check in battle lobby
	battleLobbyMech, err := boiler.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.MechID.EQ(mechID),
		boiler.BattleLobbiesMechWhere.EndedAt.IsNull(),
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
		qm.Load(boiler.BattleLobbiesMechRels.BattleLobby),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		l.Error().Err(err).Msg("Failed to load the battle lobby of the mech.")
		return nil, terror.Error(err, "Failed to load the battle lobby of the mech.")
	}

	// if in battle lobby
	if battleLobbyMech != nil {
		mai := &server.MechArenaInfo{
			Status:    server.MechArenaStatusQueue,
			CanDeploy: false,
		}

		if battleLobbyMech.R != nil && battleLobbyMech.R.BattleLobby != nil {
			mai.BattleLobbyNumber = null.IntFrom(battleLobbyMech.R.BattleLobby.Number)
		}

		// if in battle
		if battleLobbyMech.AssignedToBattleID.Valid {
			mai.Status = server.MechArenaStatusBattle

			// if battle lobby is ready but haven't started
		} else if battleLobbyMech.LockedAt.Valid {
			// load battle lobby position
			battleLobbies, err := boiler.BattleLobbies(
				boiler.BattleLobbyWhere.ReadyAt.LTE(battleLobbyMech.LockedAt),
				boiler.BattleLobbyWhere.AssignedToBattleID.IsNull(),
				boiler.BattleLobbyWhere.EndedAt.IsNull(),
			).All(gamedb.StdConn)
			if err != nil {
				return nil, terror.Error(err, "Failed to load battle lobby position")
			}

			if battleLobbies != nil {
				mai.BattleLobbyQueuePosition = null.IntFrom(len(battleLobbies))
			}
		}

		return mai, nil
	}

	// Check if damaged
	rc, err := boiler.RepairCases(
		boiler.RepairCaseWhere.MechID.EQ(mechID),
		boiler.RepairCaseWhere.CompletedAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		l.Error().Err(err).Msg("failed to check if damaged")
		return nil, err
	}

	if rc != nil {
		canDeployRatio := GetDecimalWithDefault(KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))
		totalBlocks := TotalRepairBlocks(rc.MechID)
		if decimal.NewFromInt(int64(rc.BlocksRequiredRepair - rc.BlocksRepaired)).Div(decimal.NewFromInt(int64(totalBlocks))).GreaterThan(canDeployRatio) {
			// If less than 50% repaired
			return &server.MechArenaInfo{
				Status:    server.MechArenaStatusDamaged,
				CanDeploy: false,
			}, nil
		}
		return &server.MechArenaInfo{
			Status:    server.MechArenaStatusDamaged,
			CanDeploy: true,
		}, nil
	}

	return &server.MechArenaInfo{
		Status:    server.MechArenaStatusIdle,
		CanDeploy: true,
	}, nil
}

// FilterCanDeployMechIDs return the list of mech which are able to deploy
func FilterCanDeployMechIDs(mechIDs []string) ([]string, error) {
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

	if rcs == nil {
		return mechIDs, nil
	}

	canDeployRatio := GetDecimalWithDefault(KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))

	canDeployedMechIDs := []string{}
	for _, mechID := range mechIDs {
		index := slices.IndexFunc(rcs, func(rc *boiler.RepairCase) bool { return rc.MechID == mechID })

		// if the mech does not have repair case
		if index == -1 {
			canDeployedMechIDs = append(canDeployedMechIDs, mechID)
			continue
		}

		// append mech id, if the damaged ratio of the mech is not above the can deploy ratio
		rc := rcs[index]
		if decimal.NewFromInt(int64(rc.BlocksRequiredRepair - rc.BlocksRepaired)).Div(decimal.NewFromInt(int64(TotalRepairBlocks(rc.MechID)))).LessThanOrEqual(canDeployRatio) {
			canDeployedMechIDs = append(canDeployedMechIDs, mechID)
			continue
		}
	}

	return canDeployedMechIDs, nil
}

// FilterOutMechAlreadyInQueue return error if mech is already in queue
func FilterOutMechAlreadyInQueue(mechIDs []string) ([]string, error) {
	// check any mechs is already queued
	blm, err := boiler.BattleLobbiesMechs(
		boiler.BattleLobbiesMechWhere.MechID.IN(mechIDs),
		boiler.BattleLobbiesMechWhere.EndedAt.IsNull(),
		boiler.BattleLobbiesMechWhere.RefundTXID.IsNull(),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Strs("mech id list", mechIDs).Msg("Failed to check mech queue.")
		return nil, terror.Error(err, "Failed to check mech queue.")
	}

	if blm == nil {
		return mechIDs, nil
	}

	remainMechIDs := []string{}
	for _, mechID := range mechIDs {
		// skip, if mech is already queued
		if slices.IndexFunc(blm, func(bl *boiler.BattleLobbiesMech) bool { return bl.MechID == mechID }) != -1 {
			continue
		}

		// append mechs to the list
		remainMechIDs = append(remainMechIDs, mechID)
	}

	return remainMechIDs, nil
}

// CheckMechOwnership return error if any mechs is not available to queue
func CheckMechOwnership(playerID string, mechIDs []string) error {
	// ownership check
	mcis, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(mechIDs),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Strs("mech ids", mechIDs).Err(err).Msg("unable to retrieve mech collection item from hash")
		return err
	}

	if len(mcis) != len(mechIDs) {
		return terror.Error(fmt.Errorf("contain non-mech assest"), "The list contains non-mech asset.")
	}

	for _, mci := range mcis {
		if mci.XsynLocked {
			err := fmt.Errorf("mech is locked to xsyn locked")
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is xsyn locked")
			return err
		}

		if mci.LockedToMarketplace {
			err := fmt.Errorf("mech is listed in marketplace")
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Err(err).Msg("war machine is listed in marketplace")
			return err
		}

		battleReady, err := MechBattleReady(mci.ItemID)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load battle ready status")
			return err
		}

		if !battleReady {
			gamelog.L.Error().Str("log_name", "battle arena").Str("mech_id", mci.ItemID).Msg("war machine is not available for queuing")
			return fmt.Errorf("mech is cannot be used")
		}

		if mci.OwnerID != playerID {
			return terror.Error(fmt.Errorf("does not own the mech"), "This mech is not owned by you")
		}
	}

	return nil
}
