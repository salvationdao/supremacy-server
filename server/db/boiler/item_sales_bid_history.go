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

// ItemSalesBidHistory is an object representing the database table.
type ItemSalesBidHistory struct {
	ItemSaleID  string    `boiler:"item_sale_id" boil:"item_sale_id" json:"item_sale_id" toml:"item_sale_id" yaml:"item_sale_id"`
	BidderID    string    `boiler:"bidder_id" boil:"bidder_id" json:"bidder_id" toml:"bidder_id" yaml:"bidder_id"`
	BidAt       time.Time `boiler:"bid_at" boil:"bid_at" json:"bid_at" toml:"bid_at" yaml:"bid_at"`
	BidPrice    string    `boiler:"bid_price" boil:"bid_price" json:"bid_price" toml:"bid_price" yaml:"bid_price"`
	CancelledAt null.Time `boiler:"cancelled_at" boil:"cancelled_at" json:"cancelled_at,omitempty" toml:"cancelled_at" yaml:"cancelled_at,omitempty"`

	R *itemSalesBidHistoryR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L itemSalesBidHistoryL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ItemSalesBidHistoryColumns = struct {
	ItemSaleID  string
	BidderID    string
	BidAt       string
	BidPrice    string
	CancelledAt string
}{
	ItemSaleID:  "item_sale_id",
	BidderID:    "bidder_id",
	BidAt:       "bid_at",
	BidPrice:    "bid_price",
	CancelledAt: "cancelled_at",
}

var ItemSalesBidHistoryTableColumns = struct {
	ItemSaleID  string
	BidderID    string
	BidAt       string
	BidPrice    string
	CancelledAt string
}{
	ItemSaleID:  "item_sales_bid_history.item_sale_id",
	BidderID:    "item_sales_bid_history.bidder_id",
	BidAt:       "item_sales_bid_history.bid_at",
	BidPrice:    "item_sales_bid_history.bid_price",
	CancelledAt: "item_sales_bid_history.cancelled_at",
}

// Generated where

var ItemSalesBidHistoryWhere = struct {
	ItemSaleID  whereHelperstring
	BidderID    whereHelperstring
	BidAt       whereHelpertime_Time
	BidPrice    whereHelperstring
	CancelledAt whereHelpernull_Time
}{
	ItemSaleID:  whereHelperstring{field: "\"item_sales_bid_history\".\"item_sale_id\""},
	BidderID:    whereHelperstring{field: "\"item_sales_bid_history\".\"bidder_id\""},
	BidAt:       whereHelpertime_Time{field: "\"item_sales_bid_history\".\"bid_at\""},
	BidPrice:    whereHelperstring{field: "\"item_sales_bid_history\".\"bid_price\""},
	CancelledAt: whereHelpernull_Time{field: "\"item_sales_bid_history\".\"cancelled_at\""},
}

// ItemSalesBidHistoryRels is where relationship names are stored.
var ItemSalesBidHistoryRels = struct {
	Bidder   string
	ItemSale string
}{
	Bidder:   "Bidder",
	ItemSale: "ItemSale",
}

// itemSalesBidHistoryR is where relationships are stored.
type itemSalesBidHistoryR struct {
	Bidder   *Player   `boiler:"Bidder" boil:"Bidder" json:"Bidder" toml:"Bidder" yaml:"Bidder"`
	ItemSale *ItemSale `boiler:"ItemSale" boil:"ItemSale" json:"ItemSale" toml:"ItemSale" yaml:"ItemSale"`
}

// NewStruct creates a new relationship struct
func (*itemSalesBidHistoryR) NewStruct() *itemSalesBidHistoryR {
	return &itemSalesBidHistoryR{}
}

// itemSalesBidHistoryL is where Load methods for each relationship are stored.
type itemSalesBidHistoryL struct{}

var (
	itemSalesBidHistoryAllColumns            = []string{"item_sale_id", "bidder_id", "bid_at", "bid_price", "cancelled_at"}
	itemSalesBidHistoryColumnsWithoutDefault = []string{"item_sale_id", "bidder_id", "bid_price"}
	itemSalesBidHistoryColumnsWithDefault    = []string{"bid_at", "cancelled_at"}
	itemSalesBidHistoryPrimaryKeyColumns     = []string{"item_sale_id", "bidder_id", "bid_at"}
	itemSalesBidHistoryGeneratedColumns      = []string{}
)

type (
	// ItemSalesBidHistorySlice is an alias for a slice of pointers to ItemSalesBidHistory.
	// This should almost always be used instead of []ItemSalesBidHistory.
	ItemSalesBidHistorySlice []*ItemSalesBidHistory
	// ItemSalesBidHistoryHook is the signature for custom ItemSalesBidHistory hook methods
	ItemSalesBidHistoryHook func(boil.Executor, *ItemSalesBidHistory) error

	itemSalesBidHistoryQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	itemSalesBidHistoryType                 = reflect.TypeOf(&ItemSalesBidHistory{})
	itemSalesBidHistoryMapping              = queries.MakeStructMapping(itemSalesBidHistoryType)
	itemSalesBidHistoryPrimaryKeyMapping, _ = queries.BindMapping(itemSalesBidHistoryType, itemSalesBidHistoryMapping, itemSalesBidHistoryPrimaryKeyColumns)
	itemSalesBidHistoryInsertCacheMut       sync.RWMutex
	itemSalesBidHistoryInsertCache          = make(map[string]insertCache)
	itemSalesBidHistoryUpdateCacheMut       sync.RWMutex
	itemSalesBidHistoryUpdateCache          = make(map[string]updateCache)
	itemSalesBidHistoryUpsertCacheMut       sync.RWMutex
	itemSalesBidHistoryUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var itemSalesBidHistoryAfterSelectHooks []ItemSalesBidHistoryHook

var itemSalesBidHistoryBeforeInsertHooks []ItemSalesBidHistoryHook
var itemSalesBidHistoryAfterInsertHooks []ItemSalesBidHistoryHook

var itemSalesBidHistoryBeforeUpdateHooks []ItemSalesBidHistoryHook
var itemSalesBidHistoryAfterUpdateHooks []ItemSalesBidHistoryHook

var itemSalesBidHistoryBeforeDeleteHooks []ItemSalesBidHistoryHook
var itemSalesBidHistoryAfterDeleteHooks []ItemSalesBidHistoryHook

var itemSalesBidHistoryBeforeUpsertHooks []ItemSalesBidHistoryHook
var itemSalesBidHistoryAfterUpsertHooks []ItemSalesBidHistoryHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *ItemSalesBidHistory) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range itemSalesBidHistoryAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *ItemSalesBidHistory) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range itemSalesBidHistoryBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *ItemSalesBidHistory) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range itemSalesBidHistoryAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *ItemSalesBidHistory) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range itemSalesBidHistoryBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *ItemSalesBidHistory) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range itemSalesBidHistoryAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *ItemSalesBidHistory) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range itemSalesBidHistoryBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *ItemSalesBidHistory) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range itemSalesBidHistoryAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *ItemSalesBidHistory) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range itemSalesBidHistoryBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *ItemSalesBidHistory) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range itemSalesBidHistoryAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddItemSalesBidHistoryHook registers your hook function for all future operations.
func AddItemSalesBidHistoryHook(hookPoint boil.HookPoint, itemSalesBidHistoryHook ItemSalesBidHistoryHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		itemSalesBidHistoryAfterSelectHooks = append(itemSalesBidHistoryAfterSelectHooks, itemSalesBidHistoryHook)
	case boil.BeforeInsertHook:
		itemSalesBidHistoryBeforeInsertHooks = append(itemSalesBidHistoryBeforeInsertHooks, itemSalesBidHistoryHook)
	case boil.AfterInsertHook:
		itemSalesBidHistoryAfterInsertHooks = append(itemSalesBidHistoryAfterInsertHooks, itemSalesBidHistoryHook)
	case boil.BeforeUpdateHook:
		itemSalesBidHistoryBeforeUpdateHooks = append(itemSalesBidHistoryBeforeUpdateHooks, itemSalesBidHistoryHook)
	case boil.AfterUpdateHook:
		itemSalesBidHistoryAfterUpdateHooks = append(itemSalesBidHistoryAfterUpdateHooks, itemSalesBidHistoryHook)
	case boil.BeforeDeleteHook:
		itemSalesBidHistoryBeforeDeleteHooks = append(itemSalesBidHistoryBeforeDeleteHooks, itemSalesBidHistoryHook)
	case boil.AfterDeleteHook:
		itemSalesBidHistoryAfterDeleteHooks = append(itemSalesBidHistoryAfterDeleteHooks, itemSalesBidHistoryHook)
	case boil.BeforeUpsertHook:
		itemSalesBidHistoryBeforeUpsertHooks = append(itemSalesBidHistoryBeforeUpsertHooks, itemSalesBidHistoryHook)
	case boil.AfterUpsertHook:
		itemSalesBidHistoryAfterUpsertHooks = append(itemSalesBidHistoryAfterUpsertHooks, itemSalesBidHistoryHook)
	}
}

