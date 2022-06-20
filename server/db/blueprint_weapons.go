package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func BlueprintWeapons(ids []string) ([]*server.BlueprintWeapon, error) {
	var bluePrintWeapons []*server.BlueprintWeapon
	blueprintWeapons, err := boiler.BlueprintWeapons(boiler.BlueprintWeaponWhere.ID.IN(ids)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		for _, bp := range blueprintWeapons {
			if bp.ID == id {
				bluePrintWeapons = append(bluePrintWeapons, server.BlueprintWeaponFromBoiler(bp))
			}
		}
	}

	return bluePrintWeapons, nil
}

func BlueprintWeapon(ids string) (*server.BlueprintWeapon, error) {
	blueprintWeapon, err := boiler.BlueprintWeapons(boiler.BlueprintWeaponWhere.ID.EQ(ids)).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.BlueprintWeaponFromBoiler(blueprintWeapon), nil
}
