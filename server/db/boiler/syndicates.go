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

// Syndicate is an object representing the database table.
type Syndicate struct {
	ID          string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Label       string      `boiler:"label" boil:"label" json:"label" toml:"label" yaml:"label"`
	Description string      `boiler:"description" boil:"description" json:"description" toml:"description" yaml:"description"`
	GuildID     null.String `boiler:"guild_id" boil:"guild_id" json:"guildID,omitempty" toml:"guildID" yaml:"guildID,omitempty"`
	DeletedAt   null.Time   `boiler:"deleted_at" boil:"deleted_at" json:"deletedAt,omitempty" toml:"deletedAt" yaml:"deletedAt,omitempty"`
	UpdatedAt   time.Time   `boiler:"updated_at" boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
	CreatedAt   time.Time   `boiler:"created_at" boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`

	R *syndicateR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L syndicateL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var SyndicateColumns = struct {
	ID          string
	Label       string
	Description string
	GuildID     string
	DeletedAt   string
	UpdatedAt   string
	CreatedAt   string
}{
	ID:          "id",
	Label:       "label",
	Description: "description",
	GuildID:     "guild_id",
	DeletedAt:   "deleted_at",
	UpdatedAt:   "updated_at",
	CreatedAt:   "created_at",
}

var SyndicateTableColumns = struct {
	ID          string
	Label       string
	Description string
	GuildID     string
	DeletedAt   string
	UpdatedAt   string
	CreatedAt   string
}{
	ID:          "syndicates.id",
	Label:       "syndicates.label",
	Description: "syndicates.description",
	GuildID:     "syndicates.guild_id",
	DeletedAt:   "syndicates.deleted_at",
	UpdatedAt:   "syndicates.updated_at",
	CreatedAt:   "syndicates.created_at",
}

// Generated where

var SyndicateWhere = struct {
	ID          whereHelperstring
	Label       whereHelperstring
	Description whereHelperstring
	GuildID     whereHelpernull_String
	DeletedAt   whereHelpernull_Time
	UpdatedAt   whereHelpertime_Time
	CreatedAt   whereHelpertime_Time
}{
	ID:          whereHelperstring{field: "\"syndicates\".\"id\""},
	Label:       whereHelperstring{field: "\"syndicates\".\"label\""},
	Description: whereHelperstring{field: "\"syndicates\".\"description\""},
	GuildID:     whereHelpernull_String{field: "\"syndicates\".\"guild_id\""},
	DeletedAt:   whereHelpernull_Time{field: "\"syndicates\".\"deleted_at\""},
	UpdatedAt:   whereHelpertime_Time{field: "\"syndicates\".\"updated_at\""},
	CreatedAt:   whereHelpertime_Time{field: "\"syndicates\".\"created_at\""},
}

// SyndicateRels is where relationship names are stored.
var SyndicateRels = struct {
	Brands  string
	Players string
}{
	Brands:  "Brands",
	Players: "Players",
}

// syndicateR is where relationships are stored.
type syndicateR struct {
	Brands  BrandSlice  `boiler:"Brands" boil:"Brands" json:"Brands" toml:"Brands" yaml:"Brands"`
	Players PlayerSlice `boiler:"Players" boil:"Players" json:"Players" toml:"Players" yaml:"Players"`
}

// NewStruct creates a new relationship struct
func (*syndicateR) NewStruct() *syndicateR {
	return &syndicateR{}
}

// syndicateL is where Load methods for each relationship are stored.
type syndicateL struct{}

var (
	syndicateAllColumns            = []string{"id", "label", "description", "guild_id", "deleted_at", "updated_at", "created_at"}
	syndicateColumnsWithoutDefault = []string{"label", "description"}
	syndicateColumnsWithDefault    = []string{"id", "guild_id", "deleted_at", "updated_at", "created_at"}
	syndicatePrimaryKeyColumns     = []string{"id"}
	syndicateGeneratedColumns      = []string{}
)

type (
	// SyndicateSlice is an alias for a slice of pointers to Syndicate.
	// This should almost always be used instead of []Syndicate.
	SyndicateSlice []*Syndicate
	// SyndicateHook is the signature for custom Syndicate hook methods
	SyndicateHook func(boil.Executor, *Syndicate) error

	syndicateQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	syndicateType                 = reflect.TypeOf(&Syndicate{})
	syndicateMapping              = queries.MakeStructMapping(syndicateType)
	syndicatePrimaryKeyMapping, _ = queries.BindMapping(syndicateType, syndicateMapping, syndicatePrimaryKeyColumns)
	syndicateInsertCacheMut       sync.RWMutex
	syndicateInsertCache          = make(map[string]insertCache)
	syndicateUpdateCacheMut       sync.RWMutex
	syndicateUpdateCache          = make(map[string]updateCache)
	syndicateUpsertCacheMut       sync.RWMutex
	syndicateUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var syndicateAfterSelectHooks []SyndicateHook

var syndicateBeforeInsertHooks []SyndicateHook
var syndicateAfterInsertHooks []SyndicateHook

var syndicateBeforeUpdateHooks []SyndicateHook
var syndicateAfterUpdateHooks []SyndicateHook

var syndicateBeforeDeleteHooks []SyndicateHook
var syndicateAfterDeleteHooks []SyndicateHook

var syndicateBeforeUpsertHooks []SyndicateHook
var syndicateAfterUpsertHooks []SyndicateHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Syndicate) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range syndicateAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Syndicate) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range syndicateBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Syndicate) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range syndicateAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Syndicate) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range syndicateBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Syndicate) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range syndicateAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Syndicate) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range syndicateBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Syndicate) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range syndicateAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Syndicate) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range syndicateBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Syndicate) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range syndicateAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddSyndicateHook registers your hook function for all future operations.
func AddSyndicateHook(hookPoint boil.HookPoint, syndicateHook SyndicateHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		syndicateAfterSelectHooks = append(syndicateAfterSelectHooks, syndicateHook)
	case boil.BeforeInsertHook:
		syndicateBeforeInsertHooks = append(syndicateBeforeInsertHooks, syndicateHook)
	case boil.AfterInsertHook:
		syndicateAfterInsertHooks = append(syndicateAfterInsertHooks, syndicateHook)
	case boil.BeforeUpdateHook:
		syndicateBeforeUpdateHooks = append(syndicateBeforeUpdateHooks, syndicateHook)
	case boil.AfterUpdateHook:
		syndicateAfterUpdateHooks = append(syndicateAfterUpdateHooks, syndicateHook)
	case boil.BeforeDeleteHook:
		syndicateBeforeDeleteHooks = append(syndicateBeforeDeleteHooks, syndicateHook)
	case boil.AfterDeleteHook:
		syndicateAfterDeleteHooks = append(syndicateAfterDeleteHooks, syndicateHook)
	case boil.BeforeUpsertHook:
		syndicateBeforeUpsertHooks = append(syndicateBeforeUpsertHooks, syndicateHook)
	case boil.AfterUpsertHook:
		syndicateAfterUpsertHooks = append(syndicateAfterUpsertHooks, syndicateHook)
	}
}

// One returns a single syndicate record from the query.
func (q syndicateQuery) One(exec boil.Executor) (*Syndicate, error) {
	o := &Syndicate{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for syndicates")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Syndicate records from the query.
func (q syndicateQuery) All(exec boil.Executor) (SyndicateSlice, error) {
	var o []*Syndicate

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to Syndicate slice")
	}

	if len(syndicateAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Syndicate records in the query.
func (q syndicateQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count syndicates rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q syndicateQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if syndicates exists")
	}

	return count > 0, nil
}

// Brands retrieves all the brand's Brands with an executor.
func (o *Syndicate) Brands(mods ...qm.QueryMod) brandQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"brands\".\"syndicate_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"brands\".\"deleted_at\""),
	)

	query := Brands(queryMods...)
	queries.SetFrom(query.Query, "\"brands\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"brands\".*"})
	}

	return query
}

// Players retrieves all the player's Players with an executor.
func (o *Syndicate) Players(mods ...qm.QueryMod) playerQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"players\".\"syndicate_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"players\".\"deleted_at\""),
	)

	query := Players(queryMods...)
	queries.SetFrom(query.Query, "\"players\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"players\".*"})
	}

	return query
}

// LoadBrands allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (syndicateL) LoadBrands(e boil.Executor, singular bool, maybeSyndicate interface{}, mods queries.Applicator) error {
	var slice []*Syndicate
	var object *Syndicate

	if singular {
		object = maybeSyndicate.(*Syndicate)
	} else {
		slice = *maybeSyndicate.(*[]*Syndicate)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &syndicateR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &syndicateR{}
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
		qm.From(`brands`),
		qm.WhereIn(`brands.syndicate_id in ?`, args...),
		qmhelper.WhereIsNull(`brands.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load brands")
	}

	var resultSlice []*Brand
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice brands")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on brands")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for brands")
	}

	if len(brandAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.Brands = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &brandR{}
			}
			foreign.R.Syndicate = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.SyndicateID {
				local.R.Brands = append(local.R.Brands, foreign)
				if foreign.R == nil {
					foreign.R = &brandR{}
				}
				foreign.R.Syndicate = local
				break
			}
		}
	}

	return nil
}

