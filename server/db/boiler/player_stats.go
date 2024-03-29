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
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// PlayerStat is an object representing the database table.
type PlayerStat struct {
	ID                    string `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	ViewBattleCount       int    `boiler:"view_battle_count" boil:"view_battle_count" json:"view_battle_count" toml:"view_battle_count" yaml:"view_battle_count"`
	AbilityKillCount      int    `boiler:"ability_kill_count" boil:"ability_kill_count" json:"ability_kill_count" toml:"ability_kill_count" yaml:"ability_kill_count"`
	TotalAbilityTriggered int    `boiler:"total_ability_triggered" boil:"total_ability_triggered" json:"total_ability_triggered" toml:"total_ability_triggered" yaml:"total_ability_triggered"`
	MechKillCount         int    `boiler:"mech_kill_count" boil:"mech_kill_count" json:"mech_kill_count" toml:"mech_kill_count" yaml:"mech_kill_count"`

	R *playerStatR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L playerStatL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PlayerStatColumns = struct {
	ID                    string
	ViewBattleCount       string
	AbilityKillCount      string
	TotalAbilityTriggered string
	MechKillCount         string
}{
	ID:                    "id",
	ViewBattleCount:       "view_battle_count",
	AbilityKillCount:      "ability_kill_count",
	TotalAbilityTriggered: "total_ability_triggered",
	MechKillCount:         "mech_kill_count",
}

var PlayerStatTableColumns = struct {
	ID                    string
	ViewBattleCount       string
	AbilityKillCount      string
	TotalAbilityTriggered string
	MechKillCount         string
}{
	ID:                    "player_stats.id",
	ViewBattleCount:       "player_stats.view_battle_count",
	AbilityKillCount:      "player_stats.ability_kill_count",
	TotalAbilityTriggered: "player_stats.total_ability_triggered",
	MechKillCount:         "player_stats.mech_kill_count",
}

// Generated where

var PlayerStatWhere = struct {
	ID                    whereHelperstring
	ViewBattleCount       whereHelperint
	AbilityKillCount      whereHelperint
	TotalAbilityTriggered whereHelperint
	MechKillCount         whereHelperint
}{
	ID:                    whereHelperstring{field: "\"player_stats\".\"id\""},
	ViewBattleCount:       whereHelperint{field: "\"player_stats\".\"view_battle_count\""},
	AbilityKillCount:      whereHelperint{field: "\"player_stats\".\"ability_kill_count\""},
	TotalAbilityTriggered: whereHelperint{field: "\"player_stats\".\"total_ability_triggered\""},
	MechKillCount:         whereHelperint{field: "\"player_stats\".\"mech_kill_count\""},
}

// PlayerStatRels is where relationship names are stored.
var PlayerStatRels = struct {
	IDPlayer string
}{
	IDPlayer: "IDPlayer",
}

// playerStatR is where relationships are stored.
type playerStatR struct {
	IDPlayer *Player `boiler:"IDPlayer" boil:"IDPlayer" json:"IDPlayer" toml:"IDPlayer" yaml:"IDPlayer"`
}

// NewStruct creates a new relationship struct
func (*playerStatR) NewStruct() *playerStatR {
	return &playerStatR{}
}

// playerStatL is where Load methods for each relationship are stored.
type playerStatL struct{}

var (
	playerStatAllColumns            = []string{"id", "view_battle_count", "ability_kill_count", "total_ability_triggered", "mech_kill_count"}
	playerStatColumnsWithoutDefault = []string{"id"}
	playerStatColumnsWithDefault    = []string{"view_battle_count", "ability_kill_count", "total_ability_triggered", "mech_kill_count"}
	playerStatPrimaryKeyColumns     = []string{"id"}
	playerStatGeneratedColumns      = []string{}
)

type (
	// PlayerStatSlice is an alias for a slice of pointers to PlayerStat.
	// This should almost always be used instead of []PlayerStat.
	PlayerStatSlice []*PlayerStat
	// PlayerStatHook is the signature for custom PlayerStat hook methods
	PlayerStatHook func(boil.Executor, *PlayerStat) error

	playerStatQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	playerStatType                 = reflect.TypeOf(&PlayerStat{})
	playerStatMapping              = queries.MakeStructMapping(playerStatType)
	playerStatPrimaryKeyMapping, _ = queries.BindMapping(playerStatType, playerStatMapping, playerStatPrimaryKeyColumns)
	playerStatInsertCacheMut       sync.RWMutex
	playerStatInsertCache          = make(map[string]insertCache)
	playerStatUpdateCacheMut       sync.RWMutex
	playerStatUpdateCache          = make(map[string]updateCache)
	playerStatUpsertCacheMut       sync.RWMutex
	playerStatUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var playerStatAfterSelectHooks []PlayerStatHook

var playerStatBeforeInsertHooks []PlayerStatHook
var playerStatAfterInsertHooks []PlayerStatHook

var playerStatBeforeUpdateHooks []PlayerStatHook
var playerStatAfterUpdateHooks []PlayerStatHook

var playerStatBeforeDeleteHooks []PlayerStatHook
var playerStatAfterDeleteHooks []PlayerStatHook

var playerStatBeforeUpsertHooks []PlayerStatHook
var playerStatAfterUpsertHooks []PlayerStatHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *PlayerStat) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range playerStatAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *PlayerStat) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerStatBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *PlayerStat) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerStatAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *PlayerStat) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playerStatBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *PlayerStat) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playerStatAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *PlayerStat) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playerStatBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *PlayerStat) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playerStatAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *PlayerStat) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerStatBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *PlayerStat) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerStatAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPlayerStatHook registers your hook function for all future operations.
func AddPlayerStatHook(hookPoint boil.HookPoint, playerStatHook PlayerStatHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		playerStatAfterSelectHooks = append(playerStatAfterSelectHooks, playerStatHook)
	case boil.BeforeInsertHook:
		playerStatBeforeInsertHooks = append(playerStatBeforeInsertHooks, playerStatHook)
	case boil.AfterInsertHook:
		playerStatAfterInsertHooks = append(playerStatAfterInsertHooks, playerStatHook)
	case boil.BeforeUpdateHook:
		playerStatBeforeUpdateHooks = append(playerStatBeforeUpdateHooks, playerStatHook)
	case boil.AfterUpdateHook:
		playerStatAfterUpdateHooks = append(playerStatAfterUpdateHooks, playerStatHook)
	case boil.BeforeDeleteHook:
		playerStatBeforeDeleteHooks = append(playerStatBeforeDeleteHooks, playerStatHook)
	case boil.AfterDeleteHook:
		playerStatAfterDeleteHooks = append(playerStatAfterDeleteHooks, playerStatHook)
	case boil.BeforeUpsertHook:
		playerStatBeforeUpsertHooks = append(playerStatBeforeUpsertHooks, playerStatHook)
	case boil.AfterUpsertHook:
		playerStatAfterUpsertHooks = append(playerStatAfterUpsertHooks, playerStatHook)
	}
}

// One returns a single playerStat record from the query.
func (q playerStatQuery) One(exec boil.Executor) (*PlayerStat, error) {
	o := &PlayerStat{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for player_stats")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all PlayerStat records from the query.
func (q playerStatQuery) All(exec boil.Executor) (PlayerStatSlice, error) {
	var o []*PlayerStat

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to PlayerStat slice")
	}

	if len(playerStatAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all PlayerStat records in the query.
func (q playerStatQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count player_stats rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q playerStatQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if player_stats exists")
	}

	return count > 0, nil
}

// IDPlayer pointed to by the foreign key.
func (o *PlayerStat) IDPlayer(mods ...qm.QueryMod) playerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.ID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Players(queryMods...)
	queries.SetFrom(query.Query, "\"players\"")

	return query
}

// LoadIDPlayer allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playerStatL) LoadIDPlayer(e boil.Executor, singular bool, maybePlayerStat interface{}, mods queries.Applicator) error {
	var slice []*PlayerStat
	var object *PlayerStat

	if singular {
		object = maybePlayerStat.(*PlayerStat)
	} else {
		slice = *maybePlayerStat.(*[]*PlayerStat)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playerStatR{}
		}
		args = append(args, object.ID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playerStatR{}
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

	if len(playerStatAfterSelectHooks) != 0 {
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
		object.R.IDPlayer = foreign
		if foreign.R == nil {
			foreign.R = &playerR{}
		}
		foreign.R.IDPlayerStat = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.ID == foreign.ID {
				local.R.IDPlayer = foreign
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.IDPlayerStat = local
				break
			}
		}
	}

	return nil
}

// SetIDPlayer of the playerStat to the related item.
// Sets o.R.IDPlayer to related.
// Adds o to related.R.IDPlayerStat.
func (o *PlayerStat) SetIDPlayer(exec boil.Executor, insert bool, related *Player) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"player_stats\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"id"}),
		strmangle.WhereClause("\"", "\"", 2, playerStatPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.ID = related.ID
	if o.R == nil {
		o.R = &playerStatR{
			IDPlayer: related,
		}
	} else {
		o.R.IDPlayer = related
	}

	if related.R == nil {
		related.R = &playerR{
			IDPlayerStat: o,
		}
	} else {
		related.R.IDPlayerStat = o
	}

	return nil
}

// PlayerStats retrieves all the records using an executor.
func PlayerStats(mods ...qm.QueryMod) playerStatQuery {
	mods = append(mods, qm.From("\"player_stats\""))
	return playerStatQuery{NewQuery(mods...)}
}

// FindPlayerStat retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPlayerStat(exec boil.Executor, iD string, selectCols ...string) (*PlayerStat, error) {
	playerStatObj := &PlayerStat{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"player_stats\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, playerStatObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from player_stats")
	}

	if err = playerStatObj.doAfterSelectHooks(exec); err != nil {
		return playerStatObj, err
	}

	return playerStatObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *PlayerStat) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no player_stats provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(playerStatColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	playerStatInsertCacheMut.RLock()
	cache, cached := playerStatInsertCache[key]
	playerStatInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			playerStatAllColumns,
			playerStatColumnsWithDefault,
			playerStatColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(playerStatType, playerStatMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(playerStatType, playerStatMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"player_stats\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"player_stats\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into player_stats")
	}

	if !cached {
		playerStatInsertCacheMut.Lock()
		playerStatInsertCache[key] = cache
		playerStatInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the PlayerStat.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *PlayerStat) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	playerStatUpdateCacheMut.RLock()
	cache, cached := playerStatUpdateCache[key]
	playerStatUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			playerStatAllColumns,
			playerStatPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update player_stats, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"player_stats\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, playerStatPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(playerStatType, playerStatMapping, append(wl, playerStatPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update player_stats row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for player_stats")
	}

	if !cached {
		playerStatUpdateCacheMut.Lock()
		playerStatUpdateCache[key] = cache
		playerStatUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q playerStatQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for player_stats")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for player_stats")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PlayerStatSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerStatPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"player_stats\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, playerStatPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in playerStat slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all playerStat")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *PlayerStat) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no player_stats provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(playerStatColumnsWithDefault, o)

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

	playerStatUpsertCacheMut.RLock()
	cache, cached := playerStatUpsertCache[key]
	playerStatUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			playerStatAllColumns,
			playerStatColumnsWithDefault,
			playerStatColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			playerStatAllColumns,
			playerStatPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert player_stats, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(playerStatPrimaryKeyColumns))
			copy(conflict, playerStatPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"player_stats\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(playerStatType, playerStatMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(playerStatType, playerStatMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert player_stats")
	}

	if !cached {
		playerStatUpsertCacheMut.Lock()
		playerStatUpsertCache[key] = cache
		playerStatUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single PlayerStat record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *PlayerStat) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no PlayerStat provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), playerStatPrimaryKeyMapping)
	sql := "DELETE FROM \"player_stats\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from player_stats")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for player_stats")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q playerStatQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no playerStatQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from player_stats")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for player_stats")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PlayerStatSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(playerStatBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerStatPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"player_stats\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playerStatPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from playerStat slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for player_stats")
	}

	if len(playerStatAfterDeleteHooks) != 0 {
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
func (o *PlayerStat) Reload(exec boil.Executor) error {
	ret, err := FindPlayerStat(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PlayerStatSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PlayerStatSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerStatPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"player_stats\".* FROM \"player_stats\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playerStatPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in PlayerStatSlice")
	}

	*o = slice

	return nil
}

// PlayerStatExists checks if the PlayerStat row exists.
func PlayerStatExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"player_stats\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if player_stats exists")
	}

	return exists, nil
}
