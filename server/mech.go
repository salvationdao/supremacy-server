package server

import (
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

/*
	THIS FILE SHOULD CONTAIN ZERO BOILER STRUCTS
*/

type CollectionDetails struct {
	CollectionSlug string `json:"collection_slug"`
	Hash           string `json:"hash"`
	TokenID        int64  `json:"token_id"`
}

// Mech is the struct that rpc expects for mechs
type Mech struct {
	*CollectionDetails
	ID                    string              `json:"id"`
	Label                 string              `json:"label"`
	WeaponHardpoints      int                 `json:"weapon_hardpoints"`
	UtilitySlots          int                 `json:"utility_slots"`
	Speed                 int                 `json:"speed"`
	MaxHitpoints          int                 `json:"max_hitpoints"`
	IsDefault             bool                `json:"is_default"`
	IsInsured             bool                `json:"is_insured"`
	Name                  string              `json:"name"`
	GenesisTokenID        decimal.NullDecimal `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID decimal.NullDecimal `json:"limited_release_token_id,omitempty"`
	EnergyCoreSize        string              `json:"energy_core_size"`
	Tier                  null.String         `json:"tier,omitempty"`

	BlueprintID string         `json:"blueprint_id"`
	Blueprint   *BlueprintMech `json:"blueprint_mech,omitempty"`

	BrandID string `json:"brand_id"`
	Brand   *Brand `json:"brand"`

	OwnerID string `json:"owner_id"`
	Owner   *User  `json:"user"`

	FactionID string   `json:"faction_id"`
	Faction   *Faction `json:"faction,omitempty"`

	ModelID string     `json:"model_id"`
	Model   *MechModel `json:"model"`

	// Connected objects
	DefaultChassisSkinID string             `json:"default_chassis_skin_id"`
	DefaultChassisSkin   *BlueprintMechSkin `json:"default_chassis_skin"`

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
	EnergyCoreSize       string      `json:"energy_core_size,omitempty"`
	Tier                 null.String `json:"tier,omitempty"`
	DefaultChassisSkinID string      `json:"default_chassis_skin_id"`
	Collection           string      `json:"collection"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        decimal.NullDecimal `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID decimal.NullDecimal `json:"limited_release_token_id,omitempty"`
}

func BlueprintMechFromBoiler(mech *boiler.BlueprintMech) *BlueprintMech {
	return &BlueprintMech{
		ID:               mech.ID,
		BrandID:          mech.BrandID,
		Label:            mech.Label,
		Slug:             mech.Slug,
		Skin:             mech.Skin,
		WeaponHardpoints: mech.WeaponHardpoints,
		UtilitySlots:     mech.UtilitySlots,
		Speed:            mech.Speed,
		MaxHitpoints:     mech.MaxHitpoints,
		UpdatedAt:        mech.UpdatedAt,
		CreatedAt:        mech.CreatedAt,
		ModelID:          mech.ModelID,
		EnergyCoreSize:   mech.EnergyCoreSize,
		Tier:             mech.Tier,
		Collection:       mech.Collection,
	}
}

type BlueprintUtility struct {
	ID         string      `json:"id"`
	BrandID    null.String `json:"brand_id,omitempty"`
	Label      string      `json:"label"`
	UpdatedAt  time.Time   `json:"updated_at"`
	CreatedAt  time.Time   `json:"created_at"`
	Type       string      `json:"type"`
	Collection string      `json:"collection"`

	ShieldBlueprint      *BlueprintUtilityShield      `json:"shield_blueprint,omitempty"`
	AttackDroneBlueprint *BlueprintUtilityAttackDrone `json:"attack_drone_blueprint,omitempty"`
	RepairDroneBlueprint *BlueprintUtilityRepairDrone `json:"repair_drone_blueprint,omitempty"`
	AcceleratorBlueprint *BlueprintUtilityAccelerator `json:"accelerator_blueprint,omitempty"`
	AntiMissileBlueprint *BlueprintUtilityAntiMissile `json:"anti_missile_blueprint,omitempty"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        decimal.NullDecimal `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID decimal.NullDecimal `json:"limited_release_token_id,omitempty"`
}

type MechModel struct {
	ID                   string      `json:"id"`
	Label                string      `json:"label"`
	CreatedAt            time.Time   `json:"created_at"`
	DefaultChassisSkinID null.String `json:"default_chassis_skin_id,omitempty"`
}
