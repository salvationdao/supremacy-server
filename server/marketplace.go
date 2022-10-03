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
	ListingFeeTXID       null.String                     `json:"listing_fee_tx_id" boil:"listing_fee_tx_id"`
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
	SoldFor              decimal.NullDecimal             `json:"sold_for" boil:"sold_for"`
	SoldTo               MarketplaceUser                 `json:"sold_to" boil:"sold_to,bind"`
	SoldTXID             null.String                     `json:"sold_tx_id" boil:"sold_tx_id"`
	SoldFeeTXID          null.String                     `json:"sold_fee_tx_id" boil:"sold_fee_tx_id"`
	DeletedAt            null.Time                       `json:"deleted_at" boil:"deleted_at"`
	UpdatedAt            time.Time                       `json:"updated_at" boil:"updated_at"`
	CreatedAt            time.Time                       `json:"created_at" boil:"created_at"`
	Owner                MarketplaceUser                 `json:"owner,omitempty" boil:"players,bind"`
	Mech                 MarketplaceSaleItemMech         `json:"mech,omitempty" boil:",bind"`
	MysteryCrate         MarketplaceSaleItemMysteryCrate `json:"mystery_crate,omitempty" boil:",bind"`
	Weapon               MarketplaceSaleItemWeapon       `json:"weapon,omitempty" boil:",bind"`
	CollectionItem       MarketplaceSaleCollectionItem   `json:"collection_item,omitempty" boil:",bind"`
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

type MarketplaceUser struct {
	ID            null.String `json:"id" boil:"id"`
	Username      null.String `json:"username" boil:"username"`
	FactionID     null.String `json:"faction_id" boil:"faction_id"`
	PublicAddress null.String `json:"public_address" boil:"public_address"`
	Gid           null.Int    `json:"gid" boil:"gid"`
}

func (u MarketplaceUser) MarshalJSON() ([]byte, error) {
	if !u.ID.Valid {
		return null.NullBytes, nil
	}
	type localMarketplaceUser MarketplaceUser
	return json.Marshal(localMarketplaceUser(u))
}

type MarketplaceSaleCollectionItem struct {
	Hash         string      `json:"hash" boil:"collection_items.hash"`
	Tier         null.String `json:"tier,omitempty" boil:"collection_items.tier"`
	XsynLocked   bool        `json:"-" boil:"collection_items.xsyn_locked"`
	MarketLocked bool        `json:"-" boil:"collection_items.market_locked"`
}

func (b MarketplaceSaleCollectionItem) MarshalJSON() ([]byte, error) {
	if !b.Tier.Valid {
		return null.NullBytes, nil
	}
	type localMarketplaceSaleCollectionItem MarketplaceSaleCollectionItem
	return json.Marshal(localMarketplaceSaleCollectionItem(b))
}

type MarketplaceSaleItemMech struct {
	ID        null.String `json:"id" boil:"mechs.id"`
	Label     null.String `json:"label" boil:"mechs.label"`
	Name      null.String `json:"name" boil:"mechs.name"`
	AvatarURL null.String `json:"avatar_url" boil:"mech_skin.avatar_url"`
}

func (b MarketplaceSaleItemMech) MarshalJSON() ([]byte, error) {
	if !b.ID.Valid && !b.Label.Valid && !b.Name.Valid && !b.AvatarURL.Valid {
		return null.NullBytes, nil
	}
	type localMarketplaceSaleItemMech MarketplaceSaleItemMech
	return json.Marshal(localMarketplaceSaleItemMech(b))
}

type MarketplaceSaleItemMysteryCrate struct {
	ID               null.String `json:"id" boil:"mystery_crate.id"`
	Label            null.String `json:"label" boil:"mystery_crate.label"`
	Description      null.String `json:"description" boil:"mystery_crate.description"`
	ImageURL         null.String `json:"image_url,omitempty" boil:"storefront_mystery_crates.image_url"`
	CardAnimationURL null.String `json:"card_animation_url,omitempty" boil:"storefront_mystery_crates.card_animation_url"`
	LargeImageURL    null.String `json:"large_image_url,omitempty" boil:"storefront_mystery_crates.large_image_url"`
	AnimationURL     null.String `json:"animation_url,omitempty" boil:"storefront_mystery_crates.animation_url"`
}

func (b MarketplaceSaleItemMysteryCrate) MarshalJSON() ([]byte, error) {
	if !b.ID.Valid && !b.Label.Valid {
		return null.NullBytes, nil
	}
	type localMarketplaceSaleItemMysteryCrate MarketplaceSaleItemMysteryCrate
	return json.Marshal(localMarketplaceSaleItemMysteryCrate(b))
}

type MarketplaceSaleItemWeapon struct {
	ID         null.String `json:"id" boil:"weapons.id"`
	Label      null.String `json:"label" boil:"weapons.label"`
	WeaponType null.String `json:"weapon_type" boil:"weapons.weapon_type"`
	AvatarURL  null.String `json:"avatar_url" boil:"weapons.avatar_url"`
}

