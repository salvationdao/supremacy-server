package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func IsMysteryCrateColumn(col string) bool {
	switch string(col) {
	case boiler.MysteryCrateColumns.Label:
		return true
	}
	return false
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
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s desc", boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.Label)))
	}

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize))
	}
	if page > 0 {
		queryMods = append(queryMods, qm.Offset(pageSize*(page-1)))
	}

	// Get Mystery Crates
	collectionItems, err := boiler.CollectionItems(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return total, nil, terror.Error(err)
	}

	mysteryCrateIDs := []string{}
	for _, ci := range collectionItems {
		mysteryCrateIDs = append(mysteryCrateIDs, ci.ItemID)
	}

	mysteryCrates, err := boiler.MysteryCrates(boiler.MysteryCrateWhere.ID.IN(mysteryCrateIDs)).All(gamedb.StdConn)
	if err != nil {
		return total, nil, terror.Error(err)
	}

	output := []*server.MysteryCrate{}
	for _, collectionItem := range collectionItems {
		var mysteryCrate *boiler.MysteryCrate
		for _, mc := range mysteryCrates {
			if mc.ID == collectionItem.ItemID {
				mysteryCrate = mc
				break
			}
		}
		if mysteryCrate == nil {
			return total, nil, terror.Error(fmt.Errorf("unable to find mystery crate from collection item %s", collectionItem.ItemID))
		}
		item := server.MysteryCrateFromBoiler(mysteryCrate, collectionItem)
		output = append(output, item)
	}
	return total, output, nil
}
