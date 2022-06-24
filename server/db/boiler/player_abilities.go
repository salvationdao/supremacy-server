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

// PlayerAbility is an object representing the database table.
type PlayerAbility struct {
	ID              string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	OwnerID         string    `boiler:"owner_id" boil:"owner_id" json:"owner_id" toml:"owner_id" yaml:"owner_id"`
	BlueprintID     string    `boiler:"blueprint_id" boil:"blueprint_id" json:"blueprint_id" toml:"blueprint_id" yaml:"blueprint_id"`
	Count           int       `boiler:"count" boil:"count" json:"count" toml:"count" yaml:"count"`
	LastPurchasedAt time.Time `boiler:"last_purchased_at" boil:"last_purchased_at" json:"last_purchased_at" toml:"last_purchased_at" yaml:"last_purchased_at"`

	R *playerAbilityR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L playerAbilityL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var PlayerAbilityColumns = struct {
	ID              string
	OwnerID         string
	BlueprintID     string
	Count           string
	LastPurchasedAt string
}{
	ID:              "id",
	OwnerID:         "owner_id",
	BlueprintID:     "blueprint_id",
	Count:           "count",
	LastPurchasedAt: "last_purchased_at",
}

var PlayerAbilityTableColumns = struct {
	ID              string
	OwnerID         string
	BlueprintID     string
	Count           string
	LastPurchasedAt string
}{
	ID:              "player_abilities.id",
	OwnerID:         "player_abilities.owner_id",
	BlueprintID:     "player_abilities.blueprint_id",
	Count:           "player_abilities.count",
	LastPurchasedAt: "player_abilities.last_purchased_at",
}

// Generated where

var PlayerAbilityWhere = struct {
	ID              whereHelperstring
	OwnerID         whereHelperstring
	BlueprintID     whereHelperstring
	Count           whereHelperint
	LastPurchasedAt whereHelpertime_Time
}{
	ID:              whereHelperstring{field: "\"player_abilities\".\"id\""},
	OwnerID:         whereHelperstring{field: "\"player_abilities\".\"owner_id\""},
	BlueprintID:     whereHelperstring{field: "\"player_abilities\".\"blueprint_id\""},
	Count:           whereHelperint{field: "\"player_abilities\".\"count\""},
	LastPurchasedAt: whereHelpertime_Time{field: "\"player_abilities\".\"last_purchased_at\""},
}

// PlayerAbilityRels is where relationship names are stored.
var PlayerAbilityRels = struct {
	Blueprint string
	Owner     string
}{
	Blueprint: "Blueprint",
	Owner:     "Owner",
}

// playerAbilityR is where relationships are stored.
type playerAbilityR struct {
	Blueprint *BlueprintPlayerAbility `boiler:"Blueprint" boil:"Blueprint" json:"Blueprint" toml:"Blueprint" yaml:"Blueprint"`
	Owner     *Player                 `boiler:"Owner" boil:"Owner" json:"Owner" toml:"Owner" yaml:"Owner"`
}

// NewStruct creates a new relationship struct
func (*playerAbilityR) NewStruct() *playerAbilityR {
	return &playerAbilityR{}
}

// playerAbilityL is where Load methods for each relationship are stored.
type playerAbilityL struct{}

var (
	playerAbilityAllColumns            = []string{"id", "owner_id", "blueprint_id", "count", "last_purchased_at"}
	playerAbilityColumnsWithoutDefault = []string{"owner_id", "blueprint_id"}
	playerAbilityColumnsWithDefault    = []string{"id", "count", "last_purchased_at"}
	playerAbilityPrimaryKeyColumns     = []string{"id"}
	playerAbilityGeneratedColumns      = []string{}
)

type (
	// PlayerAbilitySlice is an alias for a slice of pointers to PlayerAbility.
	// This should almost always be used instead of []PlayerAbility.
	PlayerAbilitySlice []*PlayerAbility
	// PlayerAbilityHook is the signature for custom PlayerAbility hook methods
	PlayerAbilityHook func(boil.Executor, *PlayerAbility) error

	playerAbilityQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	playerAbilityType                 = reflect.TypeOf(&PlayerAbility{})
	playerAbilityMapping              = queries.MakeStructMapping(playerAbilityType)
	playerAbilityPrimaryKeyMapping, _ = queries.BindMapping(playerAbilityType, playerAbilityMapping, playerAbilityPrimaryKeyColumns)
	playerAbilityInsertCacheMut       sync.RWMutex
	playerAbilityInsertCache          = make(map[string]insertCache)
	playerAbilityUpdateCacheMut       sync.RWMutex
	playerAbilityUpdateCache          = make(map[string]updateCache)
	playerAbilityUpsertCacheMut       sync.RWMutex
	playerAbilityUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var playerAbilityAfterSelectHooks []PlayerAbilityHook

var playerAbilityBeforeInsertHooks []PlayerAbilityHook
var playerAbilityAfterInsertHooks []PlayerAbilityHook

var playerAbilityBeforeUpdateHooks []PlayerAbilityHook
var playerAbilityAfterUpdateHooks []PlayerAbilityHook

var playerAbilityBeforeDeleteHooks []PlayerAbilityHook
var playerAbilityAfterDeleteHooks []PlayerAbilityHook

var playerAbilityBeforeUpsertHooks []PlayerAbilityHook
var playerAbilityAfterUpsertHooks []PlayerAbilityHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *PlayerAbility) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAbilityAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *PlayerAbility) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAbilityBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *PlayerAbility) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAbilityAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *PlayerAbility) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAbilityBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *PlayerAbility) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAbilityAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *PlayerAbility) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAbilityBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *PlayerAbility) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAbilityAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *PlayerAbility) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAbilityBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *PlayerAbility) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range playerAbilityAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddPlayerAbilityHook registers your hook function for all future operations.
func AddPlayerAbilityHook(hookPoint boil.HookPoint, playerAbilityHook PlayerAbilityHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		playerAbilityAfterSelectHooks = append(playerAbilityAfterSelectHooks, playerAbilityHook)
	case boil.BeforeInsertHook:
		playerAbilityBeforeInsertHooks = append(playerAbilityBeforeInsertHooks, playerAbilityHook)
	case boil.AfterInsertHook:
		playerAbilityAfterInsertHooks = append(playerAbilityAfterInsertHooks, playerAbilityHook)
	case boil.BeforeUpdateHook:
		playerAbilityBeforeUpdateHooks = append(playerAbilityBeforeUpdateHooks, playerAbilityHook)
	case boil.AfterUpdateHook:
		playerAbilityAfterUpdateHooks = append(playerAbilityAfterUpdateHooks, playerAbilityHook)
	case boil.BeforeDeleteHook:
		playerAbilityBeforeDeleteHooks = append(playerAbilityBeforeDeleteHooks, playerAbilityHook)
	case boil.AfterDeleteHook:
		playerAbilityAfterDeleteHooks = append(playerAbilityAfterDeleteHooks, playerAbilityHook)
	case boil.BeforeUpsertHook:
		playerAbilityBeforeUpsertHooks = append(playerAbilityBeforeUpsertHooks, playerAbilityHook)
	case boil.AfterUpsertHook:
		playerAbilityAfterUpsertHooks = append(playerAbilityAfterUpsertHooks, playerAbilityHook)
	}
}