func (b MarketplaceSaleItemWeapon) MarshalJSON() ([]byte, error) {
	if !b.ID.Valid {
		return null.NullBytes, nil
	}
	type localMarketplaceSaleItemWeapon MarketplaceSaleItemWeapon
	return json.Marshal(localMarketplaceSaleItemWeapon(b))
}

type MarketplaceSaleItem1155 struct {
	ID             string                `json:"id" boil:"id"`
	FactionID      string                `json:"faction_id" boil:"faction_id"`
	ItemID         string                `json:"item_id" boil:"item_id"`
	ListingFeeTXID null.String           `json:"listing_fee_tx_id" boil:"listing_fee_tx_id"`
	OwnerID        string                `json:"owner_id" boil:"owner_id"`
	BuyoutPrice    decimal.Decimal       `json:"buyout_price" boil:"buyout_price"`
	EndAt          time.Time             `json:"end_at" boil:"end_at"`
	SoldAt         null.Time             `json:"sold_at" boil:"sold_at"`
	SoldFor        decimal.NullDecimal   `json:"sold_for" boil:"sold_for"`
	SoldTXID       null.String           `json:"sold_tx_id" boil:"sold_tx_id"`
	SoldFeeTXID    null.String           `json:"sold_fee_tx_id" boil:"sold_fee_tx_id"`
	DeletedAt      null.Time             `json:"deleted_at" boil:"deleted_at"`
	UpdatedAt      time.Time             `json:"updated_at" boil:"updated_at"`
	CreatedAt      time.Time             `json:"created_at" boil:"created_at"`
	Owner          MarketplaceUser       `json:"owner,omitempty" boil:"players,bind"`
	SoldTo         MarketplaceUser       `json:"sold_to" boil:"sold_to,bind"`
	Keycard        AssetKeycardBlueprint `json:"keycard,omitempty" boil:",bind"`
}

type MechArenaStatus string

const (
	MechArenaStatusQueue   MechArenaStatus = "QUEUE"
	MechArenaStatusBattle  MechArenaStatus = "BATTLE"
	MechArenaStatusMarket  MechArenaStatus = "MARKET"
	MechArenaStatusIdle    MechArenaStatus = "IDLE"
	MechArenaStatusSold    MechArenaStatus = "SOLD"
	MechArenaStatusDamaged MechArenaStatus = "DAMAGED"
	MechArenaStatusStaked  MechArenaStatus = "STAKED"
)

type MechArenaInfo struct {
	Status              MechArenaStatus `json:"status"` // "QUEUE" | "BATTLE" | "MARKET" | "IDLE" | "SOLD"
	CanDeploy           bool            `json:"can_deploy"`
	BattleLobbyIsLocked bool            `json:"battle_lobby_is_locked"`
}

type MarketplaceEvent struct {
	ID        string                `json:"id" boil:"id"`
	EventType string                `json:"event_type" boil:"event_type"`
	Amount    decimal.NullDecimal   `json:"amount" boil:"amount"`
	CreatedAt time.Time             `json:"created_at" boil:"created_at"`
	Item      *MarketplaceEventItem `json:"item"`
}

type MarketplaceEventItem struct {
	ID                   string                          `json:"id" boil:"id"`
	FactionID            string                          `json:"faction_id" boil:"faction_id"`
	CollectionItemID     string                          `json:"collection_item_id" boil:"collection_item_id"`
	CollectionItemType   string                          `json:"collection_item_type" boil:"collection_item_type"`
	ListingFeeTXID       null.String                     `json:"listing_fee_tx_id" boil:"listing_fee_tx_id"`
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
	SoldFor              decimal.NullDecimal             `json:"sold_for" boil:"sold_for"`
	SoldTo               MarketplaceUser                 `json:"sold_to" boil:"sold_to,bind"`
	SoldTXID             null.String                     `json:"sold_tx_id" boil:"sold_tx_id"`
	SoldFeeTXID          null.String                     `json:"sold_fee_tx_id" boil:"sold_fee_tx_id"`
	DeletedAt            null.Time                       `json:"deleted_at" boil:"deleted_at"`
	UpdatedAt            time.Time                       `json:"updated_at" boil:"updated_at"`
	CreatedAt            time.Time                       `json:"created_at" boil:"created_at"`
	Owner                MarketplaceUser                 `json:"owner,omitempty" boil:"players,bind"`
	Mech                 MarketplaceSaleItemMech         `json:"mech,omitempty" boil:",bind"`
	MysteryCrate         MarketplaceSaleItemMysteryCrate `json:"mystery_crate,omitempty" boil:",bind"`
	Weapon               MarketplaceSaleItemWeapon       `json:"weapon,omitempty" boil:",bind"`
	Keycard              AssetKeycardBlueprint           `json:"keycard,omitempty" boil:",bind"`
	CollectionItem       MarketplaceSaleCollectionItem   `json:"collection_item,omitempty" boil:",bind"`
	LastBid              MarketplaceBidder               `json:"last_bid,omitempty" boil:",bind"`
}
