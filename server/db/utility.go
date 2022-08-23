package db

import (
	"database/sql"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func getDefaultUtilityQueryMods() []qm.QueryMod {
	return []qm.QueryMod{
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeUtility),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Utility,
			qm.Rels(boiler.TableNames.Utility, boiler.UtilityColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		// join utility blueprint
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintUtility,
			qm.Rels(boiler.TableNames.BlueprintUtility, boiler.BlueprintUtilityColumns.ID),
			qm.Rels(boiler.TableNames.Utility, boiler.UtilityColumns.BlueprintID),
		)),
	}
}

type UtilityListOpts struct {
	Search                 string
	Filter                 *ListFilterRequest
	Sort                   *ListSortRequest
	SortBy                 string
	SortDir                SortByDir
	PageSize               int
	Page                   int
	OwnerID                string
	DisplayXsynLocked      bool
	DisplayHidden          bool
	ExcludeMarketLocked    bool
	IncludeMarketListed    bool
	ExcludeMechLocked      bool
	ExcludeIDs             []string
	FilterRarities         []string `json:"rarities"`
	FilterTypes            []string `json:"types"`
	FilterEquippedStatuses []string `json:"equipped_statuses"`
}

type UtilityStatFilterRange struct {
	Min null.Int `json:"min"`
	Max null.Int `json:"max"`
}

func GenerateUtilityStatFilterQueryMods(column string, filter *UtilityStatFilterRange) []qm.QueryMod {
	output := []qm.QueryMod{}
	if filter == nil {
		return output
	}
	if filter.Min.Valid {
		output = append(output, qm.Where(qm.Rels(boiler.TableNames.Utility, column)+" >= ?", filter.Min))
	}
	if filter.Max.Valid {
		output = append(output, qm.Where(qm.Rels(boiler.TableNames.Utility, column)+" <= ?", filter.Max))
	}
	return output
}

func UtilityList(opts *UtilityListOpts) (int64, []*server.Utility, error) {
	queryMods := getDefaultUtilityQueryMods()

	if opts.OwnerID != "" {
		queryMods = append(queryMods, boiler.CollectionItemWhere.OwnerID.EQ(opts.OwnerID))
	}
	if !opts.DisplayXsynLocked {
		queryMods = append(queryMods, boiler.CollectionItemWhere.XsynLocked.EQ(false))
	}
	if !opts.IncludeMarketListed {
		queryMods = append(queryMods, boiler.CollectionItemWhere.LockedToMarketplace.EQ(false))
	}
	if opts.ExcludeMarketLocked {
		queryMods = append(queryMods, boiler.CollectionItemWhere.MarketLocked.EQ(false))
	}
	if !opts.DisplayHidden {
		queryMods = append(queryMods, boiler.CollectionItemWhere.AssetHidden.IsNull())
	}
	if opts.ExcludeMechLocked {
		queryMods = append(queryMods,
			boiler.UtilityWhere.LockedToMech.EQ(false),
		)
	}

	// Filters
	if opts.Filter != nil {
		// if we have filter
		for i, f := range opts.Filter.Items {
			// validate it is the right table and valid column
			if f.Table == boiler.TableNames.Utility && IsMechColumn(f.Column) {
				queryMods = append(queryMods, GenerateListFilterQueryMod(*f, i+1, opts.Filter.LinkOperator))
			}
		}
	}

	if len(opts.ExcludeIDs) > 0 {
		queryMods = append(queryMods, boiler.UtilityWhere.ID.NIN(opts.ExcludeIDs))
	}

	if len(opts.FilterRarities) > 0 {
		queryMods = append(queryMods, boiler.BlueprintUtilityWhere.Tier.IN(opts.FilterRarities))
	}

	if len(opts.FilterEquippedStatuses) > 0 {
		showEquipped := false
		showUnequipped := false
		for _, s := range opts.FilterEquippedStatuses {
			if s == "equipped" {
				showEquipped = true
			} else if s == "unequipped" {
				showUnequipped = true
			}
			if showEquipped && showUnequipped {
				break
			}
		}

		if showEquipped && !showUnequipped {
			queryMods = append(queryMods, boiler.UtilityWhere.EquippedOn.IsNotNull())
		} else if showUnequipped && !showEquipped {
			queryMods = append(queryMods, boiler.UtilityWhere.EquippedOn.IsNull())
		}
	}

	if len(opts.FilterTypes) > 0 {
		queryMods = append(queryMods, boiler.BlueprintUtilityWhere.Type.IN(opts.FilterTypes))
	}

	// Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"((to_tsvector('english', %s) @@ to_tsquery(?))",
					qm.Rels(boiler.TableNames.BlueprintUtility, boiler.BlueprintUtilityColumns.Label),
				),
					xSearch,
				))
		}
	}
	boil.DebugMode = true
	total, err := boiler.CollectionItems(
		queryMods...,
	).Count(gamedb.StdConn)
	boil.DebugMode = false
	if err != nil {
		return 0, nil, err
	}
	// Limit/Offset
	if opts.PageSize > 0 {
		queryMods = append(queryMods, qm.Limit(opts.PageSize))
	}
	if opts.Page > 0 {
		queryMods = append(queryMods, qm.Offset(opts.PageSize*(opts.Page-1)))
	}

	// Build query
	queryMods = append(queryMods,
		qm.Select(
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.CollectionSlug),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Hash),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.TokenID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.OwnerID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Tier),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.MarketLocked),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.XsynLocked),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.LockedToMarketplace),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.AssetHidden),
			qm.Rels(boiler.TableNames.Utility, boiler.UtilityColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintUtility, boiler.BlueprintUtilityColumns.Label),
		),
		qm.From(boiler.TableNames.CollectionItems),
	)

	if opts.SortBy != "" && opts.SortDir.IsValid() {
		if opts.SortBy == "alphabetical" {
			queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.BlueprintUtility, boiler.BlueprintUtilityColumns.Label), opts.SortDir)))
		} else if opts.SortBy == "rarity" {
			queryMods = append(queryMods, GenerateTierSort(qm.Rels(boiler.TableNames.BlueprintUtility, boiler.BlueprintUtilityColumns.Tier), opts.SortDir))
		}
	} else {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s ASC", qm.Rels(boiler.TableNames.BlueprintUtility, boiler.BlueprintUtilityColumns.Label))))
	}

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var utilities []*server.Utility
	for rows.Next() {
		pc := &server.Utility{
			CollectionItem: &server.CollectionItem{},
		}

		scanArgs := []interface{}{
			&pc.CollectionItem.CollectionSlug,
			&pc.CollectionItem.Hash,
			&pc.CollectionItem.TokenID,
			&pc.CollectionItem.OwnerID,
			&pc.CollectionItem.Tier,
			&pc.CollectionItem.ItemType,
			&pc.CollectionItem.MarketLocked,
			&pc.CollectionItem.XsynLocked,
			&pc.CollectionItem.LockedToMarketplace,
			&pc.CollectionItem.AssetHidden,
			&pc.ID,
			&pc.Label,
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, utilities, err
		}
		utilities = append(utilities, pc)
	}

	return total, utilities, nil
}

