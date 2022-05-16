package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertNewUtility(ownerID uuid.UUID, utility *server.BlueprintUtility) (*server.Utility, error) {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}

	// first insert the energy core
	newUtility := boiler.Utility{
		BrandID:               utility.BrandID,
		Label:                 utility.Label,
		BlueprintID:           utility.ID,
		GenesisTokenID:        utility.GenesisTokenID,
		LimitedReleaseTokenID: utility.LimitedReleaseTokenID,
		Type:                  utility.Type,
	}

	err = newUtility.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = InsertNewCollectionItem(tx, utility.Collection, boiler.ItemTypeUtility, newUtility.ID, utility.Tier, ownerID.String())
	if err != nil {
		return nil, terror.Error(err)
	}

	// insert the extra table depending on utility type
	switch newUtility.Type {
	case boiler.UtilityTypeSHIELD:
		if utility.ShieldBlueprint == nil {
			return nil, fmt.Errorf("utility type is shield but shield blueprint is nil")
		}
		newUtilityShield := boiler.UtilityShield{
			UtilityID:          newUtility.ID,
			Hitpoints:          utility.ShieldBlueprint.Hitpoints,
			RechargeRate:       utility.ShieldBlueprint.RechargeRate,
			RechargeEnergyCost: utility.ShieldBlueprint.RechargeEnergyCost,
		}
		err = newUtilityShield.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	case boiler.UtilityTypeATTACKDRONE:
		if utility.AttackDroneBlueprint == nil {
			return nil, fmt.Errorf("utility type is attack drone but attack drone blueprint is nil")
		}
		newAttackDrone := boiler.UtilityAttackDrone{
			UtilityID:        newUtility.ID,
			Damage:           utility.AttackDroneBlueprint.Damage,
			RateOfFire:       utility.AttackDroneBlueprint.RateOfFire,
			Hitpoints:        utility.AttackDroneBlueprint.Hitpoints,
			LifespanSeconds:  utility.AttackDroneBlueprint.LifespanSeconds,
			DeployEnergyCost: utility.AttackDroneBlueprint.DeployEnergyCost,
		}
		err = newAttackDrone.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	case boiler.UtilityTypeREPAIRDRONE:
		if utility.RepairDroneBlueprint == nil {
			return nil, fmt.Errorf("utility type is repair drone but repair drone blueprint is nil")
		}
		newRepairDrone := boiler.UtilityRepairDrone{
			UtilityID:        newUtility.ID,
			RepairType:       utility.RepairDroneBlueprint.RepairType,
			RepairAmount:     utility.RepairDroneBlueprint.RepairAmount,
			DeployEnergyCost: utility.RepairDroneBlueprint.DeployEnergyCost,
			LifespanSeconds:  utility.RepairDroneBlueprint.LifespanSeconds,
		}
		err = newRepairDrone.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	case boiler.UtilityTypeANTIMISSILE:
		if utility.AntiMissileBlueprint == nil {
			return nil, fmt.Errorf("utility type is anti missile but anti missile blueprint is nil")
		}
		newAntiMissile := boiler.UtilityAntiMissile{
			UtilityID:      newUtility.ID,
			RateOfFire:     utility.AntiMissileBlueprint.RateOfFire,
			FireEnergyCost: utility.AntiMissileBlueprint.FireEnergyCost,
		}
		err = newAntiMissile.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	case boiler.UtilityTypeACCELERATOR:
		if utility.AcceleratorBlueprint == nil {
			return nil, fmt.Errorf("utility type accelerator but accelerator blueprint is nil")
		}
		newAccelerator := boiler.UtilityAccelerator{
			UtilityID:    newUtility.ID,
			EnergyCost:   utility.AcceleratorBlueprint.EnergyCost,
			BoostSeconds: utility.AcceleratorBlueprint.BoostSeconds,
			BoostAmount:  utility.AcceleratorBlueprint.BoostAmount,
		}
		err = newAccelerator.Insert(tx, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	default:
		return nil, fmt.Errorf("invalid utility type")
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	utilityUUID, err := uuid.FromString(newUtility.ID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return Utility(utilityUUID)
}

func Utility(id uuid.UUID) (*server.Utility, error) {
	boilerUtility, err := boiler.FindUtility(gamedb.StdConn, id.String())
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id.String())).One(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	switch boilerUtility.Type {
	case boiler.UtilityTypeSHIELD:
		boilerShield, err := boiler.UtilityShields(boiler.UtilityShieldWhere.UtilityID.EQ(boilerUtility.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		return server.UtilityShieldFromBoiler(boilerUtility, boilerShield, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeATTACKDRONE:
		boilerAttackDrone, err := boiler.UtilityAttackDrones(boiler.UtilityAttackDroneWhere.UtilityID.EQ(boilerUtility.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		return server.UtilityAttackDroneFromBoiler(boilerUtility, boilerAttackDrone, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeREPAIRDRONE:
		boilerRepairDrone, err := boiler.UtilityRepairDrones(boiler.UtilityRepairDroneWhere.UtilityID.EQ(boilerUtility.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		return server.UtilityRepairDroneFromBoiler(boilerUtility, boilerRepairDrone, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeANTIMISSILE:
		boilerAntiMissile, err := boiler.UtilityAntiMissiles(boiler.UtilityAntiMissileWhere.UtilityID.EQ(boilerUtility.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		return server.UtilityAntiMissileFromBoiler(boilerUtility, boilerAntiMissile, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeACCELERATOR:
		boilerAccelerator, err := boiler.UtilityAccelerators(boiler.UtilityAcceleratorWhere.UtilityID.EQ(boilerUtility.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		return server.UtilityAcceleratorFromBoiler(boilerUtility, boilerAccelerator, boilerMechCollectionDetails), nil
	}

	return nil, fmt.Errorf("invalid utility type %s", boilerUtility.Type)
}
