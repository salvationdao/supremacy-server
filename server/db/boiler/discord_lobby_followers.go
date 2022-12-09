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

// DiscordLobbyFollower is an object representing the database table.
type DiscordLobbyFollower struct {
	ID                         string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	DiscordMemberID            string    `boiler:"discord_member_id" boil:"discord_member_id" json:"discord_member_id" toml:"discord_member_id" yaml:"discord_member_id"`
	DiscordLobbyAnnoucementsID string    `boiler:"discord_lobby_annoucements_id" boil:"discord_lobby_annoucements_id" json:"discord_lobby_annoucements_id" toml:"discord_lobby_annoucements_id" yaml:"discord_lobby_annoucements_id"`
	CreatedAt                  time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *discordLobbyFollowerR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L discordLobbyFollowerL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var DiscordLobbyFollowerColumns = struct {
	ID                         string
	DiscordMemberID            string
	DiscordLobbyAnnoucementsID string
	CreatedAt                  string
}{
	ID:                         "id",
	DiscordMemberID:            "discord_member_id",
	DiscordLobbyAnnoucementsID: "discord_lobby_annoucements_id",
	CreatedAt:                  "created_at",
}

var DiscordLobbyFollowerTableColumns = struct {
	ID                         string
	DiscordMemberID            string
	DiscordLobbyAnnoucementsID string
	CreatedAt                  string
}{
	ID:                         "discord_lobby_followers.id",
	DiscordMemberID:            "discord_lobby_followers.discord_member_id",
	DiscordLobbyAnnoucementsID: "discord_lobby_followers.discord_lobby_annoucements_id",
	CreatedAt:                  "discord_lobby_followers.created_at",
}

// Generated where

var DiscordLobbyFollowerWhere = struct {
	ID                         whereHelperstring
	DiscordMemberID            whereHelperstring
	DiscordLobbyAnnoucementsID whereHelperstring
	CreatedAt                  whereHelpertime_Time
}{
	ID:                         whereHelperstring{field: "\"discord_lobby_followers\".\"id\""},
	DiscordMemberID:            whereHelperstring{field: "\"discord_lobby_followers\".\"discord_member_id\""},
	DiscordLobbyAnnoucementsID: whereHelperstring{field: "\"discord_lobby_followers\".\"discord_lobby_annoucements_id\""},
	CreatedAt:                  whereHelpertime_Time{field: "\"discord_lobby_followers\".\"created_at\""},
}

// DiscordLobbyFollowerRels is where relationship names are stored.
var DiscordLobbyFollowerRels = struct {
	DiscordLobbyAnnoucement string
}{
	DiscordLobbyAnnoucement: "DiscordLobbyAnnoucement",
}

// discordLobbyFollowerR is where relationships are stored.
type discordLobbyFollowerR struct {
	DiscordLobbyAnnoucement *DiscordLobbyAnnoucement `boiler:"DiscordLobbyAnnoucement" boil:"DiscordLobbyAnnoucement" json:"DiscordLobbyAnnoucement" toml:"DiscordLobbyAnnoucement" yaml:"DiscordLobbyAnnoucement"`
}

// NewStruct creates a new relationship struct
func (*discordLobbyFollowerR) NewStruct() *discordLobbyFollowerR {
	return &discordLobbyFollowerR{}
}

// discordLobbyFollowerL is where Load methods for each relationship are stored.
type discordLobbyFollowerL struct{}

var (
	discordLobbyFollowerAllColumns            = []string{"id", "discord_member_id", "discord_lobby_annoucements_id", "created_at"}
	discordLobbyFollowerColumnsWithoutDefault = []string{"discord_member_id", "discord_lobby_annoucements_id"}
	discordLobbyFollowerColumnsWithDefault    = []string{"id", "created_at"}
	discordLobbyFollowerPrimaryKeyColumns     = []string{"id"}
	discordLobbyFollowerGeneratedColumns      = []string{}
)

type (
	// DiscordLobbyFollowerSlice is an alias for a slice of pointers to DiscordLobbyFollower.
	// This should almost always be used instead of []DiscordLobbyFollower.
	DiscordLobbyFollowerSlice []*DiscordLobbyFollower
	// DiscordLobbyFollowerHook is the signature for custom DiscordLobbyFollower hook methods
	DiscordLobbyFollowerHook func(boil.Executor, *DiscordLobbyFollower) error

	discordLobbyFollowerQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	discordLobbyFollowerType                 = reflect.TypeOf(&DiscordLobbyFollower{})
	discordLobbyFollowerMapping              = queries.MakeStructMapping(discordLobbyFollowerType)
	discordLobbyFollowerPrimaryKeyMapping, _ = queries.BindMapping(discordLobbyFollowerType, discordLobbyFollowerMapping, discordLobbyFollowerPrimaryKeyColumns)
	discordLobbyFollowerInsertCacheMut       sync.RWMutex
	discordLobbyFollowerInsertCache          = make(map[string]insertCache)
	discordLobbyFollowerUpdateCacheMut       sync.RWMutex
	discordLobbyFollowerUpdateCache          = make(map[string]updateCache)
	discordLobbyFollowerUpsertCacheMut       sync.RWMutex
	discordLobbyFollowerUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var discordLobbyFollowerAfterSelectHooks []DiscordLobbyFollowerHook

var discordLobbyFollowerBeforeInsertHooks []DiscordLobbyFollowerHook
var discordLobbyFollowerAfterInsertHooks []DiscordLobbyFollowerHook

var discordLobbyFollowerBeforeUpdateHooks []DiscordLobbyFollowerHook
var discordLobbyFollowerAfterUpdateHooks []DiscordLobbyFollowerHook

var discordLobbyFollowerBeforeDeleteHooks []DiscordLobbyFollowerHook
var discordLobbyFollowerAfterDeleteHooks []DiscordLobbyFollowerHook

var discordLobbyFollowerBeforeUpsertHooks []DiscordLobbyFollowerHook
var discordLobbyFollowerAfterUpsertHooks []DiscordLobbyFollowerHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *DiscordLobbyFollower) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range discordLobbyFollowerAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *DiscordLobbyFollower) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range discordLobbyFollowerBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *DiscordLobbyFollower) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range discordLobbyFollowerAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *DiscordLobbyFollower) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range discordLobbyFollowerBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *DiscordLobbyFollower) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range discordLobbyFollowerAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *DiscordLobbyFollower) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range discordLobbyFollowerBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *DiscordLobbyFollower) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range discordLobbyFollowerAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *DiscordLobbyFollower) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range discordLobbyFollowerBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *DiscordLobbyFollower) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range discordLobbyFollowerAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddDiscordLobbyFollowerHook registers your hook function for all future operations.
func AddDiscordLobbyFollowerHook(hookPoint boil.HookPoint, discordLobbyFollowerHook DiscordLobbyFollowerHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		discordLobbyFollowerAfterSelectHooks = append(discordLobbyFollowerAfterSelectHooks, discordLobbyFollowerHook)
	case boil.BeforeInsertHook:
		discordLobbyFollowerBeforeInsertHooks = append(discordLobbyFollowerBeforeInsertHooks, discordLobbyFollowerHook)
	case boil.AfterInsertHook:
		discordLobbyFollowerAfterInsertHooks = append(discordLobbyFollowerAfterInsertHooks, discordLobbyFollowerHook)
	case boil.BeforeUpdateHook:
		discordLobbyFollowerBeforeUpdateHooks = append(discordLobbyFollowerBeforeUpdateHooks, discordLobbyFollowerHook)
	case boil.AfterUpdateHook:
		discordLobbyFollowerAfterUpdateHooks = append(discordLobbyFollowerAfterUpdateHooks, discordLobbyFollowerHook)
	case boil.BeforeDeleteHook:
		discordLobbyFollowerBeforeDeleteHooks = append(discordLobbyFollowerBeforeDeleteHooks, discordLobbyFollowerHook)
	case boil.AfterDeleteHook:
		discordLobbyFollowerAfterDeleteHooks = append(discordLobbyFollowerAfterDeleteHooks, discordLobbyFollowerHook)
	case boil.BeforeUpsertHook:
		discordLobbyFollowerBeforeUpsertHooks = append(discordLobbyFollowerBeforeUpsertHooks, discordLobbyFollowerHook)
	case boil.AfterUpsertHook:
		discordLobbyFollowerAfterUpsertHooks = append(discordLobbyFollowerAfterUpsertHooks, discordLobbyFollowerHook)
	}
}