func InsertNewUtility(tx boil.Executor, ownerID uuid.UUID, utility *server.BlueprintUtility) (*server.Utility, error) {
	newUtility := boiler.Utility{
		BrandID:               utility.BrandID,
		Label:                 utility.Label,
		BlueprintID:           utility.ID,
		GenesisTokenID:        utility.GenesisTokenID,
		LimitedReleaseTokenID: utility.LimitedReleaseTokenID,
		Type:                  utility.Type,
	}

	err := newUtility.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		utility.Collection,
		boiler.ItemTypeUtility,
		newUtility.ID,
		utility.Tier,
		ownerID.String(),
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	// insert the extra table depending on utility type
	switch newUtility.Type {
	case boiler.UtilityTypeSHIELD:
		if utility.ShieldBlueprint == nil {
			return nil, fmt.Errorf("utility type is shield but shield blueprint is nil")
		}
		newUtilityShield := boiler.UtilityShield{
			UtilityID:          newUtility.ID,
			Hitpoints:          utility.ShieldBlueprint.Hitpoints,
			RechargeRate:       utility.ShieldBlueprint.RechargeRate,
			RechargeEnergyCost: utility.ShieldBlueprint.RechargeEnergyCost,
		}
		err = newUtilityShield.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	case boiler.UtilityTypeATTACKDRONE:
		if utility.AttackDroneBlueprint == nil {
			return nil, fmt.Errorf("utility type is attack drone but attack drone blueprint is nil")
		}
		newAttackDrone := boiler.UtilityAttackDrone{
			UtilityID:        newUtility.ID,
			Damage:           utility.AttackDroneBlueprint.Damage,
			RateOfFire:       utility.AttackDroneBlueprint.RateOfFire,
			Hitpoints:        utility.AttackDroneBlueprint.Hitpoints,
			LifespanSeconds:  utility.AttackDroneBlueprint.LifespanSeconds,
			DeployEnergyCost: utility.AttackDroneBlueprint.DeployEnergyCost,
		}
		err = newAttackDrone.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	case boiler.UtilityTypeREPAIRDRONE:
		if utility.RepairDroneBlueprint == nil {
			return nil, fmt.Errorf("utility type is repair drone but repair drone blueprint is nil")
		}
		newRepairDrone := boiler.UtilityRepairDrone{
			UtilityID:        newUtility.ID,
			RepairType:       utility.RepairDroneBlueprint.RepairType,
			RepairAmount:     utility.RepairDroneBlueprint.RepairAmount,
			DeployEnergyCost: utility.RepairDroneBlueprint.DeployEnergyCost,
			LifespanSeconds:  utility.RepairDroneBlueprint.LifespanSeconds,
		}
		err = newRepairDrone.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	case boiler.UtilityTypeANTIMISSILE:
		if utility.AntiMissileBlueprint == nil {
			return nil, fmt.Errorf("utility type is anti missile but anti missile blueprint is nil")
		}
		newAntiMissile := boiler.UtilityAntiMissile{
			UtilityID:      newUtility.ID,
			RateOfFire:     utility.AntiMissileBlueprint.RateOfFire,
			FireEnergyCost: utility.AntiMissileBlueprint.FireEnergyCost,
		}
		err = newAntiMissile.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	case boiler.UtilityTypeACCELERATOR:
		if utility.AcceleratorBlueprint == nil {
			return nil, fmt.Errorf("utility type accelerator but accelerator blueprint is nil")
		}
		newAccelerator := boiler.UtilityAccelerator{
			UtilityID:    newUtility.ID,
			EnergyCost:   utility.AcceleratorBlueprint.EnergyCost,
			BoostSeconds: utility.AcceleratorBlueprint.BoostSeconds,
			BoostAmount:  utility.AcceleratorBlueprint.BoostAmount,
		}
		err = newAccelerator.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	default:
		return nil, fmt.Errorf("invalid utility type")
	}

	return Utility(tx, newUtility.ID)
}

