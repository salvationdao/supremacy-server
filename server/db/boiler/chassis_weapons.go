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

// ChassisWeapon is an object representing the database table.
type ChassisWeapon struct {
	ID            string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	ChassisID     string    `boiler:"chassis_id" boil:"chassis_id" json:"chassisID" toml:"chassisID" yaml:"chassisID"`
	WeaponID      string    `boiler:"weapon_id" boil:"weapon_id" json:"weaponID" toml:"weaponID" yaml:"weaponID"`
	SlotNumber    int       `boiler:"slot_number" boil:"slot_number" json:"slotNumber" toml:"slotNumber" yaml:"slotNumber"`
	MountLocation string    `boiler:"mount_location" boil:"mount_location" json:"mountLocation" toml:"mountLocation" yaml:"mountLocation"`
	DeletedAt     null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deletedAt,omitempty" toml:"deletedAt" yaml:"deletedAt,omitempty"`
	UpdatedAt     time.Time `boiler:"updated_at" boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
	CreatedAt     time.Time `boiler:"created_at" boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`

	R *chassisWeaponR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L chassisWeaponL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ChassisWeaponColumns = struct {
	ID            string
	ChassisID     string
	WeaponID      string
	SlotNumber    string
	MountLocation string
	DeletedAt     string
	UpdatedAt     string
	CreatedAt     string
}{
	ID:            "id",
	ChassisID:     "chassis_id",
	WeaponID:      "weapon_id",
	SlotNumber:    "slot_number",
	MountLocation: "mount_location",
	DeletedAt:     "deleted_at",
	UpdatedAt:     "updated_at",
	CreatedAt:     "created_at",
}

var ChassisWeaponTableColumns = struct {
	ID            string
	ChassisID     string
	WeaponID      string
	SlotNumber    string
	MountLocation string
	DeletedAt     string
	UpdatedAt     string
	CreatedAt     string
}{
	ID:            "chassis_weapons.id",
	ChassisID:     "chassis_weapons.chassis_id",
	WeaponID:      "chassis_weapons.weapon_id",
	SlotNumber:    "chassis_weapons.slot_number",
	MountLocation: "chassis_weapons.mount_location",
	DeletedAt:     "chassis_weapons.deleted_at",
	UpdatedAt:     "chassis_weapons.updated_at",
	CreatedAt:     "chassis_weapons.created_at",
}

// Generated where

var ChassisWeaponWhere = struct {
	ID            whereHelperstring
	ChassisID     whereHelperstring
	WeaponID      whereHelperstring
	SlotNumber    whereHelperint
	MountLocation whereHelperstring
	DeletedAt     whereHelpernull_Time
	UpdatedAt     whereHelpertime_Time
	CreatedAt     whereHelpertime_Time
}{
	ID:            whereHelperstring{field: "\"chassis_weapons\".\"id\""},
	ChassisID:     whereHelperstring{field: "\"chassis_weapons\".\"chassis_id\""},
	WeaponID:      whereHelperstring{field: "\"chassis_weapons\".\"weapon_id\""},
	SlotNumber:    whereHelperint{field: "\"chassis_weapons\".\"slot_number\""},
	MountLocation: whereHelperstring{field: "\"chassis_weapons\".\"mount_location\""},
	DeletedAt:     whereHelpernull_Time{field: "\"chassis_weapons\".\"deleted_at\""},
	UpdatedAt:     whereHelpertime_Time{field: "\"chassis_weapons\".\"updated_at\""},
	CreatedAt:     whereHelpertime_Time{field: "\"chassis_weapons\".\"created_at\""},
}

// ChassisWeaponRels is where relationship names are stored.
var ChassisWeaponRels = struct {
	Chassis string
	Weapon  string
}{
	Chassis: "Chassis",
	Weapon:  "Weapon",
}

// chassisWeaponR is where relationships are stored.
type chassisWeaponR struct {
	Chassis *Chassis `boiler:"Chassis" boil:"Chassis" json:"Chassis" toml:"Chassis" yaml:"Chassis"`
	Weapon  *Weapon  `boiler:"Weapon" boil:"Weapon" json:"Weapon" toml:"Weapon" yaml:"Weapon"`
}

// NewStruct creates a new relationship struct
func (*chassisWeaponR) NewStruct() *chassisWeaponR {
	return &chassisWeaponR{}
}

// chassisWeaponL is where Load methods for each relationship are stored.
type chassisWeaponL struct{}

var (
	chassisWeaponAllColumns            = []string{"id", "chassis_id", "weapon_id", "slot_number", "mount_location", "deleted_at", "updated_at", "created_at"}
	chassisWeaponColumnsWithoutDefault = []string{"chassis_id", "weapon_id", "slot_number", "mount_location"}
	chassisWeaponColumnsWithDefault    = []string{"id", "deleted_at", "updated_at", "created_at"}
	chassisWeaponPrimaryKeyColumns     = []string{"id"}
	chassisWeaponGeneratedColumns      = []string{}
)

type (
	// ChassisWeaponSlice is an alias for a slice of pointers to ChassisWeapon.
	// This should almost always be used instead of []ChassisWeapon.
	ChassisWeaponSlice []*ChassisWeapon
	// ChassisWeaponHook is the signature for custom ChassisWeapon hook methods
	ChassisWeaponHook func(boil.Executor, *ChassisWeapon) error

	chassisWeaponQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	chassisWeaponType                 = reflect.TypeOf(&ChassisWeapon{})
	chassisWeaponMapping              = queries.MakeStructMapping(chassisWeaponType)
	chassisWeaponPrimaryKeyMapping, _ = queries.BindMapping(chassisWeaponType, chassisWeaponMapping, chassisWeaponPrimaryKeyColumns)
	chassisWeaponInsertCacheMut       sync.RWMutex
	chassisWeaponInsertCache          = make(map[string]insertCache)
	chassisWeaponUpdateCacheMut       sync.RWMutex
	chassisWeaponUpdateCache          = make(map[string]updateCache)
	chassisWeaponUpsertCacheMut       sync.RWMutex
	chassisWeaponUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var chassisWeaponAfterSelectHooks []ChassisWeaponHook

var chassisWeaponBeforeInsertHooks []ChassisWeaponHook
var chassisWeaponAfterInsertHooks []ChassisWeaponHook

var chassisWeaponBeforeUpdateHooks []ChassisWeaponHook
var chassisWeaponAfterUpdateHooks []ChassisWeaponHook

var chassisWeaponBeforeDeleteHooks []ChassisWeaponHook
var chassisWeaponAfterDeleteHooks []ChassisWeaponHook

var chassisWeaponBeforeUpsertHooks []ChassisWeaponHook
var chassisWeaponAfterUpsertHooks []ChassisWeaponHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *ChassisWeapon) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range chassisWeaponAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *ChassisWeapon) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range chassisWeaponBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *ChassisWeapon) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range chassisWeaponAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *ChassisWeapon) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range chassisWeaponBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *ChassisWeapon) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range chassisWeaponAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *ChassisWeapon) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range chassisWeaponBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *ChassisWeapon) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range chassisWeaponAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *ChassisWeapon) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range chassisWeaponBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *ChassisWeapon) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range chassisWeaponAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddChassisWeaponHook registers your hook function for all future operations.
func AddChassisWeaponHook(hookPoint boil.HookPoint, chassisWeaponHook ChassisWeaponHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		chassisWeaponAfterSelectHooks = append(chassisWeaponAfterSelectHooks, chassisWeaponHook)
	case boil.BeforeInsertHook:
		chassisWeaponBeforeInsertHooks = append(chassisWeaponBeforeInsertHooks, chassisWeaponHook)
	case boil.AfterInsertHook:
		chassisWeaponAfterInsertHooks = append(chassisWeaponAfterInsertHooks, chassisWeaponHook)
	case boil.BeforeUpdateHook:
		chassisWeaponBeforeUpdateHooks = append(chassisWeaponBeforeUpdateHooks, chassisWeaponHook)
	case boil.AfterUpdateHook:
		chassisWeaponAfterUpdateHooks = append(chassisWeaponAfterUpdateHooks, chassisWeaponHook)
	case boil.BeforeDeleteHook:
		chassisWeaponBeforeDeleteHooks = append(chassisWeaponBeforeDeleteHooks, chassisWeaponHook)
	case boil.AfterDeleteHook:
		chassisWeaponAfterDeleteHooks = append(chassisWeaponAfterDeleteHooks, chassisWeaponHook)
	case boil.BeforeUpsertHook:
		chassisWeaponBeforeUpsertHooks = append(chassisWeaponBeforeUpsertHooks, chassisWeaponHook)
	case boil.AfterUpsertHook:
		chassisWeaponAfterUpsertHooks = append(chassisWeaponAfterUpsertHooks, chassisWeaponHook)
	}
}

// One returns a single chassisWeapon record from the query.
func (q chassisWeaponQuery) One(exec boil.Executor) (*ChassisWeapon, error) {
	o := &ChassisWeapon{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for chassis_weapons")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all ChassisWeapon records from the query.
func (q chassisWeaponQuery) All(exec boil.Executor) (ChassisWeaponSlice, error) {
	var o []*ChassisWeapon

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to ChassisWeapon slice")
	}

	if len(chassisWeaponAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all ChassisWeapon records in the query.
func (q chassisWeaponQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count chassis_weapons rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q chassisWeaponQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if chassis_weapons exists")
	}

	return count > 0, nil
}

// Chassis pointed to by the foreign key.
func (o *ChassisWeapon) Chassis(mods ...qm.QueryMod) chassisQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.ChassisID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Chasses(queryMods...)
	queries.SetFrom(query.Query, "\"chassis\"")

	return query
}

// Weapon pointed to by the foreign key.
func (o *ChassisWeapon) Weapon(mods ...qm.QueryMod) weaponQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.WeaponID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Weapons(queryMods...)
	queries.SetFrom(query.Query, "\"weapons\"")

	return query
}

// LoadChassis allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (chassisWeaponL) LoadChassis(e boil.Executor, singular bool, maybeChassisWeapon interface{}, mods queries.Applicator) error {
	var slice []*ChassisWeapon
	var object *ChassisWeapon

	if singular {
		object = maybeChassisWeapon.(*ChassisWeapon)
	} else {
		slice = *maybeChassisWeapon.(*[]*ChassisWeapon)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &chassisWeaponR{}
		}
		args = append(args, object.ChassisID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &chassisWeaponR{}
			}

			for _, a := range args {
				if a == obj.ChassisID {
					continue Outer
				}
			}

			args = append(args, obj.ChassisID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`chassis`),
		qm.WhereIn(`chassis.id in ?`, args...),
		qmhelper.WhereIsNull(`chassis.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Chassis")
	}

	var resultSlice []*Chassis
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Chassis")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for chassis")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for chassis")
	}

	if len(chassisWeaponAfterSelectHooks) != 0 {
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
		object.R.Chassis = foreign
		if foreign.R == nil {
			foreign.R = &chassisR{}
		}
		foreign.R.ChassisWeapons = append(foreign.R.ChassisWeapons, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.ChassisID == foreign.ID {
				local.R.Chassis = foreign
				if foreign.R == nil {
					foreign.R = &chassisR{}
				}
				foreign.R.ChassisWeapons = append(foreign.R.ChassisWeapons, local)
				break
			}
		}
	}

	return nil
}

