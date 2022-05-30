package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// MarketplaceController holds handlers for marketplace
type MarketplaceController struct {
	API *API
}

// NewMarketplaceController creates the marketplace hub
func NewMarketplaceController(api *API) *MarketplaceController {
	marketplaceHub := &MarketplaceController{
		API: api,
	}

	api.SecureUserFactionCommand(HubKeyMarketplaceSalesList, marketplaceHub.SalesListHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardList, marketplaceHub.SalesListKeycardHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesCreate, marketplaceHub.SalesCreateHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardCreate, marketplaceHub.SalesKeycardCreateHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesBuy, marketplaceHub.SalesBuyHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesKeycardBuy, marketplaceHub.SalesKeycardBuyHandler)
	api.SecureUserFactionCommand(HubKeyMarketplaceSalesBid, marketplaceHub.SalesBidHandler)

	// api.SecureUserSubscribeCommand(HubKeyMarketplaceSalesItemUpdate, marketplaceHub.SalesItemUpdateSubscriber)

	return marketplaceHub
}

const HubKeyMarketplaceSalesList = "MARKETPLACE:SALES:LIST"

type MarketplaceSalesListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID         server.UserID         `json:"user_id"`
		SortDir        db.SortByDir          `json:"sort_dir"`
		SortBy         string                `json:"sort_by"`
		Filter         *db.ListFilterRequest `json:"filter,omitempty"`
		FilterRarities []string              `json:"rarities"`
		Search         string                `json:"search"`
		PageSize       int                   `json:"page_size"`
		Page           int                   `json:"page"`
	} `json:"payload"`
}

type MarketplaceSalesListResponse struct {
	Total   int64                         `json:"total"`
	Records []*server.MarketplaceSaleItem `json:"records"`
}

func (fc *MarketplaceController) SalesListHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MarketplaceSalesListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, records, err := db.MarketplaceItemSaleList(
		req.Payload.Search,
		req.Payload.Filter,
		req.Payload.FilterRarities,
		user.ID,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get list of items for sale")
		return terror.Error(err, "Failed to get list of items for sale")
	}

	resp := &MarketplaceSalesListResponse{
		Total:   total,
		Records: records,
	}
	reply(resp)

	return nil
}

const HubKeyMarketplaceSalesKeycardList = "MARKETPLACE:SALES:KEYCARD:LIST"

type MarketplaceSalesKeycardListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID         server.UserID         `json:"user_id"`
		SortDir        db.SortByDir          `json:"sort_dir"`
		SortBy         string                `json:"sort_by"`
		Filter         *db.ListFilterRequest `json:"filter,omitempty"`
		FilterRarities []string              `json:"rarities"`
		Search         string                `json:"search"`
		PageSize       int                   `json:"page_size"`
		Page           int                   `json:"page"`
	} `json:"payload"`
}

type MarketplaceSalesListKeycardResponse struct {
	Total   int64                                `json:"total"`
	Records []*server.MarketplaceKeycardSaleItem `json:"records"`
}

func (fc *MarketplaceController) SalesListKeycardHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MarketplaceSalesKeycardListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, records, err := db.MarketplaceItemKeycardSaleList(
		req.Payload.Search,
		req.Payload.Filter,
		user.ID,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get list of items for sale")
		return terror.Error(err, "Failed to get list of items for sale")
	}

	resp := &MarketplaceSalesListKeycardResponse{
		Total:   total,
		Records: records,
	}
	reply(resp)

	return nil
}

const HubKeyMarketplaceSalesCreate = "MARKETPLACE:SALES:CREATE"

type MarketplaceSalesCreateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ItemType             string           `json:"item_type"`
		ItemID               uuid.UUID        `json:"item_id"`
		HasBuyout            bool             `json:"has_buyout"`
		HasAuction           bool             `json:"has_auction"`
		HasDutchAuction      bool             `json:"has_dutch_auction"`
		AskingPrice          *decimal.Decimal `json:"asking_price"`
		AuctionReservedPrice *decimal.Decimal `json:"auction_reserved_price"`
		DutchAuctionDropRate *decimal.Decimal `json:"dutch_auction_drop_rate"`
		ListingDurationHours int64            `json:"listing_duration_hours"`
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesCreateHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue processing create sale item, try again or contact support."
	req := &MarketplaceSalesCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.ListingDurationHours <= 0 {
		err = fmt.Errorf("listing duration hours required")
		return terror.Error(err, "Invalid request received")
	}

	userID, err := uuid.FromString(user.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player requesting to sell item")
		return terror.Error(err, errMsg)
	}

	factionID, err := uuid.FromString(fID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Err(err).
			Msg("Player is not in a faction")
		return terror.Error(err, errMsg)
	}

	factionAccountID, ok := server.FactionUsers[user.FactionID.String]
	if !ok {
		err = fmt.Errorf("failed to get hard coded syndicate player id")
		gamelog.L.Error().
			Str("player_id", user.ID).
			Str("faction_id", user.FactionID.String).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
		return terror.Error(err, errMsg)
	}

	balance := mp.API.Passport.UserBalanceGet(userID)
	feePrice := db.GetDecimalWithDefault(db.KeyMarketplaceListingFee, decimal.NewFromInt(5))
	if req.Payload.HasBuyout {
		feePrice = feePrice.Add(db.GetDecimalWithDefault(db.KeyMarketplaceListingBuyoutFee, decimal.NewFromInt(5)))
	}
	feePrice = feePrice.Mul(decimal.NewFromInt(req.Payload.ListingDurationHours)).Mul(decimal.New(1, 18))

	if balance.Sub(feePrice).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Player does not have enough sups.")
		return terror.Error(err, "You do not have enough sups.")
	}

	// pay sup
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               feePrice.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_fee|%s|%s|%d", req.Payload.ItemType, req.Payload.ItemID.String(), time.Now().UnixNano())),
		Group:                string(server.TransactionGroupMarketplace),
		SubGroup:             "SUPREMACY",
		Description:          fmt.Sprintf("marketplace fee: %s: %s", req.Payload.ItemType, req.Payload.ItemID.String()),
		NotSafe:              true,
	})
	if err != nil {
		err = fmt.Errorf("failed to process marketplace fee transaction")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to process transaction for Marketplace Fee.")
		return terror.Error(err, "Failed tp process transaction for Marketplace Fee.")
	}

	// Create Sales Item
	endAt := time.Now().Add(time.Hour * time.Duration(req.Payload.ListingDurationHours))
	obj, err := db.MarketplaceSaleCreate(
		userID,
		factionID,
		txid,
		endAt,
		req.Payload.ItemID,
		req.Payload.HasBuyout,
		req.Payload.AskingPrice,
		req.Payload.HasAuction,
		req.Payload.AuctionReservedPrice,
		req.Payload.HasDutchAuction,
		req.Payload.DutchAuctionDropRate,
	)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, "Unable to create new sale item.")
	}

	obj, err = db.MarketplaceLoadItemSaleObject(obj)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item (post create).")
		return terror.Error(err, "Unable to create new sale item.")
	}

	reply(obj)

	return nil
}

const HubKeyMarketplaceSalesKeycardCreate = "MARKETPLACE:SALES:KEYCARD:CREATE"

type HubKeyMarketplaceSalesKeycardCreateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ItemType             string           `json:"item_type"`
		ItemID               uuid.UUID        `json:"item_id"`
		AskingPrice          *decimal.Decimal `json:"asking_price"`
		ListingDurationHours int64            `json:"listing_duration_hours"`
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesKeycardCreateHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue processing create keycard sale item, try again or contact support."
	req := &HubKeyMarketplaceSalesKeycardCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.ListingDurationHours <= 0 {
		err = fmt.Errorf("listing duration hours required")
		return terror.Error(err, "Invalid request received")
	}

	userID, err := uuid.FromString(user.ID)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player requesting to sell item")
		return terror.Error(err, errMsg)
	}

	factionID, err := uuid.FromString(fID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Err(err).
			Msg("Player is not in a faction")
		return terror.Error(err, errMsg)
	}

	factionAccountID, ok := server.FactionUsers[user.FactionID.String]
	if !ok {
		err = fmt.Errorf("failed to get hard coded syndicate player id")
		gamelog.L.Error().
			Str("player_id", user.ID).
			Str("faction_id", user.FactionID.String).
			Err(err).
			Msg("unable to get hard coded syndicate player ID from faction ID")
		return terror.Error(err, errMsg)
	}

	balance := mp.API.Passport.UserBalanceGet(userID)
	feePrice := db.GetDecimalWithDefault(db.KeyMarketplaceListingFee, decimal.NewFromInt(5))
	feePrice = feePrice.Add(db.GetDecimalWithDefault(db.KeyMarketplaceListingBuyoutFee, decimal.NewFromInt(5)))
	feePrice = feePrice.Mul(decimal.NewFromInt(req.Payload.ListingDurationHours)).Mul(decimal.New(1, 18))

	if balance.Sub(feePrice).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Player does not have enough sups.")
		return terror.Error(err, "You do not have enough sups.")
	}

	// pay sup
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               feePrice.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_fee|%s|%s|%d", req.Payload.ItemType, req.Payload.ItemID.String(), time.Now().UnixNano())),
		Group:                string(server.TransactionGroupMarketplace),
		SubGroup:             "SUPREMACY",
		Description:          fmt.Sprintf("marketplace fee: %s: %s", req.Payload.ItemType, req.Payload.ItemID.String()),
		NotSafe:              true,
	})
	if err != nil {
		err = fmt.Errorf("failed to process marketplace fee transaction")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to process transaction for Marketplace Fee.")
		return terror.Error(err, "Failed tp process transaction for Marketplace Fee.")
	}

	// Create Sales Item
	endAt := time.Now().Add(time.Hour * time.Duration(req.Payload.ListingDurationHours))
	obj, err := db.MarketplaceKeycardSaleCreate(
		userID,
		factionID,
		txid,
		endAt,
		req.Payload.ItemID,
		req.Payload.AskingPrice,
	)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, "Unable to create new sale item.")
	}

	// obj, err = db.MarketplaceLoadItemSaleObject(obj)
	// if err != nil {
	// 	mp.API.Passport.RefundSupsMessage(txid)
	// 	gamelog.L.Error().
	// 		Str("user_id", user.ID).
	// 		Str("item_type", string(req.Payload.ItemType)).
	// 		Str("item_id", req.Payload.ItemID.String()).
	// 		Err(err).
	// 		Msg("Unable to create new sale item (post create).")
	// 	return terror.Error(err, "Unable to create new sale item.")
	// }

	reply(obj)

	return nil
}