// One returns a single itemSalesBidHistory record from the query.
func (q itemSalesBidHistoryQuery) One(exec boil.Executor) (*ItemSalesBidHistory, error) {
	o := &ItemSalesBidHistory{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for item_sales_bid_history")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all ItemSalesBidHistory records from the query.
func (q itemSalesBidHistoryQuery) All(exec boil.Executor) (ItemSalesBidHistorySlice, error) {
	var o []*ItemSalesBidHistory

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to ItemSalesBidHistory slice")
	}

	if len(itemSalesBidHistoryAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all ItemSalesBidHistory records in the query.
func (q itemSalesBidHistoryQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count item_sales_bid_history rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q itemSalesBidHistoryQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if item_sales_bid_history exists")
	}

	return count > 0, nil
}

// Bidder pointed to by the foreign key.
func (o *ItemSalesBidHistory) Bidder(mods ...qm.QueryMod) playerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BidderID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Players(queryMods...)
	queries.SetFrom(query.Query, "\"players\"")

	return query
}

// ItemSale pointed to by the foreign key.
func (o *ItemSalesBidHistory) ItemSale(mods ...qm.QueryMod) itemSaleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.ItemSaleID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := ItemSales(queryMods...)
	queries.SetFrom(query.Query, "\"item_sales\"")

	return query
}

