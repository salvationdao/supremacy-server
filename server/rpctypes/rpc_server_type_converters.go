package rpctypes

import (
	"encoding/json"
	"server"
	"server/gamelog"
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
			CollectionSlug:   skin.CollectionDetails.CollectionSlug,
			Hash:             skin.CollectionDetails.Hash,
			TokenID:          skin.CollectionDetails.TokenID,
			ItemType:         skin.CollectionDetails.ItemType,
			ItemID:           skin.CollectionDetails.ItemID,
			Tier:             skin.CollectionDetails.Tier,
			OwnerID:          skin.CollectionDetails.OwnerID,
			OnChainStatus:    skin.CollectionDetails.OnChainStatus,
			ImageURL:         skin.CollectionDetails.ImageURL,
			CardAnimationURL: skin.CollectionDetails.CardAnimationURL,
			AvatarURL:        skin.CollectionDetails.AvatarURL,
			LargeImageURL:    skin.CollectionDetails.LargeImageURL,
			BackgroundColor:  skin.CollectionDetails.BackgroundColor,
			AnimationURL:     skin.CollectionDetails.AnimationURL,
			YoutubeURL:       skin.CollectionDetails.YoutubeURL,
		},
		ID:               skin.ID,
		BlueprintID:      skin.BlueprintID,
		GenesisTokenID:   skin.GenesisTokenID,
		Label:            skin.Label,
		OwnerID:          skin.OwnerID,
		MechModel:        skin.MechModel,
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
			ItemType:       animation.CollectionDetails.ItemType,
			ItemID:         animation.CollectionDetails.ItemID,
			Tier:           animation.CollectionDetails.Tier,
			OwnerID:        animation.CollectionDetails.OwnerID,
			OnChainStatus:  animation.CollectionDetails.OnChainStatus,

			ImageURL:         animation.CollectionDetails.ImageURL,
			CardAnimationURL: animation.CollectionDetails.CardAnimationURL,
			AvatarURL:        animation.CollectionDetails.AvatarURL,
			LargeImageURL:    animation.CollectionDetails.LargeImageURL,
			BackgroundColor:  animation.CollectionDetails.BackgroundColor,
			AnimationURL:     animation.CollectionDetails.AnimationURL,
			YoutubeURL:       animation.CollectionDetails.YoutubeURL,
		},
		ID:             animation.ID,
		BlueprintID:    animation.BlueprintID,
		Label:          animation.Label,
		OwnerID:        animation.OwnerID,
		MechModel:      animation.MechModel,
		EquippedOn:     animation.EquippedOn,
		Tier:           animation.Tier,
		IntroAnimation: animation.IntroAnimation,
		OutroAnimation: animation.OutroAnimation,
		CreatedAt:      animation.CreatedAt,
	}
}

func ServerPowerCoresToApiV1(items []*server.PowerCore) []*PowerCore {
	var converted []*PowerCore
	for _, i := range items {
		converted = append(converted, ServerPowerCoreToApiV1(i))
	}
	return converted
}

