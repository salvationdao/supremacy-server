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

	err = InsertNewCollectionItem(tx, skin.Collection, boiler.ItemTypeMechSkin, newSkin.ID, skin.Tier, ownerID.String())
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
