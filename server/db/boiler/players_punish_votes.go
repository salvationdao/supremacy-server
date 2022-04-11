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

// PlayersPunishVote is an object representing the database table.
type PlayersPunishVote struct {
	ID           string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	PunishVoteID string    `boiler:"punish_vote_id" boil:"punish_vote_id" json:"punish_vote_id" toml:"punish_vote_id" yaml:"punish_vote_id"`
	PlayerID     string    `boiler:"player_id" boil:"player_id" json:"player_id" toml:"player_id" yaml:"player_id"`
	IsAgreed     bool      `boiler:"is_agreed" boil:"is_agreed" json:"is_agreed" toml:"is_agreed" yaml:"is_agreed"`
	CreatedAt    time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt    time.Time `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	DeletedAt    null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`

	R *playersPunishVoteR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L playersPunishVoteL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PlayersPunishVoteColumns = struct {
	ID           string
	PunishVoteID string
	PlayerID     string
	IsAgreed     string
	CreatedAt    string
	UpdatedAt    string
	DeletedAt    string
}{
	ID:           "id",
	PunishVoteID: "punish_vote_id",
	PlayerID:     "player_id",
	IsAgreed:     "is_agreed",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
	DeletedAt:    "deleted_at",
}

var PlayersPunishVoteTableColumns = struct {
	ID           string
	PunishVoteID string
	PlayerID     string
	IsAgreed     string
	CreatedAt    string
	UpdatedAt    string
	DeletedAt    string
}{
	ID:           "players_punish_votes.id",
	PunishVoteID: "players_punish_votes.punish_vote_id",
	PlayerID:     "players_punish_votes.player_id",
	IsAgreed:     "players_punish_votes.is_agreed",
	CreatedAt:    "players_punish_votes.created_at",
	UpdatedAt:    "players_punish_votes.updated_at",
	DeletedAt:    "players_punish_votes.deleted_at",
}

// Generated where

var PlayersPunishVoteWhere = struct {
	ID           whereHelperstring
	PunishVoteID whereHelperstring
	PlayerID     whereHelperstring
	IsAgreed     whereHelperbool
	CreatedAt    whereHelpertime_Time
	UpdatedAt    whereHelpertime_Time
	DeletedAt    whereHelpernull_Time
}{
	ID:           whereHelperstring{field: "\"players_punish_votes\".\"id\""},
	PunishVoteID: whereHelperstring{field: "\"players_punish_votes\".\"punish_vote_id\""},
	PlayerID:     whereHelperstring{field: "\"players_punish_votes\".\"player_id\""},
	IsAgreed:     whereHelperbool{field: "\"players_punish_votes\".\"is_agreed\""},
	CreatedAt:    whereHelpertime_Time{field: "\"players_punish_votes\".\"created_at\""},
	UpdatedAt:    whereHelpertime_Time{field: "\"players_punish_votes\".\"updated_at\""},
	DeletedAt:    whereHelpernull_Time{field: "\"players_punish_votes\".\"deleted_at\""},
}

// PlayersPunishVoteRels is where relationship names are stored.
var PlayersPunishVoteRels = struct {
	Player     string
	PunishVote string
}{
	Player:     "Player",
	PunishVote: "PunishVote",
}

// playersPunishVoteR is where relationships are stored.
type playersPunishVoteR struct {
	Player     *Player     `boiler:"Player" boil:"Player" json:"Player" toml:"Player" yaml:"Player"`
	PunishVote *PunishVote `boiler:"PunishVote" boil:"PunishVote" json:"PunishVote" toml:"PunishVote" yaml:"PunishVote"`
}

// NewStruct creates a new relationship struct
func (*playersPunishVoteR) NewStruct() *playersPunishVoteR {
	return &playersPunishVoteR{}
}

// playersPunishVoteL is where Load methods for each relationship are stored.
type playersPunishVoteL struct{}

var (
	playersPunishVoteAllColumns            = []string{"id", "punish_vote_id", "player_id", "is_agreed", "created_at", "updated_at", "deleted_at"}
	playersPunishVoteColumnsWithoutDefault = []string{"punish_vote_id", "player_id", "is_agreed"}
	playersPunishVoteColumnsWithDefault    = []string{"id", "created_at", "updated_at", "deleted_at"}
	playersPunishVotePrimaryKeyColumns     = []string{"id"}
	playersPunishVoteGeneratedColumns      = []string{}
)

type (
	// PlayersPunishVoteSlice is an alias for a slice of pointers to PlayersPunishVote.
	// This should almost always be used instead of []PlayersPunishVote.
	PlayersPunishVoteSlice []*PlayersPunishVote
	// PlayersPunishVoteHook is the signature for custom PlayersPunishVote hook methods
	PlayersPunishVoteHook func(boil.Executor, *PlayersPunishVote) error

	playersPunishVoteQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	playersPunishVoteType                 = reflect.TypeOf(&PlayersPunishVote{})
	playersPunishVoteMapping              = queries.MakeStructMapping(playersPunishVoteType)
	playersPunishVotePrimaryKeyMapping, _ = queries.BindMapping(playersPunishVoteType, playersPunishVoteMapping, playersPunishVotePrimaryKeyColumns)
	playersPunishVoteInsertCacheMut       sync.RWMutex
	playersPunishVoteInsertCache          = make(map[string]insertCache)
	playersPunishVoteUpdateCacheMut       sync.RWMutex
	playersPunishVoteUpdateCache          = make(map[string]updateCache)
	playersPunishVoteUpsertCacheMut       sync.RWMutex
	playersPunishVoteUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var playersPunishVoteAfterSelectHooks []PlayersPunishVoteHook

var playersPunishVoteBeforeInsertHooks []PlayersPunishVoteHook
var playersPunishVoteAfterInsertHooks []PlayersPunishVoteHook

var playersPunishVoteBeforeUpdateHooks []PlayersPunishVoteHook
var playersPunishVoteAfterUpdateHooks []PlayersPunishVoteHook

var playersPunishVoteBeforeDeleteHooks []PlayersPunishVoteHook
var playersPunishVoteAfterDeleteHooks []PlayersPunishVoteHook

var playersPunishVoteBeforeUpsertHooks []PlayersPunishVoteHook
var playersPunishVoteAfterUpsertHooks []PlayersPunishVoteHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *PlayersPunishVote) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range playersPunishVoteAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *PlayersPunishVote) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playersPunishVoteBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *PlayersPunishVote) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playersPunishVoteAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *PlayersPunishVote) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playersPunishVoteBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *PlayersPunishVote) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playersPunishVoteAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *PlayersPunishVote) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playersPunishVoteBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *PlayersPunishVote) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playersPunishVoteAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *PlayersPunishVote) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playersPunishVoteBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *PlayersPunishVote) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playersPunishVoteAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPlayersPunishVoteHook registers your hook function for all future operations.
func AddPlayersPunishVoteHook(hookPoint boil.HookPoint, playersPunishVoteHook PlayersPunishVoteHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		playersPunishVoteAfterSelectHooks = append(playersPunishVoteAfterSelectHooks, playersPunishVoteHook)
	case boil.BeforeInsertHook:
		playersPunishVoteBeforeInsertHooks = append(playersPunishVoteBeforeInsertHooks, playersPunishVoteHook)
	case boil.AfterInsertHook:
		playersPunishVoteAfterInsertHooks = append(playersPunishVoteAfterInsertHooks, playersPunishVoteHook)
	case boil.BeforeUpdateHook:
		playersPunishVoteBeforeUpdateHooks = append(playersPunishVoteBeforeUpdateHooks, playersPunishVoteHook)
	case boil.AfterUpdateHook:
		playersPunishVoteAfterUpdateHooks = append(playersPunishVoteAfterUpdateHooks, playersPunishVoteHook)
	case boil.BeforeDeleteHook:
		playersPunishVoteBeforeDeleteHooks = append(playersPunishVoteBeforeDeleteHooks, playersPunishVoteHook)
	case boil.AfterDeleteHook:
		playersPunishVoteAfterDeleteHooks = append(playersPunishVoteAfterDeleteHooks, playersPunishVoteHook)
	case boil.BeforeUpsertHook:
		playersPunishVoteBeforeUpsertHooks = append(playersPunishVoteBeforeUpsertHooks, playersPunishVoteHook)
	case boil.AfterUpsertHook:
		playersPunishVoteAfterUpsertHooks = append(playersPunishVoteAfterUpsertHooks, playersPunishVoteHook)
	}
}

