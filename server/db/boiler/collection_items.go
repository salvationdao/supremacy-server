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

// CollectionItem is an object representing the database table.
type CollectionItem struct {
	ID               string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	CollectionSlug   string      `boiler:"collection_slug" boil:"collection_slug" json:"collection_slug" toml:"collection_slug" yaml:"collection_slug"`
	Hash             string      `boiler:"hash" boil:"hash" json:"hash" toml:"hash" yaml:"hash"`
	TokenID          int64       `boiler:"token_id" boil:"token_id" json:"token_id" toml:"token_id" yaml:"token_id"`
	ItemType         string      `boiler:"item_type" boil:"item_type" json:"item_type" toml:"item_type" yaml:"item_type"`
	ItemID           string      `boiler:"item_id" boil:"item_id" json:"item_id" toml:"item_id" yaml:"item_id"`
	Tier             string      `boiler:"tier" boil:"tier" json:"tier" toml:"tier" yaml:"tier"`
	OwnerID          string      `boiler:"owner_id" boil:"owner_id" json:"owner_id" toml:"owner_id" yaml:"owner_id"`
	OnChainStatus    string      `boiler:"on_chain_status" boil:"on_chain_status" json:"on_chain_status" toml:"on_chain_status" yaml:"on_chain_status"`
	ImageURL         null.String `boiler:"image_url" boil:"image_url" json:"image_url,omitempty" toml:"image_url" yaml:"image_url,omitempty"`
	CardAnimationURL null.String `boiler:"card_animation_url" boil:"card_animation_url" json:"card_animation_url,omitempty" toml:"card_animation_url" yaml:"card_animation_url,omitempty"`
	AvatarURL        null.String `boiler:"avatar_url" boil:"avatar_url" json:"avatar_url,omitempty" toml:"avatar_url" yaml:"avatar_url,omitempty"`
	LargeImageURL    null.String `boiler:"large_image_url" boil:"large_image_url" json:"large_image_url,omitempty" toml:"large_image_url" yaml:"large_image_url,omitempty"`
	BackgroundColor  null.String `boiler:"background_color" boil:"background_color" json:"background_color,omitempty" toml:"background_color" yaml:"background_color,omitempty"`
	AnimationURL     null.String `boiler:"animation_url" boil:"animation_url" json:"animation_url,omitempty" toml:"animation_url" yaml:"animation_url,omitempty"`
	YoutubeURL       null.String `boiler:"youtube_url" boil:"youtube_url" json:"youtube_url,omitempty" toml:"youtube_url" yaml:"youtube_url,omitempty"`

	R *collectionItemR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L collectionItemL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var CollectionItemColumns = struct {
	ID               string
	CollectionSlug   string
	Hash             string
	TokenID          string
	ItemType         string
	ItemID           string
	Tier             string
	OwnerID          string
	OnChainStatus    string
	ImageURL         string
	CardAnimationURL string
	AvatarURL        string
	LargeImageURL    string
	BackgroundColor  string
	AnimationURL     string
	YoutubeURL       string
}{
	ID:               "id",
	CollectionSlug:   "collection_slug",
	Hash:             "hash",
	TokenID:          "token_id",
	ItemType:         "item_type",
	ItemID:           "item_id",
	Tier:             "tier",
	OwnerID:          "owner_id",
	OnChainStatus:    "on_chain_status",
	ImageURL:         "image_url",
	CardAnimationURL: "card_animation_url",
	AvatarURL:        "avatar_url",
	LargeImageURL:    "large_image_url",
	BackgroundColor:  "background_color",
	AnimationURL:     "animation_url",
	YoutubeURL:       "youtube_url",
}

var CollectionItemTableColumns = struct {
	ID               string
	CollectionSlug   string
	Hash             string
	TokenID          string
	ItemType         string
	ItemID           string
	Tier             string
	OwnerID          string
	OnChainStatus    string
	ImageURL         string
	CardAnimationURL string
	AvatarURL        string
	LargeImageURL    string
	BackgroundColor  string
	AnimationURL     string
	YoutubeURL       string
}{
	ID:               "collection_items.id",
	CollectionSlug:   "collection_items.collection_slug",
	Hash:             "collection_items.hash",
	TokenID:          "collection_items.token_id",
	ItemType:         "collection_items.item_type",
	ItemID:           "collection_items.item_id",
	Tier:             "collection_items.tier",
	OwnerID:          "collection_items.owner_id",
	OnChainStatus:    "collection_items.on_chain_status",
	ImageURL:         "collection_items.image_url",
	CardAnimationURL: "collection_items.card_animation_url",
	AvatarURL:        "collection_items.avatar_url",
	LargeImageURL:    "collection_items.large_image_url",
	BackgroundColor:  "collection_items.background_color",
	AnimationURL:     "collection_items.animation_url",
	YoutubeURL:       "collection_items.youtube_url",
}

// Generated where

var CollectionItemWhere = struct {
	ID               whereHelperstring
	CollectionSlug   whereHelperstring
	Hash             whereHelperstring
	TokenID          whereHelperint64
	ItemType         whereHelperstring
	ItemID           whereHelperstring
	Tier             whereHelperstring
	OwnerID          whereHelperstring
	OnChainStatus    whereHelperstring
	ImageURL         whereHelpernull_String
	CardAnimationURL whereHelpernull_String
	AvatarURL        whereHelpernull_String
	LargeImageURL    whereHelpernull_String
	BackgroundColor  whereHelpernull_String
	AnimationURL     whereHelpernull_String
	YoutubeURL       whereHelpernull_String
}{
	ID:               whereHelperstring{field: "\"collection_items\".\"id\""},
	CollectionSlug:   whereHelperstring{field: "\"collection_items\".\"collection_slug\""},
	Hash:             whereHelperstring{field: "\"collection_items\".\"hash\""},
	TokenID:          whereHelperint64{field: "\"collection_items\".\"token_id\""},
	ItemType:         whereHelperstring{field: "\"collection_items\".\"item_type\""},
	ItemID:           whereHelperstring{field: "\"collection_items\".\"item_id\""},
	Tier:             whereHelperstring{field: "\"collection_items\".\"tier\""},
	OwnerID:          whereHelperstring{field: "\"collection_items\".\"owner_id\""},
	OnChainStatus:    whereHelperstring{field: "\"collection_items\".\"on_chain_status\""},
	ImageURL:         whereHelpernull_String{field: "\"collection_items\".\"image_url\""},
	CardAnimationURL: whereHelpernull_String{field: "\"collection_items\".\"card_animation_url\""},
	AvatarURL:        whereHelpernull_String{field: "\"collection_items\".\"avatar_url\""},
	LargeImageURL:    whereHelpernull_String{field: "\"collection_items\".\"large_image_url\""},
	BackgroundColor:  whereHelpernull_String{field: "\"collection_items\".\"background_color\""},
	AnimationURL:     whereHelpernull_String{field: "\"collection_items\".\"animation_url\""},
	YoutubeURL:       whereHelpernull_String{field: "\"collection_items\".\"youtube_url\""},
}

// CollectionItemRels is where relationship names are stored.
var CollectionItemRels = struct {
	Owner         string
	ItemItemSales string
}{
	Owner:         "Owner",
	ItemItemSales: "ItemItemSales",
}

// collectionItemR is where relationships are stored.
type collectionItemR struct {
	Owner         *Player       `boiler:"Owner" boil:"Owner" json:"Owner" toml:"Owner" yaml:"Owner"`
	ItemItemSales ItemSaleSlice `boiler:"ItemItemSales" boil:"ItemItemSales" json:"ItemItemSales" toml:"ItemItemSales" yaml:"ItemItemSales"`
}

// NewStruct creates a new relationship struct
func (*collectionItemR) NewStruct() *collectionItemR {
	return &collectionItemR{}
}

// collectionItemL is where Load methods for each relationship are stored.
type collectionItemL struct{}

var (
	collectionItemAllColumns            = []string{"id", "collection_slug", "hash", "token_id", "item_type", "item_id", "tier", "owner_id", "on_chain_status", "image_url", "card_animation_url", "avatar_url", "large_image_url", "background_color", "animation_url", "youtube_url"}
	collectionItemColumnsWithoutDefault = []string{"token_id", "item_type", "item_id", "owner_id"}
	collectionItemColumnsWithDefault    = []string{"id", "collection_slug", "hash", "tier", "on_chain_status", "image_url", "card_animation_url", "avatar_url", "large_image_url", "background_color", "animation_url", "youtube_url"}
	collectionItemPrimaryKeyColumns     = []string{"id"}
	collectionItemGeneratedColumns      = []string{}
)

type (
	// CollectionItemSlice is an alias for a slice of pointers to CollectionItem.
	// This should almost always be used instead of []CollectionItem.
	CollectionItemSlice []*CollectionItem
	// CollectionItemHook is the signature for custom CollectionItem hook methods
	CollectionItemHook func(boil.Executor, *CollectionItem) error

	collectionItemQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	collectionItemType                 = reflect.TypeOf(&CollectionItem{})
	collectionItemMapping              = queries.MakeStructMapping(collectionItemType)
	collectionItemPrimaryKeyMapping, _ = queries.BindMapping(collectionItemType, collectionItemMapping, collectionItemPrimaryKeyColumns)
	collectionItemInsertCacheMut       sync.RWMutex
	collectionItemInsertCache          = make(map[string]insertCache)
	collectionItemUpdateCacheMut       sync.RWMutex
	collectionItemUpdateCache          = make(map[string]updateCache)
	collectionItemUpsertCacheMut       sync.RWMutex
	collectionItemUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var collectionItemAfterSelectHooks []CollectionItemHook

var collectionItemBeforeInsertHooks []CollectionItemHook
var collectionItemAfterInsertHooks []CollectionItemHook

var collectionItemBeforeUpdateHooks []CollectionItemHook
var collectionItemAfterUpdateHooks []CollectionItemHook

var collectionItemBeforeDeleteHooks []CollectionItemHook
var collectionItemAfterDeleteHooks []CollectionItemHook

var collectionItemBeforeUpsertHooks []CollectionItemHook
var collectionItemAfterUpsertHooks []CollectionItemHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *CollectionItem) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range collectionItemAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *CollectionItem) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range collectionItemBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *CollectionItem) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range collectionItemAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *CollectionItem) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range collectionItemBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *CollectionItem) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range collectionItemAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *CollectionItem) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range collectionItemBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *CollectionItem) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range collectionItemAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *CollectionItem) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range collectionItemBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *CollectionItem) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range collectionItemAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddCollectionItemHook registers your hook function for all future operations.
func AddCollectionItemHook(hookPoint boil.HookPoint, collectionItemHook CollectionItemHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		collectionItemAfterSelectHooks = append(collectionItemAfterSelectHooks, collectionItemHook)
	case boil.BeforeInsertHook:
		collectionItemBeforeInsertHooks = append(collectionItemBeforeInsertHooks, collectionItemHook)
	case boil.AfterInsertHook:
		collectionItemAfterInsertHooks = append(collectionItemAfterInsertHooks, collectionItemHook)
	case boil.BeforeUpdateHook:
		collectionItemBeforeUpdateHooks = append(collectionItemBeforeUpdateHooks, collectionItemHook)
	case boil.AfterUpdateHook:
		collectionItemAfterUpdateHooks = append(collectionItemAfterUpdateHooks, collectionItemHook)
	case boil.BeforeDeleteHook:
		collectionItemBeforeDeleteHooks = append(collectionItemBeforeDeleteHooks, collectionItemHook)
	case boil.AfterDeleteHook:
		collectionItemAfterDeleteHooks = append(collectionItemAfterDeleteHooks, collectionItemHook)
	case boil.BeforeUpsertHook:
		collectionItemBeforeUpsertHooks = append(collectionItemBeforeUpsertHooks, collectionItemHook)
	case boil.AfterUpsertHook:
		collectionItemAfterUpsertHooks = append(collectionItemAfterUpsertHooks, collectionItemHook)
	}
}

