package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"

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
		OwnerID:               ownerID.String(),
		MechModel:             skin.MechModel,
		Tier:                  skin.Tier,
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

	//insert collection item
	collectionItem := boiler.CollectionItem{
		CollectionSlug: skin.Collection,
		ItemType:       boiler.ItemTypeMechSkin,
		ItemID:         newSkin.ID,
	}

	err = collectionItem.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	skinUUID, err := uuid.FromString(newSkin.ID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return MechSkin(skinUUID)
}

func MechSkin(id uuid.UUID) (*server.MechSkin, error) {
	boilerMech, err := boiler.FindMechSkin(gamedb.StdConn, id.String())
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id.String())).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.MechSkinFromBoiler(boilerMech, boilerMechCollectionDetails), nil
}
