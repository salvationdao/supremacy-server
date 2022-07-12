package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/volatiletech/null/v8"
)

// Utility is the struct that rpc expects for utility
type Utility struct {
	*CollectionItem
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

	Shield      *UtilityShield      `json:"shield,omitempty"`
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
	UtilityID          string `json:"utility_id"`
	Hitpoints          int    `json:"hitpoints"`
	RechargeRate       int    `json:"recharge_rate"`
	RechargeEnergyCost int    `json:"recharge_energy_cost"`
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

func BlueprintUtilityShieldFromBoiler(utility *boiler.BlueprintUtility, shield *boiler.BlueprintUtilityShield) *BlueprintUtility {
	return &BlueprintUtility{
		ID:         utility.ID,
		BrandID:    utility.BrandID,
		Label:      utility.Label,
		UpdatedAt:  utility.UpdatedAt,
		CreatedAt:  utility.CreatedAt,
		Type:       utility.Type,
		Collection: utility.Collection,
		Tier:       utility.Tier,
		ShieldBlueprint: &BlueprintUtilityShield{
			ID:                 shield.ID,
			BlueprintUtilityID: shield.BlueprintUtilityID,
			Hitpoints:          shield.Hitpoints,
			RechargeRate:       shield.RechargeRate,
			RechargeEnergyCost: shield.RechargeEnergyCost,
			CreatedAt:          shield.CreatedAt,
		},
		ImageURL:         utility.ImageURL,
		AnimationURL:     utility.AnimationURL,
		CardAnimationURL: utility.CardAnimationURL,
		LargeImageURL:    utility.LargeImageURL,
		AvatarURL:        utility.AvatarURL,
	}
}

func BlueprintUtilityAttackDroneFromBoiler(utility *boiler.BlueprintUtility, drone *boiler.BlueprintUtilityAttackDrone) *BlueprintUtility {
	return &BlueprintUtility{
		ID:         utility.ID,
		BrandID:    utility.BrandID,
		Label:      utility.Label,
		UpdatedAt:  utility.UpdatedAt,
		CreatedAt:  utility.CreatedAt,
		Type:       utility.Type,
		Collection: utility.Collection,
		Tier:       utility.Tier,
		AttackDroneBlueprint: &BlueprintUtilityAttackDrone{
			ID:                 drone.ID,
			BlueprintUtilityID: drone.BlueprintUtilityID,
			Damage:             drone.Damage,
			RateOfFire:         drone.RateOfFire,
			Hitpoints:          drone.Hitpoints,
			LifespanSeconds:    drone.LifespanSeconds,
			DeployEnergyCost:   drone.DeployEnergyCost,
			CreatedAt:          drone.CreatedAt,
		},
		ImageURL:         utility.ImageURL,
		AnimationURL:     utility.AnimationURL,
		CardAnimationURL: utility.CardAnimationURL,
		LargeImageURL:    utility.LargeImageURL,
		AvatarURL:        utility.AvatarURL,
	}
}

func BlueprintUtilityRepairDroneFromBoiler(utility *boiler.BlueprintUtility, drone *boiler.BlueprintUtilityRepairDrone) *BlueprintUtility {
	return &BlueprintUtility{
		ID:         utility.ID,
		BrandID:    utility.BrandID,
		Label:      utility.Label,
		UpdatedAt:  utility.UpdatedAt,
		CreatedAt:  utility.CreatedAt,
		Type:       utility.Type,
		Collection: utility.Collection,
		Tier:       utility.Tier,
		RepairDroneBlueprint: &BlueprintUtilityRepairDrone{
			ID:                 drone.ID,
			BlueprintUtilityID: drone.BlueprintUtilityID,
			RepairType:         drone.RepairType,
			RepairAmount:       drone.RepairAmount,
			DeployEnergyCost:   drone.DeployEnergyCost,
			LifespanSeconds:    drone.LifespanSeconds,
			CreatedAt:          drone.CreatedAt,
		},
		ImageURL:         utility.ImageURL,
		AnimationURL:     utility.AnimationURL,
		CardAnimationURL: utility.CardAnimationURL,
		LargeImageURL:    utility.LargeImageURL,
		AvatarURL:        utility.AvatarURL,
	}
}

func BlueprintUtilityAntiMissileFromBoiler(utility *boiler.BlueprintUtility, anti *boiler.BlueprintUtilityAntiMissile) *BlueprintUtility {
	return &BlueprintUtility{
		ID:         utility.ID,
		BrandID:    utility.BrandID,
		Label:      utility.Label,
		UpdatedAt:  utility.UpdatedAt,
		CreatedAt:  utility.CreatedAt,
		Type:       utility.Type,
		Collection: utility.Collection,
		Tier:       utility.Tier,
		AntiMissileBlueprint: &BlueprintUtilityAntiMissile{
			ID:                 anti.ID,
			BlueprintUtilityID: anti.BlueprintUtilityID,
			RateOfFire:         anti.RateOfFire,
			FireEnergyCost:     anti.FireEnergyCost,
			CreatedAt:          anti.CreatedAt,
		},
		ImageURL:         utility.ImageURL,
		AnimationURL:     utility.AnimationURL,
		CardAnimationURL: utility.CardAnimationURL,
		LargeImageURL:    utility.LargeImageURL,
		AvatarURL:        utility.AvatarURL,
	}
}

func BlueprintUtilityAcceleratorFromBoiler(utility *boiler.BlueprintUtility, anti *boiler.BlueprintUtilityAccelerator) *BlueprintUtility {
	return &BlueprintUtility{
		ID:         utility.ID,
		BrandID:    utility.BrandID,
		Label:      utility.Label,
		UpdatedAt:  utility.UpdatedAt,
		CreatedAt:  utility.CreatedAt,
		Type:       utility.Type,
		Collection: utility.Collection,
		Tier:       utility.Tier,
		AcceleratorBlueprint: &BlueprintUtilityAccelerator{
			ID:                 anti.ID,
			BlueprintUtilityID: anti.BlueprintUtilityID,
			EnergyCost:         anti.EnergyCost,
			BoostSeconds:       anti.BoostSeconds,
			BoostAmount:        anti.BoostAmount,
			CreatedAt:          anti.CreatedAt,
		},
		ImageURL:         utility.ImageURL,
		AnimationURL:     utility.AnimationURL,
		CardAnimationURL: utility.CardAnimationURL,
		LargeImageURL:    utility.LargeImageURL,
		AvatarURL:        utility.AvatarURL,
	}
}

func UtilityShieldFromBoiler(utility *boiler.Utility, shield *boiler.UtilityShield, collection *boiler.CollectionItem) *Utility {
	return &Utility{
		CollectionItem: &CollectionItem{
			CollectionSlug:   collection.CollectionSlug,
			Hash:             collection.Hash,
			TokenID:          collection.TokenID,
			ItemType:         collection.ItemType,
			ItemID:           collection.ItemID,
			Tier:             collection.Tier,
			OwnerID:          collection.OwnerID,
			MarketLocked:     collection.MarketLocked,
			XsynLocked:       collection.XsynLocked,
			AssetHidden:      collection.AssetHidden,
			ImageURL:         collection.ImageURL,
			CardAnimationURL: collection.CardAnimationURL,
			AvatarURL:        collection.AvatarURL,
			LargeImageURL:    collection.LargeImageURL,
			BackgroundColor:  collection.BackgroundColor,
			AnimationURL:     collection.AnimationURL,
			YoutubeURL:       collection.YoutubeURL,
		},
		ID:             utility.ID,
		BrandID:        utility.BrandID,
		Label:          utility.Label,
		UpdatedAt:      utility.UpdatedAt,
		CreatedAt:      utility.CreatedAt,
		BlueprintID:    utility.BlueprintID,
		GenesisTokenID: utility.GenesisTokenID,
		EquippedOn:     utility.EquippedOn,
		Type:           utility.Type,
		Shield: &UtilityShield{
			UtilityID:          shield.UtilityID,
			Hitpoints:          shield.Hitpoints,
			RechargeRate:       shield.RechargeRate,
			RechargeEnergyCost: shield.RechargeEnergyCost,
		},
	}
}

func UtilityAttackDroneFromBoiler(utility *boiler.Utility, drone *boiler.UtilityAttackDrone, collection *boiler.CollectionItem) *Utility {
	return &Utility{
		CollectionItem: &CollectionItem{
			CollectionSlug:   collection.CollectionSlug,
			Hash:             collection.Hash,
			TokenID:          collection.TokenID,
			ItemType:         collection.ItemType,
			ItemID:           collection.ItemID,
			Tier:             collection.Tier,
			OwnerID:          collection.OwnerID,
			MarketLocked:     collection.MarketLocked,
			XsynLocked:       collection.XsynLocked,
			AssetHidden:      collection.AssetHidden,
			ImageURL:         collection.ImageURL,
			CardAnimationURL: collection.CardAnimationURL,
			AvatarURL:        collection.AvatarURL,
			LargeImageURL:    collection.LargeImageURL,
			BackgroundColor:  collection.BackgroundColor,
			AnimationURL:     collection.AnimationURL,
			YoutubeURL:       collection.YoutubeURL,
		},
		ID:             utility.ID,
		BrandID:        utility.BrandID,
		Label:          utility.Label,
		UpdatedAt:      utility.UpdatedAt,
		CreatedAt:      utility.CreatedAt,
		BlueprintID:    utility.BlueprintID,
		GenesisTokenID: utility.GenesisTokenID,
		EquippedOn:     utility.EquippedOn,
		Type:           utility.Type,
		AttackDrone: &UtilityAttackDrone{
			UtilityID:        drone.UtilityID,
			Damage:           drone.Damage,
			RateOfFire:       drone.RateOfFire,
			Hitpoints:        drone.Hitpoints,
			LifespanSeconds:  drone.LifespanSeconds,
			DeployEnergyCost: drone.DeployEnergyCost,
		},
	}
}

func UtilityRepairDroneFromBoiler(utility *boiler.Utility, drone *boiler.UtilityRepairDrone, collection *boiler.CollectionItem) *Utility {
	return &Utility{
		CollectionItem: &CollectionItem{
			CollectionSlug:   collection.CollectionSlug,
			Hash:             collection.Hash,
			TokenID:          collection.TokenID,
			ItemType:         collection.ItemType,
			ItemID:           collection.ItemID,
			Tier:             collection.Tier,
			OwnerID:          collection.OwnerID,
			MarketLocked:     collection.MarketLocked,
			XsynLocked:       collection.XsynLocked,
			AssetHidden:      collection.AssetHidden,
			ImageURL:         collection.ImageURL,
			CardAnimationURL: collection.CardAnimationURL,
			AvatarURL:        collection.AvatarURL,
			LargeImageURL:    collection.LargeImageURL,
			BackgroundColor:  collection.BackgroundColor,
			AnimationURL:     collection.AnimationURL,
			YoutubeURL:       collection.YoutubeURL,
		},
		ID:             utility.ID,
		BrandID:        utility.BrandID,
		Label:          utility.Label,
		UpdatedAt:      utility.UpdatedAt,
		CreatedAt:      utility.CreatedAt,
		BlueprintID:    utility.BlueprintID,
		GenesisTokenID: utility.GenesisTokenID,
		EquippedOn:     utility.EquippedOn,
		Type:           utility.Type,
		RepairDrone: &UtilityRepairDrone{
			UtilityID:        drone.UtilityID,
			RepairType:       drone.RepairType,
			RepairAmount:     drone.RepairAmount,
			DeployEnergyCost: drone.DeployEnergyCost,
			LifespanSeconds:  drone.LifespanSeconds,
		},
	}
}

func UtilityAntiMissileFromBoiler(utility *boiler.Utility, anti *boiler.UtilityAntiMissile, collection *boiler.CollectionItem) *Utility {
	return &Utility{
		CollectionItem: &CollectionItem{
			CollectionSlug:   collection.CollectionSlug,
			Hash:             collection.Hash,
			TokenID:          collection.TokenID,
			ItemType:         collection.ItemType,
			ItemID:           collection.ItemID,
			Tier:             collection.Tier,
			OwnerID:          collection.OwnerID,
			MarketLocked:     collection.MarketLocked,
			XsynLocked:       collection.XsynLocked,
			AssetHidden:      collection.AssetHidden,
			ImageURL:         collection.ImageURL,
			CardAnimationURL: collection.CardAnimationURL,
			AvatarURL:        collection.AvatarURL,
			LargeImageURL:    collection.LargeImageURL,
			BackgroundColor:  collection.BackgroundColor,
			AnimationURL:     collection.AnimationURL,
			YoutubeURL:       collection.YoutubeURL,
		},
		ID:             utility.ID,
		BrandID:        utility.BrandID,
		Label:          utility.Label,
		UpdatedAt:      utility.UpdatedAt,
		CreatedAt:      utility.CreatedAt,
		BlueprintID:    utility.BlueprintID,
		GenesisTokenID: utility.GenesisTokenID,
		EquippedOn:     utility.EquippedOn,
		Type:           utility.Type,
		AntiMissile: &UtilityAntiMissile{
			UtilityID:      anti.UtilityID,
			RateOfFire:     anti.RateOfFire,
			FireEnergyCost: anti.FireEnergyCost,
		},
	}
}

func UtilityAcceleratorFromBoiler(utility *boiler.Utility, anti *boiler.UtilityAccelerator, collection *boiler.CollectionItem) *Utility {
	return &Utility{
		CollectionItem: &CollectionItem{
			CollectionSlug:   collection.CollectionSlug,
			Hash:             collection.Hash,
			TokenID:          collection.TokenID,
			ItemType:         collection.ItemType,
			ItemID:           collection.ItemID,
			Tier:             collection.Tier,
			OwnerID:          collection.OwnerID,
			MarketLocked:     collection.MarketLocked,
			XsynLocked:       collection.XsynLocked,
			AssetHidden:      collection.AssetHidden,
			ImageURL:         collection.ImageURL,
			CardAnimationURL: collection.CardAnimationURL,
			AvatarURL:        collection.AvatarURL,
			LargeImageURL:    collection.LargeImageURL,
			BackgroundColor:  collection.BackgroundColor,
			AnimationURL:     collection.AnimationURL,
			YoutubeURL:       collection.YoutubeURL,
		},
		ID:             utility.ID,
		BrandID:        utility.BrandID,
		Label:          utility.Label,
		UpdatedAt:      utility.UpdatedAt,
		CreatedAt:      utility.CreatedAt,
		BlueprintID:    utility.BlueprintID,
		GenesisTokenID: utility.GenesisTokenID,
		EquippedOn:     utility.EquippedOn,
		Type:           utility.Type,
		Accelerator: &UtilityAccelerator{
			UtilityID:    anti.UtilityID,
			EnergyCost:   anti.EnergyCost,
			BoostSeconds: anti.BoostSeconds,
			BoostAmount:  anti.BoostAmount,
		},
	}
}
