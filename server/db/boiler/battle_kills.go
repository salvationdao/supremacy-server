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

// BattleKill is an object representing the database table.
type BattleKill struct {
	BattleID  string    `boiler:"battle_id" boil:"battle_id" json:"battle_id" toml:"battle_id" yaml:"battle_id"`
	MechID    string    `boiler:"mech_id" boil:"mech_id" json:"mech_id" toml:"mech_id" yaml:"mech_id"`
	KilledID  string    `boiler:"killed_id" boil:"killed_id" json:"killed_id" toml:"killed_id" yaml:"killed_id"`
	CreatedAt time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *battleKillR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L battleKillL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var BattleKillColumns = struct {
	BattleID  string
	MechID    string
	KilledID  string
	CreatedAt string
}{
	BattleID:  "battle_id",
	MechID:    "mech_id",
	KilledID:  "killed_id",
	CreatedAt: "created_at",
}

var BattleKillTableColumns = struct {
	BattleID  string
	MechID    string
	KilledID  string
	CreatedAt string
}{
	BattleID:  "battle_kills.battle_id",
	MechID:    "battle_kills.mech_id",
	KilledID:  "battle_kills.killed_id",
	CreatedAt: "battle_kills.created_at",
}

// Generated where

var BattleKillWhere = struct {
	BattleID  whereHelperstring
	MechID    whereHelperstring
	KilledID  whereHelperstring
	CreatedAt whereHelpertime_Time
}{
	BattleID:  whereHelperstring{field: "\"battle_kills\".\"battle_id\""},
	MechID:    whereHelperstring{field: "\"battle_kills\".\"mech_id\""},
	KilledID:  whereHelperstring{field: "\"battle_kills\".\"killed_id\""},
	CreatedAt: whereHelpertime_Time{field: "\"battle_kills\".\"created_at\""},
}

// BattleKillRels is where relationship names are stored.
var BattleKillRels = struct {
	Battle string
	Killed string
	Mech   string
}{
	Battle: "Battle",
	Killed: "Killed",
	Mech:   "Mech",
}

// battleKillR is where relationships are stored.
type battleKillR struct {
	Battle *Battle `boiler:"Battle" boil:"Battle" json:"Battle" toml:"Battle" yaml:"Battle"`
	Killed *Mech   `boiler:"Killed" boil:"Killed" json:"Killed" toml:"Killed" yaml:"Killed"`
	Mech   *Mech   `boiler:"Mech" boil:"Mech" json:"Mech" toml:"Mech" yaml:"Mech"`
}

// NewStruct creates a new relationship struct
func (*battleKillR) NewStruct() *battleKillR {
	return &battleKillR{}
}

// battleKillL is where Load methods for each relationship are stored.
type battleKillL struct{}

var (
	battleKillAllColumns            = []string{"battle_id", "mech_id", "killed_id", "created_at"}
	battleKillColumnsWithoutDefault = []string{"battle_id", "mech_id", "killed_id"}
	battleKillColumnsWithDefault    = []string{"created_at"}
	battleKillPrimaryKeyColumns     = []string{"battle_id", "killed_id"}
	battleKillGeneratedColumns      = []string{}
)

type (
	// BattleKillSlice is an alias for a slice of pointers to BattleKill.
	// This should almost always be used instead of []BattleKill.
	BattleKillSlice []*BattleKill
	// BattleKillHook is the signature for custom BattleKill hook methods
	BattleKillHook func(boil.Executor, *BattleKill) error

	battleKillQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	battleKillType                 = reflect.TypeOf(&BattleKill{})
	battleKillMapping              = queries.MakeStructMapping(battleKillType)
	battleKillPrimaryKeyMapping, _ = queries.BindMapping(battleKillType, battleKillMapping, battleKillPrimaryKeyColumns)
	battleKillInsertCacheMut       sync.RWMutex
	battleKillInsertCache          = make(map[string]insertCache)
	battleKillUpdateCacheMut       sync.RWMutex
	battleKillUpdateCache          = make(map[string]updateCache)
	battleKillUpsertCacheMut       sync.RWMutex
	battleKillUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var battleKillAfterSelectHooks []BattleKillHook

var battleKillBeforeInsertHooks []BattleKillHook
var battleKillAfterInsertHooks []BattleKillHook

var battleKillBeforeUpdateHooks []BattleKillHook
var battleKillAfterUpdateHooks []BattleKillHook

var battleKillBeforeDeleteHooks []BattleKillHook
var battleKillAfterDeleteHooks []BattleKillHook

var battleKillBeforeUpsertHooks []BattleKillHook
var battleKillAfterUpsertHooks []BattleKillHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *BattleKill) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range battleKillAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *BattleKill) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleKillBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *BattleKill) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleKillAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *BattleKill) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range battleKillBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *BattleKill) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range battleKillAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *BattleKill) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range battleKillBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *BattleKill) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range battleKillAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *BattleKill) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleKillBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *BattleKill) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleKillAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddBattleKillHook registers your hook function for all future operations.
func AddBattleKillHook(hookPoint boil.HookPoint, battleKillHook BattleKillHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		battleKillAfterSelectHooks = append(battleKillAfterSelectHooks, battleKillHook)
	case boil.BeforeInsertHook:
		battleKillBeforeInsertHooks = append(battleKillBeforeInsertHooks, battleKillHook)
	case boil.AfterInsertHook:
		battleKillAfterInsertHooks = append(battleKillAfterInsertHooks, battleKillHook)
	case boil.BeforeUpdateHook:
		battleKillBeforeUpdateHooks = append(battleKillBeforeUpdateHooks, battleKillHook)
	case boil.AfterUpdateHook:
		battleKillAfterUpdateHooks = append(battleKillAfterUpdateHooks, battleKillHook)
	case boil.BeforeDeleteHook:
		battleKillBeforeDeleteHooks = append(battleKillBeforeDeleteHooks, battleKillHook)
	case boil.AfterDeleteHook:
		battleKillAfterDeleteHooks = append(battleKillAfterDeleteHooks, battleKillHook)
	case boil.BeforeUpsertHook:
		battleKillBeforeUpsertHooks = append(battleKillBeforeUpsertHooks, battleKillHook)
	case boil.AfterUpsertHook:
		battleKillAfterUpsertHooks = append(battleKillAfterUpsertHooks, battleKillHook)
	}
}

