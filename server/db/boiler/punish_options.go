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

// PunishOption is an object representing the database table.
type PunishOption struct {
	ID                  string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Description         string    `boiler:"description" boil:"description" json:"description" toml:"description" yaml:"description"`
	Key                 string    `boiler:"key" boil:"key" json:"key" toml:"key" yaml:"key"`
	PunishDurationHours int       `boiler:"punish_duration_hours" boil:"punish_duration_hours" json:"punish_duration_hours" toml:"punish_duration_hours" yaml:"punish_duration_hours"`
	CreatedAt           time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt           time.Time `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	DeletedAt           null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`

	R *punishOptionR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L punishOptionL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PunishOptionColumns = struct {
	ID                  string
	Description         string
	Key                 string
	PunishDurationHours string
	CreatedAt           string
	UpdatedAt           string
	DeletedAt           string
}{
	ID:                  "id",
	Description:         "description",
	Key:                 "key",
	PunishDurationHours: "punish_duration_hours",
	CreatedAt:           "created_at",
	UpdatedAt:           "updated_at",
	DeletedAt:           "deleted_at",
}

var PunishOptionTableColumns = struct {
	ID                  string
	Description         string
	Key                 string
	PunishDurationHours string
	CreatedAt           string
	UpdatedAt           string
	DeletedAt           string
}{
	ID:                  "punish_options.id",
	Description:         "punish_options.description",
	Key:                 "punish_options.key",
	PunishDurationHours: "punish_options.punish_duration_hours",
	CreatedAt:           "punish_options.created_at",
	UpdatedAt:           "punish_options.updated_at",
	DeletedAt:           "punish_options.deleted_at",
}

// Generated where

var PunishOptionWhere = struct {
	ID                  whereHelperstring
	Description         whereHelperstring
	Key                 whereHelperstring
	PunishDurationHours whereHelperint
	CreatedAt           whereHelpertime_Time
	UpdatedAt           whereHelpertime_Time
	DeletedAt           whereHelpernull_Time
}{
	ID:                  whereHelperstring{field: "\"punish_options\".\"id\""},
	Description:         whereHelperstring{field: "\"punish_options\".\"description\""},
	Key:                 whereHelperstring{field: "\"punish_options\".\"key\""},
	PunishDurationHours: whereHelperint{field: "\"punish_options\".\"punish_duration_hours\""},
	CreatedAt:           whereHelpertime_Time{field: "\"punish_options\".\"created_at\""},
	UpdatedAt:           whereHelpertime_Time{field: "\"punish_options\".\"updated_at\""},
	DeletedAt:           whereHelpernull_Time{field: "\"punish_options\".\"deleted_at\""},
}

// PunishOptionRels is where relationship names are stored.
var PunishOptionRels = struct {
	PunishVotes     string
	PunishedPlayers string
}{
	PunishVotes:     "PunishVotes",
	PunishedPlayers: "PunishedPlayers",
}

// punishOptionR is where relationships are stored.
type punishOptionR struct {
	PunishVotes     PunishVoteSlice     `boiler:"PunishVotes" boil:"PunishVotes" json:"PunishVotes" toml:"PunishVotes" yaml:"PunishVotes"`
	PunishedPlayers PunishedPlayerSlice `boiler:"PunishedPlayers" boil:"PunishedPlayers" json:"PunishedPlayers" toml:"PunishedPlayers" yaml:"PunishedPlayers"`
}

// NewStruct creates a new relationship struct
func (*punishOptionR) NewStruct() *punishOptionR {
	return &punishOptionR{}
}

// punishOptionL is where Load methods for each relationship are stored.
type punishOptionL struct{}

var (
	punishOptionAllColumns            = []string{"id", "description", "key", "punish_duration_hours", "created_at", "updated_at", "deleted_at"}
	punishOptionColumnsWithoutDefault = []string{"description", "key"}
	punishOptionColumnsWithDefault    = []string{"id", "punish_duration_hours", "created_at", "updated_at", "deleted_at"}
	punishOptionPrimaryKeyColumns     = []string{"id"}
	punishOptionGeneratedColumns      = []string{}
)

type (
	// PunishOptionSlice is an alias for a slice of pointers to PunishOption.
	// This should almost always be used instead of []PunishOption.
	PunishOptionSlice []*PunishOption
	// PunishOptionHook is the signature for custom PunishOption hook methods
	PunishOptionHook func(boil.Executor, *PunishOption) error

	punishOptionQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	punishOptionType                 = reflect.TypeOf(&PunishOption{})
	punishOptionMapping              = queries.MakeStructMapping(punishOptionType)
	punishOptionPrimaryKeyMapping, _ = queries.BindMapping(punishOptionType, punishOptionMapping, punishOptionPrimaryKeyColumns)
	punishOptionInsertCacheMut       sync.RWMutex
	punishOptionInsertCache          = make(map[string]insertCache)
	punishOptionUpdateCacheMut       sync.RWMutex
	punishOptionUpdateCache          = make(map[string]updateCache)
	punishOptionUpsertCacheMut       sync.RWMutex
	punishOptionUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var punishOptionAfterSelectHooks []PunishOptionHook

var punishOptionBeforeInsertHooks []PunishOptionHook
var punishOptionAfterInsertHooks []PunishOptionHook

var punishOptionBeforeUpdateHooks []PunishOptionHook
var punishOptionAfterUpdateHooks []PunishOptionHook

var punishOptionBeforeDeleteHooks []PunishOptionHook
var punishOptionAfterDeleteHooks []PunishOptionHook

var punishOptionBeforeUpsertHooks []PunishOptionHook
var punishOptionAfterUpsertHooks []PunishOptionHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *PunishOption) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range punishOptionAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *PunishOption) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range punishOptionBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *PunishOption) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range punishOptionAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *PunishOption) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range punishOptionBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *PunishOption) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range punishOptionAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *PunishOption) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range punishOptionBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *PunishOption) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range punishOptionAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *PunishOption) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range punishOptionBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *PunishOption) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range punishOptionAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPunishOptionHook registers your hook function for all future operations.
func AddPunishOptionHook(hookPoint boil.HookPoint, punishOptionHook PunishOptionHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		punishOptionAfterSelectHooks = append(punishOptionAfterSelectHooks, punishOptionHook)
	case boil.BeforeInsertHook:
		punishOptionBeforeInsertHooks = append(punishOptionBeforeInsertHooks, punishOptionHook)
	case boil.AfterInsertHook:
		punishOptionAfterInsertHooks = append(punishOptionAfterInsertHooks, punishOptionHook)
	case boil.BeforeUpdateHook:
		punishOptionBeforeUpdateHooks = append(punishOptionBeforeUpdateHooks, punishOptionHook)
	case boil.AfterUpdateHook:
		punishOptionAfterUpdateHooks = append(punishOptionAfterUpdateHooks, punishOptionHook)
	case boil.BeforeDeleteHook:
		punishOptionBeforeDeleteHooks = append(punishOptionBeforeDeleteHooks, punishOptionHook)
	case boil.AfterDeleteHook:
		punishOptionAfterDeleteHooks = append(punishOptionAfterDeleteHooks, punishOptionHook)
	case boil.BeforeUpsertHook:
		punishOptionBeforeUpsertHooks = append(punishOptionBeforeUpsertHooks, punishOptionHook)
	case boil.AfterUpsertHook:
		punishOptionAfterUpsertHooks = append(punishOptionAfterUpsertHooks, punishOptionHook)
	}
}

