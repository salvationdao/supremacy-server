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
	
	return Utility(tx, newUtility.ID)
}

func Utility(tx boil.Executor, id string) (*server.Utility, error) {
	boilerUtility, err := boiler.Utilities(
		boiler.UtilityWhere.ID.EQ(id),
		qm.Load(boiler.UtilityRels.Blueprint),
		).One(tx)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	switch boilerUtility.Type {
	case boiler.UtilityTypeSHIELD:
		boilerShield, err := boiler.BlueprintUtilityShields(
			boiler.BlueprintUtilityShieldWhere.BlueprintUtilityID.EQ(boilerUtility.BlueprintID),

			).One(tx)
		if err != nil {
			return nil, err
		}
		return server.UtilityShieldFromBoiler(boilerUtility, boilerShield, boilerMechCollectionDetails), nil
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
			boilerShield, err := boiler.BlueprintUtilityShields(boiler.BlueprintUtilityShieldWhere.BlueprintUtilityID.EQ(util.BlueprintID)).One(gamedb.StdConn)
			if err != nil {
				return nil, err
			}
			utilities = append(utilities, server.UtilityShieldFromBoiler(util, boilerShield, boilerMechCollectionDetails))
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
