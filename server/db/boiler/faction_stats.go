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
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// FactionStat is an object representing the database table.
type FactionStat struct {
	ID             string          `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	WinCount       int             `boiler:"win_count" boil:"win_count" json:"win_count" toml:"win_count" yaml:"win_count"`
	LossCount      int             `boiler:"loss_count" boil:"loss_count" json:"loss_count" toml:"loss_count" yaml:"loss_count"`
	KillCount      int             `boiler:"kill_count" boil:"kill_count" json:"kill_count" toml:"kill_count" yaml:"kill_count"`
	DeathCount     int             `boiler:"death_count" boil:"death_count" json:"death_count" toml:"death_count" yaml:"death_count"`
	SupsContribute decimal.Decimal `boiler:"sups_contribute" boil:"sups_contribute" json:"sups_contribute" toml:"sups_contribute" yaml:"sups_contribute"`
	MVPPlayerID    null.String     `boiler:"mvp_player_id" boil:"mvp_player_id" json:"mvp_player_id,omitempty" toml:"mvp_player_id" yaml:"mvp_player_id,omitempty"`
	MechKillCount  int             `boiler:"mech_kill_count" boil:"mech_kill_count" json:"mech_kill_count" toml:"mech_kill_count" yaml:"mech_kill_count"`

	R *factionStatR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L factionStatL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var FactionStatColumns = struct {
	ID             string
	WinCount       string
	LossCount      string
	KillCount      string
	DeathCount     string
	SupsContribute string
	MVPPlayerID    string
	MechKillCount  string
}{
	ID:             "id",
	WinCount:       "win_count",
	LossCount:      "loss_count",
	KillCount:      "kill_count",
	DeathCount:     "death_count",
	SupsContribute: "sups_contribute",
	MVPPlayerID:    "mvp_player_id",
	MechKillCount:  "mech_kill_count",
}

var FactionStatTableColumns = struct {
	ID             string
	WinCount       string
	LossCount      string
	KillCount      string
	DeathCount     string
	SupsContribute string
	MVPPlayerID    string
	MechKillCount  string
}{
	ID:             "faction_stats.id",
	WinCount:       "faction_stats.win_count",
	LossCount:      "faction_stats.loss_count",
	KillCount:      "faction_stats.kill_count",
	DeathCount:     "faction_stats.death_count",
	SupsContribute: "faction_stats.sups_contribute",
	MVPPlayerID:    "faction_stats.mvp_player_id",
	MechKillCount:  "faction_stats.mech_kill_count",
}

// Generated where

var FactionStatWhere = struct {
	ID             whereHelperstring
	WinCount       whereHelperint
	LossCount      whereHelperint
	KillCount      whereHelperint
	DeathCount     whereHelperint
	SupsContribute whereHelperdecimal_Decimal
	MVPPlayerID    whereHelpernull_String
	MechKillCount  whereHelperint
}{
	ID:             whereHelperstring{field: "\"faction_stats\".\"id\""},
	WinCount:       whereHelperint{field: "\"faction_stats\".\"win_count\""},
	LossCount:      whereHelperint{field: "\"faction_stats\".\"loss_count\""},
	KillCount:      whereHelperint{field: "\"faction_stats\".\"kill_count\""},
	DeathCount:     whereHelperint{field: "\"faction_stats\".\"death_count\""},
	SupsContribute: whereHelperdecimal_Decimal{field: "\"faction_stats\".\"sups_contribute\""},
	MVPPlayerID:    whereHelpernull_String{field: "\"faction_stats\".\"mvp_player_id\""},
	MechKillCount:  whereHelperint{field: "\"faction_stats\".\"mech_kill_count\""},
}

// FactionStatRels is where relationship names are stored.
var FactionStatRels = struct {
	IDFaction string
	MVPPlayer string
}{
	IDFaction: "IDFaction",
	MVPPlayer: "MVPPlayer",
}

// factionStatR is where relationships are stored.
type factionStatR struct {
	IDFaction *Faction `boiler:"IDFaction" boil:"IDFaction" json:"IDFaction" toml:"IDFaction" yaml:"IDFaction"`
	MVPPlayer *Player  `boiler:"MVPPlayer" boil:"MVPPlayer" json:"MVPPlayer" toml:"MVPPlayer" yaml:"MVPPlayer"`
}

// NewStruct creates a new relationship struct
func (*factionStatR) NewStruct() *factionStatR {
	return &factionStatR{}
}

// factionStatL is where Load methods for each relationship are stored.
type factionStatL struct{}

var (
	factionStatAllColumns            = []string{"id", "win_count", "loss_count", "kill_count", "death_count", "sups_contribute", "mvp_player_id", "mech_kill_count"}
	factionStatColumnsWithoutDefault = []string{"id"}
	factionStatColumnsWithDefault    = []string{"win_count", "loss_count", "kill_count", "death_count", "sups_contribute", "mvp_player_id", "mech_kill_count"}
	factionStatPrimaryKeyColumns     = []string{"id"}
	factionStatGeneratedColumns      = []string{}
)

type (
	// FactionStatSlice is an alias for a slice of pointers to FactionStat.
	// This should almost always be used instead of []FactionStat.
	FactionStatSlice []*FactionStat
	// FactionStatHook is the signature for custom FactionStat hook methods
	FactionStatHook func(boil.Executor, *FactionStat) error

	factionStatQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	factionStatType                 = reflect.TypeOf(&FactionStat{})
	factionStatMapping              = queries.MakeStructMapping(factionStatType)
	factionStatPrimaryKeyMapping, _ = queries.BindMapping(factionStatType, factionStatMapping, factionStatPrimaryKeyColumns)
	factionStatInsertCacheMut       sync.RWMutex
	factionStatInsertCache          = make(map[string]insertCache)
	factionStatUpdateCacheMut       sync.RWMutex
	factionStatUpdateCache          = make(map[string]updateCache)
	factionStatUpsertCacheMut       sync.RWMutex
	factionStatUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var factionStatAfterSelectHooks []FactionStatHook

var factionStatBeforeInsertHooks []FactionStatHook
var factionStatAfterInsertHooks []FactionStatHook

var factionStatBeforeUpdateHooks []FactionStatHook
var factionStatAfterUpdateHooks []FactionStatHook

var factionStatBeforeDeleteHooks []FactionStatHook
var factionStatAfterDeleteHooks []FactionStatHook

var factionStatBeforeUpsertHooks []FactionStatHook
var factionStatAfterUpsertHooks []FactionStatHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *FactionStat) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range factionStatAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *FactionStat) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range factionStatBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *FactionStat) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range factionStatAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *FactionStat) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range factionStatBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *FactionStat) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range factionStatAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *FactionStat) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range factionStatBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *FactionStat) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range factionStatAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *FactionStat) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range factionStatBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *FactionStat) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range factionStatAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddFactionStatHook registers your hook function for all future operations.
func AddFactionStatHook(hookPoint boil.HookPoint, factionStatHook FactionStatHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		factionStatAfterSelectHooks = append(factionStatAfterSelectHooks, factionStatHook)
	case boil.BeforeInsertHook:
		factionStatBeforeInsertHooks = append(factionStatBeforeInsertHooks, factionStatHook)
	case boil.AfterInsertHook:
		factionStatAfterInsertHooks = append(factionStatAfterInsertHooks, factionStatHook)
	case boil.BeforeUpdateHook:
		factionStatBeforeUpdateHooks = append(factionStatBeforeUpdateHooks, factionStatHook)
	case boil.AfterUpdateHook:
		factionStatAfterUpdateHooks = append(factionStatAfterUpdateHooks, factionStatHook)
	case boil.BeforeDeleteHook:
		factionStatBeforeDeleteHooks = append(factionStatBeforeDeleteHooks, factionStatHook)
	case boil.AfterDeleteHook:
		factionStatAfterDeleteHooks = append(factionStatAfterDeleteHooks, factionStatHook)
	case boil.BeforeUpsertHook:
		factionStatBeforeUpsertHooks = append(factionStatBeforeUpsertHooks, factionStatHook)
	case boil.AfterUpsertHook:
		factionStatAfterUpsertHooks = append(factionStatAfterUpsertHooks, factionStatHook)
	}
}

// One returns a single factionStat record from the query.
func (q factionStatQuery) One(exec boil.Executor) (*FactionStat, error) {
	o := &FactionStat{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for faction_stats")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all FactionStat records from the query.
func (q factionStatQuery) All(exec boil.Executor) (FactionStatSlice, error) {
	var o []*FactionStat

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to FactionStat slice")
	}

	if len(factionStatAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all FactionStat records in the query.
func (q factionStatQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count faction_stats rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q factionStatQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if faction_stats exists")
	}

	return count > 0, nil
}

// IDFaction pointed to by the foreign key.
func (o *FactionStat) IDFaction(mods ...qm.QueryMod) factionQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.ID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Factions(queryMods...)
	queries.SetFrom(query.Query, "\"factions\"")

	return query
}

// MVPPlayer pointed to by the foreign key.
func (o *FactionStat) MVPPlayer(mods ...qm.QueryMod) playerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.MVPPlayerID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Players(queryMods...)
	queries.SetFrom(query.Query, "\"players\"")

	return query
}

// LoadIDFaction allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (factionStatL) LoadIDFaction(e boil.Executor, singular bool, maybeFactionStat interface{}, mods queries.Applicator) error {
	var slice []*FactionStat
	var object *FactionStat

	if singular {
		object = maybeFactionStat.(*FactionStat)
	} else {
		slice = *maybeFactionStat.(*[]*FactionStat)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &factionStatR{}
		}
		args = append(args, object.ID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &factionStatR{}
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

	if len(factionStatAfterSelectHooks) != 0 {
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
		object.R.IDFaction = foreign
		if foreign.R == nil {
			foreign.R = &factionR{}
		}
		foreign.R.IDFactionStat = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.ID == foreign.ID {
				local.R.IDFaction = foreign
				if foreign.R == nil {
					foreign.R = &factionR{}
				}
				foreign.R.IDFactionStat = local
				break
			}
		}
	}

	return nil
}

// LoadMVPPlayer allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (factionStatL) LoadMVPPlayer(e boil.Executor, singular bool, maybeFactionStat interface{}, mods queries.Applicator) error {
	var slice []*FactionStat
	var object *FactionStat

	if singular {
		object = maybeFactionStat.(*FactionStat)
	} else {
		slice = *maybeFactionStat.(*[]*FactionStat)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &factionStatR{}
		}
		if !queries.IsNil(object.MVPPlayerID) {
			args = append(args, object.MVPPlayerID)
		}

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &factionStatR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.MVPPlayerID) {
					continue Outer
				}
			}

			if !queries.IsNil(obj.MVPPlayerID) {
				args = append(args, obj.MVPPlayerID)
			}

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

	if len(factionStatAfterSelectHooks) != 0 {
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
		object.R.MVPPlayer = foreign
		if foreign.R == nil {
			foreign.R = &playerR{}
		}
		foreign.R.MVPPlayerFactionStats = append(foreign.R.MVPPlayerFactionStats, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.MVPPlayerID, foreign.ID) {
				local.R.MVPPlayer = foreign
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.MVPPlayerFactionStats = append(foreign.R.MVPPlayerFactionStats, local)
				break
			}
		}
	}

	return nil
}

// SetIDFaction of the factionStat to the related item.
// Sets o.R.IDFaction to related.
// Adds o to related.R.IDFactionStat.
func (o *FactionStat) SetIDFaction(exec boil.Executor, insert bool, related *Faction) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"faction_stats\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"id"}),
		strmangle.WhereClause("\"", "\"", 2, factionStatPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.ID = related.ID
	if o.R == nil {
		o.R = &factionStatR{
			IDFaction: related,
		}
	} else {
		o.R.IDFaction = related
	}

	if related.R == nil {
		related.R = &factionR{
			IDFactionStat: o,
		}
	} else {
		related.R.IDFactionStat = o
	}

	return nil
}

// SetMVPPlayer of the factionStat to the related item.
// Sets o.R.MVPPlayer to related.
// Adds o to related.R.MVPPlayerFactionStats.
func (o *FactionStat) SetMVPPlayer(exec boil.Executor, insert bool, related *Player) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"faction_stats\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"mvp_player_id"}),
		strmangle.WhereClause("\"", "\"", 2, factionStatPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	queries.Assign(&o.MVPPlayerID, related.ID)
	if o.R == nil {
		o.R = &factionStatR{
			MVPPlayer: related,
		}
	} else {
		o.R.MVPPlayer = related
	}

	if related.R == nil {
		related.R = &playerR{
			MVPPlayerFactionStats: FactionStatSlice{o},
		}
	} else {
		related.R.MVPPlayerFactionStats = append(related.R.MVPPlayerFactionStats, o)
	}

	return nil
}

// RemoveMVPPlayer relationship.
// Sets o.R.MVPPlayer to nil.
// Removes o from all passed in related items' relationships struct (Optional).
func (o *FactionStat) RemoveMVPPlayer(exec boil.Executor, related *Player) error {
	var err error

	queries.SetScanner(&o.MVPPlayerID, nil)
	if _, err = o.Update(exec, boil.Whitelist("mvp_player_id")); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	if o.R != nil {
		o.R.MVPPlayer = nil
	}
	if related == nil || related.R == nil {
		return nil
	}

	for i, ri := range related.R.MVPPlayerFactionStats {
		if queries.Equal(o.MVPPlayerID, ri.MVPPlayerID) {
			continue
		}

		ln := len(related.R.MVPPlayerFactionStats)
		if ln > 1 && i < ln-1 {
			related.R.MVPPlayerFactionStats[i] = related.R.MVPPlayerFactionStats[ln-1]
		}
		related.R.MVPPlayerFactionStats = related.R.MVPPlayerFactionStats[:ln-1]
		break
	}
	return nil
}

// FactionStats retrieves all the records using an executor.
func FactionStats(mods ...qm.QueryMod) factionStatQuery {
	mods = append(mods, qm.From("\"faction_stats\""))
	return factionStatQuery{NewQuery(mods...)}
}

// FindFactionStat retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindFactionStat(exec boil.Executor, iD string, selectCols ...string) (*FactionStat, error) {
	factionStatObj := &FactionStat{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"faction_stats\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, factionStatObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from faction_stats")
	}

	if err = factionStatObj.doAfterSelectHooks(exec); err != nil {
		return factionStatObj, err
	}

	return factionStatObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *FactionStat) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no faction_stats provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(factionStatColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	factionStatInsertCacheMut.RLock()
	cache, cached := factionStatInsertCache[key]
	factionStatInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			factionStatAllColumns,
			factionStatColumnsWithDefault,
			factionStatColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(factionStatType, factionStatMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(factionStatType, factionStatMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"faction_stats\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"faction_stats\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into faction_stats")
	}

	if !cached {
		factionStatInsertCacheMut.Lock()
		factionStatInsertCache[key] = cache
		factionStatInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the FactionStat.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *FactionStat) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	factionStatUpdateCacheMut.RLock()
	cache, cached := factionStatUpdateCache[key]
	factionStatUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			factionStatAllColumns,
			factionStatPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update faction_stats, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"faction_stats\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, factionStatPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(factionStatType, factionStatMapping, append(wl, factionStatPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update faction_stats row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for faction_stats")
	}

	if !cached {
		factionStatUpdateCacheMut.Lock()
		factionStatUpdateCache[key] = cache
		factionStatUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q factionStatQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for faction_stats")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for faction_stats")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o FactionStatSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), factionStatPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"faction_stats\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, factionStatPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in factionStat slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all factionStat")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *FactionStat) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no faction_stats provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(factionStatColumnsWithDefault, o)

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

	factionStatUpsertCacheMut.RLock()
	cache, cached := factionStatUpsertCache[key]
	factionStatUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			factionStatAllColumns,
			factionStatColumnsWithDefault,
			factionStatColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			factionStatAllColumns,
			factionStatPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert faction_stats, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(factionStatPrimaryKeyColumns))
			copy(conflict, factionStatPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"faction_stats\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(factionStatType, factionStatMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(factionStatType, factionStatMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert faction_stats")
	}

	if !cached {
		factionStatUpsertCacheMut.Lock()
		factionStatUpsertCache[key] = cache
		factionStatUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single FactionStat record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *FactionStat) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no FactionStat provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), factionStatPrimaryKeyMapping)
	sql := "DELETE FROM \"faction_stats\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from faction_stats")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for faction_stats")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q factionStatQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no factionStatQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from faction_stats")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for faction_stats")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o FactionStatSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(factionStatBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), factionStatPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"faction_stats\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, factionStatPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from factionStat slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for faction_stats")
	}

	if len(factionStatAfterDeleteHooks) != 0 {
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
func (o *FactionStat) Reload(exec boil.Executor) error {
	ret, err := FindFactionStat(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *FactionStatSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := FactionStatSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), factionStatPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"faction_stats\".* FROM \"faction_stats\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, factionStatPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in FactionStatSlice")
	}

	*o = slice

	return nil
}

// FactionStatExists checks if the FactionStat row exists.
func FactionStatExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"faction_stats\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if faction_stats exists")
	}

	return exists, nil
}
