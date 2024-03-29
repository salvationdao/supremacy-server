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

// Fingerprint is an object representing the database table.
type Fingerprint struct {
	ID         string              `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	VisitorID  string              `boiler:"visitor_id" boil:"visitor_id" json:"visitor_id" toml:"visitor_id" yaml:"visitor_id"`
	OsCPU      null.String         `boiler:"os_cpu" boil:"os_cpu" json:"os_cpu,omitempty" toml:"os_cpu" yaml:"os_cpu,omitempty"`
	Platform   null.String         `boiler:"platform" boil:"platform" json:"platform,omitempty" toml:"platform" yaml:"platform,omitempty"`
	Timezone   null.String         `boiler:"timezone" boil:"timezone" json:"timezone,omitempty" toml:"timezone" yaml:"timezone,omitempty"`
	Confidence decimal.NullDecimal `boiler:"confidence" boil:"confidence" json:"confidence,omitempty" toml:"confidence" yaml:"confidence,omitempty"`
	UserAgent  null.String         `boiler:"user_agent" boil:"user_agent" json:"user_agent,omitempty" toml:"user_agent" yaml:"user_agent,omitempty"`
	DeletedAt  null.Time           `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	UpdatedAt  time.Time           `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt  time.Time           `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *fingerprintR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L fingerprintL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var FingerprintColumns = struct {
	ID         string
	VisitorID  string
	OsCPU      string
	Platform   string
	Timezone   string
	Confidence string
	UserAgent  string
	DeletedAt  string
	UpdatedAt  string
	CreatedAt  string
}{
	ID:         "id",
	VisitorID:  "visitor_id",
	OsCPU:      "os_cpu",
	Platform:   "platform",
	Timezone:   "timezone",
	Confidence: "confidence",
	UserAgent:  "user_agent",
	DeletedAt:  "deleted_at",
	UpdatedAt:  "updated_at",
	CreatedAt:  "created_at",
}

var FingerprintTableColumns = struct {
	ID         string
	VisitorID  string
	OsCPU      string
	Platform   string
	Timezone   string
	Confidence string
	UserAgent  string
	DeletedAt  string
	UpdatedAt  string
	CreatedAt  string
}{
	ID:         "fingerprints.id",
	VisitorID:  "fingerprints.visitor_id",
	OsCPU:      "fingerprints.os_cpu",
	Platform:   "fingerprints.platform",
	Timezone:   "fingerprints.timezone",
	Confidence: "fingerprints.confidence",
	UserAgent:  "fingerprints.user_agent",
	DeletedAt:  "fingerprints.deleted_at",
	UpdatedAt:  "fingerprints.updated_at",
	CreatedAt:  "fingerprints.created_at",
}

// Generated where

var FingerprintWhere = struct {
	ID         whereHelperstring
	VisitorID  whereHelperstring
	OsCPU      whereHelpernull_String
	Platform   whereHelpernull_String
	Timezone   whereHelpernull_String
	Confidence whereHelperdecimal_NullDecimal
	UserAgent  whereHelpernull_String
	DeletedAt  whereHelpernull_Time
	UpdatedAt  whereHelpertime_Time
	CreatedAt  whereHelpertime_Time
}{
	ID:         whereHelperstring{field: "\"fingerprints\".\"id\""},
	VisitorID:  whereHelperstring{field: "\"fingerprints\".\"visitor_id\""},
	OsCPU:      whereHelpernull_String{field: "\"fingerprints\".\"os_cpu\""},
	Platform:   whereHelpernull_String{field: "\"fingerprints\".\"platform\""},
	Timezone:   whereHelpernull_String{field: "\"fingerprints\".\"timezone\""},
	Confidence: whereHelperdecimal_NullDecimal{field: "\"fingerprints\".\"confidence\""},
	UserAgent:  whereHelpernull_String{field: "\"fingerprints\".\"user_agent\""},
	DeletedAt:  whereHelpernull_Time{field: "\"fingerprints\".\"deleted_at\""},
	UpdatedAt:  whereHelpertime_Time{field: "\"fingerprints\".\"updated_at\""},
	CreatedAt:  whereHelpertime_Time{field: "\"fingerprints\".\"created_at\""},
}

// FingerprintRels is where relationship names are stored.
var FingerprintRels = struct {
	FingerprintIps     string
	PlayerFingerprints string
}{
	FingerprintIps:     "FingerprintIps",
	PlayerFingerprints: "PlayerFingerprints",
}

// fingerprintR is where relationships are stored.
type fingerprintR struct {
	FingerprintIps     FingerprintIPSlice     `boiler:"FingerprintIps" boil:"FingerprintIps" json:"FingerprintIps" toml:"FingerprintIps" yaml:"FingerprintIps"`
	PlayerFingerprints PlayerFingerprintSlice `boiler:"PlayerFingerprints" boil:"PlayerFingerprints" json:"PlayerFingerprints" toml:"PlayerFingerprints" yaml:"PlayerFingerprints"`
}

// NewStruct creates a new relationship struct
func (*fingerprintR) NewStruct() *fingerprintR {
	return &fingerprintR{}
}

// fingerprintL is where Load methods for each relationship are stored.
type fingerprintL struct{}

var (
	fingerprintAllColumns            = []string{"id", "visitor_id", "os_cpu", "platform", "timezone", "confidence", "user_agent", "deleted_at", "updated_at", "created_at"}
	fingerprintColumnsWithoutDefault = []string{"visitor_id"}
	fingerprintColumnsWithDefault    = []string{"id", "os_cpu", "platform", "timezone", "confidence", "user_agent", "deleted_at", "updated_at", "created_at"}
	fingerprintPrimaryKeyColumns     = []string{"id"}
	fingerprintGeneratedColumns      = []string{}
)

type (
	// FingerprintSlice is an alias for a slice of pointers to Fingerprint.
	// This should almost always be used instead of []Fingerprint.
	FingerprintSlice []*Fingerprint
	// FingerprintHook is the signature for custom Fingerprint hook methods
	FingerprintHook func(boil.Executor, *Fingerprint) error

	fingerprintQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	fingerprintType                 = reflect.TypeOf(&Fingerprint{})
	fingerprintMapping              = queries.MakeStructMapping(fingerprintType)
	fingerprintPrimaryKeyMapping, _ = queries.BindMapping(fingerprintType, fingerprintMapping, fingerprintPrimaryKeyColumns)
	fingerprintInsertCacheMut       sync.RWMutex
	fingerprintInsertCache          = make(map[string]insertCache)
	fingerprintUpdateCacheMut       sync.RWMutex
	fingerprintUpdateCache          = make(map[string]updateCache)
	fingerprintUpsertCacheMut       sync.RWMutex
	fingerprintUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var fingerprintAfterSelectHooks []FingerprintHook

var fingerprintBeforeInsertHooks []FingerprintHook
var fingerprintAfterInsertHooks []FingerprintHook

var fingerprintBeforeUpdateHooks []FingerprintHook
var fingerprintAfterUpdateHooks []FingerprintHook

var fingerprintBeforeDeleteHooks []FingerprintHook
var fingerprintAfterDeleteHooks []FingerprintHook

var fingerprintBeforeUpsertHooks []FingerprintHook
var fingerprintAfterUpsertHooks []FingerprintHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Fingerprint) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range fingerprintAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Fingerprint) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range fingerprintBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Fingerprint) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range fingerprintAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Fingerprint) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range fingerprintBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Fingerprint) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range fingerprintAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Fingerprint) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range fingerprintBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Fingerprint) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range fingerprintAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Fingerprint) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range fingerprintBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Fingerprint) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range fingerprintAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddFingerprintHook registers your hook function for all future operations.
func AddFingerprintHook(hookPoint boil.HookPoint, fingerprintHook FingerprintHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		fingerprintAfterSelectHooks = append(fingerprintAfterSelectHooks, fingerprintHook)
	case boil.BeforeInsertHook:
		fingerprintBeforeInsertHooks = append(fingerprintBeforeInsertHooks, fingerprintHook)
	case boil.AfterInsertHook:
		fingerprintAfterInsertHooks = append(fingerprintAfterInsertHooks, fingerprintHook)
	case boil.BeforeUpdateHook:
		fingerprintBeforeUpdateHooks = append(fingerprintBeforeUpdateHooks, fingerprintHook)
	case boil.AfterUpdateHook:
		fingerprintAfterUpdateHooks = append(fingerprintAfterUpdateHooks, fingerprintHook)
	case boil.BeforeDeleteHook:
		fingerprintBeforeDeleteHooks = append(fingerprintBeforeDeleteHooks, fingerprintHook)
	case boil.AfterDeleteHook:
		fingerprintAfterDeleteHooks = append(fingerprintAfterDeleteHooks, fingerprintHook)
	case boil.BeforeUpsertHook:
		fingerprintBeforeUpsertHooks = append(fingerprintBeforeUpsertHooks, fingerprintHook)
	case boil.AfterUpsertHook:
		fingerprintAfterUpsertHooks = append(fingerprintAfterUpsertHooks, fingerprintHook)
	}
}

