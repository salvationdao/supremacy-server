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

// UtilityAttackDrone is an object representing the database table.
type UtilityAttackDrone struct {
	UtilityID        string `boiler:"utility_id" boil:"utility_id" json:"utility_id" toml:"utility_id" yaml:"utility_id"`
	Damage           int    `boiler:"damage" boil:"damage" json:"damage" toml:"damage" yaml:"damage"`
	RateOfFire       int    `boiler:"rate_of_fire" boil:"rate_of_fire" json:"rate_of_fire" toml:"rate_of_fire" yaml:"rate_of_fire"`
	Hitpoints        int    `boiler:"hitpoints" boil:"hitpoints" json:"hitpoints" toml:"hitpoints" yaml:"hitpoints"`
	LifespanSeconds  int    `boiler:"lifespan_seconds" boil:"lifespan_seconds" json:"lifespan_seconds" toml:"lifespan_seconds" yaml:"lifespan_seconds"`
	DeployEnergyCost int    `boiler:"deploy_energy_cost" boil:"deploy_energy_cost" json:"deploy_energy_cost" toml:"deploy_energy_cost" yaml:"deploy_energy_cost"`

	R *utilityAttackDroneR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L utilityAttackDroneL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var UtilityAttackDroneColumns = struct {
	UtilityID        string
	Damage           string
	RateOfFire       string
	Hitpoints        string
	LifespanSeconds  string
	DeployEnergyCost string
}{
	UtilityID:        "utility_id",
	Damage:           "damage",
	RateOfFire:       "rate_of_fire",
	Hitpoints:        "hitpoints",
	LifespanSeconds:  "lifespan_seconds",
	DeployEnergyCost: "deploy_energy_cost",
}

var UtilityAttackDroneTableColumns = struct {
	UtilityID        string
	Damage           string
	RateOfFire       string
	Hitpoints        string
	LifespanSeconds  string
	DeployEnergyCost string
}{
	UtilityID:        "utility_attack_drone.utility_id",
	Damage:           "utility_attack_drone.damage",
	RateOfFire:       "utility_attack_drone.rate_of_fire",
	Hitpoints:        "utility_attack_drone.hitpoints",
	LifespanSeconds:  "utility_attack_drone.lifespan_seconds",
	DeployEnergyCost: "utility_attack_drone.deploy_energy_cost",
}

// Generated where

var UtilityAttackDroneWhere = struct {
	UtilityID        whereHelperstring
	Damage           whereHelperint
	RateOfFire       whereHelperint
	Hitpoints        whereHelperint
	LifespanSeconds  whereHelperint
	DeployEnergyCost whereHelperint
}{
	UtilityID:        whereHelperstring{field: "\"utility_attack_drone\".\"utility_id\""},
	Damage:           whereHelperint{field: "\"utility_attack_drone\".\"damage\""},
	RateOfFire:       whereHelperint{field: "\"utility_attack_drone\".\"rate_of_fire\""},
	Hitpoints:        whereHelperint{field: "\"utility_attack_drone\".\"hitpoints\""},
	LifespanSeconds:  whereHelperint{field: "\"utility_attack_drone\".\"lifespan_seconds\""},
	DeployEnergyCost: whereHelperint{field: "\"utility_attack_drone\".\"deploy_energy_cost\""},
}

// UtilityAttackDroneRels is where relationship names are stored.
var UtilityAttackDroneRels = struct {
	Utility string
}{
	Utility: "Utility",
}

// utilityAttackDroneR is where relationships are stored.
type utilityAttackDroneR struct {
	Utility *Utility `boiler:"Utility" boil:"Utility" json:"Utility" toml:"Utility" yaml:"Utility"`
}

// NewStruct creates a new relationship struct
func (*utilityAttackDroneR) NewStruct() *utilityAttackDroneR {
	return &utilityAttackDroneR{}
}

// utilityAttackDroneL is where Load methods for each relationship are stored.
type utilityAttackDroneL struct{}

var (
	utilityAttackDroneAllColumns            = []string{"utility_id", "damage", "rate_of_fire", "hitpoints", "lifespan_seconds", "deploy_energy_cost"}
	utilityAttackDroneColumnsWithoutDefault = []string{"utility_id", "damage", "rate_of_fire", "hitpoints", "lifespan_seconds", "deploy_energy_cost"}
	utilityAttackDroneColumnsWithDefault    = []string{}
	utilityAttackDronePrimaryKeyColumns     = []string{"utility_id"}
	utilityAttackDroneGeneratedColumns      = []string{}
)

type (
	// UtilityAttackDroneSlice is an alias for a slice of pointers to UtilityAttackDrone.
	// This should almost always be used instead of []UtilityAttackDrone.
	UtilityAttackDroneSlice []*UtilityAttackDrone
	// UtilityAttackDroneHook is the signature for custom UtilityAttackDrone hook methods
	UtilityAttackDroneHook func(boil.Executor, *UtilityAttackDrone) error

	utilityAttackDroneQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	utilityAttackDroneType                 = reflect.TypeOf(&UtilityAttackDrone{})
	utilityAttackDroneMapping              = queries.MakeStructMapping(utilityAttackDroneType)
	utilityAttackDronePrimaryKeyMapping, _ = queries.BindMapping(utilityAttackDroneType, utilityAttackDroneMapping, utilityAttackDronePrimaryKeyColumns)
	utilityAttackDroneInsertCacheMut       sync.RWMutex
	utilityAttackDroneInsertCache          = make(map[string]insertCache)
	utilityAttackDroneUpdateCacheMut       sync.RWMutex
	utilityAttackDroneUpdateCache          = make(map[string]updateCache)
	utilityAttackDroneUpsertCacheMut       sync.RWMutex
	utilityAttackDroneUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var utilityAttackDroneAfterSelectHooks []UtilityAttackDroneHook

var utilityAttackDroneBeforeInsertHooks []UtilityAttackDroneHook
var utilityAttackDroneAfterInsertHooks []UtilityAttackDroneHook

var utilityAttackDroneBeforeUpdateHooks []UtilityAttackDroneHook
var utilityAttackDroneAfterUpdateHooks []UtilityAttackDroneHook

var utilityAttackDroneBeforeDeleteHooks []UtilityAttackDroneHook
var utilityAttackDroneAfterDeleteHooks []UtilityAttackDroneHook

var utilityAttackDroneBeforeUpsertHooks []UtilityAttackDroneHook
var utilityAttackDroneAfterUpsertHooks []UtilityAttackDroneHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *UtilityAttackDrone) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAttackDroneAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *UtilityAttackDrone) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAttackDroneBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *UtilityAttackDrone) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAttackDroneAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *UtilityAttackDrone) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAttackDroneBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *UtilityAttackDrone) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAttackDroneAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *UtilityAttackDrone) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAttackDroneBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *UtilityAttackDrone) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAttackDroneAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *UtilityAttackDrone) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAttackDroneBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *UtilityAttackDrone) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityAttackDroneAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddUtilityAttackDroneHook registers your hook function for all future operations.
func AddUtilityAttackDroneHook(hookPoint boil.HookPoint, utilityAttackDroneHook UtilityAttackDroneHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		utilityAttackDroneAfterSelectHooks = append(utilityAttackDroneAfterSelectHooks, utilityAttackDroneHook)
	case boil.BeforeInsertHook:
		utilityAttackDroneBeforeInsertHooks = append(utilityAttackDroneBeforeInsertHooks, utilityAttackDroneHook)
	case boil.AfterInsertHook:
		utilityAttackDroneAfterInsertHooks = append(utilityAttackDroneAfterInsertHooks, utilityAttackDroneHook)
	case boil.BeforeUpdateHook:
		utilityAttackDroneBeforeUpdateHooks = append(utilityAttackDroneBeforeUpdateHooks, utilityAttackDroneHook)
	case boil.AfterUpdateHook:
		utilityAttackDroneAfterUpdateHooks = append(utilityAttackDroneAfterUpdateHooks, utilityAttackDroneHook)
	case boil.BeforeDeleteHook:
		utilityAttackDroneBeforeDeleteHooks = append(utilityAttackDroneBeforeDeleteHooks, utilityAttackDroneHook)
	case boil.AfterDeleteHook:
		utilityAttackDroneAfterDeleteHooks = append(utilityAttackDroneAfterDeleteHooks, utilityAttackDroneHook)
	case boil.BeforeUpsertHook:
		utilityAttackDroneBeforeUpsertHooks = append(utilityAttackDroneBeforeUpsertHooks, utilityAttackDroneHook)
	case boil.AfterUpsertHook:
		utilityAttackDroneAfterUpsertHooks = append(utilityAttackDroneAfterUpsertHooks, utilityAttackDroneHook)
	}
}

