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

// BlueprintModule is an object representing the database table.
type BlueprintModule struct {
	ID               string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	BrandID          null.String `boiler:"brand_id" boil:"brand_id" json:"brand_id,omitempty" toml:"brand_id" yaml:"brand_id,omitempty"`
	Slug             string      `boiler:"slug" boil:"slug" json:"slug" toml:"slug" yaml:"slug"`
	Label            string      `boiler:"label" boil:"label" json:"label" toml:"label" yaml:"label"`
	HitpointModifier int         `boiler:"hitpoint_modifier" boil:"hitpoint_modifier" json:"hitpoint_modifier" toml:"hitpoint_modifier" yaml:"hitpoint_modifier"`
	ShieldModifier   int         `boiler:"shield_modifier" boil:"shield_modifier" json:"shield_modifier" toml:"shield_modifier" yaml:"shield_modifier"`
	DeletedAt        null.Time   `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	UpdatedAt        time.Time   `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt        time.Time   `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *blueprintModuleR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L blueprintModuleL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var BlueprintModuleColumns = struct {
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

var BlueprintModuleTableColumns = struct {
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
	ID:               "blueprint_modules.id",
	BrandID:          "blueprint_modules.brand_id",
	Slug:             "blueprint_modules.slug",
	Label:            "blueprint_modules.label",
	HitpointModifier: "blueprint_modules.hitpoint_modifier",
	ShieldModifier:   "blueprint_modules.shield_modifier",
	DeletedAt:        "blueprint_modules.deleted_at",
	UpdatedAt:        "blueprint_modules.updated_at",
	CreatedAt:        "blueprint_modules.created_at",
}

// Generated where

var BlueprintModuleWhere = struct {
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
	ID:               whereHelperstring{field: "\"blueprint_modules\".\"id\""},
	BrandID:          whereHelpernull_String{field: "\"blueprint_modules\".\"brand_id\""},
	Slug:             whereHelperstring{field: "\"blueprint_modules\".\"slug\""},
	Label:            whereHelperstring{field: "\"blueprint_modules\".\"label\""},
	HitpointModifier: whereHelperint{field: "\"blueprint_modules\".\"hitpoint_modifier\""},
	ShieldModifier:   whereHelperint{field: "\"blueprint_modules\".\"shield_modifier\""},
	DeletedAt:        whereHelpernull_Time{field: "\"blueprint_modules\".\"deleted_at\""},
	UpdatedAt:        whereHelpertime_Time{field: "\"blueprint_modules\".\"updated_at\""},
	CreatedAt:        whereHelpertime_Time{field: "\"blueprint_modules\".\"created_at\""},
}

// BlueprintModuleRels is where relationship names are stored.
var BlueprintModuleRels = struct {
	Brand string
}{
	Brand: "Brand",
}

// blueprintModuleR is where relationships are stored.
type blueprintModuleR struct {
	Brand *Brand `boiler:"Brand" boil:"Brand" json:"Brand" toml:"Brand" yaml:"Brand"`
}

// NewStruct creates a new relationship struct
func (*blueprintModuleR) NewStruct() *blueprintModuleR {
	return &blueprintModuleR{}
}

// blueprintModuleL is where Load methods for each relationship are stored.
type blueprintModuleL struct{}

var (
	blueprintModuleAllColumns            = []string{"id", "brand_id", "slug", "label", "hitpoint_modifier", "shield_modifier", "deleted_at", "updated_at", "created_at"}
	blueprintModuleColumnsWithoutDefault = []string{"slug", "label", "hitpoint_modifier", "shield_modifier"}
	blueprintModuleColumnsWithDefault    = []string{"id", "brand_id", "deleted_at", "updated_at", "created_at"}
	blueprintModulePrimaryKeyColumns     = []string{"id"}
	blueprintModuleGeneratedColumns      = []string{}
)

type (
	// BlueprintModuleSlice is an alias for a slice of pointers to BlueprintModule.
	// This should almost always be used instead of []BlueprintModule.
	BlueprintModuleSlice []*BlueprintModule
	// BlueprintModuleHook is the signature for custom BlueprintModule hook methods
	BlueprintModuleHook func(boil.Executor, *BlueprintModule) error

	blueprintModuleQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	blueprintModuleType                 = reflect.TypeOf(&BlueprintModule{})
	blueprintModuleMapping              = queries.MakeStructMapping(blueprintModuleType)
	blueprintModulePrimaryKeyMapping, _ = queries.BindMapping(blueprintModuleType, blueprintModuleMapping, blueprintModulePrimaryKeyColumns)
	blueprintModuleInsertCacheMut       sync.RWMutex
	blueprintModuleInsertCache          = make(map[string]insertCache)
	blueprintModuleUpdateCacheMut       sync.RWMutex
	blueprintModuleUpdateCache          = make(map[string]updateCache)
	blueprintModuleUpsertCacheMut       sync.RWMutex
	blueprintModuleUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var blueprintModuleAfterSelectHooks []BlueprintModuleHook

var blueprintModuleBeforeInsertHooks []BlueprintModuleHook
var blueprintModuleAfterInsertHooks []BlueprintModuleHook

var blueprintModuleBeforeUpdateHooks []BlueprintModuleHook
var blueprintModuleAfterUpdateHooks []BlueprintModuleHook

var blueprintModuleBeforeDeleteHooks []BlueprintModuleHook
var blueprintModuleAfterDeleteHooks []BlueprintModuleHook

var blueprintModuleBeforeUpsertHooks []BlueprintModuleHook
var blueprintModuleAfterUpsertHooks []BlueprintModuleHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *BlueprintModule) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range blueprintModuleAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *BlueprintModule) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range blueprintModuleBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *BlueprintModule) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range blueprintModuleAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *BlueprintModule) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range blueprintModuleBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *BlueprintModule) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range blueprintModuleAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *BlueprintModule) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range blueprintModuleBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *BlueprintModule) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range blueprintModuleAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *BlueprintModule) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range blueprintModuleBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *BlueprintModule) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range blueprintModuleAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddBlueprintModuleHook registers your hook function for all future operations.
func AddBlueprintModuleHook(hookPoint boil.HookPoint, blueprintModuleHook BlueprintModuleHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		blueprintModuleAfterSelectHooks = append(blueprintModuleAfterSelectHooks, blueprintModuleHook)
	case boil.BeforeInsertHook:
		blueprintModuleBeforeInsertHooks = append(blueprintModuleBeforeInsertHooks, blueprintModuleHook)
	case boil.AfterInsertHook:
		blueprintModuleAfterInsertHooks = append(blueprintModuleAfterInsertHooks, blueprintModuleHook)
	case boil.BeforeUpdateHook:
		blueprintModuleBeforeUpdateHooks = append(blueprintModuleBeforeUpdateHooks, blueprintModuleHook)
	case boil.AfterUpdateHook:
		blueprintModuleAfterUpdateHooks = append(blueprintModuleAfterUpdateHooks, blueprintModuleHook)
	case boil.BeforeDeleteHook:
		blueprintModuleBeforeDeleteHooks = append(blueprintModuleBeforeDeleteHooks, blueprintModuleHook)
	case boil.AfterDeleteHook:
		blueprintModuleAfterDeleteHooks = append(blueprintModuleAfterDeleteHooks, blueprintModuleHook)
	case boil.BeforeUpsertHook:
		blueprintModuleBeforeUpsertHooks = append(blueprintModuleBeforeUpsertHooks, blueprintModuleHook)
	case boil.AfterUpsertHook:
		blueprintModuleAfterUpsertHooks = append(blueprintModuleAfterUpsertHooks, blueprintModuleHook)
	}
}

// One returns a single blueprintModule record from the query.
func (q blueprintModuleQuery) One(exec boil.Executor) (*BlueprintModule, error) {
	o := &BlueprintModule{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for blueprint_modules")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all BlueprintModule records from the query.
func (q blueprintModuleQuery) All(exec boil.Executor) (BlueprintModuleSlice, error) {
	var o []*BlueprintModule

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to BlueprintModule slice")
	}

	if len(blueprintModuleAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all BlueprintModule records in the query.
func (q blueprintModuleQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count blueprint_modules rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q blueprintModuleQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if blueprint_modules exists")
	}

	return count > 0, nil
}

// Brand pointed to by the foreign key.
func (o *BlueprintModule) Brand(mods ...qm.QueryMod) brandQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BrandID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Brands(queryMods...)
	queries.SetFrom(query.Query, "\"brands\"")

	return query
}

// LoadBrand allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (blueprintModuleL) LoadBrand(e boil.Executor, singular bool, maybeBlueprintModule interface{}, mods queries.Applicator) error {
	var slice []*BlueprintModule
	var object *BlueprintModule

	if singular {
		object = maybeBlueprintModule.(*BlueprintModule)
	} else {
		slice = *maybeBlueprintModule.(*[]*BlueprintModule)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &blueprintModuleR{}
		}
		if !queries.IsNil(object.BrandID) {
			args = append(args, object.BrandID)
		}

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &blueprintModuleR{}
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

	if len(blueprintModuleAfterSelectHooks) != 0 {
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
		foreign.R.BlueprintModules = append(foreign.R.BlueprintModules, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.BrandID, foreign.ID) {
				local.R.Brand = foreign
				if foreign.R == nil {
					foreign.R = &brandR{}
				}
				foreign.R.BlueprintModules = append(foreign.R.BlueprintModules, local)
				break
			}
		}
	}

	return nil
}

// SetBrand of the blueprintModule to the related item.
// Sets o.R.Brand to related.
// Adds o to related.R.BlueprintModules.
func (o *BlueprintModule) SetBrand(exec boil.Executor, insert bool, related *Brand) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"blueprint_modules\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"brand_id"}),
		strmangle.WhereClause("\"", "\"", 2, blueprintModulePrimaryKeyColumns),
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
		o.R = &blueprintModuleR{
			Brand: related,
		}
	} else {
		o.R.Brand = related
	}

	if related.R == nil {
		related.R = &brandR{
			BlueprintModules: BlueprintModuleSlice{o},
		}
	} else {
		related.R.BlueprintModules = append(related.R.BlueprintModules, o)
	}

	return nil
}

// RemoveBrand relationship.
// Sets o.R.Brand to nil.
// Removes o from all passed in related items' relationships struct (Optional).
func (o *BlueprintModule) RemoveBrand(exec boil.Executor, related *Brand) error {
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

	for i, ri := range related.R.BlueprintModules {
		if queries.Equal(o.BrandID, ri.BrandID) {
			continue
		}

		ln := len(related.R.BlueprintModules)
		if ln > 1 && i < ln-1 {
			related.R.BlueprintModules[i] = related.R.BlueprintModules[ln-1]
		}
		related.R.BlueprintModules = related.R.BlueprintModules[:ln-1]
		break
	}
	return nil
}

// BlueprintModules retrieves all the records using an executor.
func BlueprintModules(mods ...qm.QueryMod) blueprintModuleQuery {
	mods = append(mods, qm.From("\"blueprint_modules\""), qmhelper.WhereIsNull("\"blueprint_modules\".\"deleted_at\""))
	return blueprintModuleQuery{NewQuery(mods...)}
}

// FindBlueprintModule retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindBlueprintModule(exec boil.Executor, iD string, selectCols ...string) (*BlueprintModule, error) {
	blueprintModuleObj := &BlueprintModule{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"blueprint_modules\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, blueprintModuleObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from blueprint_modules")
	}

	if err = blueprintModuleObj.doAfterSelectHooks(exec); err != nil {
		return blueprintModuleObj, err
	}

	return blueprintModuleObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *BlueprintModule) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no blueprint_modules provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(blueprintModuleColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	blueprintModuleInsertCacheMut.RLock()
	cache, cached := blueprintModuleInsertCache[key]
	blueprintModuleInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			blueprintModuleAllColumns,
			blueprintModuleColumnsWithDefault,
			blueprintModuleColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(blueprintModuleType, blueprintModuleMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(blueprintModuleType, blueprintModuleMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"blueprint_modules\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"blueprint_modules\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into blueprint_modules")
	}

	if !cached {
		blueprintModuleInsertCacheMut.Lock()
		blueprintModuleInsertCache[key] = cache
		blueprintModuleInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the BlueprintModule.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *BlueprintModule) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	blueprintModuleUpdateCacheMut.RLock()
	cache, cached := blueprintModuleUpdateCache[key]
	blueprintModuleUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			blueprintModuleAllColumns,
			blueprintModulePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update blueprint_modules, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"blueprint_modules\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, blueprintModulePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(blueprintModuleType, blueprintModuleMapping, append(wl, blueprintModulePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update blueprint_modules row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for blueprint_modules")
	}

	if !cached {
		blueprintModuleUpdateCacheMut.Lock()
		blueprintModuleUpdateCache[key] = cache
		blueprintModuleUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q blueprintModuleQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for blueprint_modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for blueprint_modules")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o BlueprintModuleSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), blueprintModulePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"blueprint_modules\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, blueprintModulePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in blueprintModule slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all blueprintModule")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *BlueprintModule) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no blueprint_modules provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(blueprintModuleColumnsWithDefault, o)

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

	blueprintModuleUpsertCacheMut.RLock()
	cache, cached := blueprintModuleUpsertCache[key]
	blueprintModuleUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			blueprintModuleAllColumns,
			blueprintModuleColumnsWithDefault,
			blueprintModuleColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			blueprintModuleAllColumns,
			blueprintModulePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert blueprint_modules, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(blueprintModulePrimaryKeyColumns))
			copy(conflict, blueprintModulePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"blueprint_modules\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(blueprintModuleType, blueprintModuleMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(blueprintModuleType, blueprintModuleMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert blueprint_modules")
	}

	if !cached {
		blueprintModuleUpsertCacheMut.Lock()
		blueprintModuleUpsertCache[key] = cache
		blueprintModuleUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single BlueprintModule record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *BlueprintModule) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no BlueprintModule provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), blueprintModulePrimaryKeyMapping)
		sql = "DELETE FROM \"blueprint_modules\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"blueprint_modules\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(blueprintModuleType, blueprintModuleMapping, append(wl, blueprintModulePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from blueprint_modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for blueprint_modules")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q blueprintModuleQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no blueprintModuleQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from blueprint_modules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for blueprint_modules")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o BlueprintModuleSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(blueprintModuleBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), blueprintModulePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"blueprint_modules\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, blueprintModulePrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), blueprintModulePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"blueprint_modules\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, blueprintModulePrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from blueprintModule slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for blueprint_modules")
	}

	if len(blueprintModuleAfterDeleteHooks) != 0 {
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
func (o *BlueprintModule) Reload(exec boil.Executor) error {
	ret, err := FindBlueprintModule(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *BlueprintModuleSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := BlueprintModuleSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), blueprintModulePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"blueprint_modules\".* FROM \"blueprint_modules\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, blueprintModulePrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in BlueprintModuleSlice")
	}

	*o = slice

	return nil
}

// BlueprintModuleExists checks if the BlueprintModule row exists.
func BlueprintModuleExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"blueprint_modules\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if blueprint_modules exists")
	}

	return exists, nil
}
