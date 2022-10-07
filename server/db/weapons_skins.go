package db

import (
	"database/sql"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func InsertNewWeaponSkin(tx *sql.Tx, ownerID uuid.UUID, blueprintWeaponSkin *server.BlueprintWeaponSkin, modelID *string) (*server.WeaponSkin, error) {
	newWeaponSkin := boiler.WeaponSkin{
		BlueprintID: blueprintWeaponSkin.ID,
		EquippedOn:  null.String{},
		CreatedAt:   blueprintWeaponSkin.CreatedAt,
	}

	err := newWeaponSkin.Insert(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		blueprintWeaponSkin.Collection,
		boiler.ItemTypeWeaponSkin,
		newWeaponSkin.ID,
		blueprintWeaponSkin.Tier,
		ownerID.String(),
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return WeaponSkin(tx, newWeaponSkin.ID, modelID)
}

func WeaponSkin(tx boil.Executor, id string, blueprintID *string) (*server.WeaponSkin, error) {
	boilerWeaponSkin, err := boiler.WeaponSkins(
		boiler.WeaponSkinWhere.ID.EQ(id),
		qm.Load(boiler.WeaponSkinRels.Blueprint),
	).One(tx)
	if err != nil {
		return nil, err
	}
	boilerWeaponCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		boiler.WeaponModelSkinCompatibilityWhere.BlueprintWeaponSkinID.EQ(boilerWeaponSkin.BlueprintID),
	}

	if blueprintID != nil && *blueprintID != "" {
		queryMods = append(queryMods, boiler.WeaponModelSkinCompatibilityWhere.WeaponModelID.EQ(*blueprintID))
	}

	weaponSkinCompatMatrix, err := boiler.WeaponModelSkinCompatibilities(
		queryMods...,
	).One(tx)
	if err != nil {
		return nil, err
	}
	return server.WeaponSkinFromBoiler(boilerWeaponSkin, boilerWeaponCollectionDetails, weaponSkinCompatMatrix, boilerWeaponSkin.R.Blueprint), nil
}

func IsWeaponSkinColumn(col string) bool {
	switch col {
	case boiler.WeaponSkinColumns.ID,
		boiler.WeaponSkinColumns.CreatedAt,
		boiler.WeaponSkinColumns.BlueprintID,
		boiler.WeaponSkinColumns.EquippedOn:
		return true
	default:
		return false
	}
}

type WeaponSkinListOpts struct {
	Search                   string
	Filter                   *ListFilterRequest
	Sort                     *ListSortRequest
	SortBy                   string
	SortDir                  SortByDir
	PageSize                 int
	Page                     int
	OwnerID                  string
	DisplayXsyn              bool
	ExcludeMarketLocked      bool
	IncludeMarketListed      bool
	DisplayGenesisAndLimited bool
	FilterRarities           []string `json:"rarities"`
	FilterSkinCompatibility  []string `json:"skin_compatibility"`
	FilterEquippedStatuses   []string `json:"equipped_statuses"`
}

func WeaponSkinListDetailed(opts *WeaponSkinListOpts) (int64, []*server.WeaponSkin, error) {

	var weaponSkins []*server.WeaponSkin

	var queryMods []qm.QueryMod

	queryMods = append(queryMods,
		// where owner id = ?
		GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.OwnerID,
			Operator: OperatorValueTypeEquals,
			Value:    opts.OwnerID,
		}, 0, ""),
		// and item type = weapon Skin
		GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.ItemType,
			Operator: OperatorValueTypeEquals,
			Value:    boiler.ItemTypeWeaponSkin,
		}, 0, "and"),
		// inner join weapon skin
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.WeaponSkin,
			qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		// inner join weapon skin blueprint
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintWeaponSkin,
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
			qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.BlueprintID),
		)),
		qm.InnerJoin(fmt.Sprintf("LATERAL (SELECT * FROM %s _wmsc WHERE _wmsc.%s = %s LIMIT 1) %s ON %s = %s",
			boiler.TableNames.WeaponModelSkinCompatibilities,
			boiler.WeaponModelSkinCompatibilityColumns.BlueprintWeaponSkinID,
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
			boiler.TableNames.WeaponModelSkinCompatibilities,
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.BlueprintWeaponSkinID),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
		)),
	)

	if len(opts.FilterSkinCompatibility) > 0 {
		var args []interface{}
		whereClause := fmt.Sprintf("WHERE %s IN (", qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.WeaponModelID))
		//// inner join weapon model
		for i, r := range opts.FilterSkinCompatibility {
			args = append(args, r)
			if i+1 == len(opts.FilterSkinCompatibility) {
				whereClause = whereClause + "?)"
				continue
			}
			whereClause = whereClause + fmt.Sprintf("?,")
		}

		queryMods = append(queryMods,
			qm.InnerJoin(fmt.Sprintf("(SELECT %s, JSONB_AGG(%s) as models FROM %s %s GROUP BY %s) sq on sq.%s = %s",
				boiler.WeaponModelSkinCompatibilityColumns.BlueprintWeaponSkinID,
				boiler.WeaponModelSkinCompatibilityColumns.WeaponModelID,
				boiler.TableNames.WeaponModelSkinCompatibilities,
				whereClause,
				boiler.WeaponModelSkinCompatibilityColumns.BlueprintWeaponSkinID,
				boiler.WeaponModelSkinCompatibilityColumns.BlueprintWeaponSkinID,
				qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
			),
				args...,
			),
		)
	}

	if !opts.DisplayXsyn || !opts.IncludeMarketListed {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.XsynLocked,
			Operator: OperatorValueTypeIsFalse,
		}, 0, ""))
	}
	if opts.ExcludeMarketLocked {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.MarketLocked,
			Operator: OperatorValueTypeIsFalse,
		}, 0, ""))
	}
	if !opts.IncludeMarketListed {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.LockedToMarketplace,
			Operator: OperatorValueTypeIsFalse,
		}, 0, ""))
	}

	// Filters
	if opts.Filter != nil {
		// if we have filter
		for i, f := range opts.Filter.Items {
			// validate it is the right table and valid column
			if f.Table == boiler.TableNames.WeaponSkin && IsWeaponSkinColumn(f.Column) {
				queryMods = append(queryMods, GenerateListFilterQueryMod(*f, i+1, opts.Filter.LinkOperator))
			}

		}
	}
	if len(opts.FilterRarities) > 0 {
		vals := []interface{}{}
		for _, r := range opts.FilterRarities {
			vals = append(vals, r)
		}
		queryMods = append(queryMods, qm.AndIn(fmt.Sprintf("%s IN ?", qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.Tier)), vals...))
	}
	if len(opts.FilterEquippedStatuses) == 1 {
		if opts.FilterEquippedStatuses[0] == "UNEQUIPPED" {
			queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
				Table:    boiler.TableNames.WeaponSkin,
				Column:   boiler.WeaponSkinColumns.EquippedOn,
				Operator: OperatorValueTypeIsNull,
			}, 0, ""))
		} else {
			queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
				Table:    boiler.TableNames.WeaponSkin,
				Column:   boiler.WeaponSkinColumns.EquippedOn,
				Operator: OperatorValueTypeIsNotNull,
			}, 0, ""))
		}
	}

	//Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"(to_tsvector('english', %s) @@ to_tsquery(?))",
					qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponColumns.Label),
				),
					xSearch,
				))
		}
	}

	total, err := boiler.CollectionItems(
		queryMods...,
	).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, err
	}

	// Limit/Offset
	if opts.PageSize > 0 {
		queryMods = append(queryMods, qm.Limit(opts.PageSize))
	}
	if opts.Page > 0 {
		queryMods = append(queryMods, qm.Offset(opts.PageSize*(opts.Page-1)))
	}

	// Build query
	queryMods = append(queryMods,
		qm.Select(
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.CollectionSlug),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Hash),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.TokenID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.OwnerID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.MarketLocked),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.XsynLocked),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.LockedToMarketplace),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.AssetHidden),
			qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.ID),
			qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.EquippedOn),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.StatModifier),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.Label),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.Tier),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ImageURL),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.CardAnimationURL),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.AvatarURL),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.LargeImageURL),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.AnimationURL),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.YoutubeURL),
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.BackgroundColor),
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.ImageURL),
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.CardAnimationURL),
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.AvatarURL),
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.LargeImageURL),
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.AnimationURL),
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.YoutubeURL),
			qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.BackgroundColor),
		),
		qm.From(boiler.TableNames.CollectionItems),
	)

	// Sort
	if opts.Sort != nil && opts.Sort.Table == boiler.TableNames.WeaponSkin && IsWeaponSkinColumn(opts.Sort.Column) && opts.Sort.Direction.IsValid() {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.WeaponSkin, opts.Sort.Column, opts.Sort.Direction)))
	} else if opts.SortBy != "" && opts.SortDir.IsValid() {
		if opts.SortBy == "alphabetical" {
			queryMods = append(queryMods,
				qm.OrderBy(
					fmt.Sprintf("(%[1]s) %[2]s",
						qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.Label),
						opts.SortDir,
					)))
		} else if opts.SortBy == "rarity" {
			queryMods = append(queryMods, GenerateTierSort(qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.Tier), opts.SortDir))
		}
	} else {
		queryMods = append(queryMods,
			qm.OrderBy(
				fmt.Sprintf("%[1]s ASC",
					qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.Label),
				)))
	}
	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to run dynamic weapon skin query")
		return 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		ws := &server.WeaponSkin{
			CollectionItem: &server.CollectionItem{},
			SkinSwatch:     &server.Images{},
			Images:         &server.Images{},
		}

		scanArgs := []interface{}{
			&ws.CollectionItem.CollectionSlug,
			&ws.CollectionItem.Hash,
			&ws.CollectionItem.TokenID,
			&ws.CollectionItem.OwnerID,
			&ws.CollectionItem.ItemType,
			&ws.CollectionItem.MarketLocked,
			&ws.CollectionItem.XsynLocked,
			&ws.CollectionItem.LockedToMarketplace,
			&ws.CollectionItem.AssetHidden,
			&ws.ID,
			&ws.EquippedOn,
			&ws.BlueprintID,
			&ws.StatModifier,
			&ws.Label,
			&ws.Tier,
			&ws.SkinSwatch.ImageURL,
			&ws.SkinSwatch.CardAnimationURL,
			&ws.SkinSwatch.AvatarURL,
			&ws.SkinSwatch.LargeImageURL,
			&ws.SkinSwatch.AnimationURL,
			&ws.SkinSwatch.YoutubeURL,
			&ws.SkinSwatch.BackgroundColor,
			&ws.Images.ImageURL,
			&ws.Images.CardAnimationURL,
			&ws.Images.AvatarURL,
			&ws.Images.LargeImageURL,
			&ws.Images.AnimationURL,
			&ws.Images.YoutubeURL,
			&ws.Images.BackgroundColor,
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			return total, weaponSkins, err
		}
		weaponSkins = append(weaponSkins, ws)
	}

	return total, weaponSkins, nil
}
