package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"
)

func BlueprintPowerCores(ids []string) ([]*server.BlueprintPowerCore, error) {
	var bluePrintPowerCores []*server.BlueprintPowerCore
	blueprintPowerCores, err := boiler.BlueprintPowerCores(boiler.BlueprintPowerCoreWhere.ID.IN(ids)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, bp := range blueprintPowerCores {
		bluePrintPowerCores = append(bluePrintPowerCores, server.BlueprintPowerCoreFromBoiler(bp))
	}

	return bluePrintPowerCores, nil
}

func BlueprintPowerCore(ids string) (*server.BlueprintPowerCore, error) {
	blueprintPowerCore, err := boiler.BlueprintPowerCores(boiler.BlueprintPowerCoreWhere.ID.EQ(ids)).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.BlueprintPowerCoreFromBoiler(blueprintPowerCore), nil
}
