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

// RepairAgentLogsOld is an object representing the database table.
type RepairAgentLogsOld struct {
	ID            string          `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	RepairAgentID string          `boiler:"repair_agent_id" boil:"repair_agent_id" json:"repair_agent_id" toml:"repair_agent_id" yaml:"repair_agent_id"`
	TriggeredWith string          `boiler:"triggered_with" boil:"triggered_with" json:"triggered_with" toml:"triggered_with" yaml:"triggered_with"`
	Score         int             `boiler:"score" boil:"score" json:"score" toml:"score" yaml:"score"`
	BlockWidth    decimal.Decimal `boiler:"block_width" boil:"block_width" json:"block_width" toml:"block_width" yaml:"block_width"`
	BlockDepth    decimal.Decimal `boiler:"block_depth" boil:"block_depth" json:"block_depth" toml:"block_depth" yaml:"block_depth"`
	IsFailed      bool            `boiler:"is_failed" boil:"is_failed" json:"is_failed" toml:"is_failed" yaml:"is_failed"`
	CreatedAt     time.Time       `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *repairAgentLogsOldR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L repairAgentLogsOldL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var RepairAgentLogsOldColumns = struct {
	ID            string
	RepairAgentID string
	TriggeredWith string
	Score         string
	BlockWidth    string
	BlockDepth    string
	IsFailed      string
	CreatedAt     string
}{
	ID:            "id",
	RepairAgentID: "repair_agent_id",
	TriggeredWith: "triggered_with",
	Score:         "score",
	BlockWidth:    "block_width",
	BlockDepth:    "block_depth",
	IsFailed:      "is_failed",
	CreatedAt:     "created_at",
}

var RepairAgentLogsOldTableColumns = struct {
	ID            string
	RepairAgentID string
	TriggeredWith string
	Score         string
	BlockWidth    string
	BlockDepth    string
	IsFailed      string
	CreatedAt     string
}{
	ID:            "repair_agent_logs_old.id",
	RepairAgentID: "repair_agent_logs_old.repair_agent_id",
	TriggeredWith: "repair_agent_logs_old.triggered_with",
	Score:         "repair_agent_logs_old.score",
	BlockWidth:    "repair_agent_logs_old.block_width",
	BlockDepth:    "repair_agent_logs_old.block_depth",
	IsFailed:      "repair_agent_logs_old.is_failed",
	CreatedAt:     "repair_agent_logs_old.created_at",
}

// Generated where

var RepairAgentLogsOldWhere = struct {
	ID            whereHelperstring
	RepairAgentID whereHelperstring
	TriggeredWith whereHelperstring
	Score         whereHelperint
	BlockWidth    whereHelperdecimal_Decimal
	BlockDepth    whereHelperdecimal_Decimal
	IsFailed      whereHelperbool
	CreatedAt     whereHelpertime_Time
}{
	ID:            whereHelperstring{field: "\"repair_agent_logs_old\".\"id\""},
	RepairAgentID: whereHelperstring{field: "\"repair_agent_logs_old\".\"repair_agent_id\""},
	TriggeredWith: whereHelperstring{field: "\"repair_agent_logs_old\".\"triggered_with\""},
	Score:         whereHelperint{field: "\"repair_agent_logs_old\".\"score\""},
	BlockWidth:    whereHelperdecimal_Decimal{field: "\"repair_agent_logs_old\".\"block_width\""},
	BlockDepth:    whereHelperdecimal_Decimal{field: "\"repair_agent_logs_old\".\"block_depth\""},
	IsFailed:      whereHelperbool{field: "\"repair_agent_logs_old\".\"is_failed\""},
	CreatedAt:     whereHelpertime_Time{field: "\"repair_agent_logs_old\".\"created_at\""},
}

// RepairAgentLogsOldRels is where relationship names are stored.
var RepairAgentLogsOldRels = struct {
	RepairAgent string
}{
	RepairAgent: "RepairAgent",
}

// repairAgentLogsOldR is where relationships are stored.
type repairAgentLogsOldR struct {
	RepairAgent *RepairAgent `boiler:"RepairAgent" boil:"RepairAgent" json:"RepairAgent" toml:"RepairAgent" yaml:"RepairAgent"`
}

// NewStruct creates a new relationship struct
func (*repairAgentLogsOldR) NewStruct() *repairAgentLogsOldR {
	return &repairAgentLogsOldR{}
}

// repairAgentLogsOldL is where Load methods for each relationship are stored.
type repairAgentLogsOldL struct{}

var (
	repairAgentLogsOldAllColumns            = []string{"id", "repair_agent_id", "triggered_with", "score", "block_width", "block_depth", "is_failed", "created_at"}
	repairAgentLogsOldColumnsWithoutDefault = []string{"repair_agent_id", "triggered_with", "score", "block_width", "block_depth"}
	repairAgentLogsOldColumnsWithDefault    = []string{"id", "is_failed", "created_at"}
	repairAgentLogsOldPrimaryKeyColumns     = []string{"id"}
	repairAgentLogsOldGeneratedColumns      = []string{}
)

type (
	// RepairAgentLogsOldSlice is an alias for a slice of pointers to RepairAgentLogsOld.
	// This should almost always be used instead of []RepairAgentLogsOld.
	RepairAgentLogsOldSlice []*RepairAgentLogsOld
	// RepairAgentLogsOldHook is the signature for custom RepairAgentLogsOld hook methods
	RepairAgentLogsOldHook func(boil.Executor, *RepairAgentLogsOld) error

	repairAgentLogsOldQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	repairAgentLogsOldType                 = reflect.TypeOf(&RepairAgentLogsOld{})
	repairAgentLogsOldMapping              = queries.MakeStructMapping(repairAgentLogsOldType)
	repairAgentLogsOldPrimaryKeyMapping, _ = queries.BindMapping(repairAgentLogsOldType, repairAgentLogsOldMapping, repairAgentLogsOldPrimaryKeyColumns)
	repairAgentLogsOldInsertCacheMut       sync.RWMutex
	repairAgentLogsOldInsertCache          = make(map[string]insertCache)
	repairAgentLogsOldUpdateCacheMut       sync.RWMutex
	repairAgentLogsOldUpdateCache          = make(map[string]updateCache)
	repairAgentLogsOldUpsertCacheMut       sync.RWMutex
	repairAgentLogsOldUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var repairAgentLogsOldAfterSelectHooks []RepairAgentLogsOldHook

var repairAgentLogsOldBeforeInsertHooks []RepairAgentLogsOldHook
var repairAgentLogsOldAfterInsertHooks []RepairAgentLogsOldHook

var repairAgentLogsOldBeforeUpdateHooks []RepairAgentLogsOldHook
var repairAgentLogsOldAfterUpdateHooks []RepairAgentLogsOldHook

var repairAgentLogsOldBeforeDeleteHooks []RepairAgentLogsOldHook
var repairAgentLogsOldAfterDeleteHooks []RepairAgentLogsOldHook

var repairAgentLogsOldBeforeUpsertHooks []RepairAgentLogsOldHook
var repairAgentLogsOldAfterUpsertHooks []RepairAgentLogsOldHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *RepairAgentLogsOld) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range repairAgentLogsOldAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *RepairAgentLogsOld) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairAgentLogsOldBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *RepairAgentLogsOld) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairAgentLogsOldAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *RepairAgentLogsOld) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range repairAgentLogsOldBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *RepairAgentLogsOld) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range repairAgentLogsOldAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *RepairAgentLogsOld) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range repairAgentLogsOldBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *RepairAgentLogsOld) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range repairAgentLogsOldAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *RepairAgentLogsOld) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairAgentLogsOldBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *RepairAgentLogsOld) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairAgentLogsOldAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddRepairAgentLogsOldHook registers your hook function for all future operations.
func AddRepairAgentLogsOldHook(hookPoint boil.HookPoint, repairAgentLogsOldHook RepairAgentLogsOldHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		repairAgentLogsOldAfterSelectHooks = append(repairAgentLogsOldAfterSelectHooks, repairAgentLogsOldHook)
	case boil.BeforeInsertHook:
		repairAgentLogsOldBeforeInsertHooks = append(repairAgentLogsOldBeforeInsertHooks, repairAgentLogsOldHook)
	case boil.AfterInsertHook:
		repairAgentLogsOldAfterInsertHooks = append(repairAgentLogsOldAfterInsertHooks, repairAgentLogsOldHook)
	case boil.BeforeUpdateHook:
		repairAgentLogsOldBeforeUpdateHooks = append(repairAgentLogsOldBeforeUpdateHooks, repairAgentLogsOldHook)
	case boil.AfterUpdateHook:
		repairAgentLogsOldAfterUpdateHooks = append(repairAgentLogsOldAfterUpdateHooks, repairAgentLogsOldHook)
	case boil.BeforeDeleteHook:
		repairAgentLogsOldBeforeDeleteHooks = append(repairAgentLogsOldBeforeDeleteHooks, repairAgentLogsOldHook)
	case boil.AfterDeleteHook:
		repairAgentLogsOldAfterDeleteHooks = append(repairAgentLogsOldAfterDeleteHooks, repairAgentLogsOldHook)
	case boil.BeforeUpsertHook:
		repairAgentLogsOldBeforeUpsertHooks = append(repairAgentLogsOldBeforeUpsertHooks, repairAgentLogsOldHook)
	case boil.AfterUpsertHook:
		repairAgentLogsOldAfterUpsertHooks = append(repairAgentLogsOldAfterUpsertHooks, repairAgentLogsOldHook)
	}
}

