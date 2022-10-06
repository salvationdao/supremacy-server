package db

import (
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func getDefaultWeaponQueryMods() []qm.QueryMod {
	return []qm.QueryMod{
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeWeapon),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Weapons,
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		// join weapon blueprint
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintWeapons,
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.BlueprintID),
		)),
		// join weapon skin
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.WeaponSkin,
			qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.ID),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.EquippedWeaponSkinID),
		)),
		// join weapon skin blueprint
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintWeaponSkin,
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
			qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.BlueprintID),
		)),
		// join weapon skin matrix
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s AND %s = %s",
			boiler.TableNames.WeaponModelSkinCompatibilities,
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.WeaponModelID),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID),
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.BlueprintWeaponSkinID),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
		)),
		// join weapon skin collection item
		qm.LeftOuterJoin(fmt.Sprintf("%s AS _wsci ON _wsci.%s = %s",
			boiler.TableNames.CollectionItems,
			boiler.CollectionItemColumns.ItemID,
			boiler.WeaponSkinTableColumns.ID,
		)),
	}
}

func WeaponEquippedOnDetails(trx boil.Executor, equippedOnID string) (*server.EquippedOnDetails, error) {
	tx := trx
	if trx == nil {
		tx = gamedb.StdConn
	}

	eid := &server.EquippedOnDetails{}

	err := boiler.NewQuery(
		qm.Select(
			boiler.CollectionItemColumns.ItemID,
			boiler.CollectionItemColumns.Hash,
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label),
		),
		qm.From(boiler.TableNames.CollectionItems),
		qm.InnerJoin(fmt.Sprintf(
			"%s on %s = %s",
			boiler.TableNames.Weapons,
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		qm.InnerJoin(fmt.Sprintf(
			"%s on %s = %s",
			boiler.TableNames.BlueprintWeapons,
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.BlueprintID),
		)),
		qm.Where(fmt.Sprintf("%s = ?", boiler.CollectionItemColumns.ItemID), equippedOnID),
	).QueryRow(tx).Scan(
		&eid.ID,
		&eid.Hash,
		&eid.Label,
	)
	if err != nil {
		return nil, err
	}

	return eid, nil
}

func InsertNewWeapon(tx *sql.Tx, ownerID uuid.UUID, weapon *server.BlueprintWeapon, weaponSkin *server.BlueprintWeaponSkin) (*server.Weapon, *server.WeaponSkin, error) {
	L := gamelog.L.With().Str("func", "InsertNewWeapon").Interface("weaponBlueprint", weapon).Interface("weaponSkin", weaponSkin).Str("ownerID", ownerID.String()).Logger()

	// first insert the new weapon skin
	wpnSkin, err := InsertNewWeaponSkin(tx, ownerID, weaponSkin, &weapon.ID)
	if err != nil {
		L.Error().Err(err).Msg("failed to insert new weapon skin")
		return nil, nil, err
	}

	newWeapon := boiler.Weapon{
		BlueprintID:           weapon.ID,
		GenesisTokenID:        weapon.GenesisTokenID,
		LimitedReleaseTokenID: weapon.LimitedReleaseTokenID,
		EquippedWeaponSkinID:  wpnSkin.ID,
		LockedToMech:          weapon.ID == "c1c78867-9de7-43d3-97e9-91381800f38e" || weapon.ID == "41099781-8586-4783-9d1c-b515a386fe9f" || weapon.ID == "e9fc2417-6a5b-489d-b82e-42942535af90",
	}

	err = newWeapon.Insert(tx, boil.Infer())
	if err != nil {
		L.Error().Err(err).Msg("failed to insert new weapon")
		return nil, nil, err
	}

	_, err = InsertNewCollectionItem(tx,
		weapon.Collection,
		boiler.ItemTypeWeapon,
		newWeapon.ID,
		"",
		ownerID.String(),
	)
	if err != nil {
		L.Error().Err(err).Msg("failed to insert new weapon collection item")
		return nil, nil, err
	}

	// update skin to say equipped to this mech
	updated, err := boiler.WeaponSkins(
		boiler.WeaponSkinWhere.ID.EQ(wpnSkin.ID),
	).UpdateAll(tx, boiler.M{
		boiler.WeaponSkinColumns.EquippedOn: newWeapon.ID,
	})
	if err != nil {
		L.Error().Err(err).Msg("failed to update weapon skin")
		return nil, nil, err
	}
	if updated != 1 {
		err = fmt.Errorf("updated %d, expected 1", updated)
		L.Error().Err(err).Msg("failed to update weapon skin")
		return nil, nil, err
	}

	wpnSkin.EquippedOn = null.StringFrom(newWeapon.ID)

	wpn, err := Weapon(tx, newWeapon.ID)
	if err != nil {
		return nil, nil, terror.Error(err)
	}

	return wpn, wpnSkin, nil
}

