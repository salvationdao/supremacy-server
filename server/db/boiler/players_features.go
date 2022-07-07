// Code generated by SQLBoiler 4.8.6 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package boiler

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// PlayersFeature is an object representing the database table.
type PlayersFeature struct {
	ID          string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	PlayerID    string    `boiler:"player_id" boil:"player_id" json:"player_id" toml:"player_id" yaml:"player_id"`
	FeatureType string    `boiler:"feature_type" boil:"feature_type" json:"feature_type" toml:"feature_type" yaml:"feature_type"`
	DeletedAt   null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	UpdatedAt   time.Time `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt   time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *playersFeatureR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L playersFeatureL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PlayersFeatureColumns = struct {
	ID          string
	PlayerID    string
	FeatureType string
	DeletedAt   string
	UpdatedAt   string
	CreatedAt   string
}{
	ID:          "id",
	PlayerID:    "player_id",
	FeatureType: "feature_type",
	DeletedAt:   "deleted_at",
	UpdatedAt:   "updated_at",
	CreatedAt:   "created_at",
}

var PlayersFeatureTableColumns = struct {
	ID          string
	PlayerID    string
	FeatureType string
	DeletedAt   string
	UpdatedAt   string
	CreatedAt   string
}{
	ID:          "players_features.id",
	PlayerID:    "players_features.player_id",
	FeatureType: "players_features.feature_type",
	DeletedAt:   "players_features.deleted_at",
	UpdatedAt:   "players_features.updated_at",
	CreatedAt:   "players_features.created_at",
}

// Generated where

var PlayersFeatureWhere = struct {
	ID          whereHelperstring
	PlayerID    whereHelperstring
	FeatureType whereHelperstring
	DeletedAt   whereHelpernull_Time
	UpdatedAt   whereHelpertime_Time
	CreatedAt   whereHelpertime_Time
}{
	ID:          whereHelperstring{field: "\"players_features\".\"id\""},
	PlayerID:    whereHelperstring{field: "\"players_features\".\"player_id\""},
	FeatureType: whereHelperstring{field: "\"players_features\".\"feature_type\""},
	DeletedAt:   whereHelpernull_Time{field: "\"players_features\".\"deleted_at\""},
	UpdatedAt:   whereHelpertime_Time{field: "\"players_features\".\"updated_at\""},
	CreatedAt:   whereHelpertime_Time{field: "\"players_features\".\"created_at\""},
}

// PlayersFeatureRels is where relationship names are stored.
var PlayersFeatureRels = struct {
	FeatureTypeFeature string
	Player             string
}{
	FeatureTypeFeature: "FeatureTypeFeature",
	Player:             "Player",
}

// playersFeatureR is where relationships are stored.
type playersFeatureR struct {
	FeatureTypeFeature *Feature `boiler:"FeatureTypeFeature" boil:"FeatureTypeFeature" json:"FeatureTypeFeature" toml:"FeatureTypeFeature" yaml:"FeatureTypeFeature"`
	Player             *Player  `boiler:"Player" boil:"Player" json:"Player" toml:"Player" yaml:"Player"`
}

// NewStruct creates a new relationship struct
func (*playersFeatureR) NewStruct() *playersFeatureR {
	return &playersFeatureR{}
}

// playersFeatureL is where Load methods for each relationship are stored.
type playersFeatureL struct{}

var (
	playersFeatureAllColumns            = []string{"id", "player_id", "feature_type", "deleted_at", "updated_at", "created_at"}
	playersFeatureColumnsWithoutDefault = []string{"player_id", "feature_type"}
	playersFeatureColumnsWithDefault    = []string{"id", "deleted_at", "updated_at", "created_at"}
	playersFeaturePrimaryKeyColumns     = []string{"id"}
	playersFeatureGeneratedColumns      = []string{}
)

type (
	// PlayersFeatureSlice is an alias for a slice of pointers to PlayersFeature.
	// This should almost always be used instead of []PlayersFeature.
	PlayersFeatureSlice []*PlayersFeature
	// PlayersFeatureHook is the signature for custom PlayersFeature hook methods
	PlayersFeatureHook func(boil.Executor, *PlayersFeature) error

	playersFeatureQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	playersFeatureType                 = reflect.TypeOf(&PlayersFeature{})
	playersFeatureMapping              = queries.MakeStructMapping(playersFeatureType)
	playersFeaturePrimaryKeyMapping, _ = queries.BindMapping(playersFeatureType, playersFeatureMapping, playersFeaturePrimaryKeyColumns)
	playersFeatureInsertCacheMut       sync.RWMutex
	playersFeatureInsertCache          = make(map[string]insertCache)
	playersFeatureUpdateCacheMut       sync.RWMutex
	playersFeatureUpdateCache          = make(map[string]updateCache)
	playersFeatureUpsertCacheMut       sync.RWMutex
	playersFeatureUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var playersFeatureAfterSelectHooks []PlayersFeatureHook

var playersFeatureBeforeInsertHooks []PlayersFeatureHook
var playersFeatureAfterInsertHooks []PlayersFeatureHook

var playersFeatureBeforeUpdateHooks []PlayersFeatureHook
var playersFeatureAfterUpdateHooks []PlayersFeatureHook

var playersFeatureBeforeDeleteHooks []PlayersFeatureHook
var playersFeatureAfterDeleteHooks []PlayersFeatureHook

var playersFeatureBeforeUpsertHooks []PlayersFeatureHook
var playersFeatureAfterUpsertHooks []PlayersFeatureHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *PlayersFeature) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range playersFeatureAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *PlayersFeature) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playersFeatureBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *PlayersFeature) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playersFeatureAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *PlayersFeature) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playersFeatureBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *PlayersFeature) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playersFeatureAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *PlayersFeature) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playersFeatureBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *PlayersFeature) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playersFeatureAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *PlayersFeature) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playersFeatureBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *PlayersFeature) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playersFeatureAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPlayersFeatureHook registers your hook function for all future operations.
func AddPlayersFeatureHook(hookPoint boil.HookPoint, playersFeatureHook PlayersFeatureHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		playersFeatureAfterSelectHooks = append(playersFeatureAfterSelectHooks, playersFeatureHook)
	case boil.BeforeInsertHook:
		playersFeatureBeforeInsertHooks = append(playersFeatureBeforeInsertHooks, playersFeatureHook)
	case boil.AfterInsertHook:
		playersFeatureAfterInsertHooks = append(playersFeatureAfterInsertHooks, playersFeatureHook)
	case boil.BeforeUpdateHook:
		playersFeatureBeforeUpdateHooks = append(playersFeatureBeforeUpdateHooks, playersFeatureHook)
	case boil.AfterUpdateHook:
		playersFeatureAfterUpdateHooks = append(playersFeatureAfterUpdateHooks, playersFeatureHook)
	case boil.BeforeDeleteHook:
		playersFeatureBeforeDeleteHooks = append(playersFeatureBeforeDeleteHooks, playersFeatureHook)
	case boil.AfterDeleteHook:
		playersFeatureAfterDeleteHooks = append(playersFeatureAfterDeleteHooks, playersFeatureHook)
	case boil.BeforeUpsertHook:
		playersFeatureBeforeUpsertHooks = append(playersFeatureBeforeUpsertHooks, playersFeatureHook)
	case boil.AfterUpsertHook:
		playersFeatureAfterUpsertHooks = append(playersFeatureAfterUpsertHooks, playersFeatureHook)
	}
}

// One returns a single playersFeature record from the query.
func (q playersFeatureQuery) One(exec boil.Executor) (*PlayersFeature, error) {
	o := &PlayersFeature{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for players_features")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all PlayersFeature records from the query.
func (q playersFeatureQuery) All(exec boil.Executor) (PlayersFeatureSlice, error) {
	var o []*PlayersFeature

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to PlayersFeature slice")
	}

	if len(playersFeatureAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all PlayersFeature records in the query.
func (q playersFeatureQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count players_features rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q playersFeatureQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if players_features exists")
	}

	return count > 0, nil
}

// FeatureTypeFeature pointed to by the foreign key.
func (o *PlayersFeature) FeatureTypeFeature(mods ...qm.QueryMod) featureQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"type\" = ?", o.FeatureType),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Features(queryMods...)
	queries.SetFrom(query.Query, "\"features\"")

	return query
}

// Player pointed to by the foreign key.
func (o *PlayersFeature) Player(mods ...qm.QueryMod) playerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.PlayerID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Players(queryMods...)
	queries.SetFrom(query.Query, "\"players\"")

	return query
}

// LoadFeatureTypeFeature allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playersFeatureL) LoadFeatureTypeFeature(e boil.Executor, singular bool, maybePlayersFeature interface{}, mods queries.Applicator) error {
	var slice []*PlayersFeature
	var object *PlayersFeature

	if singular {
		object = maybePlayersFeature.(*PlayersFeature)
	} else {
		slice = *maybePlayersFeature.(*[]*PlayersFeature)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playersFeatureR{}
		}
		args = append(args, object.FeatureType)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playersFeatureR{}
			}

			for _, a := range args {
				if a == obj.FeatureType {
					continue Outer
				}
			}

			args = append(args, obj.FeatureType)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`features`),
		qm.WhereIn(`features.type in ?`, args...),
		qmhelper.WhereIsNull(`features.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Feature")
	}

	var resultSlice []*Feature
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Feature")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for features")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for features")
	}

	if len(playersFeatureAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.FeatureTypeFeature = foreign
		if foreign.R == nil {
			foreign.R = &featureR{}
		}
		foreign.R.FeatureTypePlayersFeatures = append(foreign.R.FeatureTypePlayersFeatures, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.FeatureType == foreign.Type {
				local.R.FeatureTypeFeature = foreign
				if foreign.R == nil {
					foreign.R = &featureR{}
				}
				foreign.R.FeatureTypePlayersFeatures = append(foreign.R.FeatureTypePlayersFeatures, local)
				break
			}
		}
	}

	return nil
}

