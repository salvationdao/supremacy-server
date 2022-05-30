package server

import (
	"server/db/boiler"
)

type MarketplaceSaleItem struct {
	*boiler.ItemSale
	Owner      *boiler.Player         `json:"owner"`
	Collection *boiler.CollectionItem `json:"collection"`
	Mech       *boiler.Mech           `json:"mech,omitempty"`
}
