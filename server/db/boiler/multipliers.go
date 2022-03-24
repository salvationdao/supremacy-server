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
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// Multiplier is an object representing the database table.
type Multiplier struct {
	ID               string          `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Description      string          `boiler:"description" boil:"description" json:"description" toml:"description" yaml:"description"`
	Key              string          `boiler:"key" boil:"key" json:"key" toml:"key" yaml:"key"`
	ForGames         int             `boiler:"for_games" boil:"for_games" json:"for_games" toml:"for_games" yaml:"for_games"`
	MultiplierType   string          `boiler:"multiplier_type" boil:"multiplier_type" json:"multiplier_type" toml:"multiplier_type" yaml:"multiplier_type"`
	MustBeOnline     bool            `boiler:"must_be_online" boil:"must_be_online" json:"must_be_online" toml:"must_be_online" yaml:"must_be_online"`
	TestNumber       int             `boiler:"test_number" boil:"test_number" json:"test_number" toml:"test_number" yaml:"test_number"`
	TestString       string          `boiler:"test_string" boil:"test_string" json:"test_string" toml:"test_string" yaml:"test_string"`
	Value            decimal.Decimal `boiler:"value" boil:"value" json:"value" toml:"value" yaml:"value"`
	IsMultiplicative bool            `boiler:"is_multiplicative" boil:"is_multiplicative" json:"is_multiplicative" toml:"is_multiplicative" yaml:"is_multiplicative"`

	R *multiplierR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L multiplierL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var MultiplierColumns = struct {
	ID               string
	Description      string
	Key              string
	ForGames         string
	MultiplierType   string
	MustBeOnline     string
	TestNumber       string
	TestString       string
	Value            string
	IsMultiplicative string
}{
	ID:               "id",
	Description:      "description",
	Key:              "key",
	ForGames:         "for_games",
	MultiplierType:   "multiplier_type",
	MustBeOnline:     "must_be_online",
	TestNumber:       "test_number",
	TestString:       "test_string",
	Value:            "value",
	IsMultiplicative: "is_multiplicative",
}

var MultiplierTableColumns = struct {
	ID               string
	Description      string
	Key              string
	ForGames         string
	MultiplierType   string
	MustBeOnline     string
	TestNumber       string
	TestString       string
	Value            string
	IsMultiplicative string
}{
	ID:               "multipliers.id",
	Description:      "multipliers.description",
	Key:              "multipliers.key",
	ForGames:         "multipliers.for_games",
	MultiplierType:   "multipliers.multiplier_type",
	MustBeOnline:     "multipliers.must_be_online",
	TestNumber:       "multipliers.test_number",
	TestString:       "multipliers.test_string",
	Value:            "multipliers.value",
	IsMultiplicative: "multipliers.is_multiplicative",
}

// Generated where

var MultiplierWhere = struct {
	ID               whereHelperstring
	Description      whereHelperstring
	Key              whereHelperstring
	ForGames         whereHelperint
	MultiplierType   whereHelperstring
	MustBeOnline     whereHelperbool
	TestNumber       whereHelperint
	TestString       whereHelperstring
	Value            whereHelperdecimal_Decimal
	IsMultiplicative whereHelperbool
}{
	ID:               whereHelperstring{field: "\"multipliers\".\"id\""},
	Description:      whereHelperstring{field: "\"multipliers\".\"description\""},
	Key:              whereHelperstring{field: "\"multipliers\".\"key\""},
	ForGames:         whereHelperint{field: "\"multipliers\".\"for_games\""},
	MultiplierType:   whereHelperstring{field: "\"multipliers\".\"multiplier_type\""},
	MustBeOnline:     whereHelperbool{field: "\"multipliers\".\"must_be_online\""},
	TestNumber:       whereHelperint{field: "\"multipliers\".\"test_number\""},
	TestString:       whereHelperstring{field: "\"multipliers\".\"test_string\""},
	Value:            whereHelperdecimal_Decimal{field: "\"multipliers\".\"value\""},
	IsMultiplicative: whereHelperbool{field: "\"multipliers\".\"is_multiplicative\""},
}

// MultiplierRels is where relationship names are stored.
var MultiplierRels = struct {
	UserMultipliers string
}{
	UserMultipliers: "UserMultipliers",
}

// multiplierR is where relationships are stored.
type multiplierR struct {
	UserMultipliers UserMultiplierSlice `boiler:"UserMultipliers" boil:"UserMultipliers" json:"UserMultipliers" toml:"UserMultipliers" yaml:"UserMultipliers"`
}

// NewStruct creates a new relationship struct
func (*multiplierR) NewStruct() *multiplierR {
	return &multiplierR{}
}

// multiplierL is where Load methods for each relationship are stored.
type multiplierL struct{}

var (
	multiplierAllColumns            = []string{"id", "description", "key", "for_games", "multiplier_type", "must_be_online", "test_number", "test_string", "value", "is_multiplicative"}
	multiplierColumnsWithoutDefault = []string{"description", "key", "multiplier_type", "test_number", "test_string"}
	multiplierColumnsWithDefault    = []string{"id", "for_games", "must_be_online", "value", "is_multiplicative"}
	multiplierPrimaryKeyColumns     = []string{"id"}
	multiplierGeneratedColumns      = []string{}
)

type (
	// MultiplierSlice is an alias for a slice of pointers to Multiplier.
	// This should almost always be used instead of []Multiplier.
	MultiplierSlice []*Multiplier
	// MultiplierHook is the signature for custom Multiplier hook methods
	MultiplierHook func(boil.Executor, *Multiplier) error

	multiplierQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	multiplierType                 = reflect.TypeOf(&Multiplier{})
	multiplierMapping              = queries.MakeStructMapping(multiplierType)
	multiplierPrimaryKeyMapping, _ = queries.BindMapping(multiplierType, multiplierMapping, multiplierPrimaryKeyColumns)
	multiplierInsertCacheMut       sync.RWMutex
	multiplierInsertCache          = make(map[string]insertCache)
	multiplierUpdateCacheMut       sync.RWMutex
	multiplierUpdateCache          = make(map[string]updateCache)
	multiplierUpsertCacheMut       sync.RWMutex
	multiplierUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var multiplierAfterSelectHooks []MultiplierHook

var multiplierBeforeInsertHooks []MultiplierHook
var multiplierAfterInsertHooks []MultiplierHook

var multiplierBeforeUpdateHooks []MultiplierHook
var multiplierAfterUpdateHooks []MultiplierHook

var multiplierBeforeDeleteHooks []MultiplierHook
var multiplierAfterDeleteHooks []MultiplierHook

var multiplierBeforeUpsertHooks []MultiplierHook
var multiplierAfterUpsertHooks []MultiplierHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Multiplier) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range multiplierAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Multiplier) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range multiplierBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Multiplier) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range multiplierAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Multiplier) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range multiplierBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Multiplier) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range multiplierAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Multiplier) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range multiplierBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Multiplier) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range multiplierAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Multiplier) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range multiplierBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Multiplier) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range multiplierAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddMultiplierHook registers your hook function for all future operations.
func AddMultiplierHook(hookPoint boil.HookPoint, multiplierHook MultiplierHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		multiplierAfterSelectHooks = append(multiplierAfterSelectHooks, multiplierHook)
	case boil.BeforeInsertHook:
		multiplierBeforeInsertHooks = append(multiplierBeforeInsertHooks, multiplierHook)
	case boil.AfterInsertHook:
		multiplierAfterInsertHooks = append(multiplierAfterInsertHooks, multiplierHook)
	case boil.BeforeUpdateHook:
		multiplierBeforeUpdateHooks = append(multiplierBeforeUpdateHooks, multiplierHook)
	case boil.AfterUpdateHook:
		multiplierAfterUpdateHooks = append(multiplierAfterUpdateHooks, multiplierHook)
	case boil.BeforeDeleteHook:
		multiplierBeforeDeleteHooks = append(multiplierBeforeDeleteHooks, multiplierHook)
	case boil.AfterDeleteHook:
		multiplierAfterDeleteHooks = append(multiplierAfterDeleteHooks, multiplierHook)
	case boil.BeforeUpsertHook:
		multiplierBeforeUpsertHooks = append(multiplierBeforeUpsertHooks, multiplierHook)
	case boil.AfterUpsertHook:
		multiplierAfterUpsertHooks = append(multiplierAfterUpsertHooks, multiplierHook)
	}
}

// One returns a single multiplier record from the query.
func (q multiplierQuery) One(exec boil.Executor) (*Multiplier, error) {
	o := &Multiplier{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for multipliers")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Multiplier records from the query.
func (q multiplierQuery) All(exec boil.Executor) (MultiplierSlice, error) {
	var o []*Multiplier

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to Multiplier slice")
	}

	if len(multiplierAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Multiplier records in the query.
func (q multiplierQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count multipliers rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q multiplierQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if multipliers exists")
	}

	return count > 0, nil
}

// UserMultipliers retrieves all the user_multiplier's UserMultipliers with an executor.
func (o *Multiplier) UserMultipliers(mods ...qm.QueryMod) userMultiplierQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"user_multipliers\".\"multiplier_id\"=?", o.ID),
	)

	query := UserMultipliers(queryMods...)
	queries.SetFrom(query.Query, "\"user_multipliers\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"user_multipliers\".*"})
	}

	return query
}

// LoadUserMultipliers allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (multiplierL) LoadUserMultipliers(e boil.Executor, singular bool, maybeMultiplier interface{}, mods queries.Applicator) error {
	var slice []*Multiplier
	var object *Multiplier

	if singular {
		object = maybeMultiplier.(*Multiplier)
	} else {
		slice = *maybeMultiplier.(*[]*Multiplier)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &multiplierR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &multiplierR{}
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
		qm.From(`user_multipliers`),
		qm.WhereIn(`user_multipliers.multiplier_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load user_multipliers")
	}

	var resultSlice []*UserMultiplier
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice user_multipliers")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on user_multipliers")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for user_multipliers")
	}

	if len(userMultiplierAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.UserMultipliers = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &userMultiplierR{}
			}
			foreign.R.Multiplier = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.MultiplierID {
				local.R.UserMultipliers = append(local.R.UserMultipliers, foreign)
				if foreign.R == nil {
					foreign.R = &userMultiplierR{}
				}
				foreign.R.Multiplier = local
				break
			}
		}
	}

	return nil
}

