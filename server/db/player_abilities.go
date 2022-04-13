package db

import (
	"fmt"
	"server/db/boiler"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/net/context"
)

type (
	PlayerAbilityColumn string
)

func (p PlayerAbilityColumn) IsValid() error {

	switch string(p) {
	case
		boiler.BlueprintPlayerAbilityColumns.ID,
		boiler.BlueprintPlayerAbilityColumns.GameClientAbilityID,
		boiler.BlueprintPlayerAbilityColumns.Label,
		boiler.BlueprintPlayerAbilityColumns.Colour,
		boiler.BlueprintPlayerAbilityColumns.ImageURL,
		boiler.BlueprintPlayerAbilityColumns.Description,
		boiler.BlueprintPlayerAbilityColumns.TextColour,
		boiler.BlueprintPlayerAbilityColumns.Type:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid player ability column"))
}

type SaleAbilityDetailed struct {
	*boiler.SalePlayerAbility
	PlayerAbility boiler.BlueprintPlayerAbility `json:"player_ability"`
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
	sortBy PlayerAbilityColumn,
	sortDir SortByDir,
) (int, []SaleAbilityDetailed, error) {
	spaAlias := "spa" // alias for sale_player_abilities table
	bpaAlias := "bpa" /// alias for blueprint_player_abilities table
	fromQ := fmt.Sprintf("FROM %s %s\n", boiler.TableNames.SalePlayerAbilities, spaAlias) +
		fmt.Sprintf("INNER JOIN %[1]s %[4]s ON %[5]s.%[2]s = %[4]s.%[3]s\n", boiler.TableNames.BlueprintPlayerAbilities, boiler.SalePlayerAbilityColumns.BlueprintID, boiler.BlueprintPlayerAbilityColumns.ID, bpaAlias, spaAlias)

	selectQ := "SELECT\n" +
		fmt.Sprintf("%s.%s,\n", spaAlias, boiler.SalePlayerAbilityColumns.BlueprintID) +
		fmt.Sprintf("%s.%s,\n", spaAlias, boiler.SalePlayerAbilityColumns.CurrentPrice) +
		fmt.Sprintf("%s.%s,\n", spaAlias, boiler.SalePlayerAbilityColumns.AvailableUntil) +
		fmt.Sprintf("%s.%s,\n", bpaAlias, boiler.BlueprintPlayerAbilityColumns.ID) +
		fmt.Sprintf("%s.%s,\n", bpaAlias, boiler.BlueprintPlayerAbilityColumns.GameClientAbilityID) +
		fmt.Sprintf("%s.%s,\n", bpaAlias, boiler.BlueprintPlayerAbilityColumns.Label) +
		fmt.Sprintf("%s.%s,\n", bpaAlias, boiler.BlueprintPlayerAbilityColumns.Colour) +
		fmt.Sprintf("%s.%s,\n", bpaAlias, boiler.BlueprintPlayerAbilityColumns.ImageURL) +
		fmt.Sprintf("%s.%s,\n", bpaAlias, boiler.BlueprintPlayerAbilityColumns.Description) +
		fmt.Sprintf("%s.%s,\n", bpaAlias, boiler.BlueprintPlayerAbilityColumns.TextColour) +
		fmt.Sprintf("%s.%s,\n", bpaAlias, boiler.BlueprintPlayerAbilityColumns.Type) + fromQ

	var args []interface{}

	// Prepare Filters
	filterConditionsString := ""
	argIndex := 1
	if filter != nil {
		filterConditions := []string{}
		for _, f := range filter.Items {
			column := PlayerAbilityColumn(f.ColumnField)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}

			condition, value := GenerateListFilterSQL(f.ColumnField, f.Value, f.OperatorValue, argIndex)
			if condition != "" {
				switch f.OperatorValue {
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
		boiler.SalePlayerAbilityColumns.BlueprintID,
		selectQ,
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
		return 0, make([]SaleAbilityDetailed, 0), nil
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
		boiler.SalePlayerAbilityColumns.BlueprintID,
		filterConditionsString,
		searchCondition,
		orderBy,
		limit,
	)

	result := make([]struct {
		*boiler.SalePlayerAbility
		*boiler.BlueprintPlayerAbility
	}, 0)
	err = pgxscan.Select(ctx, conn, &result, q, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	resultResult := make([]SaleAbilityDetailed, 0)
	for _, r := range result {
		resultResult = append(resultResult, SaleAbilityDetailed{
			SalePlayerAbility: &boiler.SalePlayerAbility{
				BlueprintID:    r.BlueprintID,
				CurrentPrice:   r.CurrentPrice,
				AvailableUntil: r.AvailableUntil,
			},
			PlayerAbility: boiler.BlueprintPlayerAbility{
				ID:                  r.ID,
				GameClientAbilityID: r.GameClientAbilityID,
				Label:               r.Label,
				Colour:              r.Colour,
				ImageURL:            r.ImageURL,
				Description:         r.Description,
				TextColour:          r.TextColour,
				Type:                r.Type,
			},
		})
	}

	return totalRows, resultResult, nil
}

// PlayerAbilitiesList returns a list of IDs from the sale_player_abilities table.
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
) (int, boiler.PlayerAbilitySlice, error) {
	paAlias := "pa" // alias for sale_player_abilities table
	fromQ := fmt.Sprintf("FROM %s %s\n", boiler.TableNames.SalePlayerAbilities, paAlias)

	selectQ := "SELECT\n" +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.ID) +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.OwnerID) +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.BlueprintID) +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.GameClientAbilityID) +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.Label) +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.Colour) +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.ImageURL) +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.Description) +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.TextColour) +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.Type) +
		fmt.Sprintf("%s.%s,\n", paAlias, boiler.PlayerAbilityColumns.PurchasedAt) + fromQ

	var args []interface{}

	// Prepare Filters
	filterConditionsString := ""
	argIndex := 1
	if filter != nil {
		filterConditions := []string{}
		for _, f := range filter.Items {
			column := PlayerAbilityColumn(f.ColumnField)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}

			condition, value := GenerateListFilterSQL(f.ColumnField, f.Value, f.OperatorValue, argIndex)
			if condition != "" {
				switch f.OperatorValue {
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
			searchCondition = fmt.Sprintf(" AND ((to_tsvector('english', %[4]s.%[1]s) @@ to_tsquery($%[3]d)) OR (to_tsvector('english', %[4]s.%[2]s) @@ to_tsquery($%[3]d)))", boiler.BlueprintPlayerAbilityColumns.Label, boiler.BlueprintPlayerAbilityColumns.Description, len(args), paAlias)
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
		boiler.SalePlayerAbilityColumns.BlueprintID,
		selectQ,
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
		return 0, make(boiler.PlayerAbilitySlice, 0), nil
	}

	// Order and Limit
	orderBy := fmt.Sprintf(" ORDER BY %s.%s DESC", paAlias, boiler.SalePlayerAbilityColumns.AvailableUntil)
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
		boiler.SalePlayerAbilityColumns.BlueprintID,
		filterConditionsString,
		searchCondition,
		orderBy,
		limit,
	)

	result := make(boiler.PlayerAbilitySlice, 0)
	err = pgxscan.Select(ctx, conn, &result, q, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	return totalRows, result, nil
}
