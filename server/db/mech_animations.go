package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertNewMechAnimation(ownerID uuid.UUID, animationBlueprint *server.BlueprintMechAnimation) (*server.MechAnimation, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}

	// first insert the animation
	newAnimation := boiler.MechAnimation{
		BlueprintID:    animationBlueprint.ID,
		Label:          animationBlueprint.Label,
		MechModel:      animationBlueprint.MechModel,
		IntroAnimation: animationBlueprint.IntroAnimation,
		OutroAnimation: animationBlueprint.OutroAnimation,
	}

	err = newAnimation.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = InsertNewCollectionItem(tx,
		animationBlueprint.Collection,
		boiler.ItemTypeMechAnimation,
		newAnimation.ID,
		animationBlueprint.Tier,
		ownerID.String(),
		animationBlueprint.ImageURL,
		animationBlueprint.CardAnimationURL,
		animationBlueprint.AvatarURL,
		animationBlueprint.LargeImageURL,
		animationBlueprint.BackgroundColor,
		animationBlueprint.AnimationURL,
		animationBlueprint.YoutubeURL,
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	return MechAnimation(newAnimation.ID)
}

func MechAnimation(id string) (*server.MechAnimation, error) {
	boilerMech, err := boiler.FindMechAnimation(gamedb.StdConn, id)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.MechAnimationFromBoiler(boilerMech, boilerMechCollectionDetails), nil
}