// One returns a single collectionItem record from the query.
func (q collectionItemQuery) One(exec boil.Executor) (*CollectionItem, error) {
	o := &CollectionItem{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for collection_items")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all CollectionItem records from the query.
func (q collectionItemQuery) All(exec boil.Executor) (CollectionItemSlice, error) {
	var o []*CollectionItem

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to CollectionItem slice")
	}

	if len(collectionItemAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all CollectionItem records in the query.
func (q collectionItemQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count collection_items rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q collectionItemQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if collection_items exists")
	}

	return count > 0, nil
}

// Owner pointed to by the foreign key.
func (o *CollectionItem) Owner(mods ...qm.QueryMod) playerQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.OwnerID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Players(queryMods...)
	queries.SetFrom(query.Query, "\"players\"")

	return query
}

// ItemItemSales retrieves all the item_sale's ItemSales with an executor via item_id column.
func (o *CollectionItem) ItemItemSales(mods ...qm.QueryMod) itemSaleQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("\"item_sales\".\"item_id\"=?", o.ItemID),
		qmhelper.WhereIsNull("\"item_sales\".\"deleted_at\""),
	)

	query := ItemSales(queryMods...)
	queries.SetFrom(query.Query, "\"item_sales\"")

	if len(queries.GetSelect(query.Query)) == 0 {
		queries.SetSelect(query.Query, []string{"\"item_sales\".*"})
	}

	return query
}

