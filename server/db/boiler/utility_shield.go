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

// UtilityShield is an object representing the database table.
type UtilityShield struct {
	UtilityID          string `boiler:"utility_id" boil:"utility_id" json:"utility_id" toml:"utility_id" yaml:"utility_id"`
	Hitpoints          int    `boiler:"hitpoints" boil:"hitpoints" json:"hitpoints" toml:"hitpoints" yaml:"hitpoints"`
	RechargeRate       int    `boiler:"recharge_rate" boil:"recharge_rate" json:"recharge_rate" toml:"recharge_rate" yaml:"recharge_rate"`
	RechargeEnergyCost int    `boiler:"recharge_energy_cost" boil:"recharge_energy_cost" json:"recharge_energy_cost" toml:"recharge_energy_cost" yaml:"recharge_energy_cost"`

	R *utilityShieldR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L utilityShieldL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var UtilityShieldColumns = struct {
	UtilityID          string
	Hitpoints          string
	RechargeRate       string
	RechargeEnergyCost string
}{
	UtilityID:          "utility_id",
	Hitpoints:          "hitpoints",
	RechargeRate:       "recharge_rate",
	RechargeEnergyCost: "recharge_energy_cost",
}

var UtilityShieldTableColumns = struct {
	UtilityID          string
	Hitpoints          string
	RechargeRate       string
	RechargeEnergyCost string
}{
	UtilityID:          "utility_shield.utility_id",
	Hitpoints:          "utility_shield.hitpoints",
	RechargeRate:       "utility_shield.recharge_rate",
	RechargeEnergyCost: "utility_shield.recharge_energy_cost",
}

// Generated where

var UtilityShieldWhere = struct {
	UtilityID          whereHelperstring
	Hitpoints          whereHelperint
	RechargeRate       whereHelperint
	RechargeEnergyCost whereHelperint
}{
	UtilityID:          whereHelperstring{field: "\"utility_shield\".\"utility_id\""},
	Hitpoints:          whereHelperint{field: "\"utility_shield\".\"hitpoints\""},
	RechargeRate:       whereHelperint{field: "\"utility_shield\".\"recharge_rate\""},
	RechargeEnergyCost: whereHelperint{field: "\"utility_shield\".\"recharge_energy_cost\""},
}

// UtilityShieldRels is where relationship names are stored.
var UtilityShieldRels = struct {
	Utility string
}{
	Utility: "Utility",
}

// utilityShieldR is where relationships are stored.
type utilityShieldR struct {
	Utility *Utility `boiler:"Utility" boil:"Utility" json:"Utility" toml:"Utility" yaml:"Utility"`
}

// NewStruct creates a new relationship struct
func (*utilityShieldR) NewStruct() *utilityShieldR {
	return &utilityShieldR{}
}

// utilityShieldL is where Load methods for each relationship are stored.
type utilityShieldL struct{}

var (
	utilityShieldAllColumns            = []string{"utility_id", "hitpoints", "recharge_rate", "recharge_energy_cost"}
	utilityShieldColumnsWithoutDefault = []string{"utility_id"}
	utilityShieldColumnsWithDefault    = []string{"hitpoints", "recharge_rate", "recharge_energy_cost"}
	utilityShieldPrimaryKeyColumns     = []string{"utility_id"}
	utilityShieldGeneratedColumns      = []string{}
)

type (
	// UtilityShieldSlice is an alias for a slice of pointers to UtilityShield.
	// This should almost always be used instead of []UtilityShield.
	UtilityShieldSlice []*UtilityShield
	// UtilityShieldHook is the signature for custom UtilityShield hook methods
	UtilityShieldHook func(boil.Executor, *UtilityShield) error

	utilityShieldQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	utilityShieldType                 = reflect.TypeOf(&UtilityShield{})
	utilityShieldMapping              = queries.MakeStructMapping(utilityShieldType)
	utilityShieldPrimaryKeyMapping, _ = queries.BindMapping(utilityShieldType, utilityShieldMapping, utilityShieldPrimaryKeyColumns)
	utilityShieldInsertCacheMut       sync.RWMutex
	utilityShieldInsertCache          = make(map[string]insertCache)
	utilityShieldUpdateCacheMut       sync.RWMutex
	utilityShieldUpdateCache          = make(map[string]updateCache)
	utilityShieldUpsertCacheMut       sync.RWMutex
	utilityShieldUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var utilityShieldAfterSelectHooks []UtilityShieldHook

var utilityShieldBeforeInsertHooks []UtilityShieldHook
var utilityShieldAfterInsertHooks []UtilityShieldHook

var utilityShieldBeforeUpdateHooks []UtilityShieldHook
var utilityShieldAfterUpdateHooks []UtilityShieldHook

var utilityShieldBeforeDeleteHooks []UtilityShieldHook
var utilityShieldAfterDeleteHooks []UtilityShieldHook

var utilityShieldBeforeUpsertHooks []UtilityShieldHook
var utilityShieldAfterUpsertHooks []UtilityShieldHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *UtilityShield) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityShieldAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *UtilityShield) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityShieldBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *UtilityShield) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityShieldAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *UtilityShield) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityShieldBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *UtilityShield) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityShieldAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *UtilityShield) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityShieldBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *UtilityShield) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityShieldAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *UtilityShield) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityShieldBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *UtilityShield) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range utilityShieldAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddUtilityShieldHook registers your hook function for all future operations.
func AddUtilityShieldHook(hookPoint boil.HookPoint, utilityShieldHook UtilityShieldHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		utilityShieldAfterSelectHooks = append(utilityShieldAfterSelectHooks, utilityShieldHook)
	case boil.BeforeInsertHook:
		utilityShieldBeforeInsertHooks = append(utilityShieldBeforeInsertHooks, utilityShieldHook)
	case boil.AfterInsertHook:
		utilityShieldAfterInsertHooks = append(utilityShieldAfterInsertHooks, utilityShieldHook)
	case boil.BeforeUpdateHook:
		utilityShieldBeforeUpdateHooks = append(utilityShieldBeforeUpdateHooks, utilityShieldHook)
	case boil.AfterUpdateHook:
		utilityShieldAfterUpdateHooks = append(utilityShieldAfterUpdateHooks, utilityShieldHook)
	case boil.BeforeDeleteHook:
		utilityShieldBeforeDeleteHooks = append(utilityShieldBeforeDeleteHooks, utilityShieldHook)
	case boil.AfterDeleteHook:
		utilityShieldAfterDeleteHooks = append(utilityShieldAfterDeleteHooks, utilityShieldHook)
	case boil.BeforeUpsertHook:
		utilityShieldBeforeUpsertHooks = append(utilityShieldBeforeUpsertHooks, utilityShieldHook)
	case boil.AfterUpsertHook:
		utilityShieldAfterUpsertHooks = append(utilityShieldAfterUpsertHooks, utilityShieldHook)
	}
}

