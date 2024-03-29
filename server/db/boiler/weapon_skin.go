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

// WeaponSkin is an object representing the database table.
type WeaponSkin struct {
	ID          string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	BlueprintID string      `boiler:"blueprint_id" boil:"blueprint_id" json:"blueprint_id" toml:"blueprint_id" yaml:"blueprint_id"`
	EquippedOn  null.String `boiler:"equipped_on" boil:"equipped_on" json:"equipped_on,omitempty" toml:"equipped_on" yaml:"equipped_on,omitempty"`
	CreatedAt   time.Time   `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *weaponSkinR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L weaponSkinL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var WeaponSkinColumns = struct {
	ID          string
	BlueprintID string
	EquippedOn  string
	CreatedAt   string
}{
	ID:          "id",
	BlueprintID: "blueprint_id",
	EquippedOn:  "equipped_on",
	CreatedAt:   "created_at",
}

var WeaponSkinTableColumns = struct {
	ID          string
	BlueprintID string
	EquippedOn  string
	CreatedAt   string
}{
	ID:          "weapon_skin.id",
	BlueprintID: "weapon_skin.blueprint_id",
	EquippedOn:  "weapon_skin.equipped_on",
	CreatedAt:   "weapon_skin.created_at",
}

// Generated where

var WeaponSkinWhere = struct {
	ID          whereHelperstring
	BlueprintID whereHelperstring
	EquippedOn  whereHelpernull_String
	CreatedAt   whereHelpertime_Time
}{
	ID:          whereHelperstring{field: "\"weapon_skin\".\"id\""},
	BlueprintID: whereHelperstring{field: "\"weapon_skin\".\"blueprint_id\""},
	EquippedOn:  whereHelpernull_String{field: "\"weapon_skin\".\"equipped_on\""},
	CreatedAt:   whereHelpertime_Time{field: "\"weapon_skin\".\"created_at\""},
}

// WeaponSkinRels is where relationship names are stored.
var WeaponSkinRels = struct {
	Blueprint                 string
	EquippedOnWeapon          string
	EquippedWeaponSkinWeapons string
}{
	Blueprint:                 "Blueprint",
	EquippedOnWeapon:          "EquippedOnWeapon",
	EquippedWeaponSkinWeapons: "EquippedWeaponSkinWeapons",
}

// weaponSkinR is where relationships are stored.
type weaponSkinR struct {
	Blueprint                 *BlueprintWeaponSkin `boiler:"Blueprint" boil:"Blueprint" json:"Blueprint" toml:"Blueprint" yaml:"Blueprint"`
	EquippedOnWeapon          *Weapon              `boiler:"EquippedOnWeapon" boil:"EquippedOnWeapon" json:"EquippedOnWeapon" toml:"EquippedOnWeapon" yaml:"EquippedOnWeapon"`
	EquippedWeaponSkinWeapons WeaponSlice          `boiler:"EquippedWeaponSkinWeapons" boil:"EquippedWeaponSkinWeapons" json:"EquippedWeaponSkinWeapons" toml:"EquippedWeaponSkinWeapons" yaml:"EquippedWeaponSkinWeapons"`
}

// NewStruct creates a new relationship struct
func (*weaponSkinR) NewStruct() *weaponSkinR {
	return &weaponSkinR{}
}

// weaponSkinL is where Load methods for each relationship are stored.
type weaponSkinL struct{}

var (
	weaponSkinAllColumns            = []string{"id", "blueprint_id", "equipped_on", "created_at"}
	weaponSkinColumnsWithoutDefault = []string{"blueprint_id"}
	weaponSkinColumnsWithDefault    = []string{"id", "equipped_on", "created_at"}
	weaponSkinPrimaryKeyColumns     = []string{"id"}
	weaponSkinGeneratedColumns      = []string{}
)

type (
	// WeaponSkinSlice is an alias for a slice of pointers to WeaponSkin.
	// This should almost always be used instead of []WeaponSkin.
	WeaponSkinSlice []*WeaponSkin
	// WeaponSkinHook is the signature for custom WeaponSkin hook methods
	WeaponSkinHook func(boil.Executor, *WeaponSkin) error

	weaponSkinQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	weaponSkinType                 = reflect.TypeOf(&WeaponSkin{})
	weaponSkinMapping              = queries.MakeStructMapping(weaponSkinType)
	weaponSkinPrimaryKeyMapping, _ = queries.BindMapping(weaponSkinType, weaponSkinMapping, weaponSkinPrimaryKeyColumns)
	weaponSkinInsertCacheMut       sync.RWMutex
	weaponSkinInsertCache          = make(map[string]insertCache)
	weaponSkinUpdateCacheMut       sync.RWMutex
	weaponSkinUpdateCache          = make(map[string]updateCache)
	weaponSkinUpsertCacheMut       sync.RWMutex
	weaponSkinUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var weaponSkinAfterSelectHooks []WeaponSkinHook

var weaponSkinBeforeInsertHooks []WeaponSkinHook
var weaponSkinAfterInsertHooks []WeaponSkinHook

var weaponSkinBeforeUpdateHooks []WeaponSkinHook
var weaponSkinAfterUpdateHooks []WeaponSkinHook

var weaponSkinBeforeDeleteHooks []WeaponSkinHook
var weaponSkinAfterDeleteHooks []WeaponSkinHook

var weaponSkinBeforeUpsertHooks []WeaponSkinHook
var weaponSkinAfterUpsertHooks []WeaponSkinHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *WeaponSkin) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponSkinAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *WeaponSkin) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponSkinBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *WeaponSkin) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponSkinAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *WeaponSkin) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponSkinBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *WeaponSkin) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponSkinAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *WeaponSkin) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponSkinBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *WeaponSkin) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponSkinAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *WeaponSkin) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponSkinBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *WeaponSkin) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range weaponSkinAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddWeaponSkinHook registers your hook function for all future operations.
func AddWeaponSkinHook(hookPoint boil.HookPoint, weaponSkinHook WeaponSkinHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		weaponSkinAfterSelectHooks = append(weaponSkinAfterSelectHooks, weaponSkinHook)
	case boil.BeforeInsertHook:
		weaponSkinBeforeInsertHooks = append(weaponSkinBeforeInsertHooks, weaponSkinHook)
	case boil.AfterInsertHook:
		weaponSkinAfterInsertHooks = append(weaponSkinAfterInsertHooks, weaponSkinHook)
	case boil.BeforeUpdateHook:
		weaponSkinBeforeUpdateHooks = append(weaponSkinBeforeUpdateHooks, weaponSkinHook)
	case boil.AfterUpdateHook:
		weaponSkinAfterUpdateHooks = append(weaponSkinAfterUpdateHooks, weaponSkinHook)
	case boil.BeforeDeleteHook:
		weaponSkinBeforeDeleteHooks = append(weaponSkinBeforeDeleteHooks, weaponSkinHook)
	case boil.AfterDeleteHook:
		weaponSkinAfterDeleteHooks = append(weaponSkinAfterDeleteHooks, weaponSkinHook)
	case boil.BeforeUpsertHook:
		weaponSkinBeforeUpsertHooks = append(weaponSkinBeforeUpsertHooks, weaponSkinHook)
	case boil.AfterUpsertHook:
		weaponSkinAfterUpsertHooks = append(weaponSkinAfterUpsertHooks, weaponSkinHook)
	}
}

// One returns a single weaponSkin record from the query.
func (q weaponSkinQuery) One(exec boil.Executor) (*WeaponSkin, error) {
	o := &WeaponSkin{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for weapon_skin")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all WeaponSkin records from the query.
func (q weaponSkinQuery) All(exec boil.Executor) (WeaponSkinSlice, error) {
	var o []*WeaponSkin

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to WeaponSkin slice")
	}

	if len(weaponSkinAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all WeaponSkin records in the query.
func (q weaponSkinQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count weapon_skin rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q weaponSkinQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if weapon_skin exists")
	}

	return count > 0, nil
}

// Blueprint pointed to by the foreign key.
func (o *WeaponSkin) Blueprint(mods ...qm.QueryMod) blueprintWeaponSkinQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BlueprintID),
	}

	queryMods = append(queryMods, mods...)

	query := BlueprintWeaponSkins(queryMods...)
	queries.SetFrom(query.Query, "\"blueprint_weapon_skin\"")

	return query
}

// EquippedOnWeapon pointed to by the foreign key.
func (o *WeaponSkin) EquippedOnWeapon(mods ...qm.QueryMod) weaponQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.EquippedOn),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Weapons(queryMods...)
	queries.SetFrom(query.Query, "\"weapons\"")

	return query
}

// EquippedWeaponSkinWeapons retrieves all the weapon's Weapons with an executor via equipped_weapon_skin_id column.
func (o *WeaponSkin) EquippedWeaponSkinWeapons(mods ...qm.QueryMod) weaponQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"weapons\".\"equipped_weapon_skin_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"weapons\".\"deleted_at\""),
	)

	query := Weapons(queryMods...)
	queries.SetFrom(query.Query, "\"weapons\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"weapons\".*"})
	}

	return query
}

// LoadBlueprint allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (weaponSkinL) LoadBlueprint(e boil.Executor, singular bool, maybeWeaponSkin interface{}, mods queries.Applicator) error {
	var slice []*WeaponSkin
	var object *WeaponSkin

	if singular {
		object = maybeWeaponSkin.(*WeaponSkin)
	} else {
		slice = *maybeWeaponSkin.(*[]*WeaponSkin)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &weaponSkinR{}
		}
		args = append(args, object.BlueprintID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &weaponSkinR{}
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
		qm.From(`blueprint_weapon_skin`),
		qm.WhereIn(`blueprint_weapon_skin.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load BlueprintWeaponSkin")
	}

	var resultSlice []*BlueprintWeaponSkin
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice BlueprintWeaponSkin")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for blueprint_weapon_skin")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for blueprint_weapon_skin")
	}

	if len(weaponSkinAfterSelectHooks) != 0 {
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
			foreign.R = &blueprintWeaponSkinR{}
		}
		foreign.R.BlueprintWeaponSkins = append(foreign.R.BlueprintWeaponSkins, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.BlueprintID == foreign.ID {
				local.R.Blueprint = foreign
				if foreign.R == nil {
					foreign.R = &blueprintWeaponSkinR{}
				}
				foreign.R.BlueprintWeaponSkins = append(foreign.R.BlueprintWeaponSkins, local)
				break
			}
		}
	}

	return nil
}

