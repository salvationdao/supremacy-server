package server

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"server/db/boiler"
	"time"
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
	}
}

type StorefrontMysteryCrate struct {
	ID               string          `json:"id"`
	MysteryCrateType string          `json:"mystery_crate_type"`
	Price            decimal.Decimal `json:"price"`
	Amount           int             `json:"amount"`
	AmountSold       int             `json:"amount_sold"`
	FactionID        string          `json:"faction_id"`

	DeletedAt null.Time `json:"purchased"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
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
		DeletedAt:        storefrontMysteryCrate.DeletedAt,
		UpdatedAt:        storefrontMysteryCrate.UpdatedAt,
		CreatedAt:        storefrontMysteryCrate.CreatedAt,
	}
}

func StoreFrontMysteryCrateSliceFromBoiler(storefrontMysteryCrates []*boiler.StorefrontMysteryCrate) []*StorefrontMysteryCrate {
	var slice []*StorefrontMysteryCrate
	for _, mc := range storefrontMysteryCrates {
		crate := &StorefrontMysteryCrate{
			ID:               mc.ID,
			MysteryCrateType: mc.MysteryCrateType,
			Price:            mc.Price,
			Amount:           mc.Amount,
			AmountSold:       mc.AmountSold,
			FactionID:        mc.FactionID,
			DeletedAt:        mc.DeletedAt,
			UpdatedAt:        mc.UpdatedAt,
			CreatedAt:        mc.CreatedAt,
		}
		slice = append(slice, crate)
	}

	return slice
}
