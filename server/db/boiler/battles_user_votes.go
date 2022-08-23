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

// BattlesUserVote is an object representing the database table.
type BattlesUserVote struct {
	BattleID  string `boiler:"battle_id" boil:"battle_id" json:"battle_id" toml:"battle_id" yaml:"battle_id"`
	UserID    string `boiler:"user_id" boil:"user_id" json:"user_id" toml:"user_id" yaml:"user_id"`
	VoteCount int    `boiler:"vote_count" boil:"vote_count" json:"vote_count" toml:"vote_count" yaml:"vote_count"`

	R *battlesUserVoteR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L battlesUserVoteL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var BattlesUserVoteColumns = struct {
	BattleID  string
	UserID    string
	VoteCount string
}{
	BattleID:  "battle_id",
	UserID:    "user_id",
	VoteCount: "vote_count",
}

var BattlesUserVoteTableColumns = struct {
	BattleID  string
	UserID    string
	VoteCount string
}{
	BattleID:  "battles_user_votes.battle_id",
	UserID:    "battles_user_votes.user_id",
	VoteCount: "battles_user_votes.vote_count",
}

// Generated where

var BattlesUserVoteWhere = struct {
	BattleID  whereHelperstring
	UserID    whereHelperstring
	VoteCount whereHelperint
}{
	BattleID:  whereHelperstring{field: "\"battles_user_votes\".\"battle_id\""},
	UserID:    whereHelperstring{field: "\"battles_user_votes\".\"user_id\""},
	VoteCount: whereHelperint{field: "\"battles_user_votes\".\"vote_count\""},
}

// BattlesUserVoteRels is where relationship names are stored.
var BattlesUserVoteRels = struct {
	Battle string
	User   string
}{
	Battle: "Battle",
	User:   "User",
}

// battlesUserVoteR is where relationships are stored.
type battlesUserVoteR struct {
	Battle *Battle `boiler:"Battle" boil:"Battle" json:"Battle" toml:"Battle" yaml:"Battle"`
	User   *User   `boiler:"User" boil:"User" json:"User" toml:"User" yaml:"User"`
}

// NewStruct creates a new relationship struct
func (*battlesUserVoteR) NewStruct() *battlesUserVoteR {
	return &battlesUserVoteR{}
}

// battlesUserVoteL is where Load methods for each relationship are stored.
type battlesUserVoteL struct{}

var (
	battlesUserVoteAllColumns            = []string{"battle_id", "user_id", "vote_count"}
	battlesUserVoteColumnsWithoutDefault = []string{"battle_id", "user_id"}
	battlesUserVoteColumnsWithDefault    = []string{"vote_count"}
	battlesUserVotePrimaryKeyColumns     = []string{"battle_id", "user_id"}
	battlesUserVoteGeneratedColumns      = []string{}
)

type (
	// BattlesUserVoteSlice is an alias for a slice of pointers to BattlesUserVote.
	// This should almost always be used instead of []BattlesUserVote.
	BattlesUserVoteSlice []*BattlesUserVote
	// BattlesUserVoteHook is the signature for custom BattlesUserVote hook methods
	BattlesUserVoteHook func(boil.Executor, *BattlesUserVote) error

	battlesUserVoteQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	battlesUserVoteType                 = reflect.TypeOf(&BattlesUserVote{})
	battlesUserVoteMapping              = queries.MakeStructMapping(battlesUserVoteType)
	battlesUserVotePrimaryKeyMapping, _ = queries.BindMapping(battlesUserVoteType, battlesUserVoteMapping, battlesUserVotePrimaryKeyColumns)
	battlesUserVoteInsertCacheMut       sync.RWMutex
	battlesUserVoteInsertCache          = make(map[string]insertCache)
	battlesUserVoteUpdateCacheMut       sync.RWMutex
	battlesUserVoteUpdateCache          = make(map[string]updateCache)
	battlesUserVoteUpsertCacheMut       sync.RWMutex
	battlesUserVoteUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var battlesUserVoteAfterSelectHooks []BattlesUserVoteHook

var battlesUserVoteBeforeInsertHooks []BattlesUserVoteHook
var battlesUserVoteAfterInsertHooks []BattlesUserVoteHook

var battlesUserVoteBeforeUpdateHooks []BattlesUserVoteHook
var battlesUserVoteAfterUpdateHooks []BattlesUserVoteHook

var battlesUserVoteBeforeDeleteHooks []BattlesUserVoteHook
var battlesUserVoteAfterDeleteHooks []BattlesUserVoteHook

var battlesUserVoteBeforeUpsertHooks []BattlesUserVoteHook
var battlesUserVoteAfterUpsertHooks []BattlesUserVoteHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *BattlesUserVote) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range battlesUserVoteAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *BattlesUserVote) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battlesUserVoteBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *BattlesUserVote) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battlesUserVoteAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *BattlesUserVote) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range battlesUserVoteBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *BattlesUserVote) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range battlesUserVoteAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *BattlesUserVote) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range battlesUserVoteBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *BattlesUserVote) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range battlesUserVoteAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *BattlesUserVote) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battlesUserVoteBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *BattlesUserVote) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battlesUserVoteAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddBattlesUserVoteHook registers your hook function for all future operations.
func AddBattlesUserVoteHook(hookPoint boil.HookPoint, battlesUserVoteHook BattlesUserVoteHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		battlesUserVoteAfterSelectHooks = append(battlesUserVoteAfterSelectHooks, battlesUserVoteHook)
	case boil.BeforeInsertHook:
		battlesUserVoteBeforeInsertHooks = append(battlesUserVoteBeforeInsertHooks, battlesUserVoteHook)
	case boil.AfterInsertHook:
		battlesUserVoteAfterInsertHooks = append(battlesUserVoteAfterInsertHooks, battlesUserVoteHook)
	case boil.BeforeUpdateHook:
		battlesUserVoteBeforeUpdateHooks = append(battlesUserVoteBeforeUpdateHooks, battlesUserVoteHook)
	case boil.AfterUpdateHook:
		battlesUserVoteAfterUpdateHooks = append(battlesUserVoteAfterUpdateHooks, battlesUserVoteHook)
	case boil.BeforeDeleteHook:
		battlesUserVoteBeforeDeleteHooks = append(battlesUserVoteBeforeDeleteHooks, battlesUserVoteHook)
	case boil.AfterDeleteHook:
		battlesUserVoteAfterDeleteHooks = append(battlesUserVoteAfterDeleteHooks, battlesUserVoteHook)
	case boil.BeforeUpsertHook:
		battlesUserVoteBeforeUpsertHooks = append(battlesUserVoteBeforeUpsertHooks, battlesUserVoteHook)
	case boil.AfterUpsertHook:
		battlesUserVoteAfterUpsertHooks = append(battlesUserVoteAfterUpsertHooks, battlesUserVoteHook)
	}
}

