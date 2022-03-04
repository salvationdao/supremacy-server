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

// Player is an object representing the database table.
type Player struct {
	ID            string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	SyndicateID   string    `boiler:"syndicate_id" boil:"syndicate_id" json:"syndicateID" toml:"syndicateID" yaml:"syndicateID"`
	PublicAddress string    `boiler:"public_address" boil:"public_address" json:"publicAddress" toml:"publicAddress" yaml:"publicAddress"`
	DeletedAt     null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deletedAt,omitempty" toml:"deletedAt" yaml:"deletedAt,omitempty"`
	UpdatedAt     time.Time `boiler:"updated_at" boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
	CreatedAt     time.Time `boiler:"created_at" boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`

	R *playerR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L playerL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PlayerColumns = struct {
	ID            string
	SyndicateID   string
	PublicAddress string
	DeletedAt     string
	UpdatedAt     string
	CreatedAt     string
}{
	ID:            "id",
	SyndicateID:   "syndicate_id",
	PublicAddress: "public_address",
	DeletedAt:     "deleted_at",
	UpdatedAt:     "updated_at",
	CreatedAt:     "created_at",
}

var PlayerTableColumns = struct {
	ID            string
	SyndicateID   string
	PublicAddress string
	DeletedAt     string
	UpdatedAt     string
	CreatedAt     string
}{
	ID:            "players.id",
	SyndicateID:   "players.syndicate_id",
	PublicAddress: "players.public_address",
	DeletedAt:     "players.deleted_at",
	UpdatedAt:     "players.updated_at",
	CreatedAt:     "players.created_at",
}

// Generated where

var PlayerWhere = struct {
	ID            whereHelperstring
	SyndicateID   whereHelperstring
	PublicAddress whereHelperstring
	DeletedAt     whereHelpernull_Time
	UpdatedAt     whereHelpertime_Time
	CreatedAt     whereHelpertime_Time
}{
	ID:            whereHelperstring{field: "\"players\".\"id\""},
	SyndicateID:   whereHelperstring{field: "\"players\".\"syndicate_id\""},
	PublicAddress: whereHelperstring{field: "\"players\".\"public_address\""},
	DeletedAt:     whereHelpernull_Time{field: "\"players\".\"deleted_at\""},
	UpdatedAt:     whereHelpertime_Time{field: "\"players\".\"updated_at\""},
	CreatedAt:     whereHelpertime_Time{field: "\"players\".\"created_at\""},
}

// PlayerRels is where relationship names are stored.
var PlayerRels = struct {
	Syndicate  string
	OwnerMechs string
}{
	Syndicate:  "Syndicate",
	OwnerMechs: "OwnerMechs",
}

// playerR is where relationships are stored.
type playerR struct {
	Syndicate  *Syndicate `boiler:"Syndicate" boil:"Syndicate" json:"Syndicate" toml:"Syndicate" yaml:"Syndicate"`
	OwnerMechs MechSlice  `boiler:"OwnerMechs" boil:"OwnerMechs" json:"OwnerMechs" toml:"OwnerMechs" yaml:"OwnerMechs"`
}

// NewStruct creates a new relationship struct
func (*playerR) NewStruct() *playerR {
	return &playerR{}
}

// playerL is where Load methods for each relationship are stored.
type playerL struct{}

var (
	playerAllColumns            = []string{"id", "syndicate_id", "public_address", "deleted_at", "updated_at", "created_at"}
	playerColumnsWithoutDefault = []string{"syndicate_id", "public_address"}
	playerColumnsWithDefault    = []string{"id", "deleted_at", "updated_at", "created_at"}
	playerPrimaryKeyColumns     = []string{"id"}
	playerGeneratedColumns      = []string{}
)

type (
	// PlayerSlice is an alias for a slice of pointers to Player.
	// This should almost always be used instead of []Player.
	PlayerSlice []*Player
	// PlayerHook is the signature for custom Player hook methods
	PlayerHook func(boil.Executor, *Player) error

	playerQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	playerType                 = reflect.TypeOf(&Player{})
	playerMapping              = queries.MakeStructMapping(playerType)
	playerPrimaryKeyMapping, _ = queries.BindMapping(playerType, playerMapping, playerPrimaryKeyColumns)
	playerInsertCacheMut       sync.RWMutex
	playerInsertCache          = make(map[string]insertCache)
	playerUpdateCacheMut       sync.RWMutex
	playerUpdateCache          = make(map[string]updateCache)
	playerUpsertCacheMut       sync.RWMutex
	playerUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var playerAfterSelectHooks []PlayerHook

var playerBeforeInsertHooks []PlayerHook
var playerAfterInsertHooks []PlayerHook

var playerBeforeUpdateHooks []PlayerHook
var playerAfterUpdateHooks []PlayerHook

var playerBeforeDeleteHooks []PlayerHook
var playerAfterDeleteHooks []PlayerHook

var playerBeforeUpsertHooks []PlayerHook
var playerAfterUpsertHooks []PlayerHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Player) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Player) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Player) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Player) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playerBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Player) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Player) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playerBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Player) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Player) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Player) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPlayerHook registers your hook function for all future operations.
func AddPlayerHook(hookPoint boil.HookPoint, playerHook PlayerHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		playerAfterSelectHooks = append(playerAfterSelectHooks, playerHook)
	case boil.BeforeInsertHook:
		playerBeforeInsertHooks = append(playerBeforeInsertHooks, playerHook)
	case boil.AfterInsertHook:
		playerAfterInsertHooks = append(playerAfterInsertHooks, playerHook)
	case boil.BeforeUpdateHook:
		playerBeforeUpdateHooks = append(playerBeforeUpdateHooks, playerHook)
	case boil.AfterUpdateHook:
		playerAfterUpdateHooks = append(playerAfterUpdateHooks, playerHook)
	case boil.BeforeDeleteHook:
		playerBeforeDeleteHooks = append(playerBeforeDeleteHooks, playerHook)
	case boil.AfterDeleteHook:
		playerAfterDeleteHooks = append(playerAfterDeleteHooks, playerHook)
	case boil.BeforeUpsertHook:
		playerBeforeUpsertHooks = append(playerBeforeUpsertHooks, playerHook)
	case boil.AfterUpsertHook:
		playerAfterUpsertHooks = append(playerAfterUpsertHooks, playerHook)
	}
}

// One returns a single player record from the query.
func (q playerQuery) One(exec boil.Executor) (*Player, error) {
	o := &Player{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for players")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Player records from the query.
func (q playerQuery) All(exec boil.Executor) (PlayerSlice, error) {
	var o []*Player

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to Player slice")
	}

	if len(playerAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Player records in the query.
func (q playerQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count players rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q playerQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if players exists")
	}

	return count > 0, nil
}

// Syndicate pointed to by the foreign key.
func (o *Player) Syndicate(mods ...qm.QueryMod) syndicateQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.SyndicateID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Syndicates(queryMods...)
	queries.SetFrom(query.Query, "\"syndicates\"")

	return query
}

// OwnerMechs retrieves all the mech's Mechs with an executor via owner_id column.
func (o *Player) OwnerMechs(mods ...qm.QueryMod) mechQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"mechs\".\"owner_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"mechs\".\"deleted_at\""),
	)

	query := Mechs(queryMods...)
	queries.SetFrom(query.Query, "\"mechs\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"mechs\".*"})
	}

	return query
}

// LoadSyndicate allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playerL) LoadSyndicate(e boil.Executor, singular bool, maybePlayer interface{}, mods queries.Applicator) error {
	var slice []*Player
	var object *Player

	if singular {
		object = maybePlayer.(*Player)
	} else {
		slice = *maybePlayer.(*[]*Player)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playerR{}
		}
		args = append(args, object.SyndicateID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playerR{}
			}

			for _, a := range args {
				if a == obj.SyndicateID {
					continue Outer
				}
			}

			args = append(args, obj.SyndicateID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`syndicates`),
		qm.WhereIn(`syndicates.id in ?`, args...),
		qmhelper.WhereIsNull(`syndicates.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Syndicate")
	}

	var resultSlice []*Syndicate
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Syndicate")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for syndicates")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for syndicates")
	}

	if len(playerAfterSelectHooks) != 0 {
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
		object.R.Syndicate = foreign
		if foreign.R == nil {
			foreign.R = &syndicateR{}
		}
		foreign.R.Players = append(foreign.R.Players, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.SyndicateID == foreign.ID {
				local.R.Syndicate = foreign
				if foreign.R == nil {
					foreign.R = &syndicateR{}
				}
				foreign.R.Players = append(foreign.R.Players, local)
				break
			}
		}
	}

	return nil
}