func Weapon(tx boil.Executor, id string) (*server.Weapon, error) {
	boilerWeapon, err := boiler.Weapons(
		boiler.WeaponWhere.ID.EQ(id),
		qm.Load(boiler.WeaponRels.Blueprint),
	).One(tx)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	weaponSkin, err := WeaponSkin(tx, boilerWeapon.EquippedWeaponSkinID, &boilerWeapon.R.Blueprint.ID)
	if err != nil {
		return nil, err
	}

	itemSale, err := boiler.ItemSales(
		boiler.ItemSaleWhere.CollectionItemID.EQ(boilerMechCollectionDetails.ID),
		boiler.ItemSaleWhere.SoldAt.IsNull(),
		boiler.ItemSaleWhere.DeletedAt.IsNull(),
		boiler.ItemSaleWhere.EndAt.GT(time.Now()),
	).One(tx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	itemSaleID := null.String{}
	if itemSale != nil {
		itemSaleID = null.StringFrom(itemSale.ID)
	}
	return server.WeaponFromBoiler(boilerWeapon, boilerMechCollectionDetails, weaponSkin, itemSaleID), nil
}

// AttachWeaponToMech attaches a Weapon to a mech
func AttachWeaponToMech(trx *sql.Tx, ownerID, mechID, weaponID string) error {
	tx := trx
	var err error
	if trx == nil {
		tx, err = gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Str("mech.ID", mechID).Str("weapon ID", weaponID).Msg("failed to equip weapon to mech, issue creating tx")
			return terror.Error(err, "Issue preventing equipping this weapon to the war machine, try again or contact support.")
		}
		defer tx.Rollback()
	}

	mechCI, err := CollectionItemFromItemID(tx, mechID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to get mech collection item")
		return terror.Error(err)
	}
	weaponCI, err := CollectionItemFromItemID(tx, weaponID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg("failed to get weapon collection item")
		return terror.Error(err)
	}

	if mechCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("mechCI.OwnerID", mechCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the war machine to equip weapons to it.")
	}
	if weaponCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("weaponCI.OwnerID", weaponCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the weapon to equip it to a war machine.")
	}

	// get mech
	mech, err := boiler.Mechs(
		boiler.MechWhere.ID.EQ(mechID),
		qm.Load(boiler.MechRels.ChassisMechWeapons),
		qm.Load(boiler.MechRels.Blueprint),
	).One(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to find mech")
		return terror.Error(err)
	}

	// get Weapon
	weapon, err := boiler.FindWeapon(tx, weaponID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg("failed to find Weapon")
		return terror.Error(err)
	}

	// check weapon isn't already equipped to another war machine
	exists, err := boiler.MechWeapons(boiler.MechWeaponWhere.WeaponID.EQ(null.StringFrom(weaponID))).Exists(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg("failed to check if a mech and weapon join already exists")
		return terror.Error(err)
	}
	if exists {
		err := fmt.Errorf("weapon already equipped to a warmachine")
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg(err.Error())
		return terror.Error(err, "This weapon is already equipped to another war machine, try again or contact support.")
	}

	// get next available slot
	availableSlot, err := boiler.MechWeapons(
		boiler.MechWeaponWhere.ChassisID.EQ(mech.ID),
		boiler.MechWeaponWhere.WeaponID.IsNull(),
		qm.OrderBy(fmt.Sprintf("%s ASC", boiler.MechWeaponColumns.SlotNumber)),
	).One(tx)
	if errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Str("mechID", mech.ID).Msg("no available slots on mech to insert weapon")
		return terror.Error(err, "There are no more slots on this mech to equip this weapon.")
	} else if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mech.ID).Msg("failed to check for available weapon slots on mech")
		return terror.Error(err)
	}

	weapon.EquippedOn = null.StringFrom(mech.ID)
	_, err = weapon.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("weapon", weapon).Msg("failed to update weapon")
		return terror.Error(err, "Issue preventing equipping this weapon to the war machine, try again or contact support.")
	}

	availableSlot.WeaponID = null.StringFrom(weapon.ID)
	_, err = availableSlot.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("weapon", weapon).Msg("failed to update mech_weapon entry")
		return terror.Error(err, "Issue preventing equipping this weapon to the war machine, try again or contact support.")
	}

	if trx == nil {
		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("failed to commit transaction - AttachWeaponToMech")
			return terror.Error(err)
		}
	}

	return nil
}