// One returns a single battleKill record from the query.
func (q battleKillQuery) One(exec boil.Executor) (*BattleKill, error) {
	o := &BattleKill{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for battle_kills")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all BattleKill records from the query.
func (q battleKillQuery) All(exec boil.Executor) (BattleKillSlice, error) {
	var o []*BattleKill

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to BattleKill slice")
	}

	if len(battleKillAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all BattleKill records in the query.
func (q battleKillQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count battle_kills rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q battleKillQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if battle_kills exists")
	}

	return count > 0, nil
}

// Battle pointed to by the foreign key.
func (o *BattleKill) Battle(mods ...qm.QueryMod) battleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BattleID),
	}

	queryMods = append(queryMods, mods...)

	query := Battles(queryMods...)
	queries.SetFrom(query.Query, "\"battles\"")

	return query
}

// Killed pointed to by the foreign key.
func (o *BattleKill) Killed(mods ...qm.QueryMod) mechQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.KilledID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Mechs(queryMods...)
	queries.SetFrom(query.Query, "\"mechs\"")

	return query
}

// Mech pointed to by the foreign key.
func (o *BattleKill) Mech(mods ...qm.QueryMod) mechQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.MechID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Mechs(queryMods...)
	queries.SetFrom(query.Query, "\"mechs\"")

	return query
}

// LoadBattle allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (battleKillL) LoadBattle(e boil.Executor, singular bool, maybeBattleKill interface{}, mods queries.Applicator) error {
	var slice []*BattleKill
	var object *BattleKill

	if singular {
		object = maybeBattleKill.(*BattleKill)
	} else {
		slice = *maybeBattleKill.(*[]*BattleKill)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &battleKillR{}
		}
		args = append(args, object.BattleID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &battleKillR{}
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

	if len(battleKillAfterSelectHooks) != 0 {
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
		foreign.R.BattleKills = append(foreign.R.BattleKills, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.BattleID == foreign.ID {
				local.R.Battle = foreign
				if foreign.R == nil {
					foreign.R = &battleR{}
				}
				foreign.R.BattleKills = append(foreign.R.BattleKills, local)
				break
			}
		}
	}

	return nil
}