// LoadPlayer allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playersFeatureL) LoadPlayer(e boil.Executor, singular bool, maybePlayersFeature interface{}, mods queries.Applicator) error {
	var slice []*PlayersFeature
	var object *PlayersFeature

	if singular {
		object = maybePlayersFeature.(*PlayersFeature)
	} else {
		slice = *maybePlayersFeature.(*[]*PlayersFeature)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playersFeatureR{}
		}
		args = append(args, object.PlayerID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playersFeatureR{}
			}

			for _, a := range args {
				if a == obj.PlayerID {
					continue Outer
				}
			}

			args = append(args, obj.PlayerID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`players`),
		qm.WhereIn(`players.id in ?`, args...),
		qmhelper.WhereIsNull(`players.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Player")
	}

	var resultSlice []*Player
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Player")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for players")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for players")
	}

	if len(playersFeatureAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Player = foreign
		if foreign.R == nil {
			foreign.R = &playerR{}
		}
		foreign.R.PlayersFeatures = append(foreign.R.PlayersFeatures, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.PlayerID == foreign.ID {
				local.R.Player = foreign
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.PlayersFeatures = append(foreign.R.PlayersFeatures, local)
				break
			}
		}
	}

	return nil
}

// SetFeatureTypeFeature of the playersFeature to the related item.
// Sets o.R.FeatureTypeFeature to related.
// Adds o to related.R.FeatureTypePlayersFeatures.
func (o *PlayersFeature) SetFeatureTypeFeature(exec boil.Executor, insert bool, related *Feature) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"players_features\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"feature_type"}),
		strmangle.WhereClause("\"", "\"", 2, playersFeaturePrimaryKeyColumns),
	)
	values := []interface{}{related.Type, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.FeatureType = related.Type
	if o.R == nil {
		o.R = &playersFeatureR{
			FeatureTypeFeature: related,
		}
	} else {
		o.R.FeatureTypeFeature = related
	}

	if related.R == nil {
		related.R = &featureR{
			FeatureTypePlayersFeatures: PlayersFeatureSlice{o},
		}
	} else {
		related.R.FeatureTypePlayersFeatures = append(related.R.FeatureTypePlayersFeatures, o)
	}

	return nil
}

// SetPlayer of the playersFeature to the related item.
// Sets o.R.Player to related.
// Adds o to related.R.PlayersFeatures.
func (o *PlayersFeature) SetPlayer(exec boil.Executor, insert bool, related *Player) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"players_features\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"player_id"}),
		strmangle.WhereClause("\"", "\"", 2, playersFeaturePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.PlayerID = related.ID
	if o.R == nil {
		o.R = &playersFeatureR{
			Player: related,
		}
	} else {
		o.R.Player = related
	}

	if related.R == nil {
		related.R = &playerR{
			PlayersFeatures: PlayersFeatureSlice{o},
		}
	} else {
		related.R.PlayersFeatures = append(related.R.PlayersFeatures, o)
	}

	return nil
}

// PlayersFeatures retrieves all the records using an executor.
func PlayersFeatures(mods ...qm.QueryMod) playersFeatureQuery {
	mods = append(mods, qm.From("\"players_features\""), qmhelper.WhereIsNull("\"players_features\".\"deleted_at\""))
	return playersFeatureQuery{NewQuery(mods...)}
}

// FindPlayersFeature retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPlayersFeature(exec boil.Executor, iD string, selectCols ...string) (*PlayersFeature, error) {
	playersFeatureObj := &PlayersFeature{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"players_features\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, playersFeatureObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from players_features")
	}

	if err = playersFeatureObj.doAfterSelectHooks(exec); err != nil {
		return playersFeatureObj, err
	}

	return playersFeatureObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *PlayersFeature) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no players_features provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = currTime
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(playersFeatureColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	playersFeatureInsertCacheMut.RLock()
	cache, cached := playersFeatureInsertCache[key]
	playersFeatureInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			playersFeatureAllColumns,
			playersFeatureColumnsWithDefault,
			playersFeatureColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(playersFeatureType, playersFeatureMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(playersFeatureType, playersFeatureMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"players_features\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"players_features\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRow(cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.Exec(cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "boiler: unable to insert into players_features")
	}

	if !cached {
		playersFeatureInsertCacheMut.Lock()
		playersFeatureInsertCache[key] = cache
		playersFeatureInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the PlayersFeature.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *PlayersFeature) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	playersFeatureUpdateCacheMut.RLock()
	cache, cached := playersFeatureUpdateCache[key]
	playersFeatureUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			playersFeatureAllColumns,
			playersFeaturePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update players_features, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"players_features\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, playersFeaturePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(playersFeatureType, playersFeatureMapping, append(wl, playersFeaturePrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	var result sql.Result
	result, err = exec.Exec(cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update players_features row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for players_features")
	}

	if !cached {
		playersFeatureUpdateCacheMut.Lock()
		playersFeatureUpdateCache[key] = cache
		playersFeatureUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q playersFeatureQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for players_features")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for players_features")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PlayersFeatureSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("boiler: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playersFeaturePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"players_features\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, playersFeaturePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in playersFeature slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all playersFeature")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *PlayersFeature) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no players_features provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(playersFeatureColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	playersFeatureUpsertCacheMut.RLock()
	cache, cached := playersFeatureUpsertCache[key]
	playersFeatureUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			playersFeatureAllColumns,
			playersFeatureColumnsWithDefault,
			playersFeatureColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			playersFeatureAllColumns,
			playersFeaturePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert players_features, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(playersFeaturePrimaryKeyColumns))
			copy(conflict, playersFeaturePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"players_features\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(playersFeatureType, playersFeatureMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(playersFeatureType, playersFeatureMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRow(cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.Exec(cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "boiler: unable to upsert players_features")
	}

	if !cached {
		playersFeatureUpsertCacheMut.Lock()
		playersFeatureUpsertCache[key] = cache
		playersFeatureUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single PlayersFeature record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *PlayersFeature) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no PlayersFeature provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), playersFeaturePrimaryKeyMapping)
		sql = "DELETE FROM \"players_features\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"players_features\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(playersFeatureType, playersFeatureMapping, append(wl, playersFeaturePrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), valueMapping)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from players_features")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for players_features")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q playersFeatureQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no playersFeatureQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from players_features")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for players_features")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PlayersFeatureSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(playersFeatureBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playersFeaturePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"players_features\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playersFeaturePrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playersFeaturePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"players_features\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, playersFeaturePrimaryKeyColumns, len(o)),
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		args = append([]interface{}{currTime}, args...)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from playersFeature slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for players_features")
	}

	if len(playersFeatureAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *PlayersFeature) Reload(exec boil.Executor) error {
	ret, err := FindPlayersFeature(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PlayersFeatureSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PlayersFeatureSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playersFeaturePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"players_features\".* FROM \"players_features\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playersFeaturePrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in PlayersFeatureSlice")
	}

	*o = slice

	return nil
}

// PlayersFeatureExists checks if the PlayersFeature row exists.
func PlayersFeatureExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"players_features\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if players_features exists")
	}

	return exists, nil
}
