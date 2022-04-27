package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// MarketplaceSaleList returns a numeric paginated result of sales list.
func MarketplaceSaleList(search string, archived bool, filter *ListFilterRequest, offset int, pageSize int, sortBy string, sortDir SortByDir) (int64, []*boiler.ItemSale, error) {
	queryMods := []qm.QueryMod{
		qm.LeftOuterJoin(
			fmt.Sprintf(
				"%s ON %s = %s AND %s = ?",
				boiler.TableNames.Mechs,
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
				qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ItemID),
				qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ItemType),
			),
			server.MarketplaceItemTypeMech,
		),
	}

	// Filters
	if filter != nil {
		for i, f := range filter.Items {
			if f.Table != nil && *f.Table != "" {
				if *f.Table == boiler.TableNames.Mechs {
					column := MechColumns(f.Column)
					err := column.IsValid()
					if err != nil {
						return 0, nil, terror.Error(err)
					}
				}
			}
			queryMod := GenerateListFilterQueryMod(*f, i, filter.LinkOperator)
			queryMods = append(queryMods, queryMod)
		}
	}

	// Search
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			queryMods = append(queryMods, qm.And(
				fmt.Sprintf(
					"((to_tsvector('english', %s) @@ to_tsquery(?))",
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Label),
				),
				xsearch,
			))
		}
	}

	// Get total rows
	total, err := boiler.ItemSales(queryMods...).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if total == 0 {
		return 0, []*boiler.ItemSale{}, nil
	}

	// Sort
	orderBy := qm.OrderBy(fmt.Sprintf("%s desc", qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.CreatedAt)))
	queryMods = append(queryMods, orderBy)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	result, err := boiler.ItemSales(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	return total, result, nil
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

// MarketplaceSaleItemExists checks whether given sales item exists.
func MarketplaceSaleItemExists(id uuid.UUID) (bool, error) {
	return boiler.ItemSaleExists(gamedb.StdConn, id.String())
}
