package db

type SortByDir string

const (
	SortByDirAsc  SortByDir = "asc"
	SortByDirDesc SortByDir = "desc"
)

type ListSortRequest struct {
	Table     *string   `json:"table"`
	Column    string    `json:"column"`
	Direction SortByDir `json:"direction"`
}