// LoadEquippedOnWeapon allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (weaponSkinL) LoadEquippedOnWeapon(e boil.Executor, singular bool, maybeWeaponSkin interface{}, mods queries.Applicator) error {
	var slice []*WeaponSkin
	var object *WeaponSkin

	if singular {
		object = maybeWeaponSkin.(*WeaponSkin)
	} else {
		slice = *maybeWeaponSkin.(*[]*WeaponSkin)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &weaponSkinR{}
		}
		if !queries.IsNil(object.EquippedOn) {
			args = append(args, object.EquippedOn)
		}

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &weaponSkinR{}
			}

			for _, a := range args {
				if queries.Equal(a, obj.EquippedOn) {
					continue Outer
				}
			}

			if !queries.IsNil(obj.EquippedOn) {
				args = append(args, obj.EquippedOn)
			}

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

	if len(weaponSkinAfterSelectHooks) != 0 {
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
		object.R.EquippedOnWeapon = foreign
		if foreign.R == nil {
			foreign.R = &weaponR{}
		}
		foreign.R.EquippedOnWeaponSkins = append(foreign.R.EquippedOnWeaponSkins, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if queries.Equal(local.EquippedOn, foreign.ID) {
				local.R.EquippedOnWeapon = foreign
				if foreign.R == nil {
					foreign.R = &weaponR{}
				}
				foreign.R.EquippedOnWeaponSkins = append(foreign.R.EquippedOnWeaponSkins, local)
				break
			}
		}
	}

	return nil
}

