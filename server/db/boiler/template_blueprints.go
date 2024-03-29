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

// TemplateBlueprint is an object representing the database table.
type TemplateBlueprint struct {
	ID             string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	TemplateID     string      `boiler:"template_id" boil:"template_id" json:"template_id" toml:"template_id" yaml:"template_id"`
	Type           string      `boiler:"type" boil:"type" json:"type" toml:"type" yaml:"type"`
	BlueprintID    string      `boiler:"blueprint_id" boil:"blueprint_id" json:"blueprint_id" toml:"blueprint_id" yaml:"blueprint_id"`
	CreatedAt      time.Time   `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	BlueprintIDOld null.String `boiler:"blueprint_id_old" boil:"blueprint_id_old" json:"blueprint_id_old,omitempty" toml:"blueprint_id_old" yaml:"blueprint_id_old,omitempty"`
	DeletedAt      null.Time   `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`

	R *templateBlueprintR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L templateBlueprintL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var TemplateBlueprintColumns = struct {
	ID             string
	TemplateID     string
	Type           string
	BlueprintID    string
	CreatedAt      string
	BlueprintIDOld string
	DeletedAt      string
}{
	ID:             "id",
	TemplateID:     "template_id",
	Type:           "type",
	BlueprintID:    "blueprint_id",
	CreatedAt:      "created_at",
	BlueprintIDOld: "blueprint_id_old",
	DeletedAt:      "deleted_at",
}

var TemplateBlueprintTableColumns = struct {
	ID             string
	TemplateID     string
	Type           string
	BlueprintID    string
	CreatedAt      string
	BlueprintIDOld string
	DeletedAt      string
}{
	ID:             "template_blueprints.id",
	TemplateID:     "template_blueprints.template_id",
	Type:           "template_blueprints.type",
	BlueprintID:    "template_blueprints.blueprint_id",
	CreatedAt:      "template_blueprints.created_at",
	BlueprintIDOld: "template_blueprints.blueprint_id_old",
	DeletedAt:      "template_blueprints.deleted_at",
}

// Generated where

var TemplateBlueprintWhere = struct {
	ID             whereHelperstring
	TemplateID     whereHelperstring
	Type           whereHelperstring
	BlueprintID    whereHelperstring
	CreatedAt      whereHelpertime_Time
	BlueprintIDOld whereHelpernull_String
	DeletedAt      whereHelpernull_Time
}{
	ID:             whereHelperstring{field: "\"template_blueprints\".\"id\""},
	TemplateID:     whereHelperstring{field: "\"template_blueprints\".\"template_id\""},
	Type:           whereHelperstring{field: "\"template_blueprints\".\"type\""},
	BlueprintID:    whereHelperstring{field: "\"template_blueprints\".\"blueprint_id\""},
	CreatedAt:      whereHelpertime_Time{field: "\"template_blueprints\".\"created_at\""},
	BlueprintIDOld: whereHelpernull_String{field: "\"template_blueprints\".\"blueprint_id_old\""},
	DeletedAt:      whereHelpernull_Time{field: "\"template_blueprints\".\"deleted_at\""},
}

// TemplateBlueprintRels is where relationship names are stored.
var TemplateBlueprintRels = struct {
	Template string
}{
	Template: "Template",
}

// templateBlueprintR is where relationships are stored.
type templateBlueprintR struct {
	Template *Template `boiler:"Template" boil:"Template" json:"Template" toml:"Template" yaml:"Template"`
}

// NewStruct creates a new relationship struct
func (*templateBlueprintR) NewStruct() *templateBlueprintR {
	return &templateBlueprintR{}
}

// templateBlueprintL is where Load methods for each relationship are stored.
type templateBlueprintL struct{}

var (
	templateBlueprintAllColumns            = []string{"id", "template_id", "type", "blueprint_id", "created_at", "blueprint_id_old", "deleted_at"}
	templateBlueprintColumnsWithoutDefault = []string{"template_id", "type", "blueprint_id"}
	templateBlueprintColumnsWithDefault    = []string{"id", "created_at", "blueprint_id_old", "deleted_at"}
	templateBlueprintPrimaryKeyColumns     = []string{"id"}
	templateBlueprintGeneratedColumns      = []string{}
)

type (
	// TemplateBlueprintSlice is an alias for a slice of pointers to TemplateBlueprint.
	// This should almost always be used instead of []TemplateBlueprint.
	TemplateBlueprintSlice []*TemplateBlueprint
	// TemplateBlueprintHook is the signature for custom TemplateBlueprint hook methods
	TemplateBlueprintHook func(boil.Executor, *TemplateBlueprint) error

	templateBlueprintQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	templateBlueprintType                 = reflect.TypeOf(&TemplateBlueprint{})
	templateBlueprintMapping              = queries.MakeStructMapping(templateBlueprintType)
	templateBlueprintPrimaryKeyMapping, _ = queries.BindMapping(templateBlueprintType, templateBlueprintMapping, templateBlueprintPrimaryKeyColumns)
	templateBlueprintInsertCacheMut       sync.RWMutex
	templateBlueprintInsertCache          = make(map[string]insertCache)
	templateBlueprintUpdateCacheMut       sync.RWMutex
	templateBlueprintUpdateCache          = make(map[string]updateCache)
	templateBlueprintUpsertCacheMut       sync.RWMutex
	templateBlueprintUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var templateBlueprintAfterSelectHooks []TemplateBlueprintHook

var templateBlueprintBeforeInsertHooks []TemplateBlueprintHook
var templateBlueprintAfterInsertHooks []TemplateBlueprintHook

var templateBlueprintBeforeUpdateHooks []TemplateBlueprintHook
var templateBlueprintAfterUpdateHooks []TemplateBlueprintHook

var templateBlueprintBeforeDeleteHooks []TemplateBlueprintHook
var templateBlueprintAfterDeleteHooks []TemplateBlueprintHook

var templateBlueprintBeforeUpsertHooks []TemplateBlueprintHook
var templateBlueprintAfterUpsertHooks []TemplateBlueprintHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *TemplateBlueprint) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBlueprintAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *TemplateBlueprint) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBlueprintBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *TemplateBlueprint) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBlueprintAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *TemplateBlueprint) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBlueprintBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *TemplateBlueprint) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBlueprintAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *TemplateBlueprint) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBlueprintBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *TemplateBlueprint) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBlueprintAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *TemplateBlueprint) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBlueprintBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *TemplateBlueprint) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range templateBlueprintAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddTemplateBlueprintHook registers your hook function for all future operations.
func AddTemplateBlueprintHook(hookPoint boil.HookPoint, templateBlueprintHook TemplateBlueprintHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		templateBlueprintAfterSelectHooks = append(templateBlueprintAfterSelectHooks, templateBlueprintHook)
	case boil.BeforeInsertHook:
		templateBlueprintBeforeInsertHooks = append(templateBlueprintBeforeInsertHooks, templateBlueprintHook)
	case boil.AfterInsertHook:
		templateBlueprintAfterInsertHooks = append(templateBlueprintAfterInsertHooks, templateBlueprintHook)
	case boil.BeforeUpdateHook:
		templateBlueprintBeforeUpdateHooks = append(templateBlueprintBeforeUpdateHooks, templateBlueprintHook)
	case boil.AfterUpdateHook:
		templateBlueprintAfterUpdateHooks = append(templateBlueprintAfterUpdateHooks, templateBlueprintHook)
	case boil.BeforeDeleteHook:
		templateBlueprintBeforeDeleteHooks = append(templateBlueprintBeforeDeleteHooks, templateBlueprintHook)
	case boil.AfterDeleteHook:
		templateBlueprintAfterDeleteHooks = append(templateBlueprintAfterDeleteHooks, templateBlueprintHook)
	case boil.BeforeUpsertHook:
		templateBlueprintBeforeUpsertHooks = append(templateBlueprintBeforeUpsertHooks, templateBlueprintHook)
	case boil.AfterUpsertHook:
		templateBlueprintAfterUpsertHooks = append(templateBlueprintAfterUpsertHooks, templateBlueprintHook)
	}
}

// One returns a single templateBlueprint record from the query.
func (q templateBlueprintQuery) One(exec boil.Executor) (*TemplateBlueprint, error) {
	o := &TemplateBlueprint{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for template_blueprints")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all TemplateBlueprint records from the query.
func (q templateBlueprintQuery) All(exec boil.Executor) (TemplateBlueprintSlice, error) {
	var o []*TemplateBlueprint

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to TemplateBlueprint slice")
	}

	if len(templateBlueprintAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all TemplateBlueprint records in the query.
func (q templateBlueprintQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count template_blueprints rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q templateBlueprintQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if template_blueprints exists")
	}

	return count > 0, nil
}

// Template pointed to by the foreign key.
func (o *TemplateBlueprint) Template(mods ...qm.QueryMod) templateQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.TemplateID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Templates(queryMods...)
	queries.SetFrom(query.Query, "\"templates\"")

	return query
}

// LoadTemplate allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (templateBlueprintL) LoadTemplate(e boil.Executor, singular bool, maybeTemplateBlueprint interface{}, mods queries.Applicator) error {
	var slice []*TemplateBlueprint
	var object *TemplateBlueprint

	if singular {
		object = maybeTemplateBlueprint.(*TemplateBlueprint)
	} else {
		slice = *maybeTemplateBlueprint.(*[]*TemplateBlueprint)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &templateBlueprintR{}
		}
		args = append(args, object.TemplateID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &templateBlueprintR{}
			}

			for _, a := range args {
				if a == obj.TemplateID {
					continue Outer
				}
			}

			args = append(args, obj.TemplateID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`templates`),
		qm.WhereIn(`templates.id in ?`, args...),
		qmhelper.WhereIsNull(`templates.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Template")
	}

	var resultSlice []*Template
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Template")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for templates")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for templates")
	}

	if len(templateBlueprintAfterSelectHooks) != 0 {
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
		object.R.Template = foreign
		if foreign.R == nil {
			foreign.R = &templateR{}
		}
		foreign.R.TemplateBlueprints = append(foreign.R.TemplateBlueprints, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.TemplateID == foreign.ID {
				local.R.Template = foreign
				if foreign.R == nil {
					foreign.R = &templateR{}
				}
				foreign.R.TemplateBlueprints = append(foreign.R.TemplateBlueprints, local)
				break
			}
		}
	}

	return nil
}