// One returns a single punishOption record from the query.
func (q punishOptionQuery) One(exec boil.Executor) (*PunishOption, error) {
	o := &PunishOption{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for punish_options")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all PunishOption records from the query.
func (q punishOptionQuery) All(exec boil.Executor) (PunishOptionSlice, error) {
	var o []*PunishOption

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to PunishOption slice")
	}

	if len(punishOptionAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all PunishOption records in the query.
func (q punishOptionQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count punish_options rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q punishOptionQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if punish_options exists")
	}

	return count > 0, nil
}

// PunishVotes retrieves all the punish_vote's PunishVotes with an executor.
func (o *PunishOption) PunishVotes(mods ...qm.QueryMod) punishVoteQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"punish_votes\".\"punish_option_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"punish_votes\".\"deleted_at\""),
	)

	query := PunishVotes(queryMods...)
	queries.SetFrom(query.Query, "\"punish_votes\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"punish_votes\".*"})
	}

	return query
}

// PunishedPlayers retrieves all the punished_player's PunishedPlayers with an executor.
func (o *PunishOption) PunishedPlayers(mods ...qm.QueryMod) punishedPlayerQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"punished_players\".\"punish_option_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"punished_players\".\"deleted_at\""),
	)

	query := PunishedPlayers(queryMods...)
	queries.SetFrom(query.Query, "\"punished_players\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"punished_players\".*"})
	}

	return query
}

