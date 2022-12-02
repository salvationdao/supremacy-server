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

// RepairGameBlockLog is an object representing the database table.
type RepairGameBlockLog struct {
	ID                  string              `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	RepairAgentID       string              `boiler:"repair_agent_id" boil:"repair_agent_id" json:"repair_agent_id" toml:"repair_agent_id" yaml:"repair_agent_id"`
	RepairGameBlockType string              `boiler:"repair_game_block_type" boil:"repair_game_block_type" json:"repair_game_block_type" toml:"repair_game_block_type" yaml:"repair_game_block_type"`
	SpeedMultiplier     decimal.Decimal     `boiler:"speed_multiplier" boil:"speed_multiplier" json:"speed_multiplier" toml:"speed_multiplier" yaml:"speed_multiplier"`
	TriggerKey          string              `boiler:"trigger_key" boil:"trigger_key" json:"trigger_key" toml:"trigger_key" yaml:"trigger_key"`
	Width               decimal.Decimal     `boiler:"width" boil:"width" json:"width" toml:"width" yaml:"width"`
	Depth               decimal.Decimal     `boiler:"depth" boil:"depth" json:"depth" toml:"depth" yaml:"depth"`
	StackedAt           null.Time           `boiler:"stacked_at" boil:"stacked_at" json:"stacked_at,omitempty" toml:"stacked_at" yaml:"stacked_at,omitempty"`
	StackedWidth        decimal.NullDecimal `boiler:"stacked_width" boil:"stacked_width" json:"stacked_width,omitempty" toml:"stacked_width" yaml:"stacked_width,omitempty"`
	StackedDepth        decimal.NullDecimal `boiler:"stacked_depth" boil:"stacked_depth" json:"stacked_depth,omitempty" toml:"stacked_depth" yaml:"stacked_depth,omitempty"`
	IsFailed            bool                `boiler:"is_failed" boil:"is_failed" json:"is_failed" toml:"is_failed" yaml:"is_failed"`
	CreatedAt           time.Time           `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt           time.Time           `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	DeletedAt           null.Time           `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`

	R *repairGameBlockLogR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L repairGameBlockLogL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var RepairGameBlockLogColumns = struct {
	ID                  string
	RepairAgentID       string
	RepairGameBlockType string
	SpeedMultiplier     string
	TriggerKey          string
	Width               string
	Depth               string
	StackedAt           string
	StackedWidth        string
	StackedDepth        string
	IsFailed            string
	CreatedAt           string
	UpdatedAt           string
	DeletedAt           string
}{
	ID:                  "id",
	RepairAgentID:       "repair_agent_id",
	RepairGameBlockType: "repair_game_block_type",
	SpeedMultiplier:     "speed_multiplier",
	TriggerKey:          "trigger_key",
	Width:               "width",
	Depth:               "depth",
	StackedAt:           "stacked_at",
	StackedWidth:        "stacked_width",
	StackedDepth:        "stacked_depth",
	IsFailed:            "is_failed",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
	DeletedAt:           "deleted_at",
}

var RepairGameBlockLogTableColumns = struct {
	ID                  string
	RepairAgentID       string
	RepairGameBlockType string
	SpeedMultiplier     string
	TriggerKey          string
	Width               string
	Depth               string
	StackedAt           string
	StackedWidth        string
	StackedDepth        string
	IsFailed            string
	CreatedAt           string
	UpdatedAt           string
	DeletedAt           string
}{
	ID:                  "repair_game_block_logs.id",
	RepairAgentID:       "repair_game_block_logs.repair_agent_id",
	RepairGameBlockType: "repair_game_block_logs.repair_game_block_type",
	SpeedMultiplier:     "repair_game_block_logs.speed_multiplier",
	TriggerKey:          "repair_game_block_logs.trigger_key",
	Width:               "repair_game_block_logs.width",
	Depth:               "repair_game_block_logs.depth",
	StackedAt:           "repair_game_block_logs.stacked_at",
	StackedWidth:        "repair_game_block_logs.stacked_width",
	StackedDepth:        "repair_game_block_logs.stacked_depth",
	IsFailed:            "repair_game_block_logs.is_failed",
	CreatedAt:           "repair_game_block_logs.created_at",
	UpdatedAt:           "repair_game_block_logs.updated_at",
	DeletedAt:           "repair_game_block_logs.deleted_at",
}

// Generated where

var RepairGameBlockLogWhere = struct {
	ID                  whereHelperstring
	RepairAgentID       whereHelperstring
	RepairGameBlockType whereHelperstring
	SpeedMultiplier     whereHelperdecimal_Decimal
	TriggerKey          whereHelperstring
	Width               whereHelperdecimal_Decimal
	Depth               whereHelperdecimal_Decimal
	StackedAt           whereHelpernull_Time
	StackedWidth        whereHelperdecimal_NullDecimal
	StackedDepth        whereHelperdecimal_NullDecimal
	IsFailed            whereHelperbool
	CreatedAt           whereHelpertime_Time
	UpdatedAt           whereHelpertime_Time
	DeletedAt           whereHelpernull_Time
}{
	ID:                  whereHelperstring{field: "\"repair_game_block_logs\".\"id\""},
	RepairAgentID:       whereHelperstring{field: "\"repair_game_block_logs\".\"repair_agent_id\""},
	RepairGameBlockType: whereHelperstring{field: "\"repair_game_block_logs\".\"repair_game_block_type\""},
	SpeedMultiplier:     whereHelperdecimal_Decimal{field: "\"repair_game_block_logs\".\"speed_multiplier\""},
	TriggerKey:          whereHelperstring{field: "\"repair_game_block_logs\".\"trigger_key\""},
	Width:               whereHelperdecimal_Decimal{field: "\"repair_game_block_logs\".\"width\""},
	Depth:               whereHelperdecimal_Decimal{field: "\"repair_game_block_logs\".\"depth\""},
	StackedAt:           whereHelpernull_Time{field: "\"repair_game_block_logs\".\"stacked_at\""},
	StackedWidth:        whereHelperdecimal_NullDecimal{field: "\"repair_game_block_logs\".\"stacked_width\""},
	StackedDepth:        whereHelperdecimal_NullDecimal{field: "\"repair_game_block_logs\".\"stacked_depth\""},
	IsFailed:            whereHelperbool{field: "\"repair_game_block_logs\".\"is_failed\""},
	CreatedAt:           whereHelpertime_Time{field: "\"repair_game_block_logs\".\"created_at\""},
	UpdatedAt:           whereHelpertime_Time{field: "\"repair_game_block_logs\".\"updated_at\""},
	DeletedAt:           whereHelpernull_Time{field: "\"repair_game_block_logs\".\"deleted_at\""},
}

// RepairGameBlockLogRels is where relationship names are stored.
var RepairGameBlockLogRels = struct {
	RepairAgent string
}{
	RepairAgent: "RepairAgent",
}

// repairGameBlockLogR is where relationships are stored.
type repairGameBlockLogR struct {
	RepairAgent *RepairAgent `boiler:"RepairAgent" boil:"RepairAgent" json:"RepairAgent" toml:"RepairAgent" yaml:"RepairAgent"`
}

// NewStruct creates a new relationship struct
func (*repairGameBlockLogR) NewStruct() *repairGameBlockLogR {
	return &repairGameBlockLogR{}
}

// repairGameBlockLogL is where Load methods for each relationship are stored.
type repairGameBlockLogL struct{}

var (
	repairGameBlockLogAllColumns            = []string{"id", "repair_agent_id", "repair_game_block_type", "speed_multiplier", "trigger_key", "width", "depth", "stacked_at", "stacked_width", "stacked_depth", "is_failed", "created_at", "updated_at", "deleted_at"}
	repairGameBlockLogColumnsWithoutDefault = []string{"repair_agent_id", "trigger_key"}
	repairGameBlockLogColumnsWithDefault    = []string{"id", "repair_game_block_type", "speed_multiplier", "width", "depth", "stacked_at", "stacked_width", "stacked_depth", "is_failed", "created_at", "updated_at", "deleted_at"}
	repairGameBlockLogPrimaryKeyColumns     = []string{"id"}
	repairGameBlockLogGeneratedColumns      = []string{}
)

type (
	// RepairGameBlockLogSlice is an alias for a slice of pointers to RepairGameBlockLog.
	// This should almost always be used instead of []RepairGameBlockLog.
	RepairGameBlockLogSlice []*RepairGameBlockLog
	// RepairGameBlockLogHook is the signature for custom RepairGameBlockLog hook methods
	RepairGameBlockLogHook func(boil.Executor, *RepairGameBlockLog) error

	repairGameBlockLogQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	repairGameBlockLogType                 = reflect.TypeOf(&RepairGameBlockLog{})
	repairGameBlockLogMapping              = queries.MakeStructMapping(repairGameBlockLogType)
	repairGameBlockLogPrimaryKeyMapping, _ = queries.BindMapping(repairGameBlockLogType, repairGameBlockLogMapping, repairGameBlockLogPrimaryKeyColumns)
	repairGameBlockLogInsertCacheMut       sync.RWMutex
	repairGameBlockLogInsertCache          = make(map[string]insertCache)
	repairGameBlockLogUpdateCacheMut       sync.RWMutex
	repairGameBlockLogUpdateCache          = make(map[string]updateCache)
	repairGameBlockLogUpsertCacheMut       sync.RWMutex
	repairGameBlockLogUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var repairGameBlockLogAfterSelectHooks []RepairGameBlockLogHook

var repairGameBlockLogBeforeInsertHooks []RepairGameBlockLogHook
var repairGameBlockLogAfterInsertHooks []RepairGameBlockLogHook

var repairGameBlockLogBeforeUpdateHooks []RepairGameBlockLogHook
var repairGameBlockLogAfterUpdateHooks []RepairGameBlockLogHook

var repairGameBlockLogBeforeDeleteHooks []RepairGameBlockLogHook
var repairGameBlockLogAfterDeleteHooks []RepairGameBlockLogHook

var repairGameBlockLogBeforeUpsertHooks []RepairGameBlockLogHook
var repairGameBlockLogAfterUpsertHooks []RepairGameBlockLogHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *RepairGameBlockLog) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range repairGameBlockLogAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *RepairGameBlockLog) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairGameBlockLogBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *RepairGameBlockLog) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairGameBlockLogAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *RepairGameBlockLog) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range repairGameBlockLogBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *RepairGameBlockLog) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range repairGameBlockLogAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *RepairGameBlockLog) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range repairGameBlockLogBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *RepairGameBlockLog) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range repairGameBlockLogAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *RepairGameBlockLog) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairGameBlockLogBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *RepairGameBlockLog) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairGameBlockLogAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddRepairGameBlockLogHook registers your hook function for all future operations.
func AddRepairGameBlockLogHook(hookPoint boil.HookPoint, repairGameBlockLogHook RepairGameBlockLogHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		repairGameBlockLogAfterSelectHooks = append(repairGameBlockLogAfterSelectHooks, repairGameBlockLogHook)
	case boil.BeforeInsertHook:
		repairGameBlockLogBeforeInsertHooks = append(repairGameBlockLogBeforeInsertHooks, repairGameBlockLogHook)
	case boil.AfterInsertHook:
		repairGameBlockLogAfterInsertHooks = append(repairGameBlockLogAfterInsertHooks, repairGameBlockLogHook)
	case boil.BeforeUpdateHook:
		repairGameBlockLogBeforeUpdateHooks = append(repairGameBlockLogBeforeUpdateHooks, repairGameBlockLogHook)
	case boil.AfterUpdateHook:
		repairGameBlockLogAfterUpdateHooks = append(repairGameBlockLogAfterUpdateHooks, repairGameBlockLogHook)
	case boil.BeforeDeleteHook:
		repairGameBlockLogBeforeDeleteHooks = append(repairGameBlockLogBeforeDeleteHooks, repairGameBlockLogHook)
	case boil.AfterDeleteHook:
		repairGameBlockLogAfterDeleteHooks = append(repairGameBlockLogAfterDeleteHooks, repairGameBlockLogHook)
	case boil.BeforeUpsertHook:
		repairGameBlockLogBeforeUpsertHooks = append(repairGameBlockLogBeforeUpsertHooks, repairGameBlockLogHook)
	case boil.AfterUpsertHook:
		repairGameBlockLogAfterUpsertHooks = append(repairGameBlockLogAfterUpsertHooks, repairGameBlockLogHook)
	}
}

// One returns a single repairGameBlockLog record from the query.
func (q repairGameBlockLogQuery) One(exec boil.Executor) (*RepairGameBlockLog, error) {
	o := &RepairGameBlockLog{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for repair_game_block_logs")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all RepairGameBlockLog records from the query.
func (q repairGameBlockLogQuery) All(exec boil.Executor) (RepairGameBlockLogSlice, error) {
	var o []*RepairGameBlockLog

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to RepairGameBlockLog slice")
	}

	if len(repairGameBlockLogAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all RepairGameBlockLog records in the query.
func (q repairGameBlockLogQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count repair_game_block_logs rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q repairGameBlockLogQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if repair_game_block_logs exists")
	}

	return count > 0, nil
}

// RepairAgent pointed to by the foreign key.
func (o *RepairGameBlockLog) RepairAgent(mods ...qm.QueryMod) repairAgentQuery {
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
func (repairGameBlockLogL) LoadRepairAgent(e boil.Executor, singular bool, maybeRepairGameBlockLog interface{}, mods queries.Applicator) error {
	var slice []*RepairGameBlockLog
	var object *RepairGameBlockLog

	if singular {
		object = maybeRepairGameBlockLog.(*RepairGameBlockLog)
	} else {
		slice = *maybeRepairGameBlockLog.(*[]*RepairGameBlockLog)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &repairGameBlockLogR{}
		}
		args = append(args, object.RepairAgentID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &repairGameBlockLogR{}
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

	if len(repairGameBlockLogAfterSelectHooks) != 0 {
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
		foreign.R.RepairGameBlockLogs = append(foreign.R.RepairGameBlockLogs, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.RepairAgentID == foreign.ID {
				local.R.RepairAgent = foreign
				if foreign.R == nil {
					foreign.R = &repairAgentR{}
				}
				foreign.R.RepairGameBlockLogs = append(foreign.R.RepairGameBlockLogs, local)
				break
			}
		}
	}

	return nil
}

// SetRepairAgent of the repairGameBlockLog to the related item.
// Sets o.R.RepairAgent to related.
// Adds o to related.R.RepairGameBlockLogs.
func (o *RepairGameBlockLog) SetRepairAgent(exec boil.Executor, insert bool, related *RepairAgent) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"repair_game_block_logs\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"repair_agent_id"}),
		strmangle.WhereClause("\"", "\"", 2, repairGameBlockLogPrimaryKeyColumns),
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
		o.R = &repairGameBlockLogR{
			RepairAgent: related,
		}
	} else {
		o.R.RepairAgent = related
	}

	if related.R == nil {
		related.R = &repairAgentR{
			RepairGameBlockLogs: RepairGameBlockLogSlice{o},
		}
	} else {
		related.R.RepairGameBlockLogs = append(related.R.RepairGameBlockLogs, o)
	}

	return nil
}

// RepairGameBlockLogs retrieves all the records using an executor.
func RepairGameBlockLogs(mods ...qm.QueryMod) repairGameBlockLogQuery {
	mods = append(mods, qm.From("\"repair_game_block_logs\""), qmhelper.WhereIsNull("\"repair_game_block_logs\".\"deleted_at\""))
	return repairGameBlockLogQuery{NewQuery(mods...)}
}

// FindRepairGameBlockLog retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindRepairGameBlockLog(exec boil.Executor, iD string, selectCols ...string) (*RepairGameBlockLog, error) {
	repairGameBlockLogObj := &RepairGameBlockLog{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"repair_game_block_logs\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, repairGameBlockLogObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from repair_game_block_logs")
	}

	if err = repairGameBlockLogObj.doAfterSelectHooks(exec); err != nil {
		return repairGameBlockLogObj, err
	}

	return repairGameBlockLogObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *RepairGameBlockLog) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no repair_game_block_logs provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}
	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(repairGameBlockLogColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	repairGameBlockLogInsertCacheMut.RLock()
	cache, cached := repairGameBlockLogInsertCache[key]
	repairGameBlockLogInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			repairGameBlockLogAllColumns,
			repairGameBlockLogColumnsWithDefault,
			repairGameBlockLogColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(repairGameBlockLogType, repairGameBlockLogMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(repairGameBlockLogType, repairGameBlockLogMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"repair_game_block_logs\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"repair_game_block_logs\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into repair_game_block_logs")
	}

	if !cached {
		repairGameBlockLogInsertCacheMut.Lock()
		repairGameBlockLogInsertCache[key] = cache
		repairGameBlockLogInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the RepairGameBlockLog.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *RepairGameBlockLog) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	repairGameBlockLogUpdateCacheMut.RLock()
	cache, cached := repairGameBlockLogUpdateCache[key]
	repairGameBlockLogUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			repairGameBlockLogAllColumns,
			repairGameBlockLogPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update repair_game_block_logs, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"repair_game_block_logs\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, repairGameBlockLogPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(repairGameBlockLogType, repairGameBlockLogMapping, append(wl, repairGameBlockLogPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update repair_game_block_logs row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for repair_game_block_logs")
	}

	if !cached {
		repairGameBlockLogUpdateCacheMut.Lock()
		repairGameBlockLogUpdateCache[key] = cache
		repairGameBlockLogUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q repairGameBlockLogQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for repair_game_block_logs")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for repair_game_block_logs")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o RepairGameBlockLogSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), repairGameBlockLogPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"repair_game_block_logs\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, repairGameBlockLogPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in repairGameBlockLog slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all repairGameBlockLog")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *RepairGameBlockLog) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no repair_game_block_logs provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}
	o.UpdatedAt = currTime

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(repairGameBlockLogColumnsWithDefault, o)

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

	repairGameBlockLogUpsertCacheMut.RLock()
	cache, cached := repairGameBlockLogUpsertCache[key]
	repairGameBlockLogUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			repairGameBlockLogAllColumns,
			repairGameBlockLogColumnsWithDefault,
			repairGameBlockLogColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			repairGameBlockLogAllColumns,
			repairGameBlockLogPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert repair_game_block_logs, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(repairGameBlockLogPrimaryKeyColumns))
			copy(conflict, repairGameBlockLogPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"repair_game_block_logs\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(repairGameBlockLogType, repairGameBlockLogMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(repairGameBlockLogType, repairGameBlockLogMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert repair_game_block_logs")
	}

	if !cached {
		repairGameBlockLogUpsertCacheMut.Lock()
		repairGameBlockLogUpsertCache[key] = cache
		repairGameBlockLogUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single RepairGameBlockLog record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *RepairGameBlockLog) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no RepairGameBlockLog provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), repairGameBlockLogPrimaryKeyMapping)
		sql = "DELETE FROM \"repair_game_block_logs\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"repair_game_block_logs\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(repairGameBlockLogType, repairGameBlockLogMapping, append(wl, repairGameBlockLogPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from repair_game_block_logs")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for repair_game_block_logs")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q repairGameBlockLogQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no repairGameBlockLogQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from repair_game_block_logs")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for repair_game_block_logs")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o RepairGameBlockLogSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(repairGameBlockLogBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), repairGameBlockLogPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"repair_game_block_logs\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, repairGameBlockLogPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), repairGameBlockLogPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"repair_game_block_logs\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, repairGameBlockLogPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from repairGameBlockLog slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for repair_game_block_logs")
	}

	if len(repairGameBlockLogAfterDeleteHooks) != 0 {
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
func (o *RepairGameBlockLog) Reload(exec boil.Executor) error {
	ret, err := FindRepairGameBlockLog(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *RepairGameBlockLogSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := RepairGameBlockLogSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), repairGameBlockLogPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"repair_game_block_logs\".* FROM \"repair_game_block_logs\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, repairGameBlockLogPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in RepairGameBlockLogSlice")
	}

	*o = slice

	return nil
}

// RepairGameBlockLogExists checks if the RepairGameBlockLog row exists.
func RepairGameBlockLogExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"repair_game_block_logs\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if repair_game_block_logs exists")
	}

	return exists, nil
}
