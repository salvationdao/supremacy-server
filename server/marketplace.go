package server

import (
	"fmt"
	"server/db/boiler"

	"github.com/ninja-software/terror/v2"
)

type MarketplaceSaleType string

const (
	MarketplaceSaleTypeBuyout       MarketplaceSaleType = "BUYOUT"
	MarketplaceSaleTypeAuction      MarketplaceSaleType = "ACTION"
	MarketplaceSaleTypeDutchAuction MarketplaceSaleType = "DUTCH_AUCTION"
)

type MarketplaceSaleItem struct {
	*boiler.ItemSale
	Owner *boiler.Player `json:"owner"`
	Mech  *boiler.Mech   `json:"mech,omitempty"`
}

func (si *MarketplaceSaleItem) GetSaleType() (*MarketplaceSaleType, error) {
	var output MarketplaceSaleType
	if si.Auction {
		output = MarketplaceSaleTypeAuction
	} else if si.DutchAuction {
		output = MarketplaceSaleTypeDutchAuction
	} else if si.Buyout {
		output = MarketplaceSaleTypeBuyout
	} else {
		return nil, terror.Error(fmt.Errorf("unable to identify sale type"))
	}
	return &output, nil
}
