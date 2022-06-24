package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func BlueprintWeaponSkins(ids []string) ([]*server.BlueprintWeaponSkin, error) {
	var bluePrintWeaponSkins []*server.BlueprintWeaponSkin
	blueprintWeaponSkins, err := boiler.BlueprintWeaponSkins(boiler.BlueprintWeaponWhere.ID.IN(ids)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		for _, bp := range blueprintWeaponSkins {
			if bp.ID == id {
				bluePrintWeaponSkins = append(bluePrintWeaponSkins, server.BlueprintWeaponSkinFromBoiler(bp))
			}
		}
	}

	return bluePrintWeaponSkins, nil
}

func BlueprintWeaponSkin(ids string) (*server.BlueprintWeaponSkin, error) {
	blueprintWeaponSkin, err := boiler.BlueprintWeaponSkins(boiler.BlueprintWeaponWhere.ID.EQ(ids)).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.BlueprintWeaponSkinFromBoiler(blueprintWeaponSkin), nil
}