// LoadKilled allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (battleKillL) LoadKilled(e boil.Executor, singular bool, maybeBattleKill interface{}, mods queries.Applicator) error {
	var slice []*BattleKill
	var object *BattleKill

	if singular {
		object = maybeBattleKill.(*BattleKill)
	} else {
		slice = *maybeBattleKill.(*[]*BattleKill)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &battleKillR{}
		}
		args = append(args, object.KilledID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &battleKillR{}
			}

			for _, a := range args {
				if a == obj.KilledID {
					continue Outer
				}
			}

			args = append(args, obj.KilledID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`mechs`),
		qm.WhereIn(`mechs.id in ?`, args...),
		qmhelper.WhereIsNull(`mechs.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Mech")
	}

	var resultSlice []*Mech
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Mech")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for mechs")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for mechs")
	}

	if len(battleKillAfterSelectHooks) != 0 {
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
		object.R.Killed = foreign
		if foreign.R == nil {
			foreign.R = &mechR{}
		}
		foreign.R.KilledBattleKills = append(foreign.R.KilledBattleKills, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.KilledID == foreign.ID {
				local.R.Killed = foreign
				if foreign.R == nil {
					foreign.R = &mechR{}
				}
				foreign.R.KilledBattleKills = append(foreign.R.KilledBattleKills, local)
				break
			}
		}
	}

	return nil
}

// LoadMech allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (battleKillL) LoadMech(e boil.Executor, singular bool, maybeBattleKill interface{}, mods queries.Applicator) error {
	var slice []*BattleKill
	var object *BattleKill

	if singular {
		object = maybeBattleKill.(*BattleKill)
	} else {
		slice = *maybeBattleKill.(*[]*BattleKill)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &battleKillR{}
		}
		args = append(args, object.MechID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &battleKillR{}
			}

			for _, a := range args {
				if a == obj.MechID {
					continue Outer
				}
			}

			args = append(args, obj.MechID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`mechs`),
		qm.WhereIn(`mechs.id in ?`, args...),
		qmhelper.WhereIsNull(`mechs.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Mech")
	}

	var resultSlice []*Mech
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Mech")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for mechs")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for mechs")
	}

	if len(battleKillAfterSelectHooks) != 0 {
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
		object.R.Mech = foreign
		if foreign.R == nil {
			foreign.R = &mechR{}
		}
		foreign.R.BattleKills = append(foreign.R.BattleKills, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.MechID == foreign.ID {
				local.R.Mech = foreign
				if foreign.R == nil {
					foreign.R = &mechR{}
				}
				foreign.R.BattleKills = append(foreign.R.BattleKills, local)
				break
			}
		}
	}

	return nil
}

// SetBattle of the battleKill to the related item.
// Sets o.R.Battle to related.
// Adds o to related.R.BattleKills.
func (o *BattleKill) SetBattle(exec boil.Executor, insert bool, related *Battle) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"battle_kills\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"battle_id"}),
		strmangle.WhereClause("\"", "\"", 2, battleKillPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.BattleID, o.KilledID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.BattleID = related.ID
	if o.R == nil {
		o.R = &battleKillR{
			Battle: related,
		}
	} else {
		o.R.Battle = related
	}

	if related.R == nil {
		related.R = &battleR{
			BattleKills: BattleKillSlice{o},
		}
	} else {
		related.R.BattleKills = append(related.R.BattleKills, o)
	}

	return nil
}

// SetKilled of the battleKill to the related item.
// Sets o.R.Killed to related.
// Adds o to related.R.KilledBattleKills.
func (o *BattleKill) SetKilled(exec boil.Executor, insert bool, related *Mech) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"battle_kills\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"killed_id"}),
		strmangle.WhereClause("\"", "\"", 2, battleKillPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.BattleID, o.KilledID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.KilledID = related.ID
	if o.R == nil {
		o.R = &battleKillR{
			Killed: related,
		}
	} else {
		o.R.Killed = related
	}

	if related.R == nil {
		related.R = &mechR{
			KilledBattleKills: BattleKillSlice{o},
		}
	} else {
		related.R.KilledBattleKills = append(related.R.KilledBattleKills, o)
	}

	return nil
}

