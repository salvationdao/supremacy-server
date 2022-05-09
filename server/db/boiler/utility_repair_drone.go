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

// UtilityRepairDrone is an object representing the database table.
type UtilityRepairDrone struct {
	UtilityID        string      `boiler:"utility_id" boil:"utility_id" json:"utility_id" toml:"utility_id" yaml:"utility_id"`
	RepairType       null.String `boiler:"repair_type" boil:"repair_type" json:"repair_type,omitempty" toml:"repair_type" yaml:"repair_type,omitempty"`
	RepairAmount     int         `boiler:"repair_amount" boil:"repair_amount" json:"repair_amount" toml:"repair_amount" yaml:"repair_amount"`
	DeployEnergyCost int         `boiler:"deploy_energy_cost" boil:"deploy_energy_cost" json:"deploy_energy_cost" toml:"deploy_energy_cost" yaml:"deploy_energy_cost"`
	LifespanSeconds  int         `boiler:"lifespan_seconds" boil:"lifespan_seconds" json:"lifespan_seconds" toml:"lifespan_seconds" yaml:"lifespan_seconds"`

	R *utilityRepairDroneR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L utilityRepairDroneL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var UtilityRepairDroneColumns = struct {
	UtilityID        string
	RepairType       string
	RepairAmount     string
	DeployEnergyCost string
	LifespanSeconds  string
}{
	UtilityID:        "utility_id",
	RepairType:       "repair_type",
	RepairAmount:     "repair_amount",
	DeployEnergyCost: "deploy_energy_cost",
	LifespanSeconds:  "lifespan_seconds",
}

var UtilityRepairDroneTableColumns = struct {
	UtilityID        string
	RepairType       string
	RepairAmount     string
	DeployEnergyCost string
	LifespanSeconds  string
}{
	UtilityID:        "utility_repair_drone.utility_id",
	RepairType:       "utility_repair_drone.repair_type",
	RepairAmount:     "utility_repair_drone.repair_amount",
	DeployEnergyCost: "utility_repair_drone.deploy_energy_cost",
	LifespanSeconds:  "utility_repair_drone.lifespan_seconds",
}

// Generated where

var UtilityRepairDroneWhere = struct {
	UtilityID        whereHelperstring
	RepairType       whereHelpernull_String
	RepairAmount     whereHelperint
	DeployEnergyCost whereHelperint
	LifespanSeconds  whereHelperint
}{
	UtilityID:        whereHelperstring{field: "\"utility_repair_drone\".\"utility_id\""},
	RepairType:       whereHelpernull_String{field: "\"utility_repair_drone\".\"repair_type\""},
	RepairAmount:     whereHelperint{field: "\"utility_repair_drone\".\"repair_amount\""},
	DeployEnergyCost: whereHelperint{field: "\"utility_repair_drone\".\"deploy_energy_cost\""},
	LifespanSeconds:  whereHelperint{field: "\"utility_repair_drone\".\"lifespan_seconds\""},
}

// UtilityRepairDroneRels is where relationship names are stored.
var UtilityRepairDroneRels = struct {
	Utility string
}{
	Utility: "Utility",
}

// utilityRepairDroneR is where relationships are stored.
type utilityRepairDroneR struct {
	Utility *Utility `boiler:"Utility" boil:"Utility" json:"Utility" toml:"Utility" yaml:"Utility"`
}

// NewStruct creates a new relationship struct
func (*utilityRepairDroneR) NewStruct() *utilityRepairDroneR {
	return &utilityRepairDroneR{}
}

// utilityRepairDroneL is where Load methods for each relationship are stored.
type utilityRepairDroneL struct{}

var (
	utilityRepairDroneAllColumns            = []string{"utility_id", "repair_type", "repair_amount", "deploy_energy_cost", "lifespan_seconds"}
	utilityRepairDroneColumnsWithoutDefault = []string{"utility_id", "repair_amount", "deploy_energy_cost", "lifespan_seconds"}
	utilityRepairDroneColumnsWithDefault    = []string{"repair_type"}
	utilityRepairDronePrimaryKeyColumns     = []string{"utility_id"}
	utilityRepairDroneGeneratedColumns      = []string{}
)

type (
	// UtilityRepairDroneSlice is an alias for a slice of pointers to UtilityRepairDrone.
	// This should almost always be used instead of []UtilityRepairDrone.
	UtilityRepairDroneSlice []*UtilityRepairDrone
	// UtilityRepairDroneHook is the signature for custom UtilityRepairDrone hook methods
	UtilityRepairDroneHook func(boil.Executor, *UtilityRepairDrone) error

	utilityRepairDroneQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	utilityRepairDroneType                 = reflect.TypeOf(&UtilityRepairDrone{})
	utilityRepairDroneMapping              = queries.MakeStructMapping(utilityRepairDroneType)
	utilityRepairDronePrimaryKeyMapping, _ = queries.BindMapping(utilityRepairDroneType, utilityRepairDroneMapping, utilityRepairDronePrimaryKeyColumns)
	utilityRepairDroneInsertCacheMut       sync.RWMutex
	utilityRepairDroneInsertCache          = make(map[string]insertCache)
	utilityRepairDroneUpdateCacheMut       sync.RWMutex
	utilityRepairDroneUpdateCache          = make(map[string]updateCache)
	utilityRepairDroneUpsertCacheMut       sync.RWMutex
	utilityRepairDroneUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var utilityRepairDroneAfterSelectHooks []UtilityRepairDroneHook

var utilityRepairDroneBeforeInsertHooks []UtilityRepairDroneHook
var utilityRepairDroneAfterInsertHooks []UtilityRepairDroneHook

var utilityRepairDroneBeforeUpdateHooks []UtilityRepairDroneHook
var utilityRepairDroneAfterUpdateHooks []UtilityRepairDroneHook

var utilityRepairDroneBeforeDeleteHooks []UtilityRepairDroneHook
var utilityRepairDroneAfterDeleteHooks []UtilityRepairDroneHook

var utilityRepairDroneBeforeUpsertHooks []UtilityRepairDroneHook
var utilityRepairDroneAfterUpsertHooks []UtilityRepairDroneHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *UtilityRepairDrone) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityRepairDroneAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *UtilityRepairDrone) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityRepairDroneBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *UtilityRepairDrone) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityRepairDroneAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *UtilityRepairDrone) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityRepairDroneBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *UtilityRepairDrone) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityRepairDroneAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *UtilityRepairDrone) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityRepairDroneBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *UtilityRepairDrone) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityRepairDroneAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *UtilityRepairDrone) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityRepairDroneBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *UtilityRepairDrone) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityRepairDroneAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddUtilityRepairDroneHook registers your hook function for all future operations.
func AddUtilityRepairDroneHook(hookPoint boil.HookPoint, utilityRepairDroneHook UtilityRepairDroneHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		utilityRepairDroneAfterSelectHooks = append(utilityRepairDroneAfterSelectHooks, utilityRepairDroneHook)
	case boil.BeforeInsertHook:
		utilityRepairDroneBeforeInsertHooks = append(utilityRepairDroneBeforeInsertHooks, utilityRepairDroneHook)
	case boil.AfterInsertHook:
		utilityRepairDroneAfterInsertHooks = append(utilityRepairDroneAfterInsertHooks, utilityRepairDroneHook)
	case boil.BeforeUpdateHook:
		utilityRepairDroneBeforeUpdateHooks = append(utilityRepairDroneBeforeUpdateHooks, utilityRepairDroneHook)
	case boil.AfterUpdateHook:
		utilityRepairDroneAfterUpdateHooks = append(utilityRepairDroneAfterUpdateHooks, utilityRepairDroneHook)
	case boil.BeforeDeleteHook:
		utilityRepairDroneBeforeDeleteHooks = append(utilityRepairDroneBeforeDeleteHooks, utilityRepairDroneHook)
	case boil.AfterDeleteHook:
		utilityRepairDroneAfterDeleteHooks = append(utilityRepairDroneAfterDeleteHooks, utilityRepairDroneHook)
	case boil.BeforeUpsertHook:
		utilityRepairDroneBeforeUpsertHooks = append(utilityRepairDroneBeforeUpsertHooks, utilityRepairDroneHook)
	case boil.AfterUpsertHook:
		utilityRepairDroneAfterUpsertHooks = append(utilityRepairDroneAfterUpsertHooks, utilityRepairDroneHook)
	}
}

