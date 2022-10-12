package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type PowerCore struct {
	*CollectionItem
	*Images
	ID                    string          `json:"id"`
	BlueprintID           string          `json:"blueprint_id"`
	Label                 string          `json:"label"`
	Size                  string          `json:"size"`
	Capacity              decimal.Decimal `json:"capacity"`
	MaxDrawRate           decimal.Decimal `json:"max_draw_rate"`
	RechargeRate          decimal.Decimal `json:"recharge_rate"`
	Armour                decimal.Decimal `json:"armour"`
	MaxHitpoints          decimal.Decimal `json:"max_hitpoints"`
	EquippedOn            null.String     `json:"equipped_on,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	GenesisTokenID        null.Int64      `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64      `json:"limited_release_token_id,omitempty"`

	EquippedOnDetails *EquippedOnDetails
}

func (b *PowerCore) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

type BlueprintPowerCore struct {
	ID               string          `json:"id"`
	Collection       string          `json:"collection"`
	Label            string          `json:"label"`
	Size             string          `json:"size"`
	Capacity         decimal.Decimal `json:"capacity"`
	MaxDrawRate      decimal.Decimal `json:"max_draw_rate"`
	RechargeRate     decimal.Decimal `json:"recharge_rate"`
	Armour           decimal.Decimal `json:"armour"`
	MaxHitpoints     decimal.Decimal `json:"max_hitpoints"`
	Tier             string          `json:"tier,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	ImageURL         null.String     `json:"image_url,omitempty"`
	CardAnimationURL null.String     `json:"card_animation_url,omitempty"`
	AvatarURL        null.String     `json:"avatar_url,omitempty"`
	LargeImageURL    null.String     `json:"large_image_url,omitempty"`
	BackgroundColor  null.String     `json:"background_color,omitempty"`
	AnimationURL     null.String     `json:"animation_url,omitempty"`
	YoutubeURL       null.String     `json:"youtube_url,omitempty"`

	// only used on inserting new mechs/items, since we are still giving away some limited released and genesis
	GenesisTokenID        null.Int64 `json:"genesis_token_id,omitempty"`
	LimitedReleaseTokenID null.Int64 `json:"limited_release_token_id,omitempty"`
}

func (b *BlueprintPowerCore) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

func BlueprintPowerCoreFromBoiler(core *boiler.BlueprintPowerCore) *BlueprintPowerCore {
	return &BlueprintPowerCore{
		ID:               core.ID,
		Collection:       core.Collection,
		Label:            core.Label,
		Size:             core.Size,
		Capacity:         core.Capacity,
		MaxDrawRate:      core.MaxDrawRate,
		RechargeRate:     core.RechargeRate,
		Armour:           core.Armour,
		MaxHitpoints:     core.MaxHitpoints,
		Tier:             core.Tier,
		ImageURL:         core.ImageURL,
		AnimationURL:     core.AnimationURL,
		CardAnimationURL: core.CardAnimationURL,
		LargeImageURL:    core.LargeImageURL,
		AvatarURL:        core.AvatarURL,
		CreatedAt:        core.CreatedAt,
	}
}

func PowerCoreFromBoiler(skin *boiler.PowerCore, collection *boiler.CollectionItem) *PowerCore {
	return &PowerCore{
		CollectionItem: &CollectionItem{
			CollectionSlug: collection.CollectionSlug,
			Hash:           collection.Hash,
			TokenID:        collection.TokenID,
			ItemType:       collection.ItemType,
			ItemID:         collection.ItemID,
			Tier:           collection.Tier,
			OwnerID:        collection.OwnerID,
			MarketLocked:   collection.MarketLocked,
			XsynLocked:     collection.XsynLocked,
			AssetHidden:    collection.AssetHidden,
		},
		Images: &Images{
			ImageURL:         skin.R.Blueprint.ImageURL,
			CardAnimationURL: skin.R.Blueprint.CardAnimationURL,
			AvatarURL:        skin.R.Blueprint.AvatarURL,
			LargeImageURL:    skin.R.Blueprint.LargeImageURL,
			BackgroundColor:  skin.R.Blueprint.BackgroundColor,
			AnimationURL:     skin.R.Blueprint.AnimationURL,
			YoutubeURL:       skin.R.Blueprint.YoutubeURL,
		},
		ID:           skin.ID,
		BlueprintID:  skin.BlueprintID,
		Label:        skin.R.Blueprint.Label,
		Size:         skin.R.Blueprint.Size,
		Capacity:     skin.R.Blueprint.Capacity,
		MaxDrawRate:  skin.R.Blueprint.MaxDrawRate,
		RechargeRate: skin.R.Blueprint.RechargeRate,
		Armour:       skin.R.Blueprint.Armour,
		MaxHitpoints: skin.R.Blueprint.MaxHitpoints,
		EquippedOn:   skin.EquippedOn,
		CreatedAt:    skin.CreatedAt,
	}
}
