package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"

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
		// handle
	}
	result.BlueprintWeapon, err = BlueprintWeapons(blueprintWeaponIDS)
	if err != nil {
		// handle
	}
	result.BlueprintMechSkin, err = BlueprintMechSkinSkins(blueprintMechSkinIDS)
	if err != nil {
		// handle
	}
	result.BlueprintMechAnimation, err = BlueprintMechAnimations(blueprintMechAnimationIDS)
	if err != nil {
		// handle
	}
	result.BlueprintEnergyCore, err = BlueprintEnergyCores(blueprintEnergyCoreIDS)
	if err != nil {
		// handle
	}
	result.BlueprintUtility, err = BlueprintUtilities(blueprintUtilityIDS)
	if err != nil {
		// handle
	}
	return result, nil
}
