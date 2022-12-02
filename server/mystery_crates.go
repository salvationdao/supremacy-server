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
	*Images
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	FactionID   string      `json:"faction_id"`
	Label       string      `json:"label"`
	Opened      bool        `json:"opened"`
	LockedUntil time.Time   `json:"locked_until"`
	Purchased   bool        `json:"purchased"`
	Description string      `json:"description"`
	ItemSaleID  null.String `json:"item_sale_id"`

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

func MysteryCrateFromBoiler(mysteryCrate *boiler.MysteryCrate, collection *boiler.CollectionItem, itemSaleID null.String) *MysteryCrate {
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
			AssetHidden:         collection.AssetHidden,
			LockedToMarketplace: collection.LockedToMarketplace,
		},
		Images: &Images{
			ImageURL:         mysteryCrate.R.Blueprint.ImageURL,
			CardAnimationURL: mysteryCrate.R.Blueprint.CardAnimationURL,
			AvatarURL:        mysteryCrate.R.Blueprint.AvatarURL,
			LargeImageURL:    mysteryCrate.R.Blueprint.LargeImageURL,
			BackgroundColor:  mysteryCrate.R.Blueprint.BackgroundColor,
			AnimationURL:     mysteryCrate.R.Blueprint.AnimationURL,
			YoutubeURL:       mysteryCrate.R.Blueprint.YoutubeURL,
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
		ItemSaleID:  itemSaleID,
	}
}

type StorefrontMysteryCrate struct {
	ID               string          `json:"id"`
	FiatProductID    string          `json:"fiat_product_id"`
	FiatProduct      *FiatProduct    `json:"fiat_product"`
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
	output := &StorefrontMysteryCrate{
		ID:               storefrontMysteryCrate.ID,
		FiatProductID:    storefrontMysteryCrate.FiatProductID,
		MysteryCrateType: storefrontMysteryCrate.MysteryCrateType,
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
	if storefrontMysteryCrate.R != nil && storefrontMysteryCrate.R.FiatProduct != nil {
		pricing := []*FiatProductPricing{}
		if storefrontMysteryCrate.R.FiatProduct.R != nil && len(storefrontMysteryCrate.R.FiatProduct.R.FiatProductPricings) > 0 {
			for _, p := range storefrontMysteryCrate.R.FiatProduct.R.FiatProductPricings {
				item := &FiatProductPricing{
					CurrencyCode: p.CurrencyCode,
					Amount:       p.Amount,
				}
				pricing = append(pricing, item)
				if p.CurrencyCode == FiatCurrencyCodeSUPS {
					output.Price = p.Amount
				}
			}
		}
		output.FiatProduct = &FiatProduct{
			ID:          storefrontMysteryCrate.FiatProductID,
			ProductType: FiatProductTypeMysteryCrate,
			Name:        storefrontMysteryCrate.R.FiatProduct.Name,
			Description: storefrontMysteryCrate.R.FiatProduct.Description,
			Pricing:     pricing,
		}
	}
	return output
}

func StoreFrontMysteryCrateSliceFromBoiler(storefrontMysteryCrates []*boiler.StorefrontMysteryCrate) []*StorefrontMysteryCrate {
	var slice []*StorefrontMysteryCrate
	for _, mc := range storefrontMysteryCrates {
		slice = append(slice, StoreFrontMysteryCrateFromBoiler(mc))
	}

	return slice
}
