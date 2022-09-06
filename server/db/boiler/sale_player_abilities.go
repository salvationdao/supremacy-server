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

// SalePlayerAbility is an object representing the database table.
type SalePlayerAbility struct {
	ID           string          `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	BlueprintID  string          `boiler:"blueprint_id" boil:"blueprint_id" json:"blueprint_id" toml:"blueprint_id" yaml:"blueprint_id"`
	AmountSold   int             `boiler:"amount_sold" boil:"amount_sold" json:"amount_sold" toml:"amount_sold" yaml:"amount_sold"`
	RarityWeight int             `boiler:"rarity_weight" boil:"rarity_weight" json:"rarity_weight" toml:"rarity_weight" yaml:"rarity_weight"`
	DeletedAt    null.Time       `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	CurrentPrice decimal.Decimal `boiler:"current_price" boil:"current_price" json:"current_price" toml:"current_price" yaml:"current_price"`

	R *salePlayerAbilityR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L salePlayerAbilityL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var SalePlayerAbilityColumns = struct {
	ID           string
	BlueprintID  string
	AmountSold   string
	RarityWeight string
	DeletedAt    string
	CurrentPrice string
}{
	ID:           "id",
	BlueprintID:  "blueprint_id",
	AmountSold:   "amount_sold",
	RarityWeight: "rarity_weight",
	DeletedAt:    "deleted_at",
	CurrentPrice: "current_price",
}

var SalePlayerAbilityTableColumns = struct {
	ID           string
	BlueprintID  string
	AmountSold   string
	RarityWeight string
	DeletedAt    string
	CurrentPrice string
}{
	ID:           "sale_player_abilities.id",
	BlueprintID:  "sale_player_abilities.blueprint_id",
	AmountSold:   "sale_player_abilities.amount_sold",
	RarityWeight: "sale_player_abilities.rarity_weight",
	DeletedAt:    "sale_player_abilities.deleted_at",
	CurrentPrice: "sale_player_abilities.current_price",
}

// Generated where

var SalePlayerAbilityWhere = struct {
	ID           whereHelperstring
	BlueprintID  whereHelperstring
	AmountSold   whereHelperint
	RarityWeight whereHelperint
	DeletedAt    whereHelpernull_Time
	CurrentPrice whereHelperdecimal_Decimal
}{
	ID:           whereHelperstring{field: "\"sale_player_abilities\".\"id\""},
	BlueprintID:  whereHelperstring{field: "\"sale_player_abilities\".\"blueprint_id\""},
	AmountSold:   whereHelperint{field: "\"sale_player_abilities\".\"amount_sold\""},
	RarityWeight: whereHelperint{field: "\"sale_player_abilities\".\"rarity_weight\""},
	DeletedAt:    whereHelpernull_Time{field: "\"sale_player_abilities\".\"deleted_at\""},
	CurrentPrice: whereHelperdecimal_Decimal{field: "\"sale_player_abilities\".\"current_price\""},
}

// SalePlayerAbilityRels is where relationship names are stored.
var SalePlayerAbilityRels = struct {
	Blueprint string
}{
	Blueprint: "Blueprint",
}

// salePlayerAbilityR is where relationships are stored.
type salePlayerAbilityR struct {
	Blueprint *BlueprintPlayerAbility `boiler:"Blueprint" boil:"Blueprint" json:"Blueprint" toml:"Blueprint" yaml:"Blueprint"`
}

// NewStruct creates a new relationship struct
func (*salePlayerAbilityR) NewStruct() *salePlayerAbilityR {
	return &salePlayerAbilityR{}
}

// salePlayerAbilityL is where Load methods for each relationship are stored.
type salePlayerAbilityL struct{}

var (
	salePlayerAbilityAllColumns            = []string{"id", "blueprint_id", "amount_sold", "rarity_weight", "deleted_at", "current_price"}
	salePlayerAbilityColumnsWithoutDefault = []string{"blueprint_id"}
	salePlayerAbilityColumnsWithDefault    = []string{"id", "amount_sold", "rarity_weight", "deleted_at", "current_price"}
	salePlayerAbilityPrimaryKeyColumns     = []string{"id"}
	salePlayerAbilityGeneratedColumns      = []string{}
)

type (
	// SalePlayerAbilitySlice is an alias for a slice of pointers to SalePlayerAbility.
	// This should almost always be used instead of []SalePlayerAbility.
	SalePlayerAbilitySlice []*SalePlayerAbility
	// SalePlayerAbilityHook is the signature for custom SalePlayerAbility hook methods
	SalePlayerAbilityHook func(boil.Executor, *SalePlayerAbility) error

	salePlayerAbilityQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	salePlayerAbilityType                 = reflect.TypeOf(&SalePlayerAbility{})
	salePlayerAbilityMapping              = queries.MakeStructMapping(salePlayerAbilityType)
	salePlayerAbilityPrimaryKeyMapping, _ = queries.BindMapping(salePlayerAbilityType, salePlayerAbilityMapping, salePlayerAbilityPrimaryKeyColumns)
	salePlayerAbilityInsertCacheMut       sync.RWMutex
	salePlayerAbilityInsertCache          = make(map[string]insertCache)
	salePlayerAbilityUpdateCacheMut       sync.RWMutex
	salePlayerAbilityUpdateCache          = make(map[string]updateCache)
	salePlayerAbilityUpsertCacheMut       sync.RWMutex
	salePlayerAbilityUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var salePlayerAbilityAfterSelectHooks []SalePlayerAbilityHook

var salePlayerAbilityBeforeInsertHooks []SalePlayerAbilityHook
var salePlayerAbilityAfterInsertHooks []SalePlayerAbilityHook

var salePlayerAbilityBeforeUpdateHooks []SalePlayerAbilityHook
var salePlayerAbilityAfterUpdateHooks []SalePlayerAbilityHook

var salePlayerAbilityBeforeDeleteHooks []SalePlayerAbilityHook
var salePlayerAbilityAfterDeleteHooks []SalePlayerAbilityHook

var salePlayerAbilityBeforeUpsertHooks []SalePlayerAbilityHook
var salePlayerAbilityAfterUpsertHooks []SalePlayerAbilityHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *SalePlayerAbility) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range salePlayerAbilityAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *SalePlayerAbility) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range salePlayerAbilityBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *SalePlayerAbility) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range salePlayerAbilityAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *SalePlayerAbility) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range salePlayerAbilityBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *SalePlayerAbility) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range salePlayerAbilityAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *SalePlayerAbility) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range salePlayerAbilityBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *SalePlayerAbility) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range salePlayerAbilityAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *SalePlayerAbility) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range salePlayerAbilityBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *SalePlayerAbility) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range salePlayerAbilityAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddSalePlayerAbilityHook registers your hook function for all future operations.
func AddSalePlayerAbilityHook(hookPoint boil.HookPoint, salePlayerAbilityHook SalePlayerAbilityHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		salePlayerAbilityAfterSelectHooks = append(salePlayerAbilityAfterSelectHooks, salePlayerAbilityHook)
	case boil.BeforeInsertHook:
		salePlayerAbilityBeforeInsertHooks = append(salePlayerAbilityBeforeInsertHooks, salePlayerAbilityHook)
	case boil.AfterInsertHook:
		salePlayerAbilityAfterInsertHooks = append(salePlayerAbilityAfterInsertHooks, salePlayerAbilityHook)
	case boil.BeforeUpdateHook:
		salePlayerAbilityBeforeUpdateHooks = append(salePlayerAbilityBeforeUpdateHooks, salePlayerAbilityHook)
	case boil.AfterUpdateHook:
		salePlayerAbilityAfterUpdateHooks = append(salePlayerAbilityAfterUpdateHooks, salePlayerAbilityHook)
	case boil.BeforeDeleteHook:
		salePlayerAbilityBeforeDeleteHooks = append(salePlayerAbilityBeforeDeleteHooks, salePlayerAbilityHook)
	case boil.AfterDeleteHook:
		salePlayerAbilityAfterDeleteHooks = append(salePlayerAbilityAfterDeleteHooks, salePlayerAbilityHook)
	case boil.BeforeUpsertHook:
		salePlayerAbilityBeforeUpsertHooks = append(salePlayerAbilityBeforeUpsertHooks, salePlayerAbilityHook)
	case boil.AfterUpsertHook:
		salePlayerAbilityAfterUpsertHooks = append(salePlayerAbilityAfterUpsertHooks, salePlayerAbilityHook)
	}
}

