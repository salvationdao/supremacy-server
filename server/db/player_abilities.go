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

type SalePlayerAbilityDetailed struct {
	*boiler.SalePlayerAbility
	PlayerAbility boiler.BlueprintPlayerAbility `json:"player_ability"`
}

func SalePlayerAbilitiesList(
	ctx context.Context,
	conn pgxscan.Querier,
	search string,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy PlayerAbilityColumn,
	sortDir SortByDir,
) (int, []SalePlayerAbilityDetailed, error) {
	fromQ := fmt.Sprintf("FROM %s spa\n", boiler.TableNames.SalePlayerAbilities) +
		fmt.Sprintf("INNER JOIN %[1]s bpa ON spa.%[2]s = bpa.%[3]s\n", boiler.TableNames.BlueprintPlayerAbilities, boiler.SalePlayerAbilityColumns.BlueprintID, boiler.BlueprintPlayerAbilityColumns.ID)

	selectQ := "SELECT\n" +
		fmt.Sprintf("spa.%s,\n", boiler.SalePlayerAbilityColumns.BlueprintID) +
		fmt.Sprintf("spa.%s,\n", boiler.SalePlayerAbilityColumns.CurrentPrice) +
		fmt.Sprintf("spa.%s,\n", boiler.SalePlayerAbilityColumns.AvailableUntil) +
		fmt.Sprintf("bpa.%s,\n", boiler.BlueprintPlayerAbilityColumns.ID) +
		fmt.Sprintf("bpa.%s,\n", boiler.BlueprintPlayerAbilityColumns.GameClientAbilityID) +
		fmt.Sprintf("bpa.%s,\n", boiler.BlueprintPlayerAbilityColumns.Label) +
		fmt.Sprintf("bpa.%s,\n", boiler.BlueprintPlayerAbilityColumns.Colour) +
		fmt.Sprintf("bpa.%s,\n", boiler.BlueprintPlayerAbilityColumns.ImageURL) +
		fmt.Sprintf("bpa.%s,\n", boiler.BlueprintPlayerAbilityColumns.Description) +
		fmt.Sprintf("bpa.%s,\n", boiler.BlueprintPlayerAbilityColumns.TextColour) +
		fmt.Sprintf("bpa.%s,\n", boiler.BlueprintPlayerAbilityColumns.Type) + fromQ

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
			searchCondition = fmt.Sprintf(" AND ((to_tsvector('english', bpa.%[1]s) @@ to_tsquery($%[3]d)) OR (to_tsvector('english', bpa.%[2]s) @@ to_tsquery($%[3]d)))", boiler.BlueprintPlayerAbilityColumns.Label, boiler.BlueprintPlayerAbilityColumns.Description, len(args))
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT spa.%[1]s)
		%[2]s
		WHERE spa.%[1]s IS NOT NULL
			%[3]s
			%[4]s
		`,
		boiler.SalePlayerAbilityColumns.BlueprintID,
		selectQ,
		filterConditionsString,
		searchCondition,
	)

	var totalRows int

	err := pgxscan.Get(ctx, conn, &totalRows, countQ, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if totalRows == 0 {
		return 0, make([]SalePlayerAbilityDetailed, 0), nil
	}

	// Order and Limit
	orderBy := fmt.Sprintf(" ORDER BY spa.%s DESC", boiler.SalePlayerAbilityColumns.AvailableUntil)
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
		WHERE spa.%s IS NOT NULL
			%s
			%s
		%s
		%s`,
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

	resultResult := make([]SalePlayerAbilityDetailed, 0)
	for _, r := range result {
		resultResult = append(resultResult, SalePlayerAbilityDetailed{
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
