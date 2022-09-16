package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func BlueprintUtilities(ids []string) ([]*server.BlueprintUtility, error) {
	var blueprintUtilities []*server.BlueprintUtility
	boilerBlueprintUtilities, err := boiler.BlueprintUtilities(boiler.BlueprintUtilityWhere.ID.IN(ids)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		for _, bp := range boilerBlueprintUtilities {
			if bp.ID == id {
				switch bp.Type {

				}
			}
		}
	}

	return blueprintUtilities, nil
}
