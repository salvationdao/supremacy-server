package comms

import (
	"server"
)

func ServerMechsToApiV1(items []*server.Mech) []*Mech {
	var converted []*Mech
	for _, i := range items {
		converted = append(converted, ServerMechToApiV1(i))
	}
	return converted
}

func ServerMechSkinsToApiV1(items []*server.MechSkin) []*MechSkin {
	var converted []*MechSkin
	for _, i := range items {
		converted = append(converted, ServerMechSkinToApiV1(i))
	}
	return converted
}

func ServerMechSkinToApiV1(skin *server.MechSkin) *MechSkin {
	return &MechSkin{
		CollectionDetails: &CollectionDetails{
			CollectionSlug: skin.CollectionDetails.CollectionSlug,
			Hash:           skin.CollectionDetails.Hash,
			TokenID:        skin.CollectionDetails.TokenID,
		},
		ID:               skin.ID,
		BlueprintID:      skin.BlueprintID,
		CollectionItemID: skin.CollectionItemID,
		GenesisTokenID:   skin.GenesisTokenID,
		Label:            skin.Label,
		OwnerID:          skin.OwnerID,
		ChassisModel:     skin.ChassisModel,
		EquippedOn:       skin.EquippedOn,
		Tier:             skin.Tier,
		ImageURL:         skin.ImageURL,
		AnimationURL:     skin.AnimationURL,
		CardAnimationURL: skin.CardAnimationURL,
		AvatarURL:        skin.AvatarURL,
		LargeImageURL:    skin.LargeImageURL,
		CreatedAt:        skin.CreatedAt,
	}
}

func ServerMechAnimationsToApiV1(items []*server.MechAnimation) []*MechAnimation {
	var converted []*MechAnimation
	for _, i := range items {
		converted = append(converted, ServerMechAnimationToApiV1(i))
	}
	return converted
}

func ServerMechAnimationToApiV1(animation *server.MechAnimation) *MechAnimation {
	return &MechAnimation{
		CollectionDetails: &CollectionDetails{
			CollectionSlug: animation.CollectionDetails.CollectionSlug,
			Hash:           animation.CollectionDetails.Hash,
			TokenID:        animation.CollectionDetails.TokenID,
		},
		ID:               animation.ID,
		BlueprintID:      animation.BlueprintID,
		CollectionItemID: animation.CollectionItemID,
		Label:            animation.Label,
		OwnerID:          animation.OwnerID,
		ChassisModel:     animation.ChassisModel,
		EquippedOn:       animation.EquippedOn,
		Tier:             animation.Tier,
		IntroAnimation:   animation.IntroAnimation,
		OutroAnimation:   animation.OutroAnimation,
		CreatedAt:        animation.CreatedAt,
	}
}

func ServerEnergyCoresToApiV1(items []*server.EnergyCore) []*EnergyCore {
	var converted []*EnergyCore
	for _, i := range items {
		converted = append(converted, ServerEnergyCoreToApiV1(i))
	}
	return converted
}

func ServerEnergyCoreToApiV1(ec *server.EnergyCore) *EnergyCore {
	return &EnergyCore{
		CollectionDetails: &CollectionDetails{
			CollectionSlug: ec.CollectionDetails.CollectionSlug,
			Hash:           ec.CollectionDetails.Hash,
			TokenID:        ec.CollectionDetails.TokenID,
		},
		ID:               ec.ID,
		CollectionItemID: ec.CollectionItemID,
		OwnerID:          ec.OwnerID,
		Label:            ec.Label,
		Size:             ec.Size,
		Capacity:         ec.Capacity,
		MaxDrawRate:      ec.MaxDrawRate,
		RechargeRate:     ec.RechargeRate,
		Armour:           ec.Armour,
		MaxHitpoints:     ec.MaxHitpoints,
		Tier:             ec.Tier,
		EquippedOn:       ec.EquippedOn,
		CreatedAt:        ec.CreatedAt,
	}
}

