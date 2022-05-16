package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/shopspring/decimal"

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
		ID:                     template.ID,
		Label:                  template.Label,
		UpdatedAt:              template.UpdatedAt,
		CreatedAt:              template.CreatedAt,
		BlueprintMech:          []*server.BlueprintMech{},
		BlueprintWeapon:        []*server.BlueprintWeapon{},
		BlueprintUtility:       []*server.BlueprintUtility{},
		BlueprintMechSkin:      []*server.BlueprintMechSkin{},
		BlueprintMechAnimation: []*server.BlueprintMechAnimation{},
		BlueprintPowerCore:     []*server.BlueprintPowerCore{},
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

	genesisTokenID := decimal.NullDecimal{}
	limitedReleaseTokenID := decimal.NullDecimal{}

	// inserts mech blueprints
	for _, mechBluePrint := range tmpl.BlueprintMech {
		// templates with genesis and limited release mechs can only have ONE mech in it
		// here we check if we have a genesis/limited id that its only 1 mech
		if (genesisTokenID.Valid || limitedReleaseTokenID.Valid) && len(tmpl.BlueprintMech) > 1 {
			err := fmt.Errorf("template has already inserted a genesis mech but the template has multiple mechs")
			gamelog.L.Error().Err(err).
				Interface("mechBluePrint", mechBluePrint).
				Str("genesisTokenID", genesisTokenID.Decimal.String()).
				Str("limitedReleaseTokenID", limitedReleaseTokenID.Decimal.String()).
				Int("len(tmpl.BlueprintMech)", len(tmpl.BlueprintMech)).
				Msg("failed to insert new mech for user")
			continue
		}

		insertedMech, err := InsertNewMech(ownerID, mechBluePrint)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("mechBluePrint", mechBluePrint).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new mech for user")
			continue
		}
		mechs = append(mechs, insertedMech)

		// if the inserted mech is a genesis or limited release, we need to add that token id across all items we're inserting
		if insertedMech.GenesisTokenID.Valid {
			genesisTokenID = insertedMech.GenesisTokenID
		}
		if insertedMech.LimitedReleaseTokenID.Valid {
			limitedReleaseTokenID = insertedMech.LimitedReleaseTokenID
		}
	}

	// inserts mech animation blueprints
	for _, mechAnimation := range tmpl.BlueprintMechAnimation {
		mechAnimation.LimitedReleaseTokenID = limitedReleaseTokenID
		mechAnimation.GenesisTokenID = genesisTokenID
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
		mechSkin.LimitedReleaseTokenID = limitedReleaseTokenID
		mechSkin.GenesisTokenID = genesisTokenID
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
		powerCore.LimitedReleaseTokenID = limitedReleaseTokenID
		powerCore.GenesisTokenID = genesisTokenID
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
		weapon.LimitedReleaseTokenID = limitedReleaseTokenID
		weapon.GenesisTokenID = genesisTokenID
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
		utility.LimitedReleaseTokenID = limitedReleaseTokenID
		utility.GenesisTokenID = genesisTokenID
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

	return mechs,
		mechAnimations,
		mechSkins,
		powerCores,
		weapons,
		utilities, nil
}
