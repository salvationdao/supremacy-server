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

// UtilityAntiMissile is an object representing the database table.
type UtilityAntiMissile struct {
	UtilityID      string `boiler:"utility_id" boil:"utility_id" json:"utility_id" toml:"utility_id" yaml:"utility_id"`
	RateOfFire     int    `boiler:"rate_of_fire" boil:"rate_of_fire" json:"rate_of_fire" toml:"rate_of_fire" yaml:"rate_of_fire"`
	FireEnergyCost int    `boiler:"fire_energy_cost" boil:"fire_energy_cost" json:"fire_energy_cost" toml:"fire_energy_cost" yaml:"fire_energy_cost"`

	R *utilityAntiMissileR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L utilityAntiMissileL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var UtilityAntiMissileColumns = struct {
	UtilityID      string
	RateOfFire     string
	FireEnergyCost string
}{
	UtilityID:      "utility_id",
	RateOfFire:     "rate_of_fire",
	FireEnergyCost: "fire_energy_cost",
}

var UtilityAntiMissileTableColumns = struct {
	UtilityID      string
	RateOfFire     string
	FireEnergyCost string
}{
	UtilityID:      "utility_anti_missile.utility_id",
	RateOfFire:     "utility_anti_missile.rate_of_fire",
	FireEnergyCost: "utility_anti_missile.fire_energy_cost",
}

// Generated where

var UtilityAntiMissileWhere = struct {
	UtilityID      whereHelperstring
	RateOfFire     whereHelperint
	FireEnergyCost whereHelperint
}{
	UtilityID:      whereHelperstring{field: "\"utility_anti_missile\".\"utility_id\""},
	RateOfFire:     whereHelperint{field: "\"utility_anti_missile\".\"rate_of_fire\""},
	FireEnergyCost: whereHelperint{field: "\"utility_anti_missile\".\"fire_energy_cost\""},
}

// UtilityAntiMissileRels is where relationship names are stored.
var UtilityAntiMissileRels = struct {
	Utility string
}{
	Utility: "Utility",
}

// utilityAntiMissileR is where relationships are stored.
type utilityAntiMissileR struct {
	Utility *Utility `boiler:"Utility" boil:"Utility" json:"Utility" toml:"Utility" yaml:"Utility"`
}

// NewStruct creates a new relationship struct
func (*utilityAntiMissileR) NewStruct() *utilityAntiMissileR {
	return &utilityAntiMissileR{}
}

// utilityAntiMissileL is where Load methods for each relationship are stored.
type utilityAntiMissileL struct{}

var (
	utilityAntiMissileAllColumns            = []string{"utility_id", "rate_of_fire", "fire_energy_cost"}
	utilityAntiMissileColumnsWithoutDefault = []string{"utility_id", "rate_of_fire", "fire_energy_cost"}
	utilityAntiMissileColumnsWithDefault    = []string{}
	utilityAntiMissilePrimaryKeyColumns     = []string{"utility_id"}
	utilityAntiMissileGeneratedColumns      = []string{}
)

type (
	// UtilityAntiMissileSlice is an alias for a slice of pointers to UtilityAntiMissile.
	// This should almost always be used instead of []UtilityAntiMissile.
	UtilityAntiMissileSlice []*UtilityAntiMissile
	// UtilityAntiMissileHook is the signature for custom UtilityAntiMissile hook methods
	UtilityAntiMissileHook func(boil.Executor, *UtilityAntiMissile) error

	utilityAntiMissileQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	utilityAntiMissileType                 = reflect.TypeOf(&UtilityAntiMissile{})
	utilityAntiMissileMapping              = queries.MakeStructMapping(utilityAntiMissileType)
	utilityAntiMissilePrimaryKeyMapping, _ = queries.BindMapping(utilityAntiMissileType, utilityAntiMissileMapping, utilityAntiMissilePrimaryKeyColumns)
	utilityAntiMissileInsertCacheMut       sync.RWMutex
	utilityAntiMissileInsertCache          = make(map[string]insertCache)
	utilityAntiMissileUpdateCacheMut       sync.RWMutex
	utilityAntiMissileUpdateCache          = make(map[string]updateCache)
	utilityAntiMissileUpsertCacheMut       sync.RWMutex
	utilityAntiMissileUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var utilityAntiMissileAfterSelectHooks []UtilityAntiMissileHook

var utilityAntiMissileBeforeInsertHooks []UtilityAntiMissileHook
var utilityAntiMissileAfterInsertHooks []UtilityAntiMissileHook

var utilityAntiMissileBeforeUpdateHooks []UtilityAntiMissileHook
var utilityAntiMissileAfterUpdateHooks []UtilityAntiMissileHook

var utilityAntiMissileBeforeDeleteHooks []UtilityAntiMissileHook
var utilityAntiMissileAfterDeleteHooks []UtilityAntiMissileHook

var utilityAntiMissileBeforeUpsertHooks []UtilityAntiMissileHook
var utilityAntiMissileAfterUpsertHooks []UtilityAntiMissileHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *UtilityAntiMissile) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAntiMissileAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *UtilityAntiMissile) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAntiMissileBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *UtilityAntiMissile) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAntiMissileAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *UtilityAntiMissile) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAntiMissileBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *UtilityAntiMissile) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAntiMissileAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *UtilityAntiMissile) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAntiMissileBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *UtilityAntiMissile) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAntiMissileAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *UtilityAntiMissile) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAntiMissileBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *UtilityAntiMissile) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAntiMissileAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddUtilityAntiMissileHook registers your hook function for all future operations.
func AddUtilityAntiMissileHook(hookPoint boil.HookPoint, utilityAntiMissileHook UtilityAntiMissileHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		utilityAntiMissileAfterSelectHooks = append(utilityAntiMissileAfterSelectHooks, utilityAntiMissileHook)
	case boil.BeforeInsertHook:
		utilityAntiMissileBeforeInsertHooks = append(utilityAntiMissileBeforeInsertHooks, utilityAntiMissileHook)
	case boil.AfterInsertHook:
		utilityAntiMissileAfterInsertHooks = append(utilityAntiMissileAfterInsertHooks, utilityAntiMissileHook)
	case boil.BeforeUpdateHook:
		utilityAntiMissileBeforeUpdateHooks = append(utilityAntiMissileBeforeUpdateHooks, utilityAntiMissileHook)
	case boil.AfterUpdateHook:
		utilityAntiMissileAfterUpdateHooks = append(utilityAntiMissileAfterUpdateHooks, utilityAntiMissileHook)
	case boil.BeforeDeleteHook:
		utilityAntiMissileBeforeDeleteHooks = append(utilityAntiMissileBeforeDeleteHooks, utilityAntiMissileHook)
	case boil.AfterDeleteHook:
		utilityAntiMissileAfterDeleteHooks = append(utilityAntiMissileAfterDeleteHooks, utilityAntiMissileHook)
	case boil.BeforeUpsertHook:
		utilityAntiMissileBeforeUpsertHooks = append(utilityAntiMissileBeforeUpsertHooks, utilityAntiMissileHook)
	case boil.AfterUpsertHook:
		utilityAntiMissileAfterUpsertHooks = append(utilityAntiMissileAfterUpsertHooks, utilityAntiMissileHook)
	}
}

