package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func BlueprintMechSkinSkins(ids []string) ([]*server.BlueprintMechSkin, error) {
	blueprintMechSkins := []*server.BlueprintMechSkin{}

	boilerBlueprintMechSkins, err := boiler.BlueprintMechSkins(boiler.BlueprintMechSkinWhere.ID.IN(ids)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		for _, bp := range boilerBlueprintMechSkins {
			if bp.ID == id {
				blueprintMechSkins = append(blueprintMechSkins, server.BlueprintMechSkinFromBoiler(bp))
			}
		}
	}

	return blueprintMechSkins, nil
}

func BlueprintMechSkinSkin(ids string) (*server.BlueprintMechSkin, error) {
	blueprintMechSkin, err := boiler.BlueprintMechSkins(boiler.BlueprintMechSkinWhere.ID.EQ(ids)).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.BlueprintMechSkinFromBoiler(blueprintMechSkin), nil
}
