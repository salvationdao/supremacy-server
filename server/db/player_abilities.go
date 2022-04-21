package db

import (
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/net/context"
)

type (
	SalePlayerAbilityColumn      string
	PlayerAbilityColumn          string
	BlueprintPlayerAbilityColumn string
)

func (p SalePlayerAbilityColumn) IsValid() error {
	switch string(p) {
	case
		boiler.SalePlayerAbilityColumns.ID,
		boiler.SalePlayerAbilityColumns.BlueprintID,
		boiler.SalePlayerAbilityColumns.CurrentPrice,
		boiler.SalePlayerAbilityColumns.AvailableUntil:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid sale player ability column"))
}

func (p PlayerAbilityColumn) IsValid() error {
	switch string(p) {
	case
		boiler.PlayerAbilityColumns.ID,
		boiler.PlayerAbilityColumns.GameClientAbilityID,
		boiler.PlayerAbilityColumns.Label,
		boiler.PlayerAbilityColumns.Colour,
		boiler.PlayerAbilityColumns.ImageURL,
		boiler.PlayerAbilityColumns.Description,
		boiler.PlayerAbilityColumns.TextColour,
		boiler.PlayerAbilityColumns.LocationSelectType,
		boiler.PlayerAbilityColumns.PurchasedAt:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid player ability column"))
}

func (p BlueprintPlayerAbilityColumn) IsValid() error {
	switch string(p) {
	case
		boiler.BlueprintPlayerAbilityColumns.ID,
		boiler.BlueprintPlayerAbilityColumns.GameClientAbilityID,
		boiler.BlueprintPlayerAbilityColumns.Label,
		boiler.BlueprintPlayerAbilityColumns.Colour,
		boiler.BlueprintPlayerAbilityColumns.ImageURL,
		boiler.BlueprintPlayerAbilityColumns.Description,
		boiler.BlueprintPlayerAbilityColumns.TextColour,
		boiler.BlueprintPlayerAbilityColumns.LocationSelectType:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid blueprint player ability column"))
}

type SaleAbilityDetailed struct {
	*boiler.SalePlayerAbility
	Ability *boiler.BlueprintPlayerAbility `json:"ability,omitempty"`
}

func SaleAbilityGet(
	ctx context.Context,
	conn pgxscan.Querier,
	abilityID string,
) (*SaleAbilityDetailed, error) {
	spa, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.ID.EQ(abilityID), qm.Load(boiler.SalePlayerAbilityRels.Blueprint)).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}

	result := SaleAbilityDetailed{
		SalePlayerAbility: spa,
		Ability:           spa.R.Blueprint,
	}

	return &result, nil
}

// SaleAbilitiesList returns a list of IDs from the sale_player_abilities table.
// Filter and sorting options can be passed in to manipulate the end result.
func SaleAbilitiesList(
	ctx context.Context,
	conn pgxscan.Querier,
	search string,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy SalePlayerAbilityColumn,
	sortDir SortByDir,
) (int, []string, error) {

	spaAlias := "spa" // alias for sale_player_abilities table
	bpaAlias := "bpa" /// alias for blueprint_player_abilities table
	fromQ := fmt.Sprintf("FROM %s %s\n", boiler.TableNames.SalePlayerAbilities, spaAlias) +
		fmt.Sprintf("INNER JOIN %[1]s %[4]s ON %[5]s.%[2]s = %[4]s.%[3]s\n", boiler.TableNames.BlueprintPlayerAbilities, boiler.SalePlayerAbilityColumns.BlueprintID, boiler.BlueprintPlayerAbilityColumns.ID, bpaAlias, spaAlias)
	selectQ := "SELECT\n" +
		fmt.Sprintf("%s.%s\n", spaAlias, boiler.SalePlayerAbilityColumns.ID) + fromQ

	saleAbilities := boiler.SalePlayerAbilities(
		qm.InnerJoin(fmt.Sprintf("%[1]s %[4]s ON %[5]s.%[2]s = %[4]s.%[3]s", // "blueprint_player_abilities bpa ON spa.blueprint_id = bpa.id"
			boiler.TableNames.BlueprintPlayerAbilities,
			boiler.SalePlayerAbilityColumns.BlueprintID,
			boiler.BlueprintPlayerAbilityColumns.ID,
			bpaAlias,
			spaAlias)),
	)

	filterQueryMods := []qm.QueryMod{}
	if filter != nil {
		for i, f := range filter.Items {
			if f.Table == boiler.TableNames.BlueprintPlayerAbilities {
				column := BlueprintPlayerAbilityColumn(f.Column)
				err := column.IsValid()
				if err != nil {
					return 0, nil, terror.Error(err)
				}
			} else if f.Table == boiler.TableNames.SalePlayerAbilities {
				column := SalePlayerAbilityColumn(f.Column)
				err := column.IsValid()
				if err != nil {
					return 0, nil, terror.Error(err)
				}
			}

			queryMod := GenerateListFilterQueryMod(f.Table, f.Column, f.Value, f.Operator, i, filter.LinkOperator)
			filterQueryMods = append(filterQueryMods, queryMod)
		}
	}

	var args []interface{}
	// Prepare Filters
	filterConditionsString := ""
	argIndex := 1
	if filter != nil {
		filterConditions := []string{}
		for _, f := range filter.Items {
			column := SalePlayerAbilityColumn(f.Column)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}

			condition, value := GenerateListFilterSQL(f.Column, f.Value, f.Operator, argIndex)
			if condition != "" {
				switch f.Operator {
				case OperatorValueTypeIsNull, OperatorValueTypeIsNotNull:
					break
				default:
					argIndex += 1
					args = append(args, value)
				}
				filterConditions = append(filterConditions, condition)
			}
		}
		if len(filterConditions) > 0 {
			filterConditionsString = " AND (" + strings.Join(filterConditions, " "+string(filter.LinkOperator)+" ") + ")"
		}
	}

	searchCondition := ""
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			args = append(args, xsearch)
			searchCondition = fmt.Sprintf(" AND ((to_tsvector('english', %[4]s.%[1]s) @@ to_tsquery($%[3]d)) OR (to_tsvector('english', %[4]s.%[2]s) @@ to_tsquery($%[3]d)))", boiler.BlueprintPlayerAbilityColumns.Label, boiler.BlueprintPlayerAbilityColumns.Description, len(args), bpaAlias)
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT %[5]s.%[1]s)
		%[2]s
		WHERE %[5]s.%[1]s IS NOT NULL
			%[3]s
			%[4]s
		`,
		boiler.SalePlayerAbilityColumns.ID,
		fromQ,
		filterConditionsString,
		searchCondition,
		spaAlias,
	)

	var totalRows int
	err := pgxscan.Get(ctx, conn, &totalRows, countQ, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if totalRows == 0 {
		return 0, make([]string, 0), nil
	}

	// Order and Limit
	orderBy := fmt.Sprintf(" ORDER BY %s.%s DESC", spaAlias, boiler.SalePlayerAbilityColumns.AvailableUntil)
	if sortBy != "" {
		err := sortBy.IsValid()
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		orderBy = fmt.Sprintf(" ORDER BY %s %s", sortBy, sortDir)
	}
	limit := ""
	if pageSize > 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	// Get Paginated Result
	q := fmt.Sprintf(
		selectQ+`--sql
		WHERE %s.%s IS NOT NULL
			%s
			%s
		%s
		%s`,
		spaAlias,
		boiler.SalePlayerAbilityColumns.ID,
		filterConditionsString,
		searchCondition,
		orderBy,
		limit,
	)

	result := make([]string, 0)
	err = pgxscan.Select(ctx, conn, &result, q, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	return totalRows, result, nil
}

// PlayerAbilitiesList returns a list of IDs from the player_abilities table.
// Filter and sorting options can be passed in to manipulate the end result.
func PlayerAbilitiesList(
	ctx context.Context,
	conn pgxscan.Querier,
	search string,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy PlayerAbilityColumn,
	sortDir SortByDir,
) (int, []string, error) {
	paAlias := "pa" // alias for player_abilities table
	fromQ := fmt.Sprintf("FROM %s %s\n", boiler.TableNames.PlayerAbilities, paAlias)

	selectQ := "SELECT\n" +
		fmt.Sprintf("%s.%s\n", paAlias, boiler.PlayerAbilityColumns.ID) + fromQ

	var args []interface{}

	// Prepare Filters
	filterConditionsString := ""
	argIndex := 1
	if filter != nil {
		filterConditions := []string{}
		for _, f := range filter.Items {
			column := PlayerAbilityColumn(f.Column)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}

			condition, value := GenerateListFilterSQL(f.Column, f.Value, f.Operator, argIndex)
			if condition != "" {
				switch f.Operator {
				case OperatorValueTypeIsNull, OperatorValueTypeIsNotNull:
					break
				default:
					argIndex += 1
					args = append(args, value)
				}
				filterConditions = append(filterConditions, condition)
			}
		}
		if len(filterConditions) > 0 {
			filterConditionsString = " AND (" + strings.Join(filterConditions, " "+string(filter.LinkOperator)+" ") + ")"
		}
	}

	searchCondition := ""
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			args = append(args, xsearch)
			searchCondition = fmt.Sprintf(" AND ((to_tsvector('english', %[4]s.%[1]s) @@ to_tsquery($%[3]d)) OR (to_tsvector('english', %[4]s.%[2]s) @@ to_tsquery($%[3]d)))", boiler.PlayerAbilityColumns.Label, boiler.PlayerAbilityColumns.Description, len(args), paAlias)
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT %[5]s.%[1]s)
		%[2]s
		WHERE %[5]s.%[1]s IS NOT NULL
			%[3]s
			%[4]s
		`,
		boiler.PlayerAbilityColumns.ID,
		fromQ,
		filterConditionsString,
		searchCondition,
		paAlias,
	)

	var totalRows int

	err := pgxscan.Get(ctx, conn, &totalRows, countQ, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if totalRows == 0 {
		return 0, make([]string, 0), nil
	}

	// Order and Limit
	orderBy := fmt.Sprintf(" ORDER BY %s.%s DESC", paAlias, boiler.PlayerAbilityColumns.PurchasedAt)
	if sortBy != "" {
		err := sortBy.IsValid()
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		orderBy = fmt.Sprintf(" ORDER BY %s %s", sortBy, sortDir)
	}
	limit := ""
	if pageSize > 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	// Get Paginated Result
	q := fmt.Sprintf(
		selectQ+`--sql
		WHERE %s.%s IS NOT NULL
			%s
			%s
		%s
		%s`,
		paAlias,
		boiler.PlayerAbilityColumns.ID,
		filterConditionsString,
		searchCondition,
		orderBy,
		limit,
	)

	result := make([]string, 0)
	err = pgxscan.Select(ctx, conn, &result, q, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	return totalRows, result, nil
}
