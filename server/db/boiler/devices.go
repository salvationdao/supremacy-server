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

// Device is an object representing the database table.
type Device struct {
	ID        string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	PlayerID  string    `boiler:"player_id" boil:"player_id" json:"player_id" toml:"player_id" yaml:"player_id"`
	Token     string    `boiler:"token" boil:"token" json:"token" toml:"token" yaml:"token"`
	CreatedAt time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	DeletedAt null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`

	R *deviceR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L deviceL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var DeviceColumns = struct {
	ID        string
	PlayerID  string
	Token     string
	CreatedAt string
	UpdatedAt string
	DeletedAt string
}{
	ID:        "id",
	PlayerID:  "player_id",
	Token:     "token",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

var DeviceTableColumns = struct {
	ID        string
	PlayerID  string
	Token     string
	CreatedAt string
	UpdatedAt string
	DeletedAt string
}{
	ID:        "devices.id",
	PlayerID:  "devices.player_id",
	Token:     "devices.token",
	CreatedAt: "devices.created_at",
	UpdatedAt: "devices.updated_at",
	DeletedAt: "devices.deleted_at",
}

// Generated where

var DeviceWhere = struct {
	ID        whereHelperstring
	PlayerID  whereHelperstring
	Token     whereHelperstring
	CreatedAt whereHelpertime_Time
	UpdatedAt whereHelpertime_Time
	DeletedAt whereHelpernull_Time
}{
	ID:        whereHelperstring{field: "\"devices\".\"id\""},
	PlayerID:  whereHelperstring{field: "\"devices\".\"player_id\""},
	Token:     whereHelperstring{field: "\"devices\".\"token\""},
	CreatedAt: whereHelpertime_Time{field: "\"devices\".\"created_at\""},
	UpdatedAt: whereHelpertime_Time{field: "\"devices\".\"updated_at\""},
	DeletedAt: whereHelpernull_Time{field: "\"devices\".\"deleted_at\""},
}

// DeviceRels is where relationship names are stored.
var DeviceRels = struct {
	Player string
}{
	Player: "Player",
}

// deviceR is where relationships are stored.
type deviceR struct {
	Player *Player `boiler:"Player" boil:"Player" json:"Player" toml:"Player" yaml:"Player"`
}

// NewStruct creates a new relationship struct
func (*deviceR) NewStruct() *deviceR {
	return &deviceR{}
}

// deviceL is where Load methods for each relationship are stored.
type deviceL struct{}

var (
	deviceAllColumns            = []string{"id", "player_id", "token", "created_at", "updated_at", "deleted_at"}
	deviceColumnsWithoutDefault = []string{"player_id", "token"}
	deviceColumnsWithDefault    = []string{"id", "created_at", "updated_at", "deleted_at"}
	devicePrimaryKeyColumns     = []string{"id"}
	deviceGeneratedColumns      = []string{}
)

type (
	// DeviceSlice is an alias for a slice of pointers to Device.
	// This should almost always be used instead of []Device.
	DeviceSlice []*Device
	// DeviceHook is the signature for custom Device hook methods
	DeviceHook func(boil.Executor, *Device) error

	deviceQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	deviceType                 = reflect.TypeOf(&Device{})
	deviceMapping              = queries.MakeStructMapping(deviceType)
	devicePrimaryKeyMapping, _ = queries.BindMapping(deviceType, deviceMapping, devicePrimaryKeyColumns)
	deviceInsertCacheMut       sync.RWMutex
	deviceInsertCache          = make(map[string]insertCache)
	deviceUpdateCacheMut       sync.RWMutex
	deviceUpdateCache          = make(map[string]updateCache)
	deviceUpsertCacheMut       sync.RWMutex
	deviceUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var deviceAfterSelectHooks []DeviceHook

var deviceBeforeInsertHooks []DeviceHook
var deviceAfterInsertHooks []DeviceHook

var deviceBeforeUpdateHooks []DeviceHook
var deviceAfterUpdateHooks []DeviceHook

var deviceBeforeDeleteHooks []DeviceHook
var deviceAfterDeleteHooks []DeviceHook

var deviceBeforeUpsertHooks []DeviceHook
var deviceAfterUpsertHooks []DeviceHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Device) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range deviceAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Device) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range deviceBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Device) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range deviceAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Device) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range deviceBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Device) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range deviceAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Device) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range deviceBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Device) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range deviceAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Device) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range deviceBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Device) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range deviceAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddDeviceHook registers your hook function for all future operations.
func AddDeviceHook(hookPoint boil.HookPoint, deviceHook DeviceHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		deviceAfterSelectHooks = append(deviceAfterSelectHooks, deviceHook)
	case boil.BeforeInsertHook:
		deviceBeforeInsertHooks = append(deviceBeforeInsertHooks, deviceHook)
	case boil.AfterInsertHook:
		deviceAfterInsertHooks = append(deviceAfterInsertHooks, deviceHook)
	case boil.BeforeUpdateHook:
		deviceBeforeUpdateHooks = append(deviceBeforeUpdateHooks, deviceHook)
	case boil.AfterUpdateHook:
		deviceAfterUpdateHooks = append(deviceAfterUpdateHooks, deviceHook)
	case boil.BeforeDeleteHook:
		deviceBeforeDeleteHooks = append(deviceBeforeDeleteHooks, deviceHook)
	case boil.AfterDeleteHook:
		deviceAfterDeleteHooks = append(deviceAfterDeleteHooks, deviceHook)
	case boil.BeforeUpsertHook:
		deviceBeforeUpsertHooks = append(deviceBeforeUpsertHooks, deviceHook)
	case boil.AfterUpsertHook:
		deviceAfterUpsertHooks = append(deviceAfterUpsertHooks, deviceHook)
	}
}

// One returns a single device record from the query.
func (q deviceQuery) One(exec boil.Executor) (*Device, error) {
	o := &Device{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for devices")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Device records from the query.
func (q deviceQuery) All(exec boil.Executor) (DeviceSlice, error) {
	var o []*Device

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to Device slice")
	}

	if len(deviceAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Device records in the query.
func (q deviceQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count devices rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q deviceQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if devices exists")
	}

	return count > 0, nil
}

// Player pointed to by the foreign key.
func (o *Device) Player(mods ...qm.QueryMod) playerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.PlayerID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Players(queryMods...)
	queries.SetFrom(query.Query, "\"players\"")

	return query
}

// LoadPlayer allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (deviceL) LoadPlayer(e boil.Executor, singular bool, maybeDevice interface{}, mods queries.Applicator) error {
	var slice []*Device
	var object *Device

	if singular {
		object = maybeDevice.(*Device)
	} else {
		slice = *maybeDevice.(*[]*Device)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &deviceR{}
		}
		args = append(args, object.PlayerID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &deviceR{}
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

	if len(deviceAfterSelectHooks) != 0 {
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
		foreign.R.Devices = append(foreign.R.Devices, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.PlayerID == foreign.ID {
				local.R.Player = foreign
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.Devices = append(foreign.R.Devices, local)
				break
			}
		}
	}

	return nil
}

// SetPlayer of the device to the related item.
// Sets o.R.Player to related.
// Adds o to related.R.Devices.
func (o *Device) SetPlayer(exec boil.Executor, insert bool, related *Player) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"devices\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"player_id"}),
		strmangle.WhereClause("\"", "\"", 2, devicePrimaryKeyColumns),
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
		o.R = &deviceR{
			Player: related,
		}
	} else {
		o.R.Player = related
	}

	if related.R == nil {
		related.R = &playerR{
			Devices: DeviceSlice{o},
		}
	} else {
		related.R.Devices = append(related.R.Devices, o)
	}

	return nil
}

// Devices retrieves all the records using an executor.
func Devices(mods ...qm.QueryMod) deviceQuery {
	mods = append(mods, qm.From("\"devices\""), qmhelper.WhereIsNull("\"devices\".\"deleted_at\""))
	return deviceQuery{NewQuery(mods...)}
}

// FindDevice retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindDevice(exec boil.Executor, iD string, selectCols ...string) (*Device, error) {
	deviceObj := &Device{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"devices\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, deviceObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from devices")
	}

	if err = deviceObj.doAfterSelectHooks(exec); err != nil {
		return deviceObj, err
	}

	return deviceObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Device) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no devices provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(deviceColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	deviceInsertCacheMut.RLock()
	cache, cached := deviceInsertCache[key]
	deviceInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			deviceAllColumns,
			deviceColumnsWithDefault,
			deviceColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(deviceType, deviceMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(deviceType, deviceMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"devices\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"devices\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into devices")
	}

	if !cached {
		deviceInsertCacheMut.Lock()
		deviceInsertCache[key] = cache
		deviceInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the Device.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Device) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	deviceUpdateCacheMut.RLock()
	cache, cached := deviceUpdateCache[key]
	deviceUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			deviceAllColumns,
			devicePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update devices, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"devices\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, devicePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(deviceType, deviceMapping, append(wl, devicePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update devices row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for devices")
	}

	if !cached {
		deviceUpdateCacheMut.Lock()
		deviceUpdateCache[key] = cache
		deviceUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q deviceQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for devices")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for devices")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o DeviceSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), devicePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"devices\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, devicePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in device slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all device")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Device) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no devices provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}
	o.UpdatedAt = currTime

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(deviceColumnsWithDefault, o)

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

	deviceUpsertCacheMut.RLock()
	cache, cached := deviceUpsertCache[key]
	deviceUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			deviceAllColumns,
			deviceColumnsWithDefault,
			deviceColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			deviceAllColumns,
			devicePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert devices, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(devicePrimaryKeyColumns))
			copy(conflict, devicePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"devices\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(deviceType, deviceMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(deviceType, deviceMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert devices")
	}

	if !cached {
		deviceUpsertCacheMut.Lock()
		deviceUpsertCache[key] = cache
		deviceUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single Device record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Device) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no Device provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), devicePrimaryKeyMapping)
		sql = "DELETE FROM \"devices\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"devices\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(deviceType, deviceMapping, append(wl, devicePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from devices")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for devices")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q deviceQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no deviceQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from devices")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for devices")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o DeviceSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(deviceBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), devicePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"devices\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, devicePrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), devicePrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"devices\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, devicePrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from device slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for devices")
	}

	if len(deviceAfterDeleteHooks) != 0 {
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
func (o *Device) Reload(exec boil.Executor) error {
	ret, err := FindDevice(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *DeviceSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := DeviceSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), devicePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"devices\".* FROM \"devices\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, devicePrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in DeviceSlice")
	}

	*o = slice

	return nil
}

// DeviceExists checks if the Device row exists.
func DeviceExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"devices\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if devices exists")
	}

	return exists, nil
}
