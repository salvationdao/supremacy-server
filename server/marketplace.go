package server

import (
	"server/db/boiler"
)

type MarketplaceSaleItem struct {
	*boiler.ItemSale
	Owner      *boiler.Player         `json:"owner"`
	Collection *boiler.CollectionItem `json:"collection"`
	Mech       *Mech                  `json:"mech,omitempty"`
}

type MarketplaceKeycardSaleItem struct {
	*boiler.ItemKeycardSale
	Owner      *boiler.Player         `json:"owner"`
	Blueprints *AssetKeycardBlueprint `json:"blueprints"`
}