// One returns a single utilityRepairDrone record from the query.
func (q utilityRepairDroneQuery) One(exec boil.Executor) (*UtilityRepairDrone, error) {
	o := &UtilityRepairDrone{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for utility_repair_drone")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all UtilityRepairDrone records from the query.
func (q utilityRepairDroneQuery) All(exec boil.Executor) (UtilityRepairDroneSlice, error) {
	var o []*UtilityRepairDrone

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to UtilityRepairDrone slice")
	}

	if len(utilityRepairDroneAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all UtilityRepairDrone records in the query.
func (q utilityRepairDroneQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count utility_repair_drone rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q utilityRepairDroneQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if utility_repair_drone exists")
	}

	return count > 0, nil
}

// Utility pointed to by the foreign key.
func (o *UtilityRepairDrone) Utility(mods ...qm.QueryMod) utilityQuery {
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
func (utilityRepairDroneL) LoadUtility(e boil.Executor, singular bool, maybeUtilityRepairDrone interface{}, mods queries.Applicator) error {
	var slice []*UtilityRepairDrone
	var object *UtilityRepairDrone

	if singular {
		object = maybeUtilityRepairDrone.(*UtilityRepairDrone)
	} else {
		slice = *maybeUtilityRepairDrone.(*[]*UtilityRepairDrone)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &utilityRepairDroneR{}
		}
		args = append(args, object.UtilityID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &utilityRepairDroneR{}
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

	if len(utilityRepairDroneAfterSelectHooks) != 0 {
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
		foreign.R.UtilityRepairDrone = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UtilityID == foreign.ID {
				local.R.Utility = foreign
				if foreign.R == nil {
					foreign.R = &utilityR{}
				}
				foreign.R.UtilityRepairDrone = local
				break
			}
		}
	}

	return nil
}

// SetUtility of the utilityRepairDrone to the related item.
// Sets o.R.Utility to related.
// Adds o to related.R.UtilityRepairDrone.
func (o *UtilityRepairDrone) SetUtility(exec boil.Executor, insert bool, related *Utility) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"utility_repair_drone\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"utility_id"}),
		strmangle.WhereClause("\"", "\"", 2, utilityRepairDronePrimaryKeyColumns),
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
		o.R = &utilityRepairDroneR{
			Utility: related,
		}
	} else {
		o.R.Utility = related
	}

	if related.R == nil {
		related.R = &utilityR{
			UtilityRepairDrone: o,
		}
	} else {
		related.R.UtilityRepairDrone = o
	}

	return nil
}

// UtilityRepairDrones retrieves all the records using an executor.
func UtilityRepairDrones(mods ...qm.QueryMod) utilityRepairDroneQuery {
	mods = append(mods, qm.From("\"utility_repair_drone\""))
	return utilityRepairDroneQuery{NewQuery(mods...)}
}

// FindUtilityRepairDrone retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindUtilityRepairDrone(exec boil.Executor, utilityID string, selectCols ...string) (*UtilityRepairDrone, error) {
	utilityRepairDroneObj := &UtilityRepairDrone{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"utility_repair_drone\" where \"utility_id\"=$1", sel,
	)

	q := queries.Raw(query, utilityID)

	err := q.Bind(nil, exec, utilityRepairDroneObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from utility_repair_drone")
	}

	if err = utilityRepairDroneObj.doAfterSelectHooks(exec); err != nil {
		return utilityRepairDroneObj, err
	}

	return utilityRepairDroneObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *UtilityRepairDrone) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no utility_repair_drone provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(utilityRepairDroneColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	utilityRepairDroneInsertCacheMut.RLock()
	cache, cached := utilityRepairDroneInsertCache[key]
	utilityRepairDroneInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			utilityRepairDroneAllColumns,
			utilityRepairDroneColumnsWithDefault,
			utilityRepairDroneColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(utilityRepairDroneType, utilityRepairDroneMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(utilityRepairDroneType, utilityRepairDroneMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"utility_repair_drone\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"utility_repair_drone\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into utility_repair_drone")
	}

	if !cached {
		utilityRepairDroneInsertCacheMut.Lock()
		utilityRepairDroneInsertCache[key] = cache
		utilityRepairDroneInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the UtilityRepairDrone.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *UtilityRepairDrone) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	utilityRepairDroneUpdateCacheMut.RLock()
	cache, cached := utilityRepairDroneUpdateCache[key]
	utilityRepairDroneUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			utilityRepairDroneAllColumns,
			utilityRepairDronePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update utility_repair_drone, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"utility_repair_drone\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, utilityRepairDronePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(utilityRepairDroneType, utilityRepairDroneMapping, append(wl, utilityRepairDronePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update utility_repair_drone row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for utility_repair_drone")
	}

	if !cached {
		utilityRepairDroneUpdateCacheMut.Lock()
		utilityRepairDroneUpdateCache[key] = cache
		utilityRepairDroneUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q utilityRepairDroneQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for utility_repair_drone")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for utility_repair_drone")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o UtilityRepairDroneSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityRepairDronePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"utility_repair_drone\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, utilityRepairDronePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in utilityRepairDrone slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all utilityRepairDrone")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *UtilityRepairDrone) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no utility_repair_drone provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(utilityRepairDroneColumnsWithDefault, o)

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

	utilityRepairDroneUpsertCacheMut.RLock()
	cache, cached := utilityRepairDroneUpsertCache[key]
	utilityRepairDroneUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			utilityRepairDroneAllColumns,
			utilityRepairDroneColumnsWithDefault,
			utilityRepairDroneColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			utilityRepairDroneAllColumns,
			utilityRepairDronePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert utility_repair_drone, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(utilityRepairDronePrimaryKeyColumns))
			copy(conflict, utilityRepairDronePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"utility_repair_drone\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(utilityRepairDroneType, utilityRepairDroneMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(utilityRepairDroneType, utilityRepairDroneMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert utility_repair_drone")
	}

	if !cached {
		utilityRepairDroneUpsertCacheMut.Lock()
		utilityRepairDroneUpsertCache[key] = cache
		utilityRepairDroneUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single UtilityRepairDrone record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *UtilityRepairDrone) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no UtilityRepairDrone provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), utilityRepairDronePrimaryKeyMapping)
	sql := "DELETE FROM \"utility_repair_drone\" WHERE \"utility_id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from utility_repair_drone")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for utility_repair_drone")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q utilityRepairDroneQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no utilityRepairDroneQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from utility_repair_drone")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for utility_repair_drone")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o UtilityRepairDroneSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(utilityRepairDroneBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityRepairDronePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"utility_repair_drone\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, utilityRepairDronePrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from utilityRepairDrone slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for utility_repair_drone")
	}

	if len(utilityRepairDroneAfterDeleteHooks) != 0 {
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
func (o *UtilityRepairDrone) Reload(exec boil.Executor) error {
	ret, err := FindUtilityRepairDrone(exec, o.UtilityID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *UtilityRepairDroneSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := UtilityRepairDroneSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityRepairDronePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"utility_repair_drone\".* FROM \"utility_repair_drone\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, utilityRepairDronePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in UtilityRepairDroneSlice")
	}

	*o = slice

	return nil
}

// UtilityRepairDroneExists checks if the UtilityRepairDrone row exists.
func UtilityRepairDroneExists(exec boil.Executor, utilityID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"utility_repair_drone\" where \"utility_id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, utilityID)
	}
	row := exec.QueryRow(sql, utilityID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if utility_repair_drone exists")
	}

	return exists, nil
}