// One returns a single repairAgentLogsOld record from the query.
func (q repairAgentLogsOldQuery) One(exec boil.Executor) (*RepairAgentLogsOld, error) {
	o := &RepairAgentLogsOld{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for repair_agent_logs_old")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all RepairAgentLogsOld records from the query.
func (q repairAgentLogsOldQuery) All(exec boil.Executor) (RepairAgentLogsOldSlice, error) {
	var o []*RepairAgentLogsOld

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to RepairAgentLogsOld slice")
	}

	if len(repairAgentLogsOldAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all RepairAgentLogsOld records in the query.
func (q repairAgentLogsOldQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count repair_agent_logs_old rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q repairAgentLogsOldQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if repair_agent_logs_old exists")
	}

	return count > 0, nil
}

// RepairAgent pointed to by the foreign key.
func (o *RepairAgentLogsOld) RepairAgent(mods ...qm.QueryMod) repairAgentQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.RepairAgentID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := RepairAgents(queryMods...)
	queries.SetFrom(query.Query, "\"repair_agents\"")

	return query
}

// LoadRepairAgent allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (repairAgentLogsOldL) LoadRepairAgent(e boil.Executor, singular bool, maybeRepairAgentLogsOld interface{}, mods queries.Applicator) error {
	var slice []*RepairAgentLogsOld
	var object *RepairAgentLogsOld

	if singular {
		object = maybeRepairAgentLogsOld.(*RepairAgentLogsOld)
	} else {
		slice = *maybeRepairAgentLogsOld.(*[]*RepairAgentLogsOld)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &repairAgentLogsOldR{}
		}
		args = append(args, object.RepairAgentID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &repairAgentLogsOldR{}
			}

			for _, a := range args {
				if a == obj.RepairAgentID {
					continue Outer
				}
			}

			args = append(args, obj.RepairAgentID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`repair_agents`),
		qm.WhereIn(`repair_agents.id in ?`, args...),
		qmhelper.WhereIsNull(`repair_agents.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load RepairAgent")
	}

	var resultSlice []*RepairAgent
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice RepairAgent")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for repair_agents")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for repair_agents")
	}

	if len(repairAgentLogsOldAfterSelectHooks) != 0 {
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
		object.R.RepairAgent = foreign
		if foreign.R == nil {
			foreign.R = &repairAgentR{}
		}
		foreign.R.RepairAgentLogsOlds = append(foreign.R.RepairAgentLogsOlds, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.RepairAgentID == foreign.ID {
				local.R.RepairAgent = foreign
				if foreign.R == nil {
					foreign.R = &repairAgentR{}
				}
				foreign.R.RepairAgentLogsOlds = append(foreign.R.RepairAgentLogsOlds, local)
				break
			}
		}
	}

	return nil
}

// SetRepairAgent of the repairAgentLogsOld to the related item.
// Sets o.R.RepairAgent to related.
// Adds o to related.R.RepairAgentLogsOlds.
func (o *RepairAgentLogsOld) SetRepairAgent(exec boil.Executor, insert bool, related *RepairAgent) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"repair_agent_logs_old\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"repair_agent_id"}),
		strmangle.WhereClause("\"", "\"", 2, repairAgentLogsOldPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.RepairAgentID = related.ID
	if o.R == nil {
		o.R = &repairAgentLogsOldR{
			RepairAgent: related,
		}
	} else {
		o.R.RepairAgent = related
	}

	if related.R == nil {
		related.R = &repairAgentR{
			RepairAgentLogsOlds: RepairAgentLogsOldSlice{o},
		}
	} else {
		related.R.RepairAgentLogsOlds = append(related.R.RepairAgentLogsOlds, o)
	}

	return nil
}

// RepairAgentLogsOlds retrieves all the records using an executor.
func RepairAgentLogsOlds(mods ...qm.QueryMod) repairAgentLogsOldQuery {
	mods = append(mods, qm.From("\"repair_agent_logs_old\""))
	return repairAgentLogsOldQuery{NewQuery(mods...)}
}

// FindRepairAgentLogsOld retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindRepairAgentLogsOld(exec boil.Executor, iD string, selectCols ...string) (*RepairAgentLogsOld, error) {
	repairAgentLogsOldObj := &RepairAgentLogsOld{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"repair_agent_logs_old\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, repairAgentLogsOldObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from repair_agent_logs_old")
	}

	if err = repairAgentLogsOldObj.doAfterSelectHooks(exec); err != nil {
		return repairAgentLogsOldObj, err
	}

	return repairAgentLogsOldObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *RepairAgentLogsOld) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no repair_agent_logs_old provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(repairAgentLogsOldColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	repairAgentLogsOldInsertCacheMut.RLock()
	cache, cached := repairAgentLogsOldInsertCache[key]
	repairAgentLogsOldInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			repairAgentLogsOldAllColumns,
			repairAgentLogsOldColumnsWithDefault,
			repairAgentLogsOldColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(repairAgentLogsOldType, repairAgentLogsOldMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(repairAgentLogsOldType, repairAgentLogsOldMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"repair_agent_logs_old\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"repair_agent_logs_old\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into repair_agent_logs_old")
	}

	if !cached {
		repairAgentLogsOldInsertCacheMut.Lock()
		repairAgentLogsOldInsertCache[key] = cache
		repairAgentLogsOldInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the RepairAgentLogsOld.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *RepairAgentLogsOld) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	repairAgentLogsOldUpdateCacheMut.RLock()
	cache, cached := repairAgentLogsOldUpdateCache[key]
	repairAgentLogsOldUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			repairAgentLogsOldAllColumns,
			repairAgentLogsOldPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update repair_agent_logs_old, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"repair_agent_logs_old\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, repairAgentLogsOldPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(repairAgentLogsOldType, repairAgentLogsOldMapping, append(wl, repairAgentLogsOldPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update repair_agent_logs_old row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for repair_agent_logs_old")
	}

	if !cached {
		repairAgentLogsOldUpdateCacheMut.Lock()
		repairAgentLogsOldUpdateCache[key] = cache
		repairAgentLogsOldUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q repairAgentLogsOldQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for repair_agent_logs_old")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for repair_agent_logs_old")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o RepairAgentLogsOldSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), repairAgentLogsOldPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"repair_agent_logs_old\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, repairAgentLogsOldPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in repairAgentLogsOld slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all repairAgentLogsOld")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *RepairAgentLogsOld) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no repair_agent_logs_old provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(repairAgentLogsOldColumnsWithDefault, o)

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

	repairAgentLogsOldUpsertCacheMut.RLock()
	cache, cached := repairAgentLogsOldUpsertCache[key]
	repairAgentLogsOldUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			repairAgentLogsOldAllColumns,
			repairAgentLogsOldColumnsWithDefault,
			repairAgentLogsOldColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			repairAgentLogsOldAllColumns,
			repairAgentLogsOldPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert repair_agent_logs_old, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(repairAgentLogsOldPrimaryKeyColumns))
			copy(conflict, repairAgentLogsOldPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"repair_agent_logs_old\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(repairAgentLogsOldType, repairAgentLogsOldMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(repairAgentLogsOldType, repairAgentLogsOldMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert repair_agent_logs_old")
	}

	if !cached {
		repairAgentLogsOldUpsertCacheMut.Lock()
		repairAgentLogsOldUpsertCache[key] = cache
		repairAgentLogsOldUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single RepairAgentLogsOld record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *RepairAgentLogsOld) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no RepairAgentLogsOld provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), repairAgentLogsOldPrimaryKeyMapping)
	sql := "DELETE FROM \"repair_agent_logs_old\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from repair_agent_logs_old")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for repair_agent_logs_old")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q repairAgentLogsOldQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no repairAgentLogsOldQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from repair_agent_logs_old")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for repair_agent_logs_old")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o RepairAgentLogsOldSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(repairAgentLogsOldBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), repairAgentLogsOldPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"repair_agent_logs_old\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, repairAgentLogsOldPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from repairAgentLogsOld slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for repair_agent_logs_old")
	}

	if len(repairAgentLogsOldAfterDeleteHooks) != 0 {
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
func (o *RepairAgentLogsOld) Reload(exec boil.Executor) error {
	ret, err := FindRepairAgentLogsOld(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *RepairAgentLogsOldSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := RepairAgentLogsOldSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), repairAgentLogsOldPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"repair_agent_logs_old\".* FROM \"repair_agent_logs_old\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, repairAgentLogsOldPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in RepairAgentLogsOldSlice")
	}

	*o = slice

	return nil
}

// RepairAgentLogsOldExists checks if the RepairAgentLogsOld row exists.
func RepairAgentLogsOldExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"repair_agent_logs_old\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if repair_agent_logs_old exists")
	}

	return exists, nil
}
