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

// UserStat is an object representing the database table.
type UserStat struct {
	ID                    string `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	ViewBattleCount       int    `boiler:"view_battle_count" boil:"view_battle_count" json:"view_battle_count" toml:"view_battle_count" yaml:"view_battle_count"`
	KillCount             int    `boiler:"kill_count" boil:"kill_count" json:"kill_count" toml:"kill_count" yaml:"kill_count"`
	TotalAbilityTriggered int    `boiler:"total_ability_triggered" boil:"total_ability_triggered" json:"total_ability_triggered" toml:"total_ability_triggered" yaml:"total_ability_triggered"`
	MechKillCount         int    `boiler:"mech_kill_count" boil:"mech_kill_count" json:"mech_kill_count" toml:"mech_kill_count" yaml:"mech_kill_count"`

	R *userStatR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L userStatL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var UserStatColumns = struct {
	ID                    string
	ViewBattleCount       string
	KillCount             string
	TotalAbilityTriggered string
	MechKillCount         string
}{
	ID:                    "id",
	ViewBattleCount:       "view_battle_count",
	KillCount:             "kill_count",
	TotalAbilityTriggered: "total_ability_triggered",
	MechKillCount:         "mech_kill_count",
}

var UserStatTableColumns = struct {
	ID                    string
	ViewBattleCount       string
	KillCount             string
	TotalAbilityTriggered string
	MechKillCount         string
}{
	ID:                    "user_stats.id",
	ViewBattleCount:       "user_stats.view_battle_count",
	KillCount:             "user_stats.kill_count",
	TotalAbilityTriggered: "user_stats.total_ability_triggered",
	MechKillCount:         "user_stats.mech_kill_count",
}

// Generated where

var UserStatWhere = struct {
	ID                    whereHelperstring
	ViewBattleCount       whereHelperint
	KillCount             whereHelperint
	TotalAbilityTriggered whereHelperint
	MechKillCount         whereHelperint
}{
	ID:                    whereHelperstring{field: "\"user_stats\".\"id\""},
	ViewBattleCount:       whereHelperint{field: "\"user_stats\".\"view_battle_count\""},
	KillCount:             whereHelperint{field: "\"user_stats\".\"kill_count\""},
	TotalAbilityTriggered: whereHelperint{field: "\"user_stats\".\"total_ability_triggered\""},
	MechKillCount:         whereHelperint{field: "\"user_stats\".\"mech_kill_count\""},
}

// UserStatRels is where relationship names are stored.
var UserStatRels = struct {
	IDPlayer string
}{
	IDPlayer: "IDPlayer",
}

// userStatR is where relationships are stored.
type userStatR struct {
	IDPlayer *Player `boiler:"IDPlayer" boil:"IDPlayer" json:"IDPlayer" toml:"IDPlayer" yaml:"IDPlayer"`
}

// NewStruct creates a new relationship struct
func (*userStatR) NewStruct() *userStatR {
	return &userStatR{}
}

// userStatL is where Load methods for each relationship are stored.
type userStatL struct{}

var (
	userStatAllColumns            = []string{"id", "view_battle_count", "kill_count", "total_ability_triggered", "mech_kill_count"}
	userStatColumnsWithoutDefault = []string{"id"}
	userStatColumnsWithDefault    = []string{"view_battle_count", "kill_count", "total_ability_triggered", "mech_kill_count"}
	userStatPrimaryKeyColumns     = []string{"id"}
	userStatGeneratedColumns      = []string{}
)

type (
	// UserStatSlice is an alias for a slice of pointers to UserStat.
	// This should almost always be used instead of []UserStat.
	UserStatSlice []*UserStat
	// UserStatHook is the signature for custom UserStat hook methods
	UserStatHook func(boil.Executor, *UserStat) error

	userStatQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	userStatType                 = reflect.TypeOf(&UserStat{})
	userStatMapping              = queries.MakeStructMapping(userStatType)
	userStatPrimaryKeyMapping, _ = queries.BindMapping(userStatType, userStatMapping, userStatPrimaryKeyColumns)
	userStatInsertCacheMut       sync.RWMutex
	userStatInsertCache          = make(map[string]insertCache)
	userStatUpdateCacheMut       sync.RWMutex
	userStatUpdateCache          = make(map[string]updateCache)
	userStatUpsertCacheMut       sync.RWMutex
	userStatUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var userStatAfterSelectHooks []UserStatHook

var userStatBeforeInsertHooks []UserStatHook
var userStatAfterInsertHooks []UserStatHook

var userStatBeforeUpdateHooks []UserStatHook
var userStatAfterUpdateHooks []UserStatHook

var userStatBeforeDeleteHooks []UserStatHook
var userStatAfterDeleteHooks []UserStatHook

var userStatBeforeUpsertHooks []UserStatHook
var userStatAfterUpsertHooks []UserStatHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *UserStat) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range userStatAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *UserStat) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userStatBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *UserStat) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userStatAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *UserStat) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range userStatBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *UserStat) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range userStatAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *UserStat) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range userStatBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *UserStat) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range userStatAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *UserStat) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userStatBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *UserStat) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userStatAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddUserStatHook registers your hook function for all future operations.
func AddUserStatHook(hookPoint boil.HookPoint, userStatHook UserStatHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		userStatAfterSelectHooks = append(userStatAfterSelectHooks, userStatHook)
	case boil.BeforeInsertHook:
		userStatBeforeInsertHooks = append(userStatBeforeInsertHooks, userStatHook)
	case boil.AfterInsertHook:
		userStatAfterInsertHooks = append(userStatAfterInsertHooks, userStatHook)
	case boil.BeforeUpdateHook:
		userStatBeforeUpdateHooks = append(userStatBeforeUpdateHooks, userStatHook)
	case boil.AfterUpdateHook:
		userStatAfterUpdateHooks = append(userStatAfterUpdateHooks, userStatHook)
	case boil.BeforeDeleteHook:
		userStatBeforeDeleteHooks = append(userStatBeforeDeleteHooks, userStatHook)
	case boil.AfterDeleteHook:
		userStatAfterDeleteHooks = append(userStatAfterDeleteHooks, userStatHook)
	case boil.BeforeUpsertHook:
		userStatBeforeUpsertHooks = append(userStatBeforeUpsertHooks, userStatHook)
	case boil.AfterUpsertHook:
		userStatAfterUpsertHooks = append(userStatAfterUpsertHooks, userStatHook)
	}
}

// One returns a single userStat record from the query.
func (q userStatQuery) One(exec boil.Executor) (*UserStat, error) {
	o := &UserStat{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for user_stats")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all UserStat records from the query.
func (q userStatQuery) All(exec boil.Executor) (UserStatSlice, error) {
	var o []*UserStat

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to UserStat slice")
	}

	if len(userStatAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all UserStat records in the query.
func (q userStatQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count user_stats rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q userStatQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if user_stats exists")
	}

	return count > 0, nil
}

// IDPlayer pointed to by the foreign key.
func (o *UserStat) IDPlayer(mods ...qm.QueryMod) playerQuery {
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
func (userStatL) LoadIDPlayer(e boil.Executor, singular bool, maybeUserStat interface{}, mods queries.Applicator) error {
	var slice []*UserStat
	var object *UserStat

	if singular {
		object = maybeUserStat.(*UserStat)
	} else {
		slice = *maybeUserStat.(*[]*UserStat)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &userStatR{}
		}
		args = append(args, object.ID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &userStatR{}
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

	if len(userStatAfterSelectHooks) != 0 {
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
		foreign.R.IDUserStat = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.ID == foreign.ID {
				local.R.IDPlayer = foreign
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.IDUserStat = local
				break
			}
		}
	}

	return nil
}

// SetIDPlayer of the userStat to the related item.
// Sets o.R.IDPlayer to related.
// Adds o to related.R.IDUserStat.
func (o *UserStat) SetIDPlayer(exec boil.Executor, insert bool, related *Player) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"user_stats\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"id"}),
		strmangle.WhereClause("\"", "\"", 2, userStatPrimaryKeyColumns),
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
		o.R = &userStatR{
			IDPlayer: related,
		}
	} else {
		o.R.IDPlayer = related
	}

	if related.R == nil {
		related.R = &playerR{
			IDUserStat: o,
		}
	} else {
		related.R.IDUserStat = o
	}

	return nil
}

// UserStats retrieves all the records using an executor.
func UserStats(mods ...qm.QueryMod) userStatQuery {
	mods = append(mods, qm.From("\"user_stats\""))
	return userStatQuery{NewQuery(mods...)}
}

// FindUserStat retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindUserStat(exec boil.Executor, iD string, selectCols ...string) (*UserStat, error) {
	userStatObj := &UserStat{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"user_stats\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, userStatObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from user_stats")
	}

	if err = userStatObj.doAfterSelectHooks(exec); err != nil {
		return userStatObj, err
	}

	return userStatObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *UserStat) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no user_stats provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(userStatColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	userStatInsertCacheMut.RLock()
	cache, cached := userStatInsertCache[key]
	userStatInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			userStatAllColumns,
			userStatColumnsWithDefault,
			userStatColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(userStatType, userStatMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(userStatType, userStatMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"user_stats\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"user_stats\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into user_stats")
	}

	if !cached {
		userStatInsertCacheMut.Lock()
		userStatInsertCache[key] = cache
		userStatInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the UserStat.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *UserStat) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	userStatUpdateCacheMut.RLock()
	cache, cached := userStatUpdateCache[key]
	userStatUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			userStatAllColumns,
			userStatPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update user_stats, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"user_stats\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, userStatPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(userStatType, userStatMapping, append(wl, userStatPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update user_stats row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for user_stats")
	}

	if !cached {
		userStatUpdateCacheMut.Lock()
		userStatUpdateCache[key] = cache
		userStatUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q userStatQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for user_stats")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for user_stats")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o UserStatSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userStatPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"user_stats\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, userStatPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in userStat slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all userStat")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *UserStat) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no user_stats provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(userStatColumnsWithDefault, o)

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

	userStatUpsertCacheMut.RLock()
	cache, cached := userStatUpsertCache[key]
	userStatUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			userStatAllColumns,
			userStatColumnsWithDefault,
			userStatColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			userStatAllColumns,
			userStatPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert user_stats, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(userStatPrimaryKeyColumns))
			copy(conflict, userStatPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"user_stats\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(userStatType, userStatMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(userStatType, userStatMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert user_stats")
	}

	if !cached {
		userStatUpsertCacheMut.Lock()
		userStatUpsertCache[key] = cache
		userStatUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single UserStat record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *UserStat) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no UserStat provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), userStatPrimaryKeyMapping)
	sql := "DELETE FROM \"user_stats\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from user_stats")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for user_stats")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q userStatQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no userStatQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from user_stats")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for user_stats")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o UserStatSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(userStatBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userStatPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"user_stats\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, userStatPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from userStat slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for user_stats")
	}

	if len(userStatAfterDeleteHooks) != 0 {
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
func (o *UserStat) Reload(exec boil.Executor) error {
	ret, err := FindUserStat(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *UserStatSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := UserStatSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userStatPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"user_stats\".* FROM \"user_stats\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, userStatPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in UserStatSlice")
	}

	*o = slice

	return nil
}

// UserStatExists checks if the UserStat row exists.
func UserStatExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"user_stats\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if user_stats exists")
	}

	return exists, nil
}
