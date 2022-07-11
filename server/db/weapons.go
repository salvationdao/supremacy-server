package db

import (
	"database/sql"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

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
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.Label),
		),
		qm.From(boiler.TableNames.CollectionItems),
		qm.InnerJoin(fmt.Sprintf(
			"%s on %s = %s",
			boiler.TableNames.Weapons,
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
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

func InsertNewWeapon(trx boil.Executor, ownerID uuid.UUID, weapon *server.BlueprintWeapon) (*server.Weapon, error) {
	tx := trx
	if trx == nil {
		tx = gamedb.StdConn
	}

	//getting weapon model to get default skin id to get image url on blueprint weapon skins
	weaponModel, err := boiler.WeaponModels(
		boiler.WeaponModelWhere.ID.EQ(weapon.WeaponModelID),
		qm.Load(boiler.WeaponModelRels.DefaultSkin),
	).One(tx)
	if err != nil {
		return nil, terror.Error(err)
	}

	if weaponModel.R == nil || weaponModel.R.DefaultSkin == nil {
		return nil, terror.Error(fmt.Errorf("could not find default skin relationship to weapon"), "Could not find weapon default skin relationship, try again or contact support")
	}

	//should only have one in the arr
	bpws := weaponModel.R.DefaultSkin

	newWeapon := boiler.Weapon{
		BrandID:               weapon.BrandID,
		Label:                 weapon.Label,
		Slug:                  weapon.Slug,
		Damage:                weapon.Damage,
		BlueprintID:           weapon.ID,
		DefaultDamageType:     weapon.DefaultDamageType,
		GenesisTokenID:        weapon.GenesisTokenID,
		WeaponModelID:         null.StringFrom(weapon.WeaponModelID),
		LimitedReleaseTokenID: weapon.LimitedReleaseTokenID,
		WeaponType:            weapon.WeaponType,
		DamageFalloff:         weapon.DamageFalloff,
		DamageFalloffRate:     weapon.DamageFalloffRate,
		Spread:                weapon.Spread,
		RateOfFire:            weapon.RateOfFire,
		Radius:                weapon.Radius,
		RadiusDamageFalloff:   weapon.RadiusDamageFalloff,
		ProjectileSpeed:       weapon.ProjectileSpeed,
		EnergyCost:            weapon.EnergyCost,
		MaxAmmo:               weapon.MaxAmmo,
	}

	err = newWeapon.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	//change img, avatar etc. here but have to get it from blueprint weapon skins
	_, err = InsertNewCollectionItem(tx,
		weapon.Collection,
		boiler.ItemTypeWeapon,
		newWeapon.ID,
		weapon.Tier,
		ownerID.String(),
		bpws.ImageURL,
		bpws.CardAnimationURL,
		bpws.AvatarURL,
		bpws.LargeImageURL,
		bpws.BackgroundColor,
		bpws.AnimationURL,
		bpws.YoutubeURL,
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return Weapon(tx, newWeapon.ID)
}

func Weapon(trx boil.Executor, id string) (*server.Weapon, error) {
	tx := trx
	if trx == nil {
		tx = gamedb.StdConn
	}

	boilerWeapon, err := boiler.FindWeapon(tx, id)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	var weaponSkin *server.WeaponSkin
	if boilerWeapon.EquippedWeaponSkinID.Valid {
		weaponSkin, err = WeaponSkin(tx, boilerWeapon.EquippedWeaponSkinID.String)
		if err != nil {
			return nil, err
		}
	}

	return server.WeaponFromBoiler(boilerWeapon, boilerMechCollectionDetails, weaponSkin), nil
}

func Weapons(id ...string) ([]*server.Weapon, error) {
	var weapons []*server.Weapon
	boilerMechs, err := boiler.Weapons(boiler.WeaponWhere.ID.IN(id)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, bm := range boilerMechs {
		boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(bm.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}

		var weaponSkin *server.WeaponSkin
		if bm.EquippedWeaponSkinID.Valid {
			weaponSkin, err = WeaponSkin(gamedb.StdConn, bm.EquippedWeaponSkinID.String)
			if err != nil {
				return nil, err
			}
		}
		weapons = append(weapons, server.WeaponFromBoiler(bm, boilerMechCollectionDetails, weaponSkin))
	}

	return weapons, nil
}

// AttachWeaponToMech attaches a Weapon to a mech  TODO: create tests.
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

	// check current weapon count
	if len(mech.R.ChassisMechWeapons)+1 > mech.WeaponHardpoints {
		err := fmt.Errorf("weapon cannot fit")
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg("adding this weapon brings mechs weapons over mechs weapon hardpoints")
		return terror.Error(err, fmt.Sprintf("War machine already has %d weapons equipped and is only has %d weapon hardpoints.", len(mech.R.ChassisMechWeapons), mech.WeaponHardpoints))
	}

	// check weapon isn't already equipped to another war machine
	exists, err := boiler.MechWeapons(boiler.MechWeaponWhere.WeaponID.EQ(weaponID)).Exists(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg("failed to check if a mech and weapon join already exists")
		return terror.Error(err)
	}
	if exists {
		err := fmt.Errorf("weapon already equipped to a warmachine")
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg(err.Error())
		return terror.Error(err, "This weapon is already equipped to another war machine, try again or contact support.")
	}

	weapon.EquippedOn = null.StringFrom(mech.ID)

	_, err = weapon.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("weapon", weapon).Msg("failed to update weapon")
		return terror.Error(err, "Issue preventing equipping this weapon to the war machine, try again or contact support.")
	}

	weaponMechJoin := boiler.MechWeapon{
		ChassisID:  mech.ID,
		WeaponID:   weapon.ID,
		SlotNumber: len(mech.R.ChassisMechWeapons), // slot number starts at 0, so if we currently have 2 equipped and this is the 3rd, it will be slot 2.
	}

	err = weaponMechJoin.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("weaponMechJoin", weaponMechJoin).Msg(" failed to equip weapon to war machine")
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
			qm.Or(fmt.Sprintf(`%s IS NOT NULL`, qm.Rels(boiler.TableNames.MechWeapons, boiler.MechWeaponColumns.ID))),
		),
	).Exists(gamedb.StdConn)
	if err != nil {
		return false, terror.Error(err)
	}
	return exists, nil
}

