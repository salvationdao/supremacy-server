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

// CouponItem is an object representing the database table.
type CouponItem struct {
	ID            string              `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	CouponID      string              `boiler:"coupon_id" boil:"coupon_id" json:"coupon_id" toml:"coupon_id" yaml:"coupon_id"`
	ItemType      string              `boiler:"item_type" boil:"item_type" json:"item_type" toml:"item_type" yaml:"item_type"`
	ItemID        null.String         `boiler:"item_id" boil:"item_id" json:"item_id,omitempty" toml:"item_id" yaml:"item_id,omitempty"`
	Claimed       bool                `boiler:"claimed" boil:"claimed" json:"claimed" toml:"claimed" yaml:"claimed"`
	Amount        decimal.NullDecimal `boiler:"amount" boil:"amount" json:"amount,omitempty" toml:"amount" yaml:"amount,omitempty"`
	TransactionID null.String         `boiler:"transaction_id" boil:"transaction_id" json:"transaction_id,omitempty" toml:"transaction_id" yaml:"transaction_id,omitempty"`
	CreatedAt     time.Time           `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *couponItemR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L couponItemL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var CouponItemColumns = struct {
	ID            string
	CouponID      string
	ItemType      string
	ItemID        string
	Claimed       string
	Amount        string
	TransactionID string
	CreatedAt     string
}{
	ID:            "id",
	CouponID:      "coupon_id",
	ItemType:      "item_type",
	ItemID:        "item_id",
	Claimed:       "claimed",
	Amount:        "amount",
	TransactionID: "transaction_id",
	CreatedAt:     "created_at",
}

var CouponItemTableColumns = struct {
	ID            string
	CouponID      string
	ItemType      string
	ItemID        string
	Claimed       string
	Amount        string
	TransactionID string
	CreatedAt     string
}{
	ID:            "coupon_items.id",
	CouponID:      "coupon_items.coupon_id",
	ItemType:      "coupon_items.item_type",
	ItemID:        "coupon_items.item_id",
	Claimed:       "coupon_items.claimed",
	Amount:        "coupon_items.amount",
	TransactionID: "coupon_items.transaction_id",
	CreatedAt:     "coupon_items.created_at",
}

// Generated where

var CouponItemWhere = struct {
	ID            whereHelperstring
	CouponID      whereHelperstring
	ItemType      whereHelperstring
	ItemID        whereHelpernull_String
	Claimed       whereHelperbool
	Amount        whereHelperdecimal_NullDecimal
	TransactionID whereHelpernull_String
	CreatedAt     whereHelpertime_Time
}{
	ID:            whereHelperstring{field: "\"coupon_items\".\"id\""},
	CouponID:      whereHelperstring{field: "\"coupon_items\".\"coupon_id\""},
	ItemType:      whereHelperstring{field: "\"coupon_items\".\"item_type\""},
	ItemID:        whereHelpernull_String{field: "\"coupon_items\".\"item_id\""},
	Claimed:       whereHelperbool{field: "\"coupon_items\".\"claimed\""},
	Amount:        whereHelperdecimal_NullDecimal{field: "\"coupon_items\".\"amount\""},
	TransactionID: whereHelpernull_String{field: "\"coupon_items\".\"transaction_id\""},
	CreatedAt:     whereHelpertime_Time{field: "\"coupon_items\".\"created_at\""},
}

// CouponItemRels is where relationship names are stored.
var CouponItemRels = struct {
	Coupon string
}{
	Coupon: "Coupon",
}

// couponItemR is where relationships are stored.
type couponItemR struct {
	Coupon *Coupon `boiler:"Coupon" boil:"Coupon" json:"Coupon" toml:"Coupon" yaml:"Coupon"`
}

// NewStruct creates a new relationship struct
func (*couponItemR) NewStruct() *couponItemR {
	return &couponItemR{}
}

// couponItemL is where Load methods for each relationship are stored.
type couponItemL struct{}

var (
	couponItemAllColumns            = []string{"id", "coupon_id", "item_type", "item_id", "claimed", "amount", "transaction_id", "created_at"}
	couponItemColumnsWithoutDefault = []string{"coupon_id", "item_type"}
	couponItemColumnsWithDefault    = []string{"id", "item_id", "claimed", "amount", "transaction_id", "created_at"}
	couponItemPrimaryKeyColumns     = []string{"id"}
	couponItemGeneratedColumns      = []string{}
)

type (
	// CouponItemSlice is an alias for a slice of pointers to CouponItem.
	// This should almost always be used instead of []CouponItem.
	CouponItemSlice []*CouponItem
	// CouponItemHook is the signature for custom CouponItem hook methods
	CouponItemHook func(boil.Executor, *CouponItem) error

	couponItemQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	couponItemType                 = reflect.TypeOf(&CouponItem{})
	couponItemMapping              = queries.MakeStructMapping(couponItemType)
	couponItemPrimaryKeyMapping, _ = queries.BindMapping(couponItemType, couponItemMapping, couponItemPrimaryKeyColumns)
	couponItemInsertCacheMut       sync.RWMutex
	couponItemInsertCache          = make(map[string]insertCache)
	couponItemUpdateCacheMut       sync.RWMutex
	couponItemUpdateCache          = make(map[string]updateCache)
	couponItemUpsertCacheMut       sync.RWMutex
	couponItemUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var couponItemAfterSelectHooks []CouponItemHook

var couponItemBeforeInsertHooks []CouponItemHook
var couponItemAfterInsertHooks []CouponItemHook

var couponItemBeforeUpdateHooks []CouponItemHook
var couponItemAfterUpdateHooks []CouponItemHook

var couponItemBeforeDeleteHooks []CouponItemHook
var couponItemAfterDeleteHooks []CouponItemHook

var couponItemBeforeUpsertHooks []CouponItemHook
var couponItemAfterUpsertHooks []CouponItemHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *CouponItem) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range couponItemAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *CouponItem) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range couponItemBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *CouponItem) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range couponItemAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *CouponItem) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range couponItemBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *CouponItem) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range couponItemAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *CouponItem) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range couponItemBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *CouponItem) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range couponItemAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *CouponItem) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range couponItemBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *CouponItem) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range couponItemAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddCouponItemHook registers your hook function for all future operations.
func AddCouponItemHook(hookPoint boil.HookPoint, couponItemHook CouponItemHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		couponItemAfterSelectHooks = append(couponItemAfterSelectHooks, couponItemHook)
	case boil.BeforeInsertHook:
		couponItemBeforeInsertHooks = append(couponItemBeforeInsertHooks, couponItemHook)
	case boil.AfterInsertHook:
		couponItemAfterInsertHooks = append(couponItemAfterInsertHooks, couponItemHook)
	case boil.BeforeUpdateHook:
		couponItemBeforeUpdateHooks = append(couponItemBeforeUpdateHooks, couponItemHook)
	case boil.AfterUpdateHook:
		couponItemAfterUpdateHooks = append(couponItemAfterUpdateHooks, couponItemHook)
	case boil.BeforeDeleteHook:
		couponItemBeforeDeleteHooks = append(couponItemBeforeDeleteHooks, couponItemHook)
	case boil.AfterDeleteHook:
		couponItemAfterDeleteHooks = append(couponItemAfterDeleteHooks, couponItemHook)
	case boil.BeforeUpsertHook:
		couponItemBeforeUpsertHooks = append(couponItemBeforeUpsertHooks, couponItemHook)
	case boil.AfterUpsertHook:
		couponItemAfterUpsertHooks = append(couponItemAfterUpsertHooks, couponItemHook)
	}
}

