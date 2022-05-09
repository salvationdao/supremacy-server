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

// Template is an object representing the database table.
type Template struct {
	ID        string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Label     string    `boiler:"label" boil:"label" json:"label" toml:"label" yaml:"label"`
	DeletedAt null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	UpdatedAt time.Time `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *templateR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L templateL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var TemplateColumns = struct {
	ID        string
	Label     string
	DeletedAt string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "id",
	Label:     "label",
	DeletedAt: "deleted_at",
	UpdatedAt: "updated_at",
	CreatedAt: "created_at",
}

var TemplateTableColumns = struct {
	ID        string
	Label     string
	DeletedAt string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "templates.id",
	Label:     "templates.label",
	DeletedAt: "templates.deleted_at",
	UpdatedAt: "templates.updated_at",
	CreatedAt: "templates.created_at",
}

// Generated where

var TemplateWhere = struct {
	ID        whereHelperstring
	Label     whereHelperstring
	DeletedAt whereHelpernull_Time
	UpdatedAt whereHelpertime_Time
	CreatedAt whereHelpertime_Time
}{
	ID:        whereHelperstring{field: "\"templates\".\"id\""},
	Label:     whereHelperstring{field: "\"templates\".\"label\""},
	DeletedAt: whereHelpernull_Time{field: "\"templates\".\"deleted_at\""},
	UpdatedAt: whereHelpertime_Time{field: "\"templates\".\"updated_at\""},
	CreatedAt: whereHelpertime_Time{field: "\"templates\".\"created_at\""},
}

// TemplateRels is where relationship names are stored.
var TemplateRels = struct {
	TemplateBlueprints string
}{
	TemplateBlueprints: "TemplateBlueprints",
}

// templateR is where relationships are stored.
type templateR struct {
	TemplateBlueprints TemplateBlueprintSlice `boiler:"TemplateBlueprints" boil:"TemplateBlueprints" json:"TemplateBlueprints" toml:"TemplateBlueprints" yaml:"TemplateBlueprints"`
}

// NewStruct creates a new relationship struct
func (*templateR) NewStruct() *templateR {
	return &templateR{}
}

// templateL is where Load methods for each relationship are stored.
type templateL struct{}

var (
	templateAllColumns            = []string{"id", "label", "deleted_at", "updated_at", "created_at"}
	templateColumnsWithoutDefault = []string{"label"}
	templateColumnsWithDefault    = []string{"id", "deleted_at", "updated_at", "created_at"}
	templatePrimaryKeyColumns     = []string{"id"}
	templateGeneratedColumns      = []string{}
)

type (
	// TemplateSlice is an alias for a slice of pointers to Template.
	// This should almost always be used instead of []Template.
	TemplateSlice []*Template
	// TemplateHook is the signature for custom Template hook methods
	TemplateHook func(boil.Executor, *Template) error

	templateQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	templateType                 = reflect.TypeOf(&Template{})
	templateMapping              = queries.MakeStructMapping(templateType)
	templatePrimaryKeyMapping, _ = queries.BindMapping(templateType, templateMapping, templatePrimaryKeyColumns)
	templateInsertCacheMut       sync.RWMutex
	templateInsertCache          = make(map[string]insertCache)
	templateUpdateCacheMut       sync.RWMutex
	templateUpdateCache          = make(map[string]updateCache)
	templateUpsertCacheMut       sync.RWMutex
	templateUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var templateAfterSelectHooks []TemplateHook

var templateBeforeInsertHooks []TemplateHook
var templateAfterInsertHooks []TemplateHook

var templateBeforeUpdateHooks []TemplateHook
var templateAfterUpdateHooks []TemplateHook

var templateBeforeDeleteHooks []TemplateHook
var templateAfterDeleteHooks []TemplateHook

var templateBeforeUpsertHooks []TemplateHook
var templateAfterUpsertHooks []TemplateHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Template) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range templateAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Template) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Template) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Template) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Template) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range templateAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Template) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Template) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range templateAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Template) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Template) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddTemplateHook registers your hook function for all future operations.
func AddTemplateHook(hookPoint boil.HookPoint, templateHook TemplateHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		templateAfterSelectHooks = append(templateAfterSelectHooks, templateHook)
	case boil.BeforeInsertHook:
		templateBeforeInsertHooks = append(templateBeforeInsertHooks, templateHook)
	case boil.AfterInsertHook:
		templateAfterInsertHooks = append(templateAfterInsertHooks, templateHook)
	case boil.BeforeUpdateHook:
		templateBeforeUpdateHooks = append(templateBeforeUpdateHooks, templateHook)
	case boil.AfterUpdateHook:
		templateAfterUpdateHooks = append(templateAfterUpdateHooks, templateHook)
	case boil.BeforeDeleteHook:
		templateBeforeDeleteHooks = append(templateBeforeDeleteHooks, templateHook)
	case boil.AfterDeleteHook:
		templateAfterDeleteHooks = append(templateAfterDeleteHooks, templateHook)
	case boil.BeforeUpsertHook:
		templateBeforeUpsertHooks = append(templateBeforeUpsertHooks, templateHook)
	case boil.AfterUpsertHook:
		templateAfterUpsertHooks = append(templateAfterUpsertHooks, templateHook)
	}
}

// One returns a single template record from the query.
func (q templateQuery) One(exec boil.Executor) (*Template, error) {
	o := &Template{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for templates")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Template records from the query.
func (q templateQuery) All(exec boil.Executor) (TemplateSlice, error) {
	var o []*Template

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to Template slice")
	}

	if len(templateAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Template records in the query.
func (q templateQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count templates rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q templateQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if templates exists")
	}

	return count > 0, nil
}

// TemplateBlueprints retrieves all the template_blueprint's TemplateBlueprints with an executor.
func (o *Template) TemplateBlueprints(mods ...qm.QueryMod) templateBlueprintQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"template_blueprints\".\"template_id\"=?", o.ID),
	)

	query := TemplateBlueprints(queryMods...)
	queries.SetFrom(query.Query, "\"template_blueprints\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"template_blueprints\".*"})
	}

	return query
}

// LoadTemplateBlueprints allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (templateL) LoadTemplateBlueprints(e boil.Executor, singular bool, maybeTemplate interface{}, mods queries.Applicator) error {
	var slice []*Template
	var object *Template

	if singular {
		object = maybeTemplate.(*Template)
	} else {
		slice = *maybeTemplate.(*[]*Template)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &templateR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &templateR{}
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
		qm.From(`template_blueprints`),
		qm.WhereIn(`template_blueprints.template_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load template_blueprints")
	}

	var resultSlice []*TemplateBlueprint
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice template_blueprints")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on template_blueprints")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for template_blueprints")
	}

	if len(templateBlueprintAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.TemplateBlueprints = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &templateBlueprintR{}
			}
			foreign.R.Template = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.TemplateID {
				local.R.TemplateBlueprints = append(local.R.TemplateBlueprints, foreign)
				if foreign.R == nil {
					foreign.R = &templateBlueprintR{}
				}
				foreign.R.Template = local
				break
			}
		}
	}

	return nil
}

