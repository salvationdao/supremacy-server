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
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// StreamList is an object representing the database table.
type StreamList struct {
	Host          string          `boiler:"host" boil:"host" json:"host" toml:"host" yaml:"host"`
	Name          string          `boiler:"name" boil:"name" json:"name" toml:"name" yaml:"name"`
	URL           string          `boiler:"url" boil:"url" json:"url" toml:"url" yaml:"url"`
	StreamID      string          `boiler:"stream_id" boil:"stream_id" json:"stream_id" toml:"stream_id" yaml:"stream_id"`
	Region        string          `boiler:"region" boil:"region" json:"region" toml:"region" yaml:"region"`
	Resolution    string          `boiler:"resolution" boil:"resolution" json:"resolution" toml:"resolution" yaml:"resolution"`
	BitRatesKBits int             `boiler:"bit_rates_k_bits" boil:"bit_rates_k_bits" json:"bit_rates_k_bits" toml:"bit_rates_k_bits" yaml:"bit_rates_k_bits"`
	UserMax       int             `boiler:"user_max" boil:"user_max" json:"user_max" toml:"user_max" yaml:"user_max"`
	UsersNow      int             `boiler:"users_now" boil:"users_now" json:"users_now" toml:"users_now" yaml:"users_now"`
	Active        bool            `boiler:"active" boil:"active" json:"active" toml:"active" yaml:"active"`
	Status        string          `boiler:"status" boil:"status" json:"status" toml:"status" yaml:"status"`
	Latitude      decimal.Decimal `boiler:"latitude" boil:"latitude" json:"latitude" toml:"latitude" yaml:"latitude"`
	Longitude     decimal.Decimal `boiler:"longitude" boil:"longitude" json:"longitude" toml:"longitude" yaml:"longitude"`

	R *streamListR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L streamListL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var StreamListColumns = struct {
	Host          string
	Name          string
	URL           string
	StreamID      string
	Region        string
	Resolution    string
	BitRatesKBits string
	UserMax       string
	UsersNow      string
	Active        string
	Status        string
	Latitude      string
	Longitude     string
}{
	Host:          "host",
	Name:          "name",
	URL:           "url",
	StreamID:      "stream_id",
	Region:        "region",
	Resolution:    "resolution",
	BitRatesKBits: "bit_rates_k_bits",
	UserMax:       "user_max",
	UsersNow:      "users_now",
	Active:        "active",
	Status:        "status",
	Latitude:      "latitude",
	Longitude:     "longitude",
}

var StreamListTableColumns = struct {
	Host          string
	Name          string
	URL           string
	StreamID      string
	Region        string
	Resolution    string
	BitRatesKBits string
	UserMax       string
	UsersNow      string
	Active        string
	Status        string
	Latitude      string
	Longitude     string
}{
	Host:          "stream_list.host",
	Name:          "stream_list.name",
	URL:           "stream_list.url",
	StreamID:      "stream_list.stream_id",
	Region:        "stream_list.region",
	Resolution:    "stream_list.resolution",
	BitRatesKBits: "stream_list.bit_rates_k_bits",
	UserMax:       "stream_list.user_max",
	UsersNow:      "stream_list.users_now",
	Active:        "stream_list.active",
	Status:        "stream_list.status",
	Latitude:      "stream_list.latitude",
	Longitude:     "stream_list.longitude",
}

// Generated where

var StreamListWhere = struct {
	Host          whereHelperstring
	Name          whereHelperstring
	URL           whereHelperstring
	StreamID      whereHelperstring
	Region        whereHelperstring
	Resolution    whereHelperstring
	BitRatesKBits whereHelperint
	UserMax       whereHelperint
	UsersNow      whereHelperint
	Active        whereHelperbool
	Status        whereHelperstring
	Latitude      whereHelperdecimal_Decimal
	Longitude     whereHelperdecimal_Decimal
}{
	Host:          whereHelperstring{field: "\"stream_list\".\"host\""},
	Name:          whereHelperstring{field: "\"stream_list\".\"name\""},
	URL:           whereHelperstring{field: "\"stream_list\".\"url\""},
	StreamID:      whereHelperstring{field: "\"stream_list\".\"stream_id\""},
	Region:        whereHelperstring{field: "\"stream_list\".\"region\""},
	Resolution:    whereHelperstring{field: "\"stream_list\".\"resolution\""},
	BitRatesKBits: whereHelperint{field: "\"stream_list\".\"bit_rates_k_bits\""},
	UserMax:       whereHelperint{field: "\"stream_list\".\"user_max\""},
	UsersNow:      whereHelperint{field: "\"stream_list\".\"users_now\""},
	Active:        whereHelperbool{field: "\"stream_list\".\"active\""},
	Status:        whereHelperstring{field: "\"stream_list\".\"status\""},
	Latitude:      whereHelperdecimal_Decimal{field: "\"stream_list\".\"latitude\""},
	Longitude:     whereHelperdecimal_Decimal{field: "\"stream_list\".\"longitude\""},
}

// StreamListRels is where relationship names are stored.
var StreamListRels = struct {
}{}

// streamListR is where relationships are stored.
type streamListR struct {
}

// NewStruct creates a new relationship struct
func (*streamListR) NewStruct() *streamListR {
	return &streamListR{}
}

// streamListL is where Load methods for each relationship are stored.
type streamListL struct{}

var (
	streamListAllColumns            = []string{"host", "name", "url", "stream_id", "region", "resolution", "bit_rates_k_bits", "user_max", "users_now", "active", "status", "latitude", "longitude"}
	streamListColumnsWithoutDefault = []string{"host", "name", "url", "stream_id", "region", "resolution", "bit_rates_k_bits", "user_max", "users_now", "active", "status", "latitude", "longitude"}
	streamListColumnsWithDefault    = []string{}
	streamListPrimaryKeyColumns     = []string{"host"}
	streamListGeneratedColumns      = []string{}
)

type (
	// StreamListSlice is an alias for a slice of pointers to StreamList.
	// This should almost always be used instead of []StreamList.
	StreamListSlice []*StreamList
	// StreamListHook is the signature for custom StreamList hook methods
	StreamListHook func(boil.Executor, *StreamList) error

	streamListQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	streamListType                 = reflect.TypeOf(&StreamList{})
	streamListMapping              = queries.MakeStructMapping(streamListType)
	streamListPrimaryKeyMapping, _ = queries.BindMapping(streamListType, streamListMapping, streamListPrimaryKeyColumns)
	streamListInsertCacheMut       sync.RWMutex
	streamListInsertCache          = make(map[string]insertCache)
	streamListUpdateCacheMut       sync.RWMutex
	streamListUpdateCache          = make(map[string]updateCache)
	streamListUpsertCacheMut       sync.RWMutex
	streamListUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var streamListAfterSelectHooks []StreamListHook

var streamListBeforeInsertHooks []StreamListHook
var streamListAfterInsertHooks []StreamListHook

var streamListBeforeUpdateHooks []StreamListHook
var streamListAfterUpdateHooks []StreamListHook

var streamListBeforeDeleteHooks []StreamListHook
var streamListAfterDeleteHooks []StreamListHook

var streamListBeforeUpsertHooks []StreamListHook
var streamListAfterUpsertHooks []StreamListHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *StreamList) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range streamListAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *StreamList) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range streamListBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *StreamList) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range streamListAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *StreamList) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range streamListBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *StreamList) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range streamListAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *StreamList) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range streamListBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *StreamList) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range streamListAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *StreamList) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range streamListBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *StreamList) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range streamListAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddStreamListHook registers your hook function for all future operations.
func AddStreamListHook(hookPoint boil.HookPoint, streamListHook StreamListHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		streamListAfterSelectHooks = append(streamListAfterSelectHooks, streamListHook)
	case boil.BeforeInsertHook:
		streamListBeforeInsertHooks = append(streamListBeforeInsertHooks, streamListHook)
	case boil.AfterInsertHook:
		streamListAfterInsertHooks = append(streamListAfterInsertHooks, streamListHook)
	case boil.BeforeUpdateHook:
		streamListBeforeUpdateHooks = append(streamListBeforeUpdateHooks, streamListHook)
	case boil.AfterUpdateHook:
		streamListAfterUpdateHooks = append(streamListAfterUpdateHooks, streamListHook)
	case boil.BeforeDeleteHook:
		streamListBeforeDeleteHooks = append(streamListBeforeDeleteHooks, streamListHook)
	case boil.AfterDeleteHook:
		streamListAfterDeleteHooks = append(streamListAfterDeleteHooks, streamListHook)
	case boil.BeforeUpsertHook:
		streamListBeforeUpsertHooks = append(streamListBeforeUpsertHooks, streamListHook)
	case boil.AfterUpsertHook:
		streamListAfterUpsertHooks = append(streamListAfterUpsertHooks, streamListHook)
	}
}

