package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

// WeaponSkin is the struct that rpc expects for weapons skins
type WeaponSkin struct {
	*CollectionItem
	ID            string      `json:"id"`
	BlueprintID   string      `json:"blueprint_id"`
	Label         string      `json:"label"`
	WeaponType    string      `json:"weapon_type"`
	EquippedOn    null.String `json:"equipped_on,omitempty"`
	Tier          string      `json:"tier"`
	CreatedAt     time.Time   `json:"created_at"`

	EquippedOnDetails *EquippedOnDetails
}

func (b *WeaponSkin) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type BlueprintWeaponSkin struct {
	ID               string              `json:"id"`
	Label            string              `json:"label"`
	Tier             string              `json:"tier"`
	Collection       string              `json:"collection"`
	StatModifier     decimal.NullDecimal `json:"stat_modifier,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
}

func (b *BlueprintWeaponSkin) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type WeaponSkinSlice []*WeaponSkin

func (b *WeaponSkinSlice) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

func BlueprintWeaponSkinFromBoiler(weaponSkin *boiler.BlueprintWeaponSkin) *BlueprintWeaponSkin {
	return &BlueprintWeaponSkin{
		ID:               weaponSkin.ID,
		Label:            weaponSkin.Label,
		Tier:             weaponSkin.Tier,
		CreatedAt:        weaponSkin.CreatedAt,
		Collection:       weaponSkin.Collection,
		StatModifier:     weaponSkin.StatModifier,
	}
}

func WeaponSkinFromBoiler(weaponSkin *boiler.WeaponSkin, collection *boiler.CollectionItem) *WeaponSkin {
	return &WeaponSkin{
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
		},
		ID:            weaponSkin.ID,
		BlueprintID:   weaponSkin.BlueprintID,
		// TODO: vinnie fix me please
		//Label:         weaponSkin.Label,
		EquippedOn:    weaponSkin.EquippedOn,
		CreatedAt:     weaponSkin.CreatedAt,
	}
}
