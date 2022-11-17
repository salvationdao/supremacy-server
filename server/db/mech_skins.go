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
	DisplayUnique            bool     `json:"display_unique"`
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
		if opts.DisplayUnique {
			queryMods = append(queryMods, boiler.MechSkinWhere.BlueprintID.NIN(opts.ExcludeIDs))
		} else {
			queryMods = append(queryMods, boiler.MechSkinWhere.ID.NIN(opts.ExcludeIDs))
		}
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
	if opts.DisplayUnique {
		countQueryMods = append(countQueryMods, qm.Distinct(boiler.MechSkinTableColumns.BlueprintID))
	}

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
	selectCols := []string{}
	appendSelectCols := []string{
		qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.CollectionSlug),
		qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Hash),
		qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.TokenID),
		qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.OwnerID),
		qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
		qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.MarketLocked),
		qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.XsynLocked),
		qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.LockedToMarketplace),
		qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.AssetHidden),
		qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.ID),
		qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.EquippedOn),
		qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.LockedToMech),
		qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.GenesisTokenID),
		qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.LimitedReleaseTokenID),
		qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.Level),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.Label),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.DefaultLevel),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.Tier),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.BlueprintWeaponSkinID),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ImageURL),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.CardAnimationURL),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.AvatarURL),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.LargeImageURL),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.AnimationURL),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.YoutubeURL),
		qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.BackgroundColor),
		qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.ImageURL),
		qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.CardAnimationURL),
		qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.AvatarURL),
		qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.LargeImageURL),
		qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.AnimationURL),
		qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.YoutubeURL),
		qm.Rels(boiler.TableNames.MechModelSkinCompatibilities, boiler.MechModelSkinCompatibilityColumns.BackgroundColor),
	}
	if opts.DisplayUnique {
		selectCols = append(selectCols, fmt.Sprintf("DISTINCT ON (%[1]s) %[1]s", boiler.MechSkinTableColumns.BlueprintID))
	} else {
		selectCols = append(selectCols, qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.ID))
	}
	selectCols = append(selectCols, appendSelectCols...)

	queryMods = append(queryMods,
		qm.Select(selectCols...),
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
			if opts.DisplayUnique {
				orderBy = fmt.Sprintf("%s, (%s) %s",
					boiler.MechSkinTableColumns.BlueprintID,
					boiler.BlueprintMechSkinTableColumns.Label,
					opts.SortDir,
				)
			}
			queryMods = append(queryMods, qm.OrderBy(orderBy))
		case SortByRarity:
			queryMods = append(queryMods, GenerateTierSort(qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.Tier), opts.SortDir))
		case SortByDate:
			orderBy := fmt.Sprintf("(%s) %s",
				boiler.MechSkinTableColumns.CreatedAt,
				opts.SortDir,
			)
			if opts.DisplayUnique {
				orderBy = fmt.Sprintf("%s, (%s) %s",
					boiler.MechSkinTableColumns.BlueprintID,
					boiler.MechSkinTableColumns.CreatedAt,
					opts.SortDir,
				)
			}
			queryMods = append(queryMods, qm.OrderBy(orderBy))
		}
	} else {
		orderBy := fmt.Sprintf("%s ASC",
			qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.Label),
		)
		if opts.DisplayUnique {
			orderBy = fmt.Sprintf("%s, %s ASC",
				boiler.MechSkinTableColumns.BlueprintID,
				qm.Rels(boiler.TableNames.BlueprintMechSkin, boiler.BlueprintMechSkinColumns.Label),
			)
		}
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