// One returns a single couponItem record from the query.
func (q couponItemQuery) One(exec boil.Executor) (*CouponItem, error) {
	o := &CouponItem{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for coupon_items")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all CouponItem records from the query.
func (q couponItemQuery) All(exec boil.Executor) (CouponItemSlice, error) {
	var o []*CouponItem

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to CouponItem slice")
	}

	if len(couponItemAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all CouponItem records in the query.
func (q couponItemQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count coupon_items rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q couponItemQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if coupon_items exists")
	}

	return count > 0, nil
}

// Coupon pointed to by the foreign key.
func (o *CouponItem) Coupon(mods ...qm.QueryMod) couponQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.CouponID),
	}

	queryMods = append(queryMods, mods...)

	query := Coupons(queryMods...)
	queries.SetFrom(query.Query, "\"coupons\"")

	return query
}

// LoadCoupon allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (couponItemL) LoadCoupon(e boil.Executor, singular bool, maybeCouponItem interface{}, mods queries.Applicator) error {
	var slice []*CouponItem
	var object *CouponItem

	if singular {
		object = maybeCouponItem.(*CouponItem)
	} else {
		slice = *maybeCouponItem.(*[]*CouponItem)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &couponItemR{}
		}
		args = append(args, object.CouponID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &couponItemR{}
			}

			for _, a := range args {
				if a == obj.CouponID {
					continue Outer
				}
			}

			args = append(args, obj.CouponID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`coupons`),
		qm.WhereIn(`coupons.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Coupon")
	}

	var resultSlice []*Coupon
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Coupon")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for coupons")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for coupons")
	}

	if len(couponItemAfterSelectHooks) != 0 {
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
		object.R.Coupon = foreign
		if foreign.R == nil {
			foreign.R = &couponR{}
		}
		foreign.R.CouponItems = append(foreign.R.CouponItems, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.CouponID == foreign.ID {
				local.R.Coupon = foreign
				if foreign.R == nil {
					foreign.R = &couponR{}
				}
				foreign.R.CouponItems = append(foreign.R.CouponItems, local)
				break
			}
		}
	}

	return nil
}

// SetCoupon of the couponItem to the related item.
// Sets o.R.Coupon to related.
// Adds o to related.R.CouponItems.
func (o *CouponItem) SetCoupon(exec boil.Executor, insert bool, related *Coupon) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"coupon_items\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"coupon_id"}),
		strmangle.WhereClause("\"", "\"", 2, couponItemPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.CouponID = related.ID
	if o.R == nil {
		o.R = &couponItemR{
			Coupon: related,
		}
	} else {
		o.R.Coupon = related
	}

	if related.R == nil {
		related.R = &couponR{
			CouponItems: CouponItemSlice{o},
		}
	} else {
		related.R.CouponItems = append(related.R.CouponItems, o)
	}

	return nil
}

// CouponItems retrieves all the records using an executor.
func CouponItems(mods ...qm.QueryMod) couponItemQuery {
	mods = append(mods, qm.From("\"coupon_items\""))
	return couponItemQuery{NewQuery(mods...)}
}

// FindCouponItem retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindCouponItem(exec boil.Executor, iD string, selectCols ...string) (*CouponItem, error) {
	couponItemObj := &CouponItem{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"coupon_items\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, couponItemObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from coupon_items")
	}

	if err = couponItemObj.doAfterSelectHooks(exec); err != nil {
		return couponItemObj, err
	}

	return couponItemObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *CouponItem) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no coupon_items provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(couponItemColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	couponItemInsertCacheMut.RLock()
	cache, cached := couponItemInsertCache[key]
	couponItemInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			couponItemAllColumns,
			couponItemColumnsWithDefault,
			couponItemColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(couponItemType, couponItemMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(couponItemType, couponItemMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"coupon_items\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"coupon_items\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into coupon_items")
	}

	if !cached {
		couponItemInsertCacheMut.Lock()
		couponItemInsertCache[key] = cache
		couponItemInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the CouponItem.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *CouponItem) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	couponItemUpdateCacheMut.RLock()
	cache, cached := couponItemUpdateCache[key]
	couponItemUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			couponItemAllColumns,
			couponItemPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update coupon_items, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"coupon_items\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, couponItemPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(couponItemType, couponItemMapping, append(wl, couponItemPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update coupon_items row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for coupon_items")
	}

	if !cached {
		couponItemUpdateCacheMut.Lock()
		couponItemUpdateCache[key] = cache
		couponItemUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q couponItemQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for coupon_items")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for coupon_items")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o CouponItemSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), couponItemPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"coupon_items\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, couponItemPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in couponItem slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all couponItem")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *CouponItem) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no coupon_items provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(couponItemColumnsWithDefault, o)

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

	couponItemUpsertCacheMut.RLock()
	cache, cached := couponItemUpsertCache[key]
	couponItemUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			couponItemAllColumns,
			couponItemColumnsWithDefault,
			couponItemColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			couponItemAllColumns,
			couponItemPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert coupon_items, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(couponItemPrimaryKeyColumns))
			copy(conflict, couponItemPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"coupon_items\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(couponItemType, couponItemMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(couponItemType, couponItemMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert coupon_items")
	}

	if !cached {
		couponItemUpsertCacheMut.Lock()
		couponItemUpsertCache[key] = cache
		couponItemUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single CouponItem record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *CouponItem) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no CouponItem provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), couponItemPrimaryKeyMapping)
	sql := "DELETE FROM \"coupon_items\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from coupon_items")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for coupon_items")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q couponItemQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no couponItemQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from coupon_items")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for coupon_items")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o CouponItemSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(couponItemBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), couponItemPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"coupon_items\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, couponItemPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from couponItem slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for coupon_items")
	}

	if len(couponItemAfterDeleteHooks) != 0 {
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
func (o *CouponItem) Reload(exec boil.Executor) error {
	ret, err := FindCouponItem(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *CouponItemSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := CouponItemSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), couponItemPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"coupon_items\".* FROM \"coupon_items\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, couponItemPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in CouponItemSlice")
	}

	*o = slice

	return nil
}

// CouponItemExists checks if the CouponItem row exists.
func CouponItemExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"coupon_items\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if coupon_items exists")
	}

	return exists, nil
}