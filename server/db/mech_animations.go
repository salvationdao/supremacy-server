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
		OwnerID:        ownerID.String(),
		MechModel:      animationBlueprint.MechModel,
		Tier:           animationBlueprint.Tier,
		IntroAnimation: animationBlueprint.IntroAnimation,
		OutroAnimation: animationBlueprint.OutroAnimation,
	}

	err = newAnimation.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	//insert collection item
	collectionItem := boiler.CollectionItem{
		CollectionSlug: animationBlueprint.Collection,
		ItemType:       boiler.ItemTypeMechAnimation,
		ItemID:         newAnimation.ID,
	}

	err = collectionItem.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	animationUUID, err := uuid.FromString(newAnimation.ID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return MechAnimation(animationUUID)
}

func MechAnimation(id uuid.UUID) (*server.MechAnimation, error) {
	boilerMech, err := boiler.FindMechAnimation(gamedb.StdConn, id.String())
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id.String())).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.MechAnimationFromBoiler(boilerMech, boilerMechCollectionDetails), nil
}
