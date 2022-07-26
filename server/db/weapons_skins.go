package db

import (
	"fmt"
	"github.com/volatiletech/null/v8"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertNewWeaponSkin(trx boil.Executor, ownerID uuid.UUID, blueprintWeaponSkin *server.BlueprintWeaponSkin) (*server.WeaponSkin, error) {
	tx := trx
	if trx == nil {
		tx = gamedb.StdConn
	}

	newWeaponSkin := boiler.WeaponSkin{
		BlueprintID:   blueprintWeaponSkin.ID,
		OwnerID:       ownerID.String(),
		Label:         blueprintWeaponSkin.Label,
		EquippedOn:    null.String{},
		CreatedAt:     blueprintWeaponSkin.CreatedAt,
	}

	err := newWeaponSkin.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		blueprintWeaponSkin.Collection,
		boiler.ItemTypeWeaponSkin,
		newWeaponSkin.ID,
		blueprintWeaponSkin.Tier,
		ownerID.String(),
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return WeaponSkin(tx, newWeaponSkin.ID)
}

func WeaponSkin(trx boil.Executor, id string) (*server.WeaponSkin, error) {
	boilerWeaponSkin, err := boiler.FindWeaponSkin(trx, id)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(trx)
	if err != nil {
		return nil, err
	}

	return server.WeaponSkinFromBoiler(boilerWeaponSkin, boilerMechCollectionDetails), nil
}

func WeaponSkins(id ...string) ([]*server.WeaponSkin, error) {
	var weaponSkins []*server.WeaponSkin
	boilerWeaponSkins, err := boiler.WeaponSkins(boiler.WeaponWhere.ID.IN(id)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, bws := range boilerWeaponSkins {
		boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(bws.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		weaponSkins = append(weaponSkins, server.WeaponSkinFromBoiler(bws, boilerMechCollectionDetails))
	}

	return weaponSkins, nil
}

func AttachWeaponSkinToWeapon(tx boil.Executor, ownerID, weaponID, weaponSkinID string) error {
	// check owner
	weaponCI, err := CollectionItemFromItemID(tx, weaponID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg("failed to weapon collection item")
		return terror.Error(err)
	}
	wsCI, err := CollectionItemFromItemID(tx, weaponSkinID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponSkinID", weaponSkinID).Msg("failed to weapon skin collection item")
		return terror.Error(err)
	}

	if weaponCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("weaponCI.OwnerID", weaponCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the weapon to equip skins to it.")
	}
	if wsCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("wsCI.OwnerID", wsCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the skin to equip it to a war machine.")
	}

	// get weapon
	weapon, err := boiler.FindWeapon(tx, weaponID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg("failed to find weapon")
		return terror.Error(err)
	}

	// get weapon skin
	weaponSkin, err := boiler.FindWeaponSkin(tx, weaponSkinID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponSkinID", weaponSkinID).Msg("failed to find weapon skin")
		return terror.Error(err)
	}

	// error out, already has a weapon skin
	if weapon.EquippedWeaponSkinID.Valid && weapon.EquippedWeaponSkinID.String != "" {
		err := fmt.Errorf("weapon already has a weapon skin")
		// also check weaponSkin.EquippedOn on, if that doesn't match, update it, so it does.
		if !weaponSkin.EquippedOn.Valid {
			weaponSkin.EquippedOn = null.StringFrom(weapon.ID)
			_, err = weaponSkin.Update(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Str("weapon.ID", weapon.ID).Str("weaponSkin.ID", weaponSkin.ID).Msg("failed to update weapon skin equipped on")
				return terror.Error(err, "Weapon already has a skin equipped.")
			}
		}
		gamelog.L.Error().Err(err).Str("weapon.EquippedWeaponSkinID.String", weapon.EquippedWeaponSkinID.String).Str("new weaponSkin.ID", weaponSkin.ID).Msg(err.Error())
		return terror.Error(err, "Weapon already has a skin equipped.")
	}

	// lets join
	weapon.EquippedWeaponSkinID = null.StringFrom(weaponSkin.ID)
	weaponSkin.EquippedOn = null.StringFrom(weapon.ID)

	_, err = weapon.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Str("weapon.ChassisSkinID.String", weapon.EquippedWeaponSkinID.String).Str("new weaponSkin.ID", weaponSkin.ID).Msg("failed to equip weapon skin to weapon, issue weapon update")
		return terror.Error(err, "Issue preventing equipping this weapon skin to the war machine, try again or contact support.")
	}
	_, err = weaponSkin.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Str("weapon.ChassisSkinID.String", weapon.EquippedWeaponSkinID.String).Str("new weaponSkin.ID", weaponSkin.ID).Msg("failed to equip weapon skin to weapon, issue weapon skin update")
		return terror.Error(err, "Issue preventing equipping this weapon skin to the war machine, try again or contact support.")
	}

	return nil
}
