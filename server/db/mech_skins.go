package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// InsertNewMechSkin if modelID is nil it will return images of a random mech in this skin
func InsertNewMechSkin(tx boil.Executor, ownerID uuid.UUID, skin *server.BlueprintMechSkin, modelID *string) (*server.MechSkin, error) {
	// first insert the skin
	newSkin := boiler.MechSkin{
		BlueprintID:           skin.ID,
		GenesisTokenID:        skin.GenesisTokenID,
		LimitedReleaseTokenID: skin.LimitedReleaseTokenID,
		Level:                 skin.DefaultLevel,
	}

	err := newSkin.Insert(tx, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Interface("newSkin", newSkin).Msg("failed to insert")
		return nil, terror.Error(err)
	}

	_, err = InsertNewCollectionItem(tx,
		skin.Collection,
		boiler.ItemTypeMechSkin,
		newSkin.ID,
		skin.Tier,
		ownerID.String(),
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return MechSkin(tx, newSkin.ID, modelID)
}

// MechSkin if modelID is nil it will return images of a random mech in this skin
func MechSkin(trx boil.Executor, id string, modelID *string) (*server.MechSkin, error) {
	boilerMech, err := boiler.MechSkins(
		boiler.MechSkinWhere.ID.EQ(id),
		qm.Load(boiler.MechSkinRels.Blueprint),
	).One(trx)
	if err != nil {
		return nil, err
	}

	boilerMechCollectionDetails, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(id),
	).One(trx)
	if err != nil {
		return nil, err
	}

	queryMods := []qm.QueryMod{
		boiler.MechModelSkinCompatibilityWhere.BlueprintMechSkinID.EQ(boilerMech.BlueprintID),
	}

	// if nil was passed in, we get a random one
	if modelID != nil && *modelID != "" {
		queryMods = append(queryMods, boiler.MechModelSkinCompatibilityWhere.MechModelID.EQ(*modelID))
	}

	mechSkinCompatabilityMatrix, err := boiler.MechModelSkinCompatibilities(
		queryMods...,
	).One(trx)
	if err != nil {
		return nil, err
	}

	return server.MechSkinFromBoiler(boilerMech, boilerMechCollectionDetails, mechSkinCompatabilityMatrix, boilerMech.R.Blueprint), nil
}

func MechSkins(id ...string) ([]*server.MechSkin, error) {
	var skins []*server.MechSkin
	boilerMechSkins, err := boiler.MechSkins(
		boiler.MechSkinWhere.ID.IN(id),
		qm.Load(boiler.MechSkinRels.Blueprint),
	).All(gamedb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, ms := range boilerMechSkins {
		boilerMechCollectionDetails, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(ms.ID)).One(gamedb.StdConn)
		if err != nil {
			return nil, err
		}

		skins = append(skins, server.MechSkinFromBoiler(ms, boilerMechCollectionDetails, nil, ms.R.Blueprint))
	}
	return skins, nil
}

func IsMechSkinColumn(col string) bool {
	switch col {
	case boiler.MechSkinColumns.ID,
		boiler.MechSkinColumns.CreatedAt,
		boiler.MechSkinColumns.BlueprintID,
		boiler.MechSkinColumns.EquippedOn,
		boiler.MechSkinColumns.LockedToMech,
		boiler.MechSkinColumns.GenesisTokenID,
		boiler.MechSkinColumns.LimitedReleaseTokenID:
		return true
	default:
		return false
	}
}

type MechSkinListOpts struct {
	Search                   string
	Filter                   *ListFilterRequest
	SortBy                   SortBy
	SortDir                  SortByDir
	PageSize                 int
	Page                     int
	OwnerID                  string
	ModelID                  string
	DisplayXsyn              bool
	ExcludeMarketLocked      bool
	IncludeMarketListed      bool
	DisplayGenesisAndLimited bool
	ExcludeIDs               []string `json:"exclude_ids"`
	IncludeIDs               []string `json:"include_ids"`
	FilterRarities           []string `json:"rarities"`
	FilterSkinCompatibility  []string `json:"skin_compatibility"`
	FilterEquippedStatuses   []string `json:"equipped_statuses"`
}

