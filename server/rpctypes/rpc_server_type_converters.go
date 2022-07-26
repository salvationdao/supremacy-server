package rpctypes

import (
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/gamedb"
	"server/gamelog"

	"github.com/volatiletech/null/v8"
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
			CollectionSlug:   skin.CollectionItem.CollectionSlug,
			Hash:             skin.CollectionItem.Hash,
			TokenID:          skin.CollectionItem.TokenID,
			ItemType:         skin.CollectionItem.ItemType,
			ItemID:           skin.CollectionItem.ItemID,
			Tier:             skin.CollectionItem.Tier,
			OwnerID:          skin.CollectionItem.OwnerID,
			MarketLocked:     skin.CollectionItem.MarketLocked,
			XsynLocked:       skin.CollectionItem.XsynLocked,
		},
		ID:               skin.ID,
		BlueprintID:      skin.BlueprintID,
		GenesisTokenID:   skin.GenesisTokenID,
		Label:            skin.Label,
		OwnerID:          skin.OwnerID,
		MechModel:        skin.MechModelID,
		EquippedOn:       skin.EquippedOn,
		Tier:             skin.Tier,
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
			CollectionSlug:   animation.CollectionItem.CollectionSlug,
			Hash:             animation.CollectionItem.Hash,
			TokenID:          animation.CollectionItem.TokenID,
			ItemType:         animation.CollectionItem.ItemType,
			ItemID:           animation.CollectionItem.ItemID,
			Tier:             animation.CollectionItem.Tier,
			OwnerID:          animation.CollectionItem.OwnerID,
			MarketLocked:     animation.CollectionItem.MarketLocked,
			XsynLocked:       animation.CollectionItem.XsynLocked,
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
			CollectionSlug:   ec.CollectionItem.CollectionSlug,
			Hash:             ec.CollectionItem.Hash,
			TokenID:          ec.CollectionItem.TokenID,
			ItemType:         ec.CollectionItem.ItemType,
			ItemID:           ec.CollectionItem.ItemID,
			Tier:             ec.CollectionItem.Tier,
			OwnerID:          ec.CollectionItem.OwnerID,
			MarketLocked:     ec.CollectionItem.MarketLocked,
			XsynLocked:       ec.CollectionItem.XsynLocked,
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
			CollectionSlug:   weapon.CollectionItem.CollectionSlug,
			Hash:             weapon.CollectionItem.Hash,
			TokenID:          weapon.CollectionItem.TokenID,
			ItemType:         weapon.CollectionItem.ItemType,
			ItemID:           weapon.CollectionItem.ItemID,
			Tier:             weapon.CollectionItem.Tier,
			OwnerID:          weapon.CollectionItem.OwnerID,
			MarketLocked:     weapon.CollectionItem.MarketLocked,
			XsynLocked:       weapon.CollectionItem.XsynLocked,
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
		EquippedOn:          weapon.EquippedOn,
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
			CollectionSlug:   ec.CollectionItem.CollectionSlug,
			Hash:             ec.CollectionItem.Hash,
			TokenID:          ec.CollectionItem.TokenID,
			ItemType:         ec.CollectionItem.ItemType,
			ItemID:           ec.CollectionItem.ItemID,
			Tier:             ec.CollectionItem.Tier,
			OwnerID:          ec.CollectionItem.OwnerID,
			MarketLocked:     ec.CollectionItem.MarketLocked,
			XsynLocked:       ec.CollectionItem.XsynLocked,
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
			CollectionSlug:   mech.CollectionItem.CollectionSlug,
			Hash:             mech.CollectionItem.Hash,
			TokenID:          mech.CollectionItem.TokenID,
			ItemType:         mech.CollectionItem.ItemType,
			ItemID:           mech.CollectionItem.ItemID,
			Tier:             mech.CollectionItem.Tier,
			OwnerID:          mech.CollectionItem.OwnerID,
			MarketLocked:     mech.CollectionItem.MarketLocked,
			XsynLocked:       mech.CollectionItem.XsynLocked,
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
		isGenesisOrLimited := i.IsCompleteGenesis() || i.IsCompleteLimited()

		asJson, err := json.Marshal(i)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to convert item to json")
			continue
		}

		asset := &XsynAsset{
			ID:               i.ID,
			Name:             i.Label,
			CollectionSlug:   i.CollectionSlug,
			TokenID:          i.TokenID,
			Hash:             i.Hash,
			OwnerID:          i.OwnerID,
			Data:             asJson,
			AssetType:        null.StringFrom(i.ItemType),
		}

		if isGenesisOrLimited && i.ChassisSkin != nil {
			asset.Tier = i.ChassisSkin.Tier
		}

		// convert stats to attributes to
		asset.Attributes = []*Attribute{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				TraitType: "Name",
				Value:     i.Name,
			},
			{
				DisplayType: Number,
				TraitType:   "Weapon Hardpoints",
				Value:       i.WeaponHardpoints,
			},
			{
				DisplayType: Number,
				TraitType:   "Utility Slots",
				Value:       i.UtilitySlots,
			},
			{
				DisplayType: BoostNumber,
				TraitType:   "speed",
				Value:       i.Speed,
			},
			{
				DisplayType: BoostNumber,
				TraitType:   "Hit Points",
				Value:       i.MaxHitpoints,
			},
			{
				TraitType:   "Power Core Size",
				Value:       i.PowerCoreSize,
			},
		}

		for i, wep := range i.Weapons {
			asset.Attributes = append(asset.Attributes,
				&Attribute{
					TraitType: fmt.Sprintf("Weapon Slot %d", i+1),
					Value:     wep.Label,
					AssetHash: wep.Hash,
				})
		}

		for i, util := range i.Utility {
			asset.Attributes = append(asset.Attributes,
				&Attribute{
					TraitType: fmt.Sprintf("Utility Slot %d", i+1),
					Value:     util.Label,
					AssetHash: util.Hash,
				})
		}

		if i.PowerCore != nil {
			asset.Attributes = append(asset.Attributes,
				&Attribute{
					TraitType: "Power Core",
					Value:     i.PowerCore.Label,
					AssetHash: i.PowerCore.Hash,
				})
		}

		if i.ChassisSkinID.Valid {
			if i.ChassisSkin == nil {
				i.ChassisSkin, err = db.MechSkin(gamedb.StdConn, i.ChassisSkinID.String)
				if err != nil {
					gamelog.L.Error().Err(err).Str("i.ChassisSkinID.String", i.ChassisSkinID.String).Msg("failed to get mech skin item")
					continue
				}
			}

			asset.Attributes = append(asset.Attributes,
				&Attribute{
					TraitType: "Submodel",
					Value:     i.ChassisSkin.Label,
					AssetHash: i.Hash,
				})
		}

		err = asset.Attributes.AreValid()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("invalid asset attributes")
		}

		assets = append(assets, asset)
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
			{
				TraitType: "Tier",
				Value:     i.Tier,
			},
		}

		assets = append(assets, &XsynAsset{
			ID:             i.ID,
			CollectionSlug: i.CollectionSlug,
			TokenID:        i.TokenID,
			Tier:           i.Tier,
			Hash:           i.Hash,
			OwnerID:        i.OwnerID,
			Data:           asJson,
			AssetType:      null.StringFrom(i.ItemType),

			Name:             i.Label,
			Attributes:       attributes,
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
		attributes := Attributes{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				TraitType: "Mech Model",
				Value:     i.MechModelName,
			},
			{
				TraitType: "Tier",
				Value:     i.Tier,
			},
		}

		if i.EquippedOn.Valid {
			if i.EquippedOnDetails == nil {
				// make db call
				i.EquippedOnDetails, err = db.MechEquippedOnDetails(nil, i.EquippedOn.String)
				if err != nil {
					gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to get db.MechEquippedOnDetails")
					continue
				}
			}

			name := i.EquippedOnDetails.Name
			if name == "" {
				name = i.EquippedOnDetails.Label
			}

			attributes = append(attributes, &Attribute{
				TraitType: "Equipped On",
				Value:     i.EquippedOnDetails.Label,
				AssetHash: i.EquippedOnDetails.Hash,
			})
		}

		err = attributes.AreValid()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("invalid asset attributes")
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
			AssetType:        null.StringFrom(i.ItemType),
			XsynLocked:       i.XsynLocked,
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
		attributes := Attributes{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				TraitType: "Size",
				Value:     i.Size,
			},
			{
				DisplayType: BoostNumber,
				TraitType:   "Capacity",
				Value:       i.Capacity.InexactFloat64(),
			},
			{
				DisplayType: BoostNumber,
				TraitType:   "Max draw rate",
				Value:       i.MaxDrawRate.InexactFloat64(),
			},
			{
				DisplayType: BoostNumber,
				TraitType:   "Recharge rate",
				Value:       i.RechargeRate.InexactFloat64(),
			},
		}

		if i.EquippedOn.Valid {
			if i.EquippedOnDetails == nil {
				// make db call
				i.EquippedOnDetails, err = db.MechEquippedOnDetails(nil, i.EquippedOn.String)
				if err != nil {
					gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to get db.MechEquippedOnDetails")
					continue
				}
			}

			name := i.EquippedOnDetails.Name
			if name == "" {
				name = i.EquippedOnDetails.Label
			}

			attributes = append(attributes, &Attribute{
				TraitType: "Equipped On",
				Value:     i.EquippedOnDetails.Label,
				AssetHash: i.EquippedOnDetails.Hash,
			})
		}

		err = attributes.AreValid()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("invalid asset attributes")
		}

		assets = append(assets, &XsynAsset{
			ID:               i.ID,
			CollectionSlug:   i.CollectionSlug,
			TokenID:          i.TokenID,
			Hash:             i.Hash,
			OwnerID:          i.OwnerID,
			AssetType:        null.StringFrom(i.ItemType),
			Data:             asJson,
			Name:             i.Label,
			Attributes:       attributes,
			XsynLocked:       i.XsynLocked,
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

		attributes := Attributes{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				DisplayType: BoostNumber,
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
		}

		if i.DamageFalloff.Valid && i.DamageFalloff.Int > 0 {
			attributes = append(attributes, &Attribute{
				DisplayType: BoostNumber,
				TraitType:   "Damage Falloff",
				Value:       i.DamageFalloff.Int,
			})
		}

		if i.DamageFalloffRate.Valid && i.DamageFalloffRate.Int > 0 {
			attributes = append(attributes, &Attribute{
				DisplayType: BoostNumber,
				TraitType:   "Damage Falloff rate",
				Value:       i.DamageFalloffRate.Int,
			})
		}

		if i.Radius.Valid && i.Radius.Int > 0 {
			attributes = append(attributes, &Attribute{
				DisplayType: BoostNumber,
				TraitType:   "Area of effect",
				Value:       i.Radius.Int,
			})
		}

		if i.Spread.Valid && !i.Spread.Decimal.IsZero() {
			attributes = append(attributes, &Attribute{
				DisplayType: BoostNumber,
				TraitType:   "Spread",
				Value:       i.Spread.Decimal.InexactFloat64(),
			})
		}

		if i.RateOfFire.Valid && !i.RateOfFire.Decimal.IsZero() {
			attributes = append(attributes, &Attribute{
				DisplayType: BoostNumber,
				TraitType:   "Rate of fire",
				Value:       i.RateOfFire.Decimal.InexactFloat64(),
			})
		}

		if i.ProjectileSpeed.Valid && !i.ProjectileSpeed.Decimal.IsZero() {
			attributes = append(attributes, &Attribute{
				DisplayType: BoostNumber,
				TraitType:   "Projectile Speed",
				Value:       i.ProjectileSpeed.Decimal.InexactFloat64(),
			})
		}

		if i.EnergyCost.Valid && !i.EnergyCost.Decimal.IsZero() {
			attributes = append(attributes, &Attribute{
				DisplayType: BoostNumber,
				TraitType:   "Energy Cost",
				Value:       i.EnergyCost.Decimal.InexactFloat64(),
			})
		}

		if i.MaxAmmo.Valid && i.MaxAmmo.Int > 0 {
			attributes = append(attributes, &Attribute{
				DisplayType: BoostNumber,
				TraitType:   "Max Ammo",
				Value:       i.MaxAmmo.Int,
			})
		}

		if i.EquippedOn.Valid {
			if i.EquippedOnDetails == nil {
				// make db call
				i.EquippedOnDetails, err = db.MechEquippedOnDetails(nil, i.EquippedOn.String)
				if err != nil {
					gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to get db.MechEquippedOnDetails")
					continue
				}
			}
			name := i.EquippedOnDetails.Name
			if name == "" {
				name = i.EquippedOnDetails.Label
			}

			attributes = append(attributes, &Attribute{
				TraitType: "Equipped On",
				Value:     name,
				AssetHash: i.EquippedOnDetails.Hash,
			})
		}

		err = attributes.AreValid()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("invalid asset attributes")
		}

		assets = append(assets, &XsynAsset{
			ID:               i.ID,
			CollectionSlug:   i.CollectionSlug,
			TokenID:          i.TokenID,
			Hash:             i.Hash,
			OwnerID:          i.OwnerID,
			AssetType:        null.StringFrom(i.ItemType),
			Data:             asJson,
			Name:             i.Label,
			Attributes:       attributes,
			XsynLocked:       i.XsynLocked,
		})
	}

	return assets
}

