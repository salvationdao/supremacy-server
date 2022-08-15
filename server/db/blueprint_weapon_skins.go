package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func BlueprintWeaponSkins(ids []string) ([]*server.BlueprintWeaponSkin, error) {
	var serverBlueprintWeaponSkins []*server.BlueprintWeaponSkin
	blueprintWeaponSkins, err := boiler.BlueprintWeaponSkins(
		boiler.BlueprintWeaponSkinWhere.ID.IN(ids),
		).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		for _, bp := range blueprintWeaponSkins {
			if bp.ID == id {
				serverBlueprintWeaponSkins = append(serverBlueprintWeaponSkins, server.BlueprintWeaponSkinFromBoiler(bp))
			}
		}
	}

	return serverBlueprintWeaponSkins, nil
}

func BlueprintWeaponSkin(ids string) (*server.BlueprintWeaponSkin, error) {
	blueprintWeaponSkin, err := boiler.BlueprintWeaponSkins(boiler.BlueprintWeaponSkinWhere.ID.EQ(ids)).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.BlueprintWeaponSkinFromBoiler(blueprintWeaponSkin), nil
}
