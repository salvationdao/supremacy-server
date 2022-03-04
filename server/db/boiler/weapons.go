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

// Weapon is an object representing the database table.
type Weapon struct {
	ID         string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	BrandID    string    `boiler:"brand_id" boil:"brand_id" json:"brandID" toml:"brandID" yaml:"brandID"`
	Label      string    `boiler:"label" boil:"label" json:"label" toml:"label" yaml:"label"`
	Slug       string    `boiler:"slug" boil:"slug" json:"slug" toml:"slug" yaml:"slug"`
	Damage     int       `boiler:"damage" boil:"damage" json:"damage" toml:"damage" yaml:"damage"`
	WeaponType string    `boiler:"weapon_type" boil:"weapon_type" json:"weaponType" toml:"weaponType" yaml:"weaponType"`
	DeletedAt  null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deletedAt,omitempty" toml:"deletedAt" yaml:"deletedAt,omitempty"`
	UpdatedAt  time.Time `boiler:"updated_at" boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
	CreatedAt  time.Time `boiler:"created_at" boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`

	R *weaponR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L weaponL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var WeaponColumns = struct {
	ID         string
	BrandID    string
	Label      string
	Slug       string
	Damage     string
	WeaponType string
	DeletedAt  string
	UpdatedAt  string
	CreatedAt  string
}{
	ID:         "id",
	BrandID:    "brand_id",
	Label:      "label",
	Slug:       "slug",
	Damage:     "damage",
	WeaponType: "weapon_type",
	DeletedAt:  "deleted_at",
	UpdatedAt:  "updated_at",
	CreatedAt:  "created_at",
}

var WeaponTableColumns = struct {
	ID         string
	BrandID    string
	Label      string
	Slug       string
	Damage     string
	WeaponType string
	DeletedAt  string
	UpdatedAt  string
	CreatedAt  string
}{
	ID:         "weapons.id",
	BrandID:    "weapons.brand_id",
	Label:      "weapons.label",
	Slug:       "weapons.slug",
	Damage:     "weapons.damage",
	WeaponType: "weapons.weapon_type",
	DeletedAt:  "weapons.deleted_at",
	UpdatedAt:  "weapons.updated_at",
	CreatedAt:  "weapons.created_at",
}

// Generated where

var WeaponWhere = struct {
	ID         whereHelperstring
	BrandID    whereHelperstring
	Label      whereHelperstring
	Slug       whereHelperstring
	Damage     whereHelperint
	WeaponType whereHelperstring
	DeletedAt  whereHelpernull_Time
	UpdatedAt  whereHelpertime_Time
	CreatedAt  whereHelpertime_Time
}{
	ID:         whereHelperstring{field: "\"weapons\".\"id\""},
	BrandID:    whereHelperstring{field: "\"weapons\".\"brand_id\""},
	Label:      whereHelperstring{field: "\"weapons\".\"label\""},
	Slug:       whereHelperstring{field: "\"weapons\".\"slug\""},
	Damage:     whereHelperint{field: "\"weapons\".\"damage\""},
	WeaponType: whereHelperstring{field: "\"weapons\".\"weapon_type\""},
	DeletedAt:  whereHelpernull_Time{field: "\"weapons\".\"deleted_at\""},
	UpdatedAt:  whereHelpertime_Time{field: "\"weapons\".\"updated_at\""},
	CreatedAt:  whereHelpertime_Time{field: "\"weapons\".\"created_at\""},
}

// WeaponRels is where relationship names are stored.
var WeaponRels = struct {
	Brand        string
	MechsWeapons string
}{
	Brand:        "Brand",
	MechsWeapons: "MechsWeapons",
}

// weaponR is where relationships are stored.
type weaponR struct {
	Brand        *Brand           `boiler:"Brand" boil:"Brand" json:"Brand" toml:"Brand" yaml:"Brand"`
	MechsWeapons MechsWeaponSlice `boiler:"MechsWeapons" boil:"MechsWeapons" json:"MechsWeapons" toml:"MechsWeapons" yaml:"MechsWeapons"`
}

// NewStruct creates a new relationship struct
func (*weaponR) NewStruct() *weaponR {
	return &weaponR{}
}

// weaponL is where Load methods for each relationship are stored.
type weaponL struct{}

var (
	weaponAllColumns            = []string{"id", "brand_id", "label", "slug", "damage", "weapon_type", "deleted_at", "updated_at", "created_at"}
	weaponColumnsWithoutDefault = []string{"brand_id", "label", "slug", "damage", "weapon_type"}
	weaponColumnsWithDefault    = []string{"id", "deleted_at", "updated_at", "created_at"}
	weaponPrimaryKeyColumns     = []string{"id"}
	weaponGeneratedColumns      = []string{}
)

type (
	// WeaponSlice is an alias for a slice of pointers to Weapon.
	// This should almost always be used instead of []Weapon.
	WeaponSlice []*Weapon
	// WeaponHook is the signature for custom Weapon hook methods
	WeaponHook func(boil.Executor, *Weapon) error

	weaponQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	weaponType                 = reflect.TypeOf(&Weapon{})
	weaponMapping              = queries.MakeStructMapping(weaponType)
	weaponPrimaryKeyMapping, _ = queries.BindMapping(weaponType, weaponMapping, weaponPrimaryKeyColumns)
	weaponInsertCacheMut       sync.RWMutex
	weaponInsertCache          = make(map[string]insertCache)
	weaponUpdateCacheMut       sync.RWMutex
	weaponUpdateCache          = make(map[string]updateCache)
	weaponUpsertCacheMut       sync.RWMutex
	weaponUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var weaponAfterSelectHooks []WeaponHook

var weaponBeforeInsertHooks []WeaponHook
var weaponAfterInsertHooks []WeaponHook

var weaponBeforeUpdateHooks []WeaponHook
var weaponAfterUpdateHooks []WeaponHook

var weaponBeforeDeleteHooks []WeaponHook
var weaponAfterDeleteHooks []WeaponHook

var weaponBeforeUpsertHooks []WeaponHook
var weaponAfterUpsertHooks []WeaponHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Weapon) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Weapon) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Weapon) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Weapon) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Weapon) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Weapon) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Weapon) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Weapon) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Weapon) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddWeaponHook registers your hook function for all future operations.
func AddWeaponHook(hookPoint boil.HookPoint, weaponHook WeaponHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		weaponAfterSelectHooks = append(weaponAfterSelectHooks, weaponHook)
	case boil.BeforeInsertHook:
		weaponBeforeInsertHooks = append(weaponBeforeInsertHooks, weaponHook)
	case boil.AfterInsertHook:
		weaponAfterInsertHooks = append(weaponAfterInsertHooks, weaponHook)
	case boil.BeforeUpdateHook:
		weaponBeforeUpdateHooks = append(weaponBeforeUpdateHooks, weaponHook)
	case boil.AfterUpdateHook:
		weaponAfterUpdateHooks = append(weaponAfterUpdateHooks, weaponHook)
	case boil.BeforeDeleteHook:
		weaponBeforeDeleteHooks = append(weaponBeforeDeleteHooks, weaponHook)
	case boil.AfterDeleteHook:
		weaponAfterDeleteHooks = append(weaponAfterDeleteHooks, weaponHook)
	case boil.BeforeUpsertHook:
		weaponBeforeUpsertHooks = append(weaponBeforeUpsertHooks, weaponHook)
	case boil.AfterUpsertHook:
		weaponAfterUpsertHooks = append(weaponAfterUpsertHooks, weaponHook)
	}
}

// One returns a single weapon record from the query.
func (q weaponQuery) One(exec boil.Executor) (*Weapon, error) {
	o := &Weapon{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for weapons")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Weapon records from the query.
func (q weaponQuery) All(exec boil.Executor) (WeaponSlice, error) {
	var o []*Weapon

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to Weapon slice")
	}

	if len(weaponAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Weapon records in the query.
func (q weaponQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count weapons rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q weaponQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if weapons exists")
	}

	return count > 0, nil
}

// Brand pointed to by the foreign key.
func (o *Weapon) Brand(mods ...qm.QueryMod) brandQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BrandID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Brands(queryMods...)
	queries.SetFrom(query.Query, "\"brands\"")

	return query
}

// MechsWeapons retrieves all the mechs_weapon's MechsWeapons with an executor.
func (o *Weapon) MechsWeapons(mods ...qm.QueryMod) mechsWeaponQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"mechs_weapons\".\"weapon_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"mechs_weapons\".\"deleted_at\""),
	)

	query := MechsWeapons(queryMods...)
	queries.SetFrom(query.Query, "\"mechs_weapons\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"mechs_weapons\".*"})
	}

	return query
}

// LoadBrand allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (weaponL) LoadBrand(e boil.Executor, singular bool, maybeWeapon interface{}, mods queries.Applicator) error {
	var slice []*Weapon
	var object *Weapon

	if singular {
		object = maybeWeapon.(*Weapon)
	} else {
		slice = *maybeWeapon.(*[]*Weapon)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &weaponR{}
		}
		args = append(args, object.BrandID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &weaponR{}
			}

			for _, a := range args {
				if a == obj.BrandID {
					continue Outer
				}
			}

			args = append(args, obj.BrandID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`brands`),
		qm.WhereIn(`brands.id in ?`, args...),
		qmhelper.WhereIsNull(`brands.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Brand")
	}

	var resultSlice []*Brand
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Brand")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for brands")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for brands")
	}

	if len(weaponAfterSelectHooks) != 0 {
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
		object.R.Brand = foreign
		if foreign.R == nil {
			foreign.R = &brandR{}
		}
		foreign.R.Weapons = append(foreign.R.Weapons, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.BrandID == foreign.ID {
				local.R.Brand = foreign
				if foreign.R == nil {
					foreign.R = &brandR{}
				}
				foreign.R.Weapons = append(foreign.R.Weapons, local)
				break
			}
		}
	}

	return nil
}