func Utility(tx boil.Executor, id string) (*server.Utility, error) {
	boilerUtility, err := boiler.Utilities(boiler.UtilityWhere.ID.EQ(id), qm.Load(boiler.UtilityRels.Blueprint)).One(tx)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	switch boilerUtility.Type {
	case boiler.UtilityTypeSHIELD:
		boilerShield, err := boiler.UtilityShields(boiler.UtilityShieldWhere.UtilityID.EQ(boilerUtility.ID)).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityShieldFromBoiler(boilerUtility, boilerUtility.R.Blueprint, boilerShield, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeATTACKDRONE:
		boilerAttackDrone, err := boiler.UtilityAttackDrones(boiler.UtilityAttackDroneWhere.UtilityID.EQ(boilerUtility.ID)).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityAttackDroneFromBoiler(boilerUtility, boilerUtility.R.Blueprint, boilerAttackDrone, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeREPAIRDRONE:
		boilerRepairDrone, err := boiler.UtilityRepairDrones(boiler.UtilityRepairDroneWhere.UtilityID.EQ(boilerUtility.ID)).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityRepairDroneFromBoiler(boilerUtility, boilerUtility.R.Blueprint, boilerRepairDrone, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeANTIMISSILE:
		boilerAntiMissile, err := boiler.UtilityAntiMissiles(boiler.UtilityAntiMissileWhere.UtilityID.EQ(boilerUtility.ID)).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityAntiMissileFromBoiler(boilerUtility, boilerUtility.R.Blueprint, boilerAntiMissile, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeACCELERATOR:
		boilerAccelerator, err := boiler.UtilityAccelerators(boiler.UtilityAcceleratorWhere.UtilityID.EQ(boilerUtility.ID)).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityAcceleratorFromBoiler(boilerUtility, boilerUtility.R.Blueprint, boilerAccelerator, boilerMechCollectionDetails), nil
	}

	return nil, fmt.Errorf("invalid utility type %s", boilerUtility.Type)
}

func Utilities(id ...string) ([]*server.Utility, error) {
	var utilities []*server.Utility
	boilerUtilities, err := boiler.Utilities(boiler.UtilityWhere.ID.IN(id), qm.Load(boiler.UtilityRels.Blueprint)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, util := range boilerUtilities {
		boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(util.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}

		switch util.Type {
		case boiler.UtilityTypeSHIELD:
			boilerShield, err := boiler.UtilityShields(boiler.UtilityShieldWhere.UtilityID.EQ(util.ID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityShieldFromBoiler(util, util.R.Blueprint, boilerShield, boilerMechCollectionDetails))
		case boiler.UtilityTypeATTACKDRONE:
			boilerAttackDrone, err := boiler.UtilityAttackDrones(boiler.UtilityAttackDroneWhere.UtilityID.EQ(util.ID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityAttackDroneFromBoiler(util, util.R.Blueprint, boilerAttackDrone, boilerMechCollectionDetails))
		case boiler.UtilityTypeREPAIRDRONE:
			boilerRepairDrone, err := boiler.UtilityRepairDrones(boiler.UtilityRepairDroneWhere.UtilityID.EQ(util.ID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityRepairDroneFromBoiler(util, util.R.Blueprint, boilerRepairDrone, boilerMechCollectionDetails))
		case boiler.UtilityTypeANTIMISSILE:
			boilerAntiMissile, err := boiler.UtilityAntiMissiles(boiler.UtilityAntiMissileWhere.UtilityID.EQ(util.ID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityAntiMissileFromBoiler(util, util.R.Blueprint, boilerAntiMissile, boilerMechCollectionDetails))
		case boiler.UtilityTypeACCELERATOR:
			boilerAccelerator, err := boiler.UtilityAccelerators(boiler.UtilityAcceleratorWhere.UtilityID.EQ(util.ID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityAcceleratorFromBoiler(util, util.R.Blueprint, boilerAccelerator, boilerMechCollectionDetails))
		}
	}
	return utilities, nil
}

// AttachUtilityToMech attaches a Utility to a mech
// If lockedToMech == true utility cannot be removed from mech ever (used for genesis and limited mechs)
func AttachUtilityToMech(trx *sql.Tx, ownerID, mechID, utilityID string, lockedToMech bool) error {
	tx := trx
	var err error
	if trx == nil {
		tx, err = gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Str("mech.ID", mechID).Str("utility ID", utilityID).Msg("failed to equip utility to mech, issue creating tx")
			return terror.Error(err, "Issue preventing equipping this utility to the war machine, try again or contact support.")
		}
		defer tx.Rollback()
	}

	// TODO: possible optimize this, 6 queries to attach a part seems like a lot?
	mechCI, err := CollectionItemFromItemID(tx, mechID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to get mech collection item")
		return terror.Error(err)
	}
	utilityCI, err := CollectionItemFromItemID(tx, utilityID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("utilityID", utilityID).Msg("failed to get utility collection item")
		return terror.Error(err)
	}

	if mechCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("mechCI.OwnerID", mechCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the war machine to equip utilitys to it.")
	}
	if utilityCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("utilityCI.OwnerID", utilityCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the utility to equip it to a war machine.")
	}

	// get mech
	mech, err := boiler.Mechs(
		boiler.MechWhere.ID.EQ(mechID),
		qm.Load(boiler.MechRels.ChassisMechUtilities),
	).One(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to find mech")
		return terror.Error(err)
	}

	// get Utility
	utility, err := boiler.FindUtility(tx, utilityID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("utilityID", utilityID).Msg("failed to find Utility")
		return terror.Error(err)
	}

	// check utility isn't already equipped to another war machine
	exists, err := boiler.MechUtilities(boiler.MechUtilityWhere.UtilityID.EQ(null.StringFrom(utilityID))).Exists(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Str("utilityID", utilityID).Msg("failed to check if a mech and utility join already exists")
		return terror.Error(err)
	}
	if exists {
		err := fmt.Errorf("utility already equipped to a warmachine")
		gamelog.L.Error().Err(err).Str("utilityID", utilityID).Msg(err.Error())
		return terror.Error(err, "This utility is already equipped to another war machine, try again or contact support.")
	}

	// get next available slot
	availableSlot, err := boiler.MechUtilities(
		boiler.MechUtilityWhere.ChassisID.EQ(mech.ID),
		boiler.MechUtilityWhere.UtilityID.IsNull(),
		qm.OrderBy(fmt.Sprintf("%s ASC", boiler.MechWeaponColumns.SlotNumber)),
	).One(tx)
	if errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("mechID", mech.ID).Msg("no available slots on mech to insert utility")
		return terror.Error(err, "There are no more slots on this mech to equip this utility.")
	} else if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mech.ID).Msg("failed to check for available utility slots on mech")
		return terror.Error(err)
	}

	utility.EquippedOn = null.StringFrom(mech.ID)
	utility.LockedToMech = lockedToMech
	_, err = utility.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("utility", utility).Msg("failed to update utility")
		return terror.Error(err, "Issue preventing equipping this utility to the war machine, try again or contact support.")
	}

	availableSlot.UtilityID = null.StringFrom(utility.ID)
	_, err = availableSlot.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("utility", utility).Msg(" failed to equip utility to war machine")
		return terror.Error(err, "Issue preventing equipping this utility to the war machine, try again or contact support.")
	}

	if trx == nil {
		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("failed to commit transaction - AttachUtilityToMech")
			return terror.Error(err)
		}
	}

	return nil
}
