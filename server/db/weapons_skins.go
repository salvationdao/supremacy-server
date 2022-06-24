package db

import (
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertNewWeaponSkin(ownerID uuid.UUID, blueprintWeaponSkin *server.BlueprintWeaponSkin) (*server.WeaponSkin, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}

	//getting blueprintWeaponSkin model to get default skin id to get image url on blueprint blueprintWeaponSkin skins
	weaponModel, err := boiler.WeaponModels(
		boiler.WeaponModelWhere.ID.EQ(blueprintWeaponSkin.WeaponModelID),
		qm.Load(boiler.WeaponModelRels.DefaultSkin),
	).One(tx)
	if err != nil {
		return nil, terror.Error(err)
	}

	if weaponModel.R == nil || weaponModel.R.DefaultSkin == nil {
		return nil, terror.Error(fmt.Errorf("could not find default skin relationship to blueprintWeaponSkin"), "Could not find blueprintWeaponSkin default skin relationship, try again or contact support")
	}

	//should only have one in the arr
	bpws := weaponModel.R.DefaultSkin

	newWeaponSkin := boiler.WeaponSkin{
		BlueprintID:   blueprintWeaponSkin.ID,
		OwnerID:       ownerID.String(),
		Label:         blueprintWeaponSkin.Label,
		WeaponType:    blueprintWeaponSkin.WeaponType,
		EquippedOn:    null.String{},
		Tier:          blueprintWeaponSkin.Tier,
		CreatedAt:     blueprintWeaponSkin.CreatedAt,
		WeaponModelID: blueprintWeaponSkin.WeaponModelID,
	}

	err = newWeaponSkin.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	//change img, avatar etc. here but have to get it from blueprint blueprintWeaponSkin skins
	_, err = InsertNewCollectionItem(tx,
		blueprintWeaponSkin.Collection,
		boiler.ItemTypeWeaponSkin,
		newWeaponSkin.ID,
		blueprintWeaponSkin.Tier,
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

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	return WeaponSkin(newWeaponSkin.ID)
}

func WeaponSkin(id string) (*server.WeaponSkin, error) {
	boilerWeaponSkin, err := boiler.FindWeaponSkin(gamedb.StdConn, id)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(gamedb.StdConn)
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

func AttachWeaponSkinToWeapon(ownerID, weaponID, weaponSkinID string) error {
	// check owner
	weaponCI, err := CollectionItemFromItemID(weaponID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg("failed to weapon collection item")
		return terror.Error(err)
	}
	wsCI, err := CollectionItemFromItemID(weaponSkinID)
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
	weapon, err := boiler.FindWeapon(gamedb.StdConn, weaponID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponID", weaponID).Msg("failed to find weapon")
		return terror.Error(err)
	}

	// get weapon skin
	weaponSkin, err := boiler.FindWeaponSkin(gamedb.StdConn, weaponSkinID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("weaponSkinID", weaponSkinID).Msg("failed to find weapon skin")
		return terror.Error(err)
	}

	// wrong model
	if weapon.WeaponModelID != null.StringFrom(weaponSkin.WeaponModelID) {
		err := fmt.Errorf("weaponSkin model mismatch")
		gamelog.L.Error().Err(err).Str("weapon.WeaponModelID", weapon.WeaponModelID.String).Str("weaponSkin.WeaponModelID", weaponSkin.WeaponModelID).Msg("weapon skin doesn't fit this weapon")
		return terror.Error(err, "This weapon skin doesn't fit this weapon.")
	}

	// error out, already has a weapon skin
	if weapon.EquippedWeaponSkinID.Valid && weapon.EquippedWeaponSkinID.String != "" {
		err := fmt.Errorf("weapon already has a weapon skin")
		// also check weaponSkin.EquippedOn on, if that doesn't match, update it, so it does.
		if !weaponSkin.EquippedOn.Valid {
			weaponSkin.EquippedOn = null.StringFrom(weapon.ID)
			_, err = weaponSkin.Update(gamedb.StdConn, boil.Infer())
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

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Str("weapon.EquippedWeaponSkinID.String", weapon.EquippedWeaponSkinID.String).Str("new weaponSkin.ID", weaponSkin.ID).Msg("failed to equip weapon skin to weapon, issue creating tx")
		return terror.Error(err, "Issue preventing equipping this weapon skin to the war machine, try again or contact support.")
	}

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

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Str("weapon.ChassisSkinID.String", weapon.EquippedWeaponSkinID.String).Str("new weaponSkin.ID", weaponSkin.ID).Msg("failed to equip weapon skin to weapon, issue committing tx")
		return terror.Error(err, "Issue preventing equipping this mech skin to the war machine, try again or contact support.")
	}

	return nil
}
