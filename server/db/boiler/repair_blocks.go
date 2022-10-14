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

// RepairBlock is an object representing the database table.
type RepairBlock struct {
	ID            string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	RepairCaseID  string    `boiler:"repair_case_id" boil:"repair_case_id" json:"repair_case_id" toml:"repair_case_id" yaml:"repair_case_id"`
	RepairOfferID string    `boiler:"repair_offer_id" boil:"repair_offer_id" json:"repair_offer_id" toml:"repair_offer_id" yaml:"repair_offer_id"`
	RepairAgentID string    `boiler:"repair_agent_id" boil:"repair_agent_id" json:"repair_agent_id" toml:"repair_agent_id" yaml:"repair_agent_id"`
	CreatedAt     time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`

	R *repairBlockR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L repairBlockL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var RepairBlockColumns = struct {
	ID            string
	RepairCaseID  string
	RepairOfferID string
	RepairAgentID string
	CreatedAt     string
	UpdatedAt     string
}{
	ID:            "id",
	RepairCaseID:  "repair_case_id",
	RepairOfferID: "repair_offer_id",
	RepairAgentID: "repair_agent_id",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

var RepairBlockTableColumns = struct {
	ID            string
	RepairCaseID  string
	RepairOfferID string
	RepairAgentID string
	CreatedAt     string
	UpdatedAt     string
}{
	ID:            "repair_blocks.id",
	RepairCaseID:  "repair_blocks.repair_case_id",
	RepairOfferID: "repair_blocks.repair_offer_id",
	RepairAgentID: "repair_blocks.repair_agent_id",
	CreatedAt:     "repair_blocks.created_at",
	UpdatedAt:     "repair_blocks.updated_at",
}

// Generated where

var RepairBlockWhere = struct {
	ID            whereHelperstring
	RepairCaseID  whereHelperstring
	RepairOfferID whereHelperstring
	RepairAgentID whereHelperstring
	CreatedAt     whereHelpertime_Time
	UpdatedAt     whereHelpertime_Time
}{
	ID:            whereHelperstring{field: "\"repair_blocks\".\"id\""},
	RepairCaseID:  whereHelperstring{field: "\"repair_blocks\".\"repair_case_id\""},
	RepairOfferID: whereHelperstring{field: "\"repair_blocks\".\"repair_offer_id\""},
	RepairAgentID: whereHelperstring{field: "\"repair_blocks\".\"repair_agent_id\""},
	CreatedAt:     whereHelpertime_Time{field: "\"repair_blocks\".\"created_at\""},
	UpdatedAt:     whereHelpertime_Time{field: "\"repair_blocks\".\"updated_at\""},
}

// RepairBlockRels is where relationship names are stored.
var RepairBlockRels = struct {
	RepairAgent string
	RepairCase  string
	RepairOffer string
}{
	RepairAgent: "RepairAgent",
	RepairCase:  "RepairCase",
	RepairOffer: "RepairOffer",
}

// repairBlockR is where relationships are stored.
type repairBlockR struct {
	RepairAgent *RepairAgent `boiler:"RepairAgent" boil:"RepairAgent" json:"RepairAgent" toml:"RepairAgent" yaml:"RepairAgent"`
	RepairCase  *RepairCase  `boiler:"RepairCase" boil:"RepairCase" json:"RepairCase" toml:"RepairCase" yaml:"RepairCase"`
	RepairOffer *RepairOffer `boiler:"RepairOffer" boil:"RepairOffer" json:"RepairOffer" toml:"RepairOffer" yaml:"RepairOffer"`
}

// NewStruct creates a new relationship struct
func (*repairBlockR) NewStruct() *repairBlockR {
	return &repairBlockR{}
}

// repairBlockL is where Load methods for each relationship are stored.
type repairBlockL struct{}

var (
	repairBlockAllColumns            = []string{"id", "repair_case_id", "repair_offer_id", "repair_agent_id", "created_at", "updated_at"}
	repairBlockColumnsWithoutDefault = []string{"repair_case_id", "repair_offer_id", "repair_agent_id"}
	repairBlockColumnsWithDefault    = []string{"id", "created_at", "updated_at"}
	repairBlockPrimaryKeyColumns     = []string{"id"}
	repairBlockGeneratedColumns      = []string{}
)

type (
	// RepairBlockSlice is an alias for a slice of pointers to RepairBlock.
	// This should almost always be used instead of []RepairBlock.
	RepairBlockSlice []*RepairBlock
	// RepairBlockHook is the signature for custom RepairBlock hook methods
	RepairBlockHook func(boil.Executor, *RepairBlock) error

	repairBlockQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	repairBlockType                 = reflect.TypeOf(&RepairBlock{})
	repairBlockMapping              = queries.MakeStructMapping(repairBlockType)
	repairBlockPrimaryKeyMapping, _ = queries.BindMapping(repairBlockType, repairBlockMapping, repairBlockPrimaryKeyColumns)
	repairBlockInsertCacheMut       sync.RWMutex
	repairBlockInsertCache          = make(map[string]insertCache)
	repairBlockUpdateCacheMut       sync.RWMutex
	repairBlockUpdateCache          = make(map[string]updateCache)
	repairBlockUpsertCacheMut       sync.RWMutex
	repairBlockUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var repairBlockAfterSelectHooks []RepairBlockHook

var repairBlockBeforeInsertHooks []RepairBlockHook
var repairBlockAfterInsertHooks []RepairBlockHook

var repairBlockBeforeUpdateHooks []RepairBlockHook
var repairBlockAfterUpdateHooks []RepairBlockHook

var repairBlockBeforeDeleteHooks []RepairBlockHook
var repairBlockAfterDeleteHooks []RepairBlockHook

var repairBlockBeforeUpsertHooks []RepairBlockHook
var repairBlockAfterUpsertHooks []RepairBlockHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *RepairBlock) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range repairBlockAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *RepairBlock) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairBlockBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *RepairBlock) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairBlockAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *RepairBlock) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range repairBlockBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *RepairBlock) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range repairBlockAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *RepairBlock) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range repairBlockBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *RepairBlock) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range repairBlockAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *RepairBlock) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairBlockBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *RepairBlock) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range repairBlockAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddRepairBlockHook registers your hook function for all future operations.
func AddRepairBlockHook(hookPoint boil.HookPoint, repairBlockHook RepairBlockHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		repairBlockAfterSelectHooks = append(repairBlockAfterSelectHooks, repairBlockHook)
	case boil.BeforeInsertHook:
		repairBlockBeforeInsertHooks = append(repairBlockBeforeInsertHooks, repairBlockHook)
	case boil.AfterInsertHook:
		repairBlockAfterInsertHooks = append(repairBlockAfterInsertHooks, repairBlockHook)
	case boil.BeforeUpdateHook:
		repairBlockBeforeUpdateHooks = append(repairBlockBeforeUpdateHooks, repairBlockHook)
	case boil.AfterUpdateHook:
		repairBlockAfterUpdateHooks = append(repairBlockAfterUpdateHooks, repairBlockHook)
	case boil.BeforeDeleteHook:
		repairBlockBeforeDeleteHooks = append(repairBlockBeforeDeleteHooks, repairBlockHook)
	case boil.AfterDeleteHook:
		repairBlockAfterDeleteHooks = append(repairBlockAfterDeleteHooks, repairBlockHook)
	case boil.BeforeUpsertHook:
		repairBlockBeforeUpsertHooks = append(repairBlockBeforeUpsertHooks, repairBlockHook)
	case boil.AfterUpsertHook:
		repairBlockAfterUpsertHooks = append(repairBlockAfterUpsertHooks, repairBlockHook)
	}
}

// One returns a single repairBlock record from the query.
func (q repairBlockQuery) One(exec boil.Executor) (*RepairBlock, error) {
	o := &RepairBlock{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for repair_blocks")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all RepairBlock records from the query.
func (q repairBlockQuery) All(exec boil.Executor) (RepairBlockSlice, error) {
	var o []*RepairBlock

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to RepairBlock slice")
	}

	if len(repairBlockAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all RepairBlock records in the query.
func (q repairBlockQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count repair_blocks rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q repairBlockQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if repair_blocks exists")
	}

	return count > 0, nil
}

// RepairAgent pointed to by the foreign key.
func (o *RepairBlock) RepairAgent(mods ...qm.QueryMod) repairAgentQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.RepairAgentID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := RepairAgents(queryMods...)
	queries.SetFrom(query.Query, "\"repair_agents\"")

	return query
}

// RepairCase pointed to by the foreign key.
func (o *RepairBlock) RepairCase(mods ...qm.QueryMod) repairCaseQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.RepairCaseID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := RepairCases(queryMods...)
	queries.SetFrom(query.Query, "\"repair_cases\"")

	return query
}

// RepairOffer pointed to by the foreign key.
func (o *RepairBlock) RepairOffer(mods ...qm.QueryMod) repairOfferQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.RepairOfferID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := RepairOffers(queryMods...)
	queries.SetFrom(query.Query, "\"repair_offers\"")

	return query
}

// LoadRepairAgent allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (repairBlockL) LoadRepairAgent(e boil.Executor, singular bool, maybeRepairBlock interface{}, mods queries.Applicator) error {
	var slice []*RepairBlock
	var object *RepairBlock

	if singular {
		object = maybeRepairBlock.(*RepairBlock)
	} else {
		slice = *maybeRepairBlock.(*[]*RepairBlock)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &repairBlockR{}
		}
		args = append(args, object.RepairAgentID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &repairBlockR{}
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

	if len(repairBlockAfterSelectHooks) != 0 {
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
		foreign.R.RepairBlocks = append(foreign.R.RepairBlocks, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.RepairAgentID == foreign.ID {
				local.R.RepairAgent = foreign
				if foreign.R == nil {
					foreign.R = &repairAgentR{}
				}
				foreign.R.RepairBlocks = append(foreign.R.RepairBlocks, local)
				break
			}
		}
	}

	return nil
}

// LoadRepairCase allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (repairBlockL) LoadRepairCase(e boil.Executor, singular bool, maybeRepairBlock interface{}, mods queries.Applicator) error {
	var slice []*RepairBlock
	var object *RepairBlock

	if singular {
		object = maybeRepairBlock.(*RepairBlock)
	} else {
		slice = *maybeRepairBlock.(*[]*RepairBlock)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &repairBlockR{}
		}
		args = append(args, object.RepairCaseID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &repairBlockR{}
			}

			for _, a := range args {
				if a == obj.RepairCaseID {
					continue Outer
				}
			}

			args = append(args, obj.RepairCaseID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`repair_cases`),
		qm.WhereIn(`repair_cases.id in ?`, args...),
		qmhelper.WhereIsNull(`repair_cases.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load RepairCase")
	}

	var resultSlice []*RepairCase
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice RepairCase")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for repair_cases")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for repair_cases")
	}

	if len(repairBlockAfterSelectHooks) != 0 {
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
		object.R.RepairCase = foreign
		if foreign.R == nil {
			foreign.R = &repairCaseR{}
		}
		foreign.R.RepairBlocks = append(foreign.R.RepairBlocks, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.RepairCaseID == foreign.ID {
				local.R.RepairCase = foreign
				if foreign.R == nil {
					foreign.R = &repairCaseR{}
				}
				foreign.R.RepairBlocks = append(foreign.R.RepairBlocks, local)
				break
			}
		}
	}

	return nil
}