// One returns a single battlesUserVote record from the query.
func (q battlesUserVoteQuery) One(exec boil.Executor) (*BattlesUserVote, error) {
	o := &BattlesUserVote{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for battles_user_votes")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all BattlesUserVote records from the query.
func (q battlesUserVoteQuery) All(exec boil.Executor) (BattlesUserVoteSlice, error) {
	var o []*BattlesUserVote

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to BattlesUserVote slice")
	}

	if len(battlesUserVoteAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all BattlesUserVote records in the query.
func (q battlesUserVoteQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count battles_user_votes rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q battlesUserVoteQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if battles_user_votes exists")
	}

	return count > 0, nil
}

// Battle pointed to by the foreign key.
func (o *BattlesUserVote) Battle(mods ...qm.QueryMod) battleQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BattleID),
	}

	queryMods = append(queryMods, mods...)

	query := Battles(queryMods...)
	queries.SetFrom(query.Query, "\"battles\"")

	return query
}

// User pointed to by the foreign key.
func (o *BattlesUserVote) User(mods ...qm.QueryMod) userQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.UserID),
	}

	queryMods = append(queryMods, mods...)

	query := Users(queryMods...)
	queries.SetFrom(query.Query, "\"users\"")

	return query
}

// LoadBattle allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (battlesUserVoteL) LoadBattle(e boil.Executor, singular bool, maybeBattlesUserVote interface{}, mods queries.Applicator) error {
	var slice []*BattlesUserVote
	var object *BattlesUserVote

	if singular {
		object = maybeBattlesUserVote.(*BattlesUserVote)
	} else {
		slice = *maybeBattlesUserVote.(*[]*BattlesUserVote)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &battlesUserVoteR{}
		}
		args = append(args, object.BattleID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &battlesUserVoteR{}
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

	if len(battlesUserVoteAfterSelectHooks) != 0 {
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
		foreign.R.BattlesUserVotes = append(foreign.R.BattlesUserVotes, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.BattleID == foreign.ID {
				local.R.Battle = foreign
				if foreign.R == nil {
					foreign.R = &battleR{}
				}
				foreign.R.BattlesUserVotes = append(foreign.R.BattlesUserVotes, local)
				break
			}
		}
	}

	return nil
}