func ServerWeaponsToApiV1(items []*server.Weapon) []*Weapon {
	var converted []*Weapon
	for _, i := range items {
		converted = append(converted, ServerWeaponToApiV1(i))
	}
	return converted
}

func ServerWeaponToApiV1(weapon *server.Weapon) *Weapon {
	return &Weapon{
		CollectionDetails: &CollectionDetails{
			CollectionSlug: weapon.CollectionDetails.CollectionSlug,
			Hash:           weapon.CollectionDetails.Hash,
			TokenID:        weapon.CollectionDetails.TokenID,
		},
		ID:                   weapon.ID,
		BrandID:              weapon.BrandID,
		Label:                weapon.Label,
		Slug:                 weapon.Slug,
		Damage:               weapon.Damage,
		BlueprintID:          weapon.BlueprintID,
		DefaultDamageTyp:     weapon.DefaultDamageTyp,
		CollectionItemID:     weapon.CollectionItemID,
		GenesisTokenID:       weapon.GenesisTokenID,
		WeaponType:           weapon.WeaponType,
		OwnerID:              weapon.OwnerID,
		DamageFalloff:        weapon.DamageFalloff,
		DamageFalloffRate:    weapon.DamageFalloffRate,
		Spread:               weapon.Spread,
		RateOfFire:           weapon.RateOfFire,
		Radius:               weapon.Radius,
		RadialDoesFullDamage: weapon.RadialDoesFullDamage,
		ProjectileSpeed:      weapon.ProjectileSpeed,
		EnergyCost:           weapon.EnergyCost,
		MaxAmmo:              weapon.MaxAmmo,
		UpdatedAt:            weapon.UpdatedAt,
		CreatedAt:            weapon.CreatedAt,
	}
}

func ServerUtilitiesToApiV1(items []*server.Utility) []*Utility {
	var converted []*Utility
	for _, i := range items {
		converted = append(converted, ServerUtilityToApiV1(i))
	}
	return converted
}

func ServerUtilityToApiV1(ec *server.Utility) *Utility {
	result := &Utility{
		CollectionDetails: &CollectionDetails{
			CollectionSlug: ec.CollectionDetails.CollectionSlug,
			Hash:           ec.CollectionDetails.Hash,
			TokenID:        ec.CollectionDetails.TokenID,
		},
		ID:               ec.ID,
		BrandID:          ec.BrandID,
		Label:            ec.Label,
		UpdatedAt:        ec.UpdatedAt,
		CreatedAt:        ec.CreatedAt,
		BlueprintID:      ec.BlueprintID,
		CollectionItemID: ec.CollectionItemID,
		GenesisTokenID:   ec.GenesisTokenID,
		OwnerID:          ec.OwnerID,
		EquippedOn:       ec.EquippedOn,
		Type:             ec.Type,
	}
	switch ec.Type {
	case "SHIELD":
		if ec.Shield != nil {
			result.Shield = ServerUtilityShieldToApiV1(ec.Shield)
		}
	case "ATTACK DRONE":
		if ec.AttackDrone != nil {
			result.AttackDrone = ServerUtilityAttackDroneToApiV1(ec.AttackDrone)
		}
	case "REPAIR DRONE":
		if ec.RepairDrone != nil {
			result.RepairDrone = ServerUtilityRepairDroneToApiV1(ec.RepairDrone)
		}
	case "ANTI MISSILE":
		if ec.AntiMissile != nil {
			result.AntiMissile = ServerUtilityAntiMissileToApiV1(ec.AntiMissile)
		}
	case "ACCELERATOR":
		if ec.Accelerator != nil {
			result.Accelerator = ServerUtilityAcceleratorToApiV1(ec.Accelerator)
		}
	}

	return result
}

func ServerUtilityAcceleratorToApiV1(obj *server.UtilityAccelerator) *UtilityAccelerator {
	return &UtilityAccelerator{
		UtilityID:    obj.UtilityID,
		EnergyCost:   obj.EnergyCost,
		BoostSeconds: obj.BoostSeconds,
		BoostAmount:  obj.BoostAmount,
	}
}

