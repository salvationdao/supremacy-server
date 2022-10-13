package db

import (
	"fmt"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type SortBy string

const (
	SortByAlphabetical SortBy = "alphabetical"
	SortByRarity       SortBy = "rarity"
	SortByDate         SortBy = "date"
)

type SortByDir string

const (
	SortByDirAsc  SortByDir = "asc"
	SortByDirDesc SortByDir = "desc"
)

type ListSortRequest struct {
	Table     string    `json:"table"`
	Column    string    `json:"column"`
	Direction SortByDir `json:"direction"`
}

func (s SortByDir) IsValid() bool {
	switch s {
	case SortByDirAsc, SortByDirDesc:
		return true
	default:
		return false
	}
}

func (ls *ListSortRequest) IsValid() bool {
	return ls.Table != "" && ls.Column != "" && ls.Direction.IsValid()
}

func (ls *ListSortRequest) GenQueryMod() qm.QueryMod {
	return qm.OrderBy(fmt.Sprintf("%s.%s %s", ls.Table, ls.Column, ls.Direction))
}