// LoadWeapon allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (chassisWeaponL) LoadWeapon(e boil.Executor, singular bool, maybeChassisWeapon interface{}, mods queries.Applicator) error {
	var slice []*ChassisWeapon
	var object *ChassisWeapon

	if singular {
		object = maybeChassisWeapon.(*ChassisWeapon)
	} else {
		slice = *maybeChassisWeapon.(*[]*ChassisWeapon)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &chassisWeaponR{}
		}
		args = append(args, object.WeaponID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &chassisWeaponR{}
			}

			for _, a := range args {
				if a == obj.WeaponID {
					continue Outer
				}
			}

			args = append(args, obj.WeaponID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`weapons`),
		qm.WhereIn(`weapons.id in ?`, args...),
		qmhelper.WhereIsNull(`weapons.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Weapon")
	}

	var resultSlice []*Weapon
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Weapon")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for weapons")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for weapons")
	}

	if len(chassisWeaponAfterSelectHooks) != 0 {
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
		object.R.Weapon = foreign
		if foreign.R == nil {
			foreign.R = &weaponR{}
		}
		foreign.R.ChassisWeapon = object
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.WeaponID == foreign.ID {
				local.R.Weapon = foreign
				if foreign.R == nil {
					foreign.R = &weaponR{}
				}
				foreign.R.ChassisWeapon = local
				break
			}
		}
	}

	return nil
}

// SetChassis of the chassisWeapon to the related item.
// Sets o.R.Chassis to related.
// Adds o to related.R.ChassisWeapons.
func (o *ChassisWeapon) SetChassis(exec boil.Executor, insert bool, related *Chassis) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"chassis_weapons\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"chassis_id"}),
		strmangle.WhereClause("\"", "\"", 2, chassisWeaponPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.ChassisID = related.ID
	if o.R == nil {
		o.R = &chassisWeaponR{
			Chassis: related,
		}
	} else {
		o.R.Chassis = related
	}

	if related.R == nil {
		related.R = &chassisR{
			ChassisWeapons: ChassisWeaponSlice{o},
		}
	} else {
		related.R.ChassisWeapons = append(related.R.ChassisWeapons, o)
	}

	return nil
}

// SetWeapon of the chassisWeapon to the related item.
// Sets o.R.Weapon to related.
// Adds o to related.R.ChassisWeapon.
func (o *ChassisWeapon) SetWeapon(exec boil.Executor, insert bool, related *Weapon) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"chassis_weapons\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"weapon_id"}),
		strmangle.WhereClause("\"", "\"", 2, chassisWeaponPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.WeaponID = related.ID
	if o.R == nil {
		o.R = &chassisWeaponR{
			Weapon: related,
		}
	} else {
		o.R.Weapon = related
	}

	if related.R == nil {
		related.R = &weaponR{
			ChassisWeapon: o,
		}
	} else {
		related.R.ChassisWeapon = o
	}

	return nil
}

// ChassisWeapons retrieves all the records using an executor.
func ChassisWeapons(mods ...qm.QueryMod) chassisWeaponQuery {
	mods = append(mods, qm.From("\"chassis_weapons\""), qmhelper.WhereIsNull("\"chassis_weapons\".\"deleted_at\""))
	return chassisWeaponQuery{NewQuery(mods...)}
}

// FindChassisWeapon retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindChassisWeapon(exec boil.Executor, iD string, selectCols ...string) (*ChassisWeapon, error) {
	chassisWeaponObj := &ChassisWeapon{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"chassis_weapons\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, chassisWeaponObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from chassis_weapons")
	}

	if err = chassisWeaponObj.doAfterSelectHooks(exec); err != nil {
		return chassisWeaponObj, err
	}

	return chassisWeaponObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *ChassisWeapon) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no chassis_weapons provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(chassisWeaponColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	chassisWeaponInsertCacheMut.RLock()
	cache, cached := chassisWeaponInsertCache[key]
	chassisWeaponInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			chassisWeaponAllColumns,
			chassisWeaponColumnsWithDefault,
			chassisWeaponColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(chassisWeaponType, chassisWeaponMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(chassisWeaponType, chassisWeaponMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"chassis_weapons\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"chassis_weapons\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into chassis_weapons")
	}

	if !cached {
		chassisWeaponInsertCacheMut.Lock()
		chassisWeaponInsertCache[key] = cache
		chassisWeaponInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the ChassisWeapon.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *ChassisWeapon) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	chassisWeaponUpdateCacheMut.RLock()
	cache, cached := chassisWeaponUpdateCache[key]
	chassisWeaponUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			chassisWeaponAllColumns,
			chassisWeaponPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update chassis_weapons, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"chassis_weapons\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, chassisWeaponPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(chassisWeaponType, chassisWeaponMapping, append(wl, chassisWeaponPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update chassis_weapons row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for chassis_weapons")
	}

	if !cached {
		chassisWeaponUpdateCacheMut.Lock()
		chassisWeaponUpdateCache[key] = cache
		chassisWeaponUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q chassisWeaponQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for chassis_weapons")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for chassis_weapons")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ChassisWeaponSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), chassisWeaponPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"chassis_weapons\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, chassisWeaponPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in chassisWeapon slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all chassisWeapon")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *ChassisWeapon) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no chassis_weapons provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(chassisWeaponColumnsWithDefault, o)

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

	chassisWeaponUpsertCacheMut.RLock()
	cache, cached := chassisWeaponUpsertCache[key]
	chassisWeaponUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			chassisWeaponAllColumns,
			chassisWeaponColumnsWithDefault,
			chassisWeaponColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			chassisWeaponAllColumns,
			chassisWeaponPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert chassis_weapons, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(chassisWeaponPrimaryKeyColumns))
			copy(conflict, chassisWeaponPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"chassis_weapons\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(chassisWeaponType, chassisWeaponMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(chassisWeaponType, chassisWeaponMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert chassis_weapons")
	}

	if !cached {
		chassisWeaponUpsertCacheMut.Lock()
		chassisWeaponUpsertCache[key] = cache
		chassisWeaponUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single ChassisWeapon record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *ChassisWeapon) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no ChassisWeapon provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), chassisWeaponPrimaryKeyMapping)
		sql = "DELETE FROM \"chassis_weapons\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"chassis_weapons\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(chassisWeaponType, chassisWeaponMapping, append(wl, chassisWeaponPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from chassis_weapons")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for chassis_weapons")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q chassisWeaponQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no chassisWeaponQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from chassis_weapons")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for chassis_weapons")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ChassisWeaponSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(chassisWeaponBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), chassisWeaponPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"chassis_weapons\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, chassisWeaponPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), chassisWeaponPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"chassis_weapons\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, chassisWeaponPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from chassisWeapon slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for chassis_weapons")
	}

	if len(chassisWeaponAfterDeleteHooks) != 0 {
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
func (o *ChassisWeapon) Reload(exec boil.Executor) error {
	ret, err := FindChassisWeapon(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ChassisWeaponSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ChassisWeaponSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), chassisWeaponPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"chassis_weapons\".* FROM \"chassis_weapons\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, chassisWeaponPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in ChassisWeaponSlice")
	}

	*o = slice

	return nil
}

// ChassisWeaponExists checks if the ChassisWeapon row exists.
func ChassisWeaponExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"chassis_weapons\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if chassis_weapons exists")
	}

	return exists, nil
}