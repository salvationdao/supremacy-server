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
