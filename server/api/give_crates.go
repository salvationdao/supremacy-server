package api

import (
	"fmt"
	"net/http"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func WithDev(next func(w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		// if os.Getenv("GAMESERVER_ENVIRONMENT") != "development" {
		// 	return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		// }
		devPass := r.Header.Get("X-Authorization")
		if devPass != "NinjaDojo_!" {
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}

		return next(w, r)
	}
	return fn
}

func (api *API) DevGiveCrates(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := common.HexToAddress(chi.URLParam(r, "public_address"))
	user, err := boiler.Players(boiler.PlayerWhere.PublicAddress.EQ(null.StringFrom(publicAddress.String()))).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player by pub address")

		return http.StatusInternalServerError, err
	}

	// purchase crate
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("start tx2: %w", err)
	}
	defer tx.Rollback()

	storeCrate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.MysteryCrateType.EQ("WEAPON"),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(user.FactionID.String),
		qm.Load(boiler.StorefrontMysteryCrateRels.Faction),
	).One(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get crate for purchase, please try again or contact support.")
	}

	assignedCrate, err := assignAndRegisterPurchasedCrate(user.ID, storeCrate, tx, api)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
	}

	err = tx.Commit()
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Could not open mystery crate, please try again or contact support.")
	}

	// open crate
	tx2, err := gamedb.StdConn.Begin()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("start tx2: %w", err)
	}
	defer tx2.Rollback()
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(assignedCrate.ItemID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMysteryCrate),
	).One(tx2)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Could not find collection item, try again or contact support.")
	}

	// update mystery crate
	crate := boiler.MysteryCrate{}
	q := `
	UPDATE mystery_crate
	SET opened = TRUE
	WHERE id = $1 AND opened = FALSE 
	RETURNING id, type, faction_id, label, opened, locked_until, purchased, deleted_at, updated_at, created_at, description`
	err = gamedb.StdConn.
		QueryRow(q, collectionItem.ItemID).
		Scan(
			&crate.ID,
			&crate.Type,
			&crate.FactionID,
			&crate.Label,
			&crate.Opened,
			&crate.LockedUntil,
			&crate.Purchased,
			&crate.DeletedAt,
			&crate.UpdatedAt,
			&crate.CreatedAt,
			&crate.Description,
		)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Could not find crate, try again or contact support.")
	}

	items := OpenCrateResponse{}

	blueprintItems, err := crate.MysteryCrateBlueprints().All(tx2)
	if err != nil {
		gamelog.L.Error().Err(err).Msg(fmt.Sprintf("failed to get blueprint relationships from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
		return http.StatusInternalServerError, terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
	}

	for _, blueprintItem := range blueprintItems {
		switch blueprintItem.BlueprintType {
		case boiler.TemplateItemTypeMECH:
			bp, err := db.BlueprintMech(blueprintItem.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get mech blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
			}

			mech, err := db.InsertNewMech(tx2, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new mech from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
			}
			items.Mech = mech
		case boiler.TemplateItemTypeWEAPON:
			bp, err := db.BlueprintWeapon(blueprintItem.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get weapon blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get weapon blueprint during crate opening, try again or contact support.")
			}

			weapon, err := db.InsertNewWeapon(tx2, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new weapon from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get weapon during crate opening, try again or contact support.")
			}
			items.Weapons = append(items.Weapons, weapon)
		case boiler.TemplateItemTypeMECH_SKIN:
			bp, err := db.BlueprintMechSkinSkin(blueprintItem.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get mech skin blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get mech skin blueprint during crate opening, try again or contact support.")
			}

			mechSkin, err := db.InsertNewMechSkin(tx2, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new mech skin from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get mech skin during crate opening, try again or contact support.")
			}
			items.MechSkin = mechSkin
		case boiler.TemplateItemTypeWEAPON_SKIN:
			bp, err := db.BlueprintWeaponSkin(blueprintItem.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get weapon skin blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get weapon skin blueprint during crate opening, try again or contact support.")
			}
			weaponSkin, err := db.InsertNewWeaponSkin(tx2, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new weapon skin from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get weapon skin during crate opening, try again or contact support.")
			}
			items.WeaponSkin = weaponSkin
		case boiler.TemplateItemTypePOWER_CORE:
			bp, err := db.BlueprintPowerCore(blueprintItem.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get powercore blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get powercore blueprint during crate opening, try again or contact support.")
			}

			powerCore, err := db.InsertNewPowerCore(tx2, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new powercore from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get powercore during crate opening, try again or contact support.")
			}
			items.PowerCore = powerCore
		}
	}

	if crate.Type == boiler.CrateTypeMECH {
		//attach mech_skin to mech - mech
		err = db.AttachMechSkinToMech(tx2, user.ID, items.Mech.ID, items.MechSkin.ID, false)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach mech skin to mech during CRATE:OPEN crate: %s", crate.ID))
			return http.StatusInternalServerError, terror.Error(err, "Could not open crate, try again or contact support.")
		}

		err = db.AttachPowerCoreToMech(tx2, user.ID, items.Mech.ID, items.PowerCore.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach powercore to mech during CRATE:OPEN crate: %s", crate.ID))
			return http.StatusInternalServerError, terror.Error(err, "Could not open crate, try again or contact support.")
		}

		//attach weapons to mech -mech
		for _, weapon := range items.Weapons {
			err = db.AttachWeaponToMech(tx2, user.ID, items.Mech.ID, weapon.ID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach weapons to mech during CRATE:OPEN crate: %s", crate.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not open crate, try again or contact support.")
			}
		}

		mech, err := db.Mech(tx2, items.Mech.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to get final mech during CRATE:OPEN crate: %s", crate.ID))
			return http.StatusInternalServerError, terror.Error(err, "Could not open crate, try again or contact support.")
		}
		items.Mech = mech
	}

	if crate.Type == boiler.CrateTypeWEAPON {
		//attach weapon_skin to weapon -weapon
		if len(items.Weapons) != 1 {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("too many weapons in crate: %s", crate.ID))
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("too many weapons in weapon crate"), "Could not open crate, try again or contact support.")
		}
		err = db.AttachWeaponSkinToWeapon(tx2, user.ID, items.Weapons[0].ID, items.WeaponSkin.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach weapon skin to weapon during CRATE:OPEN crate: %s", crate.ID))
			return http.StatusInternalServerError, terror.Error(err, "Could not open crate, try again or contact support.")
		}
	}

	err = tx2.Commit()
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Could not open mystery crate, please try again or contact support.")
	}

	return http.StatusOK, nil
}
