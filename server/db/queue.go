package db

import (
	"database/sql"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"golang.org/x/exp/slices"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/shopspring/decimal"
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
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		l.Error().Err(err).Msg("Failed to load the battle lobby of the mech.")
		return nil, terror.Error(err, "Failed to load the battle lobby of the mech.")
	}

	// if in battle lobby
	if battleLobbyMech != nil {
		mai := &server.MechArenaInfo{
			Status:              server.MechArenaStatusQueue,
			CanDeploy:           false,
			BattleLobbyIsLocked: battleLobbyMech.LockedAt.Valid,
		}

		// if in battle
		if battleLobbyMech.AssignedToBattleID.Valid {
			mai.Status = server.MechArenaStatusBattle
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

// CheckMechQueueAuthorisation return error if any mechs is not available to queue
func CheckMechQueueAuthorisation(playerID string, factionID string, mechIDs []string) ([]string, error) {
	if len(mechIDs) == 0 {
		return []string{}, nil
	}

	mechIDWhereIn := fmt.Sprintf(" AND %s IN (", boiler.CollectionItemTableColumns.ItemID)
	for i, mechID := range mechIDs {
		_, err := uuid.FromString(mechID)
		if err != nil {
			return nil, terror.Error(err, "Invalid mech id")
		}

		mechIDWhereIn += "'" + mechID + "'"

		if i < len(mechIDs)-1 {
			mechIDWhereIn += ","
			continue
		}

		mechIDWhereIn += ")"
	}

	queries := []qm.QueryMod{
		qm.Select(
			boiler.MechTableColumns.ID,
			boiler.CollectionItemTableColumns.OwnerID,
			boiler.CollectionItemTableColumns.LockedToMarketplace,
			boiler.CollectionItemTableColumns.MarketLocked,
			boiler.CollectionItemTableColumns.XsynLocked,
			boiler.MechTableColumns.PowerCoreID,
			fmt.Sprintf(
				"COALESCE((SELECT COUNT(*) > 0 FROM %s WHERE %s = %s), FALSE) AS has_weapon",
				boiler.TableNames.MechWeapons,
				boiler.MechWeaponTableColumns.ChassisID,
				boiler.MechTableColumns.ID,
			),
			fmt.Sprintf(
				"COALESCE((SELECT %s <= NOW() FROM %s WHERE %s = %s), TRUE) AS is_available",
				boiler.AvailabilityTableColumns.AvailableAt,
				boiler.TableNames.Availabilities,
				boiler.AvailabilityTableColumns.ID,
				boiler.BlueprintMechTableColumns.AvailabilityID,
			),
			fmt.Sprintf(
				"(SELECT %s FROM %s WHERE %s = %s) AS staked_on_faction_id",
				boiler.StakedMechTableColumns.FactionID,
				boiler.TableNames.StakedMechs,
				boiler.StakedMechTableColumns.MechID,
				boiler.CollectionItemTableColumns.ItemID,
			),
		),
		qm.From(fmt.Sprintf(
			"(SELECT * FROM %s WHERE %s = '%s' %s) %s",
			boiler.TableNames.CollectionItems,
			boiler.CollectionItemTableColumns.ItemType,
			boiler.ItemTypeMech,
			mechIDWhereIn,
			boiler.TableNames.CollectionItems,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.Mechs,
			boiler.MechTableColumns.ID,
			boiler.CollectionItemTableColumns.ItemID,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintMechs,
			boiler.MechTableColumns.BlueprintID,
			boiler.BlueprintMechTableColumns.ID,
		)),
	}

	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("log_name", "battle arena").Strs("mech ids", mechIDs).Err(err).Msg("unable to load mech detail")
		return nil, err
	}

	availableList := []string{}

	for rows.Next() {
		mechID := ""
		ownerID := ""
		lockedToMarketplace := false
		marketLocked := false
		xsynLocked := false
		powerCoreID := null.StringFromPtr(nil)
		hasWeapon := false
		isAvailable := false
		stakedOnFactionID := null.StringFromPtr(nil)

		err = rows.Scan(
			&mechID,
			&ownerID,
			&lockedToMarketplace,
			&marketLocked,
			&xsynLocked,
			&powerCoreID,
			&hasWeapon,
			&isAvailable,
			&stakedOnFactionID,
		)
		if err != nil {
			return nil, terror.Error(err, "Failed to scan mech queue check")
		}

		if lockedToMarketplace {
			continue
		}

		if marketLocked {
			continue
		}

		if xsynLocked {
			continue
		}

		if !isAvailable {
			continue
		}

		if !powerCoreID.Valid {
			continue
		}

		if !hasWeapon {
			continue
		}

		if stakedOnFactionID.Valid {
			// check faction id if the mech is staked in faction list
			if stakedOnFactionID.String != factionID {
				continue
			}
		} else {
			// otherwise, check owner id
			if ownerID != playerID {
				continue
			}
		}

		availableList = append(availableList, mechID)
	}
	return availableList, nil
}
