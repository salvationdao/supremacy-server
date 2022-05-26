package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func Template(templateID uuid.UUID) (*server.TemplateContainer, error) {
	template, err := boiler.Templates(
		boiler.TemplateWhere.ID.EQ(templateID.String()),
		qm.Load(boiler.TemplateRels.TemplateBlueprints),
	).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}

	result := &server.TemplateContainer{
		ID:                          template.ID,
		Label:                       template.Label,
		UpdatedAt:                   template.UpdatedAt,
		CreatedAt:                   template.CreatedAt,
		IsLimitedRelease:            template.IsLimitedRelease,
		IsGenesis:                   template.IsGenesis,
		ContainsCompleteMechExactly: template.ContainsCompleteMechExactly,
		BlueprintMech:               []*server.BlueprintMech{},
		BlueprintWeapon:             []*server.BlueprintWeapon{},
		BlueprintUtility:            []*server.BlueprintUtility{},
		BlueprintMechSkin:           []*server.BlueprintMechSkin{},
		BlueprintMechAnimation:      []*server.BlueprintMechAnimation{},
		BlueprintPowerCore:          []*server.BlueprintPowerCore{},
	}

	// filter them into IDs first to optimize db queries
	var blueprintMechIDS []string
	var blueprintWeaponIDS []string
	var blueprintUtilityIDS []string
	var blueprintMechSkinIDS []string
	var blueprintMechAnimationIDS []string
	var blueprintPowerCoreIDS []string

	// filter out to ids
	for _, bp := range template.R.TemplateBlueprints {
		switch bp.Type {
		case boiler.TemplateItemTypeMECH:
			blueprintMechIDS = append(blueprintMechIDS, bp.BlueprintID)
		case boiler.TemplateItemTypeMECH_ANIMATION:
			blueprintMechAnimationIDS = append(blueprintMechAnimationIDS, bp.BlueprintID)
		case boiler.TemplateItemTypeMECH_SKIN:
			blueprintMechSkinIDS = append(blueprintMechSkinIDS, bp.BlueprintID)
		case boiler.TemplateItemTypeUTILITY:
			blueprintUtilityIDS = append(blueprintUtilityIDS, bp.BlueprintID)
		case boiler.TemplateItemTypeWEAPON:
			blueprintWeaponIDS = append(blueprintWeaponIDS, bp.BlueprintID)
		case boiler.TemplateItemTypePOWER_CORE:
			blueprintPowerCoreIDS = append(blueprintPowerCoreIDS, bp.BlueprintID)
		case boiler.TemplateItemTypeWEAPON_SKIN:
			continue
		case boiler.TemplateItemTypePLAYER_ABILITY:
			continue
		case boiler.TemplateItemTypeAMMO:
			continue
		// TODO: AMMO  //				blueprintMechAnimationIDS = append(blueprintMechAnimationIDS, bp.BlueprintID)
		default:
			return nil, terror.Error(fmt.Errorf("invalid template item type %s", bp.Type))
		}
	}

	// get the objects
	result.BlueprintMech, err = BlueprintMechs(blueprintMechIDS)
	if err != nil {
		return nil, terror.Error(err)
	}
	result.BlueprintWeapon, err = BlueprintWeapons(blueprintWeaponIDS)
	if err != nil {
		return nil, terror.Error(err)
	}
	result.BlueprintMechSkin, err = BlueprintMechSkinSkins(blueprintMechSkinIDS)
	if err != nil {
		return nil, terror.Error(err)
	}
	result.BlueprintMechAnimation, err = BlueprintMechAnimations(blueprintMechAnimationIDS)
	if err != nil {
		return nil, terror.Error(err)
	}
	result.BlueprintPowerCore, err = BlueprintPowerCores(blueprintPowerCoreIDS)
	if err != nil {
		return nil, terror.Error(err)
	}
	result.BlueprintUtility, err = BlueprintUtilities(blueprintUtilityIDS)
	if err != nil {
		return nil, terror.Error(err)
	}
	return result, nil
}

// TODO: I want TemplateRegister tested.

