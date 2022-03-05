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

// Module is an object representing the database table.
type Module struct {
	ID               string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	BrandID          null.String `boiler:"brand_id" boil:"brand_id" json:"brandID,omitempty" toml:"brandID" yaml:"brandID,omitempty"`
	Slug             string      `boiler:"slug" boil:"slug" json:"slug" toml:"slug" yaml:"slug"`
	Label            string      `boiler:"label" boil:"label" json:"label" toml:"label" yaml:"label"`
	HitpointModifier int         `boiler:"hitpoint_modifier" boil:"hitpoint_modifier" json:"hitpointModifier" toml:"hitpointModifier" yaml:"hitpointModifier"`
	ShieldModifier   int         `boiler:"shield_modifier" boil:"shield_modifier" json:"shieldModifier" toml:"shieldModifier" yaml:"shieldModifier"`
	DeletedAt        null.Time   `boiler:"deleted_at" boil:"deleted_at" json:"deletedAt,omitempty" toml:"deletedAt" yaml:"deletedAt,omitempty"`
	UpdatedAt        time.Time   `boiler:"updated_at" boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
	CreatedAt        time.Time   `boiler:"created_at" boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`

	R *moduleR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L moduleL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ModuleColumns = struct {
	ID               string
	BrandID          string
	Slug             string
	Label            string
	HitpointModifier string
	ShieldModifier   string
	DeletedAt        string
	UpdatedAt        string
	CreatedAt        string
}{
	ID:               "id",
	BrandID:          "brand_id",
	Slug:             "slug",
	Label:            "label",
	HitpointModifier: "hitpoint_modifier",
	ShieldModifier:   "shield_modifier",
	DeletedAt:        "deleted_at",
	UpdatedAt:        "updated_at",
	CreatedAt:        "created_at",
}

var ModuleTableColumns = struct {
	ID               string
	BrandID          string
	Slug             string
	Label            string
	HitpointModifier string
	ShieldModifier   string
	DeletedAt        string
	UpdatedAt        string
	CreatedAt        string
}{
	ID:               "modules.id",
	BrandID:          "modules.brand_id",
	Slug:             "modules.slug",
	Label:            "modules.label",
	HitpointModifier: "modules.hitpoint_modifier",
	ShieldModifier:   "modules.shield_modifier",
	DeletedAt:        "modules.deleted_at",
	UpdatedAt:        "modules.updated_at",
	CreatedAt:        "modules.created_at",
}

// Generated where

var ModuleWhere = struct {
	ID               whereHelperstring
	BrandID          whereHelpernull_String
	Slug             whereHelperstring
	Label            whereHelperstring
	HitpointModifier whereHelperint
	ShieldModifier   whereHelperint
	DeletedAt        whereHelpernull_Time
	UpdatedAt        whereHelpertime_Time
	CreatedAt        whereHelpertime_Time
}{
	ID:               whereHelperstring{field: "\"modules\".\"id\""},
	BrandID:          whereHelpernull_String{field: "\"modules\".\"brand_id\""},
	Slug:             whereHelperstring{field: "\"modules\".\"slug\""},
	Label:            whereHelperstring{field: "\"modules\".\"label\""},
	HitpointModifier: whereHelperint{field: "\"modules\".\"hitpoint_modifier\""},
	ShieldModifier:   whereHelperint{field: "\"modules\".\"shield_modifier\""},
	DeletedAt:        whereHelpernull_Time{field: "\"modules\".\"deleted_at\""},
	UpdatedAt:        whereHelpertime_Time{field: "\"modules\".\"updated_at\""},
	CreatedAt:        whereHelpertime_Time{field: "\"modules\".\"created_at\""},
}

// ModuleRels is where relationship names are stored.
var ModuleRels = struct {
	Brand         string
	ChassisModule string
}{
	Brand:         "Brand",
	ChassisModule: "ChassisModule",
}

// moduleR is where relationships are stored.
type moduleR struct {
	Brand         *Brand         `boiler:"Brand" boil:"Brand" json:"Brand" toml:"Brand" yaml:"Brand"`
	ChassisModule *ChassisModule `boiler:"ChassisModule" boil:"ChassisModule" json:"ChassisModule" toml:"ChassisModule" yaml:"ChassisModule"`
}

// NewStruct creates a new relationship struct
func (*moduleR) NewStruct() *moduleR {
	return &moduleR{}
}

// moduleL is where Load methods for each relationship are stored.
type moduleL struct{}

var (
	moduleAllColumns            = []string{"id", "brand_id", "slug", "label", "hitpoint_modifier", "shield_modifier", "deleted_at", "updated_at", "created_at"}
	moduleColumnsWithoutDefault = []string{"slug", "label", "hitpoint_modifier", "shield_modifier"}
	moduleColumnsWithDefault    = []string{"id", "brand_id", "deleted_at", "updated_at", "created_at"}
	modulePrimaryKeyColumns     = []string{"id"}
	moduleGeneratedColumns      = []string{}
)

type (
	// ModuleSlice is an alias for a slice of pointers to Module.
	// This should almost always be used instead of []Module.
	ModuleSlice []*Module
	// ModuleHook is the signature for custom Module hook methods
	ModuleHook func(boil.Executor, *Module) error

	moduleQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	moduleType                 = reflect.TypeOf(&Module{})
	moduleMapping              = queries.MakeStructMapping(moduleType)
	modulePrimaryKeyMapping, _ = queries.BindMapping(moduleType, moduleMapping, modulePrimaryKeyColumns)
	moduleInsertCacheMut       sync.RWMutex
	moduleInsertCache          = make(map[string]insertCache)
	moduleUpdateCacheMut       sync.RWMutex
	moduleUpdateCache          = make(map[string]updateCache)
	moduleUpsertCacheMut       sync.RWMutex
	moduleUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var moduleAfterSelectHooks []ModuleHook

var moduleBeforeInsertHooks []ModuleHook
var moduleAfterInsertHooks []ModuleHook

var moduleBeforeUpdateHooks []ModuleHook
var moduleAfterUpdateHooks []ModuleHook

var moduleBeforeDeleteHooks []ModuleHook
var moduleAfterDeleteHooks []ModuleHook

var moduleBeforeUpsertHooks []ModuleHook
var moduleAfterUpsertHooks []ModuleHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Module) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range moduleAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Module) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range moduleBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Module) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range moduleAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Module) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range moduleBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Module) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range moduleAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Module) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range moduleBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Module) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range moduleAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Module) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range moduleBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Module) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range moduleAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddModuleHook registers your hook function for all future operations.
func AddModuleHook(hookPoint boil.HookPoint, moduleHook ModuleHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		moduleAfterSelectHooks = append(moduleAfterSelectHooks, moduleHook)
	case boil.BeforeInsertHook:
		moduleBeforeInsertHooks = append(moduleBeforeInsertHooks, moduleHook)
	case boil.AfterInsertHook:
		moduleAfterInsertHooks = append(moduleAfterInsertHooks, moduleHook)
	case boil.BeforeUpdateHook:
		moduleBeforeUpdateHooks = append(moduleBeforeUpdateHooks, moduleHook)
	case boil.AfterUpdateHook:
		moduleAfterUpdateHooks = append(moduleAfterUpdateHooks, moduleHook)
	case boil.BeforeDeleteHook:
		moduleBeforeDeleteHooks = append(moduleBeforeDeleteHooks, moduleHook)
	case boil.AfterDeleteHook:
		moduleAfterDeleteHooks = append(moduleAfterDeleteHooks, moduleHook)
	case boil.BeforeUpsertHook:
		moduleBeforeUpsertHooks = append(moduleBeforeUpsertHooks, moduleHook)
	case boil.AfterUpsertHook:
		moduleAfterUpsertHooks = append(moduleAfterUpsertHooks, moduleHook)
	}
}

// One returns a single module record from the query.
func (q moduleQuery) One(exec boil.Executor) (*Module, error) {
	o := &Module{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for modules")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Module records from the query.
func (q moduleQuery) All(exec boil.Executor) (ModuleSlice, error) {
	var o []*Module

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to Module slice")
	}

	if len(moduleAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Module records in the query.
func (q moduleQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count modules rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q moduleQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if modules exists")
	}

	return count > 0, nil
}

// Brand pointed to by the foreign key.
func (o *Module) Brand(mods ...qm.QueryMod) brandQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BrandID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Brands(queryMods...)
	queries.SetFrom(query.Query, "\"brands\"")

	return query
}

// ChassisModule pointed to by the foreign key.
func (o *Module) ChassisModule(mods ...qm.QueryMod) chassisModuleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"module_id\" = ?", o.ID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := ChassisModules(queryMods...)
	queries.SetFrom(query.Query, "\"chassis_modules\"")

	return query
}

// LoadBrand allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (moduleL) LoadBrand(e boil.Executor, singular bool, maybeModule interface{}, mods queries.Applicator) error {
	var slice []*Module
	var object *Module

	if singular {
		object = maybeModule.(*Module)
	} else {
		slice = *maybeModule.(*[]*Module)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &moduleR{}
		}
		if !queries.IsNil(object.BrandID) {
			args = append(args, object.BrandID)
		}

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &moduleR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.BrandID) {
					continue Outer
				}
			}

			if !queries.IsNil(obj.BrandID) {
				args = append(args, obj.BrandID)
			}

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`brands`),
		qm.WhereIn(`brands.id in ?`, args...),
		qmhelper.WhereIsNull(`brands.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Brand")
	}

	var resultSlice []*Brand
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Brand")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for brands")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for brands")
	}

	if len(moduleAfterSelectHooks) != 0 {
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
		object.R.Brand = foreign
		if foreign.R == nil {
			foreign.R = &brandR{}
		}
		foreign.R.Modules = append(foreign.R.Modules, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.BrandID, foreign.ID) {
				local.R.Brand = foreign
				if foreign.R == nil {
					foreign.R = &brandR{}
				}
				foreign.R.Modules = append(foreign.R.Modules, local)
				break
			}
		}
	}

	return nil
}