// LoadUser allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (battlesUserVoteL) LoadUser(e boil.Executor, singular bool, maybeBattlesUserVote interface{}, mods queries.Applicator) error {
	var slice []*BattlesUserVote
	var object *BattlesUserVote

	if singular {
		object = maybeBattlesUserVote.(*BattlesUserVote)
	} else {
		slice = *maybeBattlesUserVote.(*[]*BattlesUserVote)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &battlesUserVoteR{}
		}
		args = append(args, object.UserID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &battlesUserVoteR{}
			}

			for _, a := range args {
				if a == obj.UserID {
					continue Outer
				}
			}

			args = append(args, obj.UserID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`users`),
		qm.WhereIn(`users.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load User")
	}

	var resultSlice []*User
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice User")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for users")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for users")
	}

	if len(battlesUserVoteAfterSelectHooks) != 0 {
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
		object.R.User = foreign
		if foreign.R == nil {
			foreign.R = &userR{}
		}
		foreign.R.BattlesUserVotes = append(foreign.R.BattlesUserVotes, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UserID == foreign.ID {
				local.R.User = foreign
				if foreign.R == nil {
					foreign.R = &userR{}
				}
				foreign.R.BattlesUserVotes = append(foreign.R.BattlesUserVotes, local)
				break
			}
		}
	}

	return nil
}

// SetBattle of the battlesUserVote to the related item.
// Sets o.R.Battle to related.
// Adds o to related.R.BattlesUserVotes.
func (o *BattlesUserVote) SetBattle(exec boil.Executor, insert bool, related *Battle) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"battles_user_votes\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"battle_id"}),
		strmangle.WhereClause("\"", "\"", 2, battlesUserVotePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.BattleID, o.UserID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.BattleID = related.ID
	if o.R == nil {
		o.R = &battlesUserVoteR{
			Battle: related,
		}
	} else {
		o.R.Battle = related
	}

	if related.R == nil {
		related.R = &battleR{
			BattlesUserVotes: BattlesUserVoteSlice{o},
		}
	} else {
		related.R.BattlesUserVotes = append(related.R.BattlesUserVotes, o)
	}

	return nil
}

// SetUser of the battlesUserVote to the related item.
// Sets o.R.User to related.
// Adds o to related.R.BattlesUserVotes.
func (o *BattlesUserVote) SetUser(exec boil.Executor, insert bool, related *User) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"battles_user_votes\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"user_id"}),
		strmangle.WhereClause("\"", "\"", 2, battlesUserVotePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.BattleID, o.UserID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.UserID = related.ID
	if o.R == nil {
		o.R = &battlesUserVoteR{
			User: related,
		}
	} else {
		o.R.User = related
	}

	if related.R == nil {
		related.R = &userR{
			BattlesUserVotes: BattlesUserVoteSlice{o},
		}
	} else {
		related.R.BattlesUserVotes = append(related.R.BattlesUserVotes, o)
	}

	return nil
}

// BattlesUserVotes retrieves all the records using an executor.
func BattlesUserVotes(mods ...qm.QueryMod) battlesUserVoteQuery {
	mods = append(mods, qm.From("\"battles_user_votes\""))
	return battlesUserVoteQuery{NewQuery(mods...)}
}

// FindBattlesUserVote retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindBattlesUserVote(exec boil.Executor, battleID string, userID string, selectCols ...string) (*BattlesUserVote, error) {
	battlesUserVoteObj := &BattlesUserVote{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"battles_user_votes\" where \"battle_id\"=$1 AND \"user_id\"=$2", sel,
	)

	q := queries.Raw(query, battleID, userID)

	err := q.Bind(nil, exec, battlesUserVoteObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from battles_user_votes")
	}

	if err = battlesUserVoteObj.doAfterSelectHooks(exec); err != nil {
		return battlesUserVoteObj, err
	}

	return battlesUserVoteObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *BattlesUserVote) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no battles_user_votes provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(battlesUserVoteColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	battlesUserVoteInsertCacheMut.RLock()
	cache, cached := battlesUserVoteInsertCache[key]
	battlesUserVoteInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			battlesUserVoteAllColumns,
			battlesUserVoteColumnsWithDefault,
			battlesUserVoteColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(battlesUserVoteType, battlesUserVoteMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(battlesUserVoteType, battlesUserVoteMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"battles_user_votes\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"battles_user_votes\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into battles_user_votes")
	}

	if !cached {
		battlesUserVoteInsertCacheMut.Lock()
		battlesUserVoteInsertCache[key] = cache
		battlesUserVoteInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the BattlesUserVote.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *BattlesUserVote) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	battlesUserVoteUpdateCacheMut.RLock()
	cache, cached := battlesUserVoteUpdateCache[key]
	battlesUserVoteUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			battlesUserVoteAllColumns,
			battlesUserVotePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update battles_user_votes, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"battles_user_votes\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, battlesUserVotePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(battlesUserVoteType, battlesUserVoteMapping, append(wl, battlesUserVotePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update battles_user_votes row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for battles_user_votes")
	}

	if !cached {
		battlesUserVoteUpdateCacheMut.Lock()
		battlesUserVoteUpdateCache[key] = cache
		battlesUserVoteUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q battlesUserVoteQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for battles_user_votes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for battles_user_votes")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o BattlesUserVoteSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battlesUserVotePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"battles_user_votes\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, battlesUserVotePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in battlesUserVote slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all battlesUserVote")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *BattlesUserVote) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no battles_user_votes provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(battlesUserVoteColumnsWithDefault, o)

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

	battlesUserVoteUpsertCacheMut.RLock()
	cache, cached := battlesUserVoteUpsertCache[key]
	battlesUserVoteUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			battlesUserVoteAllColumns,
			battlesUserVoteColumnsWithDefault,
			battlesUserVoteColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			battlesUserVoteAllColumns,
			battlesUserVotePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert battles_user_votes, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(battlesUserVotePrimaryKeyColumns))
			copy(conflict, battlesUserVotePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"battles_user_votes\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(battlesUserVoteType, battlesUserVoteMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(battlesUserVoteType, battlesUserVoteMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert battles_user_votes")
	}

	if !cached {
		battlesUserVoteUpsertCacheMut.Lock()
		battlesUserVoteUpsertCache[key] = cache
		battlesUserVoteUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single BattlesUserVote record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *BattlesUserVote) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no BattlesUserVote provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), battlesUserVotePrimaryKeyMapping)
	sql := "DELETE FROM \"battles_user_votes\" WHERE \"battle_id\"=$1 AND \"user_id\"=$2"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from battles_user_votes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for battles_user_votes")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q battlesUserVoteQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no battlesUserVoteQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from battles_user_votes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for battles_user_votes")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o BattlesUserVoteSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(battlesUserVoteBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battlesUserVotePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"battles_user_votes\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, battlesUserVotePrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from battlesUserVote slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for battles_user_votes")
	}

	if len(battlesUserVoteAfterDeleteHooks) != 0 {
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
func (o *BattlesUserVote) Reload(exec boil.Executor) error {
	ret, err := FindBattlesUserVote(exec, o.BattleID, o.UserID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *BattlesUserVoteSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := BattlesUserVoteSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battlesUserVotePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"battles_user_votes\".* FROM \"battles_user_votes\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, battlesUserVotePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in BattlesUserVoteSlice")
	}

	*o = slice

	return nil
}

// BattlesUserVoteExists checks if the BattlesUserVote row exists.
func BattlesUserVoteExists(exec boil.Executor, battleID string, userID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"battles_user_votes\" where \"battle_id\"=$1 AND \"user_id\"=$2 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, battleID, userID)
	}
	row := exec.QueryRow(sql, battleID, userID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if battles_user_votes exists")
	}

	return exists, nil
}