// LoadPunishVotes allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (punishOptionL) LoadPunishVotes(e boil.Executor, singular bool, maybePunishOption interface{}, mods queries.Applicator) error {
	var slice []*PunishOption
	var object *PunishOption

	if singular {
		object = maybePunishOption.(*PunishOption)
	} else {
		slice = *maybePunishOption.(*[]*PunishOption)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &punishOptionR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &punishOptionR{}
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
		qm.From(`punish_votes`),
		qm.WhereIn(`punish_votes.punish_option_id in ?`, args...),
		qmhelper.WhereIsNull(`punish_votes.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load punish_votes")
	}

	var resultSlice []*PunishVote
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice punish_votes")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on punish_votes")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for punish_votes")
	}

	if len(punishVoteAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.PunishVotes = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &punishVoteR{}
			}
			foreign.R.PunishOption = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.PunishOptionID {
				local.R.PunishVotes = append(local.R.PunishVotes, foreign)
				if foreign.R == nil {
					foreign.R = &punishVoteR{}
				}
				foreign.R.PunishOption = local
				break
			}
		}
	}

	return nil
}

// LoadPunishedPlayers allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (punishOptionL) LoadPunishedPlayers(e boil.Executor, singular bool, maybePunishOption interface{}, mods queries.Applicator) error {
	var slice []*PunishOption
	var object *PunishOption

	if singular {
		object = maybePunishOption.(*PunishOption)
	} else {
		slice = *maybePunishOption.(*[]*PunishOption)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &punishOptionR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &punishOptionR{}
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
		qm.From(`punished_players`),
		qm.WhereIn(`punished_players.punish_option_id in ?`, args...),
		qmhelper.WhereIsNull(`punished_players.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load punished_players")
	}

	var resultSlice []*PunishedPlayer
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice punished_players")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on punished_players")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for punished_players")
	}

	if len(punishedPlayerAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.PunishedPlayers = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &punishedPlayerR{}
			}
			foreign.R.PunishOption = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.PunishOptionID {
				local.R.PunishedPlayers = append(local.R.PunishedPlayers, foreign)
				if foreign.R == nil {
					foreign.R = &punishedPlayerR{}
				}
				foreign.R.PunishOption = local
				break
			}
		}
	}

	return nil
}

// AddPunishVotes adds the given related objects to the existing relationships
// of the punish_option, optionally inserting them as new records.
// Appends related to o.R.PunishVotes.
// Sets related.R.PunishOption appropriately.
func (o *PunishOption) AddPunishVotes(exec boil.Executor, insert bool, related ...*PunishVote) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.PunishOptionID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"punish_votes\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"punish_option_id"}),
				strmangle.WhereClause("\"", "\"", 2, punishVotePrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.PunishOptionID = o.ID
		}
	}

	if o.R == nil {
		o.R = &punishOptionR{
			PunishVotes: related,
		}
	} else {
		o.R.PunishVotes = append(o.R.PunishVotes, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &punishVoteR{
				PunishOption: o,
			}
		} else {
			rel.R.PunishOption = o
		}
	}
	return nil
}

// AddPunishedPlayers adds the given related objects to the existing relationships
// of the punish_option, optionally inserting them as new records.
// Appends related to o.R.PunishedPlayers.
// Sets related.R.PunishOption appropriately.
func (o *PunishOption) AddPunishedPlayers(exec boil.Executor, insert bool, related ...*PunishedPlayer) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.PunishOptionID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"punished_players\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"punish_option_id"}),
				strmangle.WhereClause("\"", "\"", 2, punishedPlayerPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.PunishOptionID = o.ID
		}
	}

	if o.R == nil {
		o.R = &punishOptionR{
			PunishedPlayers: related,
		}
	} else {
		o.R.PunishedPlayers = append(o.R.PunishedPlayers, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &punishedPlayerR{
				PunishOption: o,
			}
		} else {
			rel.R.PunishOption = o
		}
	}
	return nil
}

