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

// GameMap is an object representing the database table.
type GameMap struct {
	ID            string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Name          string    `boiler:"name" boil:"name" json:"name" toml:"name" yaml:"name"`
	MaxSpawns     int       `boiler:"max_spawns" boil:"max_spawns" json:"max_spawns" toml:"max_spawns" yaml:"max_spawns"`
	Type          string    `boiler:"type" boil:"type" json:"type" toml:"type" yaml:"type"`
	DisabledAt    null.Time `boiler:"disabled_at" boil:"disabled_at" json:"disabled_at,omitempty" toml:"disabled_at" yaml:"disabled_at,omitempty"`
	BackgroundURL string    `boiler:"background_url" boil:"background_url" json:"background_url" toml:"background_url" yaml:"background_url"`
	LogoURL       string    `boiler:"logo_url" boil:"logo_url" json:"logo_url" toml:"logo_url" yaml:"logo_url"`

	R *gameMapR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L gameMapL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var GameMapColumns = struct {
	ID            string
	Name          string
	MaxSpawns     string
	Type          string
	DisabledAt    string
	BackgroundURL string
	LogoURL       string
}{
	ID:            "id",
	Name:          "name",
	MaxSpawns:     "max_spawns",
	Type:          "type",
	DisabledAt:    "disabled_at",
	BackgroundURL: "background_url",
	LogoURL:       "logo_url",
}

var GameMapTableColumns = struct {
	ID            string
	Name          string
	MaxSpawns     string
	Type          string
	DisabledAt    string
	BackgroundURL string
	LogoURL       string
}{
	ID:            "game_maps.id",
	Name:          "game_maps.name",
	MaxSpawns:     "game_maps.max_spawns",
	Type:          "game_maps.type",
	DisabledAt:    "game_maps.disabled_at",
	BackgroundURL: "game_maps.background_url",
	LogoURL:       "game_maps.logo_url",
}

// Generated where

var GameMapWhere = struct {
	ID            whereHelperstring
	Name          whereHelperstring
	MaxSpawns     whereHelperint
	Type          whereHelperstring
	DisabledAt    whereHelpernull_Time
	BackgroundURL whereHelperstring
	LogoURL       whereHelperstring
}{
	ID:            whereHelperstring{field: "\"game_maps\".\"id\""},
	Name:          whereHelperstring{field: "\"game_maps\".\"name\""},
	MaxSpawns:     whereHelperint{field: "\"game_maps\".\"max_spawns\""},
	Type:          whereHelperstring{field: "\"game_maps\".\"type\""},
	DisabledAt:    whereHelpernull_Time{field: "\"game_maps\".\"disabled_at\""},
	BackgroundURL: whereHelperstring{field: "\"game_maps\".\"background_url\""},
	LogoURL:       whereHelperstring{field: "\"game_maps\".\"logo_url\""},
}

// GameMapRels is where relationship names are stored.
var GameMapRels = struct {
	BattleLobbies         string
	MapBattleMapQueueOlds string
	Battles               string
}{
	BattleLobbies:         "BattleLobbies",
	MapBattleMapQueueOlds: "MapBattleMapQueueOlds",
	Battles:               "Battles",
}

// gameMapR is where relationships are stored.
type gameMapR struct {
	BattleLobbies         BattleLobbySlice       `boiler:"BattleLobbies" boil:"BattleLobbies" json:"BattleLobbies" toml:"BattleLobbies" yaml:"BattleLobbies"`
	MapBattleMapQueueOlds BattleMapQueueOldSlice `boiler:"MapBattleMapQueueOlds" boil:"MapBattleMapQueueOlds" json:"MapBattleMapQueueOlds" toml:"MapBattleMapQueueOlds" yaml:"MapBattleMapQueueOlds"`
	Battles               BattleSlice            `boiler:"Battles" boil:"Battles" json:"Battles" toml:"Battles" yaml:"Battles"`
}

// NewStruct creates a new relationship struct
func (*gameMapR) NewStruct() *gameMapR {
	return &gameMapR{}
}

// gameMapL is where Load methods for each relationship are stored.
type gameMapL struct{}

var (
	gameMapAllColumns            = []string{"id", "name", "max_spawns", "type", "disabled_at", "background_url", "logo_url"}
	gameMapColumnsWithoutDefault = []string{"name"}
	gameMapColumnsWithDefault    = []string{"id", "max_spawns", "type", "disabled_at", "background_url", "logo_url"}
	gameMapPrimaryKeyColumns     = []string{"id"}
	gameMapGeneratedColumns      = []string{}
)

type (
	// GameMapSlice is an alias for a slice of pointers to GameMap.
	// This should almost always be used instead of []GameMap.
	GameMapSlice []*GameMap
	// GameMapHook is the signature for custom GameMap hook methods
	GameMapHook func(boil.Executor, *GameMap) error

	gameMapQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	gameMapType                 = reflect.TypeOf(&GameMap{})
	gameMapMapping              = queries.MakeStructMapping(gameMapType)
	gameMapPrimaryKeyMapping, _ = queries.BindMapping(gameMapType, gameMapMapping, gameMapPrimaryKeyColumns)
	gameMapInsertCacheMut       sync.RWMutex
	gameMapInsertCache          = make(map[string]insertCache)
	gameMapUpdateCacheMut       sync.RWMutex
	gameMapUpdateCache          = make(map[string]updateCache)
	gameMapUpsertCacheMut       sync.RWMutex
	gameMapUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var gameMapAfterSelectHooks []GameMapHook

var gameMapBeforeInsertHooks []GameMapHook
var gameMapAfterInsertHooks []GameMapHook

var gameMapBeforeUpdateHooks []GameMapHook
var gameMapAfterUpdateHooks []GameMapHook

var gameMapBeforeDeleteHooks []GameMapHook
var gameMapAfterDeleteHooks []GameMapHook

var gameMapBeforeUpsertHooks []GameMapHook
var gameMapAfterUpsertHooks []GameMapHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *GameMap) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range gameMapAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *GameMap) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range gameMapBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *GameMap) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range gameMapAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *GameMap) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range gameMapBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *GameMap) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range gameMapAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *GameMap) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range gameMapBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *GameMap) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range gameMapAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *GameMap) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range gameMapBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *GameMap) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range gameMapAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddGameMapHook registers your hook function for all future operations.
func AddGameMapHook(hookPoint boil.HookPoint, gameMapHook GameMapHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		gameMapAfterSelectHooks = append(gameMapAfterSelectHooks, gameMapHook)
	case boil.BeforeInsertHook:
		gameMapBeforeInsertHooks = append(gameMapBeforeInsertHooks, gameMapHook)
	case boil.AfterInsertHook:
		gameMapAfterInsertHooks = append(gameMapAfterInsertHooks, gameMapHook)
	case boil.BeforeUpdateHook:
		gameMapBeforeUpdateHooks = append(gameMapBeforeUpdateHooks, gameMapHook)
	case boil.AfterUpdateHook:
		gameMapAfterUpdateHooks = append(gameMapAfterUpdateHooks, gameMapHook)
	case boil.BeforeDeleteHook:
		gameMapBeforeDeleteHooks = append(gameMapBeforeDeleteHooks, gameMapHook)
	case boil.AfterDeleteHook:
		gameMapAfterDeleteHooks = append(gameMapAfterDeleteHooks, gameMapHook)
	case boil.BeforeUpsertHook:
		gameMapBeforeUpsertHooks = append(gameMapBeforeUpsertHooks, gameMapHook)
	case boil.AfterUpsertHook:
		gameMapAfterUpsertHooks = append(gameMapAfterUpsertHooks, gameMapHook)
	}
}