func ServerWeaponSkinsToXsynAsset(weaponSkins []*server.WeaponSkin) []*XsynAsset {
	var assets []*XsynAsset
	for _, i := range weaponSkins {
		asJson, err := json.Marshal(i)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to convert item to json")
			continue
		}

		// convert stats to attributes to
		attributes := Attributes{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				TraitType: "Tier",
				Value:     i.Tier,
			},
		}

		if i.EquippedOn.Valid {
			if i.EquippedOnDetails == nil {
				// make db call
				i.EquippedOnDetails, err = db.WeaponEquippedOnDetails(nil, i.EquippedOn.String)
				if err != nil {
					gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to get db.MechEquippedOnDetails")
					continue
				}
			}

			name := i.EquippedOnDetails.Name
			if name == "" {
				name = i.EquippedOnDetails.Label
			}

			attributes = append(attributes, &Attribute{
				TraitType: "Equipped On",
				Value:     i.EquippedOnDetails.Label,
				AssetHash: i.EquippedOnDetails.Hash,
			})
		}

		err = attributes.AreValid()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("invalid asset attributes")
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
			AssetType:        null.StringFrom(i.ItemType),
			XsynLocked:       i.XsynLocked,
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
		attributes := Attributes{
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				TraitType: "Type",
				Value:     i.Type,
			},
		}

		if i.EquippedOn.Valid {
			if i.EquippedOnDetails == nil {
				// make db call
				i.EquippedOnDetails, err = db.MechEquippedOnDetails(nil, i.EquippedOn.String)
				if err != nil {
					gamelog.L.Error().Err(err).Interface("interface", i).Msg("failed to get db.MechEquippedOnDetails")
					continue
				}
			}

			name := i.EquippedOnDetails.Name
			if name == "" {
				name = i.EquippedOnDetails.Label
			}

			attributes = append(attributes, &Attribute{
				TraitType: "Equipped On",
				Value:     i.EquippedOnDetails.Label,
				AssetHash: i.EquippedOnDetails.Hash,
			})
		}

		err = attributes.AreValid()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("invalid asset attributes")
		}

		assets = append(assets, &XsynAsset{
			ID:               i.ID,
			CollectionSlug:   i.CollectionSlug,
			TokenID:          i.TokenID,
			Hash:             i.Hash,
			OwnerID:          i.OwnerID,
			AssetType:        null.StringFrom(i.ItemType),
			Data:             asJson,
			Name:             i.Label,
			Attributes:       attributes,
			XsynLocked:       i.XsynLocked,
		})
	}

	return assets
}

func ServerMysteryCrateToXsynAsset(mysteryCrate *server.MysteryCrate, factionName string) *XsynAsset {
	asJson, err := json.Marshal(mysteryCrate)
	if err != nil {
		gamelog.L.Error().Err(err).Interface("interface", mysteryCrate).Msg("failed to convert item to json")
	}

	// convert stats to attributes to
	attributes := Attributes{
		{
			TraitType: "Type",
			Value:     mysteryCrate.Type,
		},
	}

	if factionName != "" {
		attributes = append(attributes, &Attribute{
			TraitType: "Faction",
			Value:     factionName,
		})
	}

	err = attributes.AreValid()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("invalid asset attributes")
	}

	asset := &XsynAsset{
		ID:               mysteryCrate.ID,
		CollectionSlug:   mysteryCrate.CollectionSlug,
		TokenID:          mysteryCrate.TokenID,
		Data:             asJson,
		Attributes:       attributes,
		Hash:             mysteryCrate.Hash,
		OwnerID:          mysteryCrate.OwnerID,
		AssetType:        null.StringFrom(mysteryCrate.ItemType),
		Name:             mysteryCrate.Label,
	}

	return asset
}