// CheckWeaponAttached checks whether weapon item is already equipped.
func CheckWeaponAttached(weaponID string) (bool, error) {
	exists, err := boiler.Weapons(
		qm.LeftOuterJoin(fmt.Sprintf(
			`%s on %s = %s`,
			boiler.TableNames.MechWeapons,
			qm.Rels(boiler.TableNames.MechWeapons, boiler.MechWeaponColumns.WeaponID),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
		)),
		boiler.WeaponWhere.ID.EQ(weaponID),
		qm.Expr(
			boiler.WeaponWhere.EquippedOn.IsNotNull(),
			qm.Or(fmt.Sprintf(`%s = ?`, qm.Rels(boiler.TableNames.MechWeapons, boiler.MechWeaponColumns.WeaponID)), weaponID),
		),
	).Exists(gamedb.StdConn)
	if err != nil {
		return false, terror.Error(err)
	}
	return exists, nil
}

func IsWeaponColumn(col string) bool {
	switch col {
	case
		boiler.WeaponColumns.ID,
		boiler.WeaponColumns.SlugDontUse,
		boiler.WeaponColumns.DamageDontUse,
		boiler.WeaponColumns.DeletedAt,
		boiler.WeaponColumns.UpdatedAt,
		boiler.WeaponColumns.CreatedAt,
		boiler.WeaponColumns.BlueprintID,
		boiler.WeaponColumns.EquippedOn,
		boiler.WeaponColumns.DefaultDamageTypeDontUse,
		boiler.WeaponColumns.GenesisTokenID,
		boiler.WeaponColumns.LimitedReleaseTokenID,
		boiler.WeaponColumns.DamageFalloffDontUse,
		boiler.WeaponColumns.DamageFalloffRateDontUse,
		boiler.WeaponColumns.RadiusDontUse,
		boiler.WeaponColumns.RadiusDamageFalloffDontUse,
		boiler.WeaponColumns.SpreadDontUse,
		boiler.WeaponColumns.RateOfFireDontUse,
		boiler.WeaponColumns.ProjectileSpeedDontUse,
		boiler.WeaponColumns.EnergyCostDontUse,
		boiler.WeaponColumns.IsMeleeDontUse,
		boiler.WeaponColumns.MaxAmmoDontUse,
		boiler.WeaponColumns.LockedToMech,
		boiler.WeaponColumns.EquippedWeaponSkinID,
		boiler.WeaponColumns.BlueprintIDOld:
		return true
	default:
		return false
	}
}

type WeaponListOpts struct {
	Search                        string
	Filter                        *ListFilterRequest
	Sort                          *ListSortRequest
	SortBy                        string
	SortDir                       SortByDir
	PageSize                      int
	Page                          int
	OwnerID                       string
	DisplayXsynMechs              bool
	DisplayGenesisAndLimited      bool
	DisplayHidden                 bool
	ExcludeMarketLocked           bool
	IncludeMarketListed           bool
	ExcludeMechLocked             bool
	ExcludeIDs                    []string
	FilterRarities                []string               `json:"rarities"`
	FilterWeaponTypes             []string               `json:"weapon_types"`
	FilterEquippedStatuses        []string               `json:"equipped_statuses"`
	FilterStatAmmo                *WeaponStatFilterRange `json:"stat_ammo"`
	FilterStatDamage              *WeaponStatFilterRange `json:"stat_damage"`
	FilterStatDamageFalloff       *WeaponStatFilterRange `json:"stat_damage_falloff"`
	FilterStatDamageFalloffRate   *WeaponStatFilterRange `json:"stat_damage_falloff_rate"`
	FilterStatRadius              *WeaponStatFilterRange `json:"stat_radius"`
	FilterStatRadiusDamageFalloff *WeaponStatFilterRange `json:"stat_radius_damage_falloff"`
	FilterStatRateOfFire          *WeaponStatFilterRange `json:"stat_rate_of_fire"`
	FilterStatEnergyCosts         *WeaponStatFilterRange `json:"stat_energy_cost"`
	FilterStatProjectileSpeed     *WeaponStatFilterRange `json:"stat_projectile_speed"`
	FilterStatSpread              *WeaponStatFilterRange `json:"stat_spread"`
}

type WeaponStatFilterRange struct {
	Min null.Int `json:"min"`
	Max null.Int `json:"max"`
}

func GenerateWeaponStatFilterQueryMods(column string, filter *WeaponStatFilterRange) []qm.QueryMod {
	output := []qm.QueryMod{}
	if filter == nil {
		return output
	}
	if filter.Min.Valid {
		output = append(output, qm.Where(qm.Rels(boiler.TableNames.BlueprintWeapons, column)+" >= ?", filter.Min))
	}
	if filter.Max.Valid {
		output = append(output, qm.Where(qm.Rels(boiler.TableNames.BlueprintWeapons, column)+" <= ?", filter.Max))
	}
	return output
}

