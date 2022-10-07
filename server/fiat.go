package server

import (
	"server/db/boiler"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type FiatCurrencyCode string

const (
	FiatCurrencyCodeUSD  = "USD"
	FiatCurrencyCodeSUPS = "SUPS"
)

type FiatProductType string

const (
	FiatProductTypeGeneric      = "generic"
	FiatProductTypeMysteryCrate = "mystery_crate"
)

type FiatOrder struct {
	ID            string           `json:"id"`
	OrderNumber   int64            `json:"order_number"`
	UserID        string           `json:"user_id"`
	OrderStatus   string           `json:"order_status"`
	PaymentMethod string           `json:"payment_method"`
	TXNReference  string           `json:"txn_reference"`
	Currency      string           `json:"currency"`
	Items         []*FiatOrderItem `json:"items"`
	CreatedAt     time.Time        `json:"created_at"`
}

type FiatOrderItem struct {
	ID            string          `json:"id"`
	OrderID       string          `json:"order_id"`
	FiatProductID string          `json:"fiat_product_id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Quantity      int             `json:"quantity"`
	Amount        decimal.Decimal `json:"amount"`
}

type FiatProduct struct {
	ID          string                `json:"id"`
	FactionID   string                `json:"faction_id"`
	ProductType string                `json:"product_type"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Pricing     []*FiatProductPricing `json:"pricing"`
	AvatarURL   null.String           `json:"avatar_url"`
}

type FiatProductPricing struct {
	CurrencyCode string          `json:"currency_code"`
	Amount       decimal.Decimal `json:"amount"`
}

func FiatProductFromBoiler(p *boiler.FiatProduct) *FiatProduct {
	if p == nil {
		return nil
	}
	output := &FiatProduct{
		ID:          p.ID,
		ProductType: p.ProductType,
		Name:        p.Name,
		Description: p.Description,
	}
	if p.R != nil && len(p.R.FiatProductPricings) > 0 {
		pricing := []*FiatProductPricing{}
		for _, price := range p.R.FiatProductPricings {
			pricing = append(pricing, &FiatProductPricing{
				CurrencyCode: price.CurrencyCode,
				Amount:       price.Amount,
			})
		}
		output.Pricing = pricing
	}
	return output
}

// ShoppingCart contains a user's shopping cart.
type ShoppingCart struct {
	ID        string              `json:"-"`
	UserID    string              `json:"user_id"`
	Items     []*ShoppingCartItem `json:"items"`
	CreatedAt time.Time           `json:"created_at"`
	ExpiresAt time.Time           `json:"expires_at"`
}

// ShoppingCartItem holds a single product on shopping cart.
type ShoppingCartItem struct {
	ID       string       `json:"id"`
	Quantity int          `json:"quantity"`
	Product  *FiatProduct `json:"product"`
}

func ShoppingCartFromBoiler(sc *boiler.ShoppingCart, items []*boiler.ShoppingCartItem) *ShoppingCart {
	output := &ShoppingCart{
		ID:        sc.ID,
		UserID:    sc.UserID,
		CreatedAt: sc.CreatedAt,
		ExpiresAt: sc.ExpiresAt,
		Items:     []*ShoppingCartItem{},
	}
	for _, sci := range items {
		outputItem := &ShoppingCartItem{
			ID:       sci.ID,
			Quantity: sci.Quantity,
		}
		if sci.R != nil && sci.R.Product != nil {
			outputItem.Product = FiatProductFromBoiler(sci.R.Product)
			if sci.R.Product.ProductType == boiler.FiatProductTypesMysteryCrate && sci.R.Product.R != nil && sci.R.Product.R.StorefrontMysteryCrate != nil {
				outputItem.Product.AvatarURL = sci.R.Product.R.StorefrontMysteryCrate.AvatarURL
			}
		}
		output.Items = append(output.Items, outputItem)
	}
	return output
}
