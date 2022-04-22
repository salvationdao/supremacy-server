package db

import (
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// MarketplaceSaleList returns a numeric paginated result of sales list.
func MarketplaceSaleList(search string, archived bool, filter *ListFilterRequest, offset int, pageSize int, sortBy string, sortDir SortByDir) (int, []*boiler.ItemSale, error) {
	return 0, nil, nil
}

// MarketplaceSaleCreate inserts a new sale item.
func MarketplaceSaleCreate(saleType server.MarketplaceSaleType, ownerID uuid.UUID, factionID uuid.UUID, listFeeTxnID string, itemType server.MarketplaceItemType, itemID uuid.UUID, askingPrice *decimal.Decimal, dutchOptionDropRate *decimal.Decimal) (*boiler.ItemSale, error) {
	obj := &boiler.ItemSale{
		OwnerID:        ownerID.String(),
		FactionID:      factionID.String(),
		ListingFeeTXID: listFeeTxnID,
		ItemType:       string(itemType),
		ItemID:         itemID.String(),
	}
	switch saleType {
	case server.MarketplaceSaleTypeBuyout:
		obj.Buyout = true
		obj.BuyoutPrice = null.StringFrom(askingPrice.String())
	case server.MarketplaceSaleTypeAuction:
		obj.Auction = true
		obj.AuctionCurrentPrice = null.StringFrom(askingPrice.String())
	case server.MarketplaceSaleTypeDutchAuction:
		obj.DutchAuction = true
		obj.DutchActionDropRate = null.StringFrom(dutchOptionDropRate.String())
		obj.BuyoutPrice = null.StringFrom(askingPrice.String())
	}
	err := obj.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	return obj, nil
}