// One returns a single utilityAttackDrone record from the query.
func (q utilityAttackDroneQuery) One(exec boil.Executor) (*UtilityAttackDrone, error) {
	o := &UtilityAttackDrone{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for utility_attack_drone")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all UtilityAttackDrone records from the query.
func (q utilityAttackDroneQuery) All(exec boil.Executor) (UtilityAttackDroneSlice, error) {
	var o []*UtilityAttackDrone

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to UtilityAttackDrone slice")
	}

	if len(utilityAttackDroneAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all UtilityAttackDrone records in the query.
func (q utilityAttackDroneQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count utility_attack_drone rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q utilityAttackDroneQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if utility_attack_drone exists")
	}

	return count > 0, nil
}

// Utility pointed to by the foreign key.
func (o *UtilityAttackDrone) Utility(mods ...qm.QueryMod) utilityQuery {
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
func (utilityAttackDroneL) LoadUtility(e boil.Executor, singular bool, maybeUtilityAttackDrone interface{}, mods queries.Applicator) error {
	var slice []*UtilityAttackDrone
	var object *UtilityAttackDrone

	if singular {
		object = maybeUtilityAttackDrone.(*UtilityAttackDrone)
	} else {
		slice = *maybeUtilityAttackDrone.(*[]*UtilityAttackDrone)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &utilityAttackDroneR{}
		}
		args = append(args, object.UtilityID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &utilityAttackDroneR{}
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

	if len(utilityAttackDroneAfterSelectHooks) != 0 {
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
		foreign.R.UtilityAttackDrone = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UtilityID == foreign.ID {
				local.R.Utility = foreign
				if foreign.R == nil {
					foreign.R = &utilityR{}
				}
				foreign.R.UtilityAttackDrone = local
				break
			}
		}
	}

	return nil
}

// SetUtility of the utilityAttackDrone to the related item.
// Sets o.R.Utility to related.
// Adds o to related.R.UtilityAttackDrone.
func (o *UtilityAttackDrone) SetUtility(exec boil.Executor, insert bool, related *Utility) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"utility_attack_drone\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"utility_id"}),
		strmangle.WhereClause("\"", "\"", 2, utilityAttackDronePrimaryKeyColumns),
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
		o.R = &utilityAttackDroneR{
			Utility: related,
		}
	} else {
		o.R.Utility = related
	}

	if related.R == nil {
		related.R = &utilityR{
			UtilityAttackDrone: o,
		}
	} else {
		related.R.UtilityAttackDrone = o
	}

	return nil
}

// UtilityAttackDrones retrieves all the records using an executor.
func UtilityAttackDrones(mods ...qm.QueryMod) utilityAttackDroneQuery {
	mods = append(mods, qm.From("\"utility_attack_drone\""))
	return utilityAttackDroneQuery{NewQuery(mods...)}
}

// FindUtilityAttackDrone retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindUtilityAttackDrone(exec boil.Executor, utilityID string, selectCols ...string) (*UtilityAttackDrone, error) {
	utilityAttackDroneObj := &UtilityAttackDrone{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"utility_attack_drone\" where \"utility_id\"=$1", sel,
	)

	q := queries.Raw(query, utilityID)

	err := q.Bind(nil, exec, utilityAttackDroneObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from utility_attack_drone")
	}

	if err = utilityAttackDroneObj.doAfterSelectHooks(exec); err != nil {
		return utilityAttackDroneObj, err
	}

	return utilityAttackDroneObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *UtilityAttackDrone) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no utility_attack_drone provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(utilityAttackDroneColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	utilityAttackDroneInsertCacheMut.RLock()
	cache, cached := utilityAttackDroneInsertCache[key]
	utilityAttackDroneInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			utilityAttackDroneAllColumns,
			utilityAttackDroneColumnsWithDefault,
			utilityAttackDroneColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(utilityAttackDroneType, utilityAttackDroneMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(utilityAttackDroneType, utilityAttackDroneMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"utility_attack_drone\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"utility_attack_drone\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into utility_attack_drone")
	}

	if !cached {
		utilityAttackDroneInsertCacheMut.Lock()
		utilityAttackDroneInsertCache[key] = cache
		utilityAttackDroneInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the UtilityAttackDrone.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *UtilityAttackDrone) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	utilityAttackDroneUpdateCacheMut.RLock()
	cache, cached := utilityAttackDroneUpdateCache[key]
	utilityAttackDroneUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			utilityAttackDroneAllColumns,
			utilityAttackDronePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update utility_attack_drone, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"utility_attack_drone\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, utilityAttackDronePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(utilityAttackDroneType, utilityAttackDroneMapping, append(wl, utilityAttackDronePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update utility_attack_drone row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for utility_attack_drone")
	}

	if !cached {
		utilityAttackDroneUpdateCacheMut.Lock()
		utilityAttackDroneUpdateCache[key] = cache
		utilityAttackDroneUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q utilityAttackDroneQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for utility_attack_drone")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for utility_attack_drone")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o UtilityAttackDroneSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityAttackDronePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"utility_attack_drone\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, utilityAttackDronePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in utilityAttackDrone slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all utilityAttackDrone")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *UtilityAttackDrone) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no utility_attack_drone provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(utilityAttackDroneColumnsWithDefault, o)

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

	utilityAttackDroneUpsertCacheMut.RLock()
	cache, cached := utilityAttackDroneUpsertCache[key]
	utilityAttackDroneUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			utilityAttackDroneAllColumns,
			utilityAttackDroneColumnsWithDefault,
			utilityAttackDroneColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			utilityAttackDroneAllColumns,
			utilityAttackDronePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert utility_attack_drone, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(utilityAttackDronePrimaryKeyColumns))
			copy(conflict, utilityAttackDronePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"utility_attack_drone\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(utilityAttackDroneType, utilityAttackDroneMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(utilityAttackDroneType, utilityAttackDroneMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert utility_attack_drone")
	}

	if !cached {
		utilityAttackDroneUpsertCacheMut.Lock()
		utilityAttackDroneUpsertCache[key] = cache
		utilityAttackDroneUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single UtilityAttackDrone record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *UtilityAttackDrone) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no UtilityAttackDrone provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), utilityAttackDronePrimaryKeyMapping)
	sql := "DELETE FROM \"utility_attack_drone\" WHERE \"utility_id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from utility_attack_drone")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for utility_attack_drone")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q utilityAttackDroneQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no utilityAttackDroneQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from utility_attack_drone")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for utility_attack_drone")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o UtilityAttackDroneSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(utilityAttackDroneBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityAttackDronePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"utility_attack_drone\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, utilityAttackDronePrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from utilityAttackDrone slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for utility_attack_drone")
	}

	if len(utilityAttackDroneAfterDeleteHooks) != 0 {
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
func (o *UtilityAttackDrone) Reload(exec boil.Executor) error {
	ret, err := FindUtilityAttackDrone(exec, o.UtilityID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *UtilityAttackDroneSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := UtilityAttackDroneSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityAttackDronePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"utility_attack_drone\".* FROM \"utility_attack_drone\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, utilityAttackDronePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in UtilityAttackDroneSlice")
	}

	*o = slice

	return nil
}

// UtilityAttackDroneExists checks if the UtilityAttackDrone row exists.
func UtilityAttackDroneExists(exec boil.Executor, utilityID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"utility_attack_drone\" where \"utility_id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, utilityID)
	}
	row := exec.QueryRow(sql, utilityID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if utility_attack_drone exists")
	}

	return exists, nil
}