func WeaponList(opts *WeaponListOpts) (int64, []*PlayerAsset, error) {
	queryMods := getDefaultWeaponQueryMods()

	if opts.OwnerID != "" {
		queryMods = append(queryMods, boiler.CollectionItemWhere.OwnerID.EQ(opts.OwnerID))
	}
	if !opts.DisplayXsynMechs {
		queryMods = append(queryMods, boiler.CollectionItemWhere.XsynLocked.EQ(false))
	}
	if !opts.IncludeMarketListed {
		queryMods = append(queryMods, boiler.CollectionItemWhere.LockedToMarketplace.EQ(false))
	}
	if opts.ExcludeMarketLocked {
		queryMods = append(queryMods, boiler.CollectionItemWhere.MarketLocked.EQ(false))
	}
	if !opts.DisplayGenesisAndLimited {
		queryMods = append(queryMods, boiler.WeaponWhere.GenesisTokenID.IsNull())
		queryMods = append(queryMods, boiler.WeaponWhere.LimitedReleaseTokenID.IsNull())
	}
	if !opts.DisplayHidden {
		queryMods = append(queryMods, boiler.CollectionItemWhere.AssetHidden.IsNull())
	}
	if opts.ExcludeMechLocked {
		queryMods = append(queryMods,
			boiler.WeaponWhere.LockedToMech.EQ(false),
		)
	}

	// Filters
	if opts.Filter != nil {
		// if we have filter
		for i, f := range opts.Filter.Items {
			// validate it is the right table and valid column
			if f.Table == boiler.TableNames.Weapons && IsWeaponColumn(f.Column) {
				queryMods = append(queryMods, GenerateListFilterQueryMod(*f, i+1, opts.Filter.LinkOperator))
			}
		}
	}

	if len(opts.ExcludeIDs) > 0 {
		queryMods = append(queryMods, boiler.WeaponWhere.ID.NIN(opts.ExcludeIDs))
	}

	if len(opts.FilterRarities) > 0 {
		queryMods = append(queryMods, boiler.BlueprintWeaponSkinWhere.Tier.IN(opts.FilterRarities))
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
			queryMods = append(queryMods, boiler.WeaponWhere.EquippedOn.IsNotNull())
		} else if showUnequipped && !showEquipped {
			queryMods = append(queryMods, boiler.WeaponWhere.EquippedOn.IsNull())
		}
	}

	if len(opts.FilterWeaponTypes) > 0 {
		queryMods = append(queryMods, boiler.BlueprintWeaponWhere.WeaponType.IN(opts.FilterWeaponTypes))
	}

	// Filter - Weapon Stats
	if opts.FilterStatAmmo != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.MaxAmmo, opts.FilterStatAmmo)...)
	}
	if opts.FilterStatDamage != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.Damage, opts.FilterStatDamage)...)
	}
	if opts.FilterStatDamageFalloff != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.DamageFalloff, opts.FilterStatDamageFalloff)...)
	}
	if opts.FilterStatDamageFalloffRate != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.DamageFalloffRate, opts.FilterStatDamageFalloffRate)...)
	}
	if opts.FilterStatRadius != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.Radius, opts.FilterStatRadius)...)
	}
	if opts.FilterStatRadiusDamageFalloff != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.RadiusDamageFalloff, opts.FilterStatRadiusDamageFalloff)...)
	}
	if opts.FilterStatRateOfFire != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.RateOfFire, opts.FilterStatRateOfFire)...)
	}
	if opts.FilterStatEnergyCosts != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.PowerCost, opts.FilterStatEnergyCosts)...)
	}
	if opts.FilterStatProjectileSpeed != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.ProjectileSpeed, opts.FilterStatProjectileSpeed)...)
	}
	if opts.FilterStatSpread != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.Spread, opts.FilterStatSpread)...)
	}

	// Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"((to_tsvector('english', %s) @@ to_tsquery(?) OR (to_tsvector('english', %s::text) @@ to_tsquery(?)) OR (to_tsvector('english', %s::text) @@ to_tsquery(?)) ))",
					qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label),
					qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.WeaponType),
					qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponColumns.Label),
				),
					xSearch,
					xSearch,
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
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.MarketLocked),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.XsynLocked),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.LockedToMarketplace),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.AssetHidden),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label),
			fmt.Sprintf(
				`(
					SELECT _i.%s
					FROM %s _i
					WHERE _i.%s = %s
						AND _i.%s IS NULL
						AND _i.%s IS NULL
						AND _i.%s > NOW()
				) AS item_sale_id`,
				boiler.ItemSaleColumns.ID,
				boiler.TableNames.ItemSales,
				boiler.ItemSaleColumns.CollectionItemID,
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ID),
				boiler.ItemSaleColumns.SoldAt,
				boiler.ItemSaleColumns.DeletedAt,
				boiler.ItemSaleColumns.EndAt,
			),
		),
		qm.From(boiler.TableNames.CollectionItems),
	)

	if opts.SortBy != "" && opts.SortDir.IsValid() {
		if opts.SortBy == "alphabetical" {
			queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label), opts.SortDir)))
		} else if opts.SortBy == "rarity" {
			queryMods = append(queryMods, GenerateTierSort(qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.Tier), opts.SortDir))
		}
	} else {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s ASC", qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label))))
	}

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var weapons []*PlayerAsset
	for rows.Next() {
		w := &PlayerAsset{
			CollectionItem: &server.CollectionItem{},
		}

		scanArgs := []interface{}{
			&w.CollectionItem.CollectionSlug,
			&w.CollectionItem.Hash,
			&w.CollectionItem.TokenID,
			&w.CollectionItem.OwnerID,
			&w.CollectionItem.Tier,
			&w.CollectionItem.ItemType,
			&w.CollectionItem.ItemID,
			&w.CollectionItem.MarketLocked,
			&w.CollectionItem.XsynLocked,
			&w.CollectionItem.LockedToMarketplace,
			&w.CollectionItem.AssetHidden,
			&w.ID,
			&w.Label,
			&w.ItemSaleID,
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, weapons, err
		}
		weapons = append(weapons, w)
	}

	return total, weapons, nil
}

