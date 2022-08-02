package battle

import (
	"github.com/sasha-s/go-deadlock"
	"server"
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type WarMachine struct {
	ID            string `json:"id"`
	Hash          string `json:"hash"`
	OwnedByID     string `json:"ownedByID"`
	OwnerUsername string `json:"ownerUsername"`
	Name          string `json:"name"`
	ParticipantID byte   `json:"participantID"`
	FactionID     string `json:"factionID"`
	MaxHealth     uint32 `json:"maxHealth"`
	MaxShield     uint32 `json:"maxShield"`
	Health        uint32 `json:"health"`

	ModelID string `json:"modelID"`
	Model   string `json:"model"`
	Skin    string `json:"skin"`
	Speed   int    `json:"speed"`

	Faction *Faction `json:"faction"`
	Tier    string   `json:"tier"`

	PowerCore *PowerCore     `json:"power_core,omitempty"`
	Abilities []*GameAbility `json:"abilities"`
	Weapons   []*Weapon      `json:"weapons"`
	Utility   []*Utility     `json:"utility"`

	// these objects below are used by us and not game client
	Image       string          `json:"image"`
	ImageAvatar string          `json:"imageAvatar"`
	Position    *server.Vector3 `json:"position"`
	Rotation    int             `json:"rotation"`
	IsHidden    bool            `json:"isHidden"`
	//Durability         int             `json:"durability"`
	//PowerGrid          int             `json:"powerGrid"`
	//CPU                int             `json:"cpu"`
	//WeaponHardpoint    int             `json:"weaponHardpoint"`
	//TurretHardpoint    int             `json:"turretHardpoint"`
	//UtilitySlots       int             `json:"utilitySlots"`
	//Description   *string         `json:"description"`
	//ExternalUrl   string          `json:"externalUrl"`
	Shield             uint32 `json:"shield"`
	ShieldRechargeRate uint32 `json:"shieldRechargeRate"`

	Stats *Stats `json:"stats"`
	//Energy        uint32          `json:"energy"`
	//Stat          *Stat           `json:"stat"`

	deadlock.RWMutex // lock for any mech detail changes
}

type Stats struct {
	TotalWins       int `json:"total_wins"`
	TotalDeaths     int `json:"total_deaths"`
	TotalKills      int `json:"total_kills"`
	BattlesSurvived int `json:"battles_survived"`
	TotalLosses     int `json:"total_losses"`
}

type PowerCore struct {
	ID           string          `json:"id"`
	Label        string          `json:"label"`
	Capacity     decimal.Decimal `json:"capacity"`
	MaxDrawRate  decimal.Decimal `json:"max_draw_rate"`
	RechargeRate decimal.Decimal `json:"recharge_rate"`
	Armour       decimal.Decimal `json:"armour"`
	MaxHitpoints decimal.Decimal `json:"max_hitpoints"`
	EquippedOn   null.String     `json:"equipped_on,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type DamageType byte

const (
	DamageTypeDefault   DamageType = 0
	DamageTypeEnergy    DamageType = 1
	DamageTypeExplosive DamageType = 2
)

func DamageTypeFromString(dt string) DamageType {
	switch dt {
	case boiler.DamageTypeKinetic:
		return DamageTypeDefault
	case boiler.DamageTypeEnergy:
		return DamageTypeEnergy
	case boiler.DamageTypeExplosive:
		return DamageTypeExplosive
	}
	return DamageTypeDefault
}

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
	RadiusDamageFalloff int        `json:"damageRadiusFalloff"` // Distance at which damage starts decreasing (must be greater than 0 and less than damageRadius to have any affect)
	DamageType          DamageType `json:"damageType"`          // For calculating damage weakness/resistance (eg: shields take 25% extra damage from energy weapons)
	Spread              float64    `json:"spread"`              // Projectiles are randomly offset inside a cone. Spread is the half-angle of the cone, in degrees.
	RateOfFire          float64    `json:"rateOfFire"`          // Rounds per minute
	ProjectileSpeed     int        `json:"projectileSpeed"`     // cm/s
	MaxAmmo             int        `json:"maxAmmo"`             // The max amount of ammo this weapon can hold
}

type Utility struct {
	Type  string `json:"type"`
	Label string `json:"label"`

	Shield      *UtilityShield      `json:"shield,omitempty"`
	AttackDrone *UtilityAttackDrone `json:"attack_drone,omitempty"`
	RepairDrone *UtilityRepairDrone `json:"repair_drone,omitempty"`
	Accelerator *UtilityAccelerator `json:"accelerator,omitempty"`
	AntiMissile *UtilityAntiMissile `json:"anti_missile,omitempty"`
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

func WeaponsFromServer(wpns []*server.Weapon) []*Weapon {
	var weapons []*Weapon
	for _, wpn := range wpns {
		weapons = append(weapons, WeaponFromServer(wpn))
	}
	return weapons
}

func WeaponFromServer(weapon *server.Weapon) *Weapon {
	return &Weapon{
		ID:                  weapon.ID,
		Hash:                weapon.Hash,
		Name:                weapon.Label,
		Damage:              weapon.Damage,
		DamageFalloff:       weapon.DamageFalloff.Int,
		DamageFalloffRate:   weapon.DamageFalloffRate.Int,
		DamageRadius:        weapon.Radius.Int,
		Spread:              weapon.Spread.Decimal.InexactFloat64(),
		RateOfFire:          weapon.RateOfFire.Decimal.InexactFloat64(),
		ProjectileSpeed:     int(weapon.ProjectileSpeed.Decimal.IntPart()),
		MaxAmmo:             weapon.MaxAmmo.Int,
		RadiusDamageFalloff: weapon.RadiusDamageFalloff.Int,
		DamageType:          DamageTypeFromString(weapon.DefaultDamageType),
		//Model:               	weapon.Model, // TODO: weapon models
		//Skin:              	weapon.Skin, // TODO: weapon skins
	}
}

func PowerCoreFromServer(ec *server.PowerCore) *PowerCore {
	if ec == nil {
		return nil
	}
	return &PowerCore{
		ID:           ec.ID,
		Label:        ec.Label,
		Capacity:     ec.Capacity,
		MaxDrawRate:  ec.MaxDrawRate,
		RechargeRate: ec.RechargeRate,
		Armour:       ec.Armour,
		MaxHitpoints: ec.MaxHitpoints,
		EquippedOn:   ec.EquippedOn,
		CreatedAt:    ec.CreatedAt,
	}
}

func UtilitiesFromServer(utils []*server.Utility) []*Utility {
	var utilities []*Utility
	for _, util := range utils {
		utilities = append(utilities, UtilityFromServer(util))
	}
	return utilities
}

func UtilityFromServer(util *server.Utility) *Utility {
	return &Utility{
		Type:        util.Type,
		Label:       util.Label,
		Shield:      UtilityShieldFromServer(util.Shield),
		AttackDrone: UtilityAttackDroneFromServer(util.AttackDrone),
		RepairDrone: UtilityRepairDroneFromServer(util.RepairDrone),
		Accelerator: UtilityAcceleratorFromServer(util.Accelerator),
		AntiMissile: UtilityAntiMissileFromServer(util.AntiMissile),
	}
}

func UtilityAntiMissileFromServer(util *server.UtilityAntiMissile) *UtilityAntiMissile {
	if util == nil {
		return nil
	}
	return &UtilityAntiMissile{
		UtilityID:      util.UtilityID,
		RateOfFire:     util.RateOfFire,
		FireEnergyCost: util.FireEnergyCost,
	}
}

func UtilityAcceleratorFromServer(util *server.UtilityAccelerator) *UtilityAccelerator {
	if util == nil {
		return nil
	}
	return &UtilityAccelerator{
		UtilityID:    util.UtilityID,
		EnergyCost:   util.EnergyCost,
		BoostSeconds: util.BoostSeconds,
		BoostAmount:  util.BoostAmount,
	}
}

func UtilityRepairDroneFromServer(util *server.UtilityRepairDrone) *UtilityRepairDrone {
	if util == nil {
		return nil
	}
	return &UtilityRepairDrone{
		UtilityID:        util.UtilityID,
		RepairType:       util.RepairType,
		RepairAmount:     util.RepairAmount,
		DeployEnergyCost: util.DeployEnergyCost,
		LifespanSeconds:  util.LifespanSeconds,
	}
}

func UtilityShieldFromServer(util *server.UtilityShield) *UtilityShield {
	if util == nil {
		return nil
	}
	return &UtilityShield{
		UtilityID:          util.UtilityID,
		Hitpoints:          util.Hitpoints,
		RechargeRate:       util.RechargeRate,
		RechargeEnergyCost: util.RechargeEnergyCost,
	}
}

func UtilityAttackDroneFromServer(util *server.UtilityAttackDrone) *UtilityAttackDrone {
	if util == nil {
		return nil
	}
	return &UtilityAttackDrone{
		UtilityID:        util.UtilityID,
		Damage:           util.Damage,
		RateOfFire:       util.RateOfFire,
		Hitpoints:        util.Hitpoints,
		LifespanSeconds:  util.LifespanSeconds,
		DeployEnergyCost: util.DeployEnergyCost,
	}
}