// LoadOwner allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (collectionItemL) LoadOwner(e boil.Executor, singular bool, maybeCollectionItem interface{}, mods queries.Applicator) error {
	var slice []*CollectionItem
	var object *CollectionItem

	if singular {
		object = maybeCollectionItem.(*CollectionItem)
	} else {
		slice = *maybeCollectionItem.(*[]*CollectionItem)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &collectionItemR{}
		}
		args = append(args, object.OwnerID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &collectionItemR{}
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

	if len(collectionItemAfterSelectHooks) != 0 {
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
		foreign.R.OwnerCollectionItems = append(foreign.R.OwnerCollectionItems, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.OwnerID == foreign.ID {
				local.R.Owner = foreign
				if foreign.R == nil {
					foreign.R = &playerR{}
				}
				foreign.R.OwnerCollectionItems = append(foreign.R.OwnerCollectionItems, local)
				break
			}
		}
	}

	return nil
}

// LoadItemItemSales allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (collectionItemL) LoadItemItemSales(e boil.Executor, singular bool, maybeCollectionItem interface{}, mods queries.Applicator) error {
	var slice []*CollectionItem
	var object *CollectionItem

	if singular {
		object = maybeCollectionItem.(*CollectionItem)
	} else {
		slice = *maybeCollectionItem.(*[]*CollectionItem)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &collectionItemR{}
		}
		args = append(args, object.ItemID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &collectionItemR{}
			}

			for _, a := range args {
				if a == obj.ItemID {
					continue Outer
				}
			}

			args = append(args, obj.ItemID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`item_sales`),
		qm.WhereIn(`item_sales.item_id in ?`, args...),
		qmhelper.WhereIsNull(`item_sales.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load item_sales")
	}

	var resultSlice []*ItemSale
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice item_sales")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on item_sales")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for item_sales")
	}

	if len(itemSaleAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.ItemItemSales = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &itemSaleR{}
			}
			foreign.R.Item = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ItemID == foreign.ItemID {
				local.R.ItemItemSales = append(local.R.ItemItemSales, foreign)
				if foreign.R == nil {
					foreign.R = &itemSaleR{}
				}
				foreign.R.Item = local
				break
			}
		}
	}

	return nil
}

// SetOwner of the collectionItem to the related item.
// Sets o.R.Owner to related.
// Adds o to related.R.OwnerCollectionItems.
func (o *CollectionItem) SetOwner(exec boil.Executor, insert bool, related *Player) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"collection_items\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"owner_id"}),
		strmangle.WhereClause("\"", "\"", 2, collectionItemPrimaryKeyColumns),
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
		o.R = &collectionItemR{
			Owner: related,
		}
	} else {
		o.R.Owner = related
	}

	if related.R == nil {
		related.R = &playerR{
			OwnerCollectionItems: CollectionItemSlice{o},
		}
	} else {
		related.R.OwnerCollectionItems = append(related.R.OwnerCollectionItems, o)
	}

	return nil
}

// AddItemItemSales adds the given related objects to the existing relationships
// of the collection_item, optionally inserting them as new records.
// Appends related to o.R.ItemItemSales.
// Sets related.R.Item appropriately.
func (o *CollectionItem) AddItemItemSales(exec boil.Executor, insert bool, related ...*ItemSale) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.ItemID = o.ItemID
			if err = rel.Insert(exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE \"item_sales\" SET %s WHERE %s",
				strmangle.SetParamNames("\"", "\"", 1, []string{"item_id"}),
				strmangle.WhereClause("\"", "\"", 2, itemSalePrimaryKeyColumns),
			)
			values := []interface{}{o.ItemID, rel.ID}

			if boil.DebugMode {
				fmt.Fprintln(boil.DebugWriter, updateQuery)
				fmt.Fprintln(boil.DebugWriter, values)
			}
			if _, err = exec.Exec(updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.ItemID = o.ItemID
		}
	}

	if o.R == nil {
		o.R = &collectionItemR{
			ItemItemSales: related,
		}
	} else {
		o.R.ItemItemSales = append(o.R.ItemItemSales, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &itemSaleR{
				Item: o,
			}
		} else {
			rel.R.Item = o
		}
	}
	return nil
}

// CollectionItems retrieves all the records using an executor.
func CollectionItems(mods ...qm.QueryMod) collectionItemQuery {
	mods = append(mods, qm.From("\"collection_items\""))
	return collectionItemQuery{NewQuery(mods...)}
}

// FindCollectionItem retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindCollectionItem(exec boil.Executor, iD string, selectCols ...string) (*CollectionItem, error) {
	collectionItemObj := &CollectionItem{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"collection_items\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, collectionItemObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from collection_items")
	}

	if err = collectionItemObj.doAfterSelectHooks(exec); err != nil {
		return collectionItemObj, err
	}

	return collectionItemObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *CollectionItem) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no collection_items provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(collectionItemColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	collectionItemInsertCacheMut.RLock()
	cache, cached := collectionItemInsertCache[key]
	collectionItemInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			collectionItemAllColumns,
			collectionItemColumnsWithDefault,
			collectionItemColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(collectionItemType, collectionItemMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(collectionItemType, collectionItemMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"collection_items\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"collection_items\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into collection_items")
	}

	if !cached {
		collectionItemInsertCacheMut.Lock()
		collectionItemInsertCache[key] = cache
		collectionItemInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the CollectionItem.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *CollectionItem) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	collectionItemUpdateCacheMut.RLock()
	cache, cached := collectionItemUpdateCache[key]
	collectionItemUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			collectionItemAllColumns,
			collectionItemPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update collection_items, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"collection_items\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, collectionItemPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(collectionItemType, collectionItemMapping, append(wl, collectionItemPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update collection_items row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for collection_items")
	}

	if !cached {
		collectionItemUpdateCacheMut.Lock()
		collectionItemUpdateCache[key] = cache
		collectionItemUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q collectionItemQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for collection_items")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for collection_items")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o CollectionItemSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), collectionItemPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"collection_items\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, collectionItemPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in collectionItem slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all collectionItem")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *CollectionItem) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no collection_items provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(collectionItemColumnsWithDefault, o)

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

	collectionItemUpsertCacheMut.RLock()
	cache, cached := collectionItemUpsertCache[key]
	collectionItemUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			collectionItemAllColumns,
			collectionItemColumnsWithDefault,
			collectionItemColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			collectionItemAllColumns,
			collectionItemPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert collection_items, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(collectionItemPrimaryKeyColumns))
			copy(conflict, collectionItemPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"collection_items\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(collectionItemType, collectionItemMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(collectionItemType, collectionItemMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert collection_items")
	}

	if !cached {
		collectionItemUpsertCacheMut.Lock()
		collectionItemUpsertCache[key] = cache
		collectionItemUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single CollectionItem record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *CollectionItem) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no CollectionItem provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), collectionItemPrimaryKeyMapping)
	sql := "DELETE FROM \"collection_items\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from collection_items")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for collection_items")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q collectionItemQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no collectionItemQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from collection_items")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for collection_items")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o CollectionItemSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(collectionItemBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), collectionItemPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"collection_items\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, collectionItemPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from collectionItem slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for collection_items")
	}

	if len(collectionItemAfterDeleteHooks) != 0 {
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
func (o *CollectionItem) Reload(exec boil.Executor) error {
	ret, err := FindCollectionItem(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *CollectionItemSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := CollectionItemSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), collectionItemPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"collection_items\".* FROM \"collection_items\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, collectionItemPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in CollectionItemSlice")
	}

	*o = slice

	return nil
}

// CollectionItemExists checks if the CollectionItem row exists.
func CollectionItemExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"collection_items\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if collection_items exists")
	}

	return exists, nil
}
