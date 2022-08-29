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
				case boiler.UtilityTypeSHIELD:
					shield, err := boiler.BlueprintUtilityShields(boiler.BlueprintUtilityShieldWhere.BlueprintUtilityID.EQ(bp.ID)).One(gamedb.StdConn)
					if err != nil {
						return nil, err
					}
					blueprintUtilities = append(blueprintUtilities, server.BlueprintUtilityShieldFromBoiler(bp, shield))
				}
			}
		}
	}

	return blueprintUtilities, nil
}
