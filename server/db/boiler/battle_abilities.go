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

// BattleAbility is an object representing the database table.
type BattleAbility struct {
	ID                     string `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Label                  string `boiler:"label" boil:"label" json:"label" toml:"label" yaml:"label"`
	CooldownDurationSecond int    `boiler:"cooldown_duration_second" boil:"cooldown_duration_second" json:"cooldown_duration_second" toml:"cooldown_duration_second" yaml:"cooldown_duration_second"`
	Description            string `boiler:"description" boil:"description" json:"description" toml:"description" yaml:"description"`

	R *battleAbilityR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L battleAbilityL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var BattleAbilityColumns = struct {
	ID                     string
	Label                  string
	CooldownDurationSecond string
	Description            string
}{
	ID:                     "id",
	Label:                  "label",
	CooldownDurationSecond: "cooldown_duration_second",
	Description:            "description",
}

var BattleAbilityTableColumns = struct {
	ID                     string
	Label                  string
	CooldownDurationSecond string
	Description            string
}{
	ID:                     "battle_abilities.id",
	Label:                  "battle_abilities.label",
	CooldownDurationSecond: "battle_abilities.cooldown_duration_second",
	Description:            "battle_abilities.description",
}

// Generated where

type whereHelperint struct{ field string }

func (w whereHelperint) EQ(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint) NEQ(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint) LT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint) LTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint) GT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint) GTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint) IN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint) NIN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

var BattleAbilityWhere = struct {
	ID                     whereHelperstring
	Label                  whereHelperstring
	CooldownDurationSecond whereHelperint
	Description            whereHelperstring
}{
	ID:                     whereHelperstring{field: "\"battle_abilities\".\"id\""},
	Label:                  whereHelperstring{field: "\"battle_abilities\".\"label\""},
	CooldownDurationSecond: whereHelperint{field: "\"battle_abilities\".\"cooldown_duration_second\""},
	Description:            whereHelperstring{field: "\"battle_abilities\".\"description\""},
}

// BattleAbilityRels is where relationship names are stored.
var BattleAbilityRels = struct {
	GameAbilities string
}{
	GameAbilities: "GameAbilities",
}

// battleAbilityR is where relationships are stored.
type battleAbilityR struct {
	GameAbilities GameAbilitySlice `boiler:"GameAbilities" boil:"GameAbilities" json:"GameAbilities" toml:"GameAbilities" yaml:"GameAbilities"`
}

// NewStruct creates a new relationship struct
func (*battleAbilityR) NewStruct() *battleAbilityR {
	return &battleAbilityR{}
}

// battleAbilityL is where Load methods for each relationship are stored.
type battleAbilityL struct{}

var (
	battleAbilityAllColumns            = []string{"id", "label", "cooldown_duration_second", "description"}
	battleAbilityColumnsWithoutDefault = []string{"label", "cooldown_duration_second", "description"}
	battleAbilityColumnsWithDefault    = []string{"id"}
	battleAbilityPrimaryKeyColumns     = []string{"id"}
	battleAbilityGeneratedColumns      = []string{}
)

type (
	// BattleAbilitySlice is an alias for a slice of pointers to BattleAbility.
	// This should almost always be used instead of []BattleAbility.
	BattleAbilitySlice []*BattleAbility
	// BattleAbilityHook is the signature for custom BattleAbility hook methods
	BattleAbilityHook func(boil.Executor, *BattleAbility) error

	battleAbilityQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	battleAbilityType                 = reflect.TypeOf(&BattleAbility{})
	battleAbilityMapping              = queries.MakeStructMapping(battleAbilityType)
	battleAbilityPrimaryKeyMapping, _ = queries.BindMapping(battleAbilityType, battleAbilityMapping, battleAbilityPrimaryKeyColumns)
	battleAbilityInsertCacheMut       sync.RWMutex
	battleAbilityInsertCache          = make(map[string]insertCache)
	battleAbilityUpdateCacheMut       sync.RWMutex
	battleAbilityUpdateCache          = make(map[string]updateCache)
	battleAbilityUpsertCacheMut       sync.RWMutex
	battleAbilityUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var battleAbilityAfterSelectHooks []BattleAbilityHook

var battleAbilityBeforeInsertHooks []BattleAbilityHook
var battleAbilityAfterInsertHooks []BattleAbilityHook

var battleAbilityBeforeUpdateHooks []BattleAbilityHook
var battleAbilityAfterUpdateHooks []BattleAbilityHook

var battleAbilityBeforeDeleteHooks []BattleAbilityHook
var battleAbilityAfterDeleteHooks []BattleAbilityHook

var battleAbilityBeforeUpsertHooks []BattleAbilityHook
var battleAbilityAfterUpsertHooks []BattleAbilityHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *BattleAbility) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range battleAbilityAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *BattleAbility) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleAbilityBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *BattleAbility) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleAbilityAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *BattleAbility) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range battleAbilityBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *BattleAbility) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range battleAbilityAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *BattleAbility) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range battleAbilityBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *BattleAbility) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range battleAbilityAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *BattleAbility) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleAbilityBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *BattleAbility) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleAbilityAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddBattleAbilityHook registers your hook function for all future operations.
func AddBattleAbilityHook(hookPoint boil.HookPoint, battleAbilityHook BattleAbilityHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		battleAbilityAfterSelectHooks = append(battleAbilityAfterSelectHooks, battleAbilityHook)
	case boil.BeforeInsertHook:
		battleAbilityBeforeInsertHooks = append(battleAbilityBeforeInsertHooks, battleAbilityHook)
	case boil.AfterInsertHook:
		battleAbilityAfterInsertHooks = append(battleAbilityAfterInsertHooks, battleAbilityHook)
	case boil.BeforeUpdateHook:
		battleAbilityBeforeUpdateHooks = append(battleAbilityBeforeUpdateHooks, battleAbilityHook)
	case boil.AfterUpdateHook:
		battleAbilityAfterUpdateHooks = append(battleAbilityAfterUpdateHooks, battleAbilityHook)
	case boil.BeforeDeleteHook:
		battleAbilityBeforeDeleteHooks = append(battleAbilityBeforeDeleteHooks, battleAbilityHook)
	case boil.AfterDeleteHook:
		battleAbilityAfterDeleteHooks = append(battleAbilityAfterDeleteHooks, battleAbilityHook)
	case boil.BeforeUpsertHook:
		battleAbilityBeforeUpsertHooks = append(battleAbilityBeforeUpsertHooks, battleAbilityHook)
	case boil.AfterUpsertHook:
		battleAbilityAfterUpsertHooks = append(battleAbilityAfterUpsertHooks, battleAbilityHook)
	}
}

// One returns a single battleAbility record from the query.
func (q battleAbilityQuery) One(exec boil.Executor) (*BattleAbility, error) {
	o := &BattleAbility{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for battle_abilities")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all BattleAbility records from the query.
func (q battleAbilityQuery) All(exec boil.Executor) (BattleAbilitySlice, error) {
	var o []*BattleAbility

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to BattleAbility slice")
	}

	if len(battleAbilityAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all BattleAbility records in the query.
func (q battleAbilityQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count battle_abilities rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q battleAbilityQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if battle_abilities exists")
	}

	return count > 0, nil
}

// GameAbilities retrieves all the game_ability's GameAbilities with an executor.
func (o *BattleAbility) GameAbilities(mods ...qm.QueryMod) gameAbilityQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"game_abilities\".\"battle_ability_id\"=?", o.ID),
	)

	query := GameAbilities(queryMods...)
	queries.SetFrom(query.Query, "\"game_abilities\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"game_abilities\".*"})
	}

	return query
}

// LoadGameAbilities allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (battleAbilityL) LoadGameAbilities(e boil.Executor, singular bool, maybeBattleAbility interface{}, mods queries.Applicator) error {
	var slice []*BattleAbility
	var object *BattleAbility

	if singular {
		object = maybeBattleAbility.(*BattleAbility)
	} else {
		slice = *maybeBattleAbility.(*[]*BattleAbility)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &battleAbilityR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &battleAbilityR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.ID) {
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
		qm.From(`game_abilities`),
		qm.WhereIn(`game_abilities.battle_ability_id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load game_abilities")
	}

	var resultSlice []*GameAbility
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice game_abilities")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on game_abilities")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for game_abilities")
	}

	if len(gameAbilityAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.GameAbilities = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &gameAbilityR{}
			}
			foreign.R.BattleAbility = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if queries.Equal(local.ID, foreign.BattleAbilityID) {
				local.R.GameAbilities = append(local.R.GameAbilities, foreign)
				if foreign.R == nil {
					foreign.R = &gameAbilityR{}
				}
				foreign.R.BattleAbility = local
				break
			}
		}
	}

	return nil
}

