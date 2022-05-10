package comms

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

/*
	THIS FILE SHOULD CONTAIN ZERO BOILER STRUCTS
	These are the objects things using this rpc server expect and a migration change shouldn't break external services!

	We should have convert functions on our objects that convert them to our api objects, for example
	apiMech := server.Mech.ToApiMechV1()
*/

type CollectionDetails struct {
	CollectionSlug string `json:"collection_slug"`
	Hash           string `json:"hash"`
	TokenID        int64  `json:"token_id"`
}

// Mech is the struct that rpc expects for mechs
type Mech struct {
	*CollectionDetails
	ID               string   `json:"id"`
	BrandID          string   `json:"brand_id"`
	Label            string   `json:"label"`
	WeaponHardpoints int      `json:"weapon_hardpoints"`
	UtilitySlots     int      `json:"utility_slots"`
	Speed            int      `json:"speed"`
	MaxHitpoints     int      `json:"max_hitpoints"`
	BlueprintID      string   `json:"blueprint_id"`
	IsDefault        bool     `json:"is_default"`
	IsInsured        bool     `json:"is_insured"`
	Name             string   `json:"name"`
	ModelID          string   `json:"model_id"`
	CollectionItemID string   `json:"collection_item_id"`
	GenesisTokenID   null.Int `json:"genesis_token_id,omitempty"`
	OwnerID          string   `json:"owner_id"`
	FactionID        string   `json:"faction_id"`
	EnergyCoreSize   string   `json:"energy_core_size"`

	Tier null.String `json:"tier,omitempty"`

	// Connected objects
	DefaultChassisSkinID string    `json:"default_chassis_skin_id"`
	DefaultChassisSkin   *MechSkin `json:"default_chassis_skin"`

	ChassisSkinID null.String `json:"chassis_skin_id,omitempty"`
	ChassisSkin   *MechSkin   `json:"chassis_skin,omitempty"`

	IntroAnimationID null.String    `json:"intro_animation_id,omitempty"`
	IntroAnimation   *MechAnimation `json:"intro_animation,omitempty"`

	OutroAnimationID null.String    `json:"outro_animation_id,omitempty"`
	OutroAnimation   *MechAnimation `json:"outro_animation,omitempty"`

	EnergyCoreID null.String `json:"energy_core_id,omitempty"`
	EnergyCore   *EnergyCore `json:"energy_core,omitempty"`

	Weapons []*Weapon  `json:"weapons"`
	Utility []*Utility `json:"utility"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type MechSkin struct {
	*CollectionDetails
	ID               string              `json:"id"`
	BlueprintID      string              `json:"blueprint_id"`
	CollectionItemID string              `json:"collection_item_id"`
	GenesisTokenID   decimal.NullDecimal `json:"genesis_token_id,omitempty"`
	Label            string              `json:"label"`
	OwnerID          string              `json:"owner_id"`
	ChassisModel     string              `json:"chassis_model"`
	EquippedOn       null.String         `json:"equipped_on,omitempty"`
	Tier             null.String         `json:"tier,omitempty"`
	ImageURL         null.String         `json:"image_url,omitempty"`
	AnimationURL     null.String         `json:"animation_url,omitempty"`
	CardAnimationURL null.String         `json:"card_animation_url,omitempty"`
	AvatarURL        null.String         `json:"avatar_url,omitempty"`
	LargeImageURL    null.String         `json:"large_image_url,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
}

