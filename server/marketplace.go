package server

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type MarketplaceSaleItem struct {
	ID                   string                          `json:"id" boil:"id"`
	FactionID            string                          `json:"faction_id" boil:"faction_id"`
	CollectionItemID     string                          `json:"collection_item_id" boil:"collection_item_id"`
	CollectionItemType   string                          `json:"collection_item_type" boil:"collection_item_type"`
	ListingFeeTXID       string                          `json:"listing_fee_tx_id" boil:"listing_fee_tx_id"`
	OwnerID              string                          `json:"owner_id" boil:"owner_id"`
	Auction              bool                            `json:"auction" boil:"auction"`
	AuctionCurrentPrice  decimal.NullDecimal             `json:"auction_current_price,omitempty" boil:"auction_current_price"`
	AuctionReservedPrice decimal.NullDecimal             `json:"auction_reserved_price,omitempty" boil:"auction_reserved_price"`
	TotalBids            int64                           `json:"total_bids" boil:"total_bids"`
	Buyout               bool                            `json:"buyout" boil:"buyout"`
	BuyoutPrice          decimal.NullDecimal             `json:"buyout_price" boil:"buyout_price"`
	DutchAuction         bool                            `json:"dutch_auction" boil:"dutch_auction"`
	DutchAuctionDropRate decimal.NullDecimal             `json:"dutch_auction_drop_rate,omitempty" boil:"dutch_auction_drop_rate"`
	EndAt                time.Time                       `json:"end_at" boil:"end_at"`
	SoldAt               null.Time                       `json:"sold_at" boil:"sold_at"`
	SoldFor              null.String                     `json:"sold_for" boil:"sold_for"`
	SoldBy               decimal.NullDecimal             `json:"sold_by" boil:"sold_by"`
	SoldTXID             null.String                     `json:"sold_tx_id" boil:"sold_tx_id"`
	DeletedAt            null.Time                       `json:"deleted_at" boil:"deleted_at"`
	UpdatedAt            time.Time                       `json:"updated_at" boil:"updated_at"`
	CreatedAt            time.Time                       `json:"created_at" boil:"created_at"`
	Owner                MarketplaceSaleItemOwner        `json:"owner,omitempty" boil:",bind"`
	Mech                 MarketplaceSaleItemMech         `json:"mech,omitempty" boil:",bind"`
	MysteryCrate         MarketplaceSaleItemMysteryCrate `json:"mystery_crate,omitempty" boil:",bind"`
	LastBid              MarketplaceBidder               `json:"last_bid,omitempty" boil:",bind"`
}

type MarketplaceBidder struct {
	ID            null.String `json:"id" boil:"bidder.id"`
	Username      null.String `json:"username" boil:"bidder.username"`
	FactionID     null.String `json:"faction_id" boil:"bidder.faction_id"`
	PublicAddress null.String `json:"public_address" boil:"bidder.public_address"`
	Gid           null.Int    `json:"gid" boil:"bidder.gid"`
}

func (b MarketplaceBidder) MarshalJSON() ([]byte, error) {
	if !b.ID.Valid && !b.Username.Valid && !b.FactionID.Valid && !b.PublicAddress.Valid && !b.Gid.Valid {
		return null.NullBytes, nil
	}
	type localMarketplaceBidder MarketplaceBidder
	return json.Marshal(localMarketplaceBidder(b))
}

type MarketplaceSaleItemOwner struct {
	ID            string      `json:"id" boil:"players.id"`
	Username      null.String `json:"username" boil:"players.username"`
	FactionID     null.String `json:"faction_id" boil:"players.faction_id"`
	PublicAddress null.String `json:"public_address" boil:"players.public_address"`
	Gid           int         `json:"gid" boil:"players.gid"`
}

type MarketplaceSaleItemMech struct {
	ID        null.String `json:"id" boil:"mechs.id"`
	Label     null.String `json:"label" boil:"mechs.label"`
	Name      null.String `json:"name" boil:"mechs.name"`
	Tier      null.String `json:"tier" boil:"collection_items.tier"`
	AvatarURL null.String `json:"avatar_url" boil:"mech_skin.avatar_url"`
}

func (b MarketplaceSaleItemMech) MarshalJSON() ([]byte, error) {
	if !b.ID.Valid && !b.Label.Valid && !b.Name.Valid && !b.Tier.Valid && !b.AvatarURL.Valid {
		return null.NullBytes, nil
	}
	type localMarketplaceSaleItemMech MarketplaceSaleItemMech
	return json.Marshal(localMarketplaceSaleItemMech(b))
}

type MarketplaceSaleItemMysteryCrate struct {
	ID          null.String `json:"id" boil:"mystery_crate.id"`
	Label       null.String `json:"label" boil:"mystery_crate.label"`
	ImageURL    null.String `json:"image_url" boil:"mystery_crate.image_url"`
	LockedUntil null.Time   `json:"locked_until" boil:"mystery_crate.locked_until"`
}

func (b MarketplaceSaleItemMysteryCrate) MarshalJSON() ([]byte, error) {
	if !b.ID.Valid && !b.Label.Valid && !b.ImageURL && !b.LockedUntil.Valid {
		return null.NullBytes, nil
	}
	type localMarketplaceSaleItemMysteryCrate MarketplaceSaleItemMysteryCrate
	return json.Marshal(localMarketplaceSaleItemMysteryCrate(b))
}

type MarketplaceSaleItem1155 struct {
	ID             string                   `json:"id" boil:"id"`
	FactionID      string                   `json:"faction_id" boil:"faction_id"`
	ItemID         string                   `json:"item_id" boil:"item_id"`
	ListingFeeTXID string                   `json:"listing_fee_tx_id" boil:"listing_fee_tx_id"`
	OwnerID        string                   `json:"owner_id" boil:"owner_id"`
	BuyoutPrice    decimal.Decimal          `json:"buyout_price" boil:"buyout_price"`
	EndAt          time.Time                `json:"end_at" boil:"end_at"`
	SoldAt         null.Time                `json:"sold_at" boil:"sold_at"`
	SoldFor        decimal.NullDecimal      `json:"sold_for" boil:"sold_for"`
	SoldBy         null.String              `json:"sold_by" boil:"sold_by"`
	SoldTXID       null.String              `json:"sold_tx_id" boil:"sold_tx_id"`
	DeletedAt      null.Time                `json:"deleted_at" boil:"deleted_at"`
	UpdatedAt      time.Time                `json:"updated_at" boil:"updated_at"`
	CreatedAt      time.Time                `json:"created_at" boil:"created_at"`
	Owner          MarketplaceSaleItemOwner `json:"owner,omitempty" boil:",bind"`
	Keycard        AssetKeycardBlueprint    `json:"keycard,omitempty" boil:",bind"`
}