// One returns a single gameMap record from the query.
func (q gameMapQuery) One(exec boil.Executor) (*GameMap, error) {
	o := &GameMap{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for game_maps")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all GameMap records from the query.
func (q gameMapQuery) All(exec boil.Executor) (GameMapSlice, error) {
	var o []*GameMap

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to GameMap slice")
	}

	if len(gameMapAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all GameMap records in the query.
func (q gameMapQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count game_maps rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q gameMapQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if game_maps exists")
	}

	return count > 0, nil
}

// BattleLobbies retrieves all the battle_lobby's BattleLobbies with an executor.
func (o *GameMap) BattleLobbies(mods ...qm.QueryMod) battleLobbyQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"battle_lobbies\".\"game_map_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"battle_lobbies\".\"deleted_at\""),
	)

	query := BattleLobbies(queryMods...)
	queries.SetFrom(query.Query, "\"battle_lobbies\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"battle_lobbies\".*"})
	}

	return query
}

// MapBattleMapQueueOlds retrieves all the battle_map_queue_old's BattleMapQueueOlds with an executor via map_id column.
func (o *GameMap) MapBattleMapQueueOlds(mods ...qm.QueryMod) battleMapQueueOldQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"battle_map_queue_old\".\"map_id\"=?", o.ID),
	)

	query := BattleMapQueueOlds(queryMods...)
	queries.SetFrom(query.Query, "\"battle_map_queue_old\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"battle_map_queue_old\".*"})
	}

	return query
}