// LoadBidder allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (itemSalesBidHistoryL) LoadBidder(e boil.Executor, singular bool, maybeItemSalesBidHistory interface{}, mods queries.Applicator) error {
	var slice []*ItemSalesBidHistory
	var object *ItemSalesBidHistory

	if singular {
		object = maybeItemSalesBidHistory.(*ItemSalesBidHistory)
	} else {
		slice = *maybeItemSalesBidHistory.(*[]*ItemSalesBidHistory)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &itemSalesBidHistoryR{}
		}
		args = append(args, object.BidderID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &itemSalesBidHistoryR{}
			}

			for _, a := range args {
				if a == obj.BidderID {
					continue Outer
				}
			}

			args = append(args, obj.BidderID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`players`),
		qm.WhereIn(`players.id in ?`, args...),
		qmhelper.WhereIsNull(`players.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Player")
	}

	var resultSlice []*Player
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Player")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for players")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for players")
	}

	if len(itemSalesBidHistoryAfterSelectHooks) != 0 {
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
		object.R.Bidder = foreign
		if foreign.R == nil {
			foreign.R = &playerR{}
		}
		foreign.R.BidderItemSalesBidHistories = append(foreign.R.BidderItemSalesBidHistories, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.BidderID == foreign.ID {
				local.R.Bidder = foreign
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.BidderItemSalesBidHistories = append(foreign.R.BidderItemSalesBidHistories, local)
				break
			}
		}
	}

	return nil
}

// LoadItemSale allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (itemSalesBidHistoryL) LoadItemSale(e boil.Executor, singular bool, maybeItemSalesBidHistory interface{}, mods queries.Applicator) error {
	var slice []*ItemSalesBidHistory
	var object *ItemSalesBidHistory

	if singular {
		object = maybeItemSalesBidHistory.(*ItemSalesBidHistory)
	} else {
		slice = *maybeItemSalesBidHistory.(*[]*ItemSalesBidHistory)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &itemSalesBidHistoryR{}
		}
		args = append(args, object.ItemSaleID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &itemSalesBidHistoryR{}
			}

			for _, a := range args {
				if a == obj.ItemSaleID {
					continue Outer
				}
			}

			args = append(args, obj.ItemSaleID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`item_sales`),
		qm.WhereIn(`item_sales.id in ?`, args...),
		qmhelper.WhereIsNull(`item_sales.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load ItemSale")
	}

	var resultSlice []*ItemSale
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice ItemSale")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for item_sales")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for item_sales")
	}

	if len(itemSalesBidHistoryAfterSelectHooks) != 0 {
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
		object.R.ItemSale = foreign
		if foreign.R == nil {
			foreign.R = &itemSaleR{}
		}
		foreign.R.ItemSalesBidHistories = append(foreign.R.ItemSalesBidHistories, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.ItemSaleID == foreign.ID {
				local.R.ItemSale = foreign
				if foreign.R == nil {
					foreign.R = &itemSaleR{}
				}
				foreign.R.ItemSalesBidHistories = append(foreign.R.ItemSalesBidHistories, local)
				break
			}
		}
	}

	return nil
}