// One returns a single utilityAntiMissile record from the query.
func (q utilityAntiMissileQuery) One(exec boil.Executor) (*UtilityAntiMissile, error) {
	o := &UtilityAntiMissile{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for utility_anti_missile")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all UtilityAntiMissile records from the query.
func (q utilityAntiMissileQuery) All(exec boil.Executor) (UtilityAntiMissileSlice, error) {
	var o []*UtilityAntiMissile

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to UtilityAntiMissile slice")
	}

	if len(utilityAntiMissileAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all UtilityAntiMissile records in the query.
func (q utilityAntiMissileQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count utility_anti_missile rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q utilityAntiMissileQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if utility_anti_missile exists")
	}

	return count > 0, nil
}

// Utility pointed to by the foreign key.
func (o *UtilityAntiMissile) Utility(mods ...qm.QueryMod) utilityQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.UtilityID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Utilities(queryMods...)
	queries.SetFrom(query.Query, "\"utility\"")

	return query
}

// LoadUtility allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (utilityAntiMissileL) LoadUtility(e boil.Executor, singular bool, maybeUtilityAntiMissile interface{}, mods queries.Applicator) error {
	var slice []*UtilityAntiMissile
	var object *UtilityAntiMissile

	if singular {
		object = maybeUtilityAntiMissile.(*UtilityAntiMissile)
	} else {
		slice = *maybeUtilityAntiMissile.(*[]*UtilityAntiMissile)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &utilityAntiMissileR{}
		}
		args = append(args, object.UtilityID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &utilityAntiMissileR{}
			}

			for _, a := range args {
				if a == obj.UtilityID {
					continue Outer
				}
			}

			args = append(args, obj.UtilityID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`utility`),
		qm.WhereIn(`utility.id in ?`, args...),
		qmhelper.WhereIsNull(`utility.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Utility")
	}

	var resultSlice []*Utility
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Utility")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for utility")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for utility")
	}

	if len(utilityAntiMissileAfterSelectHooks) != 0 {
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
		object.R.Utility = foreign
		if foreign.R == nil {
			foreign.R = &utilityR{}
		}
		foreign.R.UtilityAntiMissile = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UtilityID == foreign.ID {
				local.R.Utility = foreign
				if foreign.R == nil {
					foreign.R = &utilityR{}
				}
				foreign.R.UtilityAntiMissile = local
				break
			}
		}
	}

	return nil
}

// SetUtility of the utilityAntiMissile to the related item.
// Sets o.R.Utility to related.
// Adds o to related.R.UtilityAntiMissile.
func (o *UtilityAntiMissile) SetUtility(exec boil.Executor, insert bool, related *Utility) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"utility_anti_missile\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"utility_id"}),
		strmangle.WhereClause("\"", "\"", 2, utilityAntiMissilePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.UtilityID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.UtilityID = related.ID
	if o.R == nil {
		o.R = &utilityAntiMissileR{
			Utility: related,
		}
	} else {
		o.R.Utility = related
	}

	if related.R == nil {
		related.R = &utilityR{
			UtilityAntiMissile: o,
		}
	} else {
		related.R.UtilityAntiMissile = o
	}

	return nil
}

// UtilityAntiMissiles retrieves all the records using an executor.
func UtilityAntiMissiles(mods ...qm.QueryMod) utilityAntiMissileQuery {
	mods = append(mods, qm.From("\"utility_anti_missile\""))
	return utilityAntiMissileQuery{NewQuery(mods...)}
}

// FindUtilityAntiMissile retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindUtilityAntiMissile(exec boil.Executor, utilityID string, selectCols ...string) (*UtilityAntiMissile, error) {
	utilityAntiMissileObj := &UtilityAntiMissile{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"utility_anti_missile\" where \"utility_id\"=$1", sel,
	)

	q := queries.Raw(query, utilityID)

	err := q.Bind(nil, exec, utilityAntiMissileObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from utility_anti_missile")
	}

	if err = utilityAntiMissileObj.doAfterSelectHooks(exec); err != nil {
		return utilityAntiMissileObj, err
	}

	return utilityAntiMissileObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *UtilityAntiMissile) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no utility_anti_missile provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(utilityAntiMissileColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	utilityAntiMissileInsertCacheMut.RLock()
	cache, cached := utilityAntiMissileInsertCache[key]
	utilityAntiMissileInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			utilityAntiMissileAllColumns,
			utilityAntiMissileColumnsWithDefault,
			utilityAntiMissileColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(utilityAntiMissileType, utilityAntiMissileMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(utilityAntiMissileType, utilityAntiMissileMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"utility_anti_missile\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"utility_anti_missile\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into utility_anti_missile")
	}

	if !cached {
		utilityAntiMissileInsertCacheMut.Lock()
		utilityAntiMissileInsertCache[key] = cache
		utilityAntiMissileInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the UtilityAntiMissile.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *UtilityAntiMissile) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	utilityAntiMissileUpdateCacheMut.RLock()
	cache, cached := utilityAntiMissileUpdateCache[key]
	utilityAntiMissileUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			utilityAntiMissileAllColumns,
			utilityAntiMissilePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update utility_anti_missile, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"utility_anti_missile\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, utilityAntiMissilePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(utilityAntiMissileType, utilityAntiMissileMapping, append(wl, utilityAntiMissilePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update utility_anti_missile row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for utility_anti_missile")
	}

	if !cached {
		utilityAntiMissileUpdateCacheMut.Lock()
		utilityAntiMissileUpdateCache[key] = cache
		utilityAntiMissileUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q utilityAntiMissileQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for utility_anti_missile")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for utility_anti_missile")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o UtilityAntiMissileSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityAntiMissilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"utility_anti_missile\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, utilityAntiMissilePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in utilityAntiMissile slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all utilityAntiMissile")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *UtilityAntiMissile) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no utility_anti_missile provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(utilityAntiMissileColumnsWithDefault, o)

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

	utilityAntiMissileUpsertCacheMut.RLock()
	cache, cached := utilityAntiMissileUpsertCache[key]
	utilityAntiMissileUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			utilityAntiMissileAllColumns,
			utilityAntiMissileColumnsWithDefault,
			utilityAntiMissileColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			utilityAntiMissileAllColumns,
			utilityAntiMissilePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert utility_anti_missile, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(utilityAntiMissilePrimaryKeyColumns))
			copy(conflict, utilityAntiMissilePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"utility_anti_missile\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(utilityAntiMissileType, utilityAntiMissileMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(utilityAntiMissileType, utilityAntiMissileMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert utility_anti_missile")
	}

	if !cached {
		utilityAntiMissileUpsertCacheMut.Lock()
		utilityAntiMissileUpsertCache[key] = cache
		utilityAntiMissileUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single UtilityAntiMissile record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *UtilityAntiMissile) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no UtilityAntiMissile provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), utilityAntiMissilePrimaryKeyMapping)
	sql := "DELETE FROM \"utility_anti_missile\" WHERE \"utility_id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from utility_anti_missile")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for utility_anti_missile")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q utilityAntiMissileQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no utilityAntiMissileQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from utility_anti_missile")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for utility_anti_missile")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o UtilityAntiMissileSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(utilityAntiMissileBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityAntiMissilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"utility_anti_missile\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, utilityAntiMissilePrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from utilityAntiMissile slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for utility_anti_missile")
	}

	if len(utilityAntiMissileAfterDeleteHooks) != 0 {
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
func (o *UtilityAntiMissile) Reload(exec boil.Executor) error {
	ret, err := FindUtilityAntiMissile(exec, o.UtilityID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *UtilityAntiMissileSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := UtilityAntiMissileSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityAntiMissilePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"utility_anti_missile\".* FROM \"utility_anti_missile\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, utilityAntiMissilePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in UtilityAntiMissileSlice")
	}

	*o = slice

	return nil
}

// UtilityAntiMissileExists checks if the UtilityAntiMissile row exists.
func UtilityAntiMissileExists(exec boil.Executor, utilityID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"utility_anti_missile\" where \"utility_id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, utilityID)
	}
	row := exec.QueryRow(sql, utilityID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if utility_anti_missile exists")
	}

	return exists, nil
}