const (
	HubKeyMarketplaceSalesBuy        = "MARKETPLACE:SALES:BUY"
	HubKeyMarketplaceSalesKeycardBuy = "MARKETPLACE:SALES:KEYCARD:BUY"
)

type MarketplaceSalesBuyRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ItemID uuid.UUID `json:"item_id"`
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesBuyHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue buying sale item, try again or contact support."
	req := &MarketplaceSalesBuyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Check whether user can buy sale item
	saleItem, err := db.MarketplaceItemSale(req.Payload.ItemID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}
	if saleItem.FactionID != fID {
		return terror.Error(terror.ErrUnauthorised, "Item does not belong to user's faction.")
	}

	// Pay item
	// TODO: Work out Sales Cut
	userID, err := uuid.FromString(user.ID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to retrieve buyer's user id.")
		return terror.Error(err, errMsg)
	}

	saleItemCost, err := decimal.NewFromString(saleItem.BuyoutPrice.String)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to get current buyout price.")
		return terror.Error(err, errMsg)
	}

	balance := mp.API.Passport.UserBalanceGet(userID)
	if balance.Sub(saleItemCost).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Err(err).
			Msg("Player does not have enough sups.")
		return terror.Error(err, "You do not have enough sups.")
	}

	// pay sup
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(saleItem.OwnerID)),
		Amount:               saleItemCost.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item|%s|%d", saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupMarketplace),
		SubGroup:             "SUPREMACY",
		Description:          fmt.Sprintf("marketplace buy item: %s", saleItem.ID),
		NotSafe:              true,
	})
	if err != nil {
		err = fmt.Errorf("failed to process payment transaction")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to process transaction for Purchase Sale Item.")
		return terror.Error(err, "Failed tp process transaction for Purchase Sale Item.")
	}

	// update sale item
	saleItemRecord := saleItem.ItemSale
	saleItemRecord.SoldAt = null.TimeFrom(time.Now())
	saleItemRecord.SoldFor = null.StringFrom(saleItemCost.String())
	saleItemRecord.SoldTXID = null.StringFrom(txid)
	saleItemRecord.SoldBy = null.StringFrom(user.ID)

	_, err = saleItemRecord.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		err = fmt.Errorf("failed to complete payment transaction")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to process transaction for Purchase Sale Item.")
		return terror.Error(err, "Failed tp process transaction for Purchase Sale Item.")
	}

	// transfer ownership of asset
	err = db.ChangeMechOwner(req.Payload.ItemID)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to Transfer Mech to New Owner")
		return terror.Error(err, "Failed to process transaction for Purchase Sale Item.")
	}

	// success
	reply(true)

	return nil
}