// TemplateRegister copies everything out of a template into a new mech
func TemplateRegister(templateID uuid.UUID, ownerID uuid.UUID) (
	[]*server.Mech,
	[]*server.MechAnimation,
	[]*server.MechSkin,
	[]*server.PowerCore,
	[]*server.Weapon,
	[]*server.Utility,
	error) {
	var mechs []*server.Mech
	var mechAnimations []*server.MechAnimation
	var mechSkins []*server.MechSkin
	var powerCores []*server.PowerCore
	var weapons []*server.Weapon
	var utilities []*server.Utility

	exists, err := boiler.PlayerExists(gamedb.StdConn, ownerID.String())
	if err != nil {
		return mechs,
			mechAnimations,
			mechSkins,
			powerCores,
			weapons,
			utilities,
			fmt.Errorf("check player exists: %w", err)
	}
	if !exists {
		newPlayer := &boiler.Player{ID: ownerID.String()}
		err = newPlayer.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return mechs,
				mechAnimations,
				mechSkins,
				powerCores,
				weapons,
				utilities, fmt.Errorf("insert new player: %w", err)
		}
	}

	tmpl, err := Template(templateID)
	if err != nil {
		return mechs,
			mechAnimations,
			mechSkins,
			powerCores,
			weapons,
			utilities, fmt.Errorf("find template: %w", err)
	}

	// do these checks here because we can't return error after we assign the items
	if tmpl.ContainsCompleteMechExactly {
		if len(tmpl.BlueprintMech) != 1 {
			return mechs,
				mechAnimations,
				mechSkins,
				powerCores,
				weapons,
				utilities, fmt.Errorf("template contains complete mech exactly but has more than 1 mech")
		}
		if len(tmpl.BlueprintPowerCore) != 1 {
			return mechs,
				mechAnimations,
				mechSkins,
				powerCores,
				weapons,
				utilities, fmt.Errorf("template contains complete mech exactly but has wrong amount of power cores")
		}
		if len(tmpl.BlueprintWeapon) > tmpl.BlueprintMech[0].WeaponHardpoints {
			return mechs,
				mechAnimations,
				mechSkins,
				powerCores,
				weapons,
				utilities, fmt.Errorf("template contains complete mech exactly but contained mech has less weapon slots than weapons provided")
		}
		if len(tmpl.BlueprintUtility) > tmpl.BlueprintMech[0].UtilitySlots {
			return mechs,
				mechAnimations,
				mechSkins,
				powerCores,
				weapons,
				utilities, fmt.Errorf("template contains complete mech exactly but contained mech has less utility slots than utilities provided")
		}
		if len(tmpl.BlueprintMechSkin) > 1 {
			return mechs,
				mechAnimations,
				mechSkins,
				powerCores,
				weapons,
				utilities, fmt.Errorf("template contains complete mech exactly but contained more than a single skin")
		}
	}

	tokenIDs := &struct {
		GenesisTokenID null.Int64 `json:"genesis_token_id" db:"genesis_token_id"`
		LimitedTokenID null.Int64 `json:"limited_token_id" db:"limited_token_id"`
	}{}

	// if template is genesis, create it a genesis ID
	if tmpl.IsGenesis {
		// get the max genesis
		err := boiler.NewQuery(qm.SQL(`SELECT coalesce(max(genesis_token_id) + 1, 0) as genesis_token_id FROM mechs`)).Bind(nil, gamedb.StdConn, tokenIDs)
		if err != nil {
			fmt.Println("errror!")
			fmt.Println(err.Error())
			gamelog.L.Error().Err(err).Msg("failed to get new genesis token id")
			return mechs,
				mechAnimations,
				mechSkins,
				powerCores,
				weapons,
				utilities, terror.Error(err)
		}
	}
	// if template is genesis, create it a genesis ID
	if tmpl.IsLimitedRelease {
		// get the max limited
		err := boiler.NewQuery(qm.SQL(`SELECT coalesce(max(limited_release_token_id) + 1, 0) as limited_token_id FROM mechs`)).Bind(nil, gamedb.StdConn, tokenIDs)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("failed to get new limit release token id")
			return mechs,
				mechAnimations,
				mechSkins,
				powerCores,
				weapons,
				utilities, terror.Error(err)
		}
	}

	// inserts mech blueprints
	for _, mechBluePrint := range tmpl.BlueprintMech {
		// templates with genesis and limited release mechs can only have ONE mech in it
		// here we check if we have a genesis/limited id that its only 1 mech
		if (tokenIDs.GenesisTokenID.Valid || tokenIDs.LimitedTokenID.Valid) && len(tmpl.BlueprintMech) > 1 {
			err := fmt.Errorf("template has already inserted a genesis mech but the template has multiple mechs")
			gamelog.L.Error().Err(err).
				Interface("mechBluePrint", mechBluePrint).
				Int64("tokenIDs.GenesisTokenID", tokenIDs.GenesisTokenID.Int64).
				Int64("tokenIDs.LimitedTokenID", tokenIDs.LimitedTokenID.Int64).
				Int("len(tmpl.BlueprintMech)", len(tmpl.BlueprintMech)).
				Msg("failed to insert new mech for user")
			continue
		}
		mechBluePrint.LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		mechBluePrint.GenesisTokenID = tokenIDs.GenesisTokenID
		insertedMech, err := InsertNewMech(ownerID, mechBluePrint)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("mechBluePrint", mechBluePrint).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new mech for user")
			continue
		}
		mechs = append(mechs, insertedMech)
	}

	// inserts mech animation blueprints
	for _, mechAnimation := range tmpl.BlueprintMechAnimation {

		insertedMechAnimations, err := InsertNewMechAnimation(ownerID, mechAnimation)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("mechAnimation", mechAnimation).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new mech animation for user")
			continue
		}
		mechAnimations = append(mechAnimations, insertedMechAnimations)
	}

	// inserts mech animation blueprints
	for _, mechSkin := range tmpl.BlueprintMechSkin {
		mechSkin.LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		mechSkin.GenesisTokenID = tokenIDs.GenesisTokenID
		insertedMechSkin, err := InsertNewMechSkin(ownerID, mechSkin)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("mechSkin", mechSkin).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new mech skin for user")
			continue
		}
		mechSkins = append(mechSkins, insertedMechSkin)
	}

	// inserts energy core blueprints
	for _, powerCore := range tmpl.BlueprintPowerCore {
		powerCore.LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		powerCore.GenesisTokenID = tokenIDs.GenesisTokenID
		insertedPowerCore, err := InsertNewPowerCore(ownerID, powerCore)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("powerCore", powerCore).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new energy core for user")
			continue
		}
		powerCores = append(powerCores, insertedPowerCore)
	}

	// inserts weapons blueprints
	for _, weapon := range tmpl.BlueprintWeapon {
		weapon.LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		weapon.GenesisTokenID = tokenIDs.GenesisTokenID
		insertedWeapon, err := InsertNewWeapon(ownerID, weapon)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("weapon", weapon).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new weapon for user")
			continue
		}
		weapons = append(weapons, insertedWeapon)
	}

	// inserts utility blueprints
	for _, utility := range tmpl.BlueprintUtility {
		utility.LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		utility.GenesisTokenID = tokenIDs.GenesisTokenID
		insertedUtility, err := InsertNewUtility(ownerID, utility)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("utility", utility).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new utility for user")
			continue
		}
		utilities = append(utilities, insertedUtility)
	}

	// if it contains a complete mech, lets build the mech!
	if tmpl.ContainsCompleteMechExactly {
		lockedToMech := false
		if tokenIDs.GenesisTokenID.Valid || tokenIDs.LimitedTokenID.Valid {
			lockedToMech = true
		}
		// join power core
		err = AttachPowerCoreToMech(ownerID.String(), mechs[0].ID, powerCores[0].ID)
		if err != nil {
			gamelog.L.Error().Err(err).
				Str("ownerID.String()", ownerID.String()).
				Str("mechs[0].ID", mechs[0].ID).
				Str("powerCores[0].ID", powerCores[0].ID).
				Msg("failed to join powercore to mech")
		}
		// join skin
		err = AttachMechSkinToMech(ownerID.String(), mechs[0].ID, mechSkins[0].ID, lockedToMech)
		if err != nil {
			gamelog.L.Error().Err(err).
				Str("ownerID.String()", ownerID.String()).
				Str("mechs[0].ID", mechs[0].ID).
				Str("mechSkins[0].ID", mechSkins[0].ID).
				Msg("failed to join skin to mech")
		}
		// join animations
		// TODO: no animations yet
		// join weapons
		for i := 0; i < mechs[0].WeaponHardpoints; i++ {
			if len(weapons) > i {
				err = AttachWeaponToMech(ownerID.String(), mechs[0].ID, weapons[i].ID)
				if err != nil {
					gamelog.L.Error().Err(err).
						Str("ownerID.String()", ownerID.String()).
						Str("mechs[0].ID", mechs[0].ID).
						Str("weapons[i].ID", weapons[i].ID).
						Msg("failed to join weapon to mech")
					continue
				}
			}
		}
		// join utility
		for i := 0; i < mechs[0].UtilitySlots; i++ {
			if utilities[i] != nil {
				err = AttachUtilityToMech(ownerID.String(), mechs[0].ID, utilities[i].ID, lockedToMech)
				if err != nil {
					gamelog.L.Error().Err(err).
						Str("ownerID.String()", ownerID.String()).
						Str("mechs[0].ID", mechs[0].ID).
						Str("utilities[i].ID", utilities[i].ID).
						Msg("failed to join utility to mech")
					continue
				}
			}
		}
	}

	return mechs,
		mechAnimations,
		mechSkins,
		powerCores,
		weapons,
		utilities, nil
}
