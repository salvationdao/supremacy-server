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
		EnergyCoreSize:       mech.EnergyCoreSize,
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
		ID:                   weapon.ID,
		BrandID:              weapon.BrandID,
		Label:                weapon.Label,
		Slug:                 weapon.Slug,
		Damage:               weapon.Damage,
		UpdatedAt:            weapon.UpdatedAt,
		CreatedAt:            weapon.CreatedAt,
		GameClientWeaponID:   weapon.GameClientWeaponID,
		WeaponType:           weapon.WeaponType,
		DefaultDamageTyp:     weapon.DefaultDamageTyp,
		DamageFalloff:        weapon.DamageFalloff,
		DamageFalloffRate:    weapon.DamageFalloffRate,
		Spread:               weapon.Spread,
		RateOfFire:           weapon.RateOfFire,
		Radius:               weapon.Radius,
		RadialDoesFullDamage: weapon.RadialDoesFullDamage,
		ProjectileSpeed:      weapon.ProjectileSpeed,
		MaxAmmo:              weapon.MaxAmmo,
		EnergyCost:           weapon.EnergyCost,
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
		ChassisModel:     skin.ChassisModel,
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
		ChassisModel:   animation.ChassisModel,
		EquippedOn:     animation.EquippedOn,
		Tier:           animation.Tier,
		IntroAnimation: animation.IntroAnimation,
		OutroAnimation: animation.OutroAnimation,
		CreatedAt:      animation.CreatedAt,
	}
}

func ServerBlueprintEnergyCoresToApiV1(items []*server.BlueprintEnergyCore) []*BlueprintEnergyCore {
	var converted []*BlueprintEnergyCore
	for _, i := range items {
		converted = append(converted, ServerBlueprintEnergyCoreToApiV1(i))
	}
	return converted
}

func ServerBlueprintEnergyCoreToApiV1(ec *server.BlueprintEnergyCore) *BlueprintEnergyCore {
	return &BlueprintEnergyCore{
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

func ServerTemplateToApiTemplateV1(temp *server.TemplateContainer) *TemplateContainer {
	return &TemplateContainer{
		ID:                     temp.ID,
		Label:                  temp.Label,
		UpdatedAt:              temp.UpdatedAt,
		CreatedAt:              temp.CreatedAt,
		BlueprintMech:          ServerBlueprintMechsToApiV1(temp.BlueprintMech),
		BlueprintWeapon:        ServerBlueprintWeaponsToApiV1(temp.BlueprintWeapon),
		//BlueprintUtility:       temp.BlueprintUtility,
		BlueprintMechSkin:      ServerBlueprintMechSkinsToApiV1(temp.BlueprintMechSkin),
		BlueprintMechAnimation: ServerBlueprintMechAnimationsToApiV1(temp.BlueprintMechAnimation),
		BlueprintEnergyCore:    ServerBlueprintEnergyCoresToApiV1(temp.BlueprintEnergyCore),
	}
}
