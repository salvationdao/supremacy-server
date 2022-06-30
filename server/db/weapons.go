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

	boilerMech, err := boiler.FindWeapon(tx, id)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	return server.WeaponFromBoiler(boilerMech, boilerMechCollectionDetails), nil
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
		weapons = append(weapons, server.WeaponFromBoiler(bm, boilerMechCollectionDetails))
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

func WeaponList(opts *MechListOpts) (int64, []*server.Weapon, error) {
	var weapons []*server.Weapon

	var queryMods []qm.QueryMod

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
	// Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"((to_tsvector('english', %[1]s.%[2]s) @@ to_tsquery(?))",
					boiler.TableNames.Mechs,
					boiler.MechColumns.Label,
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

	fmt.Println("fuckkkkkkkkkkkkkkkkkkk")
	fmt.Println("fuckkkkkkkkkkkkkkkkkkk")
	fmt.Println("fuckkkkkkkkkkkkkkkkkkk")
	fmt.Println("fuckkkkkkkkkkkkkkkkkkk")

	// Build query
	queryMods = append(queryMods,
		qm.Select(
			// qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.CollectionSlug),
			// qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Hash),
			// qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.TokenID),
			// qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.OwnerID),
			// qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Tier),
			// qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
			// qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.MarketLocked),
			// qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.XsynLocked),
			// qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.LockedToMarketplace),
			// qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ID),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.Label),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Label),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.WeaponHardpoints),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.UtilitySlots),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Speed),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.MaxHitpoints),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IsDefault),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IsInsured),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.GenesisTokenID),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.LimitedReleaseTokenID),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.PowerCoreSize),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.PowerCoreID),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BrandID),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ModelID),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.IntroAnimationID),
			// qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.OutroAnimationID),
		),
		qm.From(boiler.TableNames.CollectionItems),
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
	)

	// Sort
	// if opts.QueueSort != nil {
	// 	queryMods = append(queryMods,
	// 		qm.Select("_bq.queue_position AS queue_position"),
	// 		qm.LeftOuterJoin(
	// 			fmt.Sprintf(`(
	// 				SELECT  _bq.mech_id, _bq.battle_contract_id, row_number () OVER (ORDER BY _bq.queued_at) AS queue_position
	// 					from battle_queue _bq
	// 					where _bq.faction_id = ?
	// 						AND _bq.battle_id IS NULL
	// 				) _bq ON _bq.mech_id = %s`,
	// 				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
	// 			),
	// 			opts.QueueSort.FactionID,
	// 		),
	// 		qm.OrderBy(fmt.Sprintf("queue_position %s NULLS LAST, %s, %s",
	// 			opts.QueueSort.SortDir,
	// 			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
	// 			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
	// 		)),
	// 	)
	// } else {
	// 	if opts.Sort != nil && opts.Sort.Table == boiler.TableNames.Mechs && IsMechColumn(opts.Sort.Column) && opts.Sort.Direction.IsValid() {
	// 		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.Mechs, opts.Sort.Column, opts.Sort.Direction)))
	// 	} else {
	// 		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s desc", boiler.TableNames.Mechs, boiler.MechColumns.Name)))
	// 	}
	// }

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
			// &wp.CollectionItem.CollectionSlug,
			// &wp.CollectionItem.Hash,
			// &wp.CollectionItem.TokenID,
			// &wp.CollectionItem.OwnerID,
			// &wp.CollectionItem.Tier,
			// &wp.CollectionItem.ItemType,
			// &wp.CollectionItem.MarketLocked,
			// &wp.CollectionItem.XsynLocked,
			// &wp.CollectionItem.LockedToMarketplace,
			// &wp.CollectionItemID,
			&wp.ID,
			// &wp.Name,
			&wp.Label,
			// &wp.WeaponHardpoints,
			// &wp.UtilitySlots,
			// &wp.Speed,
			// &wp.MaxHitpoints,
			// &wp.IsDefault,
			// &wp.IsInsured,
			// &wp.GenesisTokenID,
			// &wp.LimitedReleaseTokenID,
			// &wp.PowerCoreSize,
			// &wp.PowerCoreID,
			// &wp.BlueprintID,
			// &wp.BrandID,
			// &wp.ModelID,
			// &wp.ChassisSkinID,
			// // &wp.IntroAnimationID,
			// &wp.OutroAnimationID,
		}
		// if opts.QueueSort != nil {
		// 	scanArgs = append(scanArgs, &wp.QueuePosition)
		// }
		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, weapons, err
		}
		weapons = append(weapons, wp)
	}

	return total, weapons, nil
}
