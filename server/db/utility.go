package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertNewUtility(tx boil.Executor, ownerID uuid.UUID, utility *server.BlueprintUtility) (*server.Utility, error) {
	newUtility := boiler.Utility{
		BrandID:               utility.BrandID,
		Label:                 utility.Label,
		BlueprintID:           utility.ID,
		GenesisTokenID:        utility.GenesisTokenID,
		LimitedReleaseTokenID: utility.LimitedReleaseTokenID,
		Type:                  utility.Type,
	}

	err := newUtility.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		utility.Collection,
		boiler.ItemTypeUtility,
		newUtility.ID,
		utility.Tier,
		ownerID.String(),
	)
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
	
	return Utility(tx, newUtility.ID)
}

func Utility(tx boil.Executor, id string) (*server.Utility, error) {
	boilerUtility, err := boiler.FindUtility(tx, id)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	switch boilerUtility.Type {
	case boiler.UtilityTypeSHIELD:
		boilerShield, err := boiler.UtilityShields(boiler.UtilityShieldWhere.UtilityID.EQ(boilerUtility.ID)).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityShieldFromBoiler(boilerUtility, boilerShield, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeATTACKDRONE:
		boilerAttackDrone, err := boiler.UtilityAttackDrones(boiler.UtilityAttackDroneWhere.UtilityID.EQ(boilerUtility.ID)).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityAttackDroneFromBoiler(boilerUtility, boilerAttackDrone, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeREPAIRDRONE:
		boilerRepairDrone, err := boiler.UtilityRepairDrones(boiler.UtilityRepairDroneWhere.UtilityID.EQ(boilerUtility.ID)).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityRepairDroneFromBoiler(boilerUtility, boilerRepairDrone, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeANTIMISSILE:
		boilerAntiMissile, err := boiler.UtilityAntiMissiles(boiler.UtilityAntiMissileWhere.UtilityID.EQ(boilerUtility.ID)).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityAntiMissileFromBoiler(boilerUtility, boilerAntiMissile, boilerMechCollectionDetails), nil
	case boiler.UtilityTypeACCELERATOR:
		boilerAccelerator, err := boiler.UtilityAccelerators(boiler.UtilityAcceleratorWhere.UtilityID.EQ(boilerUtility.ID)).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityAcceleratorFromBoiler(boilerUtility, boilerAccelerator, boilerMechCollectionDetails), nil
	}

	return nil, fmt.Errorf("invalid utility type %s", boilerUtility.Type)
}

func Utilities(id ...string) ([]*server.Utility, error) {
	var utilities []*server.Utility
	boilerUtilities, err := boiler.Utilities(boiler.UtilityWhere.ID.IN(id)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, util := range boilerUtilities {
		boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(util.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}

		switch util.Type {
		case boiler.UtilityTypeSHIELD:
			boilerShield, err := boiler.UtilityShields(boiler.UtilityShieldWhere.UtilityID.EQ(util.ID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityShieldFromBoiler(util, boilerShield, boilerMechCollectionDetails))
		case boiler.UtilityTypeATTACKDRONE:
			boilerAttackDrone, err := boiler.UtilityAttackDrones(boiler.UtilityAttackDroneWhere.UtilityID.EQ(util.ID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityAttackDroneFromBoiler(util, boilerAttackDrone, boilerMechCollectionDetails))
		case boiler.UtilityTypeREPAIRDRONE:
			boilerRepairDrone, err := boiler.UtilityRepairDrones(boiler.UtilityRepairDroneWhere.UtilityID.EQ(util.ID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityRepairDroneFromBoiler(util, boilerRepairDrone, boilerMechCollectionDetails))
		case boiler.UtilityTypeANTIMISSILE:
			boilerAntiMissile, err := boiler.UtilityAntiMissiles(boiler.UtilityAntiMissileWhere.UtilityID.EQ(util.ID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityAntiMissileFromBoiler(util, boilerAntiMissile, boilerMechCollectionDetails))
		case boiler.UtilityTypeACCELERATOR:
			boilerAccelerator, err := boiler.UtilityAccelerators(boiler.UtilityAcceleratorWhere.UtilityID.EQ(util.ID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityAcceleratorFromBoiler(util, boilerAccelerator, boilerMechCollectionDetails))
		}
	}
	return utilities, nil
}

// AttachUtilityToMech attaches a Utility to a mech
// If lockedToMech == true utility cannot be removed from mech ever (used for genesis and limited mechs)
func AttachUtilityToMech(tx boil.Executor, ownerID, mechID, utilityID string, lockedToMech bool) error {
	// TODO: possible optimize this, 6 queries to attach a part seems like a lot?
	mechCI, err := CollectionItemFromItemID(tx, mechID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to get mech collection item")
		return terror.Error(err)
	}
	utilityCI, err := CollectionItemFromItemID(tx, utilityID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("utilityID", utilityID).Msg("failed to get utility collection item")
		return terror.Error(err)
	}

	if mechCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("mechCI.OwnerID", mechCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the war machine to equip utilitys to it.")
	}
	if utilityCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("utilityCI.OwnerID", utilityCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the utility to equip it to a war machine.")
	}

	// get mech
	mech, err := boiler.Mechs(
		boiler.MechWhere.ID.EQ(mechID),
		qm.Load(boiler.MechRels.ChassisMechUtilities),
		qm.Load(boiler.MechRels.Blueprint),
	).One(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to find mech")
		return terror.Error(err)
	}

	// get Utility
	utility, err := boiler.FindUtility(tx, utilityID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("utilityID", utilityID).Msg("failed to find Utility")
		return terror.Error(err)
	}

	// check current utility count
	if len(mech.R.ChassisMechUtilities)+1 > mech.R.Blueprint.UtilitySlots {
		err := fmt.Errorf("utility cannot fit")
		gamelog.L.Error().Err(err).Str("utilityID", utilityID).Msg("adding this utility brings mechs utilities over mechs utility slots")
		return terror.Error(err, fmt.Sprintf("War machine already has %d utilities equipped and is only has %d utility slots.", len(mech.R.ChassisMechUtilities), mech.R.Blueprint.UtilitySlots))
	}

	// check utility isn't already equipped to another war machine
	exists, err := boiler.MechUtilities(boiler.MechUtilityWhere.UtilityID.EQ(utilityID)).Exists(tx)
	if err != nil {
		gamelog.L.Error().Err(err).Str("utilityID", utilityID).Msg("failed to check if a mech and utility join already exists")
		return terror.Error(err)
	}
	if exists {
		err := fmt.Errorf("utility already equipped to a warmachine")
		gamelog.L.Error().Err(err).Str("utilityID", utilityID).Msg(err.Error())
		return terror.Error(err, "This utility is already equipped to another war machine, try again or contact support.")
	}

	utility.EquippedOn = null.StringFrom(mech.ID)
	utility.LockedToMech = lockedToMech

	_, err = utility.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("utility", utility).Msg("failed to update utility")
		return terror.Error(err, "Issue preventing equipping this utility to the war machine, try again or contact support.")
	}

	utilityMechJoin := boiler.MechUtility{
		ChassisID:  mech.ID,
		UtilityID:  utility.ID,
		SlotNumber: len(mech.R.ChassisMechUtilities), // slot number starts at 0, so if we currently have 2 equipped and this is the 3rd, it will be slot 2.
	}

	err = utilityMechJoin.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("utilityMechJoin", utilityMechJoin).Msg(" failed to equip utility to war machine")
		return terror.Error(err, "Issue preventing equipping this utility to the war machine, try again or contact support.")
	}

	return nil
}
