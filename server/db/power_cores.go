package db

import (
	"database/sql"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/volatiletech/null/v8"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func getDefaultPowerCoreQueryMods() []qm.QueryMod {
	return []qm.QueryMod{
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypePowerCore),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.PowerCores,
			qm.Rels(boiler.TableNames.PowerCores, boiler.PowerCoreColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		// join power core blueprint
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintPowerCores,
			qm.Rels(boiler.TableNames.BlueprintPowerCores, boiler.BlueprintPowerCoreColumns.ID),
			qm.Rels(boiler.TableNames.PowerCores, boiler.PowerCoreColumns.BlueprintID),
		)),
	}
}

type PowerCoreListOpts struct {
	Search                 string
	Filter                 *ListFilterRequest
	SortBy                 string
	SortDir                SortByDir
	PageSize               int
	Page                   int
	OwnerID                string
	DisplayXsynLocked      bool
	DisplayHidden          bool
	ExcludeMarketLocked    bool
	IncludeMarketListed    bool
	ExcludeIDs             []string                  `json:"exclude_ids"`
	FilterRarities         []string                  `json:"rarities"`
	FilterSizes            []string                  `json:"sizes"`
	FilterEquippedStatuses []string                  `json:"equipped_statuses"`
	FilterStatCapacity     *PowerCoreStatFilterRange `json:"stat_capacity"`
	FilterStatMaxDrawRate  *PowerCoreStatFilterRange `json:"stat_max_draw_rate"`
	FilterStatRechargeRate *PowerCoreStatFilterRange `json:"stat_recharge_rate"`
	FilterStatArmour       *PowerCoreStatFilterRange `json:"stat_armour"`
	FilterStatMaxHitpoints *PowerCoreStatFilterRange `json:"stat_max_hitpoints"`
}

type PowerCoreStatFilterRange struct {
	Min null.Int `json:"min"`
	Max null.Int `json:"max"`
}

func GeneratePowerCoreStatFilterQueryMods(column string, filter *PowerCoreStatFilterRange) []qm.QueryMod {
	output := []qm.QueryMod{}
	if filter == nil {
		return output
	}
	if filter.Min.Valid {
		output = append(output, qm.Where(qm.Rels(boiler.TableNames.BlueprintPowerCores, column)+" >= ?", filter.Min))
	}
	if filter.Max.Valid {
		output = append(output, qm.Where(qm.Rels(boiler.TableNames.BlueprintPowerCores, column)+" <= ?", filter.Max))
	}
	return output
}