type WeaponListOpts struct {
	Search                   string
	Filter                   *ListFilterRequest
	Sort                     *ListSortRequest
	PageSize                 int
	Page                     int
	OwnerID                  string
	DisplayXsynMechs         bool
	DisplayGenesisAndLimited bool
	DisplayHidden            bool
	ExcludeMarketLocked      bool
	IncludeMarketListed      bool
	ExcludeEquipped          bool
	FilterRarities           []string `json:"rarities"`
	FilterWeaponTypes        []string `json:"weapon_types"`
}

func WeaponList(opts *WeaponListOpts) (int64, []*server.Weapon, error) {
	var weapons []*server.Weapon

	queryMods := []qm.QueryMod{
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Weapons,
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
	}

	// create the where owner id = clause
	queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
		Table:    boiler.TableNames.CollectionItems,
		Column:   boiler.CollectionItemColumns.OwnerID,
		Operator: OperatorValueTypeEquals,
		Value:    opts.OwnerID,
	}, 0, ""),
		GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.ItemType,
			Operator: OperatorValueTypeEquals,
			Value:    boiler.ItemTypeWeapon,
		}, 0, "and"),
	)

	if !opts.DisplayXsynMechs || !opts.IncludeMarketListed {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.XsynLocked,
			Operator: OperatorValueTypeIsFalse,
		}, 0, ""))
	}
	if opts.ExcludeMarketLocked {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.MarketLocked,
			Operator: OperatorValueTypeIsFalse,
		}, 0, ""))
	}
	if !opts.IncludeMarketListed {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.LockedToMarketplace,
			Operator: OperatorValueTypeIsFalse,
		}, 0, ""))
	}
	if !opts.DisplayGenesisAndLimited {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.Weapons,
			Column:   boiler.WeaponColumns.GenesisTokenID,
			Operator: OperatorValueTypeIsNull,
		}, 0, ""))
	}
	if !opts.DisplayGenesisAndLimited {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.Weapons,
			Column:   boiler.WeaponColumns.LimitedReleaseTokenID,
			Operator: OperatorValueTypeIsNull,
		}, 0, ""))
	}
	if !opts.DisplayHidden {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.AssetHidden,
			Operator: OperatorValueTypeIsNull,
		}, 0, ""))
	}
	if opts.ExcludeEquipped {
		queryMods = append(queryMods,
			qm.LeftOuterJoin(fmt.Sprintf(
				`%s on %s = %s`,
				boiler.TableNames.MechWeapons,
				qm.Rels(boiler.TableNames.MechWeapons, boiler.MechWeaponColumns.WeaponID),
				qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
			)),
			qm.Expr(
				boiler.WeaponWhere.EquippedOn.IsNull(),
				qm.Or(fmt.Sprintf(`%s IS NULL`, qm.Rels(boiler.TableNames.MechWeapons, boiler.MechWeaponColumns.ID))),
			),
		)
	}

	// Filters
	if opts.Filter != nil {
		// if we have filter
		for i, f := range opts.Filter.Items {
			// validate it is the right table and valid column
			if f.Table == boiler.TableNames.Weapons && IsMechColumn(f.Column) {
				queryMods = append(queryMods, GenerateListFilterQueryMod(*f, i+1, opts.Filter.LinkOperator))
			}

		}
	}

	if len(opts.FilterRarities) > 0 {
		queryMods = append(queryMods, boiler.CollectionItemWhere.Tier.IN(opts.FilterRarities))
	}

	// Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,

				qm.And(fmt.Sprintf(
					"((to_tsvector('english', %[1]s.%[2]s) @@ to_tsquery(?) OR (to_tsvector('english', %[3]s.%[4]s::text) @@ to_tsquery(?)) ))",
					boiler.TableNames.Weapons,
					boiler.WeaponColumns.Label,
					boiler.TableNames.Weapons,
					boiler.WeaponColumns.WeaponType,
				),
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
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.MarketLocked),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.XsynLocked),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.LockedToMarketplace),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.AssetHidden),

			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.Label),
		),
		qm.From(boiler.TableNames.CollectionItems),
	)

	if len(opts.FilterWeaponTypes) > 0 {
		queryMods = append(queryMods, boiler.WeaponWhere.WeaponType.IN(opts.FilterWeaponTypes))
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
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, weapons, err
		}
		weapons = append(weapons, wp)
	}

	return total, weapons, nil
}

