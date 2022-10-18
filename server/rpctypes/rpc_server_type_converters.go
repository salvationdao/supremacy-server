package rpctypes

import (
	"encoding/json"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
			{
				DisplayType: BoostNumber,
				TraitType:   "Shield Hit Points",
				Value:       i.Shield,
			},
			{
				DisplayType: BoostNumber,
				TraitType:   "Shield Recharge Rate",
				Value:       i.ShieldRechargeRate,
			},
			{
				DisplayType: BoostNumber,
				TraitType:   "Shield Recharge Power Cost",
				Value:       i.ShieldRechargePowerCost,
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

		if i.ChassisSkin == nil {
			i.ChassisSkin, err = db.MechSkin(gamedb.StdConn, i.ChassisSkinID, &i.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Str("i.ChassisSkinID.String", i.ChassisSkinID).Msg("failed to get mech skin item")
				continue
			}
			asset.ImageURL = i.ChassisSkin.Images.ImageURL
			asset.BackgroundColor = i.ChassisSkin.Images.BackgroundColor
			asset.AnimationURL = i.ChassisSkin.Images.AnimationURL
			asset.YoutubeURL = i.ChassisSkin.Images.YoutubeURL
			asset.AvatarURL = i.ChassisSkin.Images.AvatarURL
			asset.CardAnimationURL = i.ChassisSkin.Images.CardAnimationURL
		} else {
			asset.ImageURL = i.Images.ImageURL
			asset.BackgroundColor = i.Images.BackgroundColor
			asset.AnimationURL = i.Images.AnimationURL
			asset.YoutubeURL = i.Images.YoutubeURL
			asset.AvatarURL = i.Images.AvatarURL
			asset.CardAnimationURL = i.Images.CardAnimationURL
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

		if isGenesisOrLimited {
			if i.GenesisTokenID.Valid {
				asset.TokenID = i.GenesisTokenID.Int64
				asset.CollectionSlug = "supremacy-genesis"
			}
			if i.LimitedReleaseTokenID.Valid {
				asset.TokenID = i.LimitedReleaseTokenID.Int64
				asset.CollectionSlug = "supremacy-limited-release"
			}

			if i.ChassisSkin != nil {
				asset.Tier = i.ChassisSkin.Tier
			}
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

func ServerMechSkinsToXsynAsset(tx boil.Executor, mechSkins []*server.MechSkin) []*XsynAsset {
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
			{
				TraitType: "Rarity",
				Value:     i.Tier,
			},
		}

		if i.EquippedOn.Valid {
			if i.EquippedOnDetails == nil {
				// make db call
				i.EquippedOnDetails, err = db.MechEquippedOnDetails(tx, i.EquippedOn.String)
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

		asset := &XsynAsset{
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
			ImageURL:         i.Images.ImageURL,
			AnimationURL:     i.Images.AnimationURL,
			LargeImageURL:    i.Images.LargeImageURL,
			CardAnimationURL: i.Images.CardAnimationURL,
			AvatarURL:        i.Images.AvatarURL,
			BackgroundColor:  i.Images.BackgroundColor,
			YoutubeURL:       i.Images.YoutubeURL,
		}

		if i.SkinSwatch != nil {
			asset.ImageURL = i.SkinSwatch.ImageURL
			asset.AnimationURL = i.SkinSwatch.AnimationURL
			asset.LargeImageURL = i.SkinSwatch.LargeImageURL
			asset.CardAnimationURL = i.SkinSwatch.CardAnimationURL
			asset.AvatarURL = i.SkinSwatch.AvatarURL
			asset.BackgroundColor = i.SkinSwatch.BackgroundColor
			asset.YoutubeURL = i.SkinSwatch.YoutubeURL
		}

		assets = append(assets, asset)
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
				i.EquippedOnDetails, err = db.MechEquippedOnDetails(gamedb.StdConn, i.EquippedOn.String)
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
			ImageURL:         i.Images.ImageURL,
			AnimationURL:     i.Images.AnimationURL,
			LargeImageURL:    i.Images.LargeImageURL,
			CardAnimationURL: i.Images.CardAnimationURL,
			AvatarURL:        i.Images.AvatarURL,
			BackgroundColor:  i.Images.BackgroundColor,
			YoutubeURL:       i.Images.YoutubeURL,
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

		if i.PowerCost.Valid && !i.PowerCost.Decimal.IsZero() {
			attributes = append(attributes, &Attribute{
				DisplayType: BoostNumber,
				TraitType:   "Energy Cost",
				Value:       i.PowerCost.Decimal.InexactFloat64(),
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
				i.EquippedOnDetails, err = db.MechEquippedOnDetails(gamedb.StdConn, i.EquippedOn.String)
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
			i.WeaponSkin, err = db.WeaponSkin(gamedb.StdConn, i.EquippedWeaponSkinID, &i.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Str("i.EquippedWeaponSkinID.String", i.EquippedWeaponSkinID).Msg("failed to get weapon skin item")
				continue
			}
			asset.ImageURL = i.WeaponSkin.Images.ImageURL
			asset.AnimationURL = i.WeaponSkin.Images.AnimationURL
			asset.LargeImageURL = i.WeaponSkin.Images.LargeImageURL
			asset.CardAnimationURL = i.WeaponSkin.Images.CardAnimationURL
			asset.AvatarURL = i.WeaponSkin.Images.AvatarURL
			asset.BackgroundColor = i.WeaponSkin.Images.BackgroundColor
			asset.YoutubeURL = i.WeaponSkin.Images.YoutubeURL

		} else {
			asset.ImageURL = i.Images.ImageURL
			asset.AnimationURL = i.Images.AnimationURL
			asset.LargeImageURL = i.Images.LargeImageURL
			asset.CardAnimationURL = i.Images.CardAnimationURL
			asset.AvatarURL = i.Images.AvatarURL
			asset.BackgroundColor = i.Images.BackgroundColor
			asset.YoutubeURL = i.Images.YoutubeURL
		}

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

func ServerWeaponSkinsToXsynAsset(tx boil.Executor, weaponSkins []*server.WeaponSkin) []*XsynAsset {
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
				i.EquippedOnDetails, err = db.WeaponEquippedOnDetails(tx, i.EquippedOn.String)
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

		asset := &XsynAsset{
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
			ImageURL:         i.Images.ImageURL,
			AnimationURL:     i.Images.AnimationURL,
			LargeImageURL:    i.Images.LargeImageURL,
			CardAnimationURL: i.Images.CardAnimationURL,
			AvatarURL:        i.Images.AvatarURL,
			BackgroundColor:  i.Images.BackgroundColor,
			YoutubeURL:       i.Images.YoutubeURL,
		}

		if i.SkinSwatch != nil {
			asset.ImageURL = i.SkinSwatch.ImageURL
			asset.AnimationURL = i.SkinSwatch.AnimationURL
			asset.LargeImageURL = i.SkinSwatch.LargeImageURL
			asset.CardAnimationURL = i.SkinSwatch.CardAnimationURL
			asset.AvatarURL = i.SkinSwatch.AvatarURL
			asset.BackgroundColor = i.SkinSwatch.BackgroundColor
			asset.YoutubeURL = i.SkinSwatch.YoutubeURL
		}

		assets = append(assets, asset)
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
				i.EquippedOnDetails, err = db.MechEquippedOnDetails(gamedb.StdConn, i.EquippedOn.String)
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
			ImageURL:         i.Images.ImageURL,
			AnimationURL:     i.Images.AnimationURL,
			LargeImageURL:    i.Images.LargeImageURL,
			CardAnimationURL: i.Images.CardAnimationURL,
			AvatarURL:        i.Images.AvatarURL,
			BackgroundColor:  i.Images.BackgroundColor,
			YoutubeURL:       i.Images.YoutubeURL,
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
		ID:               mysteryCrate.ID,
		CollectionSlug:   mysteryCrate.CollectionSlug,
		TokenID:          mysteryCrate.TokenID,
		Data:             asJson,
		Attributes:       attributes,
		Hash:             mysteryCrate.Hash,
		OwnerID:          mysteryCrate.OwnerID,
		AssetType:        null.StringFrom(mysteryCrate.ItemType),
		Name:             mysteryCrate.Label,
		ImageURL:         mysteryCrate.Images.ImageURL,
		BackgroundColor:  mysteryCrate.Images.BackgroundColor,
		AnimationURL:     mysteryCrate.Images.AnimationURL,
		YoutubeURL:       mysteryCrate.Images.YoutubeURL,
		AvatarURL:        mysteryCrate.Images.AvatarURL,
		CardAnimationURL: mysteryCrate.Images.CardAnimationURL,
	}

	return asset
}
