package server

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

/*
	THIS FILE SHOULD CONTAIN ZERO BOILER STRUCTS
	These are the objects things using this rpc server expect and a migration change shouldn't break external services!
*/

type CollectionDetails struct {
	CollectionSlug string `json:"collection_slug"`
	Hash           string `json:"hash"`
	TokenID        int64  `json:"token_id"`
}

// Mech is the struct that rpc expects for mechs
type Mech struct {
	*CollectionDetails
	ID               string      `json:"id"`
	CollectionItemID string      `json:"collection_item_id"`
	Label            string      `json:"label"`
	WeaponHardpoints int         `json:"weapon_hardpoints"`
	UtilitySlots     int         `json:"utility_slots"`
	Speed            int         `json:"speed"`
	MaxHitpoints     int         `json:"max_hitpoints"`
	IsDefault        bool        `json:"is_default"`
	IsInsured        bool        `json:"is_insured"`
	Name             string      `json:"name"`
	GenesisTokenID   null.Int    `json:"genesis_token_id,omitempty"`
	EnergyCoreSize   string      `json:"energy_core_size"`
	Tier             null.String `json:"tier,omitempty"`

	BlueprintID string         `json:"blueprint_id"`
	Blueprint   *BlueprintMech `json:"blueprint_mech,omitempty"`

	BrandID string `json:"brand_id"`
	Brand   *Brand `json:"brand"`

	OwnerID string `json:"owner_id"`
	Owner   *User  `json:"user"`

	FactionID string   `json:"faction_id"`
	Faction   *Faction `json:"faction,omitempty"`

	ModelID string `json:"model_id"`

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

type BlueprintUtility struct {
	ID        string      `json:"id"`
	BrandID   null.String `json:"brand_id,omitempty"`
	Label     string      `json:"label"`
	UpdatedAt time.Time   `json:"updated_at"`
	CreatedAt time.Time   `json:"created_at"`
	Type      string      `json:"type"`

	ShieldBlueprint      *BlueprintUtilityShield      `json:"shield_blueprint,omitempty"`
	AttackDroneBlueprint *BlueprintUtilityAttackDrone `json:"attack_drone_blueprint,omitempty"`
	RepairDroneBlueprint *BlueprintUtilityRepairDrone `json:"repair_drone_blueprint,omitempty"`
	AcceleratorBlueprint *BlueprintUtilityAccelerator `json:"accelerator_blueprint,omitempty"`
	AntiMissileBlueprint *BlueprintUtilityAntiMissile `json:"anti_missile_blueprint,omitempty"`
}
