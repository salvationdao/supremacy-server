package db

import (
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"server"
	"server/db/boiler"
)

func InsertNewMechAnimation(tx boil.Executor, ownerID uuid.UUID, animationBlueprint *server.BlueprintMechAnimation) (*server.MechAnimation, error) {
	// first insert the animation
	newAnimation := boiler.MechAnimation{
		BlueprintID:    animationBlueprint.ID,
		Label:          animationBlueprint.Label,
		MechModel:      animationBlueprint.MechModel,
		IntroAnimation: animationBlueprint.IntroAnimation,
		OutroAnimation: animationBlueprint.OutroAnimation,
	}

	err := newAnimation.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		animationBlueprint.Collection,
		boiler.ItemTypeMechAnimation,
		newAnimation.ID,
		animationBlueprint.Tier,
		ownerID.String(),
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return MechAnimation(tx, newAnimation.ID)
}

func MechAnimation(tx boil.Executor, id string) (*server.MechAnimation, error) {
	boilerMech, err := boiler.FindMechAnimation(tx, id)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	return server.MechAnimationFromBoiler(boilerMech, boilerMechCollectionDetails), nil
}
