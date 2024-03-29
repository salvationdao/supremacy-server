package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"

	"github.com/volatiletech/null/v8"
)

/*
	THIS FILE SHOULD CONTAIN ZERO BOILER STRUCTS
*/
func (b *Stats) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type Stats struct {
	TotalWins       int `json:"total_wins"`
	TotalDeaths     int `json:"total_deaths"`
	TotalKills      int `json:"total_kills"`
	BattlesSurvived int `json:"battles_survived"`
	TotalLosses     int `json:"total_losses"`
}

// Mech is the struct that rpc expects for mechs
type Mech struct {
	*CollectionItem
	*Stats
	*Images
	ID                    string     `json:"id"`
	Label                 string     `json:"label"`
	Name                  string     `json:"name"`
	GenesisTokenID        null.Int64 `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64 `json:"limited_release_token_id,omitempty"`
	CollectionItemID      string     `json:"-"`

	MechType     string          `json:"mech_type"`
	HeightMeters decimal.Decimal `json:"height_meters"`

	//// stats
	// speed
	Speed        int `json:"speed"`
	BoostedSpeed int `json:"boosted_speed"`
	// hit points
	MaxHitpoints        int `json:"max_hitpoints"`
	BoostedMaxHitpoints int `json:"boosted_max_hitpoints"`
	// shield
	Shield                    int             `json:"shield"`
	BoostedShield             int             `json:"boosted_shield"`
	ShieldRechargeRate        int             `json:"shield_recharge_rate"`
	BoostedShieldRechargeRate int             `json:"boosted_shield_recharge_rate"`
	ShieldRechargeDelay       decimal.Decimal `json:"shield_recharge_delay"`
	ShieldRechargePowerCost   int             `json:"shield_recharge_power_cost"`
	ShieldTypeID              string          `json:"shield_type"`
	ShieldTypeLabel           string          `json:"shield_type_label"`
	ShieldTypeDescription     string          `json:"shield_type_description"`
	// slots
	WeaponHardpoints int `json:"weapon_hardpoints"`
	UtilitySlots     int `json:"utility_slots"`
	// other
	RepairBlocks                int             `json:"repair_blocks"`
	PowerCoreSize               string          `json:"power_core_size"`
	BoostedStat                 string          `json:"boosted_stat"`
	WalkSpeedModifier           decimal.Decimal `json:"walk_speed_modifier"`
	BoostedWalkSpeedModifier    decimal.Decimal `json:"boosted_walk_speed_modifier"`
	SprintSpreadModifier        decimal.Decimal `json:"sprint_spread_modifier"`
	BoostedSprintSpreadModifier decimal.Decimal `json:"boosted_sprint_spread_modifier"`
	IdleDrain                   decimal.Decimal `json:"idle_drain"`
	WalkDrain                   decimal.Decimal `json:"walk_drain"`
	RunDrain                    decimal.Decimal `json:"run_drain"`

	// state
	QueuePosition null.Int    `json:"queue_position"`
	BattleReady   bool        `json:"battle_ready"`
	IsDefault     bool        `json:"is_default"`
	IsInsured     bool        `json:"is_insured"`
	ItemSaleID    null.String `json:"item_sale_id"`

	// Connected objects
	Owner                                 *User          `json:"user"`
	FactionID                             null.String    `json:"faction_id"`
	Faction                               *Faction       `json:"faction,omitempty"`
	BlueprintID                           string         `json:"blueprint_id"`
	BrandID                               string         `json:"brand_id"`
	Brand                                 *Brand         `json:"brand"`
	ChassisSkinID                         string         `json:"chassis_skin_id,omitempty"`
	ChassisSkin                           *MechSkin      `json:"chassis_skin,omitempty"`
	IntroAnimationID                      null.String    `json:"intro_animation_id,omitempty"`
	IntroAnimation                        *MechAnimation `json:"intro_animation,omitempty"`
	OutroAnimationID                      null.String    `json:"outro_animation_id,omitempty"`
	OutroAnimation                        *MechAnimation `json:"outro_animation,omitempty"`
	PowerCoreID                           null.String    `json:"power_core_id,omitempty"`
	PowerCore                             *PowerCore     `json:"power_core,omitempty"`
	Weapons                               WeaponSlice    `json:"weapons"`
	Utility                               UtilitySlice   `json:"utility"`
	UpdatedAt                             time.Time      `json:"updated_at"`
	CreatedAt                             time.Time      `json:"created_at"`
	BlueprintWeaponIDsWithSkinInheritance []string       `json:"blueprint_weapon_ids_with_skin_inheritance"`
	CompatibleBlueprintMechSkinIDs        []string       `json:"compatible_blueprint_mech_skin_ids"`
	InheritAllWeaponSkins                 bool           `json:"inherit_all_weapon_skins"`
}

type BlueprintMech struct {
	ID                      string          `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Label                   string          `boiler:"label" boil:"label" json:"label" toml:"label" yaml:"label"`
	CreatedAt               time.Time       `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	DefaultChassisSkinID    string          `boiler:"default_chassis_skin_id" boil:"default_chassis_skin_id" json:"default_chassis_skin_id" toml:"default_chassis_skin_id" yaml:"default_chassis_skin_id"`
	BrandID                 string          `boiler:"brand_id" boil:"brand_id" json:"brand_id" toml:"brand_id" yaml:"brand_id"`
	MechType                string          `boiler:"mech_type" boil:"mech_type" json:"mech_type" toml:"mech_type" yaml:"mech_type"`
	RepairBlocks            int             `boiler:"repair_blocks" boil:"repair_blocks" json:"repair_blocks" toml:"repair_blocks" yaml:"repair_blocks"`
	BoostStat               null.String     `boiler:"boost_stat" boil:"boost_stat" json:"boost_stat,omitempty" toml:"boost_stat" yaml:"boost_stat,omitempty"`
	WeaponHardpoints        int             `boiler:"weapon_hardpoints" boil:"weapon_hardpoints" json:"weapon_hardpoints" toml:"weapon_hardpoints" yaml:"weapon_hardpoints"`
	PowerCoreSize           string          `boiler:"power_core_size" boil:"power_core_size" json:"power_core_size" toml:"power_core_size" yaml:"power_core_size"`
	UtilitySlots            int             `boiler:"utility_slots" boil:"utility_slots" json:"utility_slots" toml:"utility_slots" yaml:"utility_slots"`
	Speed                   int             `boiler:"speed" boil:"speed" json:"speed" toml:"speed" yaml:"speed"`
	MaxHitpoints            int             `boiler:"max_hitpoints" boil:"max_hitpoints" json:"max_hitpoints" toml:"max_hitpoints" yaml:"max_hitpoints"`
	Collection              string          `boiler:"collection" boil:"collection" json:"collection" toml:"collection" yaml:"collection"`
	AvailabilityID          null.String     `boiler:"availability_id" boil:"availability_id" json:"availability_id,omitempty" toml:"availability_id" yaml:"availability_id,omitempty"`
	ShieldTypeID            string          `boiler:"shield_type_id" boil:"shield_type_id" json:"shield_type_id" toml:"shield_type_id" yaml:"shield_type_id"`
	ShieldMax               int             `boiler:"shield_max" boil:"shield_max" json:"shield_max" toml:"shield_max" yaml:"shield_max"`
	ShieldRechargeRate      int             `boiler:"shield_recharge_rate" boil:"shield_recharge_rate" json:"shield_recharge_rate" toml:"shield_recharge_rate" yaml:"shield_recharge_rate"`
	ShieldRechargePowerCost int             `boiler:"shield_recharge_power_cost" boil:"shield_recharge_power_cost" json:"shield_recharge_power_cost" toml:"shield_recharge_power_cost" yaml:"shield_recharge_power_cost"`
	ShieldRechargeDelay     decimal.Decimal `boiler:"shield_recharge_delay" boil:"shield_recharge_delay" json:"shield_recharge_delay" toml:"shield_recharge_delay" yaml:"shield_recharge_delay"`
	HeightMeters            decimal.Decimal `boiler:"height_meters" boil:"height_meters" json:"height_meters" toml:"height_meters" yaml:"height_meters"`

	BoostedSpeed              int64 `json:"boosted_speed"`
	BoostedMaxHitpoints       int64 `json:"boosted_max_hitpoints"`
	BoostedShieldRechargeRate int64 `json:"boosted_shield_recharge_rate"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        null.Int64 `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64 `json:"limited_release_token_id,omitempty"`
}

func (b *BlueprintMech) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

func BlueprintMechFromBoiler(mech *boiler.BlueprintMech) *BlueprintMech {
	return &BlueprintMech{
		ID:                      mech.ID,
		Label:                   mech.Label,
		CreatedAt:               mech.CreatedAt,
		DefaultChassisSkinID:    mech.DefaultChassisSkinID,
		BrandID:                 mech.BrandID,
		MechType:                mech.MechType,
		RepairBlocks:            mech.RepairBlocks,
		BoostStat:               mech.BoostStat,
		WeaponHardpoints:        mech.WeaponHardpoints,
		PowerCoreSize:           mech.PowerCoreSize,
		UtilitySlots:            mech.UtilitySlots,
		Speed:                   mech.Speed,
		MaxHitpoints:            mech.MaxHitpoints,
		Collection:              mech.Collection,
		AvailabilityID:          mech.AvailabilityID,
		ShieldTypeID:            mech.ShieldTypeID,
		ShieldMax:               mech.ShieldMax,
		ShieldRechargeRate:      mech.ShieldRechargeRate,
		ShieldRechargePowerCost: mech.ShieldRechargePowerCost,
		ShieldRechargeDelay:     mech.ShieldRechargeDelay,
		HeightMeters:            mech.HeightMeters,
	}
}

type BlueprintUtility struct {
	ID               string      `json:"id"`
	BrandID          null.String `json:"brand_id,omitempty"`
	Label            string      `json:"label"`
	UpdatedAt        time.Time   `json:"updated_at"`
	CreatedAt        time.Time   `json:"created_at"`
	Type             string      `json:"type"`
	Collection       string      `json:"collection"`
	Tier             string      `json:"tier,omitempty"`
	ImageURL         null.String `json:"image_url,omitempty"`
	CardAnimationURL null.String `json:"card_animation_url,omitempty"`
	AvatarURL        null.String `json:"avatar_url,omitempty"`
	LargeImageURL    null.String `json:"large_image_url,omitempty"`
	BackgroundColor  null.String `json:"background_color,omitempty"`
	AnimationURL     null.String `json:"animation_url,omitempty"`
	YoutubeURL       null.String `json:"youtube_url,omitempty"`

	ShieldBlueprint      *BlueprintUtilityShield      `json:"shield_blueprint,omitempty"`
	AttackDroneBlueprint *BlueprintUtilityAttackDrone `json:"attack_drone_blueprint,omitempty"`
	RepairDroneBlueprint *BlueprintUtilityRepairDrone `json:"repair_drone_blueprint,omitempty"`
	AcceleratorBlueprint *BlueprintUtilityAccelerator `json:"accelerator_blueprint,omitempty"`
	AntiMissileBlueprint *BlueprintUtilityAntiMissile `json:"anti_missile_blueprint,omitempty"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        null.Int64 `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64 `json:"limited_release_token_id,omitempty"`
}

// IsBattleReady checks if a mech has the minimum it needs for battle
func (m *Mech) IsBattleReady() bool {
	if !m.PowerCoreID.Valid {
		return false
	}
	if len(m.Weapons) <= 0 {
		return false
	}
	return true
}

func (m *Mech) CheckAndSetAsGenesisOrLimited() (genesisID null.Int64, limitedID null.Int64) {
	if !m.GenesisTokenID.Valid && !m.LimitedReleaseTokenID.Valid {
		return
	}
	if m.GenesisTokenID.Valid && m.IsCompleteGenesis() {
		genesisID = m.GenesisTokenID
		m.TokenID = m.GenesisTokenID.Int64
		m.CollectionSlug = "supremacy-genesis"
		return
	}
	if m.LimitedReleaseTokenID.Valid && m.IsCompleteLimited() {
		limitedID = m.LimitedReleaseTokenID
		m.TokenID = m.LimitedReleaseTokenID.Int64
		m.CollectionSlug = "supremacy-limited-release"
		return
	}
	return
}

// IsCompleteGenesis returns true if all parts of this mech are genesis
func (m *Mech) IsCompleteGenesis() bool {
	if !m.GenesisTokenID.Valid {
		return false
	}

	// check weapons
	if len(m.Weapons) < 1 ||
		m.Weapons[0] == nil ||
		!m.Weapons[0].GenesisTokenID.Valid {
		return false
	}
	if len(m.Weapons) < 2 ||
		m.Weapons[1] == nil ||
		!m.Weapons[1].GenesisTokenID.Valid {
		return false
	}
	if len(m.Weapons) < 3 ||
		m.Weapons[2] == nil ||
		!m.Weapons[2].GenesisTokenID.Valid {
		return false
	}

	// check power core
	if m.PowerCore == nil ||
		!m.PowerCore.GenesisTokenID.Valid {
		return false
	}

	// check skin
	if m.ChassisSkin == nil ||
		!m.ChassisSkin.GenesisTokenID.Valid {
		return false
	}
	return true
}

// IsCompleteLimited returns true if all parts of this mech are limited
func (m *Mech) IsCompleteLimited() bool {
	if !m.LimitedReleaseTokenID.Valid {
		return false
	}

	// check weapons
	if len(m.Weapons) < 1 ||
		m.Weapons[0] == nil ||
		!m.Weapons[0].LimitedReleaseTokenID.Valid {
		return false
	}
	if len(m.Weapons) < 2 ||
		m.Weapons[1] == nil ||
		!m.Weapons[1].LimitedReleaseTokenID.Valid {
		return false
	}
	if len(m.Weapons) < 3 ||
		m.Weapons[2] == nil ||
		!m.Weapons[2].LimitedReleaseTokenID.Valid {
		return false
	}

	// check power core
	if m.PowerCore == nil ||
		!m.PowerCore.LimitedReleaseTokenID.Valid {
		return false
	}

	// check skin
	if m.ChassisSkin == nil ||
		!m.ChassisSkin.LimitedReleaseTokenID.Valid {
		return false
	}
	return true
}

// SetBoostedStats takes the attached skin level and sets the boosted stats depending on the mechs boosted stat
func (m *Mech) SetBoostedStats() error {
	if m.ChassisSkin == nil {
		return fmt.Errorf("missing mech skin object")
	}
	// get the % increase or decrease
	positiveBoostPercent := (float32(m.ChassisSkin.Level) / 100) + 1
	negativeBoostPercent := 1 - (float32(m.ChassisSkin.Level) / 100)

	if m.BoostedStat == boiler.BoostStatMECH_SPEED {
		m.BoostedSpeed = int(positiveBoostPercent * float32(m.Speed)) // set the boosted stat
	} else {
		m.BoostedSpeed = m.Speed // set boosted speed to the speed, means we can always just use boosted stat instead of figuring out which one is better down the line
	}
	if m.BoostedStat == boiler.BoostStatMECH_HEALTH {
		m.BoostedMaxHitpoints = int(positiveBoostPercent * float32(m.MaxHitpoints))
	} else {
		m.BoostedMaxHitpoints = m.MaxHitpoints
	}
	if m.BoostedStat == boiler.BoostStatSHIELD_REGEN {
		m.BoostedShieldRechargeRate = int(positiveBoostPercent * float32(m.ShieldRechargeRate))
	} else {
		m.BoostedShieldRechargeRate = m.ShieldRechargeRate
	}
	if m.BoostedStat == boiler.BoostStatMECH_MAX_SHIELD {
		m.BoostedShield = int(positiveBoostPercent * float32(m.Shield))
	} else {
		m.BoostedShield = m.Shield
	}
	if m.BoostedStat == boiler.BoostStatMECH_SPRINT_SPREAD_MODIFIER {
		m.BoostedSprintSpreadModifier = m.SprintSpreadModifier.Mul(decimal.NewFromFloat32(negativeBoostPercent))
	} else {
		m.BoostedSprintSpreadModifier = m.SprintSpreadModifier
	}
	if m.BoostedStat == boiler.BoostStatMECH_WALK_SPEED_MODIFIER {
		m.BoostedWalkSpeedModifier = m.WalkSpeedModifier.Mul(decimal.NewFromFloat32(negativeBoostPercent))
	} else {
		m.BoostedWalkSpeedModifier = m.WalkSpeedModifier
	}
	if m.BoostedStat == boiler.BoostStatWEAPON_DAMAGE_FALLOFF {
		for _, w := range m.Weapons {
			if !w.DamageFalloff.Valid {
				continue
			}
			w.BoostedDamageFalloff = null.IntFrom(int(negativeBoostPercent * float32(w.DamageFalloff.Int)))
		}
	} else {
		for _, w := range m.Weapons {
			if !w.DamageFalloff.Valid {
				continue
			}
			w.BoostedDamageFalloff = w.DamageFalloff
		}
	}
	if m.BoostedStat == boiler.BoostStatWEAPON_SPREAD {
		for _, w := range m.Weapons {
			if !w.Spread.Valid {
				continue
			}
			w.BoostedSpread = decimal.NewNullDecimal(decimal.NewFromFloat32(negativeBoostPercent).Mul(w.Spread.Decimal))
		}
	} else {
		for _, w := range m.Weapons {
			if !w.Spread.Valid {
				continue
			}
			w.BoostedSpread = w.Spread
		}
	}

	return nil
}