func ServerPowerCoreToApiV1(ec *server.PowerCore) *PowerCore {
	return &PowerCore{
		CollectionDetails: &CollectionDetails{
			CollectionSlug: ec.CollectionDetails.CollectionSlug,
			Hash:           ec.CollectionDetails.Hash,
			TokenID:        ec.CollectionDetails.TokenID,
			ItemType:       ec.CollectionDetails.ItemType,
			ItemID:         ec.CollectionDetails.ItemID,
			Tier:           ec.CollectionDetails.Tier,
			OwnerID:        ec.CollectionDetails.OwnerID,
			OnChainStatus:  ec.CollectionDetails.OnChainStatus,

			ImageURL:         ec.CollectionDetails.ImageURL,
			CardAnimationURL: ec.CollectionDetails.CardAnimationURL,
			AvatarURL:        ec.CollectionDetails.AvatarURL,
			LargeImageURL:    ec.CollectionDetails.LargeImageURL,
			BackgroundColor:  ec.CollectionDetails.BackgroundColor,
			AnimationURL:     ec.CollectionDetails.AnimationURL,
			YoutubeURL:       ec.CollectionDetails.YoutubeURL,
		},
		ID:           ec.ID,
		OwnerID:      ec.OwnerID,
		Label:        ec.Label,
		Size:         ec.Size,
		Capacity:     ec.Capacity,
		MaxDrawRate:  ec.MaxDrawRate,
		RechargeRate: ec.RechargeRate,
		Armour:       ec.Armour,
		MaxHitpoints: ec.MaxHitpoints,
		Tier:         ec.Tier,
		EquippedOn:   ec.EquippedOn,
		CreatedAt:    ec.CreatedAt,
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
			ItemType:       weapon.CollectionDetails.ItemType,
			ItemID:         weapon.CollectionDetails.ItemID,
			Tier:           weapon.CollectionDetails.Tier,
			OwnerID:        weapon.CollectionDetails.OwnerID,
			OnChainStatus:  weapon.CollectionDetails.OnChainStatus,

			ImageURL:         weapon.CollectionDetails.ImageURL,
			CardAnimationURL: weapon.CollectionDetails.CardAnimationURL,
			AvatarURL:        weapon.CollectionDetails.AvatarURL,
			LargeImageURL:    weapon.CollectionDetails.LargeImageURL,
			BackgroundColor:  weapon.CollectionDetails.BackgroundColor,
			AnimationURL:     weapon.CollectionDetails.AnimationURL,
			YoutubeURL:       weapon.CollectionDetails.YoutubeURL,
		},
		ID:                  weapon.ID,
		BrandID:             weapon.BrandID,
		Label:               weapon.Label,
		Slug:                weapon.Slug,
		Damage:              weapon.Damage,
		BlueprintID:         weapon.BlueprintID,
		DefaultDamageType:   weapon.DefaultDamageType,
		GenesisTokenID:      weapon.GenesisTokenID,
		WeaponType:          weapon.WeaponType,
		OwnerID:             weapon.OwnerID,
		DamageFalloff:       weapon.DamageFalloff,
		DamageFalloffRate:   weapon.DamageFalloffRate,
		Spread:              weapon.Spread,
		RateOfFire:          weapon.RateOfFire,
		Radius:              weapon.Radius,
		RadiusDamageFalloff: weapon.RadiusDamageFalloff,
		ProjectileSpeed:     weapon.ProjectileSpeed,
		EnergyCost:          weapon.EnergyCost,
		MaxAmmo:             weapon.MaxAmmo,
		UpdatedAt:           weapon.UpdatedAt,
		CreatedAt:           weapon.CreatedAt,
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
			ItemType:       ec.CollectionDetails.ItemType,
			ItemID:         ec.CollectionDetails.ItemID,
			Tier:           ec.CollectionDetails.Tier,
			OwnerID:        ec.CollectionDetails.OwnerID,
			OnChainStatus:  ec.CollectionDetails.OnChainStatus,

			ImageURL:         ec.CollectionDetails.ImageURL,
			CardAnimationURL: ec.CollectionDetails.CardAnimationURL,
			AvatarURL:        ec.CollectionDetails.AvatarURL,
			LargeImageURL:    ec.CollectionDetails.LargeImageURL,
			BackgroundColor:  ec.CollectionDetails.BackgroundColor,
			AnimationURL:     ec.CollectionDetails.AnimationURL,
			YoutubeURL:       ec.CollectionDetails.YoutubeURL,
		},
		ID:             ec.ID,
		BrandID:        ec.BrandID,
		Label:          ec.Label,
		UpdatedAt:      ec.UpdatedAt,
		CreatedAt:      ec.CreatedAt,
		BlueprintID:    ec.BlueprintID,
		GenesisTokenID: ec.GenesisTokenID,
		OwnerID:        ec.OwnerID,
		EquippedOn:     ec.EquippedOn,
		Type:           ec.Type,
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
	m := &Mech{
		CollectionDetails: &CollectionDetails{
			CollectionSlug: mech.CollectionDetails.CollectionSlug,
			Hash:           mech.CollectionDetails.Hash,
			TokenID:        mech.CollectionDetails.TokenID,
			ItemType:       mech.CollectionDetails.ItemType,
			ItemID:         mech.CollectionDetails.ItemID,
			Tier:           mech.CollectionDetails.Tier,
			OwnerID:        mech.CollectionDetails.OwnerID,
			OnChainStatus:  mech.CollectionDetails.OnChainStatus,

			ImageURL:         mech.CollectionDetails.ImageURL,
			CardAnimationURL: mech.CollectionDetails.CardAnimationURL,
			AvatarURL:        mech.CollectionDetails.AvatarURL,
			LargeImageURL:    mech.CollectionDetails.LargeImageURL,
			BackgroundColor:  mech.CollectionDetails.BackgroundColor,
			AnimationURL:     mech.CollectionDetails.AnimationURL,
			YoutubeURL:       mech.CollectionDetails.YoutubeURL,
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
		GenesisTokenID:       mech.GenesisTokenID,
		OwnerID:              mech.OwnerID,
		FactionID:            mech.FactionID,
		PowerCoreSize:        mech.PowerCoreSize,
		Tier:                 mech.Tier,
		DefaultChassisSkinID: mech.DefaultChassisSkinID,
		DefaultChassisSkin:   ServerBlueprintMechSkinToApiV1(mech.DefaultChassisSkin),
		ChassisSkinID:        mech.ChassisSkinID,
		IntroAnimationID:     mech.IntroAnimationID,
		OutroAnimationID:     mech.OutroAnimationID,
		PowerCoreID:          mech.PowerCoreID,
		UpdatedAt:            mech.UpdatedAt,
		CreatedAt:            mech.CreatedAt,
	}

	// nullables
	if mech.PowerCore != nil {
		m.PowerCore = ServerPowerCoreToApiV1(mech.PowerCore)
	}
	if mech.Weapons != nil {
		m.Weapons = ServerWeaponsToApiV1(mech.Weapons)
	}
	if mech.Utility != nil {
		m.Utility = ServerUtilitiesToApiV1(mech.Utility)
	}
	if mech.OutroAnimation != nil {
		m.OutroAnimation = ServerMechAnimationToApiV1(mech.OutroAnimation)
	}
	if mech.IntroAnimation != nil {
		m.IntroAnimation = ServerMechAnimationToApiV1(mech.IntroAnimation)
	}
	if mech.ChassisSkin != nil {
		m.ChassisSkin = ServerMechSkinToApiV1(mech.ChassisSkin)
	}

	return m
}