func ServerUtilityAntiMissileToApiV1(obj *server.UtilityAntiMissile) *UtilityAntiMissile {
	return &UtilityAntiMissile{
		UtilityID:      obj.UtilityID,
		RateOfFire:     obj.RateOfFire,
		FireEnergyCost: obj.FireEnergyCost,
	}
}

func ServerUtilityRepairDroneToApiV1(obj *server.UtilityRepairDrone) *UtilityRepairDrone {
	return &UtilityRepairDrone{
		UtilityID:        obj.UtilityID,
		RepairType:       obj.RepairType,
		RepairAmount:     obj.RepairAmount,
		DeployEnergyCost: obj.DeployEnergyCost,
		LifespanSeconds:  obj.LifespanSeconds,
	}
}

func ServerUtilityShieldToApiV1(obj *server.UtilityShield) *UtilityShield {
	return &UtilityShield{
		UtilityID:          obj.UtilityID,
		Hitpoints:          obj.Hitpoints,
		RechargeRate:       obj.RechargeRate,
		RechargeEnergyCost: obj.RechargeEnergyCost,
	}
}

func ServerUtilityAttackDroneToApiV1(obj *server.UtilityAttackDrone) *UtilityAttackDrone {
	return &UtilityAttackDrone{
		UtilityID:        obj.UtilityID,
		Damage:           obj.Damage,
		RateOfFire:       obj.RateOfFire,
		Hitpoints:        obj.Hitpoints,
		LifespanSeconds:  obj.LifespanSeconds,
		DeployEnergyCost: obj.DeployEnergyCost,
	}
}

func ServerMechToApiV1(mech *server.Mech) *Mech {
	return &Mech{
		CollectionDetails: &CollectionDetails{
			CollectionSlug: mech.CollectionDetails.CollectionSlug,
			Hash:           mech.CollectionDetails.Hash,
			TokenID:        mech.CollectionDetails.TokenID,
		},
		ID:                   mech.ID,
		BrandID:              mech.BrandID,
		Label:                mech.Label,
		WeaponHardpoints:     mech.WeaponHardpoints,
		UtilitySlots:         mech.UtilitySlots,
		Speed:                mech.Speed,
		MaxHitpoints:         mech.MaxHitpoints,
		BlueprintID:          mech.BlueprintID,
		IsDefault:            mech.IsDefault,
		IsInsured:            mech.IsInsured,
		Name:                 mech.Name,
		ModelID:              mech.ModelID,
		CollectionItemID:     mech.CollectionItemID,
		GenesisTokenID:       mech.GenesisTokenID,
		OwnerID:              mech.OwnerID,
		FactionID:            mech.FactionID,
		EnergyCoreSize:       mech.EnergyCoreSize,
		Tier:                 mech.Tier,
		DefaultChassisSkinID: mech.DefaultChassisSkinID,
		DefaultChassisSkin:   ServerBlueprintMechSkinToApiV1(mech.DefaultChassisSkin),
		ChassisSkinID:        mech.ChassisSkinID,
		ChassisSkin:          ServerMechSkinToApiV1(mech.ChassisSkin),
		IntroAnimationID:     mech.IntroAnimationID,
		IntroAnimation:       ServerMechAnimationToApiV1(mech.IntroAnimation),
		OutroAnimationID:     mech.OutroAnimationID,
		OutroAnimation:       ServerMechAnimationToApiV1(mech.OutroAnimation),
		EnergyCoreID:         mech.EnergyCoreID,
		EnergyCore:           ServerEnergyCoreToApiV1(mech.EnergyCore),
		Weapons:              ServerWeaponsToApiV1(mech.Weapons),
		Utility:              ServerUtilitiesToApiV1(mech.Utility),
		UpdatedAt:            mech.UpdatedAt,
		CreatedAt:            mech.CreatedAt,
	}
}