// AddGameAbilities adds the given related objects to the existing relationships
// of the battle_ability, optionally inserting them as new records.
// Appends related to o.R.GameAbilities.
// Sets related.R.BattleAbility appropriately.
func (o *BattleAbility) AddGameAbilities(exec boil.Executor, insert bool, related ...*GameAbility) error {
	var err error
	for _, rel := range related {
		if insert {
			queries.Assign(&rel.BattleAbilityID, o.ID)
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"game_abilities\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"battle_ability_id"}),
				strmangle.WhereClause("\"", "\"", 2, gameAbilityPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			queries.Assign(&rel.BattleAbilityID, o.ID)
		}
	}

	if o.R == nil {
		o.R = &battleAbilityR{
			GameAbilities: related,
		}
	} else {
		o.R.GameAbilities = append(o.R.GameAbilities, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &gameAbilityR{
				BattleAbility: o,
			}
		} else {
			rel.R.BattleAbility = o
		}
	}
	return nil
}

// SetGameAbilities removes all previously related items of the
// battle_ability replacing them completely with the passed
// in related items, optionally inserting them as new records.
// Sets o.R.BattleAbility's GameAbilities accordingly.
// Replaces o.R.GameAbilities with related.
// Sets related.R.BattleAbility's GameAbilities accordingly.
func (o *BattleAbility) SetGameAbilities(exec boil.Executor, insert bool, related ...*GameAbility) error {
	query := "update \"game_abilities\" set \"battle_ability_id\" = null where \"battle_ability_id\" = $1"
	values := []interface{}{o.ID}
	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, query)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	_, err := exec.Exec(query, values...)
	if err != nil {
		return errors.Wrap(err, "failed to remove relationships before set")
	}

	if o.R != nil {
		for _, rel := range o.R.GameAbilities {
			queries.SetScanner(&rel.BattleAbilityID, nil)
			if rel.R == nil {
				continue
			}

			rel.R.BattleAbility = nil
		}

		o.R.GameAbilities = nil
	}
	return o.AddGameAbilities(exec, insert, related...)
}

