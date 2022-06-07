package server

import (
	"time"

	"github.com/volatiletech/null/v8"
)

type MarketplaceSaleItem struct {
	ID                   string                   `json:"id" boil:"id"`
	FactionID            string                   `json:"faction_id" boil:"faction_id"`
	ItemID               string                   `json:"item_id" boil:"item_id"`
	ListingFeeTXID       string                   `json:"listing_fee_tx_id" boil:"listing_fee_tx_id"`
	OwnerID              string                   `json:"owner_id" boil:"owner_id"`
	Auction              bool                     `json:"auction" boil:"auction"`
	AuctionCurrentPrice  null.String              `json:"auction_current_price,omitempty" boil:"auction_current_price"`
	AuctionReservedPrice null.String              `json:"auction_reserved_price,omitempty" boil:"auction_reserved_price"`
	Buyout               bool                     `json:"buyout" boil:"buyout"`
	BuyoutPrice          null.String              `json:"buyout_price" boil:"buyout_price"`
	DutchAuction         bool                     `json:"dutch_auction" boil:"dutch_auction"`
	DutchAuctionDropRate null.String              `json:"dutch_auction_drop_rate,omitempty" boil:"dutch_auction_drop_rate"`
	EndAt                time.Time                `json:"end_at" boil:"end_at"`
	SoldAt               null.Time                `json:"sold_at" boil:"sold_at"`
	SoldFor              null.String              `json:"sold_for" boil:"sold_for"`
	SoldBy               null.String              `json:"sold_by" boil:"sold_by"`
	SoldTXID             null.String              `json:"sold_tx_id" boil:"sold_tx_id"`
	DeletedAt            null.Time                `json:"deleted_at" boil:"deleted_at"`
	UpdatedAt            time.Time                `json:"updated_at" boil:"updated_at"`
	CreatedAt            time.Time                `json:"created_at" boil:"created_at"`
	Owner                MarketplaceSaleItemOwner `json:"owner,omitempty" boil:",bind"`
	Mech                 MarketplaceSaleItemMech  `json:"mech,omitempty" boil:",bind"`
}

type MarketplaceSaleItemOwner struct {
	ID            string      `json:"id" boil:"players.id"`
	Username      null.String `json:"username" boil:"players.username"`
	FactionID      null.String `json:"faction_id" boil:"players.faction_id"`
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
	ID             string                   `json:"id" boil:"id"`
	FactionID      string                   `json:"faction_id" boil:"faction_id"`
	ItemID         string                   `json:"item_id" boil:"item_id"`
	ListingFeeTXID string                   `json:"listing_fee_tx_id" boil:"listing_fee_tx_id"`
	OwnerID        string                   `json:"owner_id" boil:"owner_id"`
	BuyoutPrice    string                   `json:"buyout_price" boil:"buyout_price"`
	EndAt          time.Time                `json:"end_at" boil:"end_at"`
	SoldAt         null.Time                `json:"sold_at" boil:"sold_at"`
	SoldFor        null.String              `json:"sold_for" boil:"sold_for"`
	SoldBy         null.String              `json:"sold_by" boil:"sold_by"`
	SoldTXID       null.String              `json:"sold_tx_id" boil:"sold_tx_id"`
	DeletedAt      null.Time                `json:"deleted_at" boil:"deleted_at"`
	UpdatedAt      time.Time                `json:"updated_at" boil:"updated_at"`
	CreatedAt      time.Time                `json:"created_at" boil:"created_at"`
	Owner          MarketplaceSaleItemOwner `json:"owner,omitempty" boil:",bind"`
	Blueprints     AssetKeycardBlueprint    `json:"blueprints,omitempty" boil:",bind"`
}
