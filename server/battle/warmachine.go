package battle

import (
	"server"
	"server/db/boiler"
	"time"

	"github.com/sasha-s/go-deadlock"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type WarMachine struct {
	ID            string  `json:"id"`
	Hash          string  `json:"hash"`
	OwnedByID     string  `json:"ownedByID"`
	OwnerUsername string  `json:"ownerUsername"`
	Name          string  `json:"name"`
	Label         string  `json:"label"`
	ParticipantID byte    `json:"participantID"`
	MaxHealth     uint32  `json:"maxHealth"`
	MaxShield     uint32  `json:"maxShield"`
	Health        uint32  `json:"health"`
	HeightMeters  float64 `json:"height"`

	AIType *AIType `json:"aiType"`

	// shield
	Shield                  uint32  `json:"shield"`
	ShieldRechargeRate      uint32  `json:"shieldRechargeRate"`
	ShieldRechargeDelay     float64 `json:"shieldRechargeDelay"`
	ShieldRechargePowerCost uint32  `json:"shieldRechargePowerCost"`
	ShieldTypeID            string  `json:"shieldTypeID"`
	ShieldTypeLabel         string  `json:"shieldTypeLabel"`
	ShieldTypeDescription   string  `json:"shieldTypeDescription"`

	ModelID   string `json:"modelID"`
	ModelName string `json:"model_name"`
	SkinName  string `json:"skin_name"`
	SkinID    string `json:"skinID"`
	Speed     int    `json:"speed"`

	Faction   *Faction `json:"faction"`
	FactionID string   `json:"factionID"`
	Tier      string   `json:"tier"`

	PowerCore *PowerCore `json:"power_core,omitempty"`
	Weapons   []*Weapon  `json:"weapons"`
	Utility   []*Utility `json:"utility"`

	// these objects below are used by us and not game client
	Image       string          `json:"image"`
	ImageAvatar string          `json:"imageAvatar"`
	Position    *server.Vector3 `json:"position"`
	Rotation    int             `json:"rotation"`
	IsHidden    bool            `json:"isHidden"`

	Stats *Stats `json:"stats"`

	Status *Status `json:"status"`

	deadlock.RWMutex // lock for any mech detail changes

	// data for system message
	damagedBlockCount int

	isAI bool
}

type Status struct {
	IsHacked  bool `json:"is_hacked"`
	IsStunned bool `json:"is_stunned"`
}

type WarMachineGameClient struct {
	Hash         string   `json:"hash"`
	Name         string   `json:"name"`
	OwnerName    string   `json:"owner_name"`
	OwnerID      string   `json:"owner_id"`
	Faction      *Faction `json:"faction"` // will be deprecated soon
	FactionID    string   `json:"faction_id"`
	ModelName    string   `json:"model_name"` // will be deprecated soon
	ModelID      string   `json:"model_id"`
	SkinName     string   `json:"skin_name"` // will be deprecated soon
	SkinID       string   `json:"skin_id"`
	Tier         string   `json:"tier"`
	HeightMeters float64  `json:"height"`

	Weapons       []*Weapon               `json:"weapons"`
	Customisation WarMachineCustomisation `json:"customisation"`

	Health    uint32 `json:"health"`
	HealthMax uint32 `json:"health_max"`

	// shield
	Shield                  uint32  `json:"shield"`
	ShieldMax               uint32  `json:"shield_max"`
	ShieldRechargeRate      uint32  `json:"shield_recharge_rate"`
	ShieldRechargePowerCost uint32  `json:"shield_recharge_power_cost"`
	ShieldRechargeDelay     float64 `json:"shield_recharge_delay"`
	ShieldTypeID            string  `json:"shield_type_id"`

	Speed                int     `json:"speed"`
	SprintSpreadModifier float32 `json:"sprint_spread_modifier"`

	PowerCore  PowerCoreGameClient  `json:"power_core"`
	PowerStats WarMachinePowerStats `json:"power_stats"`
	Stats      *Stats               `json:"stats"`
}

type Stats struct {
	TotalWins       int `json:"total_wins"`
	TotalDeaths     int `json:"total_deaths"`
	TotalKills      int `json:"total_kills"`
	BattlesSurvived int `json:"battles_survived"`
	TotalLosses     int `json:"total_losses"`
}

type WarMachineCustomisation struct {
	IntroAnimationID string `json:"intro_animation_id"`
	OutroAnimationID string `json:"outro_animation_id"`
}

type PowerCore struct {
	ID           string          `json:"id"`
	ModelID      string          `json:"model_id"`
	Label        string          `json:"label"`
	Capacity     decimal.Decimal `json:"capacity"`
	MaxDrawRate  decimal.Decimal `json:"max_draw_rate"`
	RechargeRate decimal.Decimal `json:"recharge_rate"`
	Armour       decimal.Decimal `json:"armour"`
	MaxHitpoints decimal.Decimal `json:"max_hitpoints"`
	EquippedOn   null.String     `json:"equipped_on,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`

	MovementShare int `json:"movement_share,omitempty"`
	WeaponShare   int `json:"weapon_share,omitempty"`
	UtilityShare  int `json:"utility_share,omitempty"`

	MovementSystemCurrentPower float32
	ShieldSystemCurrentPower   float32
	WeaponSystemCurrentPower   float32
}

type PowerCoreGameClient struct {
	ID                       string  `json:"id"`
	ModelID                  string  `json:"model_id"`
	PowerCapacity            float64 `json:"power_capacity"`
	RechargeRate             float64 `json:"recharge_rate"`
	MaxDrawRate              float64 `json:"max_draw_rate"`
	WeaponSystemAllocation   int     `json:"weapon_system_allocation"`
	MovementSystemAllocation int     `json:"movement_system_allocation"`
	UtilitySystemAllocation  int     `json:"utility_system_allocation"`
}

type WarMachinePowerStats struct {
	IdleDrain float32 `json:"idle_drain"`
	WalkDrain float32 `json:"walk_drain"`
	RunDrain  float32 `json:"run_drain"`
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
	ID                  string     `json:"id"`       // UUID that client uses to apply weapon stats to the correct weapons (unique per model/blueprint)
	Hash                string     `json:"hash"`     // Unique hash of a user's weapon
	ModelID             string     `json:"model_id"` // Unused for built-in mech weapons
	SkinID              string     `json:"skin_id"`  // Unused for built-in mech weapons
	Name                string     `json:"name"`
	Damage              int        `json:"damage"`
	DamageFalloff       int        `json:"damage_falloff"`        // Distance at which damage starts decreasing
	DamageFalloffRate   int        `json:"damage_falloff_rate"`   // How much the damage decreases by per km
	DamageRadius        int        `json:"damage_radius"`         // Enemies within this radius when the projectile hits something is damaged
	RadiusDamageFalloff int        `json:"damage_radius_falloff"` // Distance at which damage starts decreasing (must be greater than 0 and less than damageRadius to have any affect)
	DamageType          DamageType `json:"damage_type"`           // For calculating damage weakness/resistance (eg: shields take 25% extra damage from energy weapons)
	Spread              float64    `json:"spread"`                // Projectiles are randomly offset inside a cone. Spread is the half-angle of the cone, in degrees.
	RateOfFire          float64    `json:"rate_of_fire"`          // Rounds per minute
	ProjectileSpeed     int        `json:"projectile_speed"`      // cm/s
	MaxAmmo             int        `json:"max_ammo"`              // The max amount of ammo this weapon can hold
	CurrentAmmo         int
	PowerCost           float64 `json:"power_cost"`
	PowerInstantDrain   bool    `json:"power_instant_drain"`
	ProjectileAmount    int     `json:"projectile_amount"`
	DotTickDamage       float64 `json:"dot_tick_damage"`
	DotMaxTicks         int     `json:"dot_max_ticks"`
	IsArced             bool    `json:"is_arced"`
	ChargeTimeSeconds   float64 `json:"charge_time"`
	BurstRateOfFire     float64 `json:"burst_rate_of_fire"`
	SocketIndex         int     `json:"socket_index"`
}

type Utility struct {
	Type  string `json:"type"`
	Label string `json:"label"`

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

func WarMachinesToClient(wms []*WarMachine) []*WarMachineGameClient {
	var warMachines []*WarMachineGameClient
	for _, wm := range wms {
		warMachines = append(warMachines, WarMachineToClient(wm))
	}
	return warMachines
}

func WarMachineToClient(wm *WarMachine) *WarMachineGameClient {
	wmgc := &WarMachineGameClient{
		Hash:      wm.Hash,
		Name:      wm.Name,
		OwnerName: wm.OwnerUsername,
		OwnerID:   wm.OwnedByID,
		Faction:   wm.Faction,
		FactionID: wm.FactionID,
		ModelName: wm.ModelName,
		ModelID:   wm.ModelID,
		SkinName:  wm.SkinName,
		SkinID:    wm.SkinID,
		Tier:      wm.Tier,

		HeightMeters: wm.HeightMeters,

		Weapons: wm.Weapons,

		Health:                  wm.Health,
		HealthMax:               wm.MaxHealth,
		ShieldMax:               wm.MaxShield,
		ShieldRechargeRate:      wm.ShieldRechargeRate,
		ShieldRechargePowerCost: wm.ShieldRechargePowerCost,
		ShieldRechargeDelay:     wm.ShieldRechargeDelay,
		ShieldTypeID:            wm.ShieldTypeID,

		Speed: wm.Speed,

		Stats: wm.Stats,
	}

	if wm.PowerCore != nil {
		wmgc.PowerCore = PowerCoreGameClient{
			ID:                       wm.PowerCore.ID,
			ModelID:                  wm.PowerCore.ModelID,
			PowerCapacity:            wm.PowerCore.Capacity.InexactFloat64(),
			RechargeRate:             wm.PowerCore.RechargeRate.InexactFloat64(),
			MaxDrawRate:              wm.PowerCore.MaxDrawRate.InexactFloat64(),
			WeaponSystemAllocation:   wm.PowerCore.WeaponShare,
			MovementSystemAllocation: wm.PowerCore.MovementShare,
			UtilitySystemAllocation:  wm.PowerCore.UtilityShare,
		}
	}

	return wmgc
}

func WeaponsFromServer(wpns []*server.Weapon, compatibilityList []string, inheritedSkinID null.String, inheritAll bool) []*Weapon {
	var weapons []*Weapon
	for _, wpn := range wpns {
		weapons = append(weapons, WeaponFromServer(wpn, compatibilityList, inheritedSkinID, inheritAll))
	}
	return weapons
}

func WeaponFromServer(weapon *server.Weapon, compatibilityList []string, inheritedSkinID null.String, inheritAll bool) *Weapon {
	skinID := weapon.WeaponSkin.BlueprintID
	if inheritedSkinID.Valid && inheritedSkinID.String != "" && (inheritAll || weapon.InheritSkin) {
		isCompatible := false
		for _, c := range compatibilityList {
			if c == weapon.BlueprintID {
				isCompatible = true
				break
			}
		}
		if isCompatible {
			skinID = inheritedSkinID.String
		}
	}
	if !weapon.InheritSkin {
		skinID = weapon.WeaponSkin.BlueprintID
	}

	return &Weapon{
		ID:          weapon.ID,
		Hash:        weapon.Hash,
		Name:        weapon.Label,
		ModelID:     weapon.BlueprintID,
		SkinID:      skinID,
		SocketIndex: weapon.SlotNumber.Int,
		//stats
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
		PowerCost:           weapon.PowerCost.Decimal.InexactFloat64(),
		ProjectileAmount:    weapon.ProjectileAmount.Int,
		DotTickDamage:       weapon.DotTickDamage.Decimal.InexactFloat64(),
		DotMaxTicks:         weapon.DotMaxTicks.Int,
		IsArced:             weapon.IsArced.Bool,
		ChargeTimeSeconds:   weapon.ChargeTimeSeconds.Decimal.InexactFloat64(),
		BurstRateOfFire:     weapon.BurstRateOfFire.Decimal.InexactFloat64(),
	}
}

func PowerCoreFromServer(pc *server.PowerCore) *PowerCore {
	if pc == nil {
		return nil
	}
	return &PowerCore{
		ID:            pc.ID,
		ModelID:       pc.BlueprintID,
		Label:         pc.Label,
		Capacity:      pc.Capacity,
		MaxDrawRate:   pc.MaxDrawRate,
		RechargeRate:  pc.RechargeRate,
		Armour:        pc.Armour,
		MaxHitpoints:  pc.MaxHitpoints,
		EquippedOn:    pc.EquippedOn,
		CreatedAt:     pc.CreatedAt,
		MovementShare: pc.MovementShare,
		WeaponShare:   pc.WeaponShare,
		UtilityShare:  pc.UtilityShare,
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
