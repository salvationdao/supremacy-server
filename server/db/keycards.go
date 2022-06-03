package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	boiler.PlayerKeycardWhere.Count.GT(0),
}

func PlayerKeycard(id uuid.UUID) (*server.AssetKeycard, error) {
	item := &server.AssetKeycard{}
	err := boiler.NewQuery(append(keycardQueryMods, boiler.PlayerKeycardWhere.ID.EQ(id.String()))...).QueryRow(gamedb.StdConn).Scan(
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
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return item, nil
}

func PlayerKeycardList(
	search string,
	filter *ListFilterRequest,
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

	if userID != nil {
		queryMods = append(
			queryMods,
			qm.And(
				fmt.Sprintf("%s = ?", qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.PlayerID)),
				userID,
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
	countQueryMods := append(queryMods[1:], qm.Select("COUNT(*)"))
	err := boiler.NewQuery(countQueryMods...).QueryRow(gamedb.StdConn).Scan(&total)
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
