package db

import "server/db/boiler"

// MarketplaceSaleList returns a numeric paginated result of sales list.
func MarketplaceSaleList(search string, archived bool, filter *ListFilterRequest, offset int, pageSize int, sortBy string, sortDir SortByDir) (int, []*boiler.ItemSale, error) {
	return 0, nil, nil
}
