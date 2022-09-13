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

type ForbiddenAssetModificationReason int8

const (
	ForbiddenAssetModificationReasonInvalid     ForbiddenAssetModificationReason = 0
	ForbiddenAssetModificationReasonMarketplace ForbiddenAssetModificationReason = 1
	ForbiddenAssetModificationReasonXsyn        ForbiddenAssetModificationReason = 2
	ForbiddenAssetModificationReasonQueue       ForbiddenAssetModificationReason = 3
	ForbiddenAssetModificationReasonBattle      ForbiddenAssetModificationReason = 4
	ForbiddenAssetModificationReasonOwner       ForbiddenAssetModificationReason = 5
)

func (f ForbiddenAssetModificationReason) String() string {
	switch f {
	case ForbiddenAssetModificationReasonMarketplace:
		return "The asset is currently being listed on the marketplace."
	case ForbiddenAssetModificationReasonXsyn:
		return "The asset is currently locked to the XSYN ecosystem."
	case ForbiddenAssetModificationReasonQueue:
		return "The asset is currently in the battle queue."
	case ForbiddenAssetModificationReasonBattle:
		return "The asset is currently in a battle."
	case ForbiddenAssetModificationReasonOwner:
		return "You do not own this asset."
	}
	return "The asset cannot be modified, unequipped, or equipped."
}

func IsValidCollectionItemType(itemType string) bool {
	switch itemType {
	case boiler.ItemTypeUtility,
		boiler.ItemTypeWeapon,
		boiler.ItemTypeMech,
		boiler.ItemTypeMechSkin,
		boiler.ItemTypeMechAnimation,
		boiler.ItemTypePowerCore,
		boiler.ItemTypeMysteryCrate,
		boiler.ItemTypeWeaponSkin:
		return true
	}
	return false
}

func CanAssetBeModifiedOrMoved(exec boil.Executor, itemID string, itemType string, ownerID ...string) (bool, ForbiddenAssetModificationReason, error) {
	if !IsValidCollectionItemType(itemType) {
		return false, -1, fmt.Errorf("unknown collection item type")
	}

	if itemType == boiler.ItemTypeMysteryCrate || itemType == boiler.ItemTypeWeaponSkin {
		return false, ForbiddenAssetModificationReasonInvalid, fmt.Errorf("invalid collection item type specified")
	}

	ci, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(itemID),
		boiler.CollectionItemWhere.ItemType.EQ(itemType),
	).One(exec)
	if err != nil {
		return false, -1, err
	}

	if len(ownerID) > 0 && ownerID[0] != "" && ci.OwnerID != ownerID[0] {
		return false, ForbiddenAssetModificationReasonOwner, nil
	}

	if ci.LockedToMarketplace {
		return false, ForbiddenAssetModificationReasonMarketplace, nil
	}
	if ci.XsynLocked {
		return false, ForbiddenAssetModificationReasonXsyn, nil
	}

	switch itemType {
	case boiler.ItemTypeUtility:
		utility, err := boiler.FindUtility(exec, itemID)
		if err != nil {
			return false, -1, err
		}
		if utility.EquippedOn.Valid {
			return CanAssetBeModifiedOrMoved(exec, utility.EquippedOn.String, boiler.ItemTypeMech)
		}
	case boiler.ItemTypeWeapon:
		weapon, err := boiler.FindWeapon(exec, itemID)
		if err != nil {
			return false, -1, err
		}
		if weapon.EquippedOn.Valid {
			return CanAssetBeModifiedOrMoved(exec, weapon.EquippedOn.String, boiler.ItemTypeMech)
		}
	case boiler.ItemTypeMech:
		mechStatus, err := GetCollectionItemStatus(*ci)
		if err != nil {
			return false, -1, err
		}
		if mechStatus.Status == server.MechArenaStatusBattle {
			return false, ForbiddenAssetModificationReasonBattle, nil
		}
	// case boiler.ItemTypeMechSkin:
	// case boiler.ItemTypeMechAnimation:
	case boiler.ItemTypePowerCore:
		powerCore, err := boiler.FindPowerCore(exec, itemID)
		if err != nil {
			return false, -1, err
		}
		if powerCore.EquippedOn.Valid {
			return CanAssetBeModifiedOrMoved(exec, powerCore.EquippedOn.String, boiler.ItemTypePowerCore)
		}
	}

	return true, -1, nil
}

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

	if tier == "" {
		tier = "MEGA"
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
