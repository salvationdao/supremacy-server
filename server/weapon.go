package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

// Weapon is the struct that rpc expects for weapons
type Weapon struct {
	*CollectionItem
	*Images
	CollectionItemID      string              `json:"collection_item_id"`
	ID                    string              `json:"id"`
	Label                 string              `json:"label"`
	Slug                  string              `json:"slug"`
	Damage                int                 `json:"damage"`
	BlueprintID           string              `json:"blueprint_id"`
	EquippedOn            null.String         `json:"equipped_on,omitempty"`
	DefaultDamageType     string              `json:"default_damage_type"`
	GenesisTokenID        null.Int64          `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64          `json:"limited_release_token_id,omitempty"`
	WeaponType            string              `json:"weapon_type"`
	DamageFalloff         null.Int            `json:"damage_falloff,omitempty"`
	DamageFalloffRate     null.Int            `json:"damage_falloff_rate,omitempty"`
	Spread                decimal.NullDecimal `json:"spread,omitempty"`
	RateOfFire            decimal.NullDecimal `json:"rate_of_fire,omitempty"`
	Radius                null.Int            `json:"radius,omitempty"`
	RadiusDamageFalloff   null.Int            `json:"radius_damage_falloff,omitempty"`
	ProjectileSpeed       decimal.NullDecimal `json:"projectile_speed,omitempty"`
	EnergyCost            decimal.NullDecimal `json:"energy_cost,omitempty"`
	MaxAmmo               null.Int            `json:"max_ammo,omitempty"`
	EquippedWeaponSkinID  string              `json:"equipped_weapon_skin_id,omitempty"`
	WeaponSkin            *WeaponSkin         `json:"weapon_skin,omitempty"`
	ItemSaleID            null.String         `json:"item_sale_id,omitempty"`
	WeaponModelID         string              `json:"weapon_model_id,omitempty"`
	LockedToMech          bool                `json:"locked_to_mech"`
	SlotNumber            int                 `json:"slot_number"`

	// TODO: AMMO //BlueprintAmmo []*
	EquippedOnDetails *EquippedOnDetails

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (b *Weapon) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type BlueprintWeapon struct {
	ID                  string              `json:"id"`
	Label               string              `json:"label"`
	Slug                string              `json:"slug"`
	Damage              int                 `json:"damage"`
	UpdatedAt           time.Time           `json:"updated_at"`
	CreatedAt           time.Time           `json:"created_at"`
	GameClientWeaponID  null.String         `json:"game_client_weapon_id,omitempty"`
	WeaponType          string              `json:"weapon_type"`
	DefaultDamageType   string              `json:"default_damage_type"`
	DamageFalloff       null.Int            `json:"damage_falloff,omitempty"`
	DamageFalloffRate   null.Int            `json:"damage_falloff_rate,omitempty"`
	Spread              decimal.NullDecimal `json:"spread,omitempty"`
	RateOfFire          decimal.NullDecimal `json:"rate_of_fire,omitempty"`
	Radius              null.Int            `json:"radius,omitempty"`
	RadiusDamageFalloff null.Int            `json:"radius_damage_falloff,omitempty"`
	ProjectileSpeed     decimal.NullDecimal `json:"projectile_speed,omitempty"`
	MaxAmmo             null.Int            `json:"max_ammo,omitempty"`
	EnergyCost          decimal.NullDecimal `json:"energy_cost,omitempty"`
	Collection          string              `json:"collection"`
	Tier                string              `json:"tier,omitempty"`
	WeaponModelID       string              `json:"weapon_model_id"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        null.Int64 `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64 `json:"limited_release_token_id,omitempty"`
}

func (b *BlueprintWeapon) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type WeaponSlice []*Weapon

func (b *WeaponSlice) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

func BlueprintWeaponFromBoiler(weapon *boiler.BlueprintWeapon) *BlueprintWeapon {
	return &BlueprintWeapon{
		ID:                  weapon.ID,
		Label:               weapon.Label,
		Slug:                weapon.Slug,
		UpdatedAt:           weapon.UpdatedAt,
		CreatedAt:           weapon.CreatedAt,
		Damage:              weapon.Damage,
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
		Collection:          weapon.Collection,
		Tier:                weapon.Tier,
		WeaponModelID:       weapon.WeaponModelID,
	}
}

func WeaponFromBoiler(weapon *boiler.Weapon, collection *boiler.CollectionItem, weaponSkin *WeaponSkin, itemSaleID null.String) *Weapon {
	return &Weapon{
		CollectionItem: &CollectionItem{
			CollectionSlug: collection.CollectionSlug,
			Hash:           collection.Hash,
			TokenID:        collection.TokenID,
			ItemType:       collection.ItemType,
			ItemID:         collection.ItemID,
			Tier:           collection.Tier,
			OwnerID:        collection.OwnerID,
			MarketLocked:   collection.MarketLocked,
			XsynLocked:     collection.XsynLocked,
			AssetHidden:    collection.AssetHidden,
		},
		Images: &Images{
			ImageURL:         weaponSkin.ImageURL,
			CardAnimationURL: weaponSkin.CardAnimationURL,
			AvatarURL:        weaponSkin.AvatarURL,
			LargeImageURL:    weaponSkin.LargeImageURL,
			BackgroundColor:  weaponSkin.BackgroundColor,
			AnimationURL:     weaponSkin.AnimationURL,
			YoutubeURL:       weaponSkin.YoutubeURL,
		},
		CollectionItemID:     collection.ID,
		ID:                   weapon.ID,
		Label:                weapon.R.Blueprint.Label,
		Slug:                 weapon.Slug,
		Damage:               weapon.Damage,
		BlueprintID:          weapon.BlueprintID,
		DefaultDamageType:    weapon.DefaultDamageType,
		GenesisTokenID:       weapon.GenesisTokenID,
		WeaponType:           weapon.R.Blueprint.WeaponType,
		DamageFalloff:        weapon.DamageFalloff,
		DamageFalloffRate:    weapon.DamageFalloffRate,
		Spread:               weapon.Spread,
		RateOfFire:           weapon.RateOfFire,
		Radius:               weapon.Radius,
		RadiusDamageFalloff:  weapon.RadiusDamageFalloff,
		ProjectileSpeed:      weapon.ProjectileSpeed,
		EnergyCost:           weapon.EnergyCost,
		MaxAmmo:              weapon.MaxAmmo,
		UpdatedAt:            weapon.UpdatedAt,
		CreatedAt:            weapon.CreatedAt,
		EquippedOn:           weapon.EquippedOn,
		EquippedWeaponSkinID: weapon.EquippedWeaponSkinID,
		WeaponSkin:           weaponSkin,
		ItemSaleID:           itemSaleID,
	}
}
