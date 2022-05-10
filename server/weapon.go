package server

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

// Weapon is the struct that rpc expects for weapons
type Weapon struct {
	*CollectionDetails
	ID                   string              `json:"id"`
	BrandID              null.String         `json:"brand_id,omitempty"`
	Label                string              `json:"label"`
	Slug                 string              `json:"slug"`
	Damage               int                 `json:"damage"`
	BlueprintID          string              `json:"blueprint_id"`
	DefaultDamageTyp     string              `json:"default_damage_typ"`
	CollectionItemID     string              `json:"collection_item_id"`
	GenesisTokenID       decimal.NullDecimal `json:"genesis_token_id,omitempty"`
	WeaponType           string              `json:"weapon_type"`
	OwnerID              string              `json:"owner_id"`
	DamageFalloff        null.Int            `json:"damage_falloff,omitempty"`
	DamageFalloffRate    null.Int            `json:"damage_falloff_rate,omitempty"`
	Spread               null.Int            `json:"spread,omitempty"`
	RateOfFire           decimal.NullDecimal `json:"rate_of_fire,omitempty"`
	Radius               null.Int            `json:"radius,omitempty"`
	RadialDoesFullDamage null.Bool           `json:"radial_does_full_damage,omitempty"`
	ProjectileSpeed      decimal.NullDecimal `json:"projectile_speed,omitempty"`
	EnergyCost           decimal.NullDecimal `json:"energy_cost,omitempty"`
	MaxAmmo              null.Int            `json:"max_ammo,omitempty"`

	//BlueprintAmmo []* // TODO: AMMO

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type BlueprintWeapon struct {
	ID                   string              `json:"id"`
	BrandID              null.String         `json:"brand_id,omitempty"`
	Label                string              `json:"label"`
	Slug                 string              `json:"slug"`
	Damage               int                 `json:"damage"`
	UpdatedAt            time.Time           `json:"updated_at"`
	CreatedAt            time.Time           `json:"created_at"`
	GameClientWeaponID   null.String         `json:"game_client_weapon_id,omitempty"`
	WeaponType           string              `json:"weapon_type"`
	DefaultDamageTyp     string              `json:"default_damage_typ"`
	DamageFalloff        null.Int            `json:"damage_falloff,omitempty"`
	DamageFalloffRate    null.Int            `json:"damage_falloff_rate,omitempty"`
	Spread               decimal.NullDecimal `json:"spread,omitempty"`
	RateOfFire           decimal.NullDecimal `json:"rate_of_fire,omitempty"`
	Radius               null.Int            `json:"radius,omitempty"`
	RadialDoesFullDamage null.Bool           `json:"radial_does_full_damage,omitempty"`
	ProjectileSpeed      null.Int            `json:"projectile_speed,omitempty"`
	MaxAmmo              null.Int            `json:"max_ammo,omitempty"`
	EnergyCost           decimal.NullDecimal `json:"energy_cost,omitempty"`
}
