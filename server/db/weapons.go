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
		// join weapon model
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.WeaponModels,
			qm.Rels(boiler.TableNames.WeaponModels, boiler.WeaponModelColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.WeaponModelID),
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
			qm.Rels(boiler.TableNames.WeaponModels, boiler.WeaponModelColumns.ID),
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.BlueprintWeaponSkinID),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
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
	wpnSkin, err := InsertNewWeaponSkin(tx, ownerID, weaponSkin, &weapon.WeaponModelID)
	if err != nil {
		L.Error().Err(err).Msg("failed to insert new weapon skin")
		return nil, nil, err
	}

	newWeapon := boiler.Weapon{
		Slug:                  weapon.Slug,
		Damage:                weapon.Damage,
		BlueprintID:           weapon.ID,
		DefaultDamageType:     weapon.DefaultDamageType,
		GenesisTokenID:        weapon.GenesisTokenID,
		LimitedReleaseTokenID: weapon.LimitedReleaseTokenID,
		DamageFalloff:         weapon.DamageFalloff,
		DamageFalloffRate:     weapon.DamageFalloffRate,
		Spread:                weapon.Spread,
		RateOfFire:            weapon.RateOfFire,
		Radius:                weapon.Radius,
		RadiusDamageFalloff:   weapon.RadiusDamageFalloff,
		ProjectileSpeed:       weapon.ProjectileSpeed,
		EnergyCost:            weapon.EnergyCost,
		MaxAmmo:               weapon.MaxAmmo,
		EquippedWeaponSkinID:  wpnSkin.ID,
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
		weapon.Tier,
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

	weaponSkin, err := WeaponSkin(tx, boilerWeapon.EquippedWeaponSkinID, &boilerWeapon.R.Blueprint.WeaponModelID)
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
		gamelog.L.Error().Err(err).Str("mechID", mech.ID).Msg("failed to check for available slots on mech")
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
		boiler.WeaponColumns.Slug,
		boiler.WeaponColumns.Damage,
		boiler.WeaponColumns.DeletedAt,
		boiler.WeaponColumns.UpdatedAt,
		boiler.WeaponColumns.CreatedAt,
		boiler.WeaponColumns.BlueprintID,
		boiler.WeaponColumns.EquippedOn,
		boiler.WeaponColumns.DefaultDamageType,
		boiler.WeaponColumns.GenesisTokenID,
		boiler.WeaponColumns.LimitedReleaseTokenID,
		boiler.WeaponColumns.DamageFalloff,
		boiler.WeaponColumns.DamageFalloffRate,
		boiler.WeaponColumns.Radius,
		boiler.WeaponColumns.RadiusDamageFalloff,
		boiler.WeaponColumns.Spread,
		boiler.WeaponColumns.RateOfFire,
		boiler.WeaponColumns.ProjectileSpeed,
		boiler.WeaponColumns.EnergyCost,
		boiler.WeaponColumns.IsMelee,
		boiler.WeaponColumns.MaxAmmo,
		boiler.WeaponColumns.LockedToMech,
		boiler.WeaponColumns.EquippedWeaponSkinID:
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
		output = append(output, qm.Where(qm.Rels(boiler.TableNames.Weapons, column)+" >= ?", filter.Min))
	}
	if filter.Max.Valid {
		output = append(output, qm.Where(qm.Rels(boiler.TableNames.Weapons, column)+" <= ?", filter.Max))
	}
	return output
}

func WeaponList(opts *WeaponListOpts) (int64, []*server.Weapon, error) {
	var weapons []*server.Weapon

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
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.WeaponColumns.MaxAmmo, opts.FilterStatAmmo)...)
	}
	if opts.FilterStatDamage != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.WeaponColumns.Damage, opts.FilterStatDamage)...)
	}
	if opts.FilterStatDamageFalloff != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.WeaponColumns.DamageFalloff, opts.FilterStatDamageFalloff)...)
	}
	if opts.FilterStatDamageFalloffRate != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.WeaponColumns.DamageFalloffRate, opts.FilterStatDamageFalloffRate)...)
	}
	if opts.FilterStatRadius != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.WeaponColumns.Radius, opts.FilterStatRadius)...)
	}
	if opts.FilterStatRadiusDamageFalloff != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.WeaponColumns.RadiusDamageFalloff, opts.FilterStatRadiusDamageFalloff)...)
	}
	if opts.FilterStatRateOfFire != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.WeaponColumns.RateOfFire, opts.FilterStatRateOfFire)...)
	}
	if opts.FilterStatEnergyCosts != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.WeaponColumns.EnergyCost, opts.FilterStatEnergyCosts)...)
	}
	if opts.FilterStatProjectileSpeed != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.WeaponColumns.ProjectileSpeed, opts.FilterStatProjectileSpeed)...)
	}
	if opts.FilterStatSpread != nil {
		queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.WeaponColumns.Spread, opts.FilterStatSpread)...)
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

	for rows.Next() {
		wp := &server.Weapon{
			CollectionItem: &server.CollectionItem{},
		}

		scanArgs := []interface{}{
			&wp.CollectionItem.CollectionSlug,
			&wp.CollectionItem.Hash,
			&wp.CollectionItem.TokenID,
			&wp.CollectionItem.OwnerID,
			&wp.CollectionItem.Tier,
			&wp.CollectionItem.ItemType,
			&wp.CollectionItem.MarketLocked,
			&wp.CollectionItem.XsynLocked,
			&wp.CollectionItem.LockedToMarketplace,
			&wp.CollectionItem.AssetHidden,
			&wp.ID,
			&wp.Label,
			&wp.ItemSaleID,
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, weapons, err
		}
		weapons = append(weapons, wp)
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
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.MaxAmmo)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.Damage)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.DamageFalloff)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.DamageFalloffRate)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.Radius)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.RadiusDamageFalloff)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.Spread)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.RateOfFire)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ProjectileSpeed)),
				fmt.Sprintf(`MAX(%s)`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.EnergyCost)),
			),
			qm.InnerJoin(fmt.Sprintf(
				"%s on %s = %s",
				boiler.TableNames.Weapons,
				qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
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

	err := boiler.Weapons(
		qm.Select(
			fmt.Sprintf(`MAX(%s)`, boiler.WeaponColumns.MaxAmmo),
			fmt.Sprintf(`MAX(%s)`, boiler.WeaponColumns.Damage),
			fmt.Sprintf(`MAX(%s)`, boiler.WeaponColumns.DamageFalloff),
			fmt.Sprintf(`MAX(%s)`, boiler.WeaponColumns.DamageFalloffRate),
			fmt.Sprintf(`MAX(%s)`, boiler.WeaponColumns.Radius),
			fmt.Sprintf(`MAX(%s)`, boiler.WeaponColumns.RadiusDamageFalloff),
			fmt.Sprintf(`MAX(%s)`, boiler.WeaponColumns.Spread),
			fmt.Sprintf(`MAX(%s)`, boiler.WeaponColumns.RateOfFire),
			fmt.Sprintf(`MAX(%s)`, boiler.WeaponColumns.ProjectileSpeed),
			fmt.Sprintf(`MAX(%s)`, boiler.WeaponColumns.EnergyCost),
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
