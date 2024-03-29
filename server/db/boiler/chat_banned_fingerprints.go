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

// ChatBannedFingerprint is an object representing the database table.
type ChatBannedFingerprint struct {
	FingerprintID string `boiler:"fingerprint_id" boil:"fingerprint_id" json:"fingerprint_id" toml:"fingerprint_id" yaml:"fingerprint_id"`

	R *chatBannedFingerprintR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L chatBannedFingerprintL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ChatBannedFingerprintColumns = struct {
	FingerprintID string
}{
	FingerprintID: "fingerprint_id",
}

var ChatBannedFingerprintTableColumns = struct {
	FingerprintID string
}{
	FingerprintID: "chat_banned_fingerprints.fingerprint_id",
}

// Generated where

var ChatBannedFingerprintWhere = struct {
	FingerprintID whereHelperstring
}{
	FingerprintID: whereHelperstring{field: "\"chat_banned_fingerprints\".\"fingerprint_id\""},
}

// ChatBannedFingerprintRels is where relationship names are stored.
var ChatBannedFingerprintRels = struct {
}{}

// chatBannedFingerprintR is where relationships are stored.
type chatBannedFingerprintR struct {
}

// NewStruct creates a new relationship struct
func (*chatBannedFingerprintR) NewStruct() *chatBannedFingerprintR {
	return &chatBannedFingerprintR{}
}

// chatBannedFingerprintL is where Load methods for each relationship are stored.
type chatBannedFingerprintL struct{}

var (
	chatBannedFingerprintAllColumns            = []string{"fingerprint_id"}
	chatBannedFingerprintColumnsWithoutDefault = []string{"fingerprint_id"}
	chatBannedFingerprintColumnsWithDefault    = []string{}
	chatBannedFingerprintPrimaryKeyColumns     = []string{"fingerprint_id"}
	chatBannedFingerprintGeneratedColumns      = []string{}
)

type (
	// ChatBannedFingerprintSlice is an alias for a slice of pointers to ChatBannedFingerprint.
	// This should almost always be used instead of []ChatBannedFingerprint.
	ChatBannedFingerprintSlice []*ChatBannedFingerprint
	// ChatBannedFingerprintHook is the signature for custom ChatBannedFingerprint hook methods
	ChatBannedFingerprintHook func(boil.Executor, *ChatBannedFingerprint) error

	chatBannedFingerprintQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	chatBannedFingerprintType                 = reflect.TypeOf(&ChatBannedFingerprint{})
	chatBannedFingerprintMapping              = queries.MakeStructMapping(chatBannedFingerprintType)
	chatBannedFingerprintPrimaryKeyMapping, _ = queries.BindMapping(chatBannedFingerprintType, chatBannedFingerprintMapping, chatBannedFingerprintPrimaryKeyColumns)
	chatBannedFingerprintInsertCacheMut       sync.RWMutex
	chatBannedFingerprintInsertCache          = make(map[string]insertCache)
	chatBannedFingerprintUpdateCacheMut       sync.RWMutex
	chatBannedFingerprintUpdateCache          = make(map[string]updateCache)
	chatBannedFingerprintUpsertCacheMut       sync.RWMutex
	chatBannedFingerprintUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var chatBannedFingerprintAfterSelectHooks []ChatBannedFingerprintHook

var chatBannedFingerprintBeforeInsertHooks []ChatBannedFingerprintHook
var chatBannedFingerprintAfterInsertHooks []ChatBannedFingerprintHook

var chatBannedFingerprintBeforeUpdateHooks []ChatBannedFingerprintHook
var chatBannedFingerprintAfterUpdateHooks []ChatBannedFingerprintHook

var chatBannedFingerprintBeforeDeleteHooks []ChatBannedFingerprintHook
var chatBannedFingerprintAfterDeleteHooks []ChatBannedFingerprintHook

var chatBannedFingerprintBeforeUpsertHooks []ChatBannedFingerprintHook
var chatBannedFingerprintAfterUpsertHooks []ChatBannedFingerprintHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *ChatBannedFingerprint) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range chatBannedFingerprintAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *ChatBannedFingerprint) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range chatBannedFingerprintBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *ChatBannedFingerprint) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range chatBannedFingerprintAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *ChatBannedFingerprint) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range chatBannedFingerprintBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *ChatBannedFingerprint) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range chatBannedFingerprintAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *ChatBannedFingerprint) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range chatBannedFingerprintBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *ChatBannedFingerprint) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range chatBannedFingerprintAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *ChatBannedFingerprint) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range chatBannedFingerprintBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *ChatBannedFingerprint) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range chatBannedFingerprintAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddChatBannedFingerprintHook registers your hook function for all future operations.
func AddChatBannedFingerprintHook(hookPoint boil.HookPoint, chatBannedFingerprintHook ChatBannedFingerprintHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		chatBannedFingerprintAfterSelectHooks = append(chatBannedFingerprintAfterSelectHooks, chatBannedFingerprintHook)
	case boil.BeforeInsertHook:
		chatBannedFingerprintBeforeInsertHooks = append(chatBannedFingerprintBeforeInsertHooks, chatBannedFingerprintHook)
	case boil.AfterInsertHook:
		chatBannedFingerprintAfterInsertHooks = append(chatBannedFingerprintAfterInsertHooks, chatBannedFingerprintHook)
	case boil.BeforeUpdateHook:
		chatBannedFingerprintBeforeUpdateHooks = append(chatBannedFingerprintBeforeUpdateHooks, chatBannedFingerprintHook)
	case boil.AfterUpdateHook:
		chatBannedFingerprintAfterUpdateHooks = append(chatBannedFingerprintAfterUpdateHooks, chatBannedFingerprintHook)
	case boil.BeforeDeleteHook:
		chatBannedFingerprintBeforeDeleteHooks = append(chatBannedFingerprintBeforeDeleteHooks, chatBannedFingerprintHook)
	case boil.AfterDeleteHook:
		chatBannedFingerprintAfterDeleteHooks = append(chatBannedFingerprintAfterDeleteHooks, chatBannedFingerprintHook)
	case boil.BeforeUpsertHook:
		chatBannedFingerprintBeforeUpsertHooks = append(chatBannedFingerprintBeforeUpsertHooks, chatBannedFingerprintHook)
	case boil.AfterUpsertHook:
		chatBannedFingerprintAfterUpsertHooks = append(chatBannedFingerprintAfterUpsertHooks, chatBannedFingerprintHook)
	}
}

