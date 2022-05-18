package comms

import (
	"server"
)

func ServerBlueprintMechsToApiV1(items []*server.BlueprintMech) []*BlueprintMech {
	var converted []*BlueprintMech
	for _, i := range items {
		converted = append(converted, ServerBlueprintMechToApiV1(i))
	}
	return converted
}

func ServerBlueprintMechToApiV1(mech *server.BlueprintMech) *BlueprintMech {
	return &BlueprintMech{
		ID:                   mech.ID,
		BrandID:              mech.BrandID,
		Label:                mech.Label,
		Slug:                 mech.Slug,
		Skin:                 mech.Skin,
		WeaponHardpoints:     mech.WeaponHardpoints,
		UtilitySlots:         mech.UtilitySlots,
		Speed:                mech.Speed,
		MaxHitpoints:         mech.MaxHitpoints,
		UpdatedAt:            mech.UpdatedAt,
		CreatedAt:            mech.CreatedAt,
		ModelID:              mech.ModelID,
		PowerCoreSize:        mech.PowerCoreSize,
		Tier:                 mech.Tier,
		DefaultChassisSkinID: mech.DefaultChassisSkinID,
	}
}

func ServerBlueprintWeaponsToApiV1(items []*server.BlueprintWeapon) []*BlueprintWeapon {
	var converted []*BlueprintWeapon
	for _, i := range items {
		converted = append(converted, ServerBlueprintWeaponToApiV1(i))
	}
	return converted
}

func ServerBlueprintWeaponToApiV1(weapon *server.BlueprintWeapon) *BlueprintWeapon {
	return &BlueprintWeapon{
		ID:                  weapon.ID,
		BrandID:             weapon.BrandID,
		Label:               weapon.Label,
		Slug:                weapon.Slug,
		Damage:              weapon.Damage,
		UpdatedAt:           weapon.UpdatedAt,
		CreatedAt:           weapon.CreatedAt,
		GameClientWeaponID:  weapon.GameClientWeaponID,
		WeaponType:          weapon.WeaponType,
		DefaultDamageType:   weapon.DefaultDamageType,
		DamageFalloff:       weapon.DamageFalloff,
		DamageFalloffRate:   weapon.DamageFalloffRate,
		Spread:              weapon.Spread,
		RateOfFire:          weapon.RateOfFire,
		Radius:              weapon.Radius,
		RadiusDamageFalloff: weapon.RadiusDamageFalloff,
		ProjectileSpeed:     weapon.ProjectileSpeed,
		MaxAmmo:             weapon.MaxAmmo,
		EnergyCost:          weapon.EnergyCost,
	}
}

func ServerBlueprintMechSkinsToApiV1(items []*server.BlueprintMechSkin) []*BlueprintMechSkin {
	var converted []*BlueprintMechSkin
	for _, i := range items {
		converted = append(converted, ServerBlueprintMechSkinToApiV1(i))
	}
	return converted
}

func ServerBlueprintMechSkinToApiV1(skin *server.BlueprintMechSkin) *BlueprintMechSkin {
	return &BlueprintMechSkin{
		ID:               skin.ID,
		Collection:       skin.Collection,
		MechModel:        skin.MechModel,
		Label:            skin.Label,
		Tier:             skin.Tier,
		ImageURL:         skin.ImageURL,
		AnimationURL:     skin.AnimationURL,
		CardAnimationURL: skin.CardAnimationURL,
		LargeImageURL:    skin.LargeImageURL,
		AvatarURL:        skin.AvatarURL,
		CreatedAt:        skin.CreatedAt,
	}
}

func ServerBlueprintMechAnimationsToApiV1(items []*server.BlueprintMechAnimation) []*BlueprintMechAnimation {
	var converted []*BlueprintMechAnimation
	for _, i := range items {
		converted = append(converted, ServerBlueprintMechAnimationToApiV1(i))
	}
	return converted
}

func ServerBlueprintMechAnimationToApiV1(animation *server.BlueprintMechAnimation) *BlueprintMechAnimation {
	return &BlueprintMechAnimation{
		ID:             animation.ID,
		Collection:     animation.Collection,
		Label:          animation.Label,
		MechModel:      animation.MechModel,
		Tier:           animation.Tier,
		IntroAnimation: animation.IntroAnimation,
		OutroAnimation: animation.OutroAnimation,
		CreatedAt:      animation.CreatedAt,
	}
}

func ServerBlueprintPowerCoresToApiV1(items []*server.BlueprintPowerCore) []*BlueprintPowerCore {
	var converted []*BlueprintPowerCore
	for _, i := range items {
		converted = append(converted, ServerBlueprintPowerCoreToApiV1(i))
	}
	return converted
}

func ServerBlueprintPowerCoreToApiV1(ec *server.BlueprintPowerCore) *BlueprintPowerCore {
	return &BlueprintPowerCore{
		ID:           ec.ID,
		Collection:   ec.Collection,
		Label:        ec.Label,
		Size:         ec.Size,
		Capacity:     ec.Capacity,
		MaxDrawRate:  ec.MaxDrawRate,
		RechargeRate: ec.RechargeRate,
		Armour:       ec.Armour,
		MaxHitpoints: ec.MaxHitpoints,
		Tier:         ec.Tier,
		CreatedAt:    ec.CreatedAt,
	}
}

