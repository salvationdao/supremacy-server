package db

import (
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AvatarFeaturesListOpts struct {
	Search   string
	Filter   *ListFilterRequest
	Sort     *ListSortRequest
	PageSize int
	Page     int
}

type AvatarHair struct {
	ImageURL null.String `json:"image_url,omitempty"`

	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

func HairList(opts *AvatarFeaturesListOpts) (int64, []*AvatarHair, error) {
	var hairs []*AvatarHair

	var queryMods []qm.QueryMod

	total, err := boiler.Hairs(
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

	// Build query
	queryMods = append(queryMods,
		qm.Select(
			qm.Rels(boiler.TableNames.Hair, boiler.HairColumns.ImageURL),
		),
		qm.From(boiler.TableNames.Hair),
	)

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		mc := &AvatarHair{}

		scanArgs := []interface{}{
			&mc.ImageURL,
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, hairs, err
		}
		hairs = append(hairs, mc)
	}

	return total, hairs, nil
}