// SetBidder of the itemSalesBidHistory to the related item.
// Sets o.R.Bidder to related.
// Adds o to related.R.BidderItemSalesBidHistories.
func (o *ItemSalesBidHistory) SetBidder(exec boil.Executor, insert bool, related *Player) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"item_sales_bid_history\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"bidder_id"}),
		strmangle.WhereClause("\"", "\"", 2, itemSalesBidHistoryPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ItemSaleID, o.BidderID, o.BidAt}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.BidderID = related.ID
	if o.R == nil {
		o.R = &itemSalesBidHistoryR{
			Bidder: related,
		}
	} else {
		o.R.Bidder = related
	}

	if related.R == nil {
		related.R = &playerR{
			BidderItemSalesBidHistories: ItemSalesBidHistorySlice{o},
		}
	} else {
		related.R.BidderItemSalesBidHistories = append(related.R.BidderItemSalesBidHistories, o)
	}

	return nil
}

// SetItemSale of the itemSalesBidHistory to the related item.
// Sets o.R.ItemSale to related.
// Adds o to related.R.ItemSalesBidHistories.
func (o *ItemSalesBidHistory) SetItemSale(exec boil.Executor, insert bool, related *ItemSale) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"item_sales_bid_history\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"item_sale_id"}),
		strmangle.WhereClause("\"", "\"", 2, itemSalesBidHistoryPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ItemSaleID, o.BidderID, o.BidAt}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.ItemSaleID = related.ID
	if o.R == nil {
		o.R = &itemSalesBidHistoryR{
			ItemSale: related,
		}
	} else {
		o.R.ItemSale = related
	}

	if related.R == nil {
		related.R = &itemSaleR{
			ItemSalesBidHistories: ItemSalesBidHistorySlice{o},
		}
	} else {
		related.R.ItemSalesBidHistories = append(related.R.ItemSalesBidHistories, o)
	}

	return nil
}

// ItemSalesBidHistories retrieves all the records using an executor.
func ItemSalesBidHistories(mods ...qm.QueryMod) itemSalesBidHistoryQuery {
	mods = append(mods, qm.From("\"item_sales_bid_history\""))
	return itemSalesBidHistoryQuery{NewQuery(mods...)}
}

