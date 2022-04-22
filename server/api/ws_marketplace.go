package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/battle"
	"server/db"
	"server/gamelog"
	"server/rpcclient"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
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

	return marketplaceHub
}

const HubKeyMarketplaceSalesList hub.HubCommandKey = "MARKETPLACE:SALES:LIST"

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

func (fc *MarketplaceController) SalesListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &MarketplaceSalesListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	return nil
}

const HubKeyMarketplaceSalesCreate hub.HubCommandKey = "MARKETPLACE:SALES:CREATE"

type MarketplaceSalesCreateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SaleType    server.MarketplaceSaleType `json:"sale_type"`
		ItemType    server.MarketplaceItemType `json:"item_type"`
		ItemID      uuid.UUID                  `json:"item_id"`
		AskingPrice *decimal.Decimal           `json:"asking_price"`
	} `json:"payload"`
}

type MarketplaceSalesCreateResponse struct {
	ID       string `json:"id"`
	ItemType string `json:"item_type"`
	SaleType string `json:"sale_type"`
}

func (fc *MarketplaceController) SalesCreateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &MarketplaceSalesCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	userID, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player requesting to sell item")
		return terror.Error(err)
	}

	balance := fc.API.Passport.UserBalanceGet(userID)
	feePrice := db.GetDecimalWithDefault(db.KeyMarketplaceListingFee, decimal.NewFromInt(5))

	if req.Payload.SaleType == server.MarketplaceSaleTypeBuyout {
		feePrice = feePrice.Add(db.GetDecimalWithDefault(db.KeyMarketplaceListingBuyoutFee, decimal.NewFromInt(5)))
	}

	if balance.Sub(feePrice).LessThan(decimal.Zero) {
		err = fmt.Errorf("insufficient funds")
		gamelog.L.Error().
			Str("user_id", hubc.Identifier()).
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
		ToUserID:             battle.SupremacyUserID,
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
			Str("user_id", hubc.Identifier()).
			Str("balance", balance.String()).
			Str("sale_type", string(req.Payload.SaleType)).
			Str("item_type", string(req.Payload.ItemType)).
			Str("item_id", req.Payload.ItemID.String()).
			Err(err).
			Msg("Failed to process transaction for Marketplace Fee.")
		return terror.Error(err, "Failed tp process transaction for Marketplace Fee.")
	}

	// Create Sales Item
	obj, err := db.MarketplaceSaleCreate(req.Payload.SaleType, req.Payload.ItemType, txid, req.Payload.ItemID, req.Payload.AskingPrice)
	if err != nil {
		return terror.Error(err, "Unable to create new sale item.")
	}

	resp := &MarketplaceSalesCreateResponse{
		ID:       obj.ID,
		ItemType: obj.ItemType,
		SaleType: string(req.Payload.SaleType),
	}
	reply(resp)

	return nil
}
