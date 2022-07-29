package db

import (
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

// create
type AvatarCreateRequest struct {
	FaceID      string
	BodyID      string
	HairID      null.String
	AccessoryID null.String
	EyeWearID   null.String
}

func CustomAvatarCreate(playerID string, req AvatarCreateRequest) error {

	ava := boiler.ProfileCustomAvatar{
		PlayerID:    playerID,
		FaceID:      req.FaceID,
		BodyID:      null.StringFrom(req.BodyID),
		HairID:      req.HairID,
		AccessoryID: req.AccessoryID,
		EyeWearID:   req.EyeWearID,
	}

	fmt.Println()
	fmt.Println()
	fmt.Println("hree44")
	fmt.Println()
	fmt.Println()
	fmt.Println()
	// insert
	err := ava.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println()
	fmt.Println("hree55")
	fmt.Println()
	fmt.Println()
	fmt.Println()

	return nil
}

// update
func CustomAvatarUpdate(id string, req AvatarCreateRequest) error {

	// get custom avatar
	ca, err := boiler.FindProfileCustomAvatar(gamedb.StdConn, id)
	if err != nil {
		return err
	}

	// update
	ca.FaceID = req.FaceID
	ca.BodyID = null.StringFrom(req.BodyID)
	ca.HairID = req.HairID
	ca.AccessoryID = req.AccessoryID
	ca.EyeWearID = req.EyeWearID
	ca.UpdatedAt = time.Now()
	_, err = ca.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}

	return nil
}
