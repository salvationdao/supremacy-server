package server

import (
	"encoding/json"
	"fmt"
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type MysteryCrate struct {
	*CollectionItem
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	FactionID   string    `json:"faction_id"`
	Label       string    `json:"label"`
	Opened      bool      `json:"opened"`
	LockedUntil time.Time `json:"locked_until"`
	Purchased   bool      `json:"purchased"`
	Description string    `json:"description"`

	DeletedAt null.Time `json:"deleted_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (b *MysteryCrate) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

func MysteryCrateFromBoiler(mysteryCrate *boiler.MysteryCrate, collection *boiler.CollectionItem) *MysteryCrate {
	return &MysteryCrate{
		CollectionItem: &CollectionItem{
			CollectionSlug:      collection.CollectionSlug,
			Hash:                collection.Hash,
			TokenID:             collection.TokenID,
			ItemType:            collection.ItemType,
			ItemID:              collection.ItemID,
			Tier:                collection.Tier,
			OwnerID:             collection.OwnerID,
			MarketLocked:        collection.MarketLocked,
			XsynLocked:          collection.XsynLocked,
			LockedToMarketplace: collection.LockedToMarketplace,
			ImageURL:            collection.ImageURL,
			CardAnimationURL:    collection.CardAnimationURL,
			AvatarURL:           collection.AvatarURL,
			LargeImageURL:       collection.LargeImageURL,
			BackgroundColor:     collection.BackgroundColor,
			AnimationURL:        collection.AnimationURL,
			YoutubeURL:          collection.YoutubeURL,
		},
		ID:          mysteryCrate.ID,
		Type:        mysteryCrate.Type,
		FactionID:   mysteryCrate.FactionID,
		Label:       mysteryCrate.Label,
		Opened:      mysteryCrate.Opened,
		LockedUntil: mysteryCrate.LockedUntil,
		Purchased:   mysteryCrate.Purchased,
		DeletedAt:   mysteryCrate.DeletedAt,
		UpdatedAt:   mysteryCrate.UpdatedAt,
		CreatedAt:   mysteryCrate.CreatedAt,
		Description: mysteryCrate.Description,
	}
}

type StorefrontMysteryCrate struct {
	ID               string          `json:"id"`
	MysteryCrateType string          `json:"mystery_crate_type"`
	Price            decimal.Decimal `json:"price"`
	Amount           int             `json:"amount"`
	AmountSold       int             `json:"amount_sold"`
	FactionID        string          `json:"faction_id"`
	Label            string          `json:"label"`
	Description      string          `json:"description"`
	ImageURL         null.String     `json:"image_url,omitempty"`
	CardAnimationURL null.String     `json:"card_animation_url,omitempty"`
	AvatarURL        null.String     `json:"avatar_url,omitempty"`
	LargeImageURL    null.String     `json:"large_image_url,omitempty"`
	BackgroundColor  null.String     `json:"background_color,omitempty"`
	AnimationURL     null.String     `json:"animation_url,omitempty"`
	YoutubeURL       null.String     `json:"youtube_url,omitempty"`
}

func (b *StorefrontMysteryCrate) Scan(value interface{}) error {
	v, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan value into byte array")
	}
	return json.Unmarshal(v, b)
}

func StoreFrontMysteryCrateFromBoiler(storefrontMysteryCrate *boiler.StorefrontMysteryCrate) *StorefrontMysteryCrate {
	return &StorefrontMysteryCrate{
		ID:               storefrontMysteryCrate.ID,
		MysteryCrateType: storefrontMysteryCrate.MysteryCrateType,
		Price:            storefrontMysteryCrate.Price,
		Amount:           storefrontMysteryCrate.Amount,
		AmountSold:       storefrontMysteryCrate.AmountSold,
		FactionID:        storefrontMysteryCrate.FactionID,
		Label:            storefrontMysteryCrate.Label,
		Description:      storefrontMysteryCrate.Description,
		ImageURL:         storefrontMysteryCrate.ImageURL,
		CardAnimationURL: storefrontMysteryCrate.CardAnimationURL,
		AvatarURL:        storefrontMysteryCrate.AvatarURL,
		LargeImageURL:    storefrontMysteryCrate.LargeImageURL,
		BackgroundColor:  storefrontMysteryCrate.BackgroundColor,
		AnimationURL:     storefrontMysteryCrate.AnimationURL,
		YoutubeURL:       storefrontMysteryCrate.YoutubeURL,
	}
}

func StoreFrontMysteryCrateSliceFromBoiler(storefrontMysteryCrates []*boiler.StorefrontMysteryCrate) []*StorefrontMysteryCrate {
	var slice []*StorefrontMysteryCrate
	for _, mc := range storefrontMysteryCrates {
		slice = append(slice, StoreFrontMysteryCrateFromBoiler(mc))
	}

	return slice
}