// One returns a single fingerprint record from the query.
func (q fingerprintQuery) One(exec boil.Executor) (*Fingerprint, error) {
	o := &Fingerprint{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for fingerprints")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Fingerprint records from the query.
func (q fingerprintQuery) All(exec boil.Executor) (FingerprintSlice, error) {
	var o []*Fingerprint

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to Fingerprint slice")
	}

	if len(fingerprintAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Fingerprint records in the query.
func (q fingerprintQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count fingerprints rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q fingerprintQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if fingerprints exists")
	}

	return count > 0, nil
}

// FingerprintIps retrieves all the fingerprint_ip's FingerprintIps with an executor.
func (o *Fingerprint) FingerprintIps(mods ...qm.QueryMod) fingerprintIPQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"fingerprint_ips\".\"fingerprint_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"fingerprint_ips\".\"deleted_at\""),
	)

	query := FingerprintIps(queryMods...)
	queries.SetFrom(query.Query, "\"fingerprint_ips\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"fingerprint_ips\".*"})
	}

	return query
}

// PlayerFingerprints retrieves all the player_fingerprint's PlayerFingerprints with an executor.
func (o *Fingerprint) PlayerFingerprints(mods ...qm.QueryMod) playerFingerprintQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"player_fingerprints\".\"fingerprint_id\"=?", o.ID),
		qmhelper.WhereIsNull("\"player_fingerprints\".\"deleted_at\""),
	)

	query := PlayerFingerprints(queryMods...)
	queries.SetFrom(query.Query, "\"player_fingerprints\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"player_fingerprints\".*"})
	}

	return query
}