func PowerCoreList(opts *PowerCoreListOpts) (int64, []*PlayerAsset, error) {
	queryMods := getDefaultPowerCoreQueryMods()

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

	// Filters
	if opts.Filter != nil {
		// if we have filter
		for i, f := range opts.Filter.Items {
			// validate it is the right table and valid column
			if f.Table == boiler.TableNames.PowerCores && IsMechColumn(f.Column) {
				queryMods = append(queryMods, GenerateListFilterQueryMod(*f, i+1, opts.Filter.LinkOperator))
			}
		}
	}

	if len(opts.ExcludeIDs) > 0 {
		queryMods = append(queryMods, boiler.PowerCoreWhere.ID.NIN(opts.ExcludeIDs))
	}

	if len(opts.FilterRarities) > 0 {
		queryMods = append(queryMods, boiler.BlueprintPowerCoreWhere.Tier.IN(opts.FilterRarities))
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
			queryMods = append(queryMods, boiler.PowerCoreWhere.EquippedOn.IsNotNull())
		} else if showUnequipped && !showEquipped {
			queryMods = append(queryMods, boiler.PowerCoreWhere.EquippedOn.IsNull())
		}
	}

	if len(opts.FilterSizes) > 0 {
		queryMods = append(queryMods, boiler.BlueprintPowerCoreWhere.Size.IN(opts.FilterSizes))
	}

	// Filter - Weapon Stats
	if opts.FilterStatCapacity != nil {
		queryMods = append(queryMods, GeneratePowerCoreStatFilterQueryMods(boiler.BlueprintPowerCoreColumns.Capacity, opts.FilterStatCapacity)...)
	}
	if opts.FilterStatMaxDrawRate != nil {
		queryMods = append(queryMods, GeneratePowerCoreStatFilterQueryMods(boiler.BlueprintPowerCoreColumns.MaxDrawRate, opts.FilterStatMaxDrawRate)...)
	}
	if opts.FilterStatRechargeRate != nil {
		queryMods = append(queryMods, GeneratePowerCoreStatFilterQueryMods(boiler.BlueprintPowerCoreColumns.RechargeRate, opts.FilterStatRechargeRate)...)
	}
	if opts.FilterStatArmour != nil {
		queryMods = append(queryMods, GeneratePowerCoreStatFilterQueryMods(boiler.BlueprintPowerCoreColumns.Armour, opts.FilterStatArmour)...)
	}
	if opts.FilterStatMaxHitpoints != nil {
		queryMods = append(queryMods, GeneratePowerCoreStatFilterQueryMods(boiler.BlueprintPowerCoreColumns.MaxHitpoints, opts.FilterStatMaxHitpoints)...)
	}

	// Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"((to_tsvector('english', %s) @@ to_tsquery(?))",
					qm.Rels(boiler.TableNames.BlueprintPowerCores, boiler.BlueprintPowerCoreColumns.Label),
				),
					xSearch,
				))
		}
	}
	total, err := boiler.CollectionItems(
		queryMods...,
	).Count(gamedb.StdConn)
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
			qm.Rels(boiler.TableNames.PowerCores, boiler.PowerCoreColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintPowerCores, boiler.BlueprintPowerCoreColumns.Label),
		),
		qm.From(boiler.TableNames.CollectionItems),
	)

	if opts.SortBy != "" && opts.SortDir.IsValid() {
		if opts.SortBy == "alphabetical" {
			queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.BlueprintPowerCores, boiler.BlueprintPowerCoreColumns.Label), opts.SortDir)))
		} else if opts.SortBy == "rarity" {
			queryMods = append(queryMods, GenerateTierSort(qm.Rels(boiler.TableNames.BlueprintPowerCores, boiler.BlueprintPowerCoreColumns.Tier), opts.SortDir))
		}
	} else {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s ASC", qm.Rels(boiler.TableNames.BlueprintPowerCores, boiler.BlueprintPowerCoreColumns.Label))))
	}

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var powerCores []*PlayerAsset
	for rows.Next() {
		pc := &PlayerAsset{}

		scanArgs := []interface{}{
			&pc.CollectionSlug,
			&pc.Hash,
			&pc.TokenID,
			&pc.ItemType,
			&pc.ItemID,
			&pc.Tier,
			&pc.OwnerID,
			&pc.MarketLocked,
			&pc.XsynLocked,
			&pc.LockedToMarketplace,
			&pc.AssetHidden,
			&pc.ID,
			&pc.Label,
			&pc.Name,
			&pc.UpdatedAt,
			&pc.CreatedAt,
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, powerCores, err
		}
		powerCores = append(powerCores, pc)
	}

	return total, powerCores, nil
}

func InsertNewPowerCore(tx boil.Executor, ownerID uuid.UUID, ec *server.BlueprintPowerCore) (*server.PowerCore, error) {
	newPowerCore := boiler.PowerCore{
		BlueprintID:           null.StringFrom(ec.ID),
		GenesisTokenID:        ec.GenesisTokenID,
		LimitedReleaseTokenID: ec.LimitedReleaseTokenID,
	}

	err := newPowerCore.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		ec.Collection,
		boiler.ItemTypePowerCore,
		newPowerCore.ID,
		ec.Tier,
		ownerID.String(),
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return PowerCore(tx, newPowerCore.ID)
}

func PowerCore(tx boil.Executor, id string) (*server.PowerCore, error) {
	boilerMech, err := boiler.PowerCores(
		boiler.PowerCoreWhere.ID.EQ(id),
		qm.Load(boiler.PowerCoreRels.Blueprint),
	).One(tx)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	return server.PowerCoreFromBoiler(boilerMech, boilerMechCollectionDetails), nil
}