// AddTemplateBlueprints adds the given related objects to the existing relationships
// of the template, optionally inserting them as new records.
// Appends related to o.R.TemplateBlueprints.
// Sets related.R.Template appropriately.
func (o *Template) AddTemplateBlueprints(exec boil.Executor, insert bool, related ...*TemplateBlueprint) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.TemplateID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"template_blueprints\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"template_id"}),
				strmangle.WhereClause("\"", "\"", 2, templateBlueprintPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.TemplateID = o.ID
		}
	}

	if o.R == nil {
		o.R = &templateR{
			TemplateBlueprints: related,
		}
	} else {
		o.R.TemplateBlueprints = append(o.R.TemplateBlueprints, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &templateBlueprintR{
				Template: o,
			}
		} else {
			rel.R.Template = o
		}
	}
	return nil
}

// Templates retrieves all the records using an executor.
func Templates(mods ...qm.QueryMod) templateQuery {
	mods = append(mods, qm.From("\"templates\""), qmhelper.WhereIsNull("\"templates\".\"deleted_at\""))
	return templateQuery{NewQuery(mods...)}
}

// FindTemplate retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindTemplate(exec boil.Executor, iD string, selectCols ...string) (*Template, error) {
	templateObj := &Template{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"templates\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, templateObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from templates")
	}

	if err = templateObj.doAfterSelectHooks(exec); err != nil {
		return templateObj, err
	}

	return templateObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Template) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no templates provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(templateColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	templateInsertCacheMut.RLock()
	cache, cached := templateInsertCache[key]
	templateInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			templateAllColumns,
			templateColumnsWithDefault,
			templateColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(templateType, templateMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(templateType, templateMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"templates\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"templates\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into templates")
	}

	if !cached {
		templateInsertCacheMut.Lock()
		templateInsertCache[key] = cache
		templateInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the Template.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Template) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	templateUpdateCacheMut.RLock()
	cache, cached := templateUpdateCache[key]
	templateUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			templateAllColumns,
			templatePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update templates, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"templates\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, templatePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(templateType, templateMapping, append(wl, templatePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update templates row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for templates")
	}

	if !cached {
		templateUpdateCacheMut.Lock()
		templateUpdateCache[key] = cache
		templateUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q templateQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for templates")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for templates")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o TemplateSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templatePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"templates\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, templatePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in template slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all template")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Template) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no templates provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(templateColumnsWithDefault, o)

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

	templateUpsertCacheMut.RLock()
	cache, cached := templateUpsertCache[key]
	templateUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			templateAllColumns,
			templateColumnsWithDefault,
			templateColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			templateAllColumns,
			templatePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert templates, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(templatePrimaryKeyColumns))
			copy(conflict, templatePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"templates\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(templateType, templateMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(templateType, templateMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert templates")
	}

	if !cached {
		templateUpsertCacheMut.Lock()
		templateUpsertCache[key] = cache
		templateUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single Template record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Template) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no Template provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), templatePrimaryKeyMapping)
		sql = "DELETE FROM \"templates\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"templates\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(templateType, templateMapping, append(wl, templatePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from templates")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for templates")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q templateQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no templateQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from templates")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for templates")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o TemplateSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(templateBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templatePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"templates\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, templatePrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templatePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"templates\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, templatePrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from template slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for templates")
	}

	if len(templateAfterDeleteHooks) != 0 {
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
func (o *Template) Reload(exec boil.Executor) error {
	ret, err := FindTemplate(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *TemplateSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := TemplateSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templatePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"templates\".* FROM \"templates\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, templatePrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in TemplateSlice")
	}

	*o = slice

	return nil
}

// TemplateExists checks if the Template row exists.
func TemplateExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"templates\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if templates exists")
	}

	return exists, nil
}