// LoadFingerprintIps allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (fingerprintL) LoadFingerprintIps(e boil.Executor, singular bool, maybeFingerprint interface{}, mods queries.Applicator) error {
	var slice []*Fingerprint
	var object *Fingerprint

	if singular {
		object = maybeFingerprint.(*Fingerprint)
	} else {
		slice = *maybeFingerprint.(*[]*Fingerprint)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &fingerprintR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &fingerprintR{}
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
		qm.From(`fingerprint_ips`),
		qm.WhereIn(`fingerprint_ips.fingerprint_id in ?`, args...),
		qmhelper.WhereIsNull(`fingerprint_ips.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load fingerprint_ips")
	}

	var resultSlice []*FingerprintIP
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice fingerprint_ips")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on fingerprint_ips")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for fingerprint_ips")
	}

	if len(fingerprintIPAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.FingerprintIps = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &fingerprintIPR{}
			}
			foreign.R.Fingerprint = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.FingerprintID {
				local.R.FingerprintIps = append(local.R.FingerprintIps, foreign)
				if foreign.R == nil {
					foreign.R = &fingerprintIPR{}
				}
				foreign.R.Fingerprint = local
				break
			}
		}
	}

	return nil
}

// LoadPlayerFingerprints allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (fingerprintL) LoadPlayerFingerprints(e boil.Executor, singular bool, maybeFingerprint interface{}, mods queries.Applicator) error {
	var slice []*Fingerprint
	var object *Fingerprint

	if singular {
		object = maybeFingerprint.(*Fingerprint)
	} else {
		slice = *maybeFingerprint.(*[]*Fingerprint)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &fingerprintR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &fingerprintR{}
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
		qm.From(`player_fingerprints`),
		qm.WhereIn(`player_fingerprints.fingerprint_id in ?`, args...),
		qmhelper.WhereIsNull(`player_fingerprints.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load player_fingerprints")
	}

	var resultSlice []*PlayerFingerprint
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice player_fingerprints")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on player_fingerprints")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for player_fingerprints")
	}

	if len(playerFingerprintAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.PlayerFingerprints = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &playerFingerprintR{}
			}
			foreign.R.Fingerprint = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.FingerprintID {
				local.R.PlayerFingerprints = append(local.R.PlayerFingerprints, foreign)
				if foreign.R == nil {
					foreign.R = &playerFingerprintR{}
				}
				foreign.R.Fingerprint = local
				break
			}
		}
	}

	return nil
}

// AddFingerprintIps adds the given related objects to the existing relationships
// of the fingerprint, optionally inserting them as new records.
// Appends related to o.R.FingerprintIps.
// Sets related.R.Fingerprint appropriately.
func (o *Fingerprint) AddFingerprintIps(exec boil.Executor, insert bool, related ...*FingerprintIP) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.FingerprintID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"fingerprint_ips\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"fingerprint_id"}),
				strmangle.WhereClause("\"", "\"", 2, fingerprintIPPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.IP, rel.FingerprintID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.FingerprintID = o.ID
		}
	}

	if o.R == nil {
		o.R = &fingerprintR{
			FingerprintIps: related,
		}
	} else {
		o.R.FingerprintIps = append(o.R.FingerprintIps, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &fingerprintIPR{
				Fingerprint: o,
			}
		} else {
			rel.R.Fingerprint = o
		}
	}
	return nil
}