// LoadMechsWeapons allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (weaponL) LoadMechsWeapons(e boil.Executor, singular bool, maybeWeapon interface{}, mods queries.Applicator) error {
	var slice []*Weapon
	var object *Weapon

	if singular {
		object = maybeWeapon.(*Weapon)
	} else {
		slice = *maybeWeapon.(*[]*Weapon)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &weaponR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &weaponR{}
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
		qm.From(`mechs_weapons`),
		qm.WhereIn(`mechs_weapons.weapon_id in ?`, args...),
		qmhelper.WhereIsNull(`mechs_weapons.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load mechs_weapons")
	}

	var resultSlice []*MechsWeapon
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice mechs_weapons")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on mechs_weapons")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for mechs_weapons")
	}

	if len(mechsWeaponAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.MechsWeapons = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &mechsWeaponR{}
			}
			foreign.R.Weapon = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.WeaponID {
				local.R.MechsWeapons = append(local.R.MechsWeapons, foreign)
				if foreign.R == nil {
					foreign.R = &mechsWeaponR{}
				}
				foreign.R.Weapon = local
				break
			}
		}
	}

	return nil
}

// SetBrand of the weapon to the related item.
// Sets o.R.Brand to related.
// Adds o to related.R.Weapons.
func (o *Weapon) SetBrand(exec boil.Executor, insert bool, related *Brand) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"weapons\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"brand_id"}),
		strmangle.WhereClause("\"", "\"", 2, weaponPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.BrandID = related.ID
	if o.R == nil {
		o.R = &weaponR{
			Brand: related,
		}
	} else {
		o.R.Brand = related
	}

	if related.R == nil {
		related.R = &brandR{
			Weapons: WeaponSlice{o},
		}
	} else {
		related.R.Weapons = append(related.R.Weapons, o)
	}

	return nil
}

// AddMechsWeapons adds the given related objects to the existing relationships
// of the weapon, optionally inserting them as new records.
// Appends related to o.R.MechsWeapons.
// Sets related.R.Weapon appropriately.
func (o *Weapon) AddMechsWeapons(exec boil.Executor, insert bool, related ...*MechsWeapon) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.WeaponID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"mechs_weapons\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"weapon_id"}),
				strmangle.WhereClause("\"", "\"", 2, mechsWeaponPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.WeaponID = o.ID
		}
	}

	if o.R == nil {
		o.R = &weaponR{
			MechsWeapons: related,
		}
	} else {
		o.R.MechsWeapons = append(o.R.MechsWeapons, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &mechsWeaponR{
				Weapon: o,
			}
		} else {
			rel.R.Weapon = o
		}
	}
	return nil
}

// Weapons retrieves all the records using an executor.
func Weapons(mods ...qm.QueryMod) weaponQuery {
	mods = append(mods, qm.From("\"weapons\""), qmhelper.WhereIsNull("\"weapons\".\"deleted_at\""))
	return weaponQuery{NewQuery(mods...)}
}

// FindWeapon retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindWeapon(exec boil.Executor, iD string, selectCols ...string) (*Weapon, error) {
	weaponObj := &Weapon{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"weapons\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, weaponObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from weapons")
	}

	if err = weaponObj.doAfterSelectHooks(exec); err != nil {
		return weaponObj, err
	}

	return weaponObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Weapon) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no weapons provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(weaponColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	weaponInsertCacheMut.RLock()
	cache, cached := weaponInsertCache[key]
	weaponInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			weaponAllColumns,
			weaponColumnsWithDefault,
			weaponColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(weaponType, weaponMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(weaponType, weaponMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"weapons\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"weapons\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into weapons")
	}

	if !cached {
		weaponInsertCacheMut.Lock()
		weaponInsertCache[key] = cache
		weaponInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the Weapon.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Weapon) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	weaponUpdateCacheMut.RLock()
	cache, cached := weaponUpdateCache[key]
	weaponUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			weaponAllColumns,
			weaponPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update weapons, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"weapons\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, weaponPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(weaponType, weaponMapping, append(wl, weaponPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update weapons row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for weapons")
	}

	if !cached {
		weaponUpdateCacheMut.Lock()
		weaponUpdateCache[key] = cache
		weaponUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q weaponQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for weapons")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for weapons")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o WeaponSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), weaponPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"weapons\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, weaponPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in weapon slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all weapon")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Weapon) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no weapons provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(weaponColumnsWithDefault, o)

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

	weaponUpsertCacheMut.RLock()
	cache, cached := weaponUpsertCache[key]
	weaponUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			weaponAllColumns,
			weaponColumnsWithDefault,
			weaponColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			weaponAllColumns,
			weaponPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert weapons, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(weaponPrimaryKeyColumns))
			copy(conflict, weaponPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"weapons\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(weaponType, weaponMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(weaponType, weaponMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert weapons")
	}

	if !cached {
		weaponUpsertCacheMut.Lock()
		weaponUpsertCache[key] = cache
		weaponUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single Weapon record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Weapon) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no Weapon provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), weaponPrimaryKeyMapping)
		sql = "DELETE FROM \"weapons\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"weapons\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(weaponType, weaponMapping, append(wl, weaponPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from weapons")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for weapons")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q weaponQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no weaponQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from weapons")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for weapons")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o WeaponSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(weaponBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), weaponPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"weapons\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, weaponPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), weaponPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"weapons\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, weaponPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from weapon slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for weapons")
	}

	if len(weaponAfterDeleteHooks) != 0 {
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
func (o *Weapon) Reload(exec boil.Executor) error {
	ret, err := FindWeapon(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *WeaponSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := WeaponSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), weaponPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"weapons\".* FROM \"weapons\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, weaponPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in WeaponSlice")
	}

	*o = slice

	return nil
}

// WeaponExists checks if the Weapon row exists.
func WeaponExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"weapons\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if weapons exists")
	}

	return exists, nil
}
