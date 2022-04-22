package db

import (
	"fmt"
	"server/db/boiler"
	"server/gamedb"

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
	sort *ListSortRequest,
	offset int,
	pageSize int,
) (int, []string, error) {
	queryMods := []qm.QueryMod{}

	// Filters
	if filter != nil {
		for i, f := range filter.Items {
			if f.Table != nil && *f.Table != "" {
				if *f.Table == boiler.TableNames.BlueprintPlayerAbilities {
					column := BlueprintPlayerAbilityColumn(f.Column)
					err := column.IsValid()
					if err != nil {
						return 0, nil, terror.Error(err)
					}
				} else if *f.Table == boiler.TableNames.SalePlayerAbilities {
					column := SalePlayerAbilityColumn(f.Column)
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

	// Search
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			queryMods = append(queryMods, qm.And(fmt.Sprintf("((to_tsvector('english', %[1]s.%[2]s) @@ to_tsquery(?)) OR (to_tsvector('english', %[1]s.%[3]s) @@ to_tsquery(?)))",
				boiler.TableNames.BlueprintPlayerAbilities,
				boiler.BlueprintPlayerAbilityColumns.Label,
				boiler.BlueprintPlayerAbilityColumns.Description,
			),
				xsearch,
			))
		}
	}

	// Sort
	orderBy := qm.OrderBy(fmt.Sprintf("%s desc", boiler.SalePlayerAbilityColumns.AvailableUntil))
	if sort != nil {
		sortColumn := sort.Column
		if sort.Table != nil && *sort.Table != "" {
			if *sort.Table == boiler.TableNames.BlueprintPlayerAbilities {
				column := BlueprintPlayerAbilityColumn(sort.Column)
				err := column.IsValid()
				if err != nil {
					return 0, nil, terror.Error(err)
				}
			} else if *sort.Table == boiler.TableNames.SalePlayerAbilities {
				column := SalePlayerAbilityColumn(sort.Column)
				err := column.IsValid()
				if err != nil {
					return 0, nil, terror.Error(err)
				}
			}
			sortColumn = fmt.Sprintf("%s.%s", *sort.Table, sort.Column)
		}
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", sortColumn, sort.Direction))
	}
	queryMods = append(queryMods, orderBy)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	queryMods = append(queryMods, qm.InnerJoin(fmt.Sprintf("%[1]s ON %[2]s.%[3]s = %[1]s.%[4]s", // "INNER JOIN blueprint_player_abilities ON sale_player_abilities.blueprint_id = blueprint_player_abilities.id"
		boiler.TableNames.BlueprintPlayerAbilities,
		boiler.TableNames.SalePlayerAbilities,
		boiler.SalePlayerAbilityColumns.BlueprintID,
		boiler.BlueprintPlayerAbilityColumns.ID,
	)))
	saleAbilities, err := boiler.SalePlayerAbilities(
		queryMods...,
	).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	sIDs := make([]string, 0)
	for _, s := range saleAbilities {
		sIDs = append(sIDs, s.ID)
	}

	return len(sIDs), sIDs, nil
}

// PlayerAbilitiesList returns a list of IDs from the player_abilities table.
// Filter and sorting options can be passed in to manipulate the end result.
func PlayerAbilitiesList(
	ctx context.Context,
	conn pgxscan.Querier,
	search string,
	filter *ListFilterRequest,
	sort *ListSortRequest,
	offset int,
	pageSize int,
) (int, []string, error) {
	queryMods := []qm.QueryMod{}

	// Filters
	if filter != nil {
		for i, f := range filter.Items {
			if f.Table != nil && *f.Table != "" {
				if *f.Table != boiler.TableNames.PlayerAbilities {
					return 0, nil, terror.Error(fmt.Errorf("invalid filter table name"))
				}
			}
			column := PlayerAbilityColumn(f.Column)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}
			queryMod := GenerateListFilterQueryMod(*f, i, filter.LinkOperator)
			queryMods = append(queryMods, queryMod)
		}
	}

	// Search
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			queryMods = append(queryMods, qm.And(fmt.Sprintf("((to_tsvector('english', %[1]s.%[2]s) @@ to_tsquery(?)) OR (to_tsvector('english', %[1]s.%[3]s) @@ to_tsquery(?)))",
				boiler.TableNames.PlayerAbilities,
				boiler.PlayerAbilityColumns.Label,
				boiler.PlayerAbilityColumns.Description,
			),
				xsearch,
			))
		}
	}

	// Sort
	orderBy := qm.OrderBy(fmt.Sprintf("%s desc", boiler.SalePlayerAbilityColumns.AvailableUntil))
	if sort != nil {
		sortColumn := sort.Column
		if sort.Table != nil && *sort.Table != "" {
			if *sort.Table != boiler.TableNames.PlayerAbilities {
				return 0, nil, terror.Error(fmt.Errorf("invalid sort table name"))
			}
			sortColumn = fmt.Sprintf("%s.%s", *sort.Table, sort.Column)
		}
		column := PlayerAbilityColumn(sort.Column)
		err := column.IsValid()
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", sortColumn, sort.Direction))
	}
	queryMods = append(queryMods, orderBy)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	playerAbilities, err := boiler.PlayerAbilities(
		queryMods...,
	).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	aIDs := make([]string, 0)
	for _, s := range playerAbilities {
		aIDs = append(aIDs, s.ID)
	}

	return len(aIDs), aIDs, nil
}
