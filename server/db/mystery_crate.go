package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func IsMysteryCrateColumn(col string) bool {
	switch string(col) {
	case boiler.MysteryCrateColumns.Label:
		return true
	}
	return false
}

func PlayerMysteryCrate(id uuid.UUID) (*server.MysteryCrate, error) {
	queryMods := []qm.QueryMod{
		qm.Select(boiler.TableNames.CollectionItems + ".*"),
		qm.InnerJoin(fmt.Sprintf(
			"%s on %s = %s",
			boiler.TableNames.MysteryCrate,
			qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMysteryCrate),
		boiler.CollectionItemWhere.ItemID.EQ(id.String()),
	}

	collection, err := boiler.CollectionItems(queryMods...).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}

	crate, err := boiler.MysteryCrates(
		boiler.MysteryCrateWhere.ID.EQ(collection.ItemID),
	).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}

	return server.MysteryCrateFromBoiler(crate, collection, null.String{}), nil
}

func PlayerMysteryCrateList(
	search string,
	excludeOpened bool,
	includeMarketListed bool,
	excludeMarketLocked bool,
	userID *string,
	page int,
	pageSize int,
	sortBy string,
	sortDir SortByDir,
) (int64, []*server.MysteryCrate, error) {
	queryMods := []qm.QueryMod{
		qm.Select(boiler.TableNames.CollectionItems + ".*"),
		qm.InnerJoin(fmt.Sprintf(
			"%s on %s = %s",
			boiler.TableNames.MysteryCrate,
			qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMysteryCrate),
		boiler.CollectionItemWhere.XsynLocked.EQ(false),
	}

	if excludeOpened {
		queryMods = append(queryMods, boiler.MysteryCrateWhere.Opened.EQ(false))
	}
	if !includeMarketListed {
		queryMods = append(queryMods, boiler.CollectionItemWhere.LockedToMarketplace.EQ(false))
	}
	if excludeMarketLocked {
		queryMods = append(queryMods,
			boiler.CollectionItemWhere.XsynLocked.EQ(false),
			boiler.CollectionItemWhere.MarketLocked.EQ(false),
		)
	}
	if userID != nil {
		queryMods = append(queryMods, boiler.CollectionItemWhere.OwnerID.EQ(*userID))
	}

	// Search
	if search != "" {
		xSearch := ParseQueryText(search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"((to_tsvector('english', %[1]s.%[2]s) @@ to_tsquery(?))",
					boiler.TableNames.MysteryCrate,
					boiler.MysteryCrateColumns.Label,
				),
					xSearch,
				))
		}
	}

	total, err := boiler.CollectionItems(
		queryMods...,
	).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if total == 0 {
		return 0, []*server.MysteryCrate{}, nil
	}

	// Sort
	if IsMysteryCrateColumn(sortBy) && sortDir.IsValid() {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.MysteryCrate, sortBy, sortDir)))
	} else {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s desc, %s.%s desc", boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.Label, boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ID)))
	}

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize))
	}
	if page > 0 {
		queryMods = append(queryMods, qm.Offset(pageSize*(page-1)))
	}

	// Get Mystery Crates
	boil.DebugMode = true
	collectionItems, err := boiler.CollectionItems(queryMods...).All(gamedb.StdConn)
	boil.DebugMode = false
	if err != nil {
		return total, nil, terror.Error(err)
	}

	collectionItemIDs := []string{}
	mysteryCrateIDs := []string{}
	for _, ci := range collectionItems {
		mysteryCrateIDs = append(mysteryCrateIDs, ci.ItemID)
		collectionItemIDs = append(collectionItemIDs, ci.ID)
	}

	mysteryCrates, err := boiler.MysteryCrates(boiler.MysteryCrateWhere.ID.IN(mysteryCrateIDs)).All(gamedb.StdConn)
	if err != nil {
		return total, nil, terror.Error(err)
	}

	itemSales, err := boiler.ItemSales(
		boiler.ItemSaleWhere.CollectionItemID.IN(collectionItemIDs),
		boiler.ItemSaleWhere.SoldAt.IsNull(),
		boiler.ItemSaleWhere.DeletedAt.IsNull(),
		boiler.ItemSaleWhere.EndAt.GT(time.Now()),
	).All(gamedb.StdConn)
	if err != nil {
		return total, nil, terror.Error(err)
	}

	output := []*server.MysteryCrate{}
	for _, collectionItem := range collectionItems {
		var (
			mysteryCrate *boiler.MysteryCrate
			itemSale     *boiler.ItemSale
		)
		for _, mc := range mysteryCrates {
			if mc.ID == collectionItem.ItemID {
				mysteryCrate = mc
				break
			}
		}
		for _, is := range itemSales {
			if is.CollectionItemID == collectionItem.ID {
				itemSale = is
			}
		}
		if mysteryCrate == nil {
			return total, nil, terror.Error(fmt.Errorf("unable to find mystery crate from collection item %s", collectionItem.ItemID))
		}
		itemSaleID := null.String{}
		if itemSale != nil {
			itemSaleID = null.StringFrom(itemSale.ID)
		}
		item := server.MysteryCrateFromBoiler(mysteryCrate, collectionItem, itemSaleID)
		output = append(output, item)
	}
	return total, output, nil
}
