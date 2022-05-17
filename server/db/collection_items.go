package db

import (
	"database/sql"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

	"github.com/ninja-software/terror/v2"
)

// InsertNewCollectionItem inserts a collection item,
// It takes a TX and DOES NOT COMMIT, commit needs to be called in the parent function.
func InsertNewCollectionItem(tx *sql.Tx, collectionSlug, itemType, itemID, tier, ownerID string) error {
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
		return fmt.Errorf("invalid collection slug %s", collectionSlug)
	}

	query := fmt.Sprintf(`
		INSERT INTO collection_items(collection_slug, token_id, item_type, item_id, tier, owner_id)
		VALUES($1, %s, $2, $3, $4, $5)`, tokenClause)

	_, err := tx.Exec(query, collectionSlug, itemType, itemID, tier, ownerID)
	if err != nil {
		gamelog.L.Error().Err(err).
		Str("itemType", itemType).
		Str("itemID", itemID).
		Str("tier", tier).
		Str("ownerID", ownerID).
		Msg("failed to insert new collection item")
		return terror.Error(err)
	}

	return nil
}

func CollectionItemFromItemID(id string) (*server.CollectionDetails, error) {
	ci, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}

	return &server.CollectionDetails{
		CollectionSlug: ci.CollectionSlug,
		Hash:           ci.Hash,
		TokenID:        ci.TokenID,
		ItemType:       ci.ItemType,
		ItemID:         ci.ItemID,
		Tier:           ci.Tier,
		OwnerID:        ci.OwnerID,
		OnChainStatus:  ci.OnChainStatus,
	}, nil
}