// One returns a single playersPunishVote record from the query.
func (q playersPunishVoteQuery) One(exec boil.Executor) (*PlayersPunishVote, error) {
	o := &PlayersPunishVote{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for players_punish_votes")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all PlayersPunishVote records from the query.
func (q playersPunishVoteQuery) All(exec boil.Executor) (PlayersPunishVoteSlice, error) {
	var o []*PlayersPunishVote

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to PlayersPunishVote slice")
	}

	if len(playersPunishVoteAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all PlayersPunishVote records in the query.
func (q playersPunishVoteQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count players_punish_votes rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q playersPunishVoteQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if players_punish_votes exists")
	}

	return count > 0, nil
}

// Player pointed to by the foreign key.
func (o *PlayersPunishVote) Player(mods ...qm.QueryMod) playerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.PlayerID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Players(queryMods...)
	queries.SetFrom(query.Query, "\"players\"")

	return query
}

// PunishVote pointed to by the foreign key.
func (o *PlayersPunishVote) PunishVote(mods ...qm.QueryMod) punishVoteQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.PunishVoteID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := PunishVotes(queryMods...)
	queries.SetFrom(query.Query, "\"punish_votes\"")

	return query
}

// LoadPlayer allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playersPunishVoteL) LoadPlayer(e boil.Executor, singular bool, maybePlayersPunishVote interface{}, mods queries.Applicator) error {
	var slice []*PlayersPunishVote
	var object *PlayersPunishVote

	if singular {
		object = maybePlayersPunishVote.(*PlayersPunishVote)
	} else {
		slice = *maybePlayersPunishVote.(*[]*PlayersPunishVote)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playersPunishVoteR{}
		}
		args = append(args, object.PlayerID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playersPunishVoteR{}
			}

			for _, a := range args {
				if a == obj.PlayerID {
					continue Outer
				}
			}

			args = append(args, obj.PlayerID)

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

	if len(playersPunishVoteAfterSelectHooks) != 0 {
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
		object.R.Player = foreign
		if foreign.R == nil {
			foreign.R = &playerR{}
		}
		foreign.R.PlayersPunishVotes = append(foreign.R.PlayersPunishVotes, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.PlayerID == foreign.ID {
				local.R.Player = foreign
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.PlayersPunishVotes = append(foreign.R.PlayersPunishVotes, local)
				break
			}
		}
	}

	return nil
}