// One returns a single chatBannedFingerprint record from the query.
func (q chatBannedFingerprintQuery) One(exec boil.Executor) (*ChatBannedFingerprint, error) {
	o := &ChatBannedFingerprint{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for chat_banned_fingerprints")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all ChatBannedFingerprint records from the query.
func (q chatBannedFingerprintQuery) All(exec boil.Executor) (ChatBannedFingerprintSlice, error) {
	var o []*ChatBannedFingerprint

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to ChatBannedFingerprint slice")
	}

	if len(chatBannedFingerprintAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all ChatBannedFingerprint records in the query.
func (q chatBannedFingerprintQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count chat_banned_fingerprints rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q chatBannedFingerprintQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if chat_banned_fingerprints exists")
	}

	return count > 0, nil
}

// ChatBannedFingerprints retrieves all the records using an executor.
func ChatBannedFingerprints(mods ...qm.QueryMod) chatBannedFingerprintQuery {
	mods = append(mods, qm.From("\"chat_banned_fingerprints\""))
	return chatBannedFingerprintQuery{NewQuery(mods...)}
}

// FindChatBannedFingerprint retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindChatBannedFingerprint(exec boil.Executor, fingerprintID string, selectCols ...string) (*ChatBannedFingerprint, error) {
	chatBannedFingerprintObj := &ChatBannedFingerprint{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"chat_banned_fingerprints\" where \"fingerprint_id\"=$1", sel,
	)

	q := queries.Raw(query, fingerprintID)

	err := q.Bind(nil, exec, chatBannedFingerprintObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from chat_banned_fingerprints")
	}

	if err = chatBannedFingerprintObj.doAfterSelectHooks(exec); err != nil {
		return chatBannedFingerprintObj, err
	}

	return chatBannedFingerprintObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *ChatBannedFingerprint) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no chat_banned_fingerprints provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(chatBannedFingerprintColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	chatBannedFingerprintInsertCacheMut.RLock()
	cache, cached := chatBannedFingerprintInsertCache[key]
	chatBannedFingerprintInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			chatBannedFingerprintAllColumns,
			chatBannedFingerprintColumnsWithDefault,
			chatBannedFingerprintColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(chatBannedFingerprintType, chatBannedFingerprintMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(chatBannedFingerprintType, chatBannedFingerprintMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"chat_banned_fingerprints\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"chat_banned_fingerprints\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into chat_banned_fingerprints")
	}

	if !cached {
		chatBannedFingerprintInsertCacheMut.Lock()
		chatBannedFingerprintInsertCache[key] = cache
		chatBannedFingerprintInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the ChatBannedFingerprint.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *ChatBannedFingerprint) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	chatBannedFingerprintUpdateCacheMut.RLock()
	cache, cached := chatBannedFingerprintUpdateCache[key]
	chatBannedFingerprintUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			chatBannedFingerprintAllColumns,
			chatBannedFingerprintPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update chat_banned_fingerprints, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"chat_banned_fingerprints\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, chatBannedFingerprintPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(chatBannedFingerprintType, chatBannedFingerprintMapping, append(wl, chatBannedFingerprintPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update chat_banned_fingerprints row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for chat_banned_fingerprints")
	}

	if !cached {
		chatBannedFingerprintUpdateCacheMut.Lock()
		chatBannedFingerprintUpdateCache[key] = cache
		chatBannedFingerprintUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q chatBannedFingerprintQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for chat_banned_fingerprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for chat_banned_fingerprints")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ChatBannedFingerprintSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), chatBannedFingerprintPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"chat_banned_fingerprints\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, chatBannedFingerprintPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in chatBannedFingerprint slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all chatBannedFingerprint")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *ChatBannedFingerprint) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no chat_banned_fingerprints provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(chatBannedFingerprintColumnsWithDefault, o)

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

	chatBannedFingerprintUpsertCacheMut.RLock()
	cache, cached := chatBannedFingerprintUpsertCache[key]
	chatBannedFingerprintUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			chatBannedFingerprintAllColumns,
			chatBannedFingerprintColumnsWithDefault,
			chatBannedFingerprintColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			chatBannedFingerprintAllColumns,
			chatBannedFingerprintPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert chat_banned_fingerprints, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(chatBannedFingerprintPrimaryKeyColumns))
			copy(conflict, chatBannedFingerprintPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"chat_banned_fingerprints\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(chatBannedFingerprintType, chatBannedFingerprintMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(chatBannedFingerprintType, chatBannedFingerprintMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert chat_banned_fingerprints")
	}

	if !cached {
		chatBannedFingerprintUpsertCacheMut.Lock()
		chatBannedFingerprintUpsertCache[key] = cache
		chatBannedFingerprintUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single ChatBannedFingerprint record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *ChatBannedFingerprint) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no ChatBannedFingerprint provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), chatBannedFingerprintPrimaryKeyMapping)
	sql := "DELETE FROM \"chat_banned_fingerprints\" WHERE \"fingerprint_id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from chat_banned_fingerprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for chat_banned_fingerprints")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q chatBannedFingerprintQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no chatBannedFingerprintQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from chat_banned_fingerprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for chat_banned_fingerprints")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ChatBannedFingerprintSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(chatBannedFingerprintBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), chatBannedFingerprintPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"chat_banned_fingerprints\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, chatBannedFingerprintPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from chatBannedFingerprint slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for chat_banned_fingerprints")
	}

	if len(chatBannedFingerprintAfterDeleteHooks) != 0 {
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
func (o *ChatBannedFingerprint) Reload(exec boil.Executor) error {
	ret, err := FindChatBannedFingerprint(exec, o.FingerprintID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ChatBannedFingerprintSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ChatBannedFingerprintSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), chatBannedFingerprintPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"chat_banned_fingerprints\".* FROM \"chat_banned_fingerprints\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, chatBannedFingerprintPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in ChatBannedFingerprintSlice")
	}

	*o = slice

	return nil
}

// ChatBannedFingerprintExists checks if the ChatBannedFingerprint row exists.
func ChatBannedFingerprintExists(exec boil.Executor, fingerprintID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"chat_banned_fingerprints\" where \"fingerprint_id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, fingerprintID)
	}
	row := exec.QueryRow(sql, fingerprintID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if chat_banned_fingerprints exists")
	}

	return exists, nil
}