// Battles retrieves all the battle's Battles with an executor.
func (o *GameMap) Battles(mods ...qm.QueryMod) battleQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"battles\".\"game_map_id\"=?", o.ID),
	)

	query := Battles(queryMods...)
	queries.SetFrom(query.Query, "\"battles\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"battles\".*"})
	}

	return query
}

// LoadBattleLobbies allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (gameMapL) LoadBattleLobbies(e boil.Executor, singular bool, maybeGameMap interface{}, mods queries.Applicator) error {
	var slice []*GameMap
	var object *GameMap

	if singular {
		object = maybeGameMap.(*GameMap)
	} else {
		slice = *maybeGameMap.(*[]*GameMap)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &gameMapR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &gameMapR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`battle_lobbies`),
		qm.WhereIn(`battle_lobbies.game_map_id in ?`, args...),
		qmhelper.WhereIsNull(`battle_lobbies.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load battle_lobbies")
	}

	var resultSlice []*BattleLobby
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice battle_lobbies")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on battle_lobbies")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for battle_lobbies")
	}

	if len(battleLobbyAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.BattleLobbies = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &battleLobbyR{}
			}
			foreign.R.GameMap = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.GameMapID {
				local.R.BattleLobbies = append(local.R.BattleLobbies, foreign)
				if foreign.R == nil {
					foreign.R = &battleLobbyR{}
				}
				foreign.R.GameMap = local
				break
			}
		}
	}

	return nil
}

// LoadMapBattleMapQueueOlds allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (gameMapL) LoadMapBattleMapQueueOlds(e boil.Executor, singular bool, maybeGameMap interface{}, mods queries.Applicator) error {
	var slice []*GameMap
	var object *GameMap

	if singular {
		object = maybeGameMap.(*GameMap)
	} else {
		slice = *maybeGameMap.(*[]*GameMap)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &gameMapR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &gameMapR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`battle_map_queue_old`),
		qm.WhereIn(`battle_map_queue_old.map_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load battle_map_queue_old")
	}

	var resultSlice []*BattleMapQueueOld
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice battle_map_queue_old")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on battle_map_queue_old")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for battle_map_queue_old")
	}

	if len(battleMapQueueOldAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.MapBattleMapQueueOlds = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &battleMapQueueOldR{}
			}
			foreign.R.Map = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.MapID {
				local.R.MapBattleMapQueueOlds = append(local.R.MapBattleMapQueueOlds, foreign)
				if foreign.R == nil {
					foreign.R = &battleMapQueueOldR{}
				}
				foreign.R.Map = local
				break
			}
		}
	}

	return nil
}

