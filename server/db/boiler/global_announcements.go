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

// GlobalAnnouncement is an object representing the database table.
type GlobalAnnouncement struct {
	ID                    string   `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Title                 string   `boiler:"title" boil:"title" json:"title" toml:"title" yaml:"title"`
	Message               string   `boiler:"message" boil:"message" json:"message" toml:"message" yaml:"message"`
	ShowFromBattleNumber  null.Int `boiler:"show_from_battle_number" boil:"show_from_battle_number" json:"show_from_battle_number,omitempty" toml:"show_from_battle_number" yaml:"show_from_battle_number,omitempty"`
	ShowUntilBattleNumber null.Int `boiler:"show_until_battle_number" boil:"show_until_battle_number" json:"show_until_battle_number,omitempty" toml:"show_until_battle_number" yaml:"show_until_battle_number,omitempty"`
	Severity              string   `boiler:"severity" boil:"severity" json:"severity" toml:"severity" yaml:"severity"`

	R *globalAnnouncementR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L globalAnnouncementL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var GlobalAnnouncementColumns = struct {
	ID                    string
	Title                 string
	Message               string
	ShowFromBattleNumber  string
	ShowUntilBattleNumber string
	Severity              string
}{
	ID:                    "id",
	Title:                 "title",
	Message:               "message",
	ShowFromBattleNumber:  "show_from_battle_number",
	ShowUntilBattleNumber: "show_until_battle_number",
	Severity:              "severity",
}

var GlobalAnnouncementTableColumns = struct {
	ID                    string
	Title                 string
	Message               string
	ShowFromBattleNumber  string
	ShowUntilBattleNumber string
	Severity              string
}{
	ID:                    "global_announcements.id",
	Title:                 "global_announcements.title",
	Message:               "global_announcements.message",
	ShowFromBattleNumber:  "global_announcements.show_from_battle_number",
	ShowUntilBattleNumber: "global_announcements.show_until_battle_number",
	Severity:              "global_announcements.severity",
}

// Generated where

var GlobalAnnouncementWhere = struct {
	ID                    whereHelperstring
	Title                 whereHelperstring
	Message               whereHelperstring
	ShowFromBattleNumber  whereHelpernull_Int
	ShowUntilBattleNumber whereHelpernull_Int
	Severity              whereHelperstring
}{
	ID:                    whereHelperstring{field: "\"global_announcements\".\"id\""},
	Title:                 whereHelperstring{field: "\"global_announcements\".\"title\""},
	Message:               whereHelperstring{field: "\"global_announcements\".\"message\""},
	ShowFromBattleNumber:  whereHelpernull_Int{field: "\"global_announcements\".\"show_from_battle_number\""},
	ShowUntilBattleNumber: whereHelpernull_Int{field: "\"global_announcements\".\"show_until_battle_number\""},
	Severity:              whereHelperstring{field: "\"global_announcements\".\"severity\""},
}

// GlobalAnnouncementRels is where relationship names are stored.
var GlobalAnnouncementRels = struct {
}{}

// globalAnnouncementR is where relationships are stored.
type globalAnnouncementR struct {
}

// NewStruct creates a new relationship struct
func (*globalAnnouncementR) NewStruct() *globalAnnouncementR {
	return &globalAnnouncementR{}
}

// globalAnnouncementL is where Load methods for each relationship are stored.
type globalAnnouncementL struct{}

var (
	globalAnnouncementAllColumns            = []string{"id", "title", "message", "show_from_battle_number", "show_until_battle_number", "severity"}
	globalAnnouncementColumnsWithoutDefault = []string{"title", "message", "severity"}
	globalAnnouncementColumnsWithDefault    = []string{"id", "show_from_battle_number", "show_until_battle_number"}
	globalAnnouncementPrimaryKeyColumns     = []string{"id"}
	globalAnnouncementGeneratedColumns      = []string{}
)

type (
	// GlobalAnnouncementSlice is an alias for a slice of pointers to GlobalAnnouncement.
	// This should almost always be used instead of []GlobalAnnouncement.
	GlobalAnnouncementSlice []*GlobalAnnouncement
	// GlobalAnnouncementHook is the signature for custom GlobalAnnouncement hook methods
	GlobalAnnouncementHook func(boil.Executor, *GlobalAnnouncement) error

	globalAnnouncementQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	globalAnnouncementType                 = reflect.TypeOf(&GlobalAnnouncement{})
	globalAnnouncementMapping              = queries.MakeStructMapping(globalAnnouncementType)
	globalAnnouncementPrimaryKeyMapping, _ = queries.BindMapping(globalAnnouncementType, globalAnnouncementMapping, globalAnnouncementPrimaryKeyColumns)
	globalAnnouncementInsertCacheMut       sync.RWMutex
	globalAnnouncementInsertCache          = make(map[string]insertCache)
	globalAnnouncementUpdateCacheMut       sync.RWMutex
	globalAnnouncementUpdateCache          = make(map[string]updateCache)
	globalAnnouncementUpsertCacheMut       sync.RWMutex
	globalAnnouncementUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var globalAnnouncementAfterSelectHooks []GlobalAnnouncementHook

var globalAnnouncementBeforeInsertHooks []GlobalAnnouncementHook
var globalAnnouncementAfterInsertHooks []GlobalAnnouncementHook

var globalAnnouncementBeforeUpdateHooks []GlobalAnnouncementHook
var globalAnnouncementAfterUpdateHooks []GlobalAnnouncementHook

var globalAnnouncementBeforeDeleteHooks []GlobalAnnouncementHook
var globalAnnouncementAfterDeleteHooks []GlobalAnnouncementHook

var globalAnnouncementBeforeUpsertHooks []GlobalAnnouncementHook
var globalAnnouncementAfterUpsertHooks []GlobalAnnouncementHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *GlobalAnnouncement) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range globalAnnouncementAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *GlobalAnnouncement) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range globalAnnouncementBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *GlobalAnnouncement) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range globalAnnouncementAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *GlobalAnnouncement) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range globalAnnouncementBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *GlobalAnnouncement) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range globalAnnouncementAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *GlobalAnnouncement) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range globalAnnouncementBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *GlobalAnnouncement) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range globalAnnouncementAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *GlobalAnnouncement) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range globalAnnouncementBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *GlobalAnnouncement) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range globalAnnouncementAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddGlobalAnnouncementHook registers your hook function for all future operations.
func AddGlobalAnnouncementHook(hookPoint boil.HookPoint, globalAnnouncementHook GlobalAnnouncementHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		globalAnnouncementAfterSelectHooks = append(globalAnnouncementAfterSelectHooks, globalAnnouncementHook)
	case boil.BeforeInsertHook:
		globalAnnouncementBeforeInsertHooks = append(globalAnnouncementBeforeInsertHooks, globalAnnouncementHook)
	case boil.AfterInsertHook:
		globalAnnouncementAfterInsertHooks = append(globalAnnouncementAfterInsertHooks, globalAnnouncementHook)
	case boil.BeforeUpdateHook:
		globalAnnouncementBeforeUpdateHooks = append(globalAnnouncementBeforeUpdateHooks, globalAnnouncementHook)
	case boil.AfterUpdateHook:
		globalAnnouncementAfterUpdateHooks = append(globalAnnouncementAfterUpdateHooks, globalAnnouncementHook)
	case boil.BeforeDeleteHook:
		globalAnnouncementBeforeDeleteHooks = append(globalAnnouncementBeforeDeleteHooks, globalAnnouncementHook)
	case boil.AfterDeleteHook:
		globalAnnouncementAfterDeleteHooks = append(globalAnnouncementAfterDeleteHooks, globalAnnouncementHook)
	case boil.BeforeUpsertHook:
		globalAnnouncementBeforeUpsertHooks = append(globalAnnouncementBeforeUpsertHooks, globalAnnouncementHook)
	case boil.AfterUpsertHook:
		globalAnnouncementAfterUpsertHooks = append(globalAnnouncementAfterUpsertHooks, globalAnnouncementHook)
	}
}

// One returns a single globalAnnouncement record from the query.
func (q globalAnnouncementQuery) One(exec boil.Executor) (*GlobalAnnouncement, error) {
	o := &GlobalAnnouncement{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for global_announcements")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all GlobalAnnouncement records from the query.
func (q globalAnnouncementQuery) All(exec boil.Executor) (GlobalAnnouncementSlice, error) {
	var o []*GlobalAnnouncement

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to GlobalAnnouncement slice")
	}

	if len(globalAnnouncementAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all GlobalAnnouncement records in the query.
func (q globalAnnouncementQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count global_announcements rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q globalAnnouncementQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if global_announcements exists")
	}

	return count > 0, nil
}

// GlobalAnnouncements retrieves all the records using an executor.
func GlobalAnnouncements(mods ...qm.QueryMod) globalAnnouncementQuery {
	mods = append(mods, qm.From("\"global_announcements\""))
	return globalAnnouncementQuery{NewQuery(mods...)}
}

// FindGlobalAnnouncement retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindGlobalAnnouncement(exec boil.Executor, iD string, selectCols ...string) (*GlobalAnnouncement, error) {
	globalAnnouncementObj := &GlobalAnnouncement{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"global_announcements\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, globalAnnouncementObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from global_announcements")
	}

	if err = globalAnnouncementObj.doAfterSelectHooks(exec); err != nil {
		return globalAnnouncementObj, err
	}

	return globalAnnouncementObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *GlobalAnnouncement) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no global_announcements provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(globalAnnouncementColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	globalAnnouncementInsertCacheMut.RLock()
	cache, cached := globalAnnouncementInsertCache[key]
	globalAnnouncementInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			globalAnnouncementAllColumns,
			globalAnnouncementColumnsWithDefault,
			globalAnnouncementColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(globalAnnouncementType, globalAnnouncementMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(globalAnnouncementType, globalAnnouncementMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"global_announcements\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"global_announcements\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into global_announcements")
	}

	if !cached {
		globalAnnouncementInsertCacheMut.Lock()
		globalAnnouncementInsertCache[key] = cache
		globalAnnouncementInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the GlobalAnnouncement.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *GlobalAnnouncement) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	globalAnnouncementUpdateCacheMut.RLock()
	cache, cached := globalAnnouncementUpdateCache[key]
	globalAnnouncementUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			globalAnnouncementAllColumns,
			globalAnnouncementPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update global_announcements, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"global_announcements\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, globalAnnouncementPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(globalAnnouncementType, globalAnnouncementMapping, append(wl, globalAnnouncementPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update global_announcements row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for global_announcements")
	}

	if !cached {
		globalAnnouncementUpdateCacheMut.Lock()
		globalAnnouncementUpdateCache[key] = cache
		globalAnnouncementUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q globalAnnouncementQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for global_announcements")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for global_announcements")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o GlobalAnnouncementSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), globalAnnouncementPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"global_announcements\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, globalAnnouncementPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in globalAnnouncement slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all globalAnnouncement")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *GlobalAnnouncement) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no global_announcements provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(globalAnnouncementColumnsWithDefault, o)

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

	globalAnnouncementUpsertCacheMut.RLock()
	cache, cached := globalAnnouncementUpsertCache[key]
	globalAnnouncementUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			globalAnnouncementAllColumns,
			globalAnnouncementColumnsWithDefault,
			globalAnnouncementColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			globalAnnouncementAllColumns,
			globalAnnouncementPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert global_announcements, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(globalAnnouncementPrimaryKeyColumns))
			copy(conflict, globalAnnouncementPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"global_announcements\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(globalAnnouncementType, globalAnnouncementMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(globalAnnouncementType, globalAnnouncementMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert global_announcements")
	}

	if !cached {
		globalAnnouncementUpsertCacheMut.Lock()
		globalAnnouncementUpsertCache[key] = cache
		globalAnnouncementUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single GlobalAnnouncement record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *GlobalAnnouncement) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no GlobalAnnouncement provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), globalAnnouncementPrimaryKeyMapping)
	sql := "DELETE FROM \"global_announcements\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from global_announcements")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for global_announcements")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q globalAnnouncementQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no globalAnnouncementQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from global_announcements")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for global_announcements")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o GlobalAnnouncementSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(globalAnnouncementBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), globalAnnouncementPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"global_announcements\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, globalAnnouncementPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from globalAnnouncement slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for global_announcements")
	}

	if len(globalAnnouncementAfterDeleteHooks) != 0 {
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
func (o *GlobalAnnouncement) Reload(exec boil.Executor) error {
	ret, err := FindGlobalAnnouncement(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *GlobalAnnouncementSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := GlobalAnnouncementSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), globalAnnouncementPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"global_announcements\".* FROM \"global_announcements\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, globalAnnouncementPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in GlobalAnnouncementSlice")
	}

	*o = slice

	return nil
}

// GlobalAnnouncementExists checks if the GlobalAnnouncement row exists.
func GlobalAnnouncementExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"global_announcements\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if global_announcements exists")
	}

	return exists, nil
}
