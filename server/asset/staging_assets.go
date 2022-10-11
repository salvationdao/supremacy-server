package asset

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpctypes"
	"server/xsyn_rpcclient"
)

// GiveUserAllAssets gives a user all weapons, skins and mechs, ONLY TO BE USERS IN DEV/STAGING
func GiveUserAllAssets(user *boiler.Player, pp *xsyn_rpcclient.XsynXrpcClient) error {
	L := gamelog.L.With().Str("func", "GiveUserAllAssets").Str("user id", user.ID).Logger()
	if server.IsProductionEnv() {
		err := fmt.Errorf("invalid environment")
		L.Error().Err(err).Msg("failed to assign assets")
		return err
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		L.Error().Err(err).Msg("failed to assign assets")
		return err
	}

	defer tx.Rollback()

	// give one of each faction mega genesis mechs
	templateIDS := []string{}
	templates, err := boiler.TemplateBlueprints(
		boiler.TemplateBlueprintWhere.BlueprintID.IN(
			[]string{
				"65159dae-d349-4b7a-a6ca-81c69d1ae843", // static id for genesis mega skin
				"1ab33cab-a76b-4557-b672-ba54545bdf6b", // static id for genesis mega skin
				"123e2d2c-ffa5-4649-9641-c2cafac64c85", // static id for genesis mega skin
			},
		),
	).All(tx)
	if err != nil {
		L.Error().Err(err).Msg("failed to get genesis templates")
		return err
	}

	for i := 0; i < 3; i++ {
		for _, tmpl := range templates {
			templateIDS = append(templateIDS, tmpl.TemplateID)
		}
	}

	err = pp.AssignTemplateToUser(&xsyn_rpcclient.AssignTemplateReq{
		TemplateIDs: templateIDS,
		UserID:      user.ID,
	})
	if err != nil {
		L.Error().Err(err).Msg("failed to assign templates to user")
		return terror.Error(err, "Failed to sync passport db")
	}

	xsynAssets := []*rpctypes.XsynAsset{}
	insertedMechs := []*server.Mech{}
	insertedMechSkins := []*server.MechSkin{}
	insertedPowerCores := []*server.PowerCore{}
	insertedWeapons := []*server.Weapon{}
	insertedWeaponSkins := []*server.WeaponSkin{}

	// Now we want to give 1 of each nexus mech
	// need mech + skin
	mechs, err := boiler.BlueprintMechs(
		boiler.BlueprintMechWhere.ShieldTypeID.EQ("b209a593-f0a0-4df9-bc8d-f6af8bab4362"), // static id nexus mech shield
		qm.Load(boiler.BlueprintMechRels.DefaultChassisSkin),
	).All(tx)
	if err != nil {
		L.Error().Err(err).Msg("failed to get blueprint mechs")
		return err
	}
	for _, mech := range mechs {
		for i := 0; i < 3; i ++ { // insert 3 of each mech
			insertedMech, insertedMechSkin, err := db.InsertNewMechAndSkin(
				tx,
				uuid.FromStringOrNil(user.ID),
				server.BlueprintMechFromBoiler(mech),
				server.BlueprintMechSkinFromBoiler(mech.R.DefaultChassisSkin),
			)
			if err != nil {
				return err
			}
			insertedMechs = append(insertedMechs, insertedMech)
			insertedMechSkins = append(insertedMechSkins, insertedMechSkin)
		}
	}

	// also insert power cores for mechs
	powerCores, err := boiler.BlueprintPowerCores().All(tx)
	if err != nil {
		L.Error().Err(err).Msg("failed to get blueprint powercores")
		return err
	}

	for _, powerCore := range powerCores {
		for i := 0; i <= len(insertedMechs); i++ { // they'll likely have a bunch of spares
			powerCore, err := db.InsertNewPowerCore(tx, uuid.FromStringOrNil(user.ID), server.BlueprintPowerCoreFromBoiler(powerCore))
			if err != nil {
				L.Error().Err(err).Msg("failed to insert powercores")
				return err
			}
			insertedPowerCores = append(insertedPowerCores, powerCore)
		}
	}

	// Now we want to give 4! of each weapon
	weapons, err := boiler.BlueprintWeapons(
		boiler.BlueprintWeaponWhere.ID.NIN(
			[]string{
				"c1c78867-9de7-43d3-97e9-91381800f38e", // don't want to give rocket pods since they're locked to genesis
				"41099781-8586-4783-9d1c-b515a386fe9f", // don't want to give rocket pods since they're locked to genesis
				"e9fc2417-6a5b-489d-b82e-42942535af90", // don't want to give rocket pods since they're locked to genesis
			},
		),
		qm.Load(boiler.BlueprintWeaponRels.DefaultSkin),
	).All(tx)
	if err != nil {
		L.Error().Err(err).Msg("failed to get blueprint weapons")
		return err
	}
	for _, weapon := range weapons {
		for i := 0; i < 4; i++ { // four hops this time
			insertedWeapon, insertedWeaponSkin, err := db.InsertNewWeapon(
				tx,
				uuid.FromStringOrNil(user.ID),
				server.BlueprintWeaponFromBoiler(weapon),
				server.BlueprintWeaponSkinFromBoiler(weapon.R.DefaultSkin),
			)
			if err != nil {
				L.Error().Err(err).Msg("failed to insert blueprint weapons")
				return err
			}
			insertedWeapons = append(insertedWeapons, insertedWeapon)
			insertedWeaponSkins = append(insertedWeaponSkins, insertedWeaponSkin)
		}
	}

	// now give them one of each weapon skin?
	weaponSkins, err := boiler.BlueprintWeaponSkins().All(tx)
	if err != nil {
		L.Error().Err(err).Msg("failed to get blueprint weapon skins")
		return err
	}
	for _, weaponSkin := range weaponSkins {
		insertedWeaponSkin, err := db.InsertNewWeaponSkin(tx, uuid.FromStringOrNil(user.ID), server.BlueprintWeaponSkinFromBoiler(weaponSkin), nil)
		if err != nil {
			L.Error().Err(err).Msg("failed to insert blueprint weapons")
			return err
		}
		insertedWeaponSkins = append(insertedWeaponSkins, insertedWeaponSkin)
	}

	// now give them one of each mech skin?
	mechSkins, err := boiler.BlueprintMechSkins().All(tx)
	if err != nil {
		L.Error().Err(err).Msg("failed to get blueprint mech skins")
		return err
	}
	for _, mechSkin := range mechSkins {
		insertedMechSkin, err := db.InsertNewMechSkin(tx, uuid.FromStringOrNil(user.ID), server.BlueprintMechSkinFromBoiler(mechSkin), nil)
		if err != nil {
			L.Error().Err(err).Msg("failed to insert blueprint mech skins")
			return err
		}
		insertedMechSkins = append(insertedMechSkins, insertedMechSkin)
	}

	// now we register them on xsyn
	xsynAssets = append(xsynAssets, rpctypes.ServerMechsToXsynAsset(insertedMechs)...)
	xsynAssets = append(xsynAssets, rpctypes.ServerMechSkinsToXsynAsset(tx, insertedMechSkins)...)
	xsynAssets = append(xsynAssets, rpctypes.ServerPowerCoresToXsynAsset(insertedPowerCores)...)
	xsynAssets = append(xsynAssets, rpctypes.ServerWeaponsToXsynAsset(insertedWeapons)...)
	xsynAssets = append(xsynAssets, rpctypes.ServerWeaServerWeaponSkinsToXsynAsset(tx, insertedWeaponSkins)...)

	err = pp.AssetsRegister(xsynAssets) // register new assets
	if err != nil {
		L.Error().Err(err).Msg("failed to register assets on xsyn")
		return err
	}

	err = tx.Commit()
	if err != nil {
		L.Error().Err(err).Msg("failed to commit when assigning assets")
		return err
	}

	return nil
}