// AddPlayerFingerprints adds the given related objects to the existing relationships
// of the fingerprint, optionally inserting them as new records.
// Appends related to o.R.PlayerFingerprints.
// Sets related.R.Fingerprint appropriately.
func (o *Fingerprint) AddPlayerFingerprints(exec boil.Executor, insert bool, related ...*PlayerFingerprint) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.FingerprintID = o.ID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"player_fingerprints\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"fingerprint_id"}),
				strmangle.WhereClause("\"", "\"", 2, playerFingerprintPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.PlayerID, rel.FingerprintID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.FingerprintID = o.ID
		}
	}

	if o.R == nil {
		o.R = &fingerprintR{
			PlayerFingerprints: related,
		}
	} else {
		o.R.PlayerFingerprints = append(o.R.PlayerFingerprints, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &playerFingerprintR{
				Fingerprint: o,
			}
		} else {
			rel.R.Fingerprint = o
		}
	}
	return nil
}

// Fingerprints retrieves all the records using an executor.
func Fingerprints(mods ...qm.QueryMod) fingerprintQuery {
	mods = append(mods, qm.From("\"fingerprints\""), qmhelper.WhereIsNull("\"fingerprints\".\"deleted_at\""))
	return fingerprintQuery{NewQuery(mods...)}
}

// FindFingerprint retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindFingerprint(exec boil.Executor, iD string, selectCols ...string) (*Fingerprint, error) {
	fingerprintObj := &Fingerprint{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"fingerprints\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, fingerprintObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from fingerprints")
	}

	if err = fingerprintObj.doAfterSelectHooks(exec); err != nil {
		return fingerprintObj, err
	}

	return fingerprintObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Fingerprint) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no fingerprints provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(fingerprintColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	fingerprintInsertCacheMut.RLock()
	cache, cached := fingerprintInsertCache[key]
	fingerprintInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			fingerprintAllColumns,
			fingerprintColumnsWithDefault,
			fingerprintColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(fingerprintType, fingerprintMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(fingerprintType, fingerprintMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"fingerprints\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"fingerprints\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into fingerprints")
	}

	if !cached {
		fingerprintInsertCacheMut.Lock()
		fingerprintInsertCache[key] = cache
		fingerprintInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the Fingerprint.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Fingerprint) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	fingerprintUpdateCacheMut.RLock()
	cache, cached := fingerprintUpdateCache[key]
	fingerprintUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			fingerprintAllColumns,
			fingerprintPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update fingerprints, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"fingerprints\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, fingerprintPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(fingerprintType, fingerprintMapping, append(wl, fingerprintPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update fingerprints row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for fingerprints")
	}

	if !cached {
		fingerprintUpdateCacheMut.Lock()
		fingerprintUpdateCache[key] = cache
		fingerprintUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q fingerprintQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for fingerprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for fingerprints")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o FingerprintSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), fingerprintPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"fingerprints\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, fingerprintPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in fingerprint slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all fingerprint")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Fingerprint) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no fingerprints provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(fingerprintColumnsWithDefault, o)

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

	fingerprintUpsertCacheMut.RLock()
	cache, cached := fingerprintUpsertCache[key]
	fingerprintUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			fingerprintAllColumns,
			fingerprintColumnsWithDefault,
			fingerprintColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			fingerprintAllColumns,
			fingerprintPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert fingerprints, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(fingerprintPrimaryKeyColumns))
			copy(conflict, fingerprintPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"fingerprints\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(fingerprintType, fingerprintMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(fingerprintType, fingerprintMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert fingerprints")
	}

	if !cached {
		fingerprintUpsertCacheMut.Lock()
		fingerprintUpsertCache[key] = cache
		fingerprintUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single Fingerprint record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Fingerprint) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no Fingerprint provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), fingerprintPrimaryKeyMapping)
		sql = "DELETE FROM \"fingerprints\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"fingerprints\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(fingerprintType, fingerprintMapping, append(wl, fingerprintPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from fingerprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for fingerprints")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q fingerprintQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no fingerprintQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from fingerprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for fingerprints")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o FingerprintSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(fingerprintBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), fingerprintPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"fingerprints\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, fingerprintPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), fingerprintPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"fingerprints\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, fingerprintPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from fingerprint slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for fingerprints")
	}

	if len(fingerprintAfterDeleteHooks) != 0 {
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
func (o *Fingerprint) Reload(exec boil.Executor) error {
	ret, err := FindFingerprint(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *FingerprintSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := FingerprintSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), fingerprintPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"fingerprints\".* FROM \"fingerprints\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, fingerprintPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in FingerprintSlice")
	}

	*o = slice

	return nil
}

// FingerprintExists checks if the Fingerprint row exists.
func FingerprintExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"fingerprints\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if fingerprints exists")
	}

	return exists, nil
}