// LoadBattles allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (gameMapL) LoadBattles(e boil.Executor, singular bool, maybeGameMap interface{}, mods queries.Applicator) error {
	var slice []*GameMap
	var object *GameMap

	if singular {
		object = maybeGameMap.(*GameMap)
	} else {
		slice = *maybeGameMap.(*[]*GameMap)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &gameMapR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &gameMapR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`battles`),
		qm.WhereIn(`battles.game_map_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load battles")
	}

	var resultSlice []*Battle
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice battles")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on battles")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for battles")
	}

	if len(battleAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.Battles = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &battleR{}
			}
			foreign.R.GameMap = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.GameMapID {
				local.R.Battles = append(local.R.Battles, foreign)
				if foreign.R == nil {
					foreign.R = &battleR{}
				}
				foreign.R.GameMap = local
				break
			}
		}
	}

	return nil
}

// AddBattleLobbies adds the given related objects to the existing relationships
// of the game_map, optionally inserting them as new records.
// Appends related to o.R.BattleLobbies.
// Sets related.R.GameMap appropriately.
func (o *GameMap) AddBattleLobbies(exec boil.Executor, insert bool, related ...*BattleLobby) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.GameMapID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"battle_lobbies\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"game_map_id"}),
				strmangle.WhereClause("\"", "\"", 2, battleLobbyPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.GameMapID = o.ID
		}
	}

	if o.R == nil {
		o.R = &gameMapR{
			BattleLobbies: related,
		}
	} else {
		o.R.BattleLobbies = append(o.R.BattleLobbies, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &battleLobbyR{
				GameMap: o,
			}
		} else {
			rel.R.GameMap = o
		}
	}
	return nil
}

// AddMapBattleMapQueueOlds adds the given related objects to the existing relationships
// of the game_map, optionally inserting them as new records.
// Appends related to o.R.MapBattleMapQueueOlds.
// Sets related.R.Map appropriately.
func (o *GameMap) AddMapBattleMapQueueOlds(exec boil.Executor, insert bool, related ...*BattleMapQueueOld) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.MapID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"battle_map_queue_old\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"map_id"}),
				strmangle.WhereClause("\"", "\"", 2, battleMapQueueOldPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.MapID = o.ID
		}
	}

	if o.R == nil {
		o.R = &gameMapR{
			MapBattleMapQueueOlds: related,
		}
	} else {
		o.R.MapBattleMapQueueOlds = append(o.R.MapBattleMapQueueOlds, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &battleMapQueueOldR{
				Map: o,
			}
		} else {
			rel.R.Map = o
		}
	}
	return nil
}

// AddBattles adds the given related objects to the existing relationships
// of the game_map, optionally inserting them as new records.
// Appends related to o.R.Battles.
// Sets related.R.GameMap appropriately.
func (o *GameMap) AddBattles(exec boil.Executor, insert bool, related ...*Battle) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.GameMapID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"battles\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"game_map_id"}),
				strmangle.WhereClause("\"", "\"", 2, battlePrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.GameMapID = o.ID
		}
	}

	if o.R == nil {
		o.R = &gameMapR{
			Battles: related,
		}
	} else {
		o.R.Battles = append(o.R.Battles, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &battleR{
				GameMap: o,
			}
		} else {
			rel.R.GameMap = o
		}
	}
	return nil
}

// GameMaps retrieves all the records using an executor.
func GameMaps(mods ...qm.QueryMod) gameMapQuery {
	mods = append(mods, qm.From("\"game_maps\""))
	return gameMapQuery{NewQuery(mods...)}
}

// FindGameMap retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindGameMap(exec boil.Executor, iD string, selectCols ...string) (*GameMap, error) {
	gameMapObj := &GameMap{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"game_maps\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, gameMapObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from game_maps")
	}

	if err = gameMapObj.doAfterSelectHooks(exec); err != nil {
		return gameMapObj, err
	}

	return gameMapObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *GameMap) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no game_maps provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(gameMapColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	gameMapInsertCacheMut.RLock()
	cache, cached := gameMapInsertCache[key]
	gameMapInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			gameMapAllColumns,
			gameMapColumnsWithDefault,
			gameMapColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(gameMapType, gameMapMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(gameMapType, gameMapMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"game_maps\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"game_maps\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into game_maps")
	}

	if !cached {
		gameMapInsertCacheMut.Lock()
		gameMapInsertCache[key] = cache
		gameMapInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the GameMap.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *GameMap) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	gameMapUpdateCacheMut.RLock()
	cache, cached := gameMapUpdateCache[key]
	gameMapUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			gameMapAllColumns,
			gameMapPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update game_maps, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"game_maps\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, gameMapPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(gameMapType, gameMapMapping, append(wl, gameMapPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update game_maps row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for game_maps")
	}

	if !cached {
		gameMapUpdateCacheMut.Lock()
		gameMapUpdateCache[key] = cache
		gameMapUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q gameMapQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for game_maps")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for game_maps")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o GameMapSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), gameMapPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"game_maps\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, gameMapPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in gameMap slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all gameMap")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *GameMap) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no game_maps provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(gameMapColumnsWithDefault, o)

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

	gameMapUpsertCacheMut.RLock()
	cache, cached := gameMapUpsertCache[key]
	gameMapUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			gameMapAllColumns,
			gameMapColumnsWithDefault,
			gameMapColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			gameMapAllColumns,
			gameMapPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert game_maps, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(gameMapPrimaryKeyColumns))
			copy(conflict, gameMapPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"game_maps\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(gameMapType, gameMapMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(gameMapType, gameMapMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert game_maps")
	}

	if !cached {
		gameMapUpsertCacheMut.Lock()
		gameMapUpsertCache[key] = cache
		gameMapUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single GameMap record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *GameMap) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no GameMap provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), gameMapPrimaryKeyMapping)
	sql := "DELETE FROM \"game_maps\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from game_maps")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for game_maps")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q gameMapQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no gameMapQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from game_maps")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for game_maps")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o GameMapSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(gameMapBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), gameMapPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"game_maps\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, gameMapPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from gameMap slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for game_maps")
	}

	if len(gameMapAfterDeleteHooks) != 0 {
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
func (o *GameMap) Reload(exec boil.Executor) error {
	ret, err := FindGameMap(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *GameMapSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := GameMapSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), gameMapPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"game_maps\".* FROM \"game_maps\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, gameMapPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in GameMapSlice")
	}

	*o = slice

	return nil
}

// GameMapExists checks if the GameMap row exists.
func GameMapExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"game_maps\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if game_maps exists")
	}

	return exists, nil
}
