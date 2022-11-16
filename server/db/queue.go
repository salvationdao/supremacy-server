package db

import (
	"database/sql"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/exp/slices"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/friendsofgo/errors"
)

const FACTION_MECH_LIMIT = 3

func GetMechQueueStatus(mechID string) (*server.MechArenaInfo, error) {
	l := gamelog.L.With().Str("func", "GetMechQueueStatus").Interface("mechID", mechID).Logger()

	_, err := uuid.FromString(mechID)
	if err != nil {
		return nil, terror.Error(fmt.Errorf("invalid mech id"), "Invalid mech id")
	}

	queries := []qm.QueryMod{
		qm.Select(
			fmt.Sprintf(
				"COALESCE((SELECT TRUE FROM %s WHERE %s = %s AND %s ISNULL AND %s ISNULL AND %s > NOW()), FALSE) AS in_marketplace",
				boiler.TableNames.ItemSales,
				boiler.ItemSaleTableColumns.CollectionItemID,
				boiler.CollectionItemTableColumns.ID,
				boiler.ItemSaleTableColumns.SoldAt,
				boiler.ItemSaleTableColumns.DeletedAt,
				boiler.ItemSaleTableColumns.EndAt,
			),
			fmt.Sprintf("%s NOTNULL AS is_queued", boiler.BattleLobbiesMechTableColumns.CreatedAt),
			fmt.Sprintf("%s NOTNULL AS is_locked_in_lobby", boiler.BattleLobbiesMechTableColumns.LockedAt),
			fmt.Sprintf("%s NOTNULL AS is_in_battle", boiler.BattleLobbiesMechTableColumns.AssignedToBattleID),
			boiler.BlueprintMechTableColumns.RepairBlocks,
			fmt.Sprintf(
				"COALESCE(%s - %s, 0) AS damaged_block",
				boiler.RepairCaseTableColumns.BlocksRequiredRepair,
				boiler.RepairCaseTableColumns.BlocksRepaired,
			),
		),

		qm.From(fmt.Sprintf(
			"(SELECT * FROM %s WHERE %s = '%s' AND %s = '%s') %s",
			boiler.TableNames.CollectionItems,
			boiler.CollectionItemTableColumns.ItemType,
			boiler.ItemTypeMech,
			boiler.CollectionItemTableColumns.ItemID,
			mechID,
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
			boiler.BlueprintMechTableColumns.ID,
			boiler.MechTableColumns.BlueprintID,
		)),

		qm.LeftOuterJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s ISNULL AND %s ISNULL",
			boiler.TableNames.RepairCases,
			boiler.RepairCaseTableColumns.MechID,
			boiler.MechTableColumns.ID,
			boiler.RepairCaseTableColumns.CompletedAt,
			boiler.RepairCaseTableColumns.DeletedAt,
		)),

		qm.LeftOuterJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s ISNULL AND %s ISNULL AND %s ISNULL",
			boiler.TableNames.BattleLobbiesMechs,
			boiler.BattleLobbiesMechTableColumns.MechID,
			boiler.MechTableColumns.ID,
			boiler.BattleLobbiesMechTableColumns.EndedAt,
			boiler.BattleLobbiesMechTableColumns.RefundTXID,
			boiler.BattleLobbiesMechTableColumns.DeletedAt,
		)),
	}

	isInMarket := false
	isQueued := false
	isLockedInLobby := false
	isInBattle := false
	totalRepairBlocks := int64(0)
	damagedRepairBlocks := int64(0)

	err = boiler.NewQuery(queries...).QueryRow(gamedb.StdConn).Scan(
		&isInMarket,
		&isQueued,
		&isLockedInLobby,
		&isInBattle,
		&totalRepairBlocks,
		&damagedRepairBlocks,
	)
	if err != nil {
		l.Error().Err(err).Msg("Failed to scan mech queue status.")
		return nil, terror.Error(err, "Failed to scan mech queue status.")
	}

	if isInMarket {
		return &server.MechArenaInfo{
			Status:    server.MechArenaStatusMarket,
			CanDeploy: false,
		}, nil
	}

	if isQueued {
		mai := &server.MechArenaInfo{
			Status:              server.MechArenaStatusQueue,
			CanDeploy:           false,
			BattleLobbyIsLocked: isLockedInLobby,
		}

		if isInBattle {
			mai.Status = server.MechArenaStatusBattle
		}

		return mai, nil
	}

	if damagedRepairBlocks > 0 {
		return &server.MechArenaInfo{
			Status:    server.MechArenaStatusDamaged,
			CanDeploy: decimal.NewFromInt(damagedRepairBlocks).Div(decimal.NewFromInt(totalRepairBlocks)).GreaterThan(GetDecimalWithDefault(KeyCanDeployDamagedRatio, decimal.NewFromFloat(0.5))),
		}, nil
	}

	return &server.MechArenaInfo{
		Status:    server.MechArenaStatusIdle,
		CanDeploy: true,
	}, nil
}

// OverDamagedMechFilter return the list of mech which are able to deploy
func OverDamagedMechFilter(mechIDs []string) ([]string, error) {
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

// NonQueuedMechFilter return error if mech is already in queue
func NonQueuedMechFilter(mechIDs []string) ([]string, error) {
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

type MechQueueAuthorisationData struct {
	MechID              string      `db:"mech_id"`
	OwnerID             string      `db:"owner_id"`
	LockedToMarketplace bool        `db:"locked_to_marketplace"`
	MarketLocked        bool        `db:"market_locked"`
	XsynLocked          bool        `db:"xsyn_locked"`
	PowerCoreID         null.String `db:"power_core_id"`
	HasWeapon           bool        `db:"has_weapon"`
	IsAvailable         bool        `db:"is_available"`
	StakedOnFactionID   null.String `db:"staked_on_faction_id"`
}

func MechsQueueAuthorisationDataGet(mechIDs []string) ([]*MechQueueAuthorisationData, error) {
	if len(mechIDs) == 0 {
		return []*MechQueueAuthorisationData{}, nil
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

	result := []*MechQueueAuthorisationData{}
	for rows.Next() {
		mqa := &MechQueueAuthorisationData{}
		err = rows.Scan(
			&mqa.MechID,
			&mqa.OwnerID,
			&mqa.LockedToMarketplace,
			&mqa.MarketLocked,
			&mqa.XsynLocked,
			&mqa.PowerCoreID,
			&mqa.HasWeapon,
			&mqa.IsAvailable,
			&mqa.StakedOnFactionID,
		)
		if err != nil {
			return nil, terror.Error(err, "Failed to scan mech queue check")
		}
		result = append(result, mqa)
	}

	return result, nil
}
