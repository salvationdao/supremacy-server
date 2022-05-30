package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func BlueprintMechs(ids []string) ([]*server.BlueprintMech, error) {
	var bluePrintMechs []*server.BlueprintMech
	blueprintMechs, err := boiler.BlueprintMechs(boiler.BlueprintMechWhere.ID.IN(ids)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		for _, bp := range blueprintMechs {
			if bp.ID == id {
				bluePrintMechs = append(bluePrintMechs, server.BlueprintMechFromBoiler(bp))
			}
		}
	}

	return bluePrintMechs, nil
}

func BlueprintMech(ids string) (*server.BlueprintMech, error) {
	blueprintMech, err := boiler.BlueprintMechs(boiler.BlueprintMechWhere.ID.EQ(ids)).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.BlueprintMechFromBoiler(blueprintMech), nil
}
