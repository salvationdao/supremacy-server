package db

import (
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var NumKeycardsOnMarketplaceSQL = fmt.Sprintf(`
	(
		SELECT COUNT(*)
		FROM %s
		WHERE %s = %s
			AND %s > NOW()
			AND %s IS NULL
	)`,
	boiler.TableNames.ItemKeycardSales,
	qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.ItemID),
	qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.ID),
	qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.EndAt),
	qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.DeletedAt),
)

var ActiveItemSaleSQL = fmt.Sprintf(`
	(
		SELECT COALESCE(array_agg(%s), '{}')
		FROM %s
		WHERE %s = %s 
			AND %s > NOW()
			AND %s IS NULL
			AND %s IS NULL
	)`,
	qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.ID),
	boiler.TableNames.ItemKeycardSales,
	qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.ItemID),
	qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.ID),
	qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.EndAt),
	qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.SoldAt),
	qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.DeletedAt),
)

var keycardQueryMods = []qm.QueryMod{
	qm.Select(
		qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.ID),
		qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.PlayerID),
		qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.BlueprintKeycardID),
		qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.Count),
		qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.CreatedAt),
		qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.ID),
		qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.Label),
		qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.Description),
		qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.Collection),
		qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.KeycardTokenID),
		qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.ImageURL),
		qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.AnimationURL),
		qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.KeycardGroup),
		qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.Syndicate),
		qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.CreatedAt),
		ActiveItemSaleSQL+" AS item_sale_ids",
		NumKeycardsOnMarketplaceSQL+" AS market_listed_count",
	),
	qm.From(boiler.TableNames.PlayerKeycards),
	qm.InnerJoin(
		fmt.Sprintf(
			"%s on %s = %s",
			boiler.TableNames.BlueprintKeycards,
			qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.ID),
			qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.BlueprintKeycardID),
		),
	),
}

func CreateOrGetKeycard(ownerID string, tokenID int) (*boiler.PlayerKeycard, error) {
	keycard, err := boiler.PlayerKeycards(
		boiler.PlayerKeycardWhere.PlayerID.EQ(ownerID),
		qm.InnerJoin(
			fmt.Sprintf(`%s ON %s = %s AND %s = $1`,
				boiler.TableNames.BlueprintKeycards,
				qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.ID),
				qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.BlueprintKeycardID),
				qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.KeycardTokenID),
			),
			tokenID,
		),
	).One(gamedb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		blueprint, err := boiler.BlueprintKeycards(boiler.BlueprintKeycardWhere.KeycardTokenID.EQ(tokenID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}

		newKeycard := &boiler.PlayerKeycard{
			PlayerID:           ownerID,
			BlueprintKeycardID: blueprint.ID,
			Count:              0,
		}

		err = newKeycard.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return nil, err
		}

		return newKeycard, nil
	}
	if err != nil {
		return nil, err
	}

	return keycard, nil
}

func UpdateKeycardReductionAmount(ownerID string, tokenID, amount int) error {
	q := `
		UPDATE player_keycards pk 
		SET count = count - $3
		WHERE pk.player_id = $1 AND pk.blueprint_keycard_id = (
			SELECT id FROM blueprint_keycards WHERE keycard_token_id = $2
		);`
	_, err := boiler.NewQuery(qm.SQL(q, ownerID, tokenID, amount)).Exec(gamedb.StdConn)
	if err != nil {
		return err
	}

	return nil
}

func PlayerKeycards(playerID string, ids ...string) ([]*server.AssetKeycard, error) {
	queries := keycardQueryMods
	if playerID != "" {
		queries = append(queries, boiler.PlayerKeycardWhere.PlayerID.EQ(playerID))
	}
	if len(ids) > 0 {
		queries = append(queries, boiler.PlayerKeycardWhere.ID.IN(ids))
	}
	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Failed to load player keycard.")
	}

	items := []*server.AssetKeycard{}
	for rows.Next() {
		item := &server.AssetKeycard{}
		err = rows.Scan(
			&item.ID,
			&item.PlayerID,
			&item.BlueprintKeycardID,
			&item.Count,
			&item.CreatedAt,
			&item.Blueprints.ID,
			&item.Blueprints.Label,
			&item.Blueprints.Description,
			&item.Blueprints.Collection,
			&item.Blueprints.KeycardTokenID,
			&item.Blueprints.ImageURL,
			&item.Blueprints.AnimationURL,
			&item.Blueprints.KeycardGroup,
			&item.Blueprints.Syndicate,
			&item.Blueprints.CreatedAt,
			&item.ItemSaleIDs,
			&item.MarketListedCount,
		)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to load asset keycard.")
			return nil, terror.Error(err, "Failed to load asset keycard.")
		}

		items = append(items, item)
	}

	return items, nil
}

func PlayerKeycardList(
	search string,
	filter *ListFilterRequest,
	includeMarketListed bool,
	userID *string,
	page int,
	pageSize int,
	sortBy string,
	sortDir SortByDir,
) (int64, []*server.AssetKeycard, error) {
	if !sortDir.IsValid() {
		return 0, nil, terror.Error(fmt.Errorf("invalid sort direction"))
	}

	queryMods := keycardQueryMods

	// Filters
	if filter != nil {
		for i, f := range filter.Items {
			queryMod := GenerateListFilterQueryMod(*f, i, filter.LinkOperator)
			queryMods = append(queryMods, queryMod)
		}
	}

	//if !includeMarketListed {
	//	queryMods = append(queryMods, qm.And(fmt.Sprintf(
	//		`%s - %s > 0`,
	//		qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.Count),
	//		NumKeycardsOnMarketplaceSQL,
	//	)))
	//}

	if userID != nil {
		queryMods = append(
			queryMods,
			qm.And(
				fmt.Sprintf("%s = ?", qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.PlayerID)),
				*userID,
			),
		)
	}

	// Search
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			queryMods = append(queryMods, qm.And(
				fmt.Sprintf(
					"(to_tsvector('english', %s) @@ to_tsquery(?))",
					qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.Label),
				),
				xsearch,
			))
		}
	}

	// Get total rows
	var total int64
	err := boiler.NewQuery(append(queryMods[1:], qm.Select("COUNT(*)"))...).QueryRow(gamedb.StdConn).Scan(&total)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if total == 0 {
		return 0, []*server.AssetKeycard{}, nil
	}

	// Sort
	orderBy := qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.CreatedAt), sortDir))
	queryMods = append(queryMods, orderBy)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize))
	}
	if page > 0 {
		queryMods = append(queryMods, qm.Offset(pageSize*(page-1)))
	}

	items := []*server.AssetKeycard{}
	err = boiler.NewQuery(queryMods...).Bind(nil, gamedb.StdConn, &items)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	return total, items, nil
}