func MechSkinListDetailed(opts *MechSkinListOpts) (int64, []*server.MechSkin, error) {
	var queryMods []qm.QueryMod

	queryMods = append(queryMods,
		// hide hidden assets
		boiler.CollectionItemWhere.AssetHidden.IsNull(),
		// where owner id = ?
		GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.OwnerID,
			Operator: OperatorValueTypeEquals,
			Value:    opts.OwnerID,
		}, 0, ""),
		// and item type = mech SkinID
		GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.CollectionItems,
			Column:   boiler.CollectionItemColumns.ItemType,
			Operator: OperatorValueTypeEquals,
			Value:    boiler.ItemTypeMechSkin,
		}, 0, "and"),
		// inner join mechs skin
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.MechSkin,
			qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
		)),
		// inner join mechs skin blueprint
		qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
			boiler.TableNames.BlueprintMechSkin,
			qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID),
			qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.BlueprintID),
		)),
	)
	if opts.ModelID != "" {
		queryMods = append(queryMods, qm.InnerJoin(fmt.Sprintf("LATERAL (SELECT * FROM %s _wmsc WHERE _wmsc.%s = %s AND _wmsc.%s = ? LIMIT 1) %s ON %s = %s",
			boiler.TableNames.MechModelSkinCompatibilities,
			boiler.MechModelSkinCompatibilityColumns.BlueprintMechSkinID,
			qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID),
			boiler.MechModelSkinCompatibilityColumns.MechModelID,
			boiler.TableNames.MechModelSkinCompatibilities,
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.BlueprintMechSkinID),
			qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID),
		), opts.ModelID))
	} else {
		queryMods = append(queryMods, qm.InnerJoin(fmt.Sprintf("LATERAL (SELECT * FROM %s _wmsc WHERE _wmsc.%s = %s LIMIT 1) %s ON %s = %s",
			boiler.TableNames.MechModelSkinCompatibilities,
			boiler.MechModelSkinCompatibilityColumns.BlueprintMechSkinID,
			qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID),
			boiler.TableNames.MechModelSkinCompatibilities,
			qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.BlueprintMechSkinID),
			qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID),
		)))
	}

	if len(opts.FilterSkinCompatibility) > 0 {
		var args []interface{}
		whereClause := fmt.Sprintf("WHERE %s IN (", qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.MechModelID))
		//// inner join mech model
		for i, r := range opts.FilterSkinCompatibility {
			args = append(args, r)
			if i+1 == len(opts.FilterSkinCompatibility) {
				whereClause = whereClause + "?)"
				continue
			}
			whereClause = whereClause + "?,"
		}

		queryMods = append(queryMods,
			qm.InnerJoin(fmt.Sprintf("(SELECT %s, JSONB_AGG(%s) as models FROM %s %s GROUP BY %s) sq on sq.%s = %s",
				boiler.MechModelSkinCompatibilityColumns.BlueprintMechSkinID,
				boiler.MechModelSkinCompatibilityColumns.MechModelID,
				boiler.TableNames.MechModelSkinCompatibilities,
				whereClause,
				boiler.MechModelSkinCompatibilityColumns.BlueprintMechSkinID,
				boiler.MechModelSkinCompatibilityColumns.BlueprintMechSkinID,
				qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID),
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
			if f.Table == boiler.TableNames.MechSkin && IsMechSkinColumn(f.Column) {
				queryMods = append(queryMods, GenerateListFilterQueryMod(*f, i+1, opts.Filter.LinkOperator))
			}

		}
	}

	if len(opts.ExcludeIDs) > 0 {
		queryMods = append(queryMods, boiler.MechSkinWhere.ID.NIN(opts.ExcludeIDs))
	}
	if len(opts.IncludeIDs) > 0 {
		queryMods = append(queryMods, boiler.MechSkinWhere.BlueprintID.IN(opts.IncludeIDs))
	}
	if len(opts.FilterRarities) > 0 {
		vals := []interface{}{}
		for _, r := range opts.FilterRarities {
			vals = append(vals, r)
		}
		queryMods = append(queryMods, qm.AndIn(fmt.Sprintf("%s IN ?", qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.Tier)), vals...))
	}
	if len(opts.FilterEquippedStatuses) == 1 {
		if opts.FilterEquippedStatuses[0] == "unequipped" {
			queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
				Table:    boiler.TableNames.MechSkin,
				Column:   boiler.MechSkinColumns.EquippedOn,
				Operator: OperatorValueTypeIsNull,
			}, 0, ""))
		} else {
			queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
				Table:    boiler.TableNames.MechSkin,
				Column:   boiler.MechSkinColumns.EquippedOn,
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
					qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechColumns.Label),
				),
					xSearch,
				))
		}
	}

	countQueryMods := []qm.QueryMod{}
	countQueryMods = append(countQueryMods, queryMods...)

	total, err := boiler.CollectionItems(
		countQueryMods...,
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
	appendSelectCols := []string{
		boiler.BlueprintMechSkinTableColumns.ID,
		boiler.CollectionItemTableColumns.CollectionSlug,
		boiler.CollectionItemTableColumns.Hash,
		boiler.CollectionItemTableColumns.TokenID,
		boiler.CollectionItemTableColumns.OwnerID,
		boiler.CollectionItemTableColumns.ItemType,
		boiler.CollectionItemTableColumns.MarketLocked,
		boiler.CollectionItemTableColumns.XsynLocked,
		boiler.CollectionItemTableColumns.LockedToMarketplace,
		boiler.CollectionItemTableColumns.AssetHidden,
		boiler.MechSkinTableColumns.ID,
		boiler.MechSkinTableColumns.EquippedOn,
		boiler.MechSkinTableColumns.LockedToMech,
		boiler.MechSkinTableColumns.GenesisTokenID,
		boiler.MechSkinTableColumns.LimitedReleaseTokenID,
		boiler.MechSkinTableColumns.Level,
		boiler.BlueprintMechSkinTableColumns.Label,
		boiler.BlueprintMechSkinTableColumns.DefaultLevel,
		boiler.BlueprintMechSkinTableColumns.Tier,
		boiler.BlueprintMechSkinTableColumns.BlueprintWeaponSkinID,
		boiler.BlueprintMechSkinTableColumns.ImageURL,
		boiler.BlueprintMechSkinTableColumns.CardAnimationURL,
		boiler.BlueprintMechSkinTableColumns.AvatarURL,
		boiler.BlueprintMechSkinTableColumns.LargeImageURL,
		boiler.BlueprintMechSkinTableColumns.AnimationURL,
		boiler.BlueprintMechSkinTableColumns.YoutubeURL,
		boiler.BlueprintMechSkinTableColumns.BackgroundColor,
		boiler.MechModelSkinCompatibilityTableColumns.ImageURL,
		boiler.MechModelSkinCompatibilityTableColumns.CardAnimationURL,
		boiler.MechModelSkinCompatibilityTableColumns.AvatarURL,
		boiler.MechModelSkinCompatibilityTableColumns.LargeImageURL,
		boiler.MechModelSkinCompatibilityTableColumns.AnimationURL,
		boiler.MechModelSkinCompatibilityTableColumns.YoutubeURL,
		boiler.MechModelSkinCompatibilityTableColumns.BackgroundColor,
	}

	queryMods = append(queryMods,
		qm.Select(appendSelectCols...),
		qm.From(boiler.TableNames.CollectionItems),
	)

	// Sort
	if opts.SortBy != "" && opts.SortDir.IsValid() {
		switch opts.SortBy {
		case SortByAlphabetical:
			orderBy := fmt.Sprintf("(%s) %s",
				boiler.BlueprintMechSkinTableColumns.Label,
				opts.SortDir,
			)
			queryMods = append(queryMods, qm.OrderBy(orderBy))
		case SortByRarity:
			queryMods = append(queryMods, qm.OrderBy(GenerateRawTierSort(qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.Tier), opts.SortDir)))
		case SortByDate:
			orderBy := fmt.Sprintf("(%s) %s",
				boiler.MechSkinTableColumns.CreatedAt,
				opts.SortDir,
			)
			queryMods = append(queryMods, qm.OrderBy(orderBy))
		}
	} else {
		orderBy := fmt.Sprintf("%s ASC",
			qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.Label),
		)
		queryMods = append(queryMods, qm.OrderBy(orderBy))
	}

	rows, err := boiler.NewQuery(
		queryMods...,
	).Query(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to run dynamic mech skin query")
		return 0, nil, err
	}
	defer rows.Close()

	mechSkins := make([]*server.MechSkin, 0)
	for rows.Next() {
		mc := &server.MechSkin{
			CollectionItem: &server.CollectionItem{},
			Images:         &server.Images{},
			SkinSwatch:     &server.Images{},
		}

		scanArgs := []interface{}{
			&mc.BlueprintID,
			&mc.CollectionItem.CollectionSlug,
			&mc.CollectionItem.Hash,
			&mc.CollectionItem.TokenID,
			&mc.CollectionItem.OwnerID,
			&mc.CollectionItem.ItemType,
			&mc.CollectionItem.MarketLocked,
			&mc.CollectionItem.XsynLocked,
			&mc.CollectionItem.LockedToMarketplace,
			&mc.CollectionItem.AssetHidden,
			&mc.ID,
			&mc.EquippedOn,
			&mc.LockedToMech,
			&mc.GenesisTokenID,
			&mc.LimitedReleaseTokenID,
			&mc.Level,
			&mc.Label,
			&mc.DefaultLevel,
			&mc.Tier,
			&mc.BlueprintWeaponSkinID,
			&mc.SkinSwatch.ImageURL,
			&mc.SkinSwatch.CardAnimationURL,
			&mc.SkinSwatch.AvatarURL,
			&mc.SkinSwatch.LargeImageURL,
			&mc.SkinSwatch.AnimationURL,
			&mc.SkinSwatch.YoutubeURL,
			&mc.SkinSwatch.BackgroundColor,
			&mc.Images.ImageURL,
			&mc.Images.CardAnimationURL,
			&mc.Images.AvatarURL,
			&mc.Images.LargeImageURL,
			&mc.Images.AnimationURL,
			&mc.Images.YoutubeURL,
			&mc.Images.BackgroundColor,
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("failed to scan mech skins")
			return total, mechSkins, err
		}
		if !mc.SkinSwatch.ImageURL.Valid &&
			!mc.SkinSwatch.CardAnimationURL.Valid &&
			!mc.SkinSwatch.AvatarURL.Valid &&
			!mc.SkinSwatch.LargeImageURL.Valid &&
			!mc.SkinSwatch.AnimationURL.Valid &&
			!mc.SkinSwatch.YoutubeURL.Valid &&
			!mc.SkinSwatch.BackgroundColor.Valid {
			mc.SkinSwatch = nil
		}
		mechSkins = append(mechSkins, mc)
	}

	return total, mechSkins, nil
}
