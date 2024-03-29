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

// SpoilsOfWar is an object representing the database table.
type SpoilsOfWar struct {
	ID                     string          `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	BattleID               string          `boiler:"battle_id" boil:"battle_id" json:"battle_id" toml:"battle_id" yaml:"battle_id"`
	BattleNumber           int             `boiler:"battle_number" boil:"battle_number" json:"battle_number" toml:"battle_number" yaml:"battle_number"`
	Amount                 decimal.Decimal `boiler:"amount" boil:"amount" json:"amount" toml:"amount" yaml:"amount"`
	AmountSent             decimal.Decimal `boiler:"amount_sent" boil:"amount_sent" json:"amount_sent" toml:"amount_sent" yaml:"amount_sent"`
	CreatedAt              time.Time       `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt              time.Time       `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CurrentTick            int             `boiler:"current_tick" boil:"current_tick" json:"current_tick" toml:"current_tick" yaml:"current_tick"`
	MaxTicks               int             `boiler:"max_ticks" boil:"max_ticks" json:"max_ticks" toml:"max_ticks" yaml:"max_ticks"`
	LeftoverAmount         decimal.Decimal `boiler:"leftover_amount" boil:"leftover_amount" json:"leftover_amount" toml:"leftover_amount" yaml:"leftover_amount"`
	LeftoversTransactionID null.String     `boiler:"leftovers_transaction_id" boil:"leftovers_transaction_id" json:"leftovers_transaction_id,omitempty" toml:"leftovers_transaction_id" yaml:"leftovers_transaction_id,omitempty"`

	R *spoilsOfWarR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L spoilsOfWarL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var SpoilsOfWarColumns = struct {
	ID                     string
	BattleID               string
	BattleNumber           string
	Amount                 string
	AmountSent             string
	CreatedAt              string
	UpdatedAt              string
	CurrentTick            string
	MaxTicks               string
	LeftoverAmount         string
	LeftoversTransactionID string
}{
	ID:                     "id",
	BattleID:               "battle_id",
	BattleNumber:           "battle_number",
	Amount:                 "amount",
	AmountSent:             "amount_sent",
	CreatedAt:              "created_at",
	UpdatedAt:              "updated_at",
	CurrentTick:            "current_tick",
	MaxTicks:               "max_ticks",
	LeftoverAmount:         "leftover_amount",
	LeftoversTransactionID: "leftovers_transaction_id",
}

var SpoilsOfWarTableColumns = struct {
	ID                     string
	BattleID               string
	BattleNumber           string
	Amount                 string
	AmountSent             string
	CreatedAt              string
	UpdatedAt              string
	CurrentTick            string
	MaxTicks               string
	LeftoverAmount         string
	LeftoversTransactionID string
}{
	ID:                     "spoils_of_war.id",
	BattleID:               "spoils_of_war.battle_id",
	BattleNumber:           "spoils_of_war.battle_number",
	Amount:                 "spoils_of_war.amount",
	AmountSent:             "spoils_of_war.amount_sent",
	CreatedAt:              "spoils_of_war.created_at",
	UpdatedAt:              "spoils_of_war.updated_at",
	CurrentTick:            "spoils_of_war.current_tick",
	MaxTicks:               "spoils_of_war.max_ticks",
	LeftoverAmount:         "spoils_of_war.leftover_amount",
	LeftoversTransactionID: "spoils_of_war.leftovers_transaction_id",
}

// Generated where

var SpoilsOfWarWhere = struct {
	ID                     whereHelperstring
	BattleID               whereHelperstring
	BattleNumber           whereHelperint
	Amount                 whereHelperdecimal_Decimal
	AmountSent             whereHelperdecimal_Decimal
	CreatedAt              whereHelpertime_Time
	UpdatedAt              whereHelpertime_Time
	CurrentTick            whereHelperint
	MaxTicks               whereHelperint
	LeftoverAmount         whereHelperdecimal_Decimal
	LeftoversTransactionID whereHelpernull_String
}{
	ID:                     whereHelperstring{field: "\"spoils_of_war\".\"id\""},
	BattleID:               whereHelperstring{field: "\"spoils_of_war\".\"battle_id\""},
	BattleNumber:           whereHelperint{field: "\"spoils_of_war\".\"battle_number\""},
	Amount:                 whereHelperdecimal_Decimal{field: "\"spoils_of_war\".\"amount\""},
	AmountSent:             whereHelperdecimal_Decimal{field: "\"spoils_of_war\".\"amount_sent\""},
	CreatedAt:              whereHelpertime_Time{field: "\"spoils_of_war\".\"created_at\""},
	UpdatedAt:              whereHelpertime_Time{field: "\"spoils_of_war\".\"updated_at\""},
	CurrentTick:            whereHelperint{field: "\"spoils_of_war\".\"current_tick\""},
	MaxTicks:               whereHelperint{field: "\"spoils_of_war\".\"max_ticks\""},
	LeftoverAmount:         whereHelperdecimal_Decimal{field: "\"spoils_of_war\".\"leftover_amount\""},
	LeftoversTransactionID: whereHelpernull_String{field: "\"spoils_of_war\".\"leftovers_transaction_id\""},
}

// SpoilsOfWarRels is where relationship names are stored.
var SpoilsOfWarRels = struct {
	Battle             string
	BattleNumberBattle string
}{
	Battle:             "Battle",
	BattleNumberBattle: "BattleNumberBattle",
}

// spoilsOfWarR is where relationships are stored.
type spoilsOfWarR struct {
	Battle             *Battle `boiler:"Battle" boil:"Battle" json:"Battle" toml:"Battle" yaml:"Battle"`
	BattleNumberBattle *Battle `boiler:"BattleNumberBattle" boil:"BattleNumberBattle" json:"BattleNumberBattle" toml:"BattleNumberBattle" yaml:"BattleNumberBattle"`
}

// NewStruct creates a new relationship struct
func (*spoilsOfWarR) NewStruct() *spoilsOfWarR {
	return &spoilsOfWarR{}
}

// spoilsOfWarL is where Load methods for each relationship are stored.
type spoilsOfWarL struct{}

var (
	spoilsOfWarAllColumns            = []string{"id", "battle_id", "battle_number", "amount", "amount_sent", "created_at", "updated_at", "current_tick", "max_ticks", "leftover_amount", "leftovers_transaction_id"}
	spoilsOfWarColumnsWithoutDefault = []string{"battle_id", "battle_number", "amount"}
	spoilsOfWarColumnsWithDefault    = []string{"id", "amount_sent", "created_at", "updated_at", "current_tick", "max_ticks", "leftover_amount", "leftovers_transaction_id"}
	spoilsOfWarPrimaryKeyColumns     = []string{"id"}
	spoilsOfWarGeneratedColumns      = []string{}
)

type (
	// SpoilsOfWarSlice is an alias for a slice of pointers to SpoilsOfWar.
	// This should almost always be used instead of []SpoilsOfWar.
	SpoilsOfWarSlice []*SpoilsOfWar
	// SpoilsOfWarHook is the signature for custom SpoilsOfWar hook methods
	SpoilsOfWarHook func(boil.Executor, *SpoilsOfWar) error

	spoilsOfWarQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	spoilsOfWarType                 = reflect.TypeOf(&SpoilsOfWar{})
	spoilsOfWarMapping              = queries.MakeStructMapping(spoilsOfWarType)
	spoilsOfWarPrimaryKeyMapping, _ = queries.BindMapping(spoilsOfWarType, spoilsOfWarMapping, spoilsOfWarPrimaryKeyColumns)
	spoilsOfWarInsertCacheMut       sync.RWMutex
	spoilsOfWarInsertCache          = make(map[string]insertCache)
	spoilsOfWarUpdateCacheMut       sync.RWMutex
	spoilsOfWarUpdateCache          = make(map[string]updateCache)
	spoilsOfWarUpsertCacheMut       sync.RWMutex
	spoilsOfWarUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var spoilsOfWarAfterSelectHooks []SpoilsOfWarHook

var spoilsOfWarBeforeInsertHooks []SpoilsOfWarHook
var spoilsOfWarAfterInsertHooks []SpoilsOfWarHook

var spoilsOfWarBeforeUpdateHooks []SpoilsOfWarHook
var spoilsOfWarAfterUpdateHooks []SpoilsOfWarHook

var spoilsOfWarBeforeDeleteHooks []SpoilsOfWarHook
var spoilsOfWarAfterDeleteHooks []SpoilsOfWarHook

var spoilsOfWarBeforeUpsertHooks []SpoilsOfWarHook
var spoilsOfWarAfterUpsertHooks []SpoilsOfWarHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *SpoilsOfWar) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range spoilsOfWarAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *SpoilsOfWar) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range spoilsOfWarBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *SpoilsOfWar) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range spoilsOfWarAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *SpoilsOfWar) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range spoilsOfWarBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *SpoilsOfWar) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range spoilsOfWarAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *SpoilsOfWar) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range spoilsOfWarBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *SpoilsOfWar) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range spoilsOfWarAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *SpoilsOfWar) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range spoilsOfWarBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *SpoilsOfWar) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range spoilsOfWarAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddSpoilsOfWarHook registers your hook function for all future operations.
func AddSpoilsOfWarHook(hookPoint boil.HookPoint, spoilsOfWarHook SpoilsOfWarHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		spoilsOfWarAfterSelectHooks = append(spoilsOfWarAfterSelectHooks, spoilsOfWarHook)
	case boil.BeforeInsertHook:
		spoilsOfWarBeforeInsertHooks = append(spoilsOfWarBeforeInsertHooks, spoilsOfWarHook)
	case boil.AfterInsertHook:
		spoilsOfWarAfterInsertHooks = append(spoilsOfWarAfterInsertHooks, spoilsOfWarHook)
	case boil.BeforeUpdateHook:
		spoilsOfWarBeforeUpdateHooks = append(spoilsOfWarBeforeUpdateHooks, spoilsOfWarHook)
	case boil.AfterUpdateHook:
		spoilsOfWarAfterUpdateHooks = append(spoilsOfWarAfterUpdateHooks, spoilsOfWarHook)
	case boil.BeforeDeleteHook:
		spoilsOfWarBeforeDeleteHooks = append(spoilsOfWarBeforeDeleteHooks, spoilsOfWarHook)
	case boil.AfterDeleteHook:
		spoilsOfWarAfterDeleteHooks = append(spoilsOfWarAfterDeleteHooks, spoilsOfWarHook)
	case boil.BeforeUpsertHook:
		spoilsOfWarBeforeUpsertHooks = append(spoilsOfWarBeforeUpsertHooks, spoilsOfWarHook)
	case boil.AfterUpsertHook:
		spoilsOfWarAfterUpsertHooks = append(spoilsOfWarAfterUpsertHooks, spoilsOfWarHook)
	}
}

// One returns a single spoilsOfWar record from the query.
func (q spoilsOfWarQuery) One(exec boil.Executor) (*SpoilsOfWar, error) {
	o := &SpoilsOfWar{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for spoils_of_war")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all SpoilsOfWar records from the query.
func (q spoilsOfWarQuery) All(exec boil.Executor) (SpoilsOfWarSlice, error) {
	var o []*SpoilsOfWar

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to SpoilsOfWar slice")
	}

	if len(spoilsOfWarAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all SpoilsOfWar records in the query.
func (q spoilsOfWarQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count spoils_of_war rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q spoilsOfWarQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if spoils_of_war exists")
	}

	return count > 0, nil
}

// Battle pointed to by the foreign key.
func (o *SpoilsOfWar) Battle(mods ...qm.QueryMod) battleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BattleID),
	}

	queryMods = append(queryMods, mods...)

	query := Battles(queryMods...)
	queries.SetFrom(query.Query, "\"battles\"")

	return query
}

// BattleNumberBattle pointed to by the foreign key.
func (o *SpoilsOfWar) BattleNumberBattle(mods ...qm.QueryMod) battleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"battle_number\" = ?", o.BattleNumber),
	}

	queryMods = append(queryMods, mods...)

	query := Battles(queryMods...)
	queries.SetFrom(query.Query, "\"battles\"")

	return query
}

// LoadBattle allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (spoilsOfWarL) LoadBattle(e boil.Executor, singular bool, maybeSpoilsOfWar interface{}, mods queries.Applicator) error {
	var slice []*SpoilsOfWar
	var object *SpoilsOfWar

	if singular {
		object = maybeSpoilsOfWar.(*SpoilsOfWar)
	} else {
		slice = *maybeSpoilsOfWar.(*[]*SpoilsOfWar)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &spoilsOfWarR{}
		}
		args = append(args, object.BattleID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &spoilsOfWarR{}
			}

			for _, a := range args {
				if a == obj.BattleID {
					continue Outer
				}
			}

			args = append(args, obj.BattleID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`battles`),
		qm.WhereIn(`battles.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Battle")
	}

	var resultSlice []*Battle
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Battle")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for battles")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for battles")
	}

	if len(spoilsOfWarAfterSelectHooks) != 0 {
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
		object.R.Battle = foreign
		if foreign.R == nil {
			foreign.R = &battleR{}
		}
		foreign.R.SpoilsOfWar = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.BattleID == foreign.ID {
				local.R.Battle = foreign
				if foreign.R == nil {
					foreign.R = &battleR{}
				}
				foreign.R.SpoilsOfWar = local
				break
			}
		}
	}

	return nil
}

