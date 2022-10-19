package db

import (
	"context"
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sort"

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
		// TODO: AMMO
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
	result.BlueprintMechSkin, err = BlueprintMechSkinSkins(gamedb.StdConn, blueprintMechSkinIDS)
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

// TemplateRegister copies everything out of a template into a new mech
func TemplateRegister(templateID uuid.UUID, ownerID uuid.UUID) (
	[]*server.Mech,
	[]*server.MechAnimation,
	[]*server.MechSkin,
	[]*server.PowerCore,
	[]*server.Weapon,
	[]*server.WeaponSkin,
	[]*server.Utility,
	error) {
	L := gamelog.L.With().Str("func", "TemplateRegister").Str("templateID", templateID.String()).Str("ownerID", ownerID.String()).Logger()

	var mechs []*server.Mech
	var mechAnimations []*server.MechAnimation
	var mechSkins []*server.MechSkin
	var powerCores []*server.PowerCore
	var weapons []*server.Weapon
	var weaponSkins []*server.WeaponSkin
	var utilities []*server.Utility

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		L.Error().Err(err).Msg("failed to begin tx")
		return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, fmt.Errorf("failed to start tx: %w", err)
	}
	defer tx.Rollback()

	exists, err := boiler.PlayerExists(tx, ownerID.String())
	if err != nil {
		L.Error().Err(err).Msg("error finding player")
		return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, fmt.Errorf("check player exists: %w", err)
	}
	if !exists {
		newPlayer := &boiler.Player{ID: ownerID.String()}
		err = newPlayer.Insert(tx, boil.Infer())
		if err != nil {
			L.Error().Err(err).Msg("error inserting player")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, fmt.Errorf("insert new player: %w", err)
		}
	}

	tmpl, err := Template(templateID)
	if err != nil {
		L.Error().Err(err).Msg("error finding template")
		return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, fmt.Errorf("find template: %w", err)
	}

	L = L.With().Interface("template", tmpl).Logger()

	if len(tmpl.BlueprintWeapon) != 3 {
		L.Error().Err(fmt.Errorf("not 3 blueprint weapons")).Msg("not 3 blueprint weapons")
		return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, fmt.Errorf("not 3 blueprint weapons")
	}

	// do these checks here because we can't return error after we assign the items
	if tmpl.ContainsCompleteMechExactly {
		if len(tmpl.BlueprintMech) != 1 {
			err = fmt.Errorf("template contains complete mech exactly but has more than 1 mech")
			L.Error().Err(err).Msg("template has wrong amount of mechs")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, err
		}
		if len(tmpl.BlueprintPowerCore) != 1 {
			err = fmt.Errorf("template contains complete mech exactly but has wrong amount of power cores")
			L.Error().Err(err).Msg("template has wrong amount of powercores")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, err
		}
		if len(tmpl.BlueprintWeapon) > tmpl.BlueprintMech[0].WeaponHardpoints {
			err = fmt.Errorf("template contains complete mech exactly but contained mech has less weapon slots than weapons provided")
			L.Error().Err(err).Msg("template has too many weapons")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, err
		}
		if len(tmpl.BlueprintUtility) > tmpl.BlueprintMech[0].UtilitySlots {
			err = fmt.Errorf("template contains complete mech exactly but contained mech has less utility slots than utilities provided")
			L.Error().Err(err).Msg("template has too many utilities")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, err
		}
		if len(tmpl.BlueprintMechSkin) > 1 {
			err = fmt.Errorf("template contains complete mech exactly but contained more than a single skin")
			L.Error().Err(err).Msg("template has too many mech skins")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, err
		}
	}

	tokenIDs := &struct {
		GenesisTokenID null.Int64 `json:"genesis_token_id" db:"genesis_token_id"`
		LimitedTokenID null.Int64 `json:"limited_token_id" db:"limited_token_id"`
	}{}

	// if template is genesis, create it a genesis ID
	if tmpl.IsGenesis {
		// get the max genesis
		err := boiler.NewQuery(qm.SQL(`SELECT coalesce(max(genesis_token_id) + 1, 0) as genesis_token_id FROM mechs`)).Bind(context.Background(), tx, tokenIDs)
		if err != nil {
			L.Error().Err(err).Msg("failed to get new genesis token id")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}
	}
	// if template is genesis, create it a genesis ID
	if tmpl.IsLimitedRelease {
		// get the max limited
		err := boiler.NewQuery(qm.SQL(`SELECT coalesce(max(limited_release_token_id) + 1, 0) as limited_token_id FROM mechs`)).Bind(context.Background(), tx, tokenIDs)
		if err != nil {
			L.Error().Err(err).Msg("failed to get new limit release token id")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}
	}

	L = L.With().Interface("tokenIDs", tokenIDs).Logger()

	// mech skins cannot be null so check we have enough skins
	if len(tmpl.BlueprintMech) != len(tmpl.BlueprintMechSkin) {
		if err != nil {
			err = fmt.Errorf("mismatch amount of mechs and skins")
			L.Error().Err(err).Int("len(tmpl.BlueprintMech)", len(tmpl.BlueprintMech)).Int("len(tmpl.BlueprintMechSkin)", len(tmpl.BlueprintMechSkin)).Msg("invalid mech and skin amounts")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}
	}

	// inserts mech blueprints
	for i, mechBluePrint := range tmpl.BlueprintMech {
		// templates with genesis and limited release mechs can only have ONE mech in it
		// here we check if we have a genesis/limited id that its only 1 mech
		if (tokenIDs.GenesisTokenID.Valid || tokenIDs.LimitedTokenID.Valid) && len(tmpl.BlueprintMech) > 1 {
			err := fmt.Errorf("template has already inserted a genesis mech but the template has multiple mechs")
			L.Error().Err(err).Msg("failed to insert new mech for user")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}
		mechBluePrint.LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		mechBluePrint.GenesisTokenID = tokenIDs.GenesisTokenID
		tmpl.BlueprintMechSkin[i].LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		tmpl.BlueprintMechSkin[i].GenesisTokenID = tokenIDs.GenesisTokenID

		insertedMech, insertedMechSkin, err := InsertNewMechAndSkin(tx, ownerID, mechBluePrint, tmpl.BlueprintMechSkin[i])
		if err != nil {
			L.Error().Err(err).Msg("failed to insert new mech for user")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}
		mechs = append(mechs, insertedMech)
		mechSkins = append(mechSkins, insertedMechSkin)
	}

	L = L.With().Interface("inserted mechs", mechs).Interface("inserted mech skin", mechSkins).Logger()

	// inserts mech animation blueprints
	for _, mechAnimation := range tmpl.BlueprintMechAnimation {
		insertedMechAnimations, err := InsertNewMechAnimation(tx, ownerID, mechAnimation)
		if err != nil {
			L.Error().Err(err).Msg("failed to insert new mech animation for user")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}
		mechAnimations = append(mechAnimations, insertedMechAnimations)
	}

	L = L.With().Interface("inserted mech animations", mechAnimations).Logger()

	// inserts power core blueprints
	for _, powerCore := range tmpl.BlueprintPowerCore {
		powerCore.LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		powerCore.GenesisTokenID = tokenIDs.GenesisTokenID
		insertedPowerCore, err := InsertNewPowerCore(tx, ownerID, powerCore)
		if err != nil {
			L.Error().Err(err).Msg("failed to insert new power core for user")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}
		powerCores = append(powerCores, insertedPowerCore)
	}

	L = L.With().Interface("inserted mech powercores", powerCores).Logger()

	// inserts weapons blueprints
	for _, weapon := range tmpl.BlueprintWeapon {
		weapon.LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		weapon.GenesisTokenID = tokenIDs.GenesisTokenID
		// get default weapon skins
		wpSkin, err := boiler.WeaponModelSkinCompatibilities(
			boiler.WeaponModelSkinCompatibilityWhere.WeaponModelID.EQ(weapon.ID),
			qm.Load(boiler.WeaponModelSkinCompatibilityRels.BlueprintWeaponSkin),
			).One(tx)
		if err != nil {
			L.Error().Err(err).Msg("failed to get weapon skin")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}
		weaponSkin := server.BlueprintWeaponSkinFromBoiler(wpSkin.R.BlueprintWeaponSkin)
		weaponSkin.LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		weaponSkin.GenesisTokenID = tokenIDs.GenesisTokenID

		insertedWeapon, insertedWeaponSkin, err := InsertNewWeapon(tx, ownerID, weapon, weaponSkin)
		if err != nil {
			L.Error().Err(err).Msg("failed to insert new weapon for user")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}
		weapons = append(weapons, insertedWeapon)
		weaponSkins = append(weaponSkins, insertedWeaponSkin)
	}

	L = L.With().Interface("inserted mech weapons", weapons).Interface("inserted weapon skins", weaponSkins).Logger()

	// inserts utility blueprints
	for _, utility := range tmpl.BlueprintUtility {
		utility.LimitedReleaseTokenID = tokenIDs.LimitedTokenID
		utility.GenesisTokenID = tokenIDs.GenesisTokenID
		insertedUtility, err := InsertNewUtility(tx, ownerID, utility)
		if err != nil {
			L.Error().Err(err).Msg("failed to insert new utility for user")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}
		utilities = append(utilities, insertedUtility)
	}

	L = L.With().Interface("inserted mech utilities", utilities).Logger()

	// if it contains a complete mech, lets build the mech!
	if tmpl.ContainsCompleteMechExactly {
		lockedToMech := false
		if tokenIDs.GenesisTokenID.Valid || tokenIDs.LimitedTokenID.Valid {
			lockedToMech = true
		}
		// join power core
		err = AttachPowerCoreToMech(tx, ownerID.String(), mechs[0].ID, powerCores[0].ID)
		if err != nil {
			L.Error().Err(err).Msg("failed to join powercore to mech")
			return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
		}

		// join animations
		// TODO: no animations yet
		// join weapons

		// sort slice to make rocket pods last
		sort.Slice(weapons, func(i, j int) bool {
			return weapons[i].ID == server.WeaponRocketPodsZai || weapons[i].ID == server.WeaponRocketPodsRM || weapons[i].ID == server.WeaponRocketPodsBC
		})
		for i := 0; i < mechs[0].WeaponHardpoints; i++ {
			if len(weapons) > i {
				err = AttachWeaponToMech(tx, ownerID.String(), mechs[0].ID, weapons[i].ID)
				if err != nil {
					L.Error().Err(err).Msg("failed to join weapon to mech")
					return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
				}
			}
		}
		// join utility
		for i := 0; i < mechs[0].UtilitySlots; i++ {
			if len(utilities) > 0 && len(utilities) >= i && utilities[i] != nil {
				err = AttachUtilityToMech(tx, ownerID.String(), mechs[0].ID, utilities[i].ID, lockedToMech)
				if err != nil {
					L.Error().Err(err).Msg("failed to join utility to mech")
					return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, terror.Error(err)
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, fmt.Errorf("failed to commit tx: %w", err)
	}

	return mechs, mechAnimations, mechSkins, powerCores, weapons, weaponSkins, utilities, nil
}