// One returns a single salePlayerAbility record from the query.
func (q salePlayerAbilityQuery) One(exec boil.Executor) (*SalePlayerAbility, error) {
	o := &SalePlayerAbility{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for sale_player_abilities")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all SalePlayerAbility records from the query.
func (q salePlayerAbilityQuery) All(exec boil.Executor) (SalePlayerAbilitySlice, error) {
	var o []*SalePlayerAbility

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to SalePlayerAbility slice")
	}

	if len(salePlayerAbilityAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all SalePlayerAbility records in the query.
func (q salePlayerAbilityQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count sale_player_abilities rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q salePlayerAbilityQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if sale_player_abilities exists")
	}

	return count > 0, nil
}

// Blueprint pointed to by the foreign key.
func (o *SalePlayerAbility) Blueprint(mods ...qm.QueryMod) blueprintPlayerAbilityQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.BlueprintID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := BlueprintPlayerAbilities(queryMods...)
	queries.SetFrom(query.Query, "\"blueprint_player_abilities\"")

	return query
}

// LoadBlueprint allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (salePlayerAbilityL) LoadBlueprint(e boil.Executor, singular bool, maybeSalePlayerAbility interface{}, mods queries.Applicator) error {
	var slice []*SalePlayerAbility
	var object *SalePlayerAbility

	if singular {
		object = maybeSalePlayerAbility.(*SalePlayerAbility)
	} else {
		slice = *maybeSalePlayerAbility.(*[]*SalePlayerAbility)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &salePlayerAbilityR{}
		}
		args = append(args, object.BlueprintID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &salePlayerAbilityR{}
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
		qmhelper.WhereIsNull(`blueprint_player_abilities.deleted_at`),
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

	if len(salePlayerAbilityAfterSelectHooks) != 0 {
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
		foreign.R.BlueprintSalePlayerAbilities = append(foreign.R.BlueprintSalePlayerAbilities, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.BlueprintID == foreign.ID {
				local.R.Blueprint = foreign
				if foreign.R == nil {
					foreign.R = &blueprintPlayerAbilityR{}
				}
				foreign.R.BlueprintSalePlayerAbilities = append(foreign.R.BlueprintSalePlayerAbilities, local)
				break
			}
		}
	}

	return nil
}

// SetBlueprint of the salePlayerAbility to the related item.
// Sets o.R.Blueprint to related.
// Adds o to related.R.BlueprintSalePlayerAbilities.
func (o *SalePlayerAbility) SetBlueprint(exec boil.Executor, insert bool, related *BlueprintPlayerAbility) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"sale_player_abilities\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"blueprint_id"}),
		strmangle.WhereClause("\"", "\"", 2, salePlayerAbilityPrimaryKeyColumns),
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
		o.R = &salePlayerAbilityR{
			Blueprint: related,
		}
	} else {
		o.R.Blueprint = related
	}

	if related.R == nil {
		related.R = &blueprintPlayerAbilityR{
			BlueprintSalePlayerAbilities: SalePlayerAbilitySlice{o},
		}
	} else {
		related.R.BlueprintSalePlayerAbilities = append(related.R.BlueprintSalePlayerAbilities, o)
	}

	return nil
}

