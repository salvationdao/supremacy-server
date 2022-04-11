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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// AssetRepair is an object representing the database table.
type AssetRepair struct {
	Hash              string          `boiler:"hash" boil:"hash" json:"hash" toml:"hash" yaml:"hash"`
	ExpectCompletedAt time.Time       `boiler:"expect_completed_at" boil:"expect_completed_at" json:"expect_completed_at" toml:"expect_completed_at" yaml:"expect_completed_at"`
	RepairMode        string          `boiler:"repair_mode" boil:"repair_mode" json:"repair_mode" toml:"repair_mode" yaml:"repair_mode"`
	IsPaidToComplete  bool            `boiler:"is_paid_to_complete" boil:"is_paid_to_complete" json:"is_paid_to_complete" toml:"is_paid_to_complete" yaml:"is_paid_to_complete"`
	CompletedAt       null.Time       `boiler:"completed_at" boil:"completed_at" json:"completed_at,omitempty" toml:"completed_at" yaml:"completed_at,omitempty"`
	CreatedAt         time.Time       `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	ID                string          `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	FullRepairFee     decimal.Decimal `boiler:"full_repair_fee" boil:"full_repair_fee" json:"full_repair_fee" toml:"full_repair_fee" yaml:"full_repair_fee"`

	R *assetRepairR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L assetRepairL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var AssetRepairColumns = struct {
	Hash              string
	ExpectCompletedAt string
	RepairMode        string
	IsPaidToComplete  string
	CompletedAt       string
	CreatedAt         string
	ID                string
	FullRepairFee     string
}{
	Hash:              "hash",
	ExpectCompletedAt: "expect_completed_at",
	RepairMode:        "repair_mode",
	IsPaidToComplete:  "is_paid_to_complete",
	CompletedAt:       "completed_at",
	CreatedAt:         "created_at",
	ID:                "id",
	FullRepairFee:     "full_repair_fee",
}

var AssetRepairTableColumns = struct {
	Hash              string
	ExpectCompletedAt string
	RepairMode        string
	IsPaidToComplete  string
	CompletedAt       string
	CreatedAt         string
	ID                string
	FullRepairFee     string
}{
	Hash:              "asset_repair.hash",
	ExpectCompletedAt: "asset_repair.expect_completed_at",
	RepairMode:        "asset_repair.repair_mode",
	IsPaidToComplete:  "asset_repair.is_paid_to_complete",
	CompletedAt:       "asset_repair.completed_at",
	CreatedAt:         "asset_repair.created_at",
	ID:                "asset_repair.id",
	FullRepairFee:     "asset_repair.full_repair_fee",
}

// Generated where

type whereHelperstring struct{ field string }

func (w whereHelperstring) EQ(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperstring) NEQ(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperstring) LT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperstring) LTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperstring) GT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperstring) GTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperstring) IN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperstring) NIN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelpertime_Time struct{ field string }

func (w whereHelpertime_Time) EQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertime_Time) NEQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertime_Time) LT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertime_Time) LTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertime_Time) GT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertime_Time) GTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

type whereHelperbool struct{ field string }

func (w whereHelperbool) EQ(x bool) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperbool) NEQ(x bool) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperbool) LT(x bool) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperbool) LTE(x bool) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperbool) GT(x bool) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperbool) GTE(x bool) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }

type whereHelpernull_Time struct{ field string }

func (w whereHelpernull_Time) EQ(x null.Time) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Time) NEQ(x null.Time) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Time) LT(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Time) LTE(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Time) GT(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Time) GTE(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpernull_Time) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Time) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

type whereHelperdecimal_Decimal struct{ field string }

