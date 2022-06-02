package server

import (
	"server/db/boiler"

	"github.com/volatiletech/null/v8"
)

type MarketplaceSaleItem struct {
	boiler.ItemSale `boil:",bind"`
	Owner           MarketplaceSaleItemOwner `json:"owner,omitempty" boil:",bind"`
	Mech            MarketplaceSaleItemMech  `json:"mech,omitempty" boil:",bind"`
}

type MarketplaceSaleItemOwner struct {
	ID            string      `json:"id" boil:"players.id"`
	Username      null.String `json:"username" boil:"players.username"`
	PublicAddress null.String `json:"public_address" boil:"players.public_address"`
	Gid           int         `json:"gid" boil:"players.gid"`
}

type MarketplaceSaleItemMech struct {
	ID        string `json:"id" boil:"mechs.id"`
	Label     string `json:"label" boil:"mechs.label"`
	Name      string `json:"name" boil:"mechs.name"`
	Tier      string `json:"tier" boil:"collection_items.tier"`
	AvatarURL string `json:"avatar_url" boil:"mech_skin.avatar_url"`
}

type MarketplaceKeycardSaleItem struct {
	*boiler.ItemKeycardSale
	Owner      *boiler.Player         `json:"owner"`
	Blueprints *AssetKeycardBlueprint `json:"blueprints"`
}
