package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertNewWeapon(ownerID uuid.UUID, weapon *server.BlueprintWeapon) (*server.Weapon, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}

	newWeapon := boiler.Weapon{
		BrandID:               weapon.BrandID,
		Label:                 weapon.Label,
		Slug:                  weapon.Slug,
		Damage:                weapon.Damage,
		BlueprintID:           weapon.ID,
		DefaultDamageType:     weapon.DefaultDamageType,
		GenesisTokenID:        weapon.GenesisTokenID,
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

	err = InsertNewCollectionItem(tx, weapon.Collection, boiler.ItemTypeWeapon, newWeapon.ID, weapon.Tier, ownerID.String())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	weaponUUID, err := uuid.FromString(newWeapon.ID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return Weapon(weaponUUID)
}

func Weapon(id uuid.UUID) (*server.Weapon, error) {
	boilerMech, err := boiler.FindWeapon(gamedb.StdConn, id.String())
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id.String())).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.WeaponFromBoiler(boilerMech, boilerMechCollectionDetails), nil
}