func (mp *MarketplaceController) SalesKeycardBuyHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue buying sale item, try again or contact support."
	req := &MarketplaceSalesBuyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Check whether user can buy sale item
	saleItem, err := db.MarketplaceItemSale(req.Payload.ItemID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}

	// Pay item
	// TODO: Work out Sales Cut
	userID, err := uuid.FromString(user.ID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to retrieve buyer's user id.")
		return terror.Error(err, errMsg)
	}

	saleItemCost, err := decimal.NewFromString(saleItem.BuyoutPrice.String)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to get current buyout price.")
		return terror.Error(err, errMsg)
	}

	balance := mp.API.Passport.UserBalanceGet(userID)
	if balance.Sub(saleItemCost).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Err(err).
			Msg("Player does not have enough sups.")
		return terror.Error(err, "You do not have enough sups.")
	}

	// pay sup
	txid, err := mp.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(saleItem.OwnerID)),
		Amount:               saleItemCost.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item|keycard|%s|%d", saleItem.ID, time.Now().UnixNano())),
		Group:                string(server.TransactionGroupMarketplace),
		SubGroup:             "SUPREMACY",
		Description:          fmt.Sprintf("marketplace buy item: %s", saleItem.ID),
		NotSafe:              true,
	})
	if err != nil {
		err = fmt.Errorf("failed to process payment transaction")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to process transaction for Purchase Sale Item.")
		return terror.Error(err, "Failed tp process transaction for Purchase Sale Item.")
	}

	// update sale item
	saleItemRecord := saleItem.ItemSale
	saleItemRecord.SoldAt = null.TimeFrom(time.Now())
	saleItemRecord.SoldFor = null.StringFrom(saleItemCost.String())
	saleItemRecord.SoldTXID = null.StringFrom(txid)
	saleItemRecord.SoldBy = null.StringFrom(user.ID)

	_, err = saleItemRecord.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		err = fmt.Errorf("failed to complete payment transaction")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to process transaction for Purchase Sale Item.")
		return terror.Error(err, "Failed tp process transaction for Purchase Sale Item.")
	}

	// transfer ownership of asset
	err = db.ChangeMechOwner(req.Payload.ItemID)
	if err != nil {
		mp.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("cost", saleItemCost.String()).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to Transfer Mech to New Owner")
		return terror.Error(err, "Failed to process transaction for Purchase Sale Item.")
	}

	// success
	reply(true)

	return nil
}

const HubKeyMarketplaceSalesBid = "MARKETPLACE:SALES:BID"

type MarketplaceSalesBidRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ItemID uuid.UUID `json:"item_id"`
		Amount string    `json:"amount"`
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesBidHandler(ctx context.Context, user *boiler.Player, fID string, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue placing bid, try again or contact support."
	req := &MarketplaceSalesBidRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	userID, err := uuid.FromString(user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Check whether user can buy sale item
	saleItem, err := db.MarketplaceItemSale(req.Payload.ItemID)
	if errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Item not found.")
	}
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}
	if !saleItem.Auction {
		return terror.Error(terror.ErrInvalidInput, "Item is not up for auction.")
	}
	if saleItem.FactionID != fID {
		return terror.Error(terror.ErrInvalidInput, "Item does not belong to user's faction.")
	}

	currentAmount, err := decimal.NewFromString(saleItem.AuctionCurrentPrice.String)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Str("current_auction_price", saleItem.AuctionCurrentPrice.String).
			Err(err).
			Msg("Unable to retrieve sale item.")
		return terror.Error(err, errMsg)
	}
	bidAmount, err := decimal.NewFromString(req.Payload.Amount)
	if err != nil {
		return terror.Error(err, "Invalid Bid Amount received.")
	}
	if currentAmount.GreaterThanOrEqual(bidAmount) {
		return terror.Error(terror.ErrInvalidInput, "Invalid bid amount, must be above the current bid price.")
	}

	// Place Bid
	_, err = db.MarketplaceSaleBidHistoryCreate(req.Payload.ItemID, userID, bidAmount)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Str("current_auction_price", saleItem.AuctionCurrentPrice.String).
			Str("bid_amount", bidAmount.String()).
			Err(err).
			Msg("Unable to place bid.")
		return terror.Error(err, errMsg)
	}

	err = db.MarketplaceSaleAuctionSync(req.Payload.ItemID)
	if err != nil {
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("item_id", req.Payload.ItemID.String()).
			Str("bid_amount", bidAmount.String()).
			Err(err).
			Msg("Unable to update current auction price.")
		return terror.Error(err, errMsg)
	}

	reply(true)

	return nil
}

const HubKeyMarketplaceSalesItemUpdate = "MARKETPLACE:SALES:ITEM:UPDATE"

type MarketplaceSalesItemUpdateSubscribe struct {
	*hub.HubCommandRequest
	Payload struct {
		ID uuid.UUID `json:"id"` // item id
	} `json:"payload"`
}

func (mp *MarketplaceController) SalesItemUpdateSubscriber(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &MarketplaceSalesItemUpdateSubscribe{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received.")
	}

	obj, err := db.MarketplaceItemSale(req.Payload.ID)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received.")
	}

	reply(obj)

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyMarketplaceSalesItemUpdate, req.Payload.ID))
	return req.TransactionID, busKey, nil
}