func WeaponListDetailed(opts *WeaponListOpts) (int64, []*server.Weapon, error) {
	queryMods := getDefaultWeaponQueryMods()

	if opts.OwnerID != "" {
		queryMods = append(queryMods, boiler.CollectionItemWhere.OwnerID.EQ(opts.OwnerID))
	}
	if !opts.DisplayXsynMechs {
		queryMods = append(queryMods, boiler.CollectionItemWhere.XsynLocked.EQ(false))
	}
	if !opts.IncludeMarketListed {
		queryMods = append(queryMods, boiler.CollectionItemWhere.LockedToMarketplace.EQ(false))
	}
	if opts.ExcludeMarketLocked {
		queryMods = append(queryMods, boiler.CollectionItemWhere.MarketLocked.EQ(false))
	}
	if !opts.DisplayGenesisAndLimited {
		queryMods = append(queryMods, boiler.WeaponWhere.GenesisTokenID.IsNull())
		queryMods = append(queryMods, boiler.WeaponWhere.LimitedReleaseTokenID.IsNull())
	}
	if !opts.DisplayHidden {
		queryMods = append(queryMods, boiler.CollectionItemWhere.AssetHidden.IsNull())
	}
	if opts.ExcludeMechLocked {
		queryMods = append(queryMods,
			boiler.WeaponWhere.LockedToMech.EQ(false),
		)
	}

	// Filters
	if opts.Filter != nil {
		// if we have filter
		for i, f := range opts.Filter.Items {
			// validate it is the right table and valid column
			if f.Table == boiler.TableNames.Weapons && IsWeaponColumn(f.Column) {
				queryMods = append(queryMods, GenerateListFilterQueryMod(*f, i+1, opts.Filter.LinkOperator))
			}
		}
	}

	if len(opts.ExcludeIDs) > 0 {
		queryMods = append(queryMods, boiler.WeaponWhere.ID.NIN(opts.ExcludeIDs))
	}

	if len(opts.FilterRarities) > 0 {
		queryMods = append(queryMods, boiler.BlueprintWeaponSkinWhere.Tier.IN(opts.FilterRarities))
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
			queryMods = append(queryMods, boiler.WeaponWhere.EquippedOn.IsNotNull())
		} else if showUnequipped && !showEquipped {
			queryMods = append(queryMods, boiler.WeaponWhere.EquippedOn.IsNull())
		}
	}

	if len(opts.FilterWeaponTypes) > 0 {
		queryMods = append(queryMods, boiler.BlueprintWeaponWhere.WeaponType.IN(opts.FilterWeaponTypes))
	}

	// Filter - Weapon Stats
	if opts.FilterStatAmmo != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.MaxAmmo, opts.FilterStatAmmo)...)
	}
	if opts.FilterStatDamage != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.Damage, opts.FilterStatDamage)...)
	}
	if opts.FilterStatDamageFalloff != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.DamageFalloff, opts.FilterStatDamageFalloff)...)
	}
	if opts.FilterStatDamageFalloffRate != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.DamageFalloffRate, opts.FilterStatDamageFalloffRate)...)
	}
	if opts.FilterStatRadius != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.Radius, opts.FilterStatRadius)...)
	}
	if opts.FilterStatRadiusDamageFalloff != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.RadiusDamageFalloff, opts.FilterStatRadiusDamageFalloff)...)
	}
	if opts.FilterStatRateOfFire != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.RateOfFire, opts.FilterStatRateOfFire)...)
	}
	if opts.FilterStatEnergyCosts != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.PowerCost, opts.FilterStatEnergyCosts)...)
	}
	if opts.FilterStatProjectileSpeed != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.ProjectileSpeed, opts.FilterStatProjectileSpeed)...)
	}
	if opts.FilterStatSpread != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.Spread, opts.FilterStatSpread)...)
	}

	// Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"((to_tsvector('english', %s) @@ to_tsquery(?) OR (to_tsvector('english', %s::text) @@ to_tsquery(?)) OR (to_tsvector('english', %s::text) @@ to_tsquery(?)) ))",
					qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label),
					qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.WeaponType),
					qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponColumns.Label),
				),
					xSearch,
					xSearch,
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
			// CollectionItem
			boiler.CollectionItemTableColumns.CollectionSlug,
			boiler.CollectionItemTableColumns.Hash,
			boiler.CollectionItemTableColumns.TokenID,
			boiler.CollectionItemTableColumns.ItemType,
			boiler.CollectionItemTableColumns.ItemID,
			boiler.CollectionItemTableColumns.Tier,
			boiler.CollectionItemTableColumns.OwnerID,
			boiler.CollectionItemTableColumns.MarketLocked,
			boiler.CollectionItemTableColumns.XsynLocked,
			boiler.CollectionItemTableColumns.LockedToMarketplace,
			boiler.CollectionItemTableColumns.AssetHidden,
			// Images
			boiler.WeaponModelSkinCompatibilityTableColumns.ImageURL,
			boiler.WeaponModelSkinCompatibilityTableColumns.CardAnimationURL,
			boiler.WeaponModelSkinCompatibilityTableColumns.AvatarURL,
			boiler.WeaponModelSkinCompatibilityTableColumns.LargeImageURL,
			boiler.WeaponModelSkinCompatibilityTableColumns.BackgroundColor,
			boiler.WeaponModelSkinCompatibilityTableColumns.AnimationURL,
			boiler.WeaponModelSkinCompatibilityTableColumns.YoutubeURL,
			// WeaponSkin
			qm.Rels("_wsci", boiler.CollectionItemColumns.CollectionSlug),
			qm.Rels("_wsci", boiler.CollectionItemColumns.Hash),
			qm.Rels("_wsci", boiler.CollectionItemColumns.TokenID),
			qm.Rels("_wsci", boiler.CollectionItemColumns.ItemType),
			qm.Rels("_wsci", boiler.CollectionItemColumns.ItemID),
			qm.Rels("_wsci", boiler.CollectionItemColumns.Tier),
			qm.Rels("_wsci", boiler.CollectionItemColumns.OwnerID),
			qm.Rels("_wsci", boiler.CollectionItemColumns.MarketLocked),
			qm.Rels("_wsci", boiler.CollectionItemColumns.XsynLocked),
			qm.Rels("_wsci", boiler.CollectionItemColumns.AssetHidden),
			boiler.BlueprintWeaponSkinTableColumns.ImageURL,
			boiler.BlueprintWeaponSkinTableColumns.CardAnimationURL,
			boiler.BlueprintWeaponSkinTableColumns.AvatarURL,
			boiler.BlueprintWeaponSkinTableColumns.LargeImageURL,
			boiler.BlueprintWeaponSkinTableColumns.BackgroundColor,
			boiler.BlueprintWeaponSkinTableColumns.AnimationURL,
			boiler.BlueprintWeaponSkinTableColumns.YoutubeURL,
			boiler.BlueprintWeaponSkinTableColumns.Label,
			boiler.BlueprintWeaponSkinTableColumns.StatModifier,
			boiler.WeaponSkinTableColumns.ID,
			boiler.WeaponSkinTableColumns.BlueprintID,
			boiler.WeaponSkinTableColumns.EquippedOn,
			boiler.WeaponSkinTableColumns.CreatedAt,
			// Other fields
			boiler.CollectionItemTableColumns.ID,
			boiler.WeaponTableColumns.ID,
			boiler.BlueprintWeaponTableColumns.Label,
			boiler.BlueprintWeaponTableColumns.Damage,
			boiler.WeaponTableColumns.BlueprintID,
			boiler.BlueprintWeaponTableColumns.DefaultDamageType,
			boiler.WeaponTableColumns.GenesisTokenID,
			boiler.BlueprintWeaponTableColumns.WeaponType,
			boiler.BlueprintWeaponTableColumns.DamageFalloff,
			boiler.BlueprintWeaponTableColumns.DamageFalloffRate,
			boiler.BlueprintWeaponTableColumns.Spread,
			boiler.BlueprintWeaponTableColumns.RateOfFire,
			boiler.BlueprintWeaponTableColumns.Radius,
			boiler.BlueprintWeaponTableColumns.RadiusDamageFalloff,
			boiler.BlueprintWeaponTableColumns.ProjectileSpeed,
			boiler.BlueprintWeaponTableColumns.PowerCost,
			boiler.BlueprintWeaponTableColumns.MaxAmmo,
			boiler.WeaponTableColumns.UpdatedAt,
			boiler.WeaponTableColumns.CreatedAt,
			boiler.WeaponTableColumns.EquippedOn,
			boiler.WeaponTableColumns.EquippedWeaponSkinID,
			boiler.WeaponTableColumns.LockedToMech,
			fmt.Sprintf(
				`(
					SELECT _i.%s
					FROM %s _i
					WHERE _i.%s = %s
						AND _i.%s IS NULL
						AND _i.%s IS NULL
						AND _i.%s > NOW()
				) AS item_sale_id`,
				boiler.ItemSaleColumns.ID,
				boiler.TableNames.ItemSales,
				boiler.ItemSaleColumns.CollectionItemID,
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ID),
				boiler.ItemSaleColumns.SoldAt,
				boiler.ItemSaleColumns.DeletedAt,
				boiler.ItemSaleColumns.EndAt,
			),
		),
		qm.From(boiler.TableNames.CollectionItems),
	)

	if opts.SortBy != "" && opts.SortDir.IsValid() {
		if opts.SortBy == "alphabetical" {
			queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label), opts.SortDir)))
		} else if opts.SortBy == "rarity" {
			queryMods = append(queryMods, GenerateTierSort(qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.Tier), opts.SortDir))
		}
	} else {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s ASC", qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label))))
	}

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	weapons := make([]*server.Weapon, 0)
	for rows.Next() {
		w := &server.Weapon{
			CollectionItem: &server.CollectionItem{},
			Images:         &server.Images{},
			WeaponSkin: &server.WeaponSkin{
				CollectionItem: &server.CollectionItem{},
				Images:         &server.Images{},
				SkinSwatch:     &server.Images{},
			},
		}

		scanArgs := []interface{}{
			&w.CollectionItem.CollectionSlug,
			&w.CollectionItem.Hash,
			&w.CollectionItem.TokenID,
			&w.CollectionItem.ItemType,
			&w.CollectionItem.ItemID,
			&w.CollectionItem.Tier,
			&w.CollectionItem.OwnerID,
			&w.CollectionItem.MarketLocked,
			&w.CollectionItem.XsynLocked,
			&w.CollectionItem.LockedToMarketplace,
			&w.CollectionItem.AssetHidden,
			&w.Images.ImageURL,
			&w.Images.CardAnimationURL,
			&w.Images.AvatarURL,
			&w.Images.LargeImageURL,
			&w.Images.BackgroundColor,
			&w.Images.AnimationURL,
			&w.Images.YoutubeURL,
			&w.WeaponSkin.CollectionItem.CollectionSlug,
			&w.WeaponSkin.CollectionItem.Hash,
			&w.WeaponSkin.CollectionItem.TokenID,
			&w.WeaponSkin.CollectionItem.ItemType,
			&w.WeaponSkin.CollectionItem.ItemID,
			&w.WeaponSkin.CollectionItem.Tier,
			&w.WeaponSkin.CollectionItem.OwnerID,
			&w.WeaponSkin.CollectionItem.MarketLocked,
			&w.WeaponSkin.CollectionItem.XsynLocked,
			&w.WeaponSkin.CollectionItem.AssetHidden,
			&w.WeaponSkin.SkinSwatch.ImageURL,
			&w.WeaponSkin.SkinSwatch.CardAnimationURL,
			&w.WeaponSkin.SkinSwatch.AvatarURL,
			&w.WeaponSkin.SkinSwatch.LargeImageURL,
			&w.WeaponSkin.SkinSwatch.BackgroundColor,
			&w.WeaponSkin.SkinSwatch.AnimationURL,
			&w.WeaponSkin.SkinSwatch.YoutubeURL,
			&w.WeaponSkin.Label,
			&w.WeaponSkin.StatModifier,
			&w.WeaponSkin.ID,
			&w.WeaponSkin.BlueprintID,
			&w.WeaponSkin.EquippedOn,
			&w.WeaponSkin.CreatedAt,
			&w.CollectionItemID,
			&w.ID,
			&w.Label,
			&w.Damage,
			&w.BlueprintID,
			&w.DefaultDamageType,
			&w.GenesisTokenID,
			&w.WeaponType,
			&w.DamageFalloff,
			&w.DamageFalloffRate,
			&w.Spread,
			&w.RateOfFire,
			&w.Radius,
			&w.RadiusDamageFalloff,
			&w.ProjectileSpeed,
			&w.PowerCost,
			&w.MaxAmmo,
			&w.UpdatedAt,
			&w.CreatedAt,
			&w.EquippedOn,
			&w.EquippedWeaponSkinID,
			&w.LockedToMech,
			&w.ItemSaleID,
		}

		w.WeaponSkin.Images = &server.Images{
			ImageURL:         w.Images.ImageURL,
			CardAnimationURL: w.Images.CardAnimationURL,
			AvatarURL:        w.Images.AvatarURL,
			LargeImageURL:    w.Images.LargeImageURL,
			BackgroundColor:  w.Images.BackgroundColor,
			AnimationURL:     w.Images.AnimationURL,
			YoutubeURL:       w.Images.YoutubeURL,
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, weapons, err
		}
		weapons = append(weapons, w)
	}

	return total, weapons, nil
}

