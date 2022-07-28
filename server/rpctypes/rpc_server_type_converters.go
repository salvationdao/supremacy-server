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
			ID:             i.ID,
			Name:           i.Label,
			CollectionSlug: i.CollectionSlug,
			TokenID:        i.TokenID,
			Hash:           i.Hash,
			OwnerID:        i.OwnerID,
			Data:           asJson,
			AssetType:      null.StringFrom(i.ItemType),
		}

		if isGenesisOrLimited && i.ChassisSkin != nil {
			asset.Tier = i.ChassisSkin.Tier
		}

		// convert stats to attributes to
		asset.Attributes = []*Attribute{
			{
				TraitType: "Type",
				Value:     fmt.Sprintf("WAR MACHINE"),
			},
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
				TraitType: "Power Core Size",
				Value:     i.PowerCoreSize,
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
					AssetHash: i.ChassisSkin.Hash,
				},
				&Attribute{
					TraitType: "Rarity",
					Value:     i.ChassisSkin.Tier,
				},
			)
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
				TraitType: "Type",
				Value:     fmt.Sprintf("WAR MACHINE ANIMATION"),
			},
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
				TraitType: "Rarity",
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

			Name:       i.Label,
			Attributes: attributes,
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
				TraitType: "Type",
				Value:     fmt.Sprintf("WAR MACHINE SUBMODEL"),
			},
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			// TODO: vinnie fix me
			//{
			//	TraitType: "Mech Model",
			//	Value:     i.MechModelName,
			//},
			{
				TraitType: "Rarity",
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
			ID:             i.ID,
			CollectionSlug: i.CollectionSlug,
			TokenID:        i.TokenID,
			Tier:           i.Tier,
			Hash:           i.Hash,
			OwnerID:        i.OwnerID,
			Data:           asJson,
			Name:           i.Label,
			Attributes:     attributes,
			AssetType:      null.StringFrom(i.ItemType),
			XsynLocked:     i.XsynLocked,
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
				TraitType: "Type",
				Value:     fmt.Sprintf("POWER CORE"),
			},
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
			ID:             i.ID,
			CollectionSlug: i.CollectionSlug,
			TokenID:        i.TokenID,
			Hash:           i.Hash,
			OwnerID:        i.OwnerID,
			AssetType:      null.StringFrom(i.ItemType),
			Data:           asJson,
			Name:           i.Label,
			Attributes:     attributes,
			XsynLocked:     i.XsynLocked,
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
				TraitType: "Type",
				Value:     fmt.Sprintf("WEAPON"),
			},
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

		asset := &XsynAsset{
			ID:             i.ID,
			CollectionSlug: i.CollectionSlug,
			TokenID:        i.TokenID,
			Hash:           i.Hash,
			OwnerID:        i.OwnerID,
			AssetType:      null.StringFrom(i.ItemType),
			Data:           asJson,
			Name:           i.Label,
			Attributes:     attributes,
			XsynLocked:     i.XsynLocked,
		}

		if i.WeaponSkin == nil {
			i.WeaponSkin, err = db.WeaponSkin(gamedb.StdConn, i.EquippedWeaponSkinID)
			if err != nil {
				gamelog.L.Error().Err(err).Str("i.EquippedWeaponSkinID.String", i.EquippedWeaponSkinID).Msg("failed to get weapon skin item")
				continue
			}
		}

		//asset.ImageURL = i.WeaponSkin.ImageURL
		//asset.BackgroundColor = i.WeaponSkin.BackgroundColor
		//asset.AnimationURL = i.WeaponSkin.AnimationURL
		//asset.YoutubeURL = i.WeaponSkin.YoutubeURL
		//asset.AvatarURL = i.WeaponSkin.AvatarURL
		//asset.CardAnimationURL = i.WeaponSkin.CardAnimationURL

		asset.Attributes = append(asset.Attributes,
			&Attribute{
				TraitType: "Submodel",
				Value:     i.WeaponSkin.Label,
				AssetHash: i.WeaponSkin.Hash,
			},
			&Attribute{
				TraitType: "Rarity",
				Value:     i.WeaponSkin.Tier,
			})

		err = attributes.AreValid()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("invalid asset attributes")
		}

		assets = append(assets, asset)
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
				TraitType: "Type",
				Value:     fmt.Sprintf("WEAPON SUBMODEL"),
			},
			{
				TraitType: "Label",
				Value:     i.Label,
			},
			{
				TraitType: "Rarity",
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
			ID:             i.ID,
			CollectionSlug: i.CollectionSlug,
			TokenID:        i.TokenID,
			Tier:           i.Tier,
			Hash:           i.Hash,
			OwnerID:        i.OwnerID,
			Data:           asJson,
			Name:           i.Label,
			Attributes:     attributes,
			AssetType:      null.StringFrom(i.ItemType),
			XsynLocked:     i.XsynLocked,
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
				TraitType: "Type",
				Value:     fmt.Sprintf("WAR MACHINE UTILITY"),
			},
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
			ID:             i.ID,
			CollectionSlug: i.CollectionSlug,
			TokenID:        i.TokenID,
			Hash:           i.Hash,
			OwnerID:        i.OwnerID,
			AssetType:      null.StringFrom(i.ItemType),
			Data:           asJson,
			Name:           i.Label,
			Attributes:     attributes,
			XsynLocked:     i.XsynLocked,
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
			Value:     fmt.Sprintf("NEXUS %s MYSTERY CRATE", mysteryCrate.Type),
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
		ID:             mysteryCrate.ID,
		CollectionSlug: mysteryCrate.CollectionSlug,
		TokenID:        mysteryCrate.TokenID,
		Data:           asJson,
		Attributes:     attributes,
		Hash:           mysteryCrate.Hash,
		OwnerID:        mysteryCrate.OwnerID,
		AssetType:      null.StringFrom(mysteryCrate.ItemType),
		Name:           mysteryCrate.Label,
	}

	return asset
}
