package db

import (
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AvatarLayersListOpts struct {
	Search    string
	Filter    *ListFilterRequest
	LayerType null.String
	Sort      *ListSortRequest
	PageSize  int
	Page      int
}

type Layer struct {
	ID       string      `json:"id"`
	ImageURL null.String `json:"image_url,omitempty"`
	Type     string      `json:"type"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func LayersList(opts *AvatarLayersListOpts) (int64, []*Layer, error) {
	var layers []*Layer

	var queryMods []qm.QueryMod

	total, err := boiler.Layers(
		queryMods...,
	).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}

	// Limit/Offset
	if opts.PageSize > 0 {
		queryMods = append(queryMods, qm.Limit(opts.PageSize))
	}
	if opts.Page > 0 {
		queryMods = append(queryMods, qm.Offset(opts.PageSize*(opts.Page-1)))
	}

	// filter type
	if opts.LayerType.Valid {
		queryMods = append(queryMods, boiler.LayerWhere.Type.EQ(opts.LayerType))

	}

	// Build query
	queryMods = append(queryMods,
		qm.Select(
			qm.Rels(boiler.TableNames.Layers, boiler.LayerColumns.ID),
			qm.Rels(boiler.TableNames.Layers, boiler.LayerColumns.ImageURL),
			qm.Rels(boiler.TableNames.Layers, boiler.LayerColumns.Type),
		),
		qm.From(boiler.TableNames.Layers),
	)

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		mc := &Layer{}

		scanArgs := []interface{}{
			&mc.ID,
			&mc.ImageURL,
			&mc.Type,
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, layers, err
		}
		layers = append(layers, mc)
	}

	return total, layers, nil
}