// PlayerWeaponsList returns a list of tallied player weapons, ordered by last purchased date from the weapons table.
func PlayerWeaponsList(
	userID string,
) ([]*boiler.Weapon, error) {

	items, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.OwnerID.EQ(userID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeWeapon),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	ids := []string{}

	for _, i := range items {
		ids = append(ids, i.ItemID)
	}

	// get weapons
	weapons, err := boiler.Weapons(boiler.WeaponWhere.ID.IN(ids)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return weapons, nil
}

func WeaponSetAllEquippedAssetsAsHidden(trx boil.Executor, weaponID string, reason null.String) error {
	tx := trx
	if trx == nil {
		tx = gamedb.StdConn
	}

	itemIDsToUpdate := []string{}

	// get equipped mech weapon skins
	mWpnSkin, err := boiler.WeaponSkins(
		boiler.WeaponSkinWhere.EquippedOn.EQ(null.StringFrom(weaponID)),
	).All(tx)
	if err != nil {
		return err
	}
	for _, itm := range mWpnSkin {
		itemIDsToUpdate = append(itemIDsToUpdate, itm.ID)
	}

	// update!
	_, err = boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.IN(itemIDsToUpdate),
	).UpdateAll(tx, boiler.M{
		"asset_hidden": reason,
	})
	if err != nil {
		return err
	}

	return nil
}