func WeaponSetAllEquippedAssetsAsHidden(conn boil.Executor, weaponID string, reason null.String) error {
	itemIDsToUpdate := []string{}

	// get equipped mech weapon skins
	mWpnSkin, err := boiler.WeaponSkins(
		boiler.WeaponSkinWhere.EquippedOn.EQ(null.StringFrom(weaponID)),
	).All(conn)
	if err != nil {
		return err
	}
	for _, itm := range mWpnSkin {
		itemIDsToUpdate = append(itemIDsToUpdate, itm.ID)
	}

	// update!
	_, err = boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(itemIDsToUpdate),
	).UpdateAll(conn, boiler.M{
		"asset_hidden": reason,
	})
	if err != nil {
		return err
	}

	return nil
}

type WeaponMaxStats struct {
	MaxAmmo             null.Int            `json:"max_ammo,omitempty"`
	Damage              int                 `json:"damage"`
	DamageFalloff       null.Int            `json:"damage_falloff,omitempty"`
	DamageFalloffRate   null.Int            `json:"damage_falloff_rate,omitempty"`
	Radius              null.Int            `json:"radius,omitempty"`
	RadiusDamageFalloff null.Int            `json:"radius_damage_falloff,omitempty"`
	Spread              decimal.NullDecimal `json:"spread,omitempty"`
	RateOfFire          decimal.NullDecimal `json:"rate_of_fire,omitempty"`
	ProjectileSpeed     decimal.NullDecimal `json:"projectile_speed,omitempty"`
	EnergyCost          decimal.NullDecimal `json:"energy_cost,omitempty"`
}

