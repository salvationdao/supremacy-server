package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

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
		return nil, err
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
		BlueprintEnergyCore:    []*server.BlueprintEnergyCore{},
	}

	// filter them into IDs first to optimize db queries
	blueprintMechIDS := []string{}
	blueprintWeaponIDS := []string{}
	blueprintUtilityIDS := []string{}
	blueprintMechSkinIDS := []string{}
	blueprintMechAnimationIDS := []string{}
	blueprintEnergyCoreIDS := []string{}

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
		case boiler.TemplateItemTypeENERGY_CORE:
			blueprintEnergyCoreIDS = append(blueprintEnergyCoreIDS, bp.BlueprintID)
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
		return nil, err
	}
	result.BlueprintWeapon, err = BlueprintWeapons(blueprintWeaponIDS)
	if err != nil {
		return nil, err
	}
	result.BlueprintMechSkin, err = BlueprintMechSkinSkins(blueprintMechSkinIDS)
	if err != nil {
		return nil, err
	}
	result.BlueprintMechAnimation, err = BlueprintMechAnimations(blueprintMechAnimationIDS)
	if err != nil {
		return nil, err
	}
	result.BlueprintEnergyCore, err = BlueprintEnergyCores(blueprintEnergyCoreIDS)
	if err != nil {
		return nil, err
	}
	result.BlueprintUtility, err = BlueprintUtilities(blueprintUtilityIDS)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TemplateRegister copies everything out of a template into a new mech
func TemplateRegister(templateID uuid.UUID, ownerID uuid.UUID) error {

	exists, err := boiler.PlayerExists(gamedb.StdConn, ownerID.String())
	if err != nil {
		return fmt.Errorf("check player exists: %w", err)
	}
	if !exists {
		newPlayer := &boiler.Player{ID: ownerID.String()}
		err = newPlayer.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return fmt.Errorf("insert new player: %w", err)
		}
	}

	tmpl, err := Template(templateID)
	if err != nil {
		return fmt.Errorf("find template: %w", err)
	}

	// inserts mech blueprints
	for _, mechBluePrint := range tmpl.BlueprintMech {
		err := InsertNewMech(ownerID, mechBluePrint)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("mechAnimation", mechBluePrint).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new mech for user")
			continue
		}
	}

	// inserts mech animation blueprints
	for _, mechAnimation := range tmpl.BlueprintMechAnimation {
		err := InsertNewMechAnimation(ownerID, mechAnimation)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("mechAnimation", mechAnimation).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new mech animation for user")
			continue
		}
	}

	// inserts mech animation blueprints
	for _, mechSkin := range tmpl.BlueprintMechSkin {
		err := InsertNewMechSkin(ownerID, mechSkin)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("mechSkin", mechSkin).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new mech skin for user")
			continue
		}
	}

	// inserts energy core blueprints
	for _, energyCore := range tmpl.BlueprintEnergyCore {
		err := InsertNewEnergyCore(ownerID, energyCore)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("energyCore", energyCore).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new energy core for user")
			continue
		}
	}

	// inserts weapons blueprints
	for _, weapon := range tmpl.BlueprintWeapon {
		err := InsertNewWeapon(ownerID, weapon)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("weapon", weapon).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new weapon for user")
			continue
		}
	}

	// inserts utility blueprints
	for _, utility := range tmpl.BlueprintUtility {
		err := InsertNewUtility(ownerID, utility)
		if err != nil {
			gamelog.L.Error().Err(err).
				Interface("utility", utility).
				Str("ownerID", ownerID.String()).
				Msg("failed to insert new utility for user")
			continue
		}
	}

	return nil
}
