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

// TemplatesTemplateModule is an object representing the database table.
type TemplatesTemplateModule struct {
	ID        string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	DeletedAt null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deletedAt,omitempty" toml:"deletedAt" yaml:"deletedAt,omitempty"`
	UpdatedAt time.Time `boiler:"updated_at" boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
	CreatedAt time.Time `boiler:"created_at" boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`

	R *templatesTemplateModuleR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L templatesTemplateModuleL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var TemplatesTemplateModuleColumns = struct {
	ID        string
	DeletedAt string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "id",
	DeletedAt: "deleted_at",
	UpdatedAt: "updated_at",
	CreatedAt: "created_at",
}

var TemplatesTemplateModuleTableColumns = struct {
	ID        string
	DeletedAt string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "templates_template_modules.id",
	DeletedAt: "templates_template_modules.deleted_at",
	UpdatedAt: "templates_template_modules.updated_at",
	CreatedAt: "templates_template_modules.created_at",
}

// Generated where

var TemplatesTemplateModuleWhere = struct {
	ID        whereHelperstring
	DeletedAt whereHelpernull_Time
	UpdatedAt whereHelpertime_Time
	CreatedAt whereHelpertime_Time
}{
	ID:        whereHelperstring{field: "\"templates_template_modules\".\"id\""},
	DeletedAt: whereHelpernull_Time{field: "\"templates_template_modules\".\"deleted_at\""},
	UpdatedAt: whereHelpertime_Time{field: "\"templates_template_modules\".\"updated_at\""},
	CreatedAt: whereHelpertime_Time{field: "\"templates_template_modules\".\"created_at\""},
}

// TemplatesTemplateModuleRels is where relationship names are stored.
var TemplatesTemplateModuleRels = struct {
}{}

// templatesTemplateModuleR is where relationships are stored.
type templatesTemplateModuleR struct {
}

// NewStruct creates a new relationship struct
func (*templatesTemplateModuleR) NewStruct() *templatesTemplateModuleR {
	return &templatesTemplateModuleR{}
}

// templatesTemplateModuleL is where Load methods for each relationship are stored.
type templatesTemplateModuleL struct{}

var (
	templatesTemplateModuleAllColumns            = []string{"id", "deleted_at", "updated_at", "created_at"}
	templatesTemplateModuleColumnsWithoutDefault = []string{}
	templatesTemplateModuleColumnsWithDefault    = []string{"id", "deleted_at", "updated_at", "created_at"}
	templatesTemplateModulePrimaryKeyColumns     = []string{"id"}
	templatesTemplateModuleGeneratedColumns      = []string{}
)

type (
	// TemplatesTemplateModuleSlice is an alias for a slice of pointers to TemplatesTemplateModule.
	// This should almost always be used instead of []TemplatesTemplateModule.
	TemplatesTemplateModuleSlice []*TemplatesTemplateModule
	// TemplatesTemplateModuleHook is the signature for custom TemplatesTemplateModule hook methods
	TemplatesTemplateModuleHook func(boil.Executor, *TemplatesTemplateModule) error

	templatesTemplateModuleQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	templatesTemplateModuleType                 = reflect.TypeOf(&TemplatesTemplateModule{})
	templatesTemplateModuleMapping              = queries.MakeStructMapping(templatesTemplateModuleType)
	templatesTemplateModulePrimaryKeyMapping, _ = queries.BindMapping(templatesTemplateModuleType, templatesTemplateModuleMapping, templatesTemplateModulePrimaryKeyColumns)
	templatesTemplateModuleInsertCacheMut       sync.RWMutex
	templatesTemplateModuleInsertCache          = make(map[string]insertCache)
	templatesTemplateModuleUpdateCacheMut       sync.RWMutex
	templatesTemplateModuleUpdateCache          = make(map[string]updateCache)
	templatesTemplateModuleUpsertCacheMut       sync.RWMutex
	templatesTemplateModuleUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var templatesTemplateModuleAfterSelectHooks []TemplatesTemplateModuleHook

var templatesTemplateModuleBeforeInsertHooks []TemplatesTemplateModuleHook
var templatesTemplateModuleAfterInsertHooks []TemplatesTemplateModuleHook

var templatesTemplateModuleBeforeUpdateHooks []TemplatesTemplateModuleHook
var templatesTemplateModuleAfterUpdateHooks []TemplatesTemplateModuleHook

var templatesTemplateModuleBeforeDeleteHooks []TemplatesTemplateModuleHook
var templatesTemplateModuleAfterDeleteHooks []TemplatesTemplateModuleHook

var templatesTemplateModuleBeforeUpsertHooks []TemplatesTemplateModuleHook
var templatesTemplateModuleAfterUpsertHooks []TemplatesTemplateModuleHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *TemplatesTemplateModule) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range templatesTemplateModuleAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *TemplatesTemplateModule) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templatesTemplateModuleBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *TemplatesTemplateModule) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templatesTemplateModuleAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *TemplatesTemplateModule) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range templatesTemplateModuleBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *TemplatesTemplateModule) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range templatesTemplateModuleAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *TemplatesTemplateModule) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range templatesTemplateModuleBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *TemplatesTemplateModule) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range templatesTemplateModuleAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *TemplatesTemplateModule) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templatesTemplateModuleBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *TemplatesTemplateModule) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templatesTemplateModuleAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddTemplatesTemplateModuleHook registers your hook function for all future operations.
func AddTemplatesTemplateModuleHook(hookPoint boil.HookPoint, templatesTemplateModuleHook TemplatesTemplateModuleHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		templatesTemplateModuleAfterSelectHooks = append(templatesTemplateModuleAfterSelectHooks, templatesTemplateModuleHook)
	case boil.BeforeInsertHook:
		templatesTemplateModuleBeforeInsertHooks = append(templatesTemplateModuleBeforeInsertHooks, templatesTemplateModuleHook)
	case boil.AfterInsertHook:
		templatesTemplateModuleAfterInsertHooks = append(templatesTemplateModuleAfterInsertHooks, templatesTemplateModuleHook)
	case boil.BeforeUpdateHook:
		templatesTemplateModuleBeforeUpdateHooks = append(templatesTemplateModuleBeforeUpdateHooks, templatesTemplateModuleHook)
	case boil.AfterUpdateHook:
		templatesTemplateModuleAfterUpdateHooks = append(templatesTemplateModuleAfterUpdateHooks, templatesTemplateModuleHook)
	case boil.BeforeDeleteHook:
		templatesTemplateModuleBeforeDeleteHooks = append(templatesTemplateModuleBeforeDeleteHooks, templatesTemplateModuleHook)
	case boil.AfterDeleteHook:
		templatesTemplateModuleAfterDeleteHooks = append(templatesTemplateModuleAfterDeleteHooks, templatesTemplateModuleHook)
	case boil.BeforeUpsertHook:
		templatesTemplateModuleBeforeUpsertHooks = append(templatesTemplateModuleBeforeUpsertHooks, templatesTemplateModuleHook)
	case boil.AfterUpsertHook:
		templatesTemplateModuleAfterUpsertHooks = append(templatesTemplateModuleAfterUpsertHooks, templatesTemplateModuleHook)
	}
}

// One returns a single templatesTemplateModule record from the query.
func (q templatesTemplateModuleQuery) One(exec boil.Executor) (*TemplatesTemplateModule, error) {
	o := &TemplatesTemplateModule{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for templates_template_modules")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all TemplatesTemplateModule records from the query.
func (q templatesTemplateModuleQuery) All(exec boil.Executor) (TemplatesTemplateModuleSlice, error) {
	var o []*TemplatesTemplateModule

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to TemplatesTemplateModule slice")
	}

	if len(templatesTemplateModuleAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all TemplatesTemplateModule records in the query.
func (q templatesTemplateModuleQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count templates_template_modules rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q templatesTemplateModuleQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if templates_template_modules exists")
	}

	return count > 0, nil
}

// TemplatesTemplateModules retrieves all the records using an executor.
func TemplatesTemplateModules(mods ...qm.QueryMod) templatesTemplateModuleQuery {
	mods = append(mods, qm.From("\"templates_template_modules\""), qmhelper.WhereIsNull("\"templates_template_modules\".\"deleted_at\""))
	return templatesTemplateModuleQuery{NewQuery(mods...)}
}

// FindTemplatesTemplateModule retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindTemplatesTemplateModule(exec boil.Executor, iD string, selectCols ...string) (*TemplatesTemplateModule, error) {
	templatesTemplateModuleObj := &TemplatesTemplateModule{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"templates_template_modules\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, templatesTemplateModuleObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from templates_template_modules")
	}

	if err = templatesTemplateModuleObj.doAfterSelectHooks(exec); err != nil {
		return templatesTemplateModuleObj, err
	}

	return templatesTemplateModuleObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *TemplatesTemplateModule) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no templates_template_modules provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(templatesTemplateModuleColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	templatesTemplateModuleInsertCacheMut.RLock()
	cache, cached := templatesTemplateModuleInsertCache[key]
	templatesTemplateModuleInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			templatesTemplateModuleAllColumns,
			templatesTemplateModuleColumnsWithDefault,
			templatesTemplateModuleColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(templatesTemplateModuleType, templatesTemplateModuleMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(templatesTemplateModuleType, templatesTemplateModuleMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"templates_template_modules\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"templates_template_modules\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into templates_template_modules")
	}

	if !cached {
		templatesTemplateModuleInsertCacheMut.Lock()
		templatesTemplateModuleInsertCache[key] = cache
		templatesTemplateModuleInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the TemplatesTemplateModule.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *TemplatesTemplateModule) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	templatesTemplateModuleUpdateCacheMut.RLock()
	cache, cached := templatesTemplateModuleUpdateCache[key]
	templatesTemplateModuleUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			templatesTemplateModuleAllColumns,
			templatesTemplateModulePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update templates_template_modules, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"templates_template_modules\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, templatesTemplateModulePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(templatesTemplateModuleType, templatesTemplateModuleMapping, append(wl, templatesTemplateModulePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update templates_template_modules row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for templates_template_modules")
	}

	if !cached {
		templatesTemplateModuleUpdateCacheMut.Lock()
		templatesTemplateModuleUpdateCache[key] = cache
		templatesTemplateModuleUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q templatesTemplateModuleQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for templates_template_modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for templates_template_modules")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o TemplatesTemplateModuleSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templatesTemplateModulePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"templates_template_modules\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, templatesTemplateModulePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in templatesTemplateModule slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all templatesTemplateModule")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *TemplatesTemplateModule) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no templates_template_modules provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(templatesTemplateModuleColumnsWithDefault, o)

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

	templatesTemplateModuleUpsertCacheMut.RLock()
	cache, cached := templatesTemplateModuleUpsertCache[key]
	templatesTemplateModuleUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			templatesTemplateModuleAllColumns,
			templatesTemplateModuleColumnsWithDefault,
			templatesTemplateModuleColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			templatesTemplateModuleAllColumns,
			templatesTemplateModulePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert templates_template_modules, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(templatesTemplateModulePrimaryKeyColumns))
			copy(conflict, templatesTemplateModulePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"templates_template_modules\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(templatesTemplateModuleType, templatesTemplateModuleMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(templatesTemplateModuleType, templatesTemplateModuleMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert templates_template_modules")
	}

	if !cached {
		templatesTemplateModuleUpsertCacheMut.Lock()
		templatesTemplateModuleUpsertCache[key] = cache
		templatesTemplateModuleUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single TemplatesTemplateModule record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *TemplatesTemplateModule) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no TemplatesTemplateModule provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), templatesTemplateModulePrimaryKeyMapping)
		sql = "DELETE FROM \"templates_template_modules\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"templates_template_modules\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(templatesTemplateModuleType, templatesTemplateModuleMapping, append(wl, templatesTemplateModulePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from templates_template_modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for templates_template_modules")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q templatesTemplateModuleQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no templatesTemplateModuleQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from templates_template_modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for templates_template_modules")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o TemplatesTemplateModuleSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(templatesTemplateModuleBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templatesTemplateModulePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"templates_template_modules\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, templatesTemplateModulePrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templatesTemplateModulePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"templates_template_modules\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, templatesTemplateModulePrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from templatesTemplateModule slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for templates_template_modules")
	}

	if len(templatesTemplateModuleAfterDeleteHooks) != 0 {
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
func (o *TemplatesTemplateModule) Reload(exec boil.Executor) error {
	ret, err := FindTemplatesTemplateModule(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *TemplatesTemplateModuleSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := TemplatesTemplateModuleSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templatesTemplateModulePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"templates_template_modules\".* FROM \"templates_template_modules\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, templatesTemplateModulePrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in TemplatesTemplateModuleSlice")
	}

	*o = slice

	return nil
}

// TemplatesTemplateModuleExists checks if the TemplatesTemplateModule row exists.
func TemplatesTemplateModuleExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"templates_template_modules\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if templates_template_modules exists")
	}

	return exists, nil
}