// One returns a single discordLobbyFollower record from the query.
func (q discordLobbyFollowerQuery) One(exec boil.Executor) (*DiscordLobbyFollower, error) {
	o := &DiscordLobbyFollower{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for discord_lobby_followers")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all DiscordLobbyFollower records from the query.
func (q discordLobbyFollowerQuery) All(exec boil.Executor) (DiscordLobbyFollowerSlice, error) {
	var o []*DiscordLobbyFollower

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to DiscordLobbyFollower slice")
	}

	if len(discordLobbyFollowerAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all DiscordLobbyFollower records in the query.
func (q discordLobbyFollowerQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count discord_lobby_followers rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q discordLobbyFollowerQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if discord_lobby_followers exists")
	}

	return count > 0, nil
}

// DiscordLobbyAnnoucement pointed to by the foreign key.
func (o *DiscordLobbyFollower) DiscordLobbyAnnoucement(mods ...qm.QueryMod) discordLobbyAnnoucementQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.DiscordLobbyAnnoucementsID),
	}

	queryMods = append(queryMods, mods...)

	query := DiscordLobbyAnnoucements(queryMods...)
	queries.SetFrom(query.Query, "\"discord_lobby_annoucements\"")

	return query
}

// LoadDiscordLobbyAnnoucement allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (discordLobbyFollowerL) LoadDiscordLobbyAnnoucement(e boil.Executor, singular bool, maybeDiscordLobbyFollower interface{}, mods queries.Applicator) error {
	var slice []*DiscordLobbyFollower
	var object *DiscordLobbyFollower

	if singular {
		object = maybeDiscordLobbyFollower.(*DiscordLobbyFollower)
	} else {
		slice = *maybeDiscordLobbyFollower.(*[]*DiscordLobbyFollower)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &discordLobbyFollowerR{}
		}
		args = append(args, object.DiscordLobbyAnnoucementsID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &discordLobbyFollowerR{}
			}

			for _, a := range args {
				if a == obj.DiscordLobbyAnnoucementsID {
					continue Outer
				}
			}

			args = append(args, obj.DiscordLobbyAnnoucementsID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`discord_lobby_annoucements`),
		qm.WhereIn(`discord_lobby_annoucements.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load DiscordLobbyAnnoucement")
	}

	var resultSlice []*DiscordLobbyAnnoucement
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice DiscordLobbyAnnoucement")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for discord_lobby_annoucements")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for discord_lobby_annoucements")
	}

	if len(discordLobbyFollowerAfterSelectHooks) != 0 {
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
		object.R.DiscordLobbyAnnoucement = foreign
		if foreign.R == nil {
			foreign.R = &discordLobbyAnnoucementR{}
		}
		foreign.R.DiscordLobbyFollowers = append(foreign.R.DiscordLobbyFollowers, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.DiscordLobbyAnnoucementsID == foreign.ID {
				local.R.DiscordLobbyAnnoucement = foreign
				if foreign.R == nil {
					foreign.R = &discordLobbyAnnoucementR{}
				}
				foreign.R.DiscordLobbyFollowers = append(foreign.R.DiscordLobbyFollowers, local)
				break
			}
		}
	}

	return nil
}

// SetDiscordLobbyAnnoucement of the discordLobbyFollower to the related item.
// Sets o.R.DiscordLobbyAnnoucement to related.
// Adds o to related.R.DiscordLobbyFollowers.
func (o *DiscordLobbyFollower) SetDiscordLobbyAnnoucement(exec boil.Executor, insert bool, related *DiscordLobbyAnnoucement) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"discord_lobby_followers\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"discord_lobby_annoucements_id"}),
		strmangle.WhereClause("\"", "\"", 2, discordLobbyFollowerPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.DiscordLobbyAnnoucementsID = related.ID
	if o.R == nil {
		o.R = &discordLobbyFollowerR{
			DiscordLobbyAnnoucement: related,
		}
	} else {
		o.R.DiscordLobbyAnnoucement = related
	}

	if related.R == nil {
		related.R = &discordLobbyAnnoucementR{
			DiscordLobbyFollowers: DiscordLobbyFollowerSlice{o},
		}
	} else {
		related.R.DiscordLobbyFollowers = append(related.R.DiscordLobbyFollowers, o)
	}

	return nil
}

// DiscordLobbyFollowers retrieves all the records using an executor.
func DiscordLobbyFollowers(mods ...qm.QueryMod) discordLobbyFollowerQuery {
	mods = append(mods, qm.From("\"discord_lobby_followers\""))
	return discordLobbyFollowerQuery{NewQuery(mods...)}
}

// FindDiscordLobbyFollower retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindDiscordLobbyFollower(exec boil.Executor, iD string, selectCols ...string) (*DiscordLobbyFollower, error) {
	discordLobbyFollowerObj := &DiscordLobbyFollower{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"discord_lobby_followers\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, discordLobbyFollowerObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from discord_lobby_followers")
	}

	if err = discordLobbyFollowerObj.doAfterSelectHooks(exec); err != nil {
		return discordLobbyFollowerObj, err
	}

	return discordLobbyFollowerObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *DiscordLobbyFollower) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no discord_lobby_followers provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(discordLobbyFollowerColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	discordLobbyFollowerInsertCacheMut.RLock()
	cache, cached := discordLobbyFollowerInsertCache[key]
	discordLobbyFollowerInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			discordLobbyFollowerAllColumns,
			discordLobbyFollowerColumnsWithDefault,
			discordLobbyFollowerColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(discordLobbyFollowerType, discordLobbyFollowerMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(discordLobbyFollowerType, discordLobbyFollowerMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"discord_lobby_followers\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"discord_lobby_followers\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into discord_lobby_followers")
	}

	if !cached {
		discordLobbyFollowerInsertCacheMut.Lock()
		discordLobbyFollowerInsertCache[key] = cache
		discordLobbyFollowerInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the DiscordLobbyFollower.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *DiscordLobbyFollower) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	discordLobbyFollowerUpdateCacheMut.RLock()
	cache, cached := discordLobbyFollowerUpdateCache[key]
	discordLobbyFollowerUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			discordLobbyFollowerAllColumns,
			discordLobbyFollowerPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update discord_lobby_followers, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"discord_lobby_followers\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, discordLobbyFollowerPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(discordLobbyFollowerType, discordLobbyFollowerMapping, append(wl, discordLobbyFollowerPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update discord_lobby_followers row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for discord_lobby_followers")
	}

	if !cached {
		discordLobbyFollowerUpdateCacheMut.Lock()
		discordLobbyFollowerUpdateCache[key] = cache
		discordLobbyFollowerUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q discordLobbyFollowerQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for discord_lobby_followers")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for discord_lobby_followers")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o DiscordLobbyFollowerSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), discordLobbyFollowerPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"discord_lobby_followers\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, discordLobbyFollowerPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in discordLobbyFollower slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all discordLobbyFollower")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *DiscordLobbyFollower) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no discord_lobby_followers provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(discordLobbyFollowerColumnsWithDefault, o)

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

	discordLobbyFollowerUpsertCacheMut.RLock()
	cache, cached := discordLobbyFollowerUpsertCache[key]
	discordLobbyFollowerUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			discordLobbyFollowerAllColumns,
			discordLobbyFollowerColumnsWithDefault,
			discordLobbyFollowerColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			discordLobbyFollowerAllColumns,
			discordLobbyFollowerPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert discord_lobby_followers, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(discordLobbyFollowerPrimaryKeyColumns))
			copy(conflict, discordLobbyFollowerPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"discord_lobby_followers\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(discordLobbyFollowerType, discordLobbyFollowerMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(discordLobbyFollowerType, discordLobbyFollowerMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert discord_lobby_followers")
	}

	if !cached {
		discordLobbyFollowerUpsertCacheMut.Lock()
		discordLobbyFollowerUpsertCache[key] = cache
		discordLobbyFollowerUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single DiscordLobbyFollower record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *DiscordLobbyFollower) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no DiscordLobbyFollower provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), discordLobbyFollowerPrimaryKeyMapping)
	sql := "DELETE FROM \"discord_lobby_followers\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from discord_lobby_followers")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for discord_lobby_followers")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q discordLobbyFollowerQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no discordLobbyFollowerQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from discord_lobby_followers")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for discord_lobby_followers")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o DiscordLobbyFollowerSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(discordLobbyFollowerBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), discordLobbyFollowerPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"discord_lobby_followers\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, discordLobbyFollowerPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from discordLobbyFollower slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for discord_lobby_followers")
	}

	if len(discordLobbyFollowerAfterDeleteHooks) != 0 {
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
func (o *DiscordLobbyFollower) Reload(exec boil.Executor) error {
	ret, err := FindDiscordLobbyFollower(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *DiscordLobbyFollowerSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := DiscordLobbyFollowerSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), discordLobbyFollowerPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"discord_lobby_followers\".* FROM \"discord_lobby_followers\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, discordLobbyFollowerPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in DiscordLobbyFollowerSlice")
	}

	*o = slice

	return nil
}

// DiscordLobbyFollowerExists checks if the DiscordLobbyFollower row exists.
func DiscordLobbyFollowerExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"discord_lobby_followers\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if discord_lobby_followers exists")
	}

	return exists, nil
}