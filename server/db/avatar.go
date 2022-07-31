package db

import (
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

	// insert
	err := ava.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return err
	}

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

// list
type CustomAvatarsListOpts struct {
	Search    string
	Filter    *ListFilterRequest
	LayerType null.String
	Sort      *ListSortRequest
	PageSize  int
	Page      int
}

type CustomAvatar struct {
	ID   string `json:"id,omitempty"`
	Face Layer  `json:"face,omitempty"`
	Body Layer  `json:"body,omitempty"`
	Hair *Layer `json:"hair,omitempty"`

	Accessory *Layer `json:"accessory,omitempty"`
	EyeWear   *Layer `json:"eye_wear,omitempty"`
	Helmet    *Layer `json:"helmet,omitempty"`

	UpdatedAt time.Time `json:"updated_at,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func CustomAvatarsList(playerID string, opts *CustomAvatarsListOpts) (int64, []string, error) {
	var ids []string

	queryMods := []qm.QueryMod{
		boiler.ProfileCustomAvatarWhere.PlayerID.EQ(playerID),
	}

	total, err := boiler.ProfileCustomAvatars(
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
			qm.Rels(boiler.TableNames.ProfileCustomAvatars, boiler.ProfileCustomAvatarColumns.ID),
		),
		qm.From(boiler.TableNames.ProfileCustomAvatars),
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
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, nil, err
		}
		ids = append(ids, mc.ID)
	}

	return total, ids, nil
}

func GetCustomAvatar(conn boil.Executor, avatarID string) (*CustomAvatar, error) {
	// get avatar
	ava, err := boiler.ProfileCustomAvatars(
		boiler.ProfileCustomAvatarWhere.ID.EQ(avatarID),

		// load layer rels
		qm.Load(boiler.ProfileCustomAvatarRels.Face),
		qm.Load(boiler.ProfileCustomAvatarRels.Body),
		qm.Load(boiler.ProfileCustomAvatarRels.Hair),
		qm.Load(boiler.ProfileCustomAvatarRels.Accessory),
		qm.Load(boiler.ProfileCustomAvatarRels.EyeWear),
		qm.Load(boiler.ProfileCustomAvatarRels.Helmet),
	).One(conn)
	if err != nil {
		return nil, err
	}

	resp := &CustomAvatar{
		ID: ava.ID,
	}

	// set layers

	// face
	if ava.R != nil && ava.R.Face != nil {
		resp.Face = Layer{
			ID:       ava.R.Face.ID,
			ImageURL: null.StringFrom(ava.R.Face.ImageURL),
		}
	}

	// body
	if ava.R != nil && ava.R.Body != nil {
		resp.Body = Layer{
			ID:       ava.R.Body.ID,
			ImageURL: null.StringFrom(ava.R.Body.ImageURL),
		}
	}

	// hair
	if ava.R != nil && ava.R.Hair != nil {
		resp.Hair = &Layer{
			ID:       ava.R.Hair.ID,
			ImageURL: null.StringFrom(ava.R.Hair.ImageURL),
		}
	}

	// accessory
	if ava.R != nil && ava.R.Accessory != nil {
		resp.Accessory = &Layer{
			ID:       ava.R.Accessory.ID,
			ImageURL: null.StringFrom(ava.R.Accessory.ImageURL),
		}
	}

	// eyewear
	if ava.R != nil && ava.R.EyeWear != nil {
		resp.EyeWear = &Layer{
			ID:       ava.R.EyeWear.ID,
			ImageURL: null.StringFrom(ava.R.EyeWear.ImageURL),
		}
	}

	// helmet
	if ava.R != nil && ava.R.Helmet != nil {
		resp.Helmet = &Layer{
			ID:       ava.R.Helmet.ID,
			ImageURL: null.StringFrom(ava.R.Helmet.ImageURL),
		}
	}

	return resp, nil
}
