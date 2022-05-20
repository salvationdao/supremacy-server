package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var itemSaleQueryMods = []qm.QueryMod{
	qm.InnerJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.CollectionItems,
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ItemID),
		),
	),
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ItemID),
		),
	),
	qm.Load(boiler.ItemSaleRels.Owner),
}

// MarketplaceLoadItemSaleObject loads the specific item type's object.
func MarketplaceLoadItemSaleObject(obj *server.MarketplaceSaleItem) (*server.MarketplaceSaleItem, error) {
	// if obj.ItemType == boiler.ItemTypeMech {
	mech, err := boiler.Mechs(
		boiler.MechWhere.ID.EQ(obj.ItemID),
	).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}
	obj.Mech = mech
	// }

	return obj, nil
}

// MarketplaceItemSale gets a specific item sale.
func MarketplaceItemSale(id uuid.UUID) (*server.MarketplaceSaleItem, error) {
	item, err := boiler.ItemSales(
		append(
			itemSaleQueryMods,
			boiler.ItemSaleWhere.ID.EQ(id.String()),
		)...,
	).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}

	// Get sale item details and the item for sale
	output := &server.MarketplaceSaleItem{
		ItemSale: item,
		Owner:    item.R.Owner,
	}
	output, err = MarketplaceLoadItemSaleObject(output)
	if err != nil {
		return nil, terror.Error(err)
	}
	return output, nil
}

// MarketplaceItemSaleList returns a numeric paginated result of sales list.
func MarketplaceItemSaleList(search string, archived bool, filter *ListFilterRequest, rarities []string, offset int, pageSize int, sortBy string, sortDir SortByDir) (int64, []*server.MarketplaceSaleItem, error) {
	queryMods := itemSaleQueryMods

	// Filters
	if filter != nil {
		for i, f := range filter.Items {
			if f.Table != "" {
				if f.Table == boiler.TableNames.Mechs {
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
	if len(rarities) > 0 {
		queryMods = append(queryMods, boiler.CollectionItemWhere.Tier.IN(rarities))
	}

	// Search
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			queryMods = append(queryMods, qm.And(
				fmt.Sprintf(
					"(to_tsvector('english', %s) @@ to_tsquery(?))",
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
		return 0, []*server.MarketplaceSaleItem{}, nil
	}

	// Sort
	orderBy := qm.OrderBy(fmt.Sprintf("%s desc", qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.CreatedAt)))
	queryMods = append(queryMods, orderBy)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	itemSales, err := boiler.ItemSales(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	// Load in related items
	records := []*server.MarketplaceSaleItem{}
	itemIDs := []string{}
	mechIDs := []string{}
	for _, row := range itemSales {
		// if row.ItemType == boiler.ItemTypeMech {
		mechIDs = append(mechIDs, row.ItemID)
		itemIDs = append(itemIDs, row.ItemID)
		// }
		records = append(records, &server.MarketplaceSaleItem{
			ItemSale: row,
			Owner:    row.R.Owner,
		})
	}
	if len(itemIDs) > 0 {
		collectionItems, err := boiler.CollectionItems(
			boiler.CollectionItemWhere.ItemID.IN(itemIDs),
		).All(gamedb.StdConn)
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		for i, row := range records {
			for _, collection := range collectionItems {
				// if row.ItemType == boiler.ItemTypeMech && row.ItemID == mech.ID {
				if row.ItemID == collection.ItemID {
					records[i].Collection = collection
				}
			}
		}
	}
	if len(mechIDs) > 0 {
		mechs, err := boiler.Mechs(
			boiler.MechWhere.ID.IN(mechIDs),
		).All(gamedb.StdConn)
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		for i, row := range records {
			for _, mech := range mechs {
				// if row.ItemType == boiler.ItemTypeMech && row.ItemID == mech.ID {
				if row.ItemID == mech.ID {
					records[i].Mech = mech
				}
			}
		}
	}

	return total, records, nil
}

// MarketplaceSaleCreate inserts a new sale item.
func MarketplaceSaleCreate(saleType server.MarketplaceSaleType, ownerID uuid.UUID, factionID uuid.UUID, listFeeTxnID string, endAt time.Time, itemID uuid.UUID, askingPrice *decimal.Decimal, dutchOptionDropRate *decimal.Decimal) (*server.MarketplaceSaleItem, error) {
	obj := &boiler.ItemSale{
		OwnerID:        ownerID.String(),
		FactionID:      factionID.String(),
		ListingFeeTXID: listFeeTxnID,
		ItemID:         itemID.String(),
		EndAt:          endAt,
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
	output := &server.MarketplaceSaleItem{
		ItemSale: obj,
	}
	return output, nil
}

// MarketplaceSaleBidHistoryCreate inserts a new bid history record.
func MarketplaceSaleBidHistoryCreate(id uuid.UUID, bidderUserID uuid.UUID, amount decimal.Decimal) (*boiler.ItemSalesBidHistory, error) {
	obj := &boiler.ItemSalesBidHistory{
		ItemSaleID: id.String(),
		BidderID:   bidderUserID.String(),
		BidPrice:   amount.String(),
	}
	err := obj.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	return obj, nil
}

// MarketplaceSaleAuctionSync updates the current auction price based on the bid history.
func MarketplaceSaleAuctionSync(id uuid.UUID) error {
	q := fmt.Sprintf(
		`UPDATE %s
		SET %s = (
			SELECT %s
			FROM %s
			WHERE %s = $1 
		)
		WHERE %s = $1`,
		boiler.TableNames.ItemSales,
		boiler.ItemSaleColumns.AuctionCurrentPrice,
		boiler.ItemSalesBidHistoryColumns.BidPrice,
		boiler.TableNames.ItemSalesBidHistory,
		boiler.ItemSalesBidHistoryColumns.ItemSaleID,
		boiler.ItemSaleColumns.ID,
	)
	_, err := gamedb.StdConn.Exec(q, id)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceSaleItemExists checks whether given sales item exists.
func MarketplaceSaleItemExists(id uuid.UUID) (bool, error) {
	return boiler.ItemSaleExists(gamedb.StdConn, id.String())
}