type MechAnimation struct {
	*CollectionDetails
	ID               string      `json:"id"`
	BlueprintID      string      `json:"blueprint_id"`
	CollectionItemID string      `json:"collection_item_id"`
	Label            string      `json:"label"`
	OwnerID          string      `json:"owner_id"`
	ChassisModel     string      `json:"chassis_model"`
	EquippedOn       null.String `json:"equipped_on,omitempty"`
	Tier             null.String `json:"tier,omitempty"`
	IntroAnimation   null.Bool   `json:"intro_animation,omitempty"`
	OutroAnimation   null.Bool   `json:"outro_animation,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
}

type EnergyCore struct {
	*CollectionDetails
	ID               string          `json:"id"`
	CollectionItemID string          `json:"collection_item_id"`
	OwnerID          string          `json:"owner_id"`
	Label            string          `json:"label"`
	Size             string          `json:"size"`
	Capacity         decimal.Decimal `json:"capacity"`
	MaxDrawRate      decimal.Decimal `json:"max_draw_rate"`
	RechargeRate     decimal.Decimal `json:"recharge_rate"`
	Armour           decimal.Decimal `json:"armour"`
	MaxHitpoints     decimal.Decimal `json:"max_hitpoints"`
	Tier             null.String     `json:"tier,omitempty"`
	EquippedOn       null.String     `json:"equipped_on,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

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

// Utility is the struct that rpc expects for utility
type Utility struct {
	*CollectionDetails
	ID               string              `json:"id"`
	BrandID          null.String         `json:"brand_id,omitempty"`
	Label            string              `json:"label"`
	UpdatedAt        time.Time           `json:"updated_at"`
	CreatedAt        time.Time           `json:"created_at"`
	BlueprintID      string              `json:"blueprint_id"`
	CollectionItemID null.String         `json:"collection_item_id,omitempty"`
	GenesisTokenID   decimal.NullDecimal `json:"genesis_token_id,omitempty"`
	OwnerID          string              `json:"owner_id"`
	EquippedOn       null.String         `json:"equipped_on,omitempty"`
	Type             string              `json:"type"`
	UtilityObject    any                 `json:"utility_object"`
}

type UtilityAttackDrone struct {
	UtilityID        string `json:"utility_id"`
	Damage           int    `json:"damage"`
	RateOfFire       int    `json:"rate_of_fire"`
	Hitpoints        int    `json:"hitpoints"`
	LifespanSeconds  int    `json:"lifespan_seconds"`
	DeployEnergyCost int    `json:"deploy_energy_cost"`
}

type UtilityShield struct {
	UtilityID          string `json:"utility_id"`
	Hitpoints          int    `json:"hitpoints"`
	RechargeRate       int    `json:"recharge_rate"`
	RechargeEnergyCost int    `json:"recharge_energy_cost"`
}

type UtilityRepairDrone struct {
	UtilityID        string      `json:"utility_id"`
	RepairType       null.String `json:"repair_type,omitempty"`
	RepairAmount     int         `json:"repair_amount"`
	DeployEnergyCost int         `json:"deploy_energy_cost"`
	LifespanSeconds  int         `json:"lifespan_seconds"`
}

type UtilityAccelerator struct {
	UtilityID    string `json:"utility_id"`
	EnergyCost   int    `json:"energy_cost"`
	BoostSeconds int    `json:"boost_seconds"`
	BoostAmount  int    `json:"boost_amount"`
}

type UtilityAntiMissile struct {
	UtilityID      string `json:"utility_id"`
	RateOfFire     int    `json:"rate_of_fire"`
	FireEnergyCost int    `json:"fire_energy_cost"`
}

type TemplateContainer struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`

	BlueprintMech          []*BlueprintMech          `json:"blueprint_mech,omitempty"`
	BlueprintWeapon        []*BlueprintWeapon        `json:"blueprint_weapon,omitempty"`
	BlueprintUtility       []*BlueprintUtility       `json:"blueprint_utility,omitempty"`
	BlueprintMechSkin      []*BlueprintMechSkin      `json:"blueprint_mech_skin,omitempty"`
	BlueprintMechAnimation []*BlueprintMechAnimation `json:"blueprint_mech_animation,omitempty"`
	BlueprintEnergyCore    []*BlueprintEnergyCore    `json:"blueprint_energy_core,omitempty"`
	//BlueprintAmmo []* // TODO: AMMO
}

type BlueprintMechSkin struct {
	ID               string      `json:"id"`
	Collection       string      `json:"collection"`
	ChassisModel     string      `json:"chassis_model"`
	Label            string      `json:"label"`
	Tier             null.String `json:"tier,omitempty"`
	ImageURL         null.String `json:"image_url,omitempty"`
	AnimationURL     null.String `json:"animation_url,omitempty"`
	CardAnimationURL null.String `json:"card_animation_url,omitempty"`
	LargeImageURL    null.String `json:"large_image_url,omitempty"`
	AvatarURL        null.String `json:"avatar_url,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
}