// LoadPlayers allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (syndicateL) LoadPlayers(e boil.Executor, singular bool, maybeSyndicate interface{}, mods queries.Applicator) error {
	var slice []*Syndicate
	var object *Syndicate

	if singular {
		object = maybeSyndicate.(*Syndicate)
	} else {
		slice = *maybeSyndicate.(*[]*Syndicate)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &syndicateR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &syndicateR{}
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
		qm.From(`players`),
		qm.WhereIn(`players.syndicate_id in ?`, args...),
		qmhelper.WhereIsNull(`players.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load players")
	}

	var resultSlice []*Player
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice players")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on players")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for players")
	}

	if len(playerAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.Players = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &playerR{}
			}
			foreign.R.Syndicate = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.SyndicateID {
				local.R.Players = append(local.R.Players, foreign)
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.Syndicate = local
				break
			}
		}
	}

	return nil
}

// AddBrands adds the given related objects to the existing relationships
// of the syndicate, optionally inserting them as new records.
// Appends related to o.R.Brands.
// Sets related.R.Syndicate appropriately.
func (o *Syndicate) AddBrands(exec boil.Executor, insert bool, related ...*Brand) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.SyndicateID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"brands\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"syndicate_id"}),
				strmangle.WhereClause("\"", "\"", 2, brandPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.SyndicateID = o.ID
		}
	}

	if o.R == nil {
		o.R = &syndicateR{
			Brands: related,
		}
	} else {
		o.R.Brands = append(o.R.Brands, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &brandR{
				Syndicate: o,
			}
		} else {
			rel.R.Syndicate = o
		}
	}
	return nil
}

// AddPlayers adds the given related objects to the existing relationships
// of the syndicate, optionally inserting them as new records.
// Appends related to o.R.Players.
// Sets related.R.Syndicate appropriately.
func (o *Syndicate) AddPlayers(exec boil.Executor, insert bool, related ...*Player) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.SyndicateID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"players\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"syndicate_id"}),
				strmangle.WhereClause("\"", "\"", 2, playerPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.SyndicateID = o.ID
		}
	}

	if o.R == nil {
		o.R = &syndicateR{
			Players: related,
		}
	} else {
		o.R.Players = append(o.R.Players, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &playerR{
				Syndicate: o,
			}
		} else {
			rel.R.Syndicate = o
		}
	}
	return nil
}

// Syndicates retrieves all the records using an executor.
func Syndicates(mods ...qm.QueryMod) syndicateQuery {
	mods = append(mods, qm.From("\"syndicates\""), qmhelper.WhereIsNull("\"syndicates\".\"deleted_at\""))
	return syndicateQuery{NewQuery(mods...)}
}

// FindSyndicate retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindSyndicate(exec boil.Executor, iD string, selectCols ...string) (*Syndicate, error) {
	syndicateObj := &Syndicate{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"syndicates\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, syndicateObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from syndicates")
	}

	if err = syndicateObj.doAfterSelectHooks(exec); err != nil {
		return syndicateObj, err
	}

	return syndicateObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Syndicate) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no syndicates provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(syndicateColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	syndicateInsertCacheMut.RLock()
	cache, cached := syndicateInsertCache[key]
	syndicateInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			syndicateAllColumns,
			syndicateColumnsWithDefault,
			syndicateColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(syndicateType, syndicateMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(syndicateType, syndicateMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"syndicates\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"syndicates\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into syndicates")
	}

	if !cached {
		syndicateInsertCacheMut.Lock()
		syndicateInsertCache[key] = cache
		syndicateInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the Syndicate.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Syndicate) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	syndicateUpdateCacheMut.RLock()
	cache, cached := syndicateUpdateCache[key]
	syndicateUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			syndicateAllColumns,
			syndicatePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update syndicates, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"syndicates\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, syndicatePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(syndicateType, syndicateMapping, append(wl, syndicatePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update syndicates row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for syndicates")
	}

	if !cached {
		syndicateUpdateCacheMut.Lock()
		syndicateUpdateCache[key] = cache
		syndicateUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q syndicateQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for syndicates")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for syndicates")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o SyndicateSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), syndicatePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"syndicates\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, syndicatePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in syndicate slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all syndicate")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Syndicate) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no syndicates provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(syndicateColumnsWithDefault, o)

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

	syndicateUpsertCacheMut.RLock()
	cache, cached := syndicateUpsertCache[key]
	syndicateUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			syndicateAllColumns,
			syndicateColumnsWithDefault,
			syndicateColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			syndicateAllColumns,
			syndicatePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert syndicates, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(syndicatePrimaryKeyColumns))
			copy(conflict, syndicatePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"syndicates\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(syndicateType, syndicateMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(syndicateType, syndicateMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert syndicates")
	}

	if !cached {
		syndicateUpsertCacheMut.Lock()
		syndicateUpsertCache[key] = cache
		syndicateUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single Syndicate record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Syndicate) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no Syndicate provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), syndicatePrimaryKeyMapping)
		sql = "DELETE FROM \"syndicates\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"syndicates\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(syndicateType, syndicateMapping, append(wl, syndicatePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from syndicates")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for syndicates")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q syndicateQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no syndicateQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from syndicates")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for syndicates")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o SyndicateSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(syndicateBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), syndicatePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"syndicates\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, syndicatePrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), syndicatePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"syndicates\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, syndicatePrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from syndicate slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for syndicates")
	}

	if len(syndicateAfterDeleteHooks) != 0 {
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
func (o *Syndicate) Reload(exec boil.Executor) error {
	ret, err := FindSyndicate(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *SyndicateSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := SyndicateSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), syndicatePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"syndicates\".* FROM \"syndicates\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, syndicatePrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in SyndicateSlice")
	}

	*o = slice

	return nil
}

// SyndicateExists checks if the Syndicate row exists.
func SyndicateExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"syndicates\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if syndicates exists")
	}

	return exists, nil
}