// SetMech of the battleKill to the related item.
// Sets o.R.Mech to related.
// Adds o to related.R.BattleKills.
func (o *BattleKill) SetMech(exec boil.Executor, insert bool, related *Mech) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"battle_kills\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"mech_id"}),
		strmangle.WhereClause("\"", "\"", 2, battleKillPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.BattleID, o.KilledID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.MechID = related.ID
	if o.R == nil {
		o.R = &battleKillR{
			Mech: related,
		}
	} else {
		o.R.Mech = related
	}

	if related.R == nil {
		related.R = &mechR{
			BattleKills: BattleKillSlice{o},
		}
	} else {
		related.R.BattleKills = append(related.R.BattleKills, o)
	}

	return nil
}

// BattleKills retrieves all the records using an executor.
func BattleKills(mods ...qm.QueryMod) battleKillQuery {
	mods = append(mods, qm.From("\"battle_kills\""))
	return battleKillQuery{NewQuery(mods...)}
}

// FindBattleKill retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindBattleKill(exec boil.Executor, battleID string, killedID string, selectCols ...string) (*BattleKill, error) {
	battleKillObj := &BattleKill{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"battle_kills\" where \"battle_id\"=$1 AND \"killed_id\"=$2", sel,
	)

	q := queries.Raw(query, battleID, killedID)

	err := q.Bind(nil, exec, battleKillObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from battle_kills")
	}

	if err = battleKillObj.doAfterSelectHooks(exec); err != nil {
		return battleKillObj, err
	}

	return battleKillObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *BattleKill) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no battle_kills provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(battleKillColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	battleKillInsertCacheMut.RLock()
	cache, cached := battleKillInsertCache[key]
	battleKillInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			battleKillAllColumns,
			battleKillColumnsWithDefault,
			battleKillColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(battleKillType, battleKillMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(battleKillType, battleKillMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"battle_kills\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"battle_kills\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into battle_kills")
	}

	if !cached {
		battleKillInsertCacheMut.Lock()
		battleKillInsertCache[key] = cache
		battleKillInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the BattleKill.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *BattleKill) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	battleKillUpdateCacheMut.RLock()
	cache, cached := battleKillUpdateCache[key]
	battleKillUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			battleKillAllColumns,
			battleKillPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update battle_kills, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"battle_kills\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, battleKillPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(battleKillType, battleKillMapping, append(wl, battleKillPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update battle_kills row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for battle_kills")
	}

	if !cached {
		battleKillUpdateCacheMut.Lock()
		battleKillUpdateCache[key] = cache
		battleKillUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q battleKillQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for battle_kills")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for battle_kills")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o BattleKillSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleKillPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"battle_kills\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, battleKillPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in battleKill slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all battleKill")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *BattleKill) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no battle_kills provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(battleKillColumnsWithDefault, o)

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

	battleKillUpsertCacheMut.RLock()
	cache, cached := battleKillUpsertCache[key]
	battleKillUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			battleKillAllColumns,
			battleKillColumnsWithDefault,
			battleKillColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			battleKillAllColumns,
			battleKillPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert battle_kills, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(battleKillPrimaryKeyColumns))
			copy(conflict, battleKillPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"battle_kills\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(battleKillType, battleKillMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(battleKillType, battleKillMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert battle_kills")
	}

	if !cached {
		battleKillUpsertCacheMut.Lock()
		battleKillUpsertCache[key] = cache
		battleKillUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single BattleKill record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *BattleKill) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no BattleKill provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), battleKillPrimaryKeyMapping)
	sql := "DELETE FROM \"battle_kills\" WHERE \"battle_id\"=$1 AND \"killed_id\"=$2"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from battle_kills")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for battle_kills")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q battleKillQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no battleKillQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from battle_kills")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for battle_kills")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o BattleKillSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(battleKillBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleKillPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"battle_kills\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, battleKillPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from battleKill slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for battle_kills")
	}

	if len(battleKillAfterDeleteHooks) != 0 {
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
func (o *BattleKill) Reload(exec boil.Executor) error {
	ret, err := FindBattleKill(exec, o.BattleID, o.KilledID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *BattleKillSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := BattleKillSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleKillPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"battle_kills\".* FROM \"battle_kills\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, battleKillPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in BattleKillSlice")
	}

	*o = slice

	return nil
}

// BattleKillExists checks if the BattleKill row exists.
func BattleKillExists(exec boil.Executor, battleID string, killedID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"battle_kills\" where \"battle_id\"=$1 AND \"killed_id\"=$2 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, battleID, killedID)
	}
	row := exec.QueryRow(sql, battleID, killedID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if battle_kills exists")
	}

	return exists, nil
}