// One returns a single utilityShield record from the query.
func (q utilityShieldQuery) One(exec boil.Executor) (*UtilityShield, error) {
	o := &UtilityShield{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for utility_shield")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all UtilityShield records from the query.
func (q utilityShieldQuery) All(exec boil.Executor) (UtilityShieldSlice, error) {
	var o []*UtilityShield

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to UtilityShield slice")
	}

	if len(utilityShieldAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all UtilityShield records in the query.
func (q utilityShieldQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count utility_shield rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q utilityShieldQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if utility_shield exists")
	}

	return count > 0, nil
}

// Utility pointed to by the foreign key.
func (o *UtilityShield) Utility(mods ...qm.QueryMod) utilityQuery {
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
func (utilityShieldL) LoadUtility(e boil.Executor, singular bool, maybeUtilityShield interface{}, mods queries.Applicator) error {
	var slice []*UtilityShield
	var object *UtilityShield

	if singular {
		object = maybeUtilityShield.(*UtilityShield)
	} else {
		slice = *maybeUtilityShield.(*[]*UtilityShield)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &utilityShieldR{}
		}
		args = append(args, object.UtilityID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &utilityShieldR{}
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

	if len(utilityShieldAfterSelectHooks) != 0 {
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
		foreign.R.UtilityShield = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UtilityID == foreign.ID {
				local.R.Utility = foreign
				if foreign.R == nil {
					foreign.R = &utilityR{}
				}
				foreign.R.UtilityShield = local
				break
			}
		}
	}

	return nil
}

// SetUtility of the utilityShield to the related item.
// Sets o.R.Utility to related.
// Adds o to related.R.UtilityShield.
func (o *UtilityShield) SetUtility(exec boil.Executor, insert bool, related *Utility) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"utility_shield\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"utility_id"}),
		strmangle.WhereClause("\"", "\"", 2, utilityShieldPrimaryKeyColumns),
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
		o.R = &utilityShieldR{
			Utility: related,
		}
	} else {
		o.R.Utility = related
	}

	if related.R == nil {
		related.R = &utilityR{
			UtilityShield: o,
		}
	} else {
		related.R.UtilityShield = o
	}

	return nil
}

// UtilityShields retrieves all the records using an executor.
func UtilityShields(mods ...qm.QueryMod) utilityShieldQuery {
	mods = append(mods, qm.From("\"utility_shield\""))
	return utilityShieldQuery{NewQuery(mods...)}
}

// FindUtilityShield retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindUtilityShield(exec boil.Executor, utilityID string, selectCols ...string) (*UtilityShield, error) {
	utilityShieldObj := &UtilityShield{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"utility_shield\" where \"utility_id\"=$1", sel,
	)

	q := queries.Raw(query, utilityID)

	err := q.Bind(nil, exec, utilityShieldObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from utility_shield")
	}

	if err = utilityShieldObj.doAfterSelectHooks(exec); err != nil {
		return utilityShieldObj, err
	}

	return utilityShieldObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *UtilityShield) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no utility_shield provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(utilityShieldColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	utilityShieldInsertCacheMut.RLock()
	cache, cached := utilityShieldInsertCache[key]
	utilityShieldInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			utilityShieldAllColumns,
			utilityShieldColumnsWithDefault,
			utilityShieldColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(utilityShieldType, utilityShieldMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(utilityShieldType, utilityShieldMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"utility_shield\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"utility_shield\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into utility_shield")
	}

	if !cached {
		utilityShieldInsertCacheMut.Lock()
		utilityShieldInsertCache[key] = cache
		utilityShieldInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the UtilityShield.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *UtilityShield) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	utilityShieldUpdateCacheMut.RLock()
	cache, cached := utilityShieldUpdateCache[key]
	utilityShieldUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			utilityShieldAllColumns,
			utilityShieldPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update utility_shield, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"utility_shield\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, utilityShieldPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(utilityShieldType, utilityShieldMapping, append(wl, utilityShieldPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update utility_shield row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for utility_shield")
	}

	if !cached {
		utilityShieldUpdateCacheMut.Lock()
		utilityShieldUpdateCache[key] = cache
		utilityShieldUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q utilityShieldQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for utility_shield")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for utility_shield")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o UtilityShieldSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityShieldPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"utility_shield\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, utilityShieldPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in utilityShield slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all utilityShield")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *UtilityShield) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no utility_shield provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(utilityShieldColumnsWithDefault, o)

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

	utilityShieldUpsertCacheMut.RLock()
	cache, cached := utilityShieldUpsertCache[key]
	utilityShieldUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			utilityShieldAllColumns,
			utilityShieldColumnsWithDefault,
			utilityShieldColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			utilityShieldAllColumns,
			utilityShieldPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert utility_shield, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(utilityShieldPrimaryKeyColumns))
			copy(conflict, utilityShieldPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"utility_shield\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(utilityShieldType, utilityShieldMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(utilityShieldType, utilityShieldMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert utility_shield")
	}

	if !cached {
		utilityShieldUpsertCacheMut.Lock()
		utilityShieldUpsertCache[key] = cache
		utilityShieldUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single UtilityShield record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *UtilityShield) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no UtilityShield provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), utilityShieldPrimaryKeyMapping)
	sql := "DELETE FROM \"utility_shield\" WHERE \"utility_id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from utility_shield")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for utility_shield")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q utilityShieldQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no utilityShieldQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from utility_shield")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for utility_shield")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o UtilityShieldSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(utilityShieldBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityShieldPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"utility_shield\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, utilityShieldPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from utilityShield slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for utility_shield")
	}

	if len(utilityShieldAfterDeleteHooks) != 0 {
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
func (o *UtilityShield) Reload(exec boil.Executor) error {
	ret, err := FindUtilityShield(exec, o.UtilityID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *UtilityShieldSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := UtilityShieldSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), utilityShieldPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"utility_shield\".* FROM \"utility_shield\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, utilityShieldPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in UtilityShieldSlice")
	}

	*o = slice

	return nil
}

// UtilityShieldExists checks if the UtilityShield row exists.
func UtilityShieldExists(exec boil.Executor, utilityID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"utility_shield\" where \"utility_id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, utilityID)
	}
	row := exec.QueryRow(sql, utilityID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if utility_shield exists")
	}

	return exists, nil
}