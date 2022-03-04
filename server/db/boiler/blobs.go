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

// Blob is an object representing the database table.
type Blob struct {
	ID            string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	FileName      string      `boiler:"file_name" boil:"file_name" json:"fileName" toml:"fileName" yaml:"fileName"`
	MimeType      string      `boiler:"mime_type" boil:"mime_type" json:"mimeType" toml:"mimeType" yaml:"mimeType"`
	FileSizeBytes int64       `boiler:"file_size_bytes" boil:"file_size_bytes" json:"fileSizeBytes" toml:"fileSizeBytes" yaml:"fileSizeBytes"`
	Extension     string      `boiler:"extension" boil:"extension" json:"extension" toml:"extension" yaml:"extension"`
	File          []byte      `boiler:"file" boil:"file" json:"file" toml:"file" yaml:"file"`
	Views         int         `boiler:"views" boil:"views" json:"views" toml:"views" yaml:"views"`
	Hash          null.String `boiler:"hash" boil:"hash" json:"hash,omitempty" toml:"hash" yaml:"hash,omitempty"`
	DeletedAt     null.Time   `boiler:"deleted_at" boil:"deleted_at" json:"deletedAt,omitempty" toml:"deletedAt" yaml:"deletedAt,omitempty"`
	UpdatedAt     time.Time   `boiler:"updated_at" boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
	CreatedAt     time.Time   `boiler:"created_at" boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`

	R *blobR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L blobL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var BlobColumns = struct {
	ID            string
	FileName      string
	MimeType      string
	FileSizeBytes string
	Extension     string
	File          string
	Views         string
	Hash          string
	DeletedAt     string
	UpdatedAt     string
	CreatedAt     string
}{
	ID:            "id",
	FileName:      "file_name",
	MimeType:      "mime_type",
	FileSizeBytes: "file_size_bytes",
	Extension:     "extension",
	File:          "file",
	Views:         "views",
	Hash:          "hash",
	DeletedAt:     "deleted_at",
	UpdatedAt:     "updated_at",
	CreatedAt:     "created_at",
}

var BlobTableColumns = struct {
	ID            string
	FileName      string
	MimeType      string
	FileSizeBytes string
	Extension     string
	File          string
	Views         string
	Hash          string
	DeletedAt     string
	UpdatedAt     string
	CreatedAt     string
}{
	ID:            "blobs.id",
	FileName:      "blobs.file_name",
	MimeType:      "blobs.mime_type",
	FileSizeBytes: "blobs.file_size_bytes",
	Extension:     "blobs.extension",
	File:          "blobs.file",
	Views:         "blobs.views",
	Hash:          "blobs.hash",
	DeletedAt:     "blobs.deleted_at",
	UpdatedAt:     "blobs.updated_at",
	CreatedAt:     "blobs.created_at",
}

// Generated where

type whereHelperint64 struct{ field string }

func (w whereHelperint64) EQ(x int64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint64) NEQ(x int64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint64) LT(x int64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint64) LTE(x int64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint64) GT(x int64) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint64) GTE(x int64) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint64) IN(slice []int64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint64) NIN(slice []int64) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelper__byte struct{ field string }

func (w whereHelper__byte) EQ(x []byte) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelper__byte) NEQ(x []byte) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelper__byte) LT(x []byte) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelper__byte) LTE(x []byte) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelper__byte) GT(x []byte) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelper__byte) GTE(x []byte) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }

var BlobWhere = struct {
	ID            whereHelperstring
	FileName      whereHelperstring
	MimeType      whereHelperstring
	FileSizeBytes whereHelperint64
	Extension     whereHelperstring
	File          whereHelper__byte
	Views         whereHelperint
	Hash          whereHelpernull_String
	DeletedAt     whereHelpernull_Time
	UpdatedAt     whereHelpertime_Time
	CreatedAt     whereHelpertime_Time
}{
	ID:            whereHelperstring{field: "\"blobs\".\"id\""},
	FileName:      whereHelperstring{field: "\"blobs\".\"file_name\""},
	MimeType:      whereHelperstring{field: "\"blobs\".\"mime_type\""},
	FileSizeBytes: whereHelperint64{field: "\"blobs\".\"file_size_bytes\""},
	Extension:     whereHelperstring{field: "\"blobs\".\"extension\""},
	File:          whereHelper__byte{field: "\"blobs\".\"file\""},
	Views:         whereHelperint{field: "\"blobs\".\"views\""},
	Hash:          whereHelpernull_String{field: "\"blobs\".\"hash\""},
	DeletedAt:     whereHelpernull_Time{field: "\"blobs\".\"deleted_at\""},
	UpdatedAt:     whereHelpertime_Time{field: "\"blobs\".\"updated_at\""},
	CreatedAt:     whereHelpertime_Time{field: "\"blobs\".\"created_at\""},
}

// BlobRels is where relationship names are stored.
var BlobRels = struct {
}{}

// blobR is where relationships are stored.
type blobR struct {
}

// NewStruct creates a new relationship struct
func (*blobR) NewStruct() *blobR {
	return &blobR{}
}

// blobL is where Load methods for each relationship are stored.
type blobL struct{}

var (
	blobAllColumns            = []string{"id", "file_name", "mime_type", "file_size_bytes", "extension", "file", "views", "hash", "deleted_at", "updated_at", "created_at"}
	blobColumnsWithoutDefault = []string{"file_name", "mime_type", "file_size_bytes", "extension", "file"}
	blobColumnsWithDefault    = []string{"id", "views", "hash", "deleted_at", "updated_at", "created_at"}
	blobPrimaryKeyColumns     = []string{"id"}
	blobGeneratedColumns      = []string{}
)

type (
	// BlobSlice is an alias for a slice of pointers to Blob.
	// This should almost always be used instead of []Blob.
	BlobSlice []*Blob
	// BlobHook is the signature for custom Blob hook methods
	BlobHook func(boil.Executor, *Blob) error

	blobQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	blobType                 = reflect.TypeOf(&Blob{})
	blobMapping              = queries.MakeStructMapping(blobType)
	blobPrimaryKeyMapping, _ = queries.BindMapping(blobType, blobMapping, blobPrimaryKeyColumns)
	blobInsertCacheMut       sync.RWMutex
	blobInsertCache          = make(map[string]insertCache)
	blobUpdateCacheMut       sync.RWMutex
	blobUpdateCache          = make(map[string]updateCache)
	blobUpsertCacheMut       sync.RWMutex
	blobUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var blobAfterSelectHooks []BlobHook

var blobBeforeInsertHooks []BlobHook
var blobAfterInsertHooks []BlobHook

var blobBeforeUpdateHooks []BlobHook
var blobAfterUpdateHooks []BlobHook

var blobBeforeDeleteHooks []BlobHook
var blobAfterDeleteHooks []BlobHook

var blobBeforeUpsertHooks []BlobHook
var blobAfterUpsertHooks []BlobHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Blob) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range blobAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Blob) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range blobBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Blob) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range blobAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Blob) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range blobBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Blob) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range blobAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Blob) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range blobBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Blob) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range blobAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Blob) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range blobBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Blob) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range blobAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddBlobHook registers your hook function for all future operations.
func AddBlobHook(hookPoint boil.HookPoint, blobHook BlobHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		blobAfterSelectHooks = append(blobAfterSelectHooks, blobHook)
	case boil.BeforeInsertHook:
		blobBeforeInsertHooks = append(blobBeforeInsertHooks, blobHook)
	case boil.AfterInsertHook:
		blobAfterInsertHooks = append(blobAfterInsertHooks, blobHook)
	case boil.BeforeUpdateHook:
		blobBeforeUpdateHooks = append(blobBeforeUpdateHooks, blobHook)
	case boil.AfterUpdateHook:
		blobAfterUpdateHooks = append(blobAfterUpdateHooks, blobHook)
	case boil.BeforeDeleteHook:
		blobBeforeDeleteHooks = append(blobBeforeDeleteHooks, blobHook)
	case boil.AfterDeleteHook:
		blobAfterDeleteHooks = append(blobAfterDeleteHooks, blobHook)
	case boil.BeforeUpsertHook:
		blobBeforeUpsertHooks = append(blobBeforeUpsertHooks, blobHook)
	case boil.AfterUpsertHook:
		blobAfterUpsertHooks = append(blobAfterUpsertHooks, blobHook)
	}
}

// One returns a single blob record from the query.
func (q blobQuery) One(exec boil.Executor) (*Blob, error) {
	o := &Blob{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for blobs")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Blob records from the query.
func (q blobQuery) All(exec boil.Executor) (BlobSlice, error) {
	var o []*Blob

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to Blob slice")
	}

	if len(blobAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Blob records in the query.
func (q blobQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count blobs rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q blobQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if blobs exists")
	}

	return count > 0, nil
}

// Blobs retrieves all the records using an executor.
func Blobs(mods ...qm.QueryMod) blobQuery {
	mods = append(mods, qm.From("\"blobs\""), qmhelper.WhereIsNull("\"blobs\".\"deleted_at\""))
	return blobQuery{NewQuery(mods...)}
}

// FindBlob retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindBlob(exec boil.Executor, iD string, selectCols ...string) (*Blob, error) {
	blobObj := &Blob{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"blobs\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, blobObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from blobs")
	}

	if err = blobObj.doAfterSelectHooks(exec); err != nil {
		return blobObj, err
	}

	return blobObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Blob) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no blobs provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(blobColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	blobInsertCacheMut.RLock()
	cache, cached := blobInsertCache[key]
	blobInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			blobAllColumns,
			blobColumnsWithDefault,
			blobColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(blobType, blobMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(blobType, blobMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"blobs\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"blobs\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into blobs")
	}

	if !cached {
		blobInsertCacheMut.Lock()
		blobInsertCache[key] = cache
		blobInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the Blob.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Blob) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	blobUpdateCacheMut.RLock()
	cache, cached := blobUpdateCache[key]
	blobUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			blobAllColumns,
			blobPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update blobs, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"blobs\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, blobPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(blobType, blobMapping, append(wl, blobPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update blobs row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for blobs")
	}

	if !cached {
		blobUpdateCacheMut.Lock()
		blobUpdateCache[key] = cache
		blobUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q blobQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for blobs")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for blobs")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o BlobSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), blobPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"blobs\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, blobPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in blob slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all blob")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Blob) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no blobs provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(blobColumnsWithDefault, o)

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

	blobUpsertCacheMut.RLock()
	cache, cached := blobUpsertCache[key]
	blobUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			blobAllColumns,
			blobColumnsWithDefault,
			blobColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			blobAllColumns,
			blobPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert blobs, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(blobPrimaryKeyColumns))
			copy(conflict, blobPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"blobs\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(blobType, blobMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(blobType, blobMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert blobs")
	}

	if !cached {
		blobUpsertCacheMut.Lock()
		blobUpsertCache[key] = cache
		blobUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single Blob record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Blob) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no Blob provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), blobPrimaryKeyMapping)
		sql = "DELETE FROM \"blobs\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"blobs\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(blobType, blobMapping, append(wl, blobPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from blobs")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for blobs")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q blobQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no blobQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from blobs")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for blobs")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o BlobSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(blobBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), blobPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"blobs\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, blobPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), blobPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"blobs\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, blobPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from blob slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for blobs")
	}

	if len(blobAfterDeleteHooks) != 0 {
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
func (o *Blob) Reload(exec boil.Executor) error {
	ret, err := FindBlob(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *BlobSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := BlobSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), blobPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"blobs\".* FROM \"blobs\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, blobPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in BlobSlice")
	}

	*o = slice

	return nil
}

// BlobExists checks if the Blob row exists.
func BlobExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"blobs\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if blobs exists")
	}

	return exists, nil
}
