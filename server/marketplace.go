package server

import (
	"server/db/boiler"

	"github.com/volatiletech/null/v8"
)

type MarketplaceSaleItem struct {
	boiler.ItemSale `boil:",bind"`
	Owner           MarketplaceSaleItemOwner `json:"owner,omitempty" boil:"players,bind"`
	Mech            MarketplaceSaleItemMech  `json:"mech,omitempty" boil:"mech,bind"`
}

type MarketplaceSaleItemOwner struct {
	ID            string      `json:"id" boil:"players.id"`
	Username      null.String `json:"username" boil:"players.username"`
	PublicAddress null.String `json:"public_address" boil:"players.public_address"`
}

type MarketplaceSaleItemMech struct {
	ID    string `json:"id" boil:"mechs.id"`
	Label string `json:"label" boil:"mechs.label"`
	Name  string `json:"name" boil:"mechs.name"`
}

type MarketplaceKeycardSaleItem struct {
	*boiler.ItemKeycardSale
	Owner      *boiler.Player         `json:"owner"`
	Blueprints *AssetKeycardBlueprint `json:"blueprints"`
}