// PunishOptions retrieves all the records using an executor.
func PunishOptions(mods ...qm.QueryMod) punishOptionQuery {
	mods = append(mods, qm.From("\"punish_options\""), qmhelper.WhereIsNull("\"punish_options\".\"deleted_at\""))
	return punishOptionQuery{NewQuery(mods...)}
}

// FindPunishOption retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPunishOption(exec boil.Executor, iD string, selectCols ...string) (*PunishOption, error) {
	punishOptionObj := &PunishOption{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"punish_options\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, punishOptionObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from punish_options")
	}

	if err = punishOptionObj.doAfterSelectHooks(exec); err != nil {
		return punishOptionObj, err
	}

	return punishOptionObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *PunishOption) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no punish_options provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(punishOptionColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	punishOptionInsertCacheMut.RLock()
	cache, cached := punishOptionInsertCache[key]
	punishOptionInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			punishOptionAllColumns,
			punishOptionColumnsWithDefault,
			punishOptionColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(punishOptionType, punishOptionMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(punishOptionType, punishOptionMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"punish_options\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"punish_options\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into punish_options")
	}

	if !cached {
		punishOptionInsertCacheMut.Lock()
		punishOptionInsertCache[key] = cache
		punishOptionInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the PunishOption.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *PunishOption) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	punishOptionUpdateCacheMut.RLock()
	cache, cached := punishOptionUpdateCache[key]
	punishOptionUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			punishOptionAllColumns,
			punishOptionPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update punish_options, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"punish_options\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, punishOptionPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(punishOptionType, punishOptionMapping, append(wl, punishOptionPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update punish_options row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for punish_options")
	}

	if !cached {
		punishOptionUpdateCacheMut.Lock()
		punishOptionUpdateCache[key] = cache
		punishOptionUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q punishOptionQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for punish_options")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for punish_options")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PunishOptionSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), punishOptionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"punish_options\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, punishOptionPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in punishOption slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all punishOption")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *PunishOption) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no punish_options provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}
	o.UpdatedAt = currTime

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(punishOptionColumnsWithDefault, o)

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

	punishOptionUpsertCacheMut.RLock()
	cache, cached := punishOptionUpsertCache[key]
	punishOptionUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			punishOptionAllColumns,
			punishOptionColumnsWithDefault,
			punishOptionColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			punishOptionAllColumns,
			punishOptionPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert punish_options, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(punishOptionPrimaryKeyColumns))
			copy(conflict, punishOptionPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"punish_options\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(punishOptionType, punishOptionMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(punishOptionType, punishOptionMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert punish_options")
	}

	if !cached {
		punishOptionUpsertCacheMut.Lock()
		punishOptionUpsertCache[key] = cache
		punishOptionUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single PunishOption record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *PunishOption) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no PunishOption provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), punishOptionPrimaryKeyMapping)
		sql = "DELETE FROM \"punish_options\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"punish_options\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(punishOptionType, punishOptionMapping, append(wl, punishOptionPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from punish_options")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for punish_options")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q punishOptionQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no punishOptionQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from punish_options")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for punish_options")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PunishOptionSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(punishOptionBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), punishOptionPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"punish_options\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, punishOptionPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), punishOptionPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"punish_options\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, punishOptionPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from punishOption slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for punish_options")
	}

	if len(punishOptionAfterDeleteHooks) != 0 {
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
func (o *PunishOption) Reload(exec boil.Executor) error {
	ret, err := FindPunishOption(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PunishOptionSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PunishOptionSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), punishOptionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"punish_options\".* FROM \"punish_options\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, punishOptionPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in PunishOptionSlice")
	}

	*o = slice

	return nil
}

// PunishOptionExists checks if the PunishOption row exists.
func PunishOptionExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"punish_options\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if punish_options exists")
	}

	return exists, nil
}