// FindItemSalesBidHistory retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindItemSalesBidHistory(exec boil.Executor, itemSaleID string, bidderID string, bidAt time.Time, selectCols ...string) (*ItemSalesBidHistory, error) {
	itemSalesBidHistoryObj := &ItemSalesBidHistory{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"item_sales_bid_history\" where \"item_sale_id\"=$1 AND \"bidder_id\"=$2 AND \"bid_at\"=$3", sel,
	)

	q := queries.Raw(query, itemSaleID, bidderID, bidAt)

	err := q.Bind(nil, exec, itemSalesBidHistoryObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from item_sales_bid_history")
	}

	if err = itemSalesBidHistoryObj.doAfterSelectHooks(exec); err != nil {
		return itemSalesBidHistoryObj, err
	}

	return itemSalesBidHistoryObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *ItemSalesBidHistory) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no item_sales_bid_history provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(itemSalesBidHistoryColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	itemSalesBidHistoryInsertCacheMut.RLock()
	cache, cached := itemSalesBidHistoryInsertCache[key]
	itemSalesBidHistoryInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			itemSalesBidHistoryAllColumns,
			itemSalesBidHistoryColumnsWithDefault,
			itemSalesBidHistoryColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(itemSalesBidHistoryType, itemSalesBidHistoryMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(itemSalesBidHistoryType, itemSalesBidHistoryMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"item_sales_bid_history\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"item_sales_bid_history\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into item_sales_bid_history")
	}

	if !cached {
		itemSalesBidHistoryInsertCacheMut.Lock()
		itemSalesBidHistoryInsertCache[key] = cache
		itemSalesBidHistoryInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the ItemSalesBidHistory.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *ItemSalesBidHistory) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	itemSalesBidHistoryUpdateCacheMut.RLock()
	cache, cached := itemSalesBidHistoryUpdateCache[key]
	itemSalesBidHistoryUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			itemSalesBidHistoryAllColumns,
			itemSalesBidHistoryPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update item_sales_bid_history, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"item_sales_bid_history\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, itemSalesBidHistoryPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(itemSalesBidHistoryType, itemSalesBidHistoryMapping, append(wl, itemSalesBidHistoryPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update item_sales_bid_history row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for item_sales_bid_history")
	}

	if !cached {
		itemSalesBidHistoryUpdateCacheMut.Lock()
		itemSalesBidHistoryUpdateCache[key] = cache
		itemSalesBidHistoryUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q itemSalesBidHistoryQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for item_sales_bid_history")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for item_sales_bid_history")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ItemSalesBidHistorySlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), itemSalesBidHistoryPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"item_sales_bid_history\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, itemSalesBidHistoryPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in itemSalesBidHistory slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all itemSalesBidHistory")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *ItemSalesBidHistory) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no item_sales_bid_history provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(itemSalesBidHistoryColumnsWithDefault, o)

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

	itemSalesBidHistoryUpsertCacheMut.RLock()
	cache, cached := itemSalesBidHistoryUpsertCache[key]
	itemSalesBidHistoryUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			itemSalesBidHistoryAllColumns,
			itemSalesBidHistoryColumnsWithDefault,
			itemSalesBidHistoryColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			itemSalesBidHistoryAllColumns,
			itemSalesBidHistoryPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert item_sales_bid_history, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(itemSalesBidHistoryPrimaryKeyColumns))
			copy(conflict, itemSalesBidHistoryPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"item_sales_bid_history\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(itemSalesBidHistoryType, itemSalesBidHistoryMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(itemSalesBidHistoryType, itemSalesBidHistoryMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert item_sales_bid_history")
	}

	if !cached {
		itemSalesBidHistoryUpsertCacheMut.Lock()
		itemSalesBidHistoryUpsertCache[key] = cache
		itemSalesBidHistoryUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single ItemSalesBidHistory record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *ItemSalesBidHistory) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no ItemSalesBidHistory provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), itemSalesBidHistoryPrimaryKeyMapping)
	sql := "DELETE FROM \"item_sales_bid_history\" WHERE \"item_sale_id\"=$1 AND \"bidder_id\"=$2 AND \"bid_at\"=$3"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from item_sales_bid_history")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for item_sales_bid_history")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q itemSalesBidHistoryQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no itemSalesBidHistoryQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from item_sales_bid_history")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for item_sales_bid_history")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ItemSalesBidHistorySlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(itemSalesBidHistoryBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), itemSalesBidHistoryPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"item_sales_bid_history\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, itemSalesBidHistoryPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from itemSalesBidHistory slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for item_sales_bid_history")
	}

	if len(itemSalesBidHistoryAfterDeleteHooks) != 0 {
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
func (o *ItemSalesBidHistory) Reload(exec boil.Executor) error {
	ret, err := FindItemSalesBidHistory(exec, o.ItemSaleID, o.BidderID, o.BidAt)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ItemSalesBidHistorySlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ItemSalesBidHistorySlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), itemSalesBidHistoryPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"item_sales_bid_history\".* FROM \"item_sales_bid_history\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, itemSalesBidHistoryPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in ItemSalesBidHistorySlice")
	}

	*o = slice

	return nil
}

// ItemSalesBidHistoryExists checks if the ItemSalesBidHistory row exists.
func ItemSalesBidHistoryExists(exec boil.Executor, itemSaleID string, bidderID string, bidAt time.Time) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"item_sales_bid_history\" where \"item_sale_id\"=$1 AND \"bidder_id\"=$2 AND \"bid_at\"=$3 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, itemSaleID, bidderID, bidAt)
	}
	row := exec.QueryRow(sql, itemSaleID, bidderID, bidAt)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if item_sales_bid_history exists")
	}

	return exists, nil
}
