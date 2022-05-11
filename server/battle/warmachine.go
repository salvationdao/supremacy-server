package battle

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type WarMachine struct {
	ID            string `json:"id"`
	Hash          string `json:"hash"`
	OwnedByID     string `json:"ownedByID"`
	Name          string `json:"name"`
	ParticipantID byte   `json:"participantID"`
	FactionID     string `json:"factionID"`
	MaxHealth     uint32 `json:"maxHealth"`
	Health        uint32 `json:"health"`

	Model string `json:"model"`
	Skin  string `json:"skin"`
	Speed int    `json:"speed"`

	Faction *Faction `json:"faction"`
	Tier    string   `json:"tier"`

	EnergyCore *EnergyCore    `json:"energy_core,omitempty"`
	Abilities  []*GameAbility `json:"abilities"`
	Weapons    []*Weapon      `json:"weapons"`

	//Durability         int             `json:"durability"`
	//PowerGrid          int             `json:"powerGrid"`
	//CPU                int             `json:"cpu"`
	//WeaponHardpoint    int             `json:"weaponHardpoint"`
	//TurretHardpoint    int             `json:"turretHardpoint"`
	//UtilitySlots       int             `json:"utilitySlots"`
	//ShieldRechargeRate float64         `json:"shieldRechargeRate"`
	//Description   *string         `json:"description"`
	//ExternalUrl   string          `json:"externalUrl"`
	//Image         string          `json:"image"`
	//MaxShield     uint32          `json:"maxShield"`
	//Shield        uint32          `json:"shield"`
	//Energy        uint32          `json:"energy"`
	//Stat          *Stat           `json:"stat"`
	//ImageAvatar   string          `json:"imageAvatar"`
	//Position      *server.Vector3 `json:"position"`
	//Rotation      int             `json:"rotation"`
}

type EnergyCore struct {
	ID           string          `json:"id"`
	Label        string          `json:"label"`
	Capacity     decimal.Decimal `json:"capacity"`
	MaxDrawRate  decimal.Decimal `json:"max_draw_rate"`
	RechargeRate decimal.Decimal `json:"recharge_rate"`
	Armour       decimal.Decimal `json:"armour"`
	MaxHitpoints decimal.Decimal `json:"max_hitpoints"`
	Tier         null.String     `json:"tier,omitempty"`
	EquippedOn   null.String     `json:"equipped_on,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type DamageType byte

const (
	DamageTypeDefault   DamageType = 0
	DamageTypeEnergy    DamageType = 1
	DamageTypeExplosive DamageType = 2
)

type Weapon struct {
	ID                  string     `json:"id"`    // UUID that client uses to apply weapon stats to the correct weapons (unique per model/blueprint)
	Hash                string     `json:"hash"`  // Unique hash of a user's weapon
	Model               string     `json:"model"` // Unused for built-in mech weapons
	Skin                string     `json:"skin"`  // Unused for built-in mech weapons
	Name                string     `json:"name"`
	Damage              int        `json:"damage"`
	DamageFalloff       int        `json:"damageFalloff"`       // Distance at which damage starts decreasing
	DamageFalloffRate   int        `json:"damageFalloffRate"`   // How much the damage decreases by per km
	DamageRadius        int        `json:"damageRadius"`        // Enemies within this radius when the projectile hits something is damaged
	DamageRadiusFalloff int        `json:"damageRadiusFalloff"` // Distance at which damage starts decreasing (must be greater than 0 and less than damageRadius to have any affect)
	DamageType          DamageType `json:"damageType"`          // For calculating damage weakness/resistance (eg: shields take 25% extra damage from energy weapons)
	Spread              float32    `json:"spread"`              // Projectiles are randomly offset inside a cone. Spread is the half-angle of the cone, in degrees.
	RateOfFire          float32    `json:"rateOfFire"`          // Rounds per minute
	ProjectileSpeed     int        `json:"projectileSpeed"`     // cm/s
	MaxAmmo             int        `json:"maxAmmo"`             // The max amount of ammo this weapon can hold
}