func PlayerMechSkins(playerID string, modelID string, mechSkinIDs ...string) ([]*server.MechSkin, error) {
	playerIDWhere := ""
	if playerID != "" {
		_, err := uuid.FromString(playerID)
		if err != nil {
			return nil, terror.Error(err, "Invalid player id")
		}

		playerIDWhere = fmt.Sprintf(" AND %s = '%s'", boiler.CollectionItemTableColumns.OwnerID, playerID)
	}

	modelIDWhere := ""
	if modelID != "" {
		_, err := uuid.FromString(modelID)
		if err != nil {
			return nil, terror.Error(err, "Invalid model id.")
		}

		modelIDWhere = fmt.Sprintf(" AND %s = '%s'", boiler.MechModelSkinCompatibilityTableColumns.MechModelID, modelID)
	}

	mechSkinIDWhereIn := ""
	if len(mechSkinIDs) > 0 {
		mechSkinIDWhereIn = fmt.Sprintf(" AND %s IN (", boiler.CollectionItemTableColumns.ItemID)
		for i, mechSkinID := range mechSkinIDs {

			_, err := uuid.FromString(mechSkinID)
			if err != nil {
				return nil, terror.Error(err, "Invalid mech skin id.")
			}

			mechSkinIDWhereIn = "'" + mechSkinID + "'"

			if i < len(mechSkinIDs)-1 {
				mechSkinIDWhereIn += ","
				continue
			}

			mechSkinIDWhereIn += ")"
		}
	}

	queries := []qm.QueryMod{
		qm.Select(
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
			boiler.BlueprintMechSkinTableColumns.AvatarURL,
			boiler.BlueprintMechSkinTableColumns.CardAnimationURL,
			boiler.BlueprintMechSkinTableColumns.LargeImageURL,
			boiler.BlueprintMechSkinTableColumns.AnimationURL,
			boiler.BlueprintMechSkinTableColumns.YoutubeURL,
			boiler.BlueprintMechSkinTableColumns.BackgroundColor,

			fmt.Sprintf(
				"(SELECT TO_JSON(%s) FROM %s WHERE %s = %s %s LIMIT 1) AS images",
				boiler.TableNames.MechModelSkinCompatibilities,
				boiler.TableNames.MechModelSkinCompatibilities,
				boiler.MechModelSkinCompatibilityTableColumns.BlueprintMechSkinID,
				boiler.BlueprintMechSkinTableColumns.ID,
				modelIDWhere,
			),
		),

		qm.From(fmt.Sprintf(
			"(SELECT * FROM %s WHERE %s ISNULL AND %s = '%s' %s %s) %s",
			boiler.TableNames.CollectionItems,
			boiler.CollectionItemTableColumns.AssetHidden,
			boiler.CollectionItemTableColumns.ItemType,
			boiler.ItemTypeMechSkin,
			playerIDWhere,
			mechSkinIDWhereIn,
			boiler.TableNames.CollectionItems,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.MechSkin,
			boiler.MechSkinTableColumns.ID,
			boiler.CollectionItemTableColumns.ItemID,
		)),

		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintMechSkin,
			boiler.BlueprintMechSkinTableColumns.ID,
			boiler.MechSkinTableColumns.BlueprintID,
		)),
	}

	rows, err := boiler.NewQuery(queries...).Query(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to load mech skins.")
		return nil, terror.Error(err, "Failed to load mech skins")
	}

	result := []*server.MechSkin{}
	for rows.Next() {
		ms := &server.MechSkin{
			CollectionItem: &server.CollectionItem{},
			SkinSwatch:     &server.Images{},
		}

		images := &server.Images{}

		err = rows.Scan(
			&ms.CollectionItem.CollectionSlug,
			&ms.CollectionItem.Hash,
			&ms.CollectionItem.TokenID,
			&ms.CollectionItem.OwnerID,
			&ms.CollectionItem.ItemType,
			&ms.CollectionItem.MarketLocked,
			&ms.CollectionItem.XsynLocked,
			&ms.CollectionItem.LockedToMarketplace,
			&ms.CollectionItem.AssetHidden,

			&ms.ID,
			&ms.EquippedOn,
			&ms.LockedToMech,
			&ms.GenesisTokenID,
			&ms.LimitedReleaseTokenID,
			&ms.Level,

			&ms.Label,
			&ms.DefaultLevel,
			&ms.CollectionItem.Tier,
			&ms.BlueprintWeaponSkinID,

			&ms.SkinSwatch.ImageURL,
			&ms.SkinSwatch.AvatarURL,
			&ms.SkinSwatch.CardAnimationURL,
			&ms.SkinSwatch.LargeImageURL,
			&ms.SkinSwatch.AnimationURL,
			&ms.SkinSwatch.YoutubeURL,
			&ms.SkinSwatch.BackgroundColor,

			&images,
		)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to scan mech skin.")
			return nil, terror.Error(err, "Failed to scan mech skin.")
		}

		ms.Images = images

		result = append(result, ms)
	}

	return result, nil
}
