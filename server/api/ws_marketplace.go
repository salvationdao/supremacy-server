package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamelog"
	"server/rpcclient"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
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

	api.SecureUserCommand(HubKeyMarketplaceSalesList, marketplaceHub.SalesListHandler)
	api.SecureUserCommand(HubKeyMarketplaceSalesCreate, marketplaceHub.SalesCreateHandler)

	// api.SecureUserSubscribeCommand(HubKeyMarketplaceSalesItemUpdate, marketplaceHub.SalesItemUpdateSubscriber)

	return marketplaceHub
}

const HubKeyMarketplaceSalesList = "MARKETPLACE:SALES:LIST"

type MarketplaceSalesListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID   server.UserID         `json:"user_id"`
		SortDir  db.SortByDir          `json:"sort_dir"`
		SortBy   string                `json:"sort_by"`
		Filter   *db.ListFilterRequest `json:"filter,omitempty"`
		Archived bool                  `json:"archived"`
		Search   string                `json:"search"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

type MarketplaceSalesListResponse struct {
	Total   int64                     `json:"total"`
	Records []*db.MarketplaceSaleItem `json:"records"`
}

func (fc *MarketplaceController) SalesListHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &MarketplaceSalesListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, records, err := db.MarketplaceItemSaleList(req.Payload.Search, req.Payload.Archived, req.Payload.Filter, offset, req.Payload.PageSize, req.Payload.SortBy, req.Payload.SortDir)
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

const HubKeyMarketplaceSalesCreate = "MARKETPLACE:SALES:CREATE"

type MarketplaceSalesCreateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SaleType             server.MarketplaceSaleType `json:"sale_type"`
		ItemType             server.MarketplaceItemType `json:"item_type"`
		ItemID               uuid.UUID                  `json:"item_id"`
		AskingPrice          *decimal.Decimal           `json:"asking_price"`
		DutchAuctionDropRate *decimal.Decimal           `json:"dutch_auction_drop_rate"`
		ListingDurationHours int64                      `json:"listing_duration_hours"`
	} `json:"payload"`
}

func (fc *MarketplaceController) SalesCreateHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue processing list sale item, try again or contact support."
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

	if user.FactionID.IsZero() {
		err := fmt.Errorf("player is not enlisted in a faction")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Err(err).
			Msg("Player is not in a faction")
		return terror.Error(err, "You are not enlisted in a faction.")
	}
	factionID, err := uuid.FromString(user.FactionID.String)
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

	balance := fc.API.Passport.UserBalanceGet(userID)
	feePrice := db.GetDecimalWithDefault(db.KeyMarketplaceListingFee, decimal.NewFromInt(5))

	if req.Payload.SaleType == server.MarketplaceSaleTypeBuyout {
		feePrice = feePrice.Add(db.GetDecimalWithDefault(db.KeyMarketplaceListingBuyoutFee, decimal.NewFromInt(5)))
	}
	feePrice = feePrice.Mul(decimal.NewFromInt(req.Payload.ListingDurationHours))

	if balance.Sub(feePrice).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("sale_type", string(req.Payload.SaleType)).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Player does not have enough sups.")
		return terror.Error(err, "You do not have enough sups.")
	}

	// pay sup
	txid, err := fc.API.Passport.SpendSupMessage(rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               feePrice.String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_fee|%s|%s|%s|%d", req.Payload.SaleType, req.Payload.ItemType, req.Payload.ItemID.String(), time.Now().UnixNano())),
		Group:                string(server.TransactionGroupMarketplace),
		SubGroup:             "SUPREMACY",
		Description:          fmt.Sprintf("marketplace fee: %s - %s: %s", req.Payload.SaleType, req.Payload.ItemType, req.Payload.ItemID.String()),
		NotSafe:              true,
	})
	if err != nil {
		err = fmt.Errorf("failed to process marketplace fee transaction")
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("balance", balance.String()).
			Str("sale_type", string(req.Payload.SaleType)).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to process transaction for Marketplace Fee.")
		return terror.Error(err, "Failed tp process transaction for Marketplace Fee.")
	}

	// Create Sales Item
	endAt := time.Now().Add(time.Hour * time.Duration(req.Payload.ListingDurationHours))
	obj, err := db.MarketplaceSaleCreate(req.Payload.SaleType, userID, factionID, txid, endAt, req.Payload.ItemType, req.Payload.ItemID, req.Payload.AskingPrice, req.Payload.DutchAuctionDropRate)
	if err != nil {
		fc.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("sale_type", string(req.Payload.SaleType)).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item.")
		return terror.Error(err, "Unable to create new sale item.")
	}

	obj, err = db.MarketplaceLoadItemSaleObject(obj)
	if err != nil {
		fc.API.Passport.RefundSupsMessage(txid)
		gamelog.L.Error().
			Str("user_id", user.ID).
			Str("sale_type", string(req.Payload.SaleType)).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Unable to create new sale item (post create).")
		return terror.Error(err, "Unable to create new sale item.")
	}

	reply(obj)

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
		return "", "", terror.Error(err, "Invalid request received")
	}

	obj, err := db.MarketplaceItemSale(req.Payload.ID)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	reply(obj)

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyMarketplaceSalesItemUpdate, req.Payload.ID))
	return req.TransactionID, busKey, nil
}
