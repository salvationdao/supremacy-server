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

// PlayerActiveLog is an object representing the database table.
type PlayerActiveLog struct {
	ID         string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	PlayerID   string      `boiler:"player_id" boil:"player_id" json:"player_id" toml:"player_id" yaml:"player_id"`
	FactionID  null.String `boiler:"faction_id" boil:"faction_id" json:"faction_id,omitempty" toml:"faction_id" yaml:"faction_id,omitempty"`
	ActiveAt   time.Time   `boiler:"active_at" boil:"active_at" json:"active_at" toml:"active_at" yaml:"active_at"`
	InactiveAt null.Time   `boiler:"inactive_at" boil:"inactive_at" json:"inactive_at,omitempty" toml:"inactive_at" yaml:"inactive_at,omitempty"`

	R *playerActiveLogR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L playerActiveLogL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PlayerActiveLogColumns = struct {
	ID         string
	PlayerID   string
	FactionID  string
	ActiveAt   string
	InactiveAt string
}{
	ID:         "id",
	PlayerID:   "player_id",
	FactionID:  "faction_id",
	ActiveAt:   "active_at",
	InactiveAt: "inactive_at",
}

var PlayerActiveLogTableColumns = struct {
	ID         string
	PlayerID   string
	FactionID  string
	ActiveAt   string
	InactiveAt string
}{
	ID:         "player_active_logs.id",
	PlayerID:   "player_active_logs.player_id",
	FactionID:  "player_active_logs.faction_id",
	ActiveAt:   "player_active_logs.active_at",
	InactiveAt: "player_active_logs.inactive_at",
}

// Generated where

var PlayerActiveLogWhere = struct {
	ID         whereHelperstring
	PlayerID   whereHelperstring
	FactionID  whereHelpernull_String
	ActiveAt   whereHelpertime_Time
	InactiveAt whereHelpernull_Time
}{
	ID:         whereHelperstring{field: "\"player_active_logs\".\"id\""},
	PlayerID:   whereHelperstring{field: "\"player_active_logs\".\"player_id\""},
	FactionID:  whereHelpernull_String{field: "\"player_active_logs\".\"faction_id\""},
	ActiveAt:   whereHelpertime_Time{field: "\"player_active_logs\".\"active_at\""},
	InactiveAt: whereHelpernull_Time{field: "\"player_active_logs\".\"inactive_at\""},
}

// PlayerActiveLogRels is where relationship names are stored.
var PlayerActiveLogRels = struct {
	Faction string
	Player  string
}{
	Faction: "Faction",
	Player:  "Player",
}

// playerActiveLogR is where relationships are stored.
type playerActiveLogR struct {
	Faction *Faction `boiler:"Faction" boil:"Faction" json:"Faction" toml:"Faction" yaml:"Faction"`
	Player  *Player  `boiler:"Player" boil:"Player" json:"Player" toml:"Player" yaml:"Player"`
}

// NewStruct creates a new relationship struct
func (*playerActiveLogR) NewStruct() *playerActiveLogR {
	return &playerActiveLogR{}
}

// playerActiveLogL is where Load methods for each relationship are stored.
type playerActiveLogL struct{}

var (
	playerActiveLogAllColumns            = []string{"id", "player_id", "faction_id", "active_at", "inactive_at"}
	playerActiveLogColumnsWithoutDefault = []string{"player_id"}
	playerActiveLogColumnsWithDefault    = []string{"id", "faction_id", "active_at", "inactive_at"}
	playerActiveLogPrimaryKeyColumns     = []string{"id"}
	playerActiveLogGeneratedColumns      = []string{}
)

type (
	// PlayerActiveLogSlice is an alias for a slice of pointers to PlayerActiveLog.
	// This should almost always be used instead of []PlayerActiveLog.
	PlayerActiveLogSlice []*PlayerActiveLog
	// PlayerActiveLogHook is the signature for custom PlayerActiveLog hook methods
	PlayerActiveLogHook func(boil.Executor, *PlayerActiveLog) error

	playerActiveLogQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	playerActiveLogType                 = reflect.TypeOf(&PlayerActiveLog{})
	playerActiveLogMapping              = queries.MakeStructMapping(playerActiveLogType)
	playerActiveLogPrimaryKeyMapping, _ = queries.BindMapping(playerActiveLogType, playerActiveLogMapping, playerActiveLogPrimaryKeyColumns)
	playerActiveLogInsertCacheMut       sync.RWMutex
	playerActiveLogInsertCache          = make(map[string]insertCache)
	playerActiveLogUpdateCacheMut       sync.RWMutex
	playerActiveLogUpdateCache          = make(map[string]updateCache)
	playerActiveLogUpsertCacheMut       sync.RWMutex
	playerActiveLogUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var playerActiveLogAfterSelectHooks []PlayerActiveLogHook

var playerActiveLogBeforeInsertHooks []PlayerActiveLogHook
var playerActiveLogAfterInsertHooks []PlayerActiveLogHook

var playerActiveLogBeforeUpdateHooks []PlayerActiveLogHook
var playerActiveLogAfterUpdateHooks []PlayerActiveLogHook

var playerActiveLogBeforeDeleteHooks []PlayerActiveLogHook
var playerActiveLogAfterDeleteHooks []PlayerActiveLogHook

var playerActiveLogBeforeUpsertHooks []PlayerActiveLogHook
var playerActiveLogAfterUpsertHooks []PlayerActiveLogHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *PlayerActiveLog) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range playerActiveLogAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *PlayerActiveLog) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerActiveLogBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *PlayerActiveLog) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerActiveLogAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *PlayerActiveLog) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playerActiveLogBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *PlayerActiveLog) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playerActiveLogAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *PlayerActiveLog) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playerActiveLogBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *PlayerActiveLog) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playerActiveLogAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *PlayerActiveLog) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerActiveLogBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *PlayerActiveLog) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerActiveLogAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPlayerActiveLogHook registers your hook function for all future operations.
func AddPlayerActiveLogHook(hookPoint boil.HookPoint, playerActiveLogHook PlayerActiveLogHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		playerActiveLogAfterSelectHooks = append(playerActiveLogAfterSelectHooks, playerActiveLogHook)
	case boil.BeforeInsertHook:
		playerActiveLogBeforeInsertHooks = append(playerActiveLogBeforeInsertHooks, playerActiveLogHook)
	case boil.AfterInsertHook:
		playerActiveLogAfterInsertHooks = append(playerActiveLogAfterInsertHooks, playerActiveLogHook)
	case boil.BeforeUpdateHook:
		playerActiveLogBeforeUpdateHooks = append(playerActiveLogBeforeUpdateHooks, playerActiveLogHook)
	case boil.AfterUpdateHook:
		playerActiveLogAfterUpdateHooks = append(playerActiveLogAfterUpdateHooks, playerActiveLogHook)
	case boil.BeforeDeleteHook:
		playerActiveLogBeforeDeleteHooks = append(playerActiveLogBeforeDeleteHooks, playerActiveLogHook)
	case boil.AfterDeleteHook:
		playerActiveLogAfterDeleteHooks = append(playerActiveLogAfterDeleteHooks, playerActiveLogHook)
	case boil.BeforeUpsertHook:
		playerActiveLogBeforeUpsertHooks = append(playerActiveLogBeforeUpsertHooks, playerActiveLogHook)
	case boil.AfterUpsertHook:
		playerActiveLogAfterUpsertHooks = append(playerActiveLogAfterUpsertHooks, playerActiveLogHook)
	}
}

