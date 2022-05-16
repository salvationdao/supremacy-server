package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertNewPowerCore(ownerID uuid.UUID, ec *server.BlueprintPowerCore) (*server.PowerCore, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}

	// first insert the energy core
	newPowerCore := boiler.PowerCore{
		Label:        ec.Label,
		Size:         ec.Size,
		Capacity:     ec.Capacity,
		MaxDrawRate:  ec.MaxDrawRate,
		RechargeRate: ec.RechargeRate,
		Armour:       ec.Armour,
		MaxHitpoints: ec.MaxHitpoints,
	}

	err = newPowerCore.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = InsertNewCollectionItem(tx, ec.Collection, boiler.ItemTypePowerCore, newPowerCore.ID, ec.Tier, ownerID.String())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	powerCoreUUID, err := uuid.FromString(newPowerCore.ID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return PowerCore(powerCoreUUID)
}

func PowerCore(id uuid.UUID) (*server.PowerCore, error) {
	boilerMech, err := boiler.FindPowerCore(gamedb.StdConn, id.String())
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id.String())).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.PowerCoreFromBoiler(boilerMech, boilerMechCollectionDetails), nil
}