// SetTemplate of the templateBlueprint to the related item.
// Sets o.R.Template to related.
// Adds o to related.R.TemplateBlueprints.
func (o *TemplateBlueprint) SetTemplate(exec boil.Executor, insert bool, related *Template) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"template_blueprints\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"template_id"}),
		strmangle.WhereClause("\"", "\"", 2, templateBlueprintPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.TemplateID = related.ID
	if o.R == nil {
		o.R = &templateBlueprintR{
			Template: related,
		}
	} else {
		o.R.Template = related
	}

	if related.R == nil {
		related.R = &templateR{
			TemplateBlueprints: TemplateBlueprintSlice{o},
		}
	} else {
		related.R.TemplateBlueprints = append(related.R.TemplateBlueprints, o)
	}

	return nil
}

// TemplateBlueprints retrieves all the records using an executor.
func TemplateBlueprints(mods ...qm.QueryMod) templateBlueprintQuery {
	mods = append(mods, qm.From("\"template_blueprints\""), qmhelper.WhereIsNull("\"template_blueprints\".\"deleted_at\""))
	return templateBlueprintQuery{NewQuery(mods...)}
}

// FindTemplateBlueprint retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindTemplateBlueprint(exec boil.Executor, iD string, selectCols ...string) (*TemplateBlueprint, error) {
	templateBlueprintObj := &TemplateBlueprint{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"template_blueprints\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, templateBlueprintObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from template_blueprints")
	}

	if err = templateBlueprintObj.doAfterSelectHooks(exec); err != nil {
		return templateBlueprintObj, err
	}

	return templateBlueprintObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *TemplateBlueprint) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no template_blueprints provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(templateBlueprintColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	templateBlueprintInsertCacheMut.RLock()
	cache, cached := templateBlueprintInsertCache[key]
	templateBlueprintInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			templateBlueprintAllColumns,
			templateBlueprintColumnsWithDefault,
			templateBlueprintColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(templateBlueprintType, templateBlueprintMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(templateBlueprintType, templateBlueprintMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"template_blueprints\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"template_blueprints\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into template_blueprints")
	}

	if !cached {
		templateBlueprintInsertCacheMut.Lock()
		templateBlueprintInsertCache[key] = cache
		templateBlueprintInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the TemplateBlueprint.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *TemplateBlueprint) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	templateBlueprintUpdateCacheMut.RLock()
	cache, cached := templateBlueprintUpdateCache[key]
	templateBlueprintUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			templateBlueprintAllColumns,
			templateBlueprintPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update template_blueprints, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"template_blueprints\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, templateBlueprintPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(templateBlueprintType, templateBlueprintMapping, append(wl, templateBlueprintPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update template_blueprints row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for template_blueprints")
	}

	if !cached {
		templateBlueprintUpdateCacheMut.Lock()
		templateBlueprintUpdateCache[key] = cache
		templateBlueprintUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q templateBlueprintQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for template_blueprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for template_blueprints")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o TemplateBlueprintSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templateBlueprintPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"template_blueprints\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, templateBlueprintPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in templateBlueprint slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all templateBlueprint")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *TemplateBlueprint) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no template_blueprints provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(templateBlueprintColumnsWithDefault, o)

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

	templateBlueprintUpsertCacheMut.RLock()
	cache, cached := templateBlueprintUpsertCache[key]
	templateBlueprintUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			templateBlueprintAllColumns,
			templateBlueprintColumnsWithDefault,
			templateBlueprintColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			templateBlueprintAllColumns,
			templateBlueprintPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert template_blueprints, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(templateBlueprintPrimaryKeyColumns))
			copy(conflict, templateBlueprintPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"template_blueprints\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(templateBlueprintType, templateBlueprintMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(templateBlueprintType, templateBlueprintMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert template_blueprints")
	}

	if !cached {
		templateBlueprintUpsertCacheMut.Lock()
		templateBlueprintUpsertCache[key] = cache
		templateBlueprintUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single TemplateBlueprint record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *TemplateBlueprint) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no TemplateBlueprint provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), templateBlueprintPrimaryKeyMapping)
		sql = "DELETE FROM \"template_blueprints\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"template_blueprints\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(templateBlueprintType, templateBlueprintMapping, append(wl, templateBlueprintPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from template_blueprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for template_blueprints")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q templateBlueprintQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no templateBlueprintQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from template_blueprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for template_blueprints")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o TemplateBlueprintSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(templateBlueprintBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templateBlueprintPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"template_blueprints\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, templateBlueprintPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templateBlueprintPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"template_blueprints\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, templateBlueprintPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from templateBlueprint slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for template_blueprints")
	}

	if len(templateBlueprintAfterDeleteHooks) != 0 {
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
func (o *TemplateBlueprint) Reload(exec boil.Executor) error {
	ret, err := FindTemplateBlueprint(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *TemplateBlueprintSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := TemplateBlueprintSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), templateBlueprintPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"template_blueprints\".* FROM \"template_blueprints\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, templateBlueprintPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in TemplateBlueprintSlice")
	}

	*o = slice

	return nil
}

// TemplateBlueprintExists checks if the TemplateBlueprint row exists.
func TemplateBlueprintExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"template_blueprints\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if template_blueprints exists")
	}

	return exists, nil
}
