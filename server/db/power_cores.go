package db

import (
	"database/sql"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/volatiletech/null/v8"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func InsertNewPowerCore(tx boil.Executor, ownerID uuid.UUID, ec *server.BlueprintPowerCore) (*server.PowerCore, error) {
	newPowerCore := boiler.PowerCore{
		BlueprintID:           null.StringFrom(ec.ID),
		Label:                 ec.Label,
		Size:                  ec.Size,
		Capacity:              ec.Capacity,
		MaxDrawRate:           ec.MaxDrawRate,
		RechargeRate:          ec.RechargeRate,
		Armour:                ec.Armour,
		MaxHitpoints:          ec.MaxHitpoints,
		GenesisTokenID:        ec.GenesisTokenID,
		LimitedReleaseTokenID: ec.LimitedReleaseTokenID,
	}

	err := newPowerCore.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		ec.Collection,
		boiler.ItemTypePowerCore,
		newPowerCore.ID,
		ec.Tier,
		ownerID.String(),
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return PowerCore(tx, newPowerCore.ID)
}

func PowerCore(tx boil.Executor, id string) (*server.PowerCore, error) {
	boilerMech, err := boiler.PowerCores(
		boiler.PowerCoreWhere.ID.EQ(id),
		qm.Load(boiler.PowerCoreRels.Blueprint),
		).One(tx)
	if err != nil {
		return nil, err
	}
	boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	return server.PowerCoreFromBoiler(boilerMech, boilerMechCollectionDetails), nil
}

func PowerCores(id ...string) ([]*server.PowerCore, error) {
	var powerCores []*server.PowerCore
	boilerPowerCores, err := boiler.PowerCores(boiler.PowerCoreWhere.ID.IN(id)).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}
	for _, pc := range boilerPowerCores {
		boilerPowerCoreCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(pc.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}
		powerCores = append(powerCores, server.PowerCoreFromBoiler(pc, boilerPowerCoreCollectionDetails))
	}

	return powerCores, nil
}

// AttachPowerCoreToMech attaches a power core to a mech  TODO: create tests.
func AttachPowerCoreToMech(trx *sql.Tx, ownerID, mechID, powerCoreID string) error {
	// TODO: possible optimize this, 6 queries to attach a part seems like a lot?
	// check owner
	tx := trx
	var err error
	if trx == nil {
		tx, err = gamedb.StdConn.Begin()
		if err != nil {
			gamelog.L.Error().Err(err).Str("mech.ID", mechID).Str("powercore ID", powerCoreID).Msg("failed to equip powercore to mech, issue creating tx")
			return terror.Error(err, "Issue preventing equipping this powercore to the war machine, try again or contact support.")
		}
		defer tx.Rollback()
	}

	mechCI, err := CollectionItemFromItemID(tx, mechID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to mech collection item")
		return terror.Error(err)
	}
	pcCI, err := CollectionItemFromItemID(tx, powerCoreID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("powerCoreID", powerCoreID).Msg("failed to power core collection item")
		return terror.Error(err)
	}

	if mechCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("mechCI.OwnerID", mechCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the war machine to equip power cores to it.")

	}
	if pcCI.OwnerID != ownerID {
		err := fmt.Errorf("owner id mismatch")
		gamelog.L.Error().Err(err).Str("pcCI.OwnerID", pcCI.OwnerID).Str("ownerID", ownerID).Msg("user doesn't own the item")
		return terror.Error(err, "You need to be the owner of the power core to equip it to a war machine.")
	}

	// get mech
	mech, err := boiler.FindMech(tx, mechID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("mechID", mechID).Msg("failed to find mech")
		return terror.Error(err)
	}

	// get power core
	powerCore, err := boiler.FindPowerCore(tx, powerCoreID)
	if err != nil {
		gamelog.L.Error().Err(err).Str("powerCoreID", powerCoreID).Msg("failed to find power core")
		return terror.Error(err)
	}

	// wrong size
	if mech.PowerCoreSize != powerCore.Size {
		err := fmt.Errorf("powercore size mismatch")
		gamelog.L.Error().Err(err).Str("mech.PowerCoreSize", mech.PowerCoreSize).Str("powerCore.Size", powerCore.Size).Msg("this powercore doesn't fit")
		return terror.Error(err, "This power core doesn't fit this war machine.")
	}

	// error out, already has a power core
	if mech.PowerCoreID.Valid && mech.PowerCoreID.String != "" {
		err := fmt.Errorf("warmachine already has a power core")
		// also check powerCore.EquippedOn on, if that doesn't match, update it, so it does.
		if !powerCore.EquippedOn.Valid {
			powerCore.EquippedOn = null.StringFrom(mech.ID)
			_, err = powerCore.Update(tx, boil.Infer())
			if err != nil {
				gamelog.L.Error().Err(err).Str("mech.ID", mech.ID).Str("powerCore.ID", powerCore.ID).Msg("failed to update power core equipped on")
				return terror.Error(err, "War machine already has a power core.")
			}
		}
		gamelog.L.Error().Err(err).Str("mech.PowerCoreID.String", mech.PowerCoreID.String).Str("new powerCore.ID", powerCore.ID).Msg(err.Error())
		return terror.Error(err, "War machine already has a power core.")
	}

	// lets join
	mech.PowerCoreID = null.StringFrom(powerCore.ID)
	powerCore.EquippedOn = null.StringFrom(mech.ID)

	_, err = mech.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech.PowerCoreID.String", mech.PowerCoreID.String).Str("new powerCore.ID", powerCore.ID).Msg("failed to equip power core to mech, issue mech update")
		return terror.Error(err, "Issue preventing equipping this power core to the war machine, try again or contact support.")
	}
	_, err = powerCore.Update(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Str("mech.PowerCoreID.String", mech.PowerCoreID.String).Str("new powerCore.ID", powerCore.ID).Msg("failed to equip power core to mech, issue power core update")
		return terror.Error(err, "Issue preventing equipping this power core to the war machine, try again or contact support.")
	}

	if trx == nil {
		err = tx.Commit()
		if err != nil {
			gamelog.L.Error().Err(err).Str("mech.PowerCoreID.String", mech.PowerCoreID.String).Str("new powerCore.ID", powerCore.ID).Msg("failed to equip power core to mech, issue committing tx")
			return terror.Error(err, "Issue preventing equipping this power core to the war machine, try again or contact support.")
		}
	}

	return nil
}
