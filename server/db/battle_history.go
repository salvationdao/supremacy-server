package db

import (
	"fmt"
	"server/db/boiler"
	"server/gamedb"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type (
	BattleMechColumn string
	BattleColumn     string
)

func (p BattleMechColumn) IsValid() error {
	switch string(p) {
	case
		boiler.BattleMechColumns.BattleID,
		boiler.BattleMechColumns.MechID,
		boiler.BattleMechColumns.OwnerID,
		boiler.BattleMechColumns.FactionID,
		boiler.BattleMechColumns.Killed,
		boiler.BattleMechColumns.KilledByID,
		boiler.BattleMechColumns.Kills,
		boiler.BattleMechColumns.DamageTaken,
		boiler.BattleMechColumns.UpdatedAt,
		boiler.BattleMechColumns.CreatedAt,
		boiler.BattleMechColumns.FactionWon,
		boiler.BattleMechColumns.MechSurvived:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid battle mech column"))
}

type BattleDetailed struct {
	*boiler.Battle
	GameMap *boiler.GameMap `json:"game_map"`
}

type BattleMechDetailed struct {
	*boiler.BattleMech
	Battle *BattleDetailed `json:"battle"`
}

// BattleMechsListPaginated returns a list of IDs from the battle_mechs table.
// Filter and sorting options can be passed in to manipulate the end result.
func BattleMechsListPaginated(
	filter *ListFilterRequest,
	sort *ListSortRequest,
	offset int,
	pageSize int,
) (int64, []*BattleMechDetailed, error) {
	queryMods := []qm.QueryMod{}

	// Filters
	if filter != nil {
		for i, f := range filter.Items {
			if f.Table != nil && *f.Table != "" {
				if *f.Table != boiler.TableNames.BattleMechs {
					return 0, nil, terror.Error(fmt.Errorf("invalid filter table name"))
				}
			}
			column := BattleMechColumn(f.Column)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}
			queryMod := GenerateListFilterQueryMod(*f, i, filter.LinkOperator)
			queryMods = append(queryMods, queryMod)
		}
	}

	total, err := boiler.BattleMechs(queryMods...).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}

	// Sort
	orderBy := qm.OrderBy(fmt.Sprintf("%s desc", boiler.BattleMechColumns.CreatedAt))
	if sort != nil {
		sortColumn := sort.Column
		if sort.Table != nil && *sort.Table != "" {
			if *sort.Table != boiler.TableNames.BattleMechs {
				return 0, nil, terror.Error(fmt.Errorf("invalid filter table name"))
			}
			column := BattleMechColumn(sort.Column)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}
		}
		sortColumn = fmt.Sprintf("%s.%s", *sort.Table, sort.Column)
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", sortColumn, sort.Direction))
	}
	queryMods = append(queryMods, orderBy)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	queryMods = append(queryMods, qm.Load(qm.Rels(boiler.BattleMechRels.Battle, boiler.BattleRels.GameMap)))
	battleMechs, err := boiler.BattleMechs(
		queryMods...,
	).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	detailedBattleMechs := []*BattleMechDetailed{}
	for _, b := range battleMechs {
		detailedBattleMechs = append(detailedBattleMechs, &BattleMechDetailed{
			BattleMech: b,
			Battle: &BattleDetailed{
				Battle:  b.R.Battle,
				GameMap: b.R.Battle.R.GameMap,
			},
		})
	}

	return total, detailedBattleMechs, nil
}
