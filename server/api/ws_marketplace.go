package api

import (
	"context"

	"github.com/ninja-syndicate/hub"
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

	return marketplaceHub
}

const HubKeyMarketplaceSalesList hub.HubCommandKey = "MARKETPLACE:SALES:LIST"

func (fc *MarketplaceController) SalesListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	return nil
}
