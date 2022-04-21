package api

import (
	"context"
	"encoding/json"
	"server"
	"server/db"

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
		SaleType server.MarketplaceSaleType `json:"sale_type"`
		ItemType server.MarketplaceItemType `json:"item_type"`
		ItemID   uuid.UUID                  `json:"item_id"`
		Price    *decimal.Decimal           `json:"price"`
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

	obj, err := db.MarketplaceSaleCreate(req.Payload.SaleType, req.Payload.ItemType, req.Payload.ItemID, req.Payload.Price)
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
