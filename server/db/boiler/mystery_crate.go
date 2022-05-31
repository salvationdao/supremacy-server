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

// MysteryCrate is an object representing the database table.
type MysteryCrate struct {
	ID          string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Type        string    `boiler:"type" boil:"type" json:"type" toml:"type" yaml:"type"`
	FactionID   string    `boiler:"faction_id" boil:"faction_id" json:"faction_id" toml:"faction_id" yaml:"faction_id"`
	Label       string    `boiler:"label" boil:"label" json:"label" toml:"label" yaml:"label"`
	Opened      bool      `boiler:"opened" boil:"opened" json:"opened" toml:"opened" yaml:"opened"`
	LockedUntil time.Time `boiler:"locked_until" boil:"locked_until" json:"locked_until" toml:"locked_until" yaml:"locked_until"`
	Purchased   bool      `boiler:"purchased" boil:"purchased" json:"purchased" toml:"purchased" yaml:"purchased"`
	DeletedAt   null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	UpdatedAt   time.Time `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt   time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *mysteryCrateR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L mysteryCrateL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var MysteryCrateColumns = struct {
	ID          string
	Type        string
	FactionID   string
	Label       string
	Opened      string
	LockedUntil string
	Purchased   string
	DeletedAt   string
	UpdatedAt   string
	CreatedAt   string
}{
	ID:          "id",
	Type:        "type",
	FactionID:   "faction_id",
	Label:       "label",
	Opened:      "opened",
	LockedUntil: "locked_until",
	Purchased:   "purchased",
	DeletedAt:   "deleted_at",
	UpdatedAt:   "updated_at",
	CreatedAt:   "created_at",
}

var MysteryCrateTableColumns = struct {
	ID          string
	Type        string
	FactionID   string
	Label       string
	Opened      string
	LockedUntil string
	Purchased   string
	DeletedAt   string
	UpdatedAt   string
	CreatedAt   string
}{
	ID:          "mystery_crate.id",
	Type:        "mystery_crate.type",
	FactionID:   "mystery_crate.faction_id",
	Label:       "mystery_crate.label",
	Opened:      "mystery_crate.opened",
	LockedUntil: "mystery_crate.locked_until",
	Purchased:   "mystery_crate.purchased",
	DeletedAt:   "mystery_crate.deleted_at",
	UpdatedAt:   "mystery_crate.updated_at",
	CreatedAt:   "mystery_crate.created_at",
}

// Generated where

var MysteryCrateWhere = struct {
	ID          whereHelperstring
	Type        whereHelperstring
	FactionID   whereHelperstring
	Label       whereHelperstring
	Opened      whereHelperbool
	LockedUntil whereHelpertime_Time
	Purchased   whereHelperbool
	DeletedAt   whereHelpernull_Time
	UpdatedAt   whereHelpertime_Time
	CreatedAt   whereHelpertime_Time
}{
	ID:          whereHelperstring{field: "\"mystery_crate\".\"id\""},
	Type:        whereHelperstring{field: "\"mystery_crate\".\"type\""},
	FactionID:   whereHelperstring{field: "\"mystery_crate\".\"faction_id\""},
	Label:       whereHelperstring{field: "\"mystery_crate\".\"label\""},
	Opened:      whereHelperbool{field: "\"mystery_crate\".\"opened\""},
	LockedUntil: whereHelpertime_Time{field: "\"mystery_crate\".\"locked_until\""},
	Purchased:   whereHelperbool{field: "\"mystery_crate\".\"purchased\""},
	DeletedAt:   whereHelpernull_Time{field: "\"mystery_crate\".\"deleted_at\""},
	UpdatedAt:   whereHelpertime_Time{field: "\"mystery_crate\".\"updated_at\""},
	CreatedAt:   whereHelpertime_Time{field: "\"mystery_crate\".\"created_at\""},
}

// MysteryCrateRels is where relationship names are stored.
var MysteryCrateRels = struct {
	Faction                string
	MysteryCrateBlueprints string
}{
	Faction:                "Faction",
	MysteryCrateBlueprints: "MysteryCrateBlueprints",
}

// mysteryCrateR is where relationships are stored.
type mysteryCrateR struct {
	Faction                *Faction                   `boiler:"Faction" boil:"Faction" json:"Faction" toml:"Faction" yaml:"Faction"`
	MysteryCrateBlueprints MysteryCrateBlueprintSlice `boiler:"MysteryCrateBlueprints" boil:"MysteryCrateBlueprints" json:"MysteryCrateBlueprints" toml:"MysteryCrateBlueprints" yaml:"MysteryCrateBlueprints"`
}

// NewStruct creates a new relationship struct
func (*mysteryCrateR) NewStruct() *mysteryCrateR {
	return &mysteryCrateR{}
}

// mysteryCrateL is where Load methods for each relationship are stored.
type mysteryCrateL struct{}

var (
	mysteryCrateAllColumns            = []string{"id", "type", "faction_id", "label", "opened", "locked_until", "purchased", "deleted_at", "updated_at", "created_at"}
	mysteryCrateColumnsWithoutDefault = []string{"type", "faction_id", "label"}
	mysteryCrateColumnsWithDefault    = []string{"id", "opened", "locked_until", "purchased", "deleted_at", "updated_at", "created_at"}
	mysteryCratePrimaryKeyColumns     = []string{"id"}
	mysteryCrateGeneratedColumns      = []string{}
)

type (
	// MysteryCrateSlice is an alias for a slice of pointers to MysteryCrate.
	// This should almost always be used instead of []MysteryCrate.
	MysteryCrateSlice []*MysteryCrate
	// MysteryCrateHook is the signature for custom MysteryCrate hook methods
	MysteryCrateHook func(boil.Executor, *MysteryCrate) error

	mysteryCrateQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	mysteryCrateType                 = reflect.TypeOf(&MysteryCrate{})
	mysteryCrateMapping              = queries.MakeStructMapping(mysteryCrateType)
	mysteryCratePrimaryKeyMapping, _ = queries.BindMapping(mysteryCrateType, mysteryCrateMapping, mysteryCratePrimaryKeyColumns)
	mysteryCrateInsertCacheMut       sync.RWMutex
	mysteryCrateInsertCache          = make(map[string]insertCache)
	mysteryCrateUpdateCacheMut       sync.RWMutex
	mysteryCrateUpdateCache          = make(map[string]updateCache)
	mysteryCrateUpsertCacheMut       sync.RWMutex
	mysteryCrateUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var mysteryCrateAfterSelectHooks []MysteryCrateHook

var mysteryCrateBeforeInsertHooks []MysteryCrateHook
var mysteryCrateAfterInsertHooks []MysteryCrateHook

var mysteryCrateBeforeUpdateHooks []MysteryCrateHook
var mysteryCrateAfterUpdateHooks []MysteryCrateHook

var mysteryCrateBeforeDeleteHooks []MysteryCrateHook
var mysteryCrateAfterDeleteHooks []MysteryCrateHook

var mysteryCrateBeforeUpsertHooks []MysteryCrateHook
var mysteryCrateAfterUpsertHooks []MysteryCrateHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *MysteryCrate) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range mysteryCrateAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *MysteryCrate) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range mysteryCrateBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *MysteryCrate) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range mysteryCrateAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *MysteryCrate) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range mysteryCrateBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *MysteryCrate) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range mysteryCrateAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *MysteryCrate) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range mysteryCrateBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *MysteryCrate) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range mysteryCrateAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *MysteryCrate) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range mysteryCrateBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *MysteryCrate) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range mysteryCrateAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddMysteryCrateHook registers your hook function for all future operations.
func AddMysteryCrateHook(hookPoint boil.HookPoint, mysteryCrateHook MysteryCrateHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		mysteryCrateAfterSelectHooks = append(mysteryCrateAfterSelectHooks, mysteryCrateHook)
	case boil.BeforeInsertHook:
		mysteryCrateBeforeInsertHooks = append(mysteryCrateBeforeInsertHooks, mysteryCrateHook)
	case boil.AfterInsertHook:
		mysteryCrateAfterInsertHooks = append(mysteryCrateAfterInsertHooks, mysteryCrateHook)
	case boil.BeforeUpdateHook:
		mysteryCrateBeforeUpdateHooks = append(mysteryCrateBeforeUpdateHooks, mysteryCrateHook)
	case boil.AfterUpdateHook:
		mysteryCrateAfterUpdateHooks = append(mysteryCrateAfterUpdateHooks, mysteryCrateHook)
	case boil.BeforeDeleteHook:
		mysteryCrateBeforeDeleteHooks = append(mysteryCrateBeforeDeleteHooks, mysteryCrateHook)
	case boil.AfterDeleteHook:
		mysteryCrateAfterDeleteHooks = append(mysteryCrateAfterDeleteHooks, mysteryCrateHook)
	case boil.BeforeUpsertHook:
		mysteryCrateBeforeUpsertHooks = append(mysteryCrateBeforeUpsertHooks, mysteryCrateHook)
	case boil.AfterUpsertHook:
		mysteryCrateAfterUpsertHooks = append(mysteryCrateAfterUpsertHooks, mysteryCrateHook)
	}
}

// One returns a single mysteryCrate record from the query.
func (q mysteryCrateQuery) One(exec boil.Executor) (*MysteryCrate, error) {
	o := &MysteryCrate{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for mystery_crate")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all MysteryCrate records from the query.
func (q mysteryCrateQuery) All(exec boil.Executor) (MysteryCrateSlice, error) {
	var o []*MysteryCrate

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to MysteryCrate slice")
	}

	if len(mysteryCrateAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all MysteryCrate records in the query.
func (q mysteryCrateQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count mystery_crate rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q mysteryCrateQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if mystery_crate exists")
	}

	return count > 0, nil
}

// Faction pointed to by the foreign key.
func (o *MysteryCrate) Faction(mods ...qm.QueryMod) factionQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.FactionID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Factions(queryMods...)
	queries.SetFrom(query.Query, "\"factions\"")

	return query
}

// MysteryCrateBlueprints retrieves all the mystery_crate_blueprint's MysteryCrateBlueprints with an executor.
func (o *MysteryCrate) MysteryCrateBlueprints(mods ...qm.QueryMod) mysteryCrateBlueprintQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"mystery_crate_blueprints\".\"mystery_crate_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"mystery_crate_blueprints\".\"deleted_at\""),
	)

	query := MysteryCrateBlueprints(queryMods...)
	queries.SetFrom(query.Query, "\"mystery_crate_blueprints\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"mystery_crate_blueprints\".*"})
	}

	return query
}

// LoadFaction allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (mysteryCrateL) LoadFaction(e boil.Executor, singular bool, maybeMysteryCrate interface{}, mods queries.Applicator) error {
	var slice []*MysteryCrate
	var object *MysteryCrate

	if singular {
		object = maybeMysteryCrate.(*MysteryCrate)
	} else {
		slice = *maybeMysteryCrate.(*[]*MysteryCrate)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &mysteryCrateR{}
		}
		args = append(args, object.FactionID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &mysteryCrateR{}
			}

			for _, a := range args {
				if a == obj.FactionID {
					continue Outer
				}
			}

			args = append(args, obj.FactionID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`factions`),
		qm.WhereIn(`factions.id in ?`, args...),
		qmhelper.WhereIsNull(`factions.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Faction")
	}

	var resultSlice []*Faction
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Faction")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for factions")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for factions")
	}

	if len(mysteryCrateAfterSelectHooks) != 0 {
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
		object.R.Faction = foreign
		if foreign.R == nil {
			foreign.R = &factionR{}
		}
		foreign.R.MysteryCrates = append(foreign.R.MysteryCrates, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.FactionID == foreign.ID {
				local.R.Faction = foreign
				if foreign.R == nil {
					foreign.R = &factionR{}
				}
				foreign.R.MysteryCrates = append(foreign.R.MysteryCrates, local)
				break
			}
		}
	}

	return nil
}

