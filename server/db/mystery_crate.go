package db

import (
	"fmt"
	"server"

	"github.com/ninja-software/terror/v2"
)

func PlayerMysteryCrateList(
	search string,
	excludeMarketListed bool,
	userID *string,
	page int,
	pageSize int,
	sortBy string,
	sortDir SortByDir,
) (int64, []server.MysteryCrate, error) {
	if !sortDir.IsValid() {
		return 0, nil, terror.Error(fmt.Errorf("invalid sort direction"))
	}

	// if excludeMarketListed {
	// 	queryMods = append(queryMods, qm.And(fmt.Sprintf(
	// 		`%s - %s > 0`,
	// 		qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.Count),
	// 		NumKeycardsOnMarketplaceSQL,
	// 	)))
	// }

	return 0, nil, nil
}