type BlueprintMechAnimation struct {
	ID             string      `json:"id"`
	Collection     string      `json:"collection"`
	Label          string      `json:"label"`
	ChassisModel   string      `json:"chassis_model"`
	EquippedOn     null.String `json:"equipped_on,omitempty"`
	Tier           null.String `json:"tier,omitempty"`
	IntroAnimation null.Bool   `json:"intro_animation,omitempty"`
	OutroAnimation null.Bool   `json:"outro_animation,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

type BlueprintEnergyCore struct {
	ID           string          `json:"id"`
	Collection   string          `json:"collection"`
	Label        string          `json:"label"`
	Size         string          `json:"size"`
	Capacity     decimal.Decimal `json:"capacity"`
	MaxDrawRate  decimal.Decimal `json:"max_draw_rate"`
	RechargeRate decimal.Decimal `json:"recharge_rate"`
	Armour       decimal.Decimal `json:"armour"`
	MaxHitpoints decimal.Decimal `json:"max_hitpoints"`
	Tier         null.String     `json:"tier,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type BlueprintMech struct {
	ID                   string      `json:"id"`
	BrandID              string      `json:"brand_id"`
	Label                string      `json:"label"`
	Slug                 string      `json:"slug"`
	Skin                 string      `json:"skin"`
	WeaponHardpoints     int         `json:"weapon_hardpoints"`
	UtilitySlots         int         `json:"utility_slots"`
	Speed                int         `json:"speed"`
	MaxHitpoints         int         `json:"max_hitpoints"`
	UpdatedAt            time.Time   `json:"updated_at"`
	CreatedAt            time.Time   `json:"created_at"`
	ModelID              string      `json:"model_id"`
	EnergyCoreSize       null.String `json:"energy_core_size,omitempty"`
	Tier                 null.String `json:"tier,omitempty"`
	DefaultChassisSkinID string      `json:"default_chassis_skin_id"`
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

type BlueprintUtility struct {
	ID            string      `json:"id"`
	BrandID       null.String `json:"brand_id,omitempty"`
	Label         string      `json:"label"`
	UpdatedAt     time.Time   `json:"updated_at"`
	CreatedAt     time.Time   `json:"created_at"`
	Type          string      `json:"type"`
	UtilityObject any         `json:"utility_object"`
}

type BlueprintUtilityAttackDrone struct {
	ID                 string    `json:"id"`
	BlueprintUtilityID string    `json:"blueprint_utility_id"`
	Damage             int       `json:"damage"`
	RateOfFire         int       `json:"rate_of_fire"`
	Hitpoints          int       `json:"hitpoints"`
	LifespanSeconds    int       `json:"lifespan_seconds"`
	DeployEnergyCost   int       `json:"deploy_energy_cost"`
	CreatedAt          time.Time `json:"created_at"`
}

type BlueprintUtilityShield struct {
	ID                 string    `json:"id"`
	BlueprintUtilityID string    `json:"blueprint_utility_id"`
	Hitpoints          int       `json:"hitpoints"`
	RechargeRate       int       `json:"recharge_rate"`
	RechargeEnergyCost int       `json:"recharge_energy_cost"`
	CreatedAt          time.Time `json:"created_at"`
}

type BlueprintUtilityRepairDrone struct {
	ID                 string      `json:"id"`
	BlueprintUtilityID string      `json:"blueprint_utility_id"`
	RepairType         null.String `json:"repair_type,omitempty"`
	RepairAmount       int         `json:"repair_amount"`
	DeployEnergyCost   int         `json:"deploy_energy_cost"`
	LifespanSeconds    int         `json:"lifespan_seconds"`
	CreatedAt          time.Time   `json:"created_at"`
}

type BlueprintUtilityAccelerator struct {
	ID                 string    `json:"id"`
	BlueprintUtilityID string    `json:"blueprint_utility_id"`
	EnergyCost         int       `json:"energy_cost"`
	BoostSeconds       int       `json:"boost_seconds"`
	BoostAmount        int       `json:"boost_amount"`
	CreatedAt          time.Time `json:"created_at"`
}

type BlueprintUtilityAntiMissile struct {
	ID                 string    `json:"id"`
	BlueprintUtilityID string    `json:"blueprint_utility_id"`
	RateOfFire         int       `json:"rate_of_fire"`
	FireEnergyCost     int       `json:"fire_energy_cost"`
	CreatedAt          time.Time `json:"created_at"`
}