// SalePlayerAbilities retrieves all the records using an executor.
func SalePlayerAbilities(mods ...qm.QueryMod) salePlayerAbilityQuery {
	mods = append(mods, qm.From("\"sale_player_abilities\""), qmhelper.WhereIsNull("\"sale_player_abilities\".\"deleted_at\""))
	return salePlayerAbilityQuery{NewQuery(mods...)}
}

// FindSalePlayerAbility retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindSalePlayerAbility(exec boil.Executor, iD string, selectCols ...string) (*SalePlayerAbility, error) {
	salePlayerAbilityObj := &SalePlayerAbility{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"sale_player_abilities\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, salePlayerAbilityObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from sale_player_abilities")
	}

	if err = salePlayerAbilityObj.doAfterSelectHooks(exec); err != nil {
		return salePlayerAbilityObj, err
	}

	return salePlayerAbilityObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *SalePlayerAbility) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no sale_player_abilities provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(salePlayerAbilityColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	salePlayerAbilityInsertCacheMut.RLock()
	cache, cached := salePlayerAbilityInsertCache[key]
	salePlayerAbilityInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			salePlayerAbilityAllColumns,
			salePlayerAbilityColumnsWithDefault,
			salePlayerAbilityColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(salePlayerAbilityType, salePlayerAbilityMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(salePlayerAbilityType, salePlayerAbilityMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"sale_player_abilities\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"sale_player_abilities\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into sale_player_abilities")
	}

	if !cached {
		salePlayerAbilityInsertCacheMut.Lock()
		salePlayerAbilityInsertCache[key] = cache
		salePlayerAbilityInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the SalePlayerAbility.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *SalePlayerAbility) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	salePlayerAbilityUpdateCacheMut.RLock()
	cache, cached := salePlayerAbilityUpdateCache[key]
	salePlayerAbilityUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			salePlayerAbilityAllColumns,
			salePlayerAbilityPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update sale_player_abilities, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"sale_player_abilities\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, salePlayerAbilityPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(salePlayerAbilityType, salePlayerAbilityMapping, append(wl, salePlayerAbilityPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update sale_player_abilities row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for sale_player_abilities")
	}

	if !cached {
		salePlayerAbilityUpdateCacheMut.Lock()
		salePlayerAbilityUpdateCache[key] = cache
		salePlayerAbilityUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q salePlayerAbilityQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for sale_player_abilities")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for sale_player_abilities")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o SalePlayerAbilitySlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), salePlayerAbilityPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"sale_player_abilities\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, salePlayerAbilityPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in salePlayerAbility slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all salePlayerAbility")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *SalePlayerAbility) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no sale_player_abilities provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(salePlayerAbilityColumnsWithDefault, o)

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

	salePlayerAbilityUpsertCacheMut.RLock()
	cache, cached := salePlayerAbilityUpsertCache[key]
	salePlayerAbilityUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			salePlayerAbilityAllColumns,
			salePlayerAbilityColumnsWithDefault,
			salePlayerAbilityColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			salePlayerAbilityAllColumns,
			salePlayerAbilityPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert sale_player_abilities, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(salePlayerAbilityPrimaryKeyColumns))
			copy(conflict, salePlayerAbilityPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"sale_player_abilities\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(salePlayerAbilityType, salePlayerAbilityMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(salePlayerAbilityType, salePlayerAbilityMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert sale_player_abilities")
	}

	if !cached {
		salePlayerAbilityUpsertCacheMut.Lock()
		salePlayerAbilityUpsertCache[key] = cache
		salePlayerAbilityUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single SalePlayerAbility record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *SalePlayerAbility) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no SalePlayerAbility provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), salePlayerAbilityPrimaryKeyMapping)
		sql = "DELETE FROM \"sale_player_abilities\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"sale_player_abilities\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(salePlayerAbilityType, salePlayerAbilityMapping, append(wl, salePlayerAbilityPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from sale_player_abilities")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for sale_player_abilities")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q salePlayerAbilityQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no salePlayerAbilityQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from sale_player_abilities")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for sale_player_abilities")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o SalePlayerAbilitySlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(salePlayerAbilityBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), salePlayerAbilityPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"sale_player_abilities\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, salePlayerAbilityPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), salePlayerAbilityPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"sale_player_abilities\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, salePlayerAbilityPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from salePlayerAbility slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for sale_player_abilities")
	}

	if len(salePlayerAbilityAfterDeleteHooks) != 0 {
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
func (o *SalePlayerAbility) Reload(exec boil.Executor) error {
	ret, err := FindSalePlayerAbility(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *SalePlayerAbilitySlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := SalePlayerAbilitySlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), salePlayerAbilityPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"sale_player_abilities\".* FROM \"sale_player_abilities\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, salePlayerAbilityPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in SalePlayerAbilitySlice")
	}

	*o = slice

	return nil
}

// SalePlayerAbilityExists checks if the SalePlayerAbility row exists.
func SalePlayerAbilityExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"sale_player_abilities\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if sale_player_abilities exists")
	}

	return exists, nil
}
