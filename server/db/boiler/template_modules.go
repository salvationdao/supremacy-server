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

// TemplateModule is an object representing the database table.
type TemplateModule struct {
	ID             string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Slug           string    `boiler:"slug" boil:"slug" json:"slug" toml:"slug" yaml:"slug"`
	Label          string    `boiler:"label" boil:"label" json:"label" toml:"label" yaml:"label"`
	HPModifier     int       `boiler:"hp_modifier" boil:"hp_modifier" json:"hpModifier" toml:"hpModifier" yaml:"hpModifier"`
	ShieldModifier int       `boiler:"shield_modifier" boil:"shield_modifier" json:"shieldModifier" toml:"shieldModifier" yaml:"shieldModifier"`
	DeletedAt      null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deletedAt,omitempty" toml:"deletedAt" yaml:"deletedAt,omitempty"`
	UpdatedAt      time.Time `boiler:"updated_at" boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
	CreatedAt      time.Time `boiler:"created_at" boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`

	R *templateModuleR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L templateModuleL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var TemplateModuleColumns = struct {
	ID             string
	Slug           string
	Label          string
	HPModifier     string
	ShieldModifier string
	DeletedAt      string
	UpdatedAt      string
	CreatedAt      string
}{
	ID:             "id",
	Slug:           "slug",
	Label:          "label",
	HPModifier:     "hp_modifier",
	ShieldModifier: "shield_modifier",
	DeletedAt:      "deleted_at",
	UpdatedAt:      "updated_at",
	CreatedAt:      "created_at",
}

var TemplateModuleTableColumns = struct {
	ID             string
	Slug           string
	Label          string
	HPModifier     string
	ShieldModifier string
	DeletedAt      string
	UpdatedAt      string
	CreatedAt      string
}{
	ID:             "template_modules.id",
	Slug:           "template_modules.slug",
	Label:          "template_modules.label",
	HPModifier:     "template_modules.hp_modifier",
	ShieldModifier: "template_modules.shield_modifier",
	DeletedAt:      "template_modules.deleted_at",
	UpdatedAt:      "template_modules.updated_at",
	CreatedAt:      "template_modules.created_at",
}

// Generated where

var TemplateModuleWhere = struct {
	ID             whereHelperstring
	Slug           whereHelperstring
	Label          whereHelperstring
	HPModifier     whereHelperint
	ShieldModifier whereHelperint
	DeletedAt      whereHelpernull_Time
	UpdatedAt      whereHelpertime_Time
	CreatedAt      whereHelpertime_Time
}{
	ID:             whereHelperstring{field: "\"template_modules\".\"id\""},
	Slug:           whereHelperstring{field: "\"template_modules\".\"slug\""},
	Label:          whereHelperstring{field: "\"template_modules\".\"label\""},
	HPModifier:     whereHelperint{field: "\"template_modules\".\"hp_modifier\""},
	ShieldModifier: whereHelperint{field: "\"template_modules\".\"shield_modifier\""},
	DeletedAt:      whereHelpernull_Time{field: "\"template_modules\".\"deleted_at\""},
	UpdatedAt:      whereHelpertime_Time{field: "\"template_modules\".\"updated_at\""},
	CreatedAt:      whereHelpertime_Time{field: "\"template_modules\".\"created_at\""},
}

// TemplateModuleRels is where relationship names are stored.
var TemplateModuleRels = struct {
}{}

// templateModuleR is where relationships are stored.
type templateModuleR struct {
}

// NewStruct creates a new relationship struct
func (*templateModuleR) NewStruct() *templateModuleR {
	return &templateModuleR{}
}

// templateModuleL is where Load methods for each relationship are stored.
type templateModuleL struct{}

var (
	templateModuleAllColumns            = []string{"id", "slug", "label", "hp_modifier", "shield_modifier", "deleted_at", "updated_at", "created_at"}
	templateModuleColumnsWithoutDefault = []string{"slug", "label", "hp_modifier", "shield_modifier"}
	templateModuleColumnsWithDefault    = []string{"id", "deleted_at", "updated_at", "created_at"}
	templateModulePrimaryKeyColumns     = []string{"id"}
	templateModuleGeneratedColumns      = []string{}
)

type (
	// TemplateModuleSlice is an alias for a slice of pointers to TemplateModule.
	// This should almost always be used instead of []TemplateModule.
	TemplateModuleSlice []*TemplateModule
	// TemplateModuleHook is the signature for custom TemplateModule hook methods
	TemplateModuleHook func(boil.Executor, *TemplateModule) error

	templateModuleQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	templateModuleType                 = reflect.TypeOf(&TemplateModule{})
	templateModuleMapping              = queries.MakeStructMapping(templateModuleType)
	templateModulePrimaryKeyMapping, _ = queries.BindMapping(templateModuleType, templateModuleMapping, templateModulePrimaryKeyColumns)
	templateModuleInsertCacheMut       sync.RWMutex
	templateModuleInsertCache          = make(map[string]insertCache)
	templateModuleUpdateCacheMut       sync.RWMutex
	templateModuleUpdateCache          = make(map[string]updateCache)
	templateModuleUpsertCacheMut       sync.RWMutex
	templateModuleUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var templateModuleAfterSelectHooks []TemplateModuleHook

var templateModuleBeforeInsertHooks []TemplateModuleHook
var templateModuleAfterInsertHooks []TemplateModuleHook

var templateModuleBeforeUpdateHooks []TemplateModuleHook
var templateModuleAfterUpdateHooks []TemplateModuleHook

var templateModuleBeforeDeleteHooks []TemplateModuleHook
var templateModuleAfterDeleteHooks []TemplateModuleHook

var templateModuleBeforeUpsertHooks []TemplateModuleHook
var templateModuleAfterUpsertHooks []TemplateModuleHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *TemplateModule) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range templateModuleAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *TemplateModule) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateModuleBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *TemplateModule) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateModuleAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *TemplateModule) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range templateModuleBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *TemplateModule) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range templateModuleAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *TemplateModule) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range templateModuleBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *TemplateModule) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range templateModuleAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *TemplateModule) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateModuleBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *TemplateModule) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateModuleAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddTemplateModuleHook registers your hook function for all future operations.
func AddTemplateModuleHook(hookPoint boil.HookPoint, templateModuleHook TemplateModuleHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		templateModuleAfterSelectHooks = append(templateModuleAfterSelectHooks, templateModuleHook)
	case boil.BeforeInsertHook:
		templateModuleBeforeInsertHooks = append(templateModuleBeforeInsertHooks, templateModuleHook)
	case boil.AfterInsertHook:
		templateModuleAfterInsertHooks = append(templateModuleAfterInsertHooks, templateModuleHook)
	case boil.BeforeUpdateHook:
		templateModuleBeforeUpdateHooks = append(templateModuleBeforeUpdateHooks, templateModuleHook)
	case boil.AfterUpdateHook:
		templateModuleAfterUpdateHooks = append(templateModuleAfterUpdateHooks, templateModuleHook)
	case boil.BeforeDeleteHook:
		templateModuleBeforeDeleteHooks = append(templateModuleBeforeDeleteHooks, templateModuleHook)
	case boil.AfterDeleteHook:
		templateModuleAfterDeleteHooks = append(templateModuleAfterDeleteHooks, templateModuleHook)
	case boil.BeforeUpsertHook:
		templateModuleBeforeUpsertHooks = append(templateModuleBeforeUpsertHooks, templateModuleHook)
	case boil.AfterUpsertHook:
		templateModuleAfterUpsertHooks = append(templateModuleAfterUpsertHooks, templateModuleHook)
	}
}

// One returns a single templateModule record from the query.
func (q templateModuleQuery) One(exec boil.Executor) (*TemplateModule, error) {
	o := &TemplateModule{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for template_modules")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all TemplateModule records from the query.
func (q templateModuleQuery) All(exec boil.Executor) (TemplateModuleSlice, error) {
	var o []*TemplateModule

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to TemplateModule slice")
	}

	if len(templateModuleAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all TemplateModule records in the query.
func (q templateModuleQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count template_modules rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q templateModuleQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if template_modules exists")
	}

	return count > 0, nil
}

// TemplateModules retrieves all the records using an executor.
func TemplateModules(mods ...qm.QueryMod) templateModuleQuery {
	mods = append(mods, qm.From("\"template_modules\""), qmhelper.WhereIsNull("\"template_modules\".\"deleted_at\""))
	return templateModuleQuery{NewQuery(mods...)}
}

// FindTemplateModule retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindTemplateModule(exec boil.Executor, iD string, selectCols ...string) (*TemplateModule, error) {
	templateModuleObj := &TemplateModule{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"template_modules\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, templateModuleObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from template_modules")
	}

	if err = templateModuleObj.doAfterSelectHooks(exec); err != nil {
		return templateModuleObj, err
	}

	return templateModuleObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *TemplateModule) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no template_modules provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(templateModuleColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	templateModuleInsertCacheMut.RLock()
	cache, cached := templateModuleInsertCache[key]
	templateModuleInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			templateModuleAllColumns,
			templateModuleColumnsWithDefault,
			templateModuleColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(templateModuleType, templateModuleMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(templateModuleType, templateModuleMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"template_modules\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"template_modules\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into template_modules")
	}

	if !cached {
		templateModuleInsertCacheMut.Lock()
		templateModuleInsertCache[key] = cache
		templateModuleInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the TemplateModule.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *TemplateModule) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	templateModuleUpdateCacheMut.RLock()
	cache, cached := templateModuleUpdateCache[key]
	templateModuleUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			templateModuleAllColumns,
			templateModulePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update template_modules, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"template_modules\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, templateModulePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(templateModuleType, templateModuleMapping, append(wl, templateModulePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update template_modules row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for template_modules")
	}

	if !cached {
		templateModuleUpdateCacheMut.Lock()
		templateModuleUpdateCache[key] = cache
		templateModuleUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q templateModuleQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for template_modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for template_modules")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o TemplateModuleSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templateModulePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"template_modules\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, templateModulePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in templateModule slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all templateModule")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *TemplateModule) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no template_modules provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(templateModuleColumnsWithDefault, o)

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

	templateModuleUpsertCacheMut.RLock()
	cache, cached := templateModuleUpsertCache[key]
	templateModuleUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			templateModuleAllColumns,
			templateModuleColumnsWithDefault,
			templateModuleColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			templateModuleAllColumns,
			templateModulePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert template_modules, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(templateModulePrimaryKeyColumns))
			copy(conflict, templateModulePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"template_modules\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(templateModuleType, templateModuleMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(templateModuleType, templateModuleMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert template_modules")
	}

	if !cached {
		templateModuleUpsertCacheMut.Lock()
		templateModuleUpsertCache[key] = cache
		templateModuleUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single TemplateModule record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *TemplateModule) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no TemplateModule provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), templateModulePrimaryKeyMapping)
		sql = "DELETE FROM \"template_modules\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"template_modules\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(templateModuleType, templateModuleMapping, append(wl, templateModulePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from template_modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for template_modules")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q templateModuleQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no templateModuleQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from template_modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for template_modules")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o TemplateModuleSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(templateModuleBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templateModulePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"template_modules\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, templateModulePrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templateModulePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"template_modules\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, templateModulePrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from templateModule slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for template_modules")
	}

	if len(templateModuleAfterDeleteHooks) != 0 {
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
func (o *TemplateModule) Reload(exec boil.Executor) error {
	ret, err := FindTemplateModule(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *TemplateModuleSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := TemplateModuleSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templateModulePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"template_modules\".* FROM \"template_modules\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, templateModulePrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in TemplateModuleSlice")
	}

	*o = slice

	return nil
}

// TemplateModuleExists checks if the TemplateModule row exists.
func TemplateModuleExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"template_modules\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if template_modules exists")
	}

	return exists, nil
}