// RemoveGameAbilities relationships from objects passed in.
// Removes related items from R.GameAbilities (uses pointer comparison, removal does not keep order)
// Sets related.R.BattleAbility.
func (o *BattleAbility) RemoveGameAbilities(exec boil.Executor, related ...*GameAbility) error {
	if len(related) == 0 {
		return nil
	}

	var err error
	for _, rel := range related {
		queries.SetScanner(&rel.BattleAbilityID, nil)
		if rel.R != nil {
			rel.R.BattleAbility = nil
		}
		if _, err = rel.Update(exec, boil.Whitelist("battle_ability_id")); err != nil {
			return err
		}
	}
	if o.R == nil {
		return nil
	}

	for _, rel := range related {
		for i, ri := range o.R.GameAbilities {
			if rel != ri {
				continue
			}

			ln := len(o.R.GameAbilities)
			if ln > 1 && i < ln-1 {
				o.R.GameAbilities[i] = o.R.GameAbilities[ln-1]
			}
			o.R.GameAbilities = o.R.GameAbilities[:ln-1]
			break
		}
	}

	return nil
}

// BattleAbilities retrieves all the records using an executor.
func BattleAbilities(mods ...qm.QueryMod) battleAbilityQuery {
	mods = append(mods, qm.From("\"battle_abilities\""))
	return battleAbilityQuery{NewQuery(mods...)}
}

