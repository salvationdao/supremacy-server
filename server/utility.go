package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/volatiletech/null/v8"
)

// Utility is the struct that rpc expects for utility
type Utility struct {
	*CollectionItem
	*Images
	ID                    string      `json:"id"`
	BrandID               null.String `json:"brand_id,omitempty"`
	Label                 string      `json:"label"`
	UpdatedAt             time.Time   `json:"updated_at"`
	CreatedAt             time.Time   `json:"created_at"`
	BlueprintID           string      `json:"blueprint_id"`
	GenesisTokenID        null.Int64  `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64  `json:"limited_release_token_id,omitempty"`
	EquippedOn            null.String `json:"equipped_on,omitempty"`
	Type                  string      `json:"type"`
	LockedToMech          bool        `json:"locked_to_mech"`
	SlotNumber            null.Int    `json:"slot_number,omitempty"`

	AttackDrone *UtilityAttackDrone `json:"attack_drone,omitempty"`
	RepairDrone *UtilityRepairDrone `json:"repair_drone,omitempty"`
	Accelerator *UtilityAccelerator `json:"accelerator,omitempty"`
	AntiMissile *UtilityAntiMissile `json:"anti_missile,omitempty"`

	EquippedOnDetails *EquippedOnDetails
}

func (b *Utility) Scan(value interface{}) error {
	if value == nil {
		//b = nil
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type UtilitySlice []*Utility

func (b *UtilitySlice) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type UtilityAttackDrone struct {
	UtilityID        string `json:"utility_id"`
	Damage           int    `json:"damage"`
	RateOfFire       int    `json:"rate_of_fire"`
	Hitpoints        int    `json:"hitpoints"`
	LifespanSeconds  int    `json:"lifespan_seconds"`
	DeployEnergyCost int    `json:"deploy_energy_cost"`
}

func (b *UtilityAttackDrone) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type UtilityShield struct {
	UtilityID           string `json:"utility_id"`
	Hitpoints           int    `json:"hitpoints"`
	BoostedHitpoints    int    `json:"boosted_hitpoints"`
	RechargeRate        int    `json:"recharge_rate"`
	BoostedRechargeRate int    `json:"boosted_recharge_rate"`
	RechargeEnergyCost  int    `json:"recharge_energy_cost"`
}

func (b *UtilityShield) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type UtilityRepairDrone struct {
	UtilityID        string      `json:"utility_id"`
	RepairType       null.String `json:"repair_type,omitempty"`
	RepairAmount     int         `json:"repair_amount"`
	DeployEnergyCost int         `json:"deploy_energy_cost"`
	LifespanSeconds  int         `json:"lifespan_seconds"`
}

func (b *UtilityRepairDrone) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type UtilityAccelerator struct {
	UtilityID    string `json:"utility_id"`
	EnergyCost   int    `json:"energy_cost"`
	BoostSeconds int    `json:"boost_seconds"`
	BoostAmount  int    `json:"boost_amount"`
}

func (b *UtilityAccelerator) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type UtilityAntiMissile struct {
	UtilityID      string `json:"utility_id"`
	RateOfFire     int    `json:"rate_of_fire"`
	FireEnergyCost int    `json:"fire_energy_cost"`
}

func (b *UtilityAntiMissile) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
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

func (b *BlueprintUtilityAttackDrone) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type BlueprintUtilityShield struct {
	ID                 string    `json:"id"`
	BlueprintUtilityID string    `json:"blueprint_utility_id"`
	Hitpoints          int       `json:"hitpoints"`
	RechargeRate       int       `json:"recharge_rate"`
	RechargeEnergyCost int       `json:"recharge_energy_cost"`
	CreatedAt          time.Time `json:"created_at"`
}

func (b *BlueprintUtilityShield) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
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

func (b *BlueprintUtilityRepairDrone) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type BlueprintUtilityAccelerator struct {
	ID                 string    `json:"id"`
	BlueprintUtilityID string    `json:"blueprint_utility_id"`
	EnergyCost         int       `json:"energy_cost"`
	BoostSeconds       int       `json:"boost_seconds"`
	BoostAmount        int       `json:"boost_amount"`
	CreatedAt          time.Time `json:"created_at"`
}

func (b *BlueprintUtilityAccelerator) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type BlueprintUtilityAntiMissile struct {
	ID                 string    `json:"id"`
	BlueprintUtilityID string    `json:"blueprint_utility_id"`
	RateOfFire         int       `json:"rate_of_fire"`
	FireEnergyCost     int       `json:"fire_energy_cost"`
	CreatedAt          time.Time `json:"created_at"`
}

func (b *BlueprintUtilityAntiMissile) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}