func PowerCores(id ...string) ([]*server.PowerCore, error) {
	var powerCores []*server.PowerCore
	boilerPowerCores, err := boiler.PowerCores(
		boiler.PowerCoreWhere.ID.IN(id),
		qm.Load(boiler.PowerCoreRels.Blueprint),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}
	for _, pc := range boilerPowerCores {
		boilerPowerCoreCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(pc.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		powerCores = append(powerCores, server.PowerCoreFromBoiler(pc, boilerPowerCoreCollectionDetails))
	}

	return powerCores, nil
}

// AttachPowerCoreToMech attaches a power core to a mech
func AttachPowerCoreToMech(trx *sql.Tx, ownerID, mechID, powerCoreID string) error {
	// TODO: possible optimize this, 6 queries to attach a part seems like a lot?
	// check owner
	tx := trx
	var err error
	if trx == nil {
		tx, err = gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Str("mech.ID", mechID).Str("powercore ID", powerCoreID).Msg("failed to equip powercore to mech, issue creating tx")
			return terror.Error(err, "Issue preventing equipping this powercore to the war machine, try again or contact support.")
		}
		defer tx.Rollback()
	}

	mechCI, err := CollectionItemFromItemID(tx, mechID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to mech collection item")
		return terror.Error(err)
	}
	pcCI, err := CollectionItemFromItemID(tx, powerCoreID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("powerCoreID", powerCoreID).Msg("failed to power core collection item")
		return terror.Error(err)
	}

	if mechCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("mechCI.OwnerID", mechCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the war machine to equip power cores to it.")

	}
	if pcCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("pcCI.OwnerID", pcCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the power core to equip it to a war machine.")
	}

	// get mech
	mech, err := boiler.Mechs(
		boiler.MechWhere.ID.EQ(mechID),
		qm.Load(boiler.MechRels.Blueprint),
	).One(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to find mech")
		return terror.Error(err)
	}

	// get power core
	powerCore, err := boiler.PowerCores(
		boiler.PowerCoreWhere.ID.EQ(powerCoreID),
		qm.Load(boiler.PowerCoreRels.Blueprint),
	).One(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Str("powerCoreID", powerCoreID).Msg("failed to find power core")
		return terror.Error(err)
	}

	// wrong size
	if mech.R.Blueprint.PowerCoreSize != powerCore.R.Blueprint.Size {
		err := fmt.Errorf("powercore size mismatch")
		gamelog.L.Error().Err(err).Str("mech.PowerCoreSize", mech.R.Blueprint.PowerCoreSize).Str("powerCore.Size", powerCore.R.Blueprint.Size).Msg("this powercore doesn't fit")
		return terror.Error(err, "This power core doesn't fit this war machine.")
	}

	// error out, already has a power core
	if mech.PowerCoreID.Valid && mech.PowerCoreID.String != "" {
		err := fmt.Errorf("warmachine already has a power core")
		// also check powerCore.EquippedOn on, if that doesn't match, update it, so it does.
		if !powerCore.EquippedOn.Valid {
			powerCore.EquippedOn = null.StringFrom(mech.ID)
			_, err = powerCore.Update(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Str("mech.ID", mech.ID).Str("powerCore.ID", powerCore.ID).Msg("failed to update power core equipped on")
				return terror.Error(err, "War machine already has a power core.")
			}
		}
		gamelog.L.Error().Err(err).Str("mech.PowerCoreID.String", mech.PowerCoreID.String).Str("new powerCore.ID", powerCore.ID).Msg(err.Error())
		return terror.Error(err, "War machine already has a power core.")
	}

	// lets join
	mech.PowerCoreID = null.StringFrom(powerCore.ID)
	powerCore.EquippedOn = null.StringFrom(mech.ID)

	_, err = mech.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech.PowerCoreID.String", mech.PowerCoreID.String).Str("new powerCore.ID", powerCore.ID).Msg("failed to equip power core to mech, issue mech update")
		return terror.Error(err, "Issue preventing equipping this power core to the war machine, try again or contact support.")
	}
	_, err = powerCore.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech.PowerCoreID.String", mech.PowerCoreID.String).Str("new powerCore.ID", powerCore.ID).Msg("failed to equip power core to mech, issue power core update")
		return terror.Error(err, "Issue preventing equipping this power core to the war machine, try again or contact support.")
	}

	if trx == nil {
		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Err(err).Str("mech.PowerCoreID.String", mech.PowerCoreID.String).Str("new powerCore.ID", powerCore.ID).Msg("failed to equip power core to mech, issue committing tx")
			return terror.Error(err, "Issue preventing equipping this power core to the war machine, try again or contact support.")
		}
	}

	return nil
}