// LoadPunishVote allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playersPunishVoteL) LoadPunishVote(e boil.Executor, singular bool, maybePlayersPunishVote interface{}, mods queries.Applicator) error {
	var slice []*PlayersPunishVote
	var object *PlayersPunishVote

	if singular {
		object = maybePlayersPunishVote.(*PlayersPunishVote)
	} else {
		slice = *maybePlayersPunishVote.(*[]*PlayersPunishVote)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playersPunishVoteR{}
		}
		args = append(args, object.PunishVoteID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playersPunishVoteR{}
			}

			for _, a := range args {
				if a == obj.PunishVoteID {
					continue Outer
				}
			}

			args = append(args, obj.PunishVoteID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`punish_votes`),
		qm.WhereIn(`punish_votes.id in ?`, args...),
		qmhelper.WhereIsNull(`punish_votes.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load PunishVote")
	}

	var resultSlice []*PunishVote
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice PunishVote")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for punish_votes")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for punish_votes")
	}

	if len(playersPunishVoteAfterSelectHooks) != 0 {
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
		object.R.PunishVote = foreign
		if foreign.R == nil {
			foreign.R = &punishVoteR{}
		}
		foreign.R.PlayersPunishVotes = append(foreign.R.PlayersPunishVotes, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.PunishVoteID == foreign.ID {
				local.R.PunishVote = foreign
				if foreign.R == nil {
					foreign.R = &punishVoteR{}
				}
				foreign.R.PlayersPunishVotes = append(foreign.R.PlayersPunishVotes, local)
				break
			}
		}
	}

	return nil
}

// SetPlayer of the playersPunishVote to the related item.
// Sets o.R.Player to related.
// Adds o to related.R.PlayersPunishVotes.
func (o *PlayersPunishVote) SetPlayer(exec boil.Executor, insert bool, related *Player) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"players_punish_votes\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"player_id"}),
		strmangle.WhereClause("\"", "\"", 2, playersPunishVotePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.PlayerID = related.ID
	if o.R == nil {
		o.R = &playersPunishVoteR{
			Player: related,
		}
	} else {
		o.R.Player = related
	}

	if related.R == nil {
		related.R = &playerR{
			PlayersPunishVotes: PlayersPunishVoteSlice{o},
		}
	} else {
		related.R.PlayersPunishVotes = append(related.R.PlayersPunishVotes, o)
	}

	return nil
}

// SetPunishVote of the playersPunishVote to the related item.
// Sets o.R.PunishVote to related.
// Adds o to related.R.PlayersPunishVotes.
func (o *PlayersPunishVote) SetPunishVote(exec boil.Executor, insert bool, related *PunishVote) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"players_punish_votes\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"punish_vote_id"}),
		strmangle.WhereClause("\"", "\"", 2, playersPunishVotePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.PunishVoteID = related.ID
	if o.R == nil {
		o.R = &playersPunishVoteR{
			PunishVote: related,
		}
	} else {
		o.R.PunishVote = related
	}

	if related.R == nil {
		related.R = &punishVoteR{
			PlayersPunishVotes: PlayersPunishVoteSlice{o},
		}
	} else {
		related.R.PlayersPunishVotes = append(related.R.PlayersPunishVotes, o)
	}

	return nil
}

// PlayersPunishVotes retrieves all the records using an executor.
func PlayersPunishVotes(mods ...qm.QueryMod) playersPunishVoteQuery {
	mods = append(mods, qm.From("\"players_punish_votes\""), qmhelper.WhereIsNull("\"players_punish_votes\".\"deleted_at\""))
	return playersPunishVoteQuery{NewQuery(mods...)}
}

// FindPlayersPunishVote retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPlayersPunishVote(exec boil.Executor, iD string, selectCols ...string) (*PlayersPunishVote, error) {
	playersPunishVoteObj := &PlayersPunishVote{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"players_punish_votes\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, playersPunishVoteObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from players_punish_votes")
	}

	if err = playersPunishVoteObj.doAfterSelectHooks(exec); err != nil {
		return playersPunishVoteObj, err
	}

	return playersPunishVoteObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *PlayersPunishVote) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no players_punish_votes provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(playersPunishVoteColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	playersPunishVoteInsertCacheMut.RLock()
	cache, cached := playersPunishVoteInsertCache[key]
	playersPunishVoteInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			playersPunishVoteAllColumns,
			playersPunishVoteColumnsWithDefault,
			playersPunishVoteColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(playersPunishVoteType, playersPunishVoteMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(playersPunishVoteType, playersPunishVoteMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"players_punish_votes\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"players_punish_votes\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into players_punish_votes")
	}

	if !cached {
		playersPunishVoteInsertCacheMut.Lock()
		playersPunishVoteInsertCache[key] = cache
		playersPunishVoteInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the PlayersPunishVote.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *PlayersPunishVote) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	playersPunishVoteUpdateCacheMut.RLock()
	cache, cached := playersPunishVoteUpdateCache[key]
	playersPunishVoteUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			playersPunishVoteAllColumns,
			playersPunishVotePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update players_punish_votes, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"players_punish_votes\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, playersPunishVotePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(playersPunishVoteType, playersPunishVoteMapping, append(wl, playersPunishVotePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update players_punish_votes row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for players_punish_votes")
	}

	if !cached {
		playersPunishVoteUpdateCacheMut.Lock()
		playersPunishVoteUpdateCache[key] = cache
		playersPunishVoteUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q playersPunishVoteQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for players_punish_votes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for players_punish_votes")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PlayersPunishVoteSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playersPunishVotePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"players_punish_votes\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, playersPunishVotePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in playersPunishVote slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all playersPunishVote")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *PlayersPunishVote) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no players_punish_votes provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}
	o.UpdatedAt = currTime

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(playersPunishVoteColumnsWithDefault, o)

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

	playersPunishVoteUpsertCacheMut.RLock()
	cache, cached := playersPunishVoteUpsertCache[key]
	playersPunishVoteUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			playersPunishVoteAllColumns,
			playersPunishVoteColumnsWithDefault,
			playersPunishVoteColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			playersPunishVoteAllColumns,
			playersPunishVotePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert players_punish_votes, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(playersPunishVotePrimaryKeyColumns))
			copy(conflict, playersPunishVotePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"players_punish_votes\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(playersPunishVoteType, playersPunishVoteMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(playersPunishVoteType, playersPunishVoteMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert players_punish_votes")
	}

	if !cached {
		playersPunishVoteUpsertCacheMut.Lock()
		playersPunishVoteUpsertCache[key] = cache
		playersPunishVoteUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single PlayersPunishVote record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *PlayersPunishVote) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no PlayersPunishVote provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), playersPunishVotePrimaryKeyMapping)
		sql = "DELETE FROM \"players_punish_votes\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"players_punish_votes\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(playersPunishVoteType, playersPunishVoteMapping, append(wl, playersPunishVotePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from players_punish_votes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for players_punish_votes")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q playersPunishVoteQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no playersPunishVoteQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from players_punish_votes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for players_punish_votes")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PlayersPunishVoteSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(playersPunishVoteBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playersPunishVotePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"players_punish_votes\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playersPunishVotePrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playersPunishVotePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"players_punish_votes\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, playersPunishVotePrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from playersPunishVote slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for players_punish_votes")
	}

	if len(playersPunishVoteAfterDeleteHooks) != 0 {
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
func (o *PlayersPunishVote) Reload(exec boil.Executor) error {
	ret, err := FindPlayersPunishVote(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PlayersPunishVoteSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PlayersPunishVoteSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playersPunishVotePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"players_punish_votes\".* FROM \"players_punish_votes\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playersPunishVotePrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in PlayersPunishVoteSlice")
	}

	*o = slice

	return nil
}

// PlayersPunishVoteExists checks if the PlayersPunishVote row exists.
func PlayersPunishVoteExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"players_punish_votes\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if players_punish_votes exists")
	}

	return exists, nil
}