// FindBattleAbility retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindBattleAbility(exec boil.Executor, iD string, selectCols ...string) (*BattleAbility, error) {
	battleAbilityObj := &BattleAbility{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"battle_abilities\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, battleAbilityObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from battle_abilities")
	}

	if err = battleAbilityObj.doAfterSelectHooks(exec); err != nil {
		return battleAbilityObj, err
	}

	return battleAbilityObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *BattleAbility) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no battle_abilities provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(battleAbilityColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	battleAbilityInsertCacheMut.RLock()
	cache, cached := battleAbilityInsertCache[key]
	battleAbilityInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			battleAbilityAllColumns,
			battleAbilityColumnsWithDefault,
			battleAbilityColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(battleAbilityType, battleAbilityMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(battleAbilityType, battleAbilityMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"battle_abilities\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"battle_abilities\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into battle_abilities")
	}

	if !cached {
		battleAbilityInsertCacheMut.Lock()
		battleAbilityInsertCache[key] = cache
		battleAbilityInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the BattleAbility.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *BattleAbility) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	battleAbilityUpdateCacheMut.RLock()
	cache, cached := battleAbilityUpdateCache[key]
	battleAbilityUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			battleAbilityAllColumns,
			battleAbilityPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update battle_abilities, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"battle_abilities\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, battleAbilityPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(battleAbilityType, battleAbilityMapping, append(wl, battleAbilityPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update battle_abilities row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for battle_abilities")
	}

	if !cached {
		battleAbilityUpdateCacheMut.Lock()
		battleAbilityUpdateCache[key] = cache
		battleAbilityUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q battleAbilityQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for battle_abilities")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for battle_abilities")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o BattleAbilitySlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleAbilityPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"battle_abilities\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, battleAbilityPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in battleAbility slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all battleAbility")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *BattleAbility) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no battle_abilities provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(battleAbilityColumnsWithDefault, o)

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

	battleAbilityUpsertCacheMut.RLock()
	cache, cached := battleAbilityUpsertCache[key]
	battleAbilityUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			battleAbilityAllColumns,
			battleAbilityColumnsWithDefault,
			battleAbilityColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			battleAbilityAllColumns,
			battleAbilityPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert battle_abilities, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(battleAbilityPrimaryKeyColumns))
			copy(conflict, battleAbilityPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"battle_abilities\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(battleAbilityType, battleAbilityMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(battleAbilityType, battleAbilityMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert battle_abilities")
	}

	if !cached {
		battleAbilityUpsertCacheMut.Lock()
		battleAbilityUpsertCache[key] = cache
		battleAbilityUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single BattleAbility record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *BattleAbility) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no BattleAbility provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), battleAbilityPrimaryKeyMapping)
	sql := "DELETE FROM \"battle_abilities\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from battle_abilities")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for battle_abilities")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q battleAbilityQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no battleAbilityQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from battle_abilities")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for battle_abilities")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o BattleAbilitySlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(battleAbilityBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleAbilityPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"battle_abilities\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, battleAbilityPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from battleAbility slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for battle_abilities")
	}

	if len(battleAbilityAfterDeleteHooks) != 0 {
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
func (o *BattleAbility) Reload(exec boil.Executor) error {
	ret, err := FindBattleAbility(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *BattleAbilitySlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := BattleAbilitySlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleAbilityPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"battle_abilities\".* FROM \"battle_abilities\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, battleAbilityPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in BattleAbilitySlice")
	}

	*o = slice

	return nil
}

// BattleAbilityExists checks if the BattleAbility row exists.
func BattleAbilityExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"battle_abilities\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if battle_abilities exists")
	}

	return exists, nil
}