func GetWeaponMaxStats(conn boil.Executor, userID string) (*WeaponMaxStats, error) {
	output := &WeaponMaxStats{}

	if userID != "" {
		err := boiler.CollectionItems(
			qm.Select(
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.MaxAmmo)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Damage)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.DamageFalloff)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.DamageFalloffRate)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Radius)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.RadiusDamageFalloff)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Spread)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.RateOfFire)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ProjectileSpeed)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.PowerCost)),
			),
			qm.InnerJoin(fmt.Sprintf(
				"%s on %s = %s",
				boiler.TableNames.Weapons,
				qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
			)),
			qm.InnerJoin(fmt.Sprintf(
				"%s on %s = %s",
				boiler.TableNames.BlueprintWeapons,
				qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID),
				qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.BlueprintID),
			)),
			boiler.CollectionItemWhere.OwnerID.EQ(userID),
		).QueryRow(conn).Scan(
			&output.MaxAmmo,
			&output.Damage,
			&output.DamageFalloff,
			&output.DamageFalloffRate,
			&output.Radius,
			&output.RadiusDamageFalloff,
			&output.Spread,
			&output.RateOfFire,
			&output.ProjectileSpeed,
			&output.EnergyCost,
		)
		if err != nil {
			return nil, err
		}
		return output, nil
	}

	err := boiler.BlueprintWeapons(
		qm.Select(
			fmt.Sprintf(`MAX(%s)`, boiler.BlueprintWeaponColumns.MaxAmmo),
			fmt.Sprintf(`MAX(%s)`, boiler.BlueprintWeaponColumns.Damage),
			fmt.Sprintf(`MAX(%s)`, boiler.BlueprintWeaponColumns.DamageFalloff),
			fmt.Sprintf(`MAX(%s)`, boiler.BlueprintWeaponColumns.DamageFalloffRate),
			fmt.Sprintf(`MAX(%s)`, boiler.BlueprintWeaponColumns.Radius),
			fmt.Sprintf(`MAX(%s)`, boiler.BlueprintWeaponColumns.RadiusDamageFalloff),
			fmt.Sprintf(`MAX(%s)`, boiler.BlueprintWeaponColumns.Spread),
			fmt.Sprintf(`MAX(%s)`, boiler.BlueprintWeaponColumns.RateOfFire),
			fmt.Sprintf(`MAX(%s)`, boiler.BlueprintWeaponColumns.ProjectileSpeed),
			fmt.Sprintf(`MAX(%s)`, boiler.BlueprintWeaponColumns.PowerCost),
		),
	).QueryRow(conn).Scan(
		&output.MaxAmmo,
		&output.Damage,
		&output.DamageFalloff,
		&output.DamageFalloffRate,
		&output.Radius,
		&output.RadiusDamageFalloff,
		&output.Spread,
		&output.RateOfFire,
		&output.ProjectileSpeed,
		&output.EnergyCost,
	)
	if err != nil {
		return nil, err
	}
	return output, nil
}
