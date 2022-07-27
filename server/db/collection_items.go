package db

import (
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamelog"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ninja-software/terror/v2"
)

// InsertNewCollectionItem inserts a collection item,
// It takes a TX and DOES NOT COMMIT, commit needs to be called in the parent function.
func InsertNewCollectionItem(tx boil.Executor,
	collectionSlug,
	itemType,
	itemID,
	tier,
	ownerID string,
) (*boiler.CollectionItem, error) {
	item := &boiler.CollectionItem{}

	// I couldn't find the boiler enum types for some reason, so just doing strings
	tokenClause := ""
	switch collectionSlug {
	case "supremacy-general":
		tokenClause = "NEXTVAL('collection_general')"
	case "supremacy-limited-release":
		tokenClause = "NEXTVAL('collection_limited_release')"
	case "supremacy-genesis":
		tokenClause = "NEXTVAL('collection_genesis')"
	case "supremacy-consumables":
		tokenClause = "NEXTVAL('collection_consumables')"
	default:
		return nil, fmt.Errorf("invalid collection slug %s", collectionSlug)
	}

	query := fmt.Sprintf(`
		INSERT INTO collection_items(
			collection_slug, 
			token_id, 
			item_type, 
			item_id, 
			tier, 
			owner_id
			)
		VALUES($1, %s, $2, $3, $4, $5) RETURNING 
			id,
			collection_slug,
			hash,
			token_id,
			item_type,
			item_id,
			tier,
			owner_id,
			market_locked,
			xsyn_locked
			`, tokenClause)

	err := tx.QueryRow(query,
		collectionSlug,
		itemType,
		itemID,
		tier,
		ownerID,
	).Scan(&item.ID,
		&item.CollectionSlug,
		&item.Hash,
		&item.TokenID,
		&item.ItemType,
		&item.ItemID,
		&item.Tier,
		&item.OwnerID,
		&item.MarketLocked,
		&item.XsynLocked,
	)

	if err != nil {
		gamelog.L.Error().Err(err).
			Str("itemType", itemType).
			Str("itemID", itemID).
			Str("tier", tier).
			Str("ownerID", ownerID).
			Msg("failed to insert new collection item")
		return nil, terror.Error(err)
	}

	return item, nil
}

func CollectionItemFromItemID(tx boil.Executor, id string) (*server.CollectionItem, error) {
	ci, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(tx)
	if err != nil {
		return nil, terror.Error(err)
	}

	return &server.CollectionItem{
		CollectionSlug: ci.CollectionSlug,
		Hash:           ci.Hash,
		TokenID:        ci.TokenID,
		ItemType:       ci.ItemType,
		ItemID:         ci.ItemID,
		Tier:           ci.Tier,
		OwnerID:        ci.OwnerID,
		XsynLocked:     ci.XsynLocked,
		MarketLocked:   ci.MarketLocked,
	}, nil
}

func CollectionItemFromBoiler(ci *boiler.CollectionItem) *server.CollectionItem {
	return &server.CollectionItem{
		CollectionSlug: ci.CollectionSlug,
		Hash:           ci.Hash,
		TokenID:        ci.TokenID,
		ItemType:       ci.ItemType,
		ItemID:         ci.ItemID,
		Tier:           ci.Tier,
		OwnerID:        ci.OwnerID,
		XsynLocked:     ci.XsynLocked,
		MarketLocked:   ci.MarketLocked,
	}
}

func GenerateTierSort(col string, sortDir SortByDir) qm.QueryMod {
	return qm.OrderBy(fmt.Sprintf(`(
		CASE %s
			WHEN 'MEGA' THEN 1
			WHEN 'COLOSSAL' THEN 2
			WHEN 'RARE' THEN 3
			WHEN 'LEGENDARY' THEN 4
			WHEN 'ELITE_LEGENDARY' THEN 5
			WHEN 'ULTRA_RARE' THEN 6
			WHEN 'EXOTIC' THEN 7
			WHEN 'GUARDIAN' THEN 8
			WHEN 'MYTHIC' THEN 9
			WHEN 'DEUS_EX' THEN 10
			WHEN 'TITAN' THEN 11
		END
	) %s NULLS LAST`, col, sortDir))
}