func ServerBlueprintUtilitiesToApiV1(items []*server.BlueprintUtility) []*BlueprintUtility {
	var converted []*BlueprintUtility
	for _, i := range items {
		converted = append(converted, ServerBlueprintUtilityToApiV1(i))
	}
	return converted
}

func ServerBlueprintUtilityToApiV1(ec *server.BlueprintUtility) *BlueprintUtility {
	result := &BlueprintUtility{
		ID:        ec.ID,
		BrandID:   ec.BrandID,
		Label:     ec.Label,
		UpdatedAt: ec.UpdatedAt,
		CreatedAt: ec.CreatedAt,
		Type:      ec.Type,
	}
	switch ec.Type {
	case "SHIELD":
		if ec.ShieldBlueprint != nil {
			result.UtilityObject = ServerBlueprintUtilityShieldToApiV1(ec.ShieldBlueprint)
		}
	case "ATTACK DRONE":
		if ec.AttackDroneBlueprint != nil {
			result.UtilityObject = ServerBlueprintUtilityAttackDroneToApiV1(ec.AttackDroneBlueprint)
		}
	case "REPAIR DRONE":
		if ec.RepairDroneBlueprint != nil {
			result.UtilityObject = ServerBlueprintUtilityRepairDroneToApiV1(ec.RepairDroneBlueprint)
		}
	case "ANTI MISSILE":
		if ec.AntiMissileBlueprint != nil {
			result.UtilityObject = ServerBlueprintUtilityAntiMissileToApiV1(ec.AntiMissileBlueprint)
		}
	case "ACCELERATOR":
		if ec.AcceleratorBlueprint != nil {
			result.UtilityObject = ServerBlueprintUtilityAcceleratorToApiV1(ec.AcceleratorBlueprint)
		}
	}

	return result
}

func ServerBlueprintUtilityAcceleratorToApiV1(obj *server.BlueprintUtilityAccelerator) *BlueprintUtilityAccelerator {
	return &BlueprintUtilityAccelerator{
		ID:                 obj.ID,
		BlueprintUtilityID: obj.BlueprintUtilityID,
		EnergyCost:         obj.EnergyCost,
		BoostSeconds:       obj.BoostSeconds,
		BoostAmount:        obj.BoostAmount,
		CreatedAt:          obj.CreatedAt,
	}
}

func ServerBlueprintUtilityAntiMissileToApiV1(obj *server.BlueprintUtilityAntiMissile) *BlueprintUtilityAntiMissile {
	return &BlueprintUtilityAntiMissile{
		ID:                 obj.ID,
		BlueprintUtilityID: obj.BlueprintUtilityID,
		RateOfFire:         obj.RateOfFire,
		FireEnergyCost:     obj.FireEnergyCost,
		CreatedAt:          obj.CreatedAt,
	}
}

func ServerBlueprintUtilityRepairDroneToApiV1(obj *server.BlueprintUtilityRepairDrone) *BlueprintUtilityRepairDrone {
	return &BlueprintUtilityRepairDrone{
		ID:                 obj.ID,
		BlueprintUtilityID: obj.BlueprintUtilityID,
		RepairType:         obj.RepairType,
		RepairAmount:       obj.RepairAmount,
		DeployEnergyCost:   obj.DeployEnergyCost,
		LifespanSeconds:    obj.LifespanSeconds,
		CreatedAt:          obj.CreatedAt,
	}
}

func ServerBlueprintUtilityShieldToApiV1(obj *server.BlueprintUtilityShield) *BlueprintUtilityShield {
	return &BlueprintUtilityShield{
		ID:                 obj.ID,
		BlueprintUtilityID: obj.BlueprintUtilityID,
		Hitpoints:          obj.Hitpoints,
		RechargeRate:       obj.RechargeRate,
		RechargeEnergyCost: obj.RechargeEnergyCost,
		CreatedAt:          obj.CreatedAt,
	}
}

func ServerBlueprintUtilityAttackDroneToApiV1(obj *server.BlueprintUtilityAttackDrone) *BlueprintUtilityAttackDrone {
	return &BlueprintUtilityAttackDrone{
		ID:                 obj.ID,
		BlueprintUtilityID: obj.BlueprintUtilityID,
		Damage:             obj.Damage,
		RateOfFire:         obj.RateOfFire,
		Hitpoints:          obj.Hitpoints,
		LifespanSeconds:    obj.LifespanSeconds,
		DeployEnergyCost:   obj.DeployEnergyCost,
		CreatedAt:          obj.CreatedAt,
	}
}

func ServerTemplateToApiTemplateV1(temp *server.TemplateContainer) *TemplateContainer {
	return &TemplateContainer{
		ID:                     temp.ID,
		Label:                  temp.Label,
		UpdatedAt:              temp.UpdatedAt,
		CreatedAt:              temp.CreatedAt,
		BlueprintMech:          ServerBlueprintMechsToApiV1(temp.BlueprintMech),
		BlueprintWeapon:        ServerBlueprintWeaponsToApiV1(temp.BlueprintWeapon),
		BlueprintUtility:       ServerBlueprintUtilitiesToApiV1(temp.BlueprintUtility),
		BlueprintMechSkin:      ServerBlueprintMechSkinsToApiV1(temp.BlueprintMechSkin),
		BlueprintMechAnimation: ServerBlueprintMechAnimationsToApiV1(temp.BlueprintMechAnimation),
		BlueprintPowerCore:     ServerBlueprintPowerCoresToApiV1(temp.BlueprintPowerCore),
	}
}