// One returns a single streamList record from the query.
func (q streamListQuery) One(exec boil.Executor) (*StreamList, error) {
	o := &StreamList{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for stream_list")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all StreamList records from the query.
func (q streamListQuery) All(exec boil.Executor) (StreamListSlice, error) {
	var o []*StreamList

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to StreamList slice")
	}

	if len(streamListAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all StreamList records in the query.
func (q streamListQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count stream_list rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q streamListQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if stream_list exists")
	}

	return count > 0, nil
}

// StreamLists retrieves all the records using an executor.
func StreamLists(mods ...qm.QueryMod) streamListQuery {
	mods = append(mods, qm.From("\"stream_list\""))
	return streamListQuery{NewQuery(mods...)}
}

// FindStreamList retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindStreamList(exec boil.Executor, host string, selectCols ...string) (*StreamList, error) {
	streamListObj := &StreamList{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"stream_list\" where \"host\"=$1", sel,
	)

	q := queries.Raw(query, host)

	err := q.Bind(nil, exec, streamListObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from stream_list")
	}

	if err = streamListObj.doAfterSelectHooks(exec); err != nil {
		return streamListObj, err
	}

	return streamListObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *StreamList) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no stream_list provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(streamListColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	streamListInsertCacheMut.RLock()
	cache, cached := streamListInsertCache[key]
	streamListInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			streamListAllColumns,
			streamListColumnsWithDefault,
			streamListColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(streamListType, streamListMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(streamListType, streamListMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"stream_list\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"stream_list\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into stream_list")
	}

	if !cached {
		streamListInsertCacheMut.Lock()
		streamListInsertCache[key] = cache
		streamListInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the StreamList.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *StreamList) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	streamListUpdateCacheMut.RLock()
	cache, cached := streamListUpdateCache[key]
	streamListUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			streamListAllColumns,
			streamListPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update stream_list, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"stream_list\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, streamListPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(streamListType, streamListMapping, append(wl, streamListPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update stream_list row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for stream_list")
	}

	if !cached {
		streamListUpdateCacheMut.Lock()
		streamListUpdateCache[key] = cache
		streamListUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q streamListQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for stream_list")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for stream_list")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o StreamListSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), streamListPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"stream_list\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, streamListPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in streamList slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all streamList")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *StreamList) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no stream_list provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(streamListColumnsWithDefault, o)

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

	streamListUpsertCacheMut.RLock()
	cache, cached := streamListUpsertCache[key]
	streamListUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			streamListAllColumns,
			streamListColumnsWithDefault,
			streamListColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			streamListAllColumns,
			streamListPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert stream_list, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(streamListPrimaryKeyColumns))
			copy(conflict, streamListPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"stream_list\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(streamListType, streamListMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(streamListType, streamListMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert stream_list")
	}

	if !cached {
		streamListUpsertCacheMut.Lock()
		streamListUpsertCache[key] = cache
		streamListUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single StreamList record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *StreamList) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no StreamList provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), streamListPrimaryKeyMapping)
	sql := "DELETE FROM \"stream_list\" WHERE \"host\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from stream_list")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for stream_list")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q streamListQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no streamListQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from stream_list")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for stream_list")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o StreamListSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(streamListBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), streamListPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"stream_list\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, streamListPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from streamList slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for stream_list")
	}

	if len(streamListAfterDeleteHooks) != 0 {
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
func (o *StreamList) Reload(exec boil.Executor) error {
	ret, err := FindStreamList(exec, o.Host)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *StreamListSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := StreamListSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), streamListPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"stream_list\".* FROM \"stream_list\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, streamListPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in StreamListSlice")
	}

	*o = slice

	return nil
}

// StreamListExists checks if the StreamList row exists.
func StreamListExists(exec boil.Executor, host string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"stream_list\" where \"host\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, host)
	}
	row := exec.QueryRow(sql, host)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if stream_list exists")
	}

	return exists, nil
}