// LoadChassisModule allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-1 relationship.
func (moduleL) LoadChassisModule(e boil.Executor, singular bool, maybeModule interface{}, mods queries.Applicator) error {
	var slice []*Module
	var object *Module

	if singular {
		object = maybeModule.(*Module)
	} else {
		slice = *maybeModule.(*[]*Module)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &moduleR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &moduleR{}
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
		qm.From(`chassis_modules`),
		qm.WhereIn(`chassis_modules.module_id in ?`, args...),
		qmhelper.WhereIsNull(`chassis_modules.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load ChassisModule")
	}

	var resultSlice []*ChassisModule
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice ChassisModule")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for chassis_modules")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for chassis_modules")
	}

	if len(moduleAfterSelectHooks) != 0 {
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
		object.R.ChassisModule = foreign
		if foreign.R == nil {
			foreign.R = &chassisModuleR{}
		}
		foreign.R.Module = object
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.ID == foreign.ModuleID {
				local.R.ChassisModule = foreign
				if foreign.R == nil {
					foreign.R = &chassisModuleR{}
				}
				foreign.R.Module = local
				break
			}
		}
	}

	return nil
}

// SetBrand of the module to the related item.
// Sets o.R.Brand to related.
// Adds o to related.R.Modules.
func (o *Module) SetBrand(exec boil.Executor, insert bool, related *Brand) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"modules\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"brand_id"}),
		strmangle.WhereClause("\"", "\"", 2, modulePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	queries.Assign(&o.BrandID, related.ID)
	if o.R == nil {
		o.R = &moduleR{
			Brand: related,
		}
	} else {
		o.R.Brand = related
	}

	if related.R == nil {
		related.R = &brandR{
			Modules: ModuleSlice{o},
		}
	} else {
		related.R.Modules = append(related.R.Modules, o)
	}

	return nil
}

// RemoveBrand relationship.
// Sets o.R.Brand to nil.
// Removes o from all passed in related items' relationships struct (Optional).
func (o *Module) RemoveBrand(exec boil.Executor, related *Brand) error {
	var err error

	queries.SetScanner(&o.BrandID, nil)
	if _, err = o.Update(exec, boil.Whitelist("brand_id")); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	if o.R != nil {
		o.R.Brand = nil
	}
	if related == nil || related.R == nil {
		return nil
	}

	for i, ri := range related.R.Modules {
		if queries.Equal(o.BrandID, ri.BrandID) {
			continue
		}

		ln := len(related.R.Modules)
		if ln > 1 && i < ln-1 {
			related.R.Modules[i] = related.R.Modules[ln-1]
		}
		related.R.Modules = related.R.Modules[:ln-1]
		break
	}
	return nil
}

// SetChassisModule of the module to the related item.
// Sets o.R.ChassisModule to related.
// Adds o to related.R.Module.
func (o *Module) SetChassisModule(exec boil.Executor, insert bool, related *ChassisModule) error {
	var err error

	if insert {
		related.ModuleID = o.ID

		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	} else {
		updateQuery := fmt.Sprintf(
			"UPDATE \"chassis_modules\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, []string{"module_id"}),
			strmangle.WhereClause("\"", "\"", 2, chassisModulePrimaryKeyColumns),
		)
		values := []interface{}{o.ID, related.ID}

		if boil.DebugMode {
			fmt.Fprintln(boil.DebugWriter, updateQuery)
			fmt.Fprintln(boil.DebugWriter, values)
		}
		if _, err = exec.Exec(updateQuery, values...); err != nil {
			return errors.Wrap(err, "failed to update foreign table")
		}

		related.ModuleID = o.ID

	}

	if o.R == nil {
		o.R = &moduleR{
			ChassisModule: related,
		}
	} else {
		o.R.ChassisModule = related
	}

	if related.R == nil {
		related.R = &chassisModuleR{
			Module: o,
		}
	} else {
		related.R.Module = o
	}
	return nil
}

// Modules retrieves all the records using an executor.
func Modules(mods ...qm.QueryMod) moduleQuery {
	mods = append(mods, qm.From("\"modules\""), qmhelper.WhereIsNull("\"modules\".\"deleted_at\""))
	return moduleQuery{NewQuery(mods...)}
}

// FindModule retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindModule(exec boil.Executor, iD string, selectCols ...string) (*Module, error) {
	moduleObj := &Module{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"modules\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, moduleObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from modules")
	}

	if err = moduleObj.doAfterSelectHooks(exec); err != nil {
		return moduleObj, err
	}

	return moduleObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Module) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no modules provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(moduleColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	moduleInsertCacheMut.RLock()
	cache, cached := moduleInsertCache[key]
	moduleInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			moduleAllColumns,
			moduleColumnsWithDefault,
			moduleColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(moduleType, moduleMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(moduleType, moduleMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"modules\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"modules\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into modules")
	}

	if !cached {
		moduleInsertCacheMut.Lock()
		moduleInsertCache[key] = cache
		moduleInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the Module.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Module) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	moduleUpdateCacheMut.RLock()
	cache, cached := moduleUpdateCache[key]
	moduleUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			moduleAllColumns,
			modulePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update modules, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"modules\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, modulePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(moduleType, moduleMapping, append(wl, modulePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update modules row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for modules")
	}

	if !cached {
		moduleUpdateCacheMut.Lock()
		moduleUpdateCache[key] = cache
		moduleUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q moduleQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for modules")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ModuleSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), modulePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"modules\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, modulePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in module slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all module")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Module) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no modules provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(moduleColumnsWithDefault, o)

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

	moduleUpsertCacheMut.RLock()
	cache, cached := moduleUpsertCache[key]
	moduleUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			moduleAllColumns,
			moduleColumnsWithDefault,
			moduleColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			moduleAllColumns,
			modulePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert modules, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(modulePrimaryKeyColumns))
			copy(conflict, modulePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"modules\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(moduleType, moduleMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(moduleType, moduleMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert modules")
	}

	if !cached {
		moduleUpsertCacheMut.Lock()
		moduleUpsertCache[key] = cache
		moduleUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single Module record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Module) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no Module provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), modulePrimaryKeyMapping)
		sql = "DELETE FROM \"modules\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"modules\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(moduleType, moduleMapping, append(wl, modulePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for modules")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q moduleQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no moduleQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for modules")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ModuleSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(moduleBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), modulePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"modules\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, modulePrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), modulePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"modules\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, modulePrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from module slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for modules")
	}

	if len(moduleAfterDeleteHooks) != 0 {
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
func (o *Module) Reload(exec boil.Executor) error {
	ret, err := FindModule(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ModuleSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ModuleSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), modulePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"modules\".* FROM \"modules\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, modulePrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in ModuleSlice")
	}

	*o = slice

	return nil
}

// ModuleExists checks if the Module row exists.
func ModuleExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"modules\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if modules exists")
	}

	return exists, nil
}
