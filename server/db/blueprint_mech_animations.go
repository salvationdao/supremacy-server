package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func BlueprintMechAnimations(ids []string) ([]*server.BlueprintMechAnimation, error) {
	var blueprintMechAnimations []*server.BlueprintMechAnimation
	boilerBlueprintMechAnimations, err := boiler.BlueprintMechAnimations(boiler.BlueprintMechAnimationWhere.ID.IN(ids)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		for _, bp := range boilerBlueprintMechAnimations {
			if bp.ID == id {
				blueprintMechAnimations = append(blueprintMechAnimations, server.BlueprintMechAnimationFromBoiler(bp))
			}
		}
	}

	return blueprintMechAnimations, nil
}

func BlueprintMechAnimation(ids string) (*server.BlueprintMechAnimation, error) {
	blueprintMechAnimation, err := boiler.BlueprintMechAnimations(boiler.BlueprintMechAnimationWhere.ID.EQ(ids)).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.BlueprintMechAnimationFromBoiler(blueprintMechAnimation), nil
}
