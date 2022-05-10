package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func BlueprintEnergyCores(ids []string) ([]*server.BlueprintEnergyCore, error) {
	var bluePrintEnergyCores []*server.BlueprintEnergyCore
	blueprintEnergyCores, err := boiler.BlueprintEnergyCores(boiler.BlueprintEnergyCoreWhere.ID.IN(ids)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, bp := range blueprintEnergyCores {
		bluePrintEnergyCores = append(bluePrintEnergyCores, server.BlueprintEnergyCoreFromBoiler(bp))
	}

	return bluePrintEnergyCores, nil
}

func BlueprintEnergyCore(ids string) (*server.BlueprintEnergyCore, error) {
	blueprintEnergyCore, err := boiler.BlueprintEnergyCores(boiler.BlueprintEnergyCoreWhere.ID.EQ(ids)).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.BlueprintEnergyCoreFromBoiler(blueprintEnergyCore), nil
}
