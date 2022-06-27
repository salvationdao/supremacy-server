package db

import (
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"

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
	imageURL,
	cardAnimationURL,
	avatarURL,
	largeImageURL,
	backgroundURL,
	animationURL,
	youtubeURL null.String) (*boiler.CollectionItem, error) {
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
			owner_id,
			image_url,
			card_animation_url,
			avatar_url,
			large_image_url,
			background_color,
			animation_url,
			youtube_url
			)
		VALUES($1, %s, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING 
			id,
			collection_slug,
			hash,
			token_id,
			item_type,
			item_id,
			tier,
			owner_id,
			market_locked,
			xsyn_locked,
			image_url,
			card_animation_url,
			avatar_url,
			large_image_url,
			background_color,
			animation_url,
			youtube_url
			`, tokenClause)

	err := tx.QueryRow(query,
		collectionSlug,
		itemType,
		itemID,
		tier,
		ownerID,
		imageURL,
		cardAnimationURL,
		avatarURL,
		largeImageURL,
		backgroundURL,
		animationURL,
		youtubeURL,
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
		&item.ImageURL,
		&item.CardAnimationURL,
		&item.AvatarURL,
		&item.LargeImageURL,
		&item.BackgroundColor,
		&item.AnimationURL,
		&item.YoutubeURL)

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

func CollectionItemFromItemID(id string) (*server.CollectionItem, error) {
	ci, err := boiler.CollectionItems(boiler.CollectionItemWhere.ItemID.EQ(id)).One(gamedb.StdConn)
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
		CollectionSlug:   ci.CollectionSlug,
		Hash:             ci.Hash,
		TokenID:          ci.TokenID,
		ItemType:         ci.ItemType,
		ItemID:           ci.ItemID,
		Tier:             ci.Tier,
		OwnerID:          ci.OwnerID,
		XsynLocked:       ci.XsynLocked,
		MarketLocked:     ci.MarketLocked,
		ImageURL:         ci.ImageURL,
		CardAnimationURL: ci.CardAnimationURL,
		AvatarURL:        ci.AvatarURL,
		LargeImageURL:    ci.LargeImageURL,
		BackgroundColor:  ci.BackgroundColor,
		AnimationURL:     ci.AnimationURL,
		YoutubeURL:       ci.YoutubeURL,
	}
}