// One returns a single playerAbility record from the query.
func (q playerAbilityQuery) One(exec boil.Executor) (*PlayerAbility, error) {
	o := &PlayerAbility{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for player_abilities")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all PlayerAbility records from the query.
func (q playerAbilityQuery) All(exec boil.Executor) (PlayerAbilitySlice, error) {
	var o []*PlayerAbility

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to PlayerAbility slice")
	}

	if len(playerAbilityAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all PlayerAbility records in the query.
func (q playerAbilityQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count player_abilities rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q playerAbilityQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if player_abilities exists")
	}

	return count > 0, nil
}

// Blueprint pointed to by the foreign key.
func (o *PlayerAbility) Blueprint(mods ...qm.QueryMod) blueprintPlayerAbilityQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BlueprintID),
	}

	queryMods = append(queryMods, mods...)

	query := BlueprintPlayerAbilities(queryMods...)
	queries.SetFrom(query.Query, "\"blueprint_player_abilities\"")

	return query
}

// Owner pointed to by the foreign key.
func (o *PlayerAbility) Owner(mods ...qm.QueryMod) playerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.OwnerID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Players(queryMods...)
	queries.SetFrom(query.Query, "\"players\"")

	return query
}

// LoadBlueprint allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playerAbilityL) LoadBlueprint(e boil.Executor, singular bool, maybePlayerAbility interface{}, mods queries.Applicator) error {
	var slice []*PlayerAbility
	var object *PlayerAbility

	if singular {
		object = maybePlayerAbility.(*PlayerAbility)
	} else {
		slice = *maybePlayerAbility.(*[]*PlayerAbility)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playerAbilityR{}
		}
		args = append(args, object.BlueprintID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playerAbilityR{}
			}

			for _, a := range args {
				if a == obj.BlueprintID {
					continue Outer
				}
			}

			args = append(args, obj.BlueprintID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`blueprint_player_abilities`),
		qm.WhereIn(`blueprint_player_abilities.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load BlueprintPlayerAbility")
	}

	var resultSlice []*BlueprintPlayerAbility
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice BlueprintPlayerAbility")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for blueprint_player_abilities")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for blueprint_player_abilities")
	}

	if len(playerAbilityAfterSelectHooks) != 0 {
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
		object.R.Blueprint = foreign
		if foreign.R == nil {
			foreign.R = &blueprintPlayerAbilityR{}
		}
		foreign.R.BlueprintPlayerAbilities = append(foreign.R.BlueprintPlayerAbilities, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.BlueprintID == foreign.ID {
				local.R.Blueprint = foreign
				if foreign.R == nil {
					foreign.R = &blueprintPlayerAbilityR{}
				}
				foreign.R.BlueprintPlayerAbilities = append(foreign.R.BlueprintPlayerAbilities, local)
				break
			}
		}
	}

	return nil
}

// LoadOwner allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (playerAbilityL) LoadOwner(e boil.Executor, singular bool, maybePlayerAbility interface{}, mods queries.Applicator) error {
	var slice []*PlayerAbility
	var object *PlayerAbility

	if singular {
		object = maybePlayerAbility.(*PlayerAbility)
	} else {
		slice = *maybePlayerAbility.(*[]*PlayerAbility)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &playerAbilityR{}
		}
		args = append(args, object.OwnerID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &playerAbilityR{}
			}

			for _, a := range args {
				if a == obj.OwnerID {
					continue Outer
				}
			}

			args = append(args, obj.OwnerID)

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

	if len(playerAbilityAfterSelectHooks) != 0 {
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
		object.R.Owner = foreign
		if foreign.R == nil {
			foreign.R = &playerR{}
		}
		foreign.R.OwnerPlayerAbilities = append(foreign.R.OwnerPlayerAbilities, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.OwnerID == foreign.ID {
				local.R.Owner = foreign
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.OwnerPlayerAbilities = append(foreign.R.OwnerPlayerAbilities, local)
				break
			}
		}
	}

	return nil
}

// SetBlueprint of the playerAbility to the related item.
// Sets o.R.Blueprint to related.
// Adds o to related.R.BlueprintPlayerAbilities.
func (o *PlayerAbility) SetBlueprint(exec boil.Executor, insert bool, related *BlueprintPlayerAbility) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"player_abilities\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"blueprint_id"}),
		strmangle.WhereClause("\"", "\"", 2, playerAbilityPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.BlueprintID = related.ID
	if o.R == nil {
		o.R = &playerAbilityR{
			Blueprint: related,
		}
	} else {
		o.R.Blueprint = related
	}

	if related.R == nil {
		related.R = &blueprintPlayerAbilityR{
			BlueprintPlayerAbilities: PlayerAbilitySlice{o},
		}
	} else {
		related.R.BlueprintPlayerAbilities = append(related.R.BlueprintPlayerAbilities, o)
	}

	return nil
}

// SetOwner of the playerAbility to the related item.
// Sets o.R.Owner to related.
// Adds o to related.R.OwnerPlayerAbilities.
func (o *PlayerAbility) SetOwner(exec boil.Executor, insert bool, related *Player) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"player_abilities\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"owner_id"}),
		strmangle.WhereClause("\"", "\"", 2, playerAbilityPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.OwnerID = related.ID
	if o.R == nil {
		o.R = &playerAbilityR{
			Owner: related,
		}
	} else {
		o.R.Owner = related
	}

	if related.R == nil {
		related.R = &playerR{
			OwnerPlayerAbilities: PlayerAbilitySlice{o},
		}
	} else {
		related.R.OwnerPlayerAbilities = append(related.R.OwnerPlayerAbilities, o)
	}

	return nil
}

// PlayerAbilities retrieves all the records using an executor.
func PlayerAbilities(mods ...qm.QueryMod) playerAbilityQuery {
	mods = append(mods, qm.From("\"player_abilities\""))
	return playerAbilityQuery{NewQuery(mods...)}
}

// FindPlayerAbility retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindPlayerAbility(exec boil.Executor, iD string, selectCols ...string) (*PlayerAbility, error) {
	playerAbilityObj := &PlayerAbility{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"player_abilities\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, playerAbilityObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from player_abilities")
	}

	if err = playerAbilityObj.doAfterSelectHooks(exec); err != nil {
		return playerAbilityObj, err
	}

	return playerAbilityObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *PlayerAbility) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no player_abilities provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(playerAbilityColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	playerAbilityInsertCacheMut.RLock()
	cache, cached := playerAbilityInsertCache[key]
	playerAbilityInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			playerAbilityAllColumns,
			playerAbilityColumnsWithDefault,
			playerAbilityColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(playerAbilityType, playerAbilityMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(playerAbilityType, playerAbilityMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"player_abilities\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"player_abilities\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into player_abilities")
	}

	if !cached {
		playerAbilityInsertCacheMut.Lock()
		playerAbilityInsertCache[key] = cache
		playerAbilityInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the PlayerAbility.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *PlayerAbility) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	playerAbilityUpdateCacheMut.RLock()
	cache, cached := playerAbilityUpdateCache[key]
	playerAbilityUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			playerAbilityAllColumns,
			playerAbilityPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update player_abilities, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"player_abilities\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, playerAbilityPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(playerAbilityType, playerAbilityMapping, append(wl, playerAbilityPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update player_abilities row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for player_abilities")
	}

	if !cached {
		playerAbilityUpdateCacheMut.Lock()
		playerAbilityUpdateCache[key] = cache
		playerAbilityUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q playerAbilityQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for player_abilities")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for player_abilities")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o PlayerAbilitySlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerAbilityPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"player_abilities\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, playerAbilityPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in playerAbility slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all playerAbility")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *PlayerAbility) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no player_abilities provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(playerAbilityColumnsWithDefault, o)

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

	playerAbilityUpsertCacheMut.RLock()
	cache, cached := playerAbilityUpsertCache[key]
	playerAbilityUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			playerAbilityAllColumns,
			playerAbilityColumnsWithDefault,
			playerAbilityColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			playerAbilityAllColumns,
			playerAbilityPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert player_abilities, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(playerAbilityPrimaryKeyColumns))
			copy(conflict, playerAbilityPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"player_abilities\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(playerAbilityType, playerAbilityMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(playerAbilityType, playerAbilityMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert player_abilities")
	}

	if !cached {
		playerAbilityUpsertCacheMut.Lock()
		playerAbilityUpsertCache[key] = cache
		playerAbilityUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single PlayerAbility record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *PlayerAbility) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no PlayerAbility provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), playerAbilityPrimaryKeyMapping)
	sql := "DELETE FROM \"player_abilities\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from player_abilities")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for player_abilities")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q playerAbilityQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no playerAbilityQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from player_abilities")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for player_abilities")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o PlayerAbilitySlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(playerAbilityBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerAbilityPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"player_abilities\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playerAbilityPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from playerAbility slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for player_abilities")
	}

	if len(playerAbilityAfterDeleteHooks) != 0 {
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
func (o *PlayerAbility) Reload(exec boil.Executor) error {
	ret, err := FindPlayerAbility(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *PlayerAbilitySlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := PlayerAbilitySlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), playerAbilityPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"player_abilities\".* FROM \"player_abilities\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, playerAbilityPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in PlayerAbilitySlice")
	}

	*o = slice

	return nil
}

// PlayerAbilityExists checks if the PlayerAbility row exists.
func PlayerAbilityExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"player_abilities\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if player_abilities exists")
	}

	return exists, nil
}