func (w whereHelperdecimal_Decimal) EQ(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelperdecimal_Decimal) NEQ(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelperdecimal_Decimal) LT(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelperdecimal_Decimal) LTE(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelperdecimal_Decimal) GT(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelperdecimal_Decimal) GTE(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var AssetRepairWhere = struct {
	Hash              whereHelperstring
	ExpectCompletedAt whereHelpertime_Time
	RepairMode        whereHelperstring
	IsPaidToComplete  whereHelperbool
	CompletedAt       whereHelpernull_Time
	CreatedAt         whereHelpertime_Time
	ID                whereHelperstring
	FullRepairFee     whereHelperdecimal_Decimal
}{
	Hash:              whereHelperstring{field: "\"asset_repair\".\"hash\""},
	ExpectCompletedAt: whereHelpertime_Time{field: "\"asset_repair\".\"expect_completed_at\""},
	RepairMode:        whereHelperstring{field: "\"asset_repair\".\"repair_mode\""},
	IsPaidToComplete:  whereHelperbool{field: "\"asset_repair\".\"is_paid_to_complete\""},
	CompletedAt:       whereHelpernull_Time{field: "\"asset_repair\".\"completed_at\""},
	CreatedAt:         whereHelpertime_Time{field: "\"asset_repair\".\"created_at\""},
	ID:                whereHelperstring{field: "\"asset_repair\".\"id\""},
	FullRepairFee:     whereHelperdecimal_Decimal{field: "\"asset_repair\".\"full_repair_fee\""},
}

// AssetRepairRels is where relationship names are stored.
var AssetRepairRels = struct {
}{}

// assetRepairR is where relationships are stored.
type assetRepairR struct {
}

// NewStruct creates a new relationship struct
func (*assetRepairR) NewStruct() *assetRepairR {
	return &assetRepairR{}
}

// assetRepairL is where Load methods for each relationship are stored.
type assetRepairL struct{}

var (
	assetRepairAllColumns            = []string{"hash", "expect_completed_at", "repair_mode", "is_paid_to_complete", "completed_at", "created_at", "id", "full_repair_fee"}
	assetRepairColumnsWithoutDefault = []string{"hash", "expect_completed_at", "repair_mode"}
	assetRepairColumnsWithDefault    = []string{"is_paid_to_complete", "completed_at", "created_at", "id", "full_repair_fee"}
	assetRepairPrimaryKeyColumns     = []string{"id"}
	assetRepairGeneratedColumns      = []string{}
)

type (
	// AssetRepairSlice is an alias for a slice of pointers to AssetRepair.
	// This should almost always be used instead of []AssetRepair.
	AssetRepairSlice []*AssetRepair
	// AssetRepairHook is the signature for custom AssetRepair hook methods
	AssetRepairHook func(boil.Executor, *AssetRepair) error

	assetRepairQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	assetRepairType                 = reflect.TypeOf(&AssetRepair{})
	assetRepairMapping              = queries.MakeStructMapping(assetRepairType)
	assetRepairPrimaryKeyMapping, _ = queries.BindMapping(assetRepairType, assetRepairMapping, assetRepairPrimaryKeyColumns)
	assetRepairInsertCacheMut       sync.RWMutex
	assetRepairInsertCache          = make(map[string]insertCache)
	assetRepairUpdateCacheMut       sync.RWMutex
	assetRepairUpdateCache          = make(map[string]updateCache)
	assetRepairUpsertCacheMut       sync.RWMutex
	assetRepairUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var assetRepairAfterSelectHooks []AssetRepairHook

var assetRepairBeforeInsertHooks []AssetRepairHook
var assetRepairAfterInsertHooks []AssetRepairHook

var assetRepairBeforeUpdateHooks []AssetRepairHook
var assetRepairAfterUpdateHooks []AssetRepairHook

var assetRepairBeforeDeleteHooks []AssetRepairHook
var assetRepairAfterDeleteHooks []AssetRepairHook

var assetRepairBeforeUpsertHooks []AssetRepairHook
var assetRepairAfterUpsertHooks []AssetRepairHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *AssetRepair) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range assetRepairAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *AssetRepair) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range assetRepairBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *AssetRepair) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range assetRepairAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *AssetRepair) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range assetRepairBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *AssetRepair) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range assetRepairAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *AssetRepair) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range assetRepairBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *AssetRepair) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range assetRepairAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *AssetRepair) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range assetRepairBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *AssetRepair) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range assetRepairAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddAssetRepairHook registers your hook function for all future operations.
func AddAssetRepairHook(hookPoint boil.HookPoint, assetRepairHook AssetRepairHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		assetRepairAfterSelectHooks = append(assetRepairAfterSelectHooks, assetRepairHook)
	case boil.BeforeInsertHook:
		assetRepairBeforeInsertHooks = append(assetRepairBeforeInsertHooks, assetRepairHook)
	case boil.AfterInsertHook:
		assetRepairAfterInsertHooks = append(assetRepairAfterInsertHooks, assetRepairHook)
	case boil.BeforeUpdateHook:
		assetRepairBeforeUpdateHooks = append(assetRepairBeforeUpdateHooks, assetRepairHook)
	case boil.AfterUpdateHook:
		assetRepairAfterUpdateHooks = append(assetRepairAfterUpdateHooks, assetRepairHook)
	case boil.BeforeDeleteHook:
		assetRepairBeforeDeleteHooks = append(assetRepairBeforeDeleteHooks, assetRepairHook)
	case boil.AfterDeleteHook:
		assetRepairAfterDeleteHooks = append(assetRepairAfterDeleteHooks, assetRepairHook)
	case boil.BeforeUpsertHook:
		assetRepairBeforeUpsertHooks = append(assetRepairBeforeUpsertHooks, assetRepairHook)
	case boil.AfterUpsertHook:
		assetRepairAfterUpsertHooks = append(assetRepairAfterUpsertHooks, assetRepairHook)
	}
}

// One returns a single assetRepair record from the query.
func (q assetRepairQuery) One(exec boil.Executor) (*AssetRepair, error) {
	o := &AssetRepair{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for asset_repair")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all AssetRepair records from the query.
func (q assetRepairQuery) All(exec boil.Executor) (AssetRepairSlice, error) {
	var o []*AssetRepair

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to AssetRepair slice")
	}

	if len(assetRepairAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all AssetRepair records in the query.
func (q assetRepairQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count asset_repair rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q assetRepairQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if asset_repair exists")
	}

	return count > 0, nil
}

// AssetRepairs retrieves all the records using an executor.
func AssetRepairs(mods ...qm.QueryMod) assetRepairQuery {
	mods = append(mods, qm.From("\"asset_repair\""))
	return assetRepairQuery{NewQuery(mods...)}
}

// FindAssetRepair retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindAssetRepair(exec boil.Executor, iD string, selectCols ...string) (*AssetRepair, error) {
	assetRepairObj := &AssetRepair{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"asset_repair\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, assetRepairObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from asset_repair")
	}

	if err = assetRepairObj.doAfterSelectHooks(exec); err != nil {
		return assetRepairObj, err
	}

	return assetRepairObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *AssetRepair) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no asset_repair provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(assetRepairColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	assetRepairInsertCacheMut.RLock()
	cache, cached := assetRepairInsertCache[key]
	assetRepairInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			assetRepairAllColumns,
			assetRepairColumnsWithDefault,
			assetRepairColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(assetRepairType, assetRepairMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(assetRepairType, assetRepairMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"asset_repair\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"asset_repair\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into asset_repair")
	}

	if !cached {
		assetRepairInsertCacheMut.Lock()
		assetRepairInsertCache[key] = cache
		assetRepairInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the AssetRepair.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *AssetRepair) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	assetRepairUpdateCacheMut.RLock()
	cache, cached := assetRepairUpdateCache[key]
	assetRepairUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			assetRepairAllColumns,
			assetRepairPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update asset_repair, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"asset_repair\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, assetRepairPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(assetRepairType, assetRepairMapping, append(wl, assetRepairPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update asset_repair row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for asset_repair")
	}

	if !cached {
		assetRepairUpdateCacheMut.Lock()
		assetRepairUpdateCache[key] = cache
		assetRepairUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q assetRepairQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for asset_repair")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for asset_repair")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o AssetRepairSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), assetRepairPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"asset_repair\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, assetRepairPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in assetRepair slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all assetRepair")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *AssetRepair) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no asset_repair provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(assetRepairColumnsWithDefault, o)

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

	assetRepairUpsertCacheMut.RLock()
	cache, cached := assetRepairUpsertCache[key]
	assetRepairUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			assetRepairAllColumns,
			assetRepairColumnsWithDefault,
			assetRepairColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			assetRepairAllColumns,
			assetRepairPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert asset_repair, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(assetRepairPrimaryKeyColumns))
			copy(conflict, assetRepairPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"asset_repair\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(assetRepairType, assetRepairMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(assetRepairType, assetRepairMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert asset_repair")
	}

	if !cached {
		assetRepairUpsertCacheMut.Lock()
		assetRepairUpsertCache[key] = cache
		assetRepairUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single AssetRepair record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *AssetRepair) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no AssetRepair provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), assetRepairPrimaryKeyMapping)
	sql := "DELETE FROM \"asset_repair\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from asset_repair")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for asset_repair")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q assetRepairQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no assetRepairQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from asset_repair")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for asset_repair")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o AssetRepairSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(assetRepairBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), assetRepairPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"asset_repair\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, assetRepairPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from assetRepair slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for asset_repair")
	}

	if len(assetRepairAfterDeleteHooks) != 0 {
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
func (o *AssetRepair) Reload(exec boil.Executor) error {
	ret, err := FindAssetRepair(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *AssetRepairSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := AssetRepairSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), assetRepairPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"asset_repair\".* FROM \"asset_repair\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, assetRepairPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in AssetRepairSlice")
	}

	*o = slice

	return nil
}

// AssetRepairExists checks if the AssetRepair row exists.
func AssetRepairExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"asset_repair\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if asset_repair exists")
	}

	return exists, nil
}