// AddUserMultipliers adds the given related objects to the existing relationships
// of the multiplier, optionally inserting them as new records.
// Appends related to o.R.UserMultipliers.
// Sets related.R.Multiplier appropriately.
func (o *Multiplier) AddUserMultipliers(exec boil.Executor, insert bool, related ...*UserMultiplier) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.MultiplierID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"user_multipliers\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"multiplier_id"}),
				strmangle.WhereClause("\"", "\"", 2, userMultiplierPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.PlayerID, rel.FromBattleNumber, rel.MultiplierID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.MultiplierID = o.ID
		}
	}

	if o.R == nil {
		o.R = &multiplierR{
			UserMultipliers: related,
		}
	} else {
		o.R.UserMultipliers = append(o.R.UserMultipliers, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &userMultiplierR{
				Multiplier: o,
			}
		} else {
			rel.R.Multiplier = o
		}
	}
	return nil
}

// Multipliers retrieves all the records using an executor.
func Multipliers(mods ...qm.QueryMod) multiplierQuery {
	mods = append(mods, qm.From("\"multipliers\""))
	return multiplierQuery{NewQuery(mods...)}
}

// FindMultiplier retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindMultiplier(exec boil.Executor, iD string, selectCols ...string) (*Multiplier, error) {
	multiplierObj := &Multiplier{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"multipliers\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, multiplierObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from multipliers")
	}

	if err = multiplierObj.doAfterSelectHooks(exec); err != nil {
		return multiplierObj, err
	}

	return multiplierObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Multiplier) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no multipliers provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(multiplierColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	multiplierInsertCacheMut.RLock()
	cache, cached := multiplierInsertCache[key]
	multiplierInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			multiplierAllColumns,
			multiplierColumnsWithDefault,
			multiplierColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(multiplierType, multiplierMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(multiplierType, multiplierMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"multipliers\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"multipliers\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into multipliers")
	}

	if !cached {
		multiplierInsertCacheMut.Lock()
		multiplierInsertCache[key] = cache
		multiplierInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the Multiplier.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Multiplier) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	multiplierUpdateCacheMut.RLock()
	cache, cached := multiplierUpdateCache[key]
	multiplierUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			multiplierAllColumns,
			multiplierPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update multipliers, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"multipliers\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, multiplierPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(multiplierType, multiplierMapping, append(wl, multiplierPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update multipliers row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for multipliers")
	}

	if !cached {
		multiplierUpdateCacheMut.Lock()
		multiplierUpdateCache[key] = cache
		multiplierUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q multiplierQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for multipliers")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for multipliers")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o MultiplierSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), multiplierPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"multipliers\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, multiplierPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in multiplier slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all multiplier")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Multiplier) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no multipliers provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(multiplierColumnsWithDefault, o)

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

	multiplierUpsertCacheMut.RLock()
	cache, cached := multiplierUpsertCache[key]
	multiplierUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			multiplierAllColumns,
			multiplierColumnsWithDefault,
			multiplierColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			multiplierAllColumns,
			multiplierPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert multipliers, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(multiplierPrimaryKeyColumns))
			copy(conflict, multiplierPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"multipliers\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(multiplierType, multiplierMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(multiplierType, multiplierMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert multipliers")
	}

	if !cached {
		multiplierUpsertCacheMut.Lock()
		multiplierUpsertCache[key] = cache
		multiplierUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single Multiplier record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Multiplier) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no Multiplier provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), multiplierPrimaryKeyMapping)
	sql := "DELETE FROM \"multipliers\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from multipliers")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for multipliers")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q multiplierQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no multiplierQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from multipliers")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for multipliers")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o MultiplierSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(multiplierBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), multiplierPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"multipliers\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, multiplierPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from multiplier slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for multipliers")
	}

	if len(multiplierAfterDeleteHooks) != 0 {
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
func (o *Multiplier) Reload(exec boil.Executor) error {
	ret, err := FindMultiplier(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *MultiplierSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := MultiplierSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), multiplierPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"multipliers\".* FROM \"multipliers\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, multiplierPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in MultiplierSlice")
	}

	*o = slice

	return nil
}

// MultiplierExists checks if the Multiplier row exists.
func MultiplierExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"multipliers\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if multipliers exists")
	}

	return exists, nil
}
