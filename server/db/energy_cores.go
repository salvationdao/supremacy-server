package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertNewEnergyCore(ownerID uuid.UUID, ec *server.BlueprintEnergyCore) (*server.EnergyCore, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}

	// first insert the energy core
	newEnergyCore := boiler.EnergyCore{
		OwnerID:      ownerID.String(),
		Label:        ec.Label,
		Size:         ec.Size,
		Capacity:     ec.Capacity,
		MaxDrawRate:  ec.MaxDrawRate,
		RechargeRate: ec.RechargeRate,
		Armour:       ec.Armour,
		MaxHitpoints: ec.MaxHitpoints,
		Tier:         ec.Tier,
	}

	err = newEnergyCore.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	//insert collection item
	collectionItem := boiler.CollectionItem{
		CollectionSlug: ec.Collection,
		ItemType:       boiler.ItemTypeEnergyCore,
		ItemID:         newEnergyCore.ID,
	}

	err = collectionItem.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	energyCoreUUID, err := uuid.FromString(newEnergyCore.ID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return EnergyCore(energyCoreUUID)
}

func EnergyCore(id uuid.UUID) (*server.EnergyCore, error) {
	boilerMech, err := boiler.FindEnergyCore(gamedb.StdConn, id.String())
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id.String())).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	return server.EnergyCoreFromBoiler(boilerMech, boilerMechCollectionDetails), nil
}