// LoadEquippedWeaponSkinWeapons allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (weaponSkinL) LoadEquippedWeaponSkinWeapons(e boil.Executor, singular bool, maybeWeaponSkin interface{}, mods queries.Applicator) error {
	var slice []*WeaponSkin
	var object *WeaponSkin

	if singular {
		object = maybeWeaponSkin.(*WeaponSkin)
	} else {
		slice = *maybeWeaponSkin.(*[]*WeaponSkin)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &weaponSkinR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &weaponSkinR{}
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
		qm.From(`weapons`),
		qm.WhereIn(`weapons.equipped_weapon_skin_id in ?`, args...),
		qmhelper.WhereIsNull(`weapons.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load weapons")
	}

	var resultSlice []*Weapon
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice weapons")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on weapons")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for weapons")
	}

	if len(weaponAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.EquippedWeaponSkinWeapons = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &weaponR{}
			}
			foreign.R.EquippedWeaponSkin = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.EquippedWeaponSkinID {
				local.R.EquippedWeaponSkinWeapons = append(local.R.EquippedWeaponSkinWeapons, foreign)
				if foreign.R == nil {
					foreign.R = &weaponR{}
				}
				foreign.R.EquippedWeaponSkin = local
				break
			}
		}
	}

	return nil
}

// SetBlueprint of the weaponSkin to the related item.
// Sets o.R.Blueprint to related.
// Adds o to related.R.BlueprintWeaponSkins.
func (o *WeaponSkin) SetBlueprint(exec boil.Executor, insert bool, related *BlueprintWeaponSkin) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"weapon_skin\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"blueprint_id"}),
		strmangle.WhereClause("\"", "\"", 2, weaponSkinPrimaryKeyColumns),
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
		o.R = &weaponSkinR{
			Blueprint: related,
		}
	} else {
		o.R.Blueprint = related
	}

	if related.R == nil {
		related.R = &blueprintWeaponSkinR{
			BlueprintWeaponSkins: WeaponSkinSlice{o},
		}
	} else {
		related.R.BlueprintWeaponSkins = append(related.R.BlueprintWeaponSkins, o)
	}

	return nil
}

// SetEquippedOnWeapon of the weaponSkin to the related item.
// Sets o.R.EquippedOnWeapon to related.
// Adds o to related.R.EquippedOnWeaponSkins.
func (o *WeaponSkin) SetEquippedOnWeapon(exec boil.Executor, insert bool, related *Weapon) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"weapon_skin\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"equipped_on"}),
		strmangle.WhereClause("\"", "\"", 2, weaponSkinPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	queries.Assign(&o.EquippedOn, related.ID)
	if o.R == nil {
		o.R = &weaponSkinR{
			EquippedOnWeapon: related,
		}
	} else {
		o.R.EquippedOnWeapon = related
	}

	if related.R == nil {
		related.R = &weaponR{
			EquippedOnWeaponSkins: WeaponSkinSlice{o},
		}
	} else {
		related.R.EquippedOnWeaponSkins = append(related.R.EquippedOnWeaponSkins, o)
	}

	return nil
}

// RemoveEquippedOnWeapon relationship.
// Sets o.R.EquippedOnWeapon to nil.
// Removes o from all passed in related items' relationships struct (Optional).
func (o *WeaponSkin) RemoveEquippedOnWeapon(exec boil.Executor, related *Weapon) error {
	var err error

	queries.SetScanner(&o.EquippedOn, nil)
	if _, err = o.Update(exec, boil.Whitelist("equipped_on")); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	if o.R != nil {
		o.R.EquippedOnWeapon = nil
	}
	if related == nil || related.R == nil {
		return nil
	}

	for i, ri := range related.R.EquippedOnWeaponSkins {
		if queries.Equal(o.EquippedOn, ri.EquippedOn) {
			continue
		}

		ln := len(related.R.EquippedOnWeaponSkins)
		if ln > 1 && i < ln-1 {
			related.R.EquippedOnWeaponSkins[i] = related.R.EquippedOnWeaponSkins[ln-1]
		}
		related.R.EquippedOnWeaponSkins = related.R.EquippedOnWeaponSkins[:ln-1]
		break
	}
	return nil
}

// AddEquippedWeaponSkinWeapons adds the given related objects to the existing relationships
// of the weapon_skin, optionally inserting them as new records.
// Appends related to o.R.EquippedWeaponSkinWeapons.
// Sets related.R.EquippedWeaponSkin appropriately.
func (o *WeaponSkin) AddEquippedWeaponSkinWeapons(exec boil.Executor, insert bool, related ...*Weapon) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.EquippedWeaponSkinID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"weapons\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"equipped_weapon_skin_id"}),
				strmangle.WhereClause("\"", "\"", 2, weaponPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.EquippedWeaponSkinID = o.ID
		}
	}

	if o.R == nil {
		o.R = &weaponSkinR{
			EquippedWeaponSkinWeapons: related,
		}
	} else {
		o.R.EquippedWeaponSkinWeapons = append(o.R.EquippedWeaponSkinWeapons, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &weaponR{
				EquippedWeaponSkin: o,
			}
		} else {
			rel.R.EquippedWeaponSkin = o
		}
	}
	return nil
}

// WeaponSkins retrieves all the records using an executor.
func WeaponSkins(mods ...qm.QueryMod) weaponSkinQuery {
	mods = append(mods, qm.From("\"weapon_skin\""))
	return weaponSkinQuery{NewQuery(mods...)}
}

// FindWeaponSkin retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindWeaponSkin(exec boil.Executor, iD string, selectCols ...string) (*WeaponSkin, error) {
	weaponSkinObj := &WeaponSkin{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"weapon_skin\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, weaponSkinObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from weapon_skin")
	}

	if err = weaponSkinObj.doAfterSelectHooks(exec); err != nil {
		return weaponSkinObj, err
	}

	return weaponSkinObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *WeaponSkin) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no weapon_skin provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(weaponSkinColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	weaponSkinInsertCacheMut.RLock()
	cache, cached := weaponSkinInsertCache[key]
	weaponSkinInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			weaponSkinAllColumns,
			weaponSkinColumnsWithDefault,
			weaponSkinColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(weaponSkinType, weaponSkinMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(weaponSkinType, weaponSkinMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"weapon_skin\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"weapon_skin\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into weapon_skin")
	}

	if !cached {
		weaponSkinInsertCacheMut.Lock()
		weaponSkinInsertCache[key] = cache
		weaponSkinInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the WeaponSkin.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *WeaponSkin) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	weaponSkinUpdateCacheMut.RLock()
	cache, cached := weaponSkinUpdateCache[key]
	weaponSkinUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			weaponSkinAllColumns,
			weaponSkinPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update weapon_skin, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"weapon_skin\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, weaponSkinPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(weaponSkinType, weaponSkinMapping, append(wl, weaponSkinPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update weapon_skin row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for weapon_skin")
	}

	if !cached {
		weaponSkinUpdateCacheMut.Lock()
		weaponSkinUpdateCache[key] = cache
		weaponSkinUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q weaponSkinQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for weapon_skin")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for weapon_skin")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o WeaponSkinSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), weaponSkinPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"weapon_skin\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, weaponSkinPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in weaponSkin slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all weaponSkin")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *WeaponSkin) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no weapon_skin provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(weaponSkinColumnsWithDefault, o)

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

	weaponSkinUpsertCacheMut.RLock()
	cache, cached := weaponSkinUpsertCache[key]
	weaponSkinUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			weaponSkinAllColumns,
			weaponSkinColumnsWithDefault,
			weaponSkinColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			weaponSkinAllColumns,
			weaponSkinPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert weapon_skin, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(weaponSkinPrimaryKeyColumns))
			copy(conflict, weaponSkinPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"weapon_skin\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(weaponSkinType, weaponSkinMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(weaponSkinType, weaponSkinMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert weapon_skin")
	}

	if !cached {
		weaponSkinUpsertCacheMut.Lock()
		weaponSkinUpsertCache[key] = cache
		weaponSkinUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single WeaponSkin record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *WeaponSkin) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no WeaponSkin provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), weaponSkinPrimaryKeyMapping)
	sql := "DELETE FROM \"weapon_skin\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from weapon_skin")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for weapon_skin")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q weaponSkinQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no weaponSkinQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from weapon_skin")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for weapon_skin")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o WeaponSkinSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(weaponSkinBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), weaponSkinPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"weapon_skin\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, weaponSkinPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from weaponSkin slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for weapon_skin")
	}

	if len(weaponSkinAfterDeleteHooks) != 0 {
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
func (o *WeaponSkin) Reload(exec boil.Executor) error {
	ret, err := FindWeaponSkin(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *WeaponSkinSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := WeaponSkinSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), weaponSkinPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"weapon_skin\".* FROM \"weapon_skin\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, weaponSkinPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in WeaponSkinSlice")
	}

	*o = slice

	return nil
}

// WeaponSkinExists checks if the WeaponSkin row exists.
func WeaponSkinExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"weapon_skin\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if weapon_skin exists")
	}

	return exists, nil
}