// LoadOwnerMechs allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (playerL) LoadOwnerMechs(e boil.Executor, singular bool, maybePlayer interface{}, mods queries.Applicator) error {
	var slice []*Player
	var object *Player

	if singular {
		object = maybePlayer.(*Player)
	} else {
		slice = *maybePlayer.(*[]*Player)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playerR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playerR{}
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
		qm.From(`mechs`),
		qm.WhereIn(`mechs.owner_id in ?`, args...),
		qmhelper.WhereIsNull(`mechs.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load mechs")
	}

	var resultSlice []*Mech
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice mechs")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on mechs")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for mechs")
	}

	if len(mechAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.OwnerMechs = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &mechR{}
			}
			foreign.R.Owner = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.OwnerID {
				local.R.OwnerMechs = append(local.R.OwnerMechs, foreign)
				if foreign.R == nil {
					foreign.R = &mechR{}
				}
				foreign.R.Owner = local
				break
			}
		}
	}

	return nil
}

// SetSyndicate of the player to the related item.
// Sets o.R.Syndicate to related.
// Adds o to related.R.Players.
func (o *Player) SetSyndicate(exec boil.Executor, insert bool, related *Syndicate) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"players\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"syndicate_id"}),
		strmangle.WhereClause("\"", "\"", 2, playerPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.SyndicateID = related.ID
	if o.R == nil {
		o.R = &playerR{
			Syndicate: related,
		}
	} else {
		o.R.Syndicate = related
	}

	if related.R == nil {
		related.R = &syndicateR{
			Players: PlayerSlice{o},
		}
	} else {
		related.R.Players = append(related.R.Players, o)
	}

	return nil
}

// AddOwnerMechs adds the given related objects to the existing relationships
// of the player, optionally inserting them as new records.
// Appends related to o.R.OwnerMechs.
// Sets related.R.Owner appropriately.
func (o *Player) AddOwnerMechs(exec boil.Executor, insert bool, related ...*Mech) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.OwnerID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"mechs\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"owner_id"}),
				strmangle.WhereClause("\"", "\"", 2, mechPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.OwnerID = o.ID
		}
	}

	if o.R == nil {
		o.R = &playerR{
			OwnerMechs: related,
		}
	} else {
		o.R.OwnerMechs = append(o.R.OwnerMechs, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &mechR{
				Owner: o,
			}
		} else {
			rel.R.Owner = o
		}
	}
	return nil
}

// Players retrieves all the records using an executor.
func Players(mods ...qm.QueryMod) playerQuery {
	mods = append(mods, qm.From("\"players\""), qmhelper.WhereIsNull("\"players\".\"deleted_at\""))
	return playerQuery{NewQuery(mods...)}
}

// FindPlayer retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPlayer(exec boil.Executor, iD string, selectCols ...string) (*Player, error) {
	playerObj := &Player{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"players\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, playerObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from players")
	}

	if err = playerObj.doAfterSelectHooks(exec); err != nil {
		return playerObj, err
	}

	return playerObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Player) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no players provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(playerColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	playerInsertCacheMut.RLock()
	cache, cached := playerInsertCache[key]
	playerInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			playerAllColumns,
			playerColumnsWithDefault,
			playerColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(playerType, playerMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(playerType, playerMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"players\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"players\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into players")
	}

	if !cached {
		playerInsertCacheMut.Lock()
		playerInsertCache[key] = cache
		playerInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the Player.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Player) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	playerUpdateCacheMut.RLock()
	cache, cached := playerUpdateCache[key]
	playerUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			playerAllColumns,
			playerPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update players, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"players\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, playerPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(playerType, playerMapping, append(wl, playerPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update players row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for players")
	}

	if !cached {
		playerUpdateCacheMut.Lock()
		playerUpdateCache[key] = cache
		playerUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q playerQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for players")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for players")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PlayerSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"players\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, playerPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in player slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all player")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Player) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no players provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(playerColumnsWithDefault, o)

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

	playerUpsertCacheMut.RLock()
	cache, cached := playerUpsertCache[key]
	playerUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			playerAllColumns,
			playerColumnsWithDefault,
			playerColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			playerAllColumns,
			playerPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert players, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(playerPrimaryKeyColumns))
			copy(conflict, playerPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"players\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(playerType, playerMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(playerType, playerMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert players")
	}

	if !cached {
		playerUpsertCacheMut.Lock()
		playerUpsertCache[key] = cache
		playerUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single Player record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Player) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no Player provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), playerPrimaryKeyMapping)
		sql = "DELETE FROM \"players\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"players\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(playerType, playerMapping, append(wl, playerPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from players")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for players")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q playerQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no playerQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from players")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for players")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PlayerSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(playerBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"players\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playerPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"players\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, playerPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from player slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for players")
	}

	if len(playerAfterDeleteHooks) != 0 {
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
func (o *Player) Reload(exec boil.Executor) error {
	ret, err := FindPlayer(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PlayerSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PlayerSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"players\".* FROM \"players\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playerPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in PlayerSlice")
	}

	*o = slice

	return nil
}

// PlayerExists checks if the Player row exists.
func PlayerExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"players\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if players exists")
	}

	return exists, nil
}