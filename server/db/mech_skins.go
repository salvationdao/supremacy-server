package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/volatiletech/null/v8"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"
)

func InsertNewMechSkin(ownerID uuid.UUID, skin *server.BlueprintMechSkin) (*server.MechSkin, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}

	// first insert the skin
	newSkin := boiler.MechSkin{
		BlueprintID:           skin.ID,
		GenesisTokenID:        skin.GenesisTokenID,
		LimitedReleaseTokenID: skin.LimitedReleaseTokenID,
		Label:                 skin.Label,
		MechModel:             skin.MechModel,
		ImageURL:              skin.ImageURL,
		AnimationURL:          skin.AnimationURL,
		CardAnimationURL:      skin.CardAnimationURL,
		AvatarURL:             skin.AvatarURL,
		LargeImageURL:         skin.LargeImageURL,
	}

	err = newSkin.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = InsertNewCollectionItem(tx,
		skin.Collection,
		boiler.ItemTypeMechSkin,
		newSkin.ID,
		skin.Tier,
		ownerID.String(),
		skin.ImageURL,
		skin.CardAnimationURL,
		skin.AvatarURL,
		skin.LargeImageURL,
		skin.BackgroundColor,
		skin.AnimationURL,
		skin.YoutubeURL,
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	return MechSkin(newSkin.ID)
}

func MechSkin(id string) (*server.MechSkin, error) {
	boilerMech, err := boiler.FindMechSkin(gamedb.StdConn, id)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.MechSkinFromBoiler(boilerMech, boilerMechCollectionDetails), nil
}

func MechSkins(id ...string) ([]*server.MechSkin, error) {
	var skins []*server.MechSkin
	boilerMechSkins, err := boiler.MechSkins(boiler.MechSkinWhere.ID.IN(id)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, ms := range boilerMechSkins {
		boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(ms.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		skins = append(skins, server.MechSkinFromBoiler(ms, boilerMechCollectionDetails))
	}
	return skins, nil
}

// AttachMechSkinToMech attaches a mech skin to a mech  TODO: create tests.
func AttachMechSkinToMech(ownerID, mechID, chassisSkinID string) error {
	// TODO: possible optimize this, 6 queries to attach a part seems like a lot?
	// check owner
	mechCI, err := CollectionItemFromItemID(mechID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to mech collection item")
		return terror.Error(err)
	}
	msCI, err := CollectionItemFromItemID(chassisSkinID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("chassisSkinID", chassisSkinID).Msg("failed to mech skin collection item")
		return terror.Error(err)
	}

	if mechCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("mechCI.OwnerID", mechCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the war machine to equip skins to it.")
	}
	if msCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("msCI.OwnerID", msCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the skin to equip it to a war machine.")
	}

	// get mech
	mech, err := boiler.FindMech(gamedb.StdConn, mechID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to find mech")
		return terror.Error(err)
	}

	// get mech skin
	mechSkin, err := boiler.FindMechSkin(gamedb.StdConn, chassisSkinID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("chassisSkinID", chassisSkinID).Msg("failed to find mech skin")
		return terror.Error(err)
	}

	// wrong model
	if mech.ModelID != mechSkin.MechModel {
		err := fmt.Errorf("mechSkin model mismatch")
		gamelog.L.Error().Err(err).Str("mech.ModelID", mech.ModelID).Str("mechSkin.MechModel", mechSkin.MechModel).Msg("mech skin doesn't fit this mech")
		return terror.Error(err, "This war machine skin doesn't fit this war machine.")
	}

	// error out, already has a mech skin
	if mech.ChassisSkinID.Valid && mech.ChassisSkinID.String != "" {
		err := fmt.Errorf("warmachine already has a mech skin")
		// also check mechSkin.EquippedOn on, if that doesn't match, update it, so it does.
		if !mechSkin.EquippedOn.Valid {
			mechSkin.EquippedOn = null.StringFrom(mech.ID)
			_, err = mechSkin.Update(gamedb.StdConn, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Str("mech.ID", mech.ID).Str("mechSkin.ID", mechSkin.ID).Msg("failed to update mech skin equipped on")
				return terror.Error(err, "War machine already has a skin equipped.")
			}
		}
		gamelog.L.Error().Err(err).Str("mech.ChassisSkinID.String", mech.ChassisSkinID.String).Str("new mechSkin.ID", mechSkin.ID).Msg(err.Error())
		return terror.Error(err, "War machine already has a skin equipped.")
	}

	// lets join
	mech.ChassisSkinID = null.StringFrom(mechSkin.ID)
	mechSkin.EquippedOn = null.StringFrom(mech.ID)

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech.ChassisSkinID.String", mech.ChassisSkinID.String).Str("new mechSkin.ID", mechSkin.ID).Msg("failed to equip mech skin to mech, issue creating tx")
		return terror.Error(err, "Issue preventing equipping this mech skin to the war machine, try again or contact support.")
	}

	_, err = mech.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech.ChassisSkinID.String", mech.ChassisSkinID.String).Str("new mechSkin.ID", mechSkin.ID).Msg("failed to equip mech skin to mech, issue mech update")
		return terror.Error(err, "Issue preventing equipping this mech skin to the war machine, try again or contact support.")
	}
	_, err = mechSkin.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech.ChassisSkinID.String", mech.ChassisSkinID.String).Str("new mechSkin.ID", mechSkin.ID).Msg("failed to equip mech skin to mech, issue mech skin update")
		return terror.Error(err, "Issue preventing equipping this mech skin to the war machine, try again or contact support.")
	}

	err = tx.Commit()
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech.ChassisSkinID.String", mech.ChassisSkinID.String).Str("new mechSkin.ID", mechSkin.ID).Msg("failed to equip mech skin to mech, issue committing tx")
		return terror.Error(err, "Issue preventing equipping this mech skin to the war machine, try again or contact support.")
	}

	return nil
}