// LoadRepairOffer allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (repairBlockL) LoadRepairOffer(e boil.Executor, singular bool, maybeRepairBlock interface{}, mods queries.Applicator) error {
	var slice []*RepairBlock
	var object *RepairBlock

	if singular {
		object = maybeRepairBlock.(*RepairBlock)
	} else {
		slice = *maybeRepairBlock.(*[]*RepairBlock)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &repairBlockR{}
		}
		args = append(args, object.RepairOfferID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &repairBlockR{}
			}

			for _, a := range args {
				if a == obj.RepairOfferID {
					continue Outer
				}
			}

			args = append(args, obj.RepairOfferID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`repair_offers`),
		qm.WhereIn(`repair_offers.id in ?`, args...),
		qmhelper.WhereIsNull(`repair_offers.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load RepairOffer")
	}

	var resultSlice []*RepairOffer
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice RepairOffer")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for repair_offers")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for repair_offers")
	}

	if len(repairBlockAfterSelectHooks) != 0 {
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
		object.R.RepairOffer = foreign
		if foreign.R == nil {
			foreign.R = &repairOfferR{}
		}
		foreign.R.RepairBlocks = append(foreign.R.RepairBlocks, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.RepairOfferID == foreign.ID {
				local.R.RepairOffer = foreign
				if foreign.R == nil {
					foreign.R = &repairOfferR{}
				}
				foreign.R.RepairBlocks = append(foreign.R.RepairBlocks, local)
				break
			}
		}
	}

	return nil
}

// SetRepairAgent of the repairBlock to the related item.
// Sets o.R.RepairAgent to related.
// Adds o to related.R.RepairBlocks.
func (o *RepairBlock) SetRepairAgent(exec boil.Executor, insert bool, related *RepairAgent) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"repair_blocks\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"repair_agent_id"}),
		strmangle.WhereClause("\"", "\"", 2, repairBlockPrimaryKeyColumns),
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
		o.R = &repairBlockR{
			RepairAgent: related,
		}
	} else {
		o.R.RepairAgent = related
	}

	if related.R == nil {
		related.R = &repairAgentR{
			RepairBlocks: RepairBlockSlice{o},
		}
	} else {
		related.R.RepairBlocks = append(related.R.RepairBlocks, o)
	}

	return nil
}

// SetRepairCase of the repairBlock to the related item.
// Sets o.R.RepairCase to related.
// Adds o to related.R.RepairBlocks.
func (o *RepairBlock) SetRepairCase(exec boil.Executor, insert bool, related *RepairCase) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"repair_blocks\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"repair_case_id"}),
		strmangle.WhereClause("\"", "\"", 2, repairBlockPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.RepairCaseID = related.ID
	if o.R == nil {
		o.R = &repairBlockR{
			RepairCase: related,
		}
	} else {
		o.R.RepairCase = related
	}

	if related.R == nil {
		related.R = &repairCaseR{
			RepairBlocks: RepairBlockSlice{o},
		}
	} else {
		related.R.RepairBlocks = append(related.R.RepairBlocks, o)
	}

	return nil
}

// SetRepairOffer of the repairBlock to the related item.
// Sets o.R.RepairOffer to related.
// Adds o to related.R.RepairBlocks.
func (o *RepairBlock) SetRepairOffer(exec boil.Executor, insert bool, related *RepairOffer) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"repair_blocks\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"repair_offer_id"}),
		strmangle.WhereClause("\"", "\"", 2, repairBlockPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.RepairOfferID = related.ID
	if o.R == nil {
		o.R = &repairBlockR{
			RepairOffer: related,
		}
	} else {
		o.R.RepairOffer = related
	}

	if related.R == nil {
		related.R = &repairOfferR{
			RepairBlocks: RepairBlockSlice{o},
		}
	} else {
		related.R.RepairBlocks = append(related.R.RepairBlocks, o)
	}

	return nil
}

// RepairBlocks retrieves all the records using an executor.
func RepairBlocks(mods ...qm.QueryMod) repairBlockQuery {
	mods = append(mods, qm.From("\"repair_blocks\""))
	return repairBlockQuery{NewQuery(mods...)}
}

// FindRepairBlock retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindRepairBlock(exec boil.Executor, iD string, selectCols ...string) (*RepairBlock, error) {
	repairBlockObj := &RepairBlock{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"repair_blocks\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, repairBlockObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from repair_blocks")
	}

	if err = repairBlockObj.doAfterSelectHooks(exec); err != nil {
		return repairBlockObj, err
	}

	return repairBlockObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *RepairBlock) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no repair_blocks provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(repairBlockColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	repairBlockInsertCacheMut.RLock()
	cache, cached := repairBlockInsertCache[key]
	repairBlockInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			repairBlockAllColumns,
			repairBlockColumnsWithDefault,
			repairBlockColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(repairBlockType, repairBlockMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(repairBlockType, repairBlockMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"repair_blocks\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"repair_blocks\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into repair_blocks")
	}

	if !cached {
		repairBlockInsertCacheMut.Lock()
		repairBlockInsertCache[key] = cache
		repairBlockInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the RepairBlock.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *RepairBlock) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	repairBlockUpdateCacheMut.RLock()
	cache, cached := repairBlockUpdateCache[key]
	repairBlockUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			repairBlockAllColumns,
			repairBlockPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update repair_blocks, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"repair_blocks\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, repairBlockPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(repairBlockType, repairBlockMapping, append(wl, repairBlockPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update repair_blocks row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for repair_blocks")
	}

	if !cached {
		repairBlockUpdateCacheMut.Lock()
		repairBlockUpdateCache[key] = cache
		repairBlockUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q repairBlockQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for repair_blocks")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for repair_blocks")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o RepairBlockSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), repairBlockPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"repair_blocks\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, repairBlockPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in repairBlock slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all repairBlock")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *RepairBlock) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no repair_blocks provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}
	o.UpdatedAt = currTime

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(repairBlockColumnsWithDefault, o)

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

	repairBlockUpsertCacheMut.RLock()
	cache, cached := repairBlockUpsertCache[key]
	repairBlockUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			repairBlockAllColumns,
			repairBlockColumnsWithDefault,
			repairBlockColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			repairBlockAllColumns,
			repairBlockPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert repair_blocks, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(repairBlockPrimaryKeyColumns))
			copy(conflict, repairBlockPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"repair_blocks\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(repairBlockType, repairBlockMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(repairBlockType, repairBlockMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert repair_blocks")
	}

	if !cached {
		repairBlockUpsertCacheMut.Lock()
		repairBlockUpsertCache[key] = cache
		repairBlockUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single RepairBlock record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *RepairBlock) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no RepairBlock provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), repairBlockPrimaryKeyMapping)
	sql := "DELETE FROM \"repair_blocks\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from repair_blocks")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for repair_blocks")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q repairBlockQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no repairBlockQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from repair_blocks")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for repair_blocks")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o RepairBlockSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(repairBlockBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), repairBlockPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"repair_blocks\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, repairBlockPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from repairBlock slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for repair_blocks")
	}

	if len(repairBlockAfterDeleteHooks) != 0 {
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
func (o *RepairBlock) Reload(exec boil.Executor) error {
	ret, err := FindRepairBlock(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *RepairBlockSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := RepairBlockSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), repairBlockPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"repair_blocks\".* FROM \"repair_blocks\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, repairBlockPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in RepairBlockSlice")
	}

	*o = slice

	return nil
}

// RepairBlockExists checks if the RepairBlock row exists.
func RepairBlockExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"repair_blocks\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if repair_blocks exists")
	}

	return exists, nil
}