// LoadMysteryCrateBlueprints allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (mysteryCrateL) LoadMysteryCrateBlueprints(e boil.Executor, singular bool, maybeMysteryCrate interface{}, mods queries.Applicator) error {
	var slice []*MysteryCrate
	var object *MysteryCrate

	if singular {
		object = maybeMysteryCrate.(*MysteryCrate)
	} else {
		slice = *maybeMysteryCrate.(*[]*MysteryCrate)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &mysteryCrateR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &mysteryCrateR{}
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
		qm.From(`mystery_crate_blueprints`),
		qm.WhereIn(`mystery_crate_blueprints.mystery_crate_id in ?`, args...),
		qmhelper.WhereIsNull(`mystery_crate_blueprints.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load mystery_crate_blueprints")
	}

	var resultSlice []*MysteryCrateBlueprint
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice mystery_crate_blueprints")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on mystery_crate_blueprints")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for mystery_crate_blueprints")
	}

	if len(mysteryCrateBlueprintAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.MysteryCrateBlueprints = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &mysteryCrateBlueprintR{}
			}
			foreign.R.MysteryCrate = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.MysteryCrateID {
				local.R.MysteryCrateBlueprints = append(local.R.MysteryCrateBlueprints, foreign)
				if foreign.R == nil {
					foreign.R = &mysteryCrateBlueprintR{}
				}
				foreign.R.MysteryCrate = local
				break
			}
		}
	}

	return nil
}

// SetFaction of the mysteryCrate to the related item.
// Sets o.R.Faction to related.
// Adds o to related.R.MysteryCrates.
func (o *MysteryCrate) SetFaction(exec boil.Executor, insert bool, related *Faction) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"mystery_crate\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"faction_id"}),
		strmangle.WhereClause("\"", "\"", 2, mysteryCratePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.FactionID = related.ID
	if o.R == nil {
		o.R = &mysteryCrateR{
			Faction: related,
		}
	} else {
		o.R.Faction = related
	}

	if related.R == nil {
		related.R = &factionR{
			MysteryCrates: MysteryCrateSlice{o},
		}
	} else {
		related.R.MysteryCrates = append(related.R.MysteryCrates, o)
	}

	return nil
}

// AddMysteryCrateBlueprints adds the given related objects to the existing relationships
// of the mystery_crate, optionally inserting them as new records.
// Appends related to o.R.MysteryCrateBlueprints.
// Sets related.R.MysteryCrate appropriately.
func (o *MysteryCrate) AddMysteryCrateBlueprints(exec boil.Executor, insert bool, related ...*MysteryCrateBlueprint) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.MysteryCrateID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"mystery_crate_blueprints\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"mystery_crate_id"}),
				strmangle.WhereClause("\"", "\"", 2, mysteryCrateBlueprintPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.MysteryCrateID = o.ID
		}
	}

	if o.R == nil {
		o.R = &mysteryCrateR{
			MysteryCrateBlueprints: related,
		}
	} else {
		o.R.MysteryCrateBlueprints = append(o.R.MysteryCrateBlueprints, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &mysteryCrateBlueprintR{
				MysteryCrate: o,
			}
		} else {
			rel.R.MysteryCrate = o
		}
	}
	return nil
}

// MysteryCrates retrieves all the records using an executor.
func MysteryCrates(mods ...qm.QueryMod) mysteryCrateQuery {
	mods = append(mods, qm.From("\"mystery_crate\""), qmhelper.WhereIsNull("\"mystery_crate\".\"deleted_at\""))
	return mysteryCrateQuery{NewQuery(mods...)}
}

// FindMysteryCrate retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindMysteryCrate(exec boil.Executor, iD string, selectCols ...string) (*MysteryCrate, error) {
	mysteryCrateObj := &MysteryCrate{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"mystery_crate\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, mysteryCrateObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from mystery_crate")
	}

	if err = mysteryCrateObj.doAfterSelectHooks(exec); err != nil {
		return mysteryCrateObj, err
	}

	return mysteryCrateObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *MysteryCrate) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no mystery_crate provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = currTime
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(mysteryCrateColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	mysteryCrateInsertCacheMut.RLock()
	cache, cached := mysteryCrateInsertCache[key]
	mysteryCrateInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			mysteryCrateAllColumns,
			mysteryCrateColumnsWithDefault,
			mysteryCrateColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(mysteryCrateType, mysteryCrateMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(mysteryCrateType, mysteryCrateMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"mystery_crate\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"mystery_crate\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into mystery_crate")
	}

	if !cached {
		mysteryCrateInsertCacheMut.Lock()
		mysteryCrateInsertCache[key] = cache
		mysteryCrateInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the MysteryCrate.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *MysteryCrate) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	mysteryCrateUpdateCacheMut.RLock()
	cache, cached := mysteryCrateUpdateCache[key]
	mysteryCrateUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			mysteryCrateAllColumns,
			mysteryCratePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update mystery_crate, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"mystery_crate\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, mysteryCratePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(mysteryCrateType, mysteryCrateMapping, append(wl, mysteryCratePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update mystery_crate row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for mystery_crate")
	}

	if !cached {
		mysteryCrateUpdateCacheMut.Lock()
		mysteryCrateUpdateCache[key] = cache
		mysteryCrateUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q mysteryCrateQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for mystery_crate")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for mystery_crate")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o MysteryCrateSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), mysteryCratePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"mystery_crate\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, mysteryCratePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in mysteryCrate slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all mysteryCrate")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *MysteryCrate) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no mystery_crate provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(mysteryCrateColumnsWithDefault, o)

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

	mysteryCrateUpsertCacheMut.RLock()
	cache, cached := mysteryCrateUpsertCache[key]
	mysteryCrateUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			mysteryCrateAllColumns,
			mysteryCrateColumnsWithDefault,
			mysteryCrateColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			mysteryCrateAllColumns,
			mysteryCratePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert mystery_crate, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(mysteryCratePrimaryKeyColumns))
			copy(conflict, mysteryCratePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"mystery_crate\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(mysteryCrateType, mysteryCrateMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(mysteryCrateType, mysteryCrateMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert mystery_crate")
	}

	if !cached {
		mysteryCrateUpsertCacheMut.Lock()
		mysteryCrateUpsertCache[key] = cache
		mysteryCrateUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single MysteryCrate record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *MysteryCrate) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no MysteryCrate provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), mysteryCratePrimaryKeyMapping)
		sql = "DELETE FROM \"mystery_crate\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"mystery_crate\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(mysteryCrateType, mysteryCrateMapping, append(wl, mysteryCratePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from mystery_crate")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for mystery_crate")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q mysteryCrateQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no mysteryCrateQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from mystery_crate")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for mystery_crate")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o MysteryCrateSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(mysteryCrateBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), mysteryCratePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"mystery_crate\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, mysteryCratePrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), mysteryCratePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"mystery_crate\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, mysteryCratePrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from mysteryCrate slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for mystery_crate")
	}

	if len(mysteryCrateAfterDeleteHooks) != 0 {
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
func (o *MysteryCrate) Reload(exec boil.Executor) error {
	ret, err := FindMysteryCrate(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *MysteryCrateSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := MysteryCrateSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), mysteryCratePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"mystery_crate\".* FROM \"mystery_crate\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, mysteryCratePrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in MysteryCrateSlice")
	}

	*o = slice

	return nil
}

// MysteryCrateExists checks if the MysteryCrate row exists.
func MysteryCrateExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"mystery_crate\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if mystery_crate exists")
	}

	return exists, nil
}