// One returns a single playerActiveLog record from the query.
func (q playerActiveLogQuery) One(exec boil.Executor) (*PlayerActiveLog, error) {
	o := &PlayerActiveLog{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for player_active_logs")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all PlayerActiveLog records from the query.
func (q playerActiveLogQuery) All(exec boil.Executor) (PlayerActiveLogSlice, error) {
	var o []*PlayerActiveLog

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to PlayerActiveLog slice")
	}

	if len(playerActiveLogAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all PlayerActiveLog records in the query.
func (q playerActiveLogQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count player_active_logs rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q playerActiveLogQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if player_active_logs exists")
	}

	return count > 0, nil
}

// Faction pointed to by the foreign key.
func (o *PlayerActiveLog) Faction(mods ...qm.QueryMod) factionQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.FactionID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Factions(queryMods...)
	queries.SetFrom(query.Query, "\"factions\"")

	return query
}

// Player pointed to by the foreign key.
func (o *PlayerActiveLog) Player(mods ...qm.QueryMod) playerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.PlayerID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Players(queryMods...)
	queries.SetFrom(query.Query, "\"players\"")

	return query
}

// LoadFaction allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playerActiveLogL) LoadFaction(e boil.Executor, singular bool, maybePlayerActiveLog interface{}, mods queries.Applicator) error {
	var slice []*PlayerActiveLog
	var object *PlayerActiveLog

	if singular {
		object = maybePlayerActiveLog.(*PlayerActiveLog)
	} else {
		slice = *maybePlayerActiveLog.(*[]*PlayerActiveLog)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playerActiveLogR{}
		}
		if !queries.IsNil(object.FactionID) {
			args = append(args, object.FactionID)
		}

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playerActiveLogR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.FactionID) {
					continue Outer
				}
			}

			if !queries.IsNil(obj.FactionID) {
				args = append(args, obj.FactionID)
			}

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`factions`),
		qm.WhereIn(`factions.id in ?`, args...),
		qmhelper.WhereIsNull(`factions.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Faction")
	}

	var resultSlice []*Faction
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Faction")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for factions")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for factions")
	}

	if len(playerActiveLogAfterSelectHooks) != 0 {
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
		object.R.Faction = foreign
		if foreign.R == nil {
			foreign.R = &factionR{}
		}
		foreign.R.PlayerActiveLogs = append(foreign.R.PlayerActiveLogs, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.FactionID, foreign.ID) {
				local.R.Faction = foreign
				if foreign.R == nil {
					foreign.R = &factionR{}
				}
				foreign.R.PlayerActiveLogs = append(foreign.R.PlayerActiveLogs, local)
				break
			}
		}
	}

	return nil
}

// LoadPlayer allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playerActiveLogL) LoadPlayer(e boil.Executor, singular bool, maybePlayerActiveLog interface{}, mods queries.Applicator) error {
	var slice []*PlayerActiveLog
	var object *PlayerActiveLog

	if singular {
		object = maybePlayerActiveLog.(*PlayerActiveLog)
	} else {
		slice = *maybePlayerActiveLog.(*[]*PlayerActiveLog)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playerActiveLogR{}
		}
		args = append(args, object.PlayerID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playerActiveLogR{}
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

	if len(playerActiveLogAfterSelectHooks) != 0 {
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
		foreign.R.PlayerActiveLogs = append(foreign.R.PlayerActiveLogs, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.PlayerID == foreign.ID {
				local.R.Player = foreign
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.PlayerActiveLogs = append(foreign.R.PlayerActiveLogs, local)
				break
			}
		}
	}

	return nil
}

// SetFaction of the playerActiveLog to the related item.
// Sets o.R.Faction to related.
// Adds o to related.R.PlayerActiveLogs.
func (o *PlayerActiveLog) SetFaction(exec boil.Executor, insert bool, related *Faction) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"player_active_logs\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"faction_id"}),
		strmangle.WhereClause("\"", "\"", 2, playerActiveLogPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	queries.Assign(&o.FactionID, related.ID)
	if o.R == nil {
		o.R = &playerActiveLogR{
			Faction: related,
		}
	} else {
		o.R.Faction = related
	}

	if related.R == nil {
		related.R = &factionR{
			PlayerActiveLogs: PlayerActiveLogSlice{o},
		}
	} else {
		related.R.PlayerActiveLogs = append(related.R.PlayerActiveLogs, o)
	}

	return nil
}

// RemoveFaction relationship.
// Sets o.R.Faction to nil.
// Removes o from all passed in related items' relationships struct (Optional).
func (o *PlayerActiveLog) RemoveFaction(exec boil.Executor, related *Faction) error {
	var err error

	queries.SetScanner(&o.FactionID, nil)
	if _, err = o.Update(exec, boil.Whitelist("faction_id")); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	if o.R != nil {
		o.R.Faction = nil
	}
	if related == nil || related.R == nil {
		return nil
	}

	for i, ri := range related.R.PlayerActiveLogs {
		if queries.Equal(o.FactionID, ri.FactionID) {
			continue
		}

		ln := len(related.R.PlayerActiveLogs)
		if ln > 1 && i < ln-1 {
			related.R.PlayerActiveLogs[i] = related.R.PlayerActiveLogs[ln-1]
		}
		related.R.PlayerActiveLogs = related.R.PlayerActiveLogs[:ln-1]
		break
	}
	return nil
}

// SetPlayer of the playerActiveLog to the related item.
// Sets o.R.Player to related.
// Adds o to related.R.PlayerActiveLogs.
func (o *PlayerActiveLog) SetPlayer(exec boil.Executor, insert bool, related *Player) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"player_active_logs\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"player_id"}),
		strmangle.WhereClause("\"", "\"", 2, playerActiveLogPrimaryKeyColumns),
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
		o.R = &playerActiveLogR{
			Player: related,
		}
	} else {
		o.R.Player = related
	}

	if related.R == nil {
		related.R = &playerR{
			PlayerActiveLogs: PlayerActiveLogSlice{o},
		}
	} else {
		related.R.PlayerActiveLogs = append(related.R.PlayerActiveLogs, o)
	}

	return nil
}

// PlayerActiveLogs retrieves all the records using an executor.
func PlayerActiveLogs(mods ...qm.QueryMod) playerActiveLogQuery {
	mods = append(mods, qm.From("\"player_active_logs\""))
	return playerActiveLogQuery{NewQuery(mods...)}
}

// FindPlayerActiveLog retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPlayerActiveLog(exec boil.Executor, iD string, selectCols ...string) (*PlayerActiveLog, error) {
	playerActiveLogObj := &PlayerActiveLog{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"player_active_logs\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, playerActiveLogObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from player_active_logs")
	}

	if err = playerActiveLogObj.doAfterSelectHooks(exec); err != nil {
		return playerActiveLogObj, err
	}

	return playerActiveLogObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *PlayerActiveLog) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no player_active_logs provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(playerActiveLogColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	playerActiveLogInsertCacheMut.RLock()
	cache, cached := playerActiveLogInsertCache[key]
	playerActiveLogInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			playerActiveLogAllColumns,
			playerActiveLogColumnsWithDefault,
			playerActiveLogColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(playerActiveLogType, playerActiveLogMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(playerActiveLogType, playerActiveLogMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"player_active_logs\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"player_active_logs\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into player_active_logs")
	}

	if !cached {
		playerActiveLogInsertCacheMut.Lock()
		playerActiveLogInsertCache[key] = cache
		playerActiveLogInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the PlayerActiveLog.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *PlayerActiveLog) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	playerActiveLogUpdateCacheMut.RLock()
	cache, cached := playerActiveLogUpdateCache[key]
	playerActiveLogUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			playerActiveLogAllColumns,
			playerActiveLogPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update player_active_logs, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"player_active_logs\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, playerActiveLogPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(playerActiveLogType, playerActiveLogMapping, append(wl, playerActiveLogPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update player_active_logs row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for player_active_logs")
	}

	if !cached {
		playerActiveLogUpdateCacheMut.Lock()
		playerActiveLogUpdateCache[key] = cache
		playerActiveLogUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q playerActiveLogQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for player_active_logs")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for player_active_logs")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PlayerActiveLogSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerActiveLogPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"player_active_logs\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, playerActiveLogPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in playerActiveLog slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all playerActiveLog")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *PlayerActiveLog) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no player_active_logs provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(playerActiveLogColumnsWithDefault, o)

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

	playerActiveLogUpsertCacheMut.RLock()
	cache, cached := playerActiveLogUpsertCache[key]
	playerActiveLogUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			playerActiveLogAllColumns,
			playerActiveLogColumnsWithDefault,
			playerActiveLogColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			playerActiveLogAllColumns,
			playerActiveLogPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert player_active_logs, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(playerActiveLogPrimaryKeyColumns))
			copy(conflict, playerActiveLogPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"player_active_logs\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(playerActiveLogType, playerActiveLogMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(playerActiveLogType, playerActiveLogMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert player_active_logs")
	}

	if !cached {
		playerActiveLogUpsertCacheMut.Lock()
		playerActiveLogUpsertCache[key] = cache
		playerActiveLogUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single PlayerActiveLog record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *PlayerActiveLog) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no PlayerActiveLog provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), playerActiveLogPrimaryKeyMapping)
	sql := "DELETE FROM \"player_active_logs\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from player_active_logs")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for player_active_logs")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q playerActiveLogQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no playerActiveLogQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from player_active_logs")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for player_active_logs")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PlayerActiveLogSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(playerActiveLogBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerActiveLogPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"player_active_logs\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playerActiveLogPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from playerActiveLog slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for player_active_logs")
	}

	if len(playerActiveLogAfterDeleteHooks) != 0 {
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
func (o *PlayerActiveLog) Reload(exec boil.Executor) error {
	ret, err := FindPlayerActiveLog(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PlayerActiveLogSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PlayerActiveLogSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerActiveLogPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"player_active_logs\".* FROM \"player_active_logs\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playerActiveLogPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in PlayerActiveLogSlice")
	}

	*o = slice

	return nil
}

// PlayerActiveLogExists checks if the PlayerActiveLog row exists.
func PlayerActiveLogExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"player_active_logs\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if player_active_logs exists")
	}

	return exists, nil
}