// LoadBattleNumberBattle allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (spoilsOfWarL) LoadBattleNumberBattle(e boil.Executor, singular bool, maybeSpoilsOfWar interface{}, mods queries.Applicator) error {
	var slice []*SpoilsOfWar
	var object *SpoilsOfWar

	if singular {
		object = maybeSpoilsOfWar.(*SpoilsOfWar)
	} else {
		slice = *maybeSpoilsOfWar.(*[]*SpoilsOfWar)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &spoilsOfWarR{}
		}
		args = append(args, object.BattleNumber)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &spoilsOfWarR{}
			}

			for _, a := range args {
				if a == obj.BattleNumber {
					continue Outer
				}
			}

			args = append(args, obj.BattleNumber)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`battles`),
		qm.WhereIn(`battles.battle_number in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Battle")
	}

	var resultSlice []*Battle
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Battle")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for battles")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for battles")
	}

	if len(spoilsOfWarAfterSelectHooks) != 0 {
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
		object.R.BattleNumberBattle = foreign
		if foreign.R == nil {
			foreign.R = &battleR{}
		}
		foreign.R.BattleNumberSpoilsOfWar = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.BattleNumber == foreign.BattleNumber {
				local.R.BattleNumberBattle = foreign
				if foreign.R == nil {
					foreign.R = &battleR{}
				}
				foreign.R.BattleNumberSpoilsOfWar = local
				break
			}
		}
	}

	return nil
}

// SetBattle of the spoilsOfWar to the related item.
// Sets o.R.Battle to related.
// Adds o to related.R.SpoilsOfWar.
func (o *SpoilsOfWar) SetBattle(exec boil.Executor, insert bool, related *Battle) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"spoils_of_war\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"battle_id"}),
		strmangle.WhereClause("\"", "\"", 2, spoilsOfWarPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.BattleID = related.ID
	if o.R == nil {
		o.R = &spoilsOfWarR{
			Battle: related,
		}
	} else {
		o.R.Battle = related
	}

	if related.R == nil {
		related.R = &battleR{
			SpoilsOfWar: o,
		}
	} else {
		related.R.SpoilsOfWar = o
	}

	return nil
}

// SetBattleNumberBattle of the spoilsOfWar to the related item.
// Sets o.R.BattleNumberBattle to related.
// Adds o to related.R.BattleNumberSpoilsOfWar.
func (o *SpoilsOfWar) SetBattleNumberBattle(exec boil.Executor, insert bool, related *Battle) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"spoils_of_war\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"battle_number"}),
		strmangle.WhereClause("\"", "\"", 2, spoilsOfWarPrimaryKeyColumns),
	)
	values := []interface{}{related.BattleNumber, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.BattleNumber = related.BattleNumber
	if o.R == nil {
		o.R = &spoilsOfWarR{
			BattleNumberBattle: related,
		}
	} else {
		o.R.BattleNumberBattle = related
	}

	if related.R == nil {
		related.R = &battleR{
			BattleNumberSpoilsOfWar: o,
		}
	} else {
		related.R.BattleNumberSpoilsOfWar = o
	}

	return nil
}

// SpoilsOfWars retrieves all the records using an executor.
func SpoilsOfWars(mods ...qm.QueryMod) spoilsOfWarQuery {
	mods = append(mods, qm.From("\"spoils_of_war\""))
	return spoilsOfWarQuery{NewQuery(mods...)}
}

// FindSpoilsOfWar retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindSpoilsOfWar(exec boil.Executor, iD string, selectCols ...string) (*SpoilsOfWar, error) {
	spoilsOfWarObj := &SpoilsOfWar{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"spoils_of_war\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, spoilsOfWarObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from spoils_of_war")
	}

	if err = spoilsOfWarObj.doAfterSelectHooks(exec); err != nil {
		return spoilsOfWarObj, err
	}

	return spoilsOfWarObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *SpoilsOfWar) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no spoils_of_war provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(spoilsOfWarColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	spoilsOfWarInsertCacheMut.RLock()
	cache, cached := spoilsOfWarInsertCache[key]
	spoilsOfWarInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			spoilsOfWarAllColumns,
			spoilsOfWarColumnsWithDefault,
			spoilsOfWarColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(spoilsOfWarType, spoilsOfWarMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(spoilsOfWarType, spoilsOfWarMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"spoils_of_war\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"spoils_of_war\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into spoils_of_war")
	}

	if !cached {
		spoilsOfWarInsertCacheMut.Lock()
		spoilsOfWarInsertCache[key] = cache
		spoilsOfWarInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the SpoilsOfWar.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *SpoilsOfWar) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	spoilsOfWarUpdateCacheMut.RLock()
	cache, cached := spoilsOfWarUpdateCache[key]
	spoilsOfWarUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			spoilsOfWarAllColumns,
			spoilsOfWarPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update spoils_of_war, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"spoils_of_war\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, spoilsOfWarPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(spoilsOfWarType, spoilsOfWarMapping, append(wl, spoilsOfWarPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update spoils_of_war row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for spoils_of_war")
	}

	if !cached {
		spoilsOfWarUpdateCacheMut.Lock()
		spoilsOfWarUpdateCache[key] = cache
		spoilsOfWarUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q spoilsOfWarQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for spoils_of_war")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for spoils_of_war")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o SpoilsOfWarSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), spoilsOfWarPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"spoils_of_war\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, spoilsOfWarPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in spoilsOfWar slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all spoilsOfWar")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *SpoilsOfWar) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no spoils_of_war provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}
	o.UpdatedAt = currTime

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(spoilsOfWarColumnsWithDefault, o)

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

	spoilsOfWarUpsertCacheMut.RLock()
	cache, cached := spoilsOfWarUpsertCache[key]
	spoilsOfWarUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			spoilsOfWarAllColumns,
			spoilsOfWarColumnsWithDefault,
			spoilsOfWarColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			spoilsOfWarAllColumns,
			spoilsOfWarPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert spoils_of_war, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(spoilsOfWarPrimaryKeyColumns))
			copy(conflict, spoilsOfWarPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"spoils_of_war\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(spoilsOfWarType, spoilsOfWarMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(spoilsOfWarType, spoilsOfWarMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert spoils_of_war")
	}

	if !cached {
		spoilsOfWarUpsertCacheMut.Lock()
		spoilsOfWarUpsertCache[key] = cache
		spoilsOfWarUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single SpoilsOfWar record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *SpoilsOfWar) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no SpoilsOfWar provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), spoilsOfWarPrimaryKeyMapping)
	sql := "DELETE FROM \"spoils_of_war\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from spoils_of_war")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for spoils_of_war")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q spoilsOfWarQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no spoilsOfWarQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from spoils_of_war")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for spoils_of_war")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o SpoilsOfWarSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(spoilsOfWarBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), spoilsOfWarPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"spoils_of_war\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, spoilsOfWarPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from spoilsOfWar slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for spoils_of_war")
	}

	if len(spoilsOfWarAfterDeleteHooks) != 0 {
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
func (o *SpoilsOfWar) Reload(exec boil.Executor) error {
	ret, err := FindSpoilsOfWar(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *SpoilsOfWarSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := SpoilsOfWarSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), spoilsOfWarPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"spoils_of_war\".* FROM \"spoils_of_war\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, spoilsOfWarPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in SpoilsOfWarSlice")
	}

	*o = slice

	return nil
}

// SpoilsOfWarExists checks if the SpoilsOfWar row exists.
func SpoilsOfWarExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"spoils_of_war\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if spoils_of_war exists")
	}

	return exists, nil
}