func ServerMechsToXsynAsset(mechs []*server.Mech) []*XsynAsset {
	var assets []*XsynAsset
	for _, i := range mechs {
		asJson, err := json.Marshal(i)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to convert item to json")
			continue
		}

		// convert stats to attributes to
		attributes := []*Attribute{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				DisplayType: "Number",
				TraitType:   "Weapon Hardpoints",
				Value:       i.WeaponHardpoints,
			},
			{
				DisplayType: "Number",
				TraitType:   "Utility Slots",
				Value:       i.UtilitySlots,
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "speed",
				Value:       i.Speed,
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Hit Points",
				Value:       i.MaxHitpoints,
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Power Core Size",
				Value:       i.PowerCoreSize,
			},
		}

		assets = append(assets, &XsynAsset{
			ID:             i.ID,
			Name:           i.Name,
			CollectionSlug: i.CollectionSlug,
			TokenID:        i.TokenID,
			Tier:           i.Tier,
			Hash:           i.Hash,
			OwnerID:        i.OwnerID,
			Data:           asJson,

			Attributes:      attributes,
			ImageURL:        i.ImageURL,
			BackgroundColor: i.BackgroundColor,
			AnimationURL:    i.AnimationURL,
			YoutubeURL:      i.YoutubeURL,
			OnChainStatus:   i.OnChainStatus,
			//XsynLocked: i.XsynLocked, // TODO: add a way for gameserver to see if they have the lock status of the asset
		})
	}

	return assets
}

func ServerMechAnimationsToXsynAsset(mechAnimations []*server.MechAnimation) []*XsynAsset {
	var assets []*XsynAsset
	for _, i := range mechAnimations {
		asJson, err := json.Marshal(i)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to convert item to json")
			continue
		}

		// convert stats to attributes to
		attributes := []*Attribute{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				TraitType: "Intro Animation",
				Value:     i.IntroAnimation.Bool,
			},
			{
				TraitType: "Outro Animation",
				Value:     i.IntroAnimation.Bool,
			},
		}

		assets = append(assets, &XsynAsset{
			ID:              i.ID,
			CollectionSlug:  i.CollectionSlug,
			TokenID:         i.TokenID,
			Tier:            i.Tier,
			Hash:            i.Hash,
			OwnerID:         i.OwnerID,
			Data:            asJson,
			Name:            i.Label,
			Attributes:      attributes,
			ImageURL:        i.ImageURL,
			BackgroundColor: i.BackgroundColor,
			AnimationURL:    i.AnimationURL,
			YoutubeURL:      i.YoutubeURL,
			OnChainStatus:   i.OnChainStatus,
			//XsynLocked: i.XsynLocked, // TODO: add a way for gameserver to see if they have the lock status of the asset
		})
	}

	return assets
}

func ServerMechSkinsToXsynAsset(mechSkins []*server.MechSkin) []*XsynAsset {
	var assets []*XsynAsset
	for _, i := range mechSkins {
		asJson, err := json.Marshal(i)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to convert item to json")
			continue
		}

		// convert stats to attributes to
		attributes := []*Attribute{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				TraitType: "Mech Model",
				Value:     i.MechModel, // TODO: get mech model name instead
			},
		}

		assets = append(assets, &XsynAsset{
			ID:               i.ID,
			CollectionSlug:   i.CollectionSlug,
			TokenID:          i.TokenID,
			Tier:             i.Tier,
			Hash:             i.Hash,
			OwnerID:          i.OwnerID,
			Data:             asJson,
			Name:             i.Label,
			Attributes:       attributes,
			ImageURL:         i.ImageURL,
			AnimationURL:     i.AnimationURL,
			LargeImageURL:    i.LargeImageURL,
			CardAnimationURL: i.CardAnimationURL,
			AvatarURL:        i.AvatarURL,
			BackgroundColor:  i.BackgroundColor,
			YoutubeURL:       i.YoutubeURL,
			OnChainStatus:    i.OnChainStatus,
			//XsynLocked: i.XsynLocked, // TODO: add a way for gameserver to see if they have the lock status of the asset
		})
	}

	return assets
}

func ServerPowerCoresToXsynAsset(powerCore []*server.PowerCore) []*XsynAsset {
	var assets []*XsynAsset
	for _, i := range powerCore {
		asJson, err := json.Marshal(i)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to convert item to json")
			continue
		}

		// convert stats to attributes to
		attributes := []*Attribute{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				TraitType: "Size",
				Value:     i.Size,
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Capacity",
				Value:       i.Capacity.InexactFloat64(),
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Max draw rate",
				Value:       i.MaxDrawRate.InexactFloat64(),
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Recharge rate",
				Value:       i.RechargeRate.InexactFloat64(),
			},
		}

		assets = append(assets, &XsynAsset{
			ID:              i.ID,
			CollectionSlug:  i.CollectionSlug,
			TokenID:         i.TokenID,
			Tier:            i.Tier,
			Hash:            i.Hash,
			OwnerID:         i.OwnerID,
			Data:            asJson,
			Name:            i.Label,
			Attributes:      attributes,
			ImageURL:        i.ImageURL,
			BackgroundColor: i.BackgroundColor,
			AnimationURL:    i.AnimationURL,
			YoutubeURL:      i.YoutubeURL,
			OnChainStatus:   i.OnChainStatus,
			//XsynLocked: i.XsynLocked, // TODO: add a way for gameserver to see if they have the lock status of the asset
		})

	}

	return assets
}

func ServerWeaponsToXsynAsset(weapons []*server.Weapon) []*XsynAsset {
	var assets []*XsynAsset
	for _, i := range weapons {
		asJson, err := json.Marshal(i)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to convert item to json")
			continue
		}
		// TODO create these dynamically depending on weapon type
		attributes := []*Attribute{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Damage",
				Value:       i.Damage,
			},
			{
				TraitType: "Damage Type",
				Value:     i.DefaultDamageType,
			},
			{
				TraitType: "Weapon Type",
				Value:     i.WeaponType,
			},
			{
				TraitType: "Damage Falloff",
				Value:     i.DamageFalloff.Int,
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Damage Falloff rate",
				Value:       i.DamageFalloffRate.Int,
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Area of effect",
				Value:       i.Radius.Int,
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Spread",
				Value:       i.Spread.Decimal.InexactFloat64(),
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Rate of fire",
				Value:       i.RateOfFire.Decimal.InexactFloat64(),
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Projectile Speed",
				Value:       i.ProjectileSpeed.Decimal.InexactFloat64(),
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Energy Cost",
				Value:       i.EnergyCost.Decimal.InexactFloat64(),
			},
			{
				DisplayType: "BoostNumber",
				TraitType:   "Max Ammo",
				Value:       i.MaxAmmo.Int,
			},
			{
				TraitType: "Tier",
				Value:     i.Tier,
			},
		}

		assets = append(assets, &XsynAsset{
			ID:              i.ID,
			CollectionSlug:  i.CollectionSlug,
			TokenID:         i.TokenID,
			Tier:            i.Tier,
			Hash:            i.Hash,
			OwnerID:         i.OwnerID,
			Data:            asJson,
			Name:            i.Label,
			Attributes:      attributes,
			ImageURL:        i.ImageURL,
			BackgroundColor: i.BackgroundColor,
			AnimationURL:    i.AnimationURL,
			YoutubeURL:      i.YoutubeURL,
			OnChainStatus:   i.OnChainStatus,
			//XsynLocked: i.XsynLocked, // TODO: add a way for gameserver to see if they have the lock status of the asset
		})
	}

	return assets
}

func ServerUtilitiesToXsynAsset(utils []*server.Utility) []*XsynAsset {
	var assets []*XsynAsset
	for _, i := range utils {
		asJson, err := json.Marshal(i)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to convert item to json")
			continue
		}

		// TODO create these dynamically depending on utility type
		attributes := []*Attribute{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				TraitType: "Type",
				Value:     i.Type,
			},
		}
		assets = append(assets, &XsynAsset{
			ID:              i.ID,
			CollectionSlug:  i.CollectionSlug,
			TokenID:         i.TokenID,
			Tier:            i.Tier,
			Hash:            i.Hash,
			OwnerID:         i.OwnerID,
			Data:            asJson,
			Name:            i.Label,
			Attributes:      attributes,
			ImageURL:        i.ImageURL,
			BackgroundColor: i.BackgroundColor,
			AnimationURL:    i.AnimationURL,
			YoutubeURL:      i.YoutubeURL,
			OnChainStatus:   i.OnChainStatus,
			//XsynLocked: i.XsynLocked, // TODO: add a way for gameserver to see if they have the lock status of the asset
		})
	}

	return assets
}
