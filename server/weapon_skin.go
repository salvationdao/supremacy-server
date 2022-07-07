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
	OwnerID       string      `json:"owner_id"`
	Label         string      `json:"label"`
	WeaponType    string      `json:"weapon_type"`
	EquippedOn    null.String `json:"equipped_on,omitempty"`
	Tier          string      `json:"tier"`
	CreatedAt     time.Time   `json:"created_at"`
	WeaponModelID string      `json:"weapon_model_id"`

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
	WeaponType       string              `json:"weapon_type"`
	Tier             string              `json:"tier"`
	CreatedAt        time.Time           `json:"created_at"`
	ImageURL         null.String         `json:"image_url,omitempty"`
	CardAnimationURL null.String         `json:"card_animation_url,omitempty"`
	AvatarURL        null.String         `json:"avatar_url,omitempty"`
	LargeImageURL    null.String         `json:"large_image_url,omitempty"`
	BackgroundColor  null.String         `json:"background_color,omitempty"`
	AnimationURL     null.String         `json:"animation_url,omitempty"`
	YoutubeURL       null.String         `json:"youtube_url,omitempty"`
	Collection       string              `json:"collection"`
	WeaponModelID    string              `json:"weapon_model_id"`
	StatModifier     decimal.NullDecimal `json:"stat_modifier,omitempty"`
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
		WeaponType:       weaponSkin.WeaponType,
		Tier:             weaponSkin.Tier,
		CreatedAt:        weaponSkin.CreatedAt,
		ImageURL:         weaponSkin.ImageURL,
		CardAnimationURL: weaponSkin.CardAnimationURL,
		AvatarURL:        weaponSkin.AvatarURL,
		LargeImageURL:    weaponSkin.LargeImageURL,
		BackgroundColor:  weaponSkin.BackgroundColor,
		AnimationURL:     weaponSkin.AnimationURL,
		YoutubeURL:       weaponSkin.YoutubeURL,
		Collection:       weaponSkin.Collection,
		WeaponModelID:    weaponSkin.WeaponModelID,
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
			ImageURL:         collection.ImageURL,
			CardAnimationURL: collection.CardAnimationURL,
			AvatarURL:        collection.AvatarURL,
			LargeImageURL:    collection.LargeImageURL,
			BackgroundColor:  collection.BackgroundColor,
			AnimationURL:     collection.AnimationURL,
			YoutubeURL:       collection.YoutubeURL,
		},
		ID:            weaponSkin.ID,
		BlueprintID:   weaponSkin.BlueprintID,
		OwnerID:       weaponSkin.OwnerID,
		Label:         weaponSkin.Label,
		WeaponType:    weaponSkin.WeaponType,
		EquippedOn:    weaponSkin.EquippedOn,
		Tier:          weaponSkin.Tier,
		CreatedAt:     weaponSkin.CreatedAt,
		WeaponModelID: weaponSkin.WeaponModelID,
	}
}
