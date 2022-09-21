package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

const soldToTable = "st"
const bidderTable = "bidder"
const weaponSkinColTable = "wsc"
const mechSkinColTable = "msc"

var ItemSaleQueryMods = []qm.QueryMod{
	qm.Select(
		fmt.Sprintf(`%s AS id`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ID)),
		fmt.Sprintf(`%s AS faction_id`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.FactionID)),
		fmt.Sprintf(`%s AS collection_item_id`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.CollectionItemID)),
		fmt.Sprintf(`%s AS listing_fee_tx_id`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ListingFeeTXID)),
		fmt.Sprintf(`%s AS owner_id`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.OwnerID)),
		fmt.Sprintf(`%s AS auction`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.Auction)),
		fmt.Sprintf(`%s AS auction_current_price`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.AuctionCurrentPrice)),
		fmt.Sprintf(`%s AS auction_reserved_price`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.AuctionReservedPrice)),
		fmt.Sprintf(`%s AS buyout`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.Buyout)),
		fmt.Sprintf(`%s AS buyout_price`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.BuyoutPrice)),
		fmt.Sprintf(`%s AS dutch_auction`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.DutchAuction)),
		fmt.Sprintf(`%s AS dutch_auction_drop_rate`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.DutchAuctionDropRate)),
		fmt.Sprintf(`%s AS end_at`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.EndAt)),
		fmt.Sprintf(`%s AS sold_at`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.SoldAt)),
		fmt.Sprintf(`%s AS sold_for`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.SoldFor)),
		fmt.Sprintf(`%s "sold_to.id"`, qm.Rels(soldToTable, boiler.PlayerColumns.ID)),
		fmt.Sprintf(`%s "sold_to.username"`, qm.Rels(soldToTable, boiler.PlayerColumns.Username)),
		fmt.Sprintf(`%s "sold_to.faction_id"`, qm.Rels(soldToTable, boiler.PlayerColumns.FactionID)),
		fmt.Sprintf(`%s "sold_to.public_address"`, qm.Rels(soldToTable, boiler.PlayerColumns.PublicAddress)),
		fmt.Sprintf(`%s "sold_to.gid"`, qm.Rels(soldToTable, boiler.PlayerColumns.Gid)),
		fmt.Sprintf(`%s AS sold_tx_id`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.SoldTXID)),
		fmt.Sprintf(`%s AS sold_fee_tx_id`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.SoldFeeTXID)),
		fmt.Sprintf(`%s AS deleted_at`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.DeletedAt)),
		fmt.Sprintf(`%s AS updated_at`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.UpdatedAt)),
		fmt.Sprintf(`%s AS created_at`, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.CreatedAt)),
		fmt.Sprintf(`%s AS collection_item_type`, qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType)),
		fmt.Sprintf(`(SELECT COUNT(*) FROM %s _isbh WHERE _isbh.%s = %s) AS total_bids`, boiler.TableNames.ItemSalesBidHistory, boiler.ItemSalesBidHistoryColumns.ItemSaleID, qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ID)),
		fmt.Sprintf(`%s AS "players.id"`, qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID)),
		fmt.Sprintf(`%s AS "players.username"`, qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Username)),
		fmt.Sprintf(`%s AS "players.faction_id"`, qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.FactionID)),
		fmt.Sprintf(`%s AS "players.public_address"`, qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.PublicAddress)),
		fmt.Sprintf(`%s AS "players.gid"`, qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Gid)),
		fmt.Sprintf(`%s AS "mechs.id"`, qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID)),
		fmt.Sprintf(`%s AS "mechs.name"`, qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name)),
		fmt.Sprintf(`%s "mechs.label"`, qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label)),
		fmt.Sprintf(`%s AS "mystery_crate.id"`, qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.ID)),
		fmt.Sprintf(`%s AS "mystery_crate.label"`, qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.Label)),
		fmt.Sprintf(`%s AS "mystery_crate.description"`, qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.Description)),
		qm.Rels(boiler.TableNames.StorefrontMysteryCrates, boiler.StorefrontMysteryCrateColumns.ImageURL),
		qm.Rels(boiler.TableNames.StorefrontMysteryCrates, boiler.StorefrontMysteryCrateColumns.CardAnimationURL),
		qm.Rels(boiler.TableNames.StorefrontMysteryCrates, boiler.StorefrontMysteryCrateColumns.LargeImageURL),
		qm.Rels(boiler.TableNames.StorefrontMysteryCrates, boiler.StorefrontMysteryCrateColumns.AnimationURL),
		fmt.Sprintf(`%s AS "weapons.id"`, qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID)),
		fmt.Sprintf(`%s AS "weapons.label"`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label)),
		fmt.Sprintf(`%s AS "weapons.weapon_type"`, qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.WeaponType)),
		fmt.Sprintf(`
		(
			CASE
				WHEN %[1]s = 'weapon' THEN COALESCE(%[2]s, %[3]s)
				WHEN %[1]s = 'mech' THEN COALESCE(%[4]s, %[3]s)
				ELSE %[3]s
			END
		) AS "collection_items.tier"`,
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
			qm.Rels(weaponSkinColTable, boiler.CollectionItemColumns.Tier),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Tier),
			qm.Rels(mechSkinColTable, boiler.CollectionItemColumns.Tier),
		),
		fmt.Sprintf(`%s AS "collection_items.xsyn_locked"`, qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.XsynLocked)),
		fmt.Sprintf(`%s AS "collection_items.market_locked"`, qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.MarketLocked)),
		fmt.Sprintf(`%s AS "collection_items.hash"`, qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Hash)),
		fmt.Sprintf(`%s AS "bidder.id"`, qm.Rels(bidderTable, boiler.PlayerColumns.ID)),
		fmt.Sprintf(`%s AS "bidder.username"`, qm.Rels(bidderTable, boiler.PlayerColumns.Username)),
		fmt.Sprintf(`%s AS "bidder.faction_id"`, qm.Rels(bidderTable, boiler.PlayerColumns.FactionID)),
		fmt.Sprintf(`%s AS "bidder.public_address"`, qm.Rels(bidderTable, boiler.PlayerColumns.PublicAddress)),
		fmt.Sprintf(`%s AS "bidder.gid"`, qm.Rels(bidderTable, boiler.PlayerColumns.Gid)),
	),
	// collection items
	qm.InnerJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.CollectionItems,
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ID),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.CollectionItemID),
		),
	),
	// mechs
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s AND %s = ?",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
		),
		boiler.ItemTypeMech,
	),
	// mech skin
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.MechSkin,
			qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
		),
	),
	// mech blueprint
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintMechs,
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
		),
	),
	// mech skin collection items
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s %s ON %s = %s AND %s = ?",
			boiler.TableNames.CollectionItems,
			mechSkinColTable,
			qm.Rels(mechSkinColTable, boiler.CollectionItemColumns.ItemID),
			qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.ID),
			qm.Rels(mechSkinColTable, boiler.CollectionItemColumns.ItemType),
		),
		boiler.ItemTypeMechSkin,
	),
	// mystery crate
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s AND %s = ?",
			boiler.TableNames.MysteryCrate,
			qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
		),
		boiler.ItemTypeMysteryCrate,
	),
	// storefront mystery crate (for images)
	qm.LeftOuterJoin(fmt.Sprintf("%s ON %s = %s",
		boiler.TableNames.StorefrontMysteryCrates,
		qm.Rels(boiler.TableNames.StorefrontMysteryCrates, boiler.StorefrontMysteryCrateColumns.ID),
		qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.BlueprintID),
	)),
	// weapons
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s AND %s = ?",
			boiler.TableNames.Weapons,
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
		),
		boiler.ItemTypeWeapon,
	),
	// blueprint weapons
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintWeapons,
			qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.BlueprintID),
		),
	),
	// weapon skin
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.WeaponSkin,
			qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.ID),
			qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.EquippedWeaponSkinID),
		),
	),
	// weapon skin collection item
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s %s ON %s = %s AND %s = ?",
			boiler.TableNames.CollectionItems,
			weaponSkinColTable,
			qm.Rels(weaponSkinColTable, boiler.CollectionItemColumns.ItemID),
			qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.ID),
			qm.Rels(weaponSkinColTable, boiler.CollectionItemColumns.ItemType),
		),
		boiler.ItemTypeWeaponSkin,
	),
	// blueprint weapon skin
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintWeaponSkin,
			qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
			qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.BlueprintID),
		),
	),
	// players
	qm.InnerJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.Players,
			qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.OwnerID),
		),
	),
	// sold to player
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s AS %s ON %s = %s",
			boiler.TableNames.Players,
			soldToTable,
			qm.Rels(soldToTable, boiler.PlayerColumns.ID),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.SoldTo),
		),
	),
	// Last Auction Bidder
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s AND %s IS NULL AND %s IS NULL",
			boiler.TableNames.ItemSalesBidHistory,
			qm.Rels(boiler.TableNames.ItemSalesBidHistory, boiler.ItemSalesBidHistoryColumns.ItemSaleID),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ID),
			qm.Rels(boiler.TableNames.ItemSalesBidHistory, boiler.ItemSalesBidHistoryColumns.CancelledAt),
			qm.Rels(boiler.TableNames.ItemSalesBidHistory, boiler.ItemSalesBidHistoryColumns.RefundBidTXID),
		),
	),
	// bidder player
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s AS %s ON %s = %s",
			boiler.TableNames.Players,
			bidderTable,
			qm.Rels(bidderTable, boiler.PlayerColumns.ID),
			qm.Rels(boiler.TableNames.ItemSalesBidHistory, boiler.ItemSalesBidHistoryColumns.BidderID),
		),
	),
}

var ItemKeycardSaleQueryMods = []qm.QueryMod{
	qm.Select(
		`item_keycard_sales.id,
		item_keycard_sales.faction_id,
		item_keycard_sales.item_id,
		item_keycard_sales.listing_fee_tx_id,
		item_keycard_sales.owner_id,
		item_keycard_sales.buyout_price,
		item_keycard_sales.end_at,
		item_keycard_sales.sold_at,
		item_keycard_sales.sold_for,
		item_keycard_sales.sold_tx_id,
		item_keycard_sales.sold_fee_tx_id,
		item_keycard_sales.deleted_at,
		item_keycard_sales.updated_at,
		item_keycard_sales.created_at,
		players.id AS "players.id",
		players.faction_id AS "players.faction_id",
		players.username AS "players.username",
		players.public_address AS "players.public_address",
		players.gid AS "players.gid",
		st.id AS "sold_to.id",
		st.username AS "sold_to.username",
		st.faction_id AS "sold_to.faction_id",
		st.public_address AS "sold_to.public_address",
		st.gid AS "sold_to.gid",
		blueprint_keycards.id AS "blueprint_keycards.id",
		blueprint_keycards.label AS "blueprint_keycards.label",
		blueprint_keycards.description AS "blueprint_keycards.description",
		blueprint_keycards.collection AS "blueprint_keycards.collection",
		blueprint_keycards.keycard_token_id AS "blueprint_keycards.keycard_token_id",
		blueprint_keycards.image_url AS "blueprint_keycards.image_url",
		blueprint_keycards.animation_url AS "blueprint_keycards.animation_url",
		blueprint_keycards.keycard_group AS "blueprint_keycards.keycard_group",
		blueprint_keycards.syndicate AS "blueprint_keycards.syndicate",
		blueprint_keycards.created_at AS "blueprint_keycards.created_at"`,
	),
	qm.InnerJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.PlayerKeycards,
			qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.ID),
			qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.ItemID),
		),
	),
	qm.InnerJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.BlueprintKeycards,
			qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.ID),
			qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.BlueprintKeycardID),
		),
	),
	qm.InnerJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.Players,
			qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
			qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.OwnerID),
		),
	),
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s AS st ON %s = %s",
			boiler.TableNames.Players,
			qm.Rels("st", boiler.PlayerColumns.ID),
			qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.SoldTo),
		),
	),
}

var ItemSaleOtherAssetsSQL = `
SELECT (
	CASE ci.item_type
		WHEN 'mech' THEN (
			SELECT array_agg(_ci.hash)
			FROM (
				(
					SELECT 'mech_skin'::item_type AS item_type, _ms.id as item_id
					FROM mechs _m 
						INNER JOIN mech_skin _ms on _ms.id = _m.chassis_skin_id 
					WHERE _m.id = ci.item_id
				)
				UNION ALL 
				(
					SELECT 'weapon'::item_type AS item_type, _mw.weapon_id AS item_id 
					FROM mech_weapons _mw 
					WHERE _mw.chassis_id = ci.item_id
				)
				UNION ALL 
				(
					SELECT 'utility'::item_type AS item_type, _mu.utility_id AS item_id 
					FROM mech_utility _mu 
					WHERE _mu.chassis_id = ci.item_id
				)
				UNION ALL 
				(
					SELECT 'power_core'::item_type AS item_type, _pc.id AS item_id 
					FROM mechs _m 
						INNER JOIN power_cores _pc ON _pc.id = _m.power_core_id
					WHERE _m.id = ci.item_id
				)
			) a 
				INNER JOIN collection_items _ci on _ci.item_id = a.item_id AND _ci.item_type = a.item_type
		)
		WHEN 'weapon' THEN (
			SELECT array_agg( _ci.hash)
			FROM weapons _w 
				INNER JOIN weapon_skin _ws on _ws.id = _w.equipped_weapon_skin_id
				INNER JOIN collection_items _ci on _ci.item_id = _ws.id AND _ci.item_type = 'weapon_skin'
			WHERE _w.id = ci.item_id
		)
		ELSE ARRAY[]::text[]
	END)	
FROM item_sales s 
	LEFT JOIN collection_items ci ON ci.id = s.collection_item_id 
	LEFT JOIN weapons w ON w.id = ci.item_id AND ci.item_type = 'weapon'
	LEFT JOIN mechs m ON m.id = ci.item_id AND ci.item_type = 'mech'
WHERE s.id = $1
	AND (
	    (ci.item_type = 'weapon' AND w.id IS NOT NULL AND w.genesis_token_id IS NULL)
	    OR (ci.item_type = 'mech' AND m.id IS NOT NULL AND m.genesis_token_id IS NULL)
    )`

// MarketplaceItemSale gets a specific item sale.
func MarketplaceItemSale(id uuid.UUID) (*server.MarketplaceSaleItem, error) {
	output := &server.MarketplaceSaleItem{}
	err := boiler.ItemSales(
		append(
			ItemSaleQueryMods,
			boiler.ItemSaleWhere.ID.EQ(id.String()),
		)...,
	).QueryRow(gamedb.StdConn).Scan(
		&output.ID,
		&output.FactionID,
		&output.CollectionItemID,
		&output.ListingFeeTXID,
		&output.OwnerID,
		&output.Auction,
		&output.AuctionCurrentPrice,
		&output.AuctionReservedPrice,
		&output.Buyout,
		&output.BuyoutPrice,
		&output.DutchAuction,
		&output.DutchAuctionDropRate,
		&output.EndAt,
		&output.SoldAt,
		&output.SoldFor,
		&output.SoldTo.ID,
		&output.SoldTo.Username,
		&output.SoldTo.FactionID,
		&output.SoldTo.PublicAddress,
		&output.SoldTo.Gid,
		&output.SoldTXID,
		&output.SoldFeeTXID,
		&output.DeletedAt,
		&output.UpdatedAt,
		&output.CreatedAt,
		&output.CollectionItemType,
		&output.TotalBids,
		&output.Owner.ID,
		&output.Owner.Username,
		&output.Owner.FactionID,
		&output.Owner.PublicAddress,
		&output.Owner.Gid,
		&output.Mech.ID,
		&output.Mech.Name,
		&output.Mech.Label,
		&output.MysteryCrate.ID,
		&output.MysteryCrate.Label,
		&output.MysteryCrate.Description,
		&output.MysteryCrate.ImageURL,
		&output.MysteryCrate.LargeImageURL,
		&output.MysteryCrate.CardAnimationURL,
		&output.MysteryCrate.AnimationURL,
		&output.Weapon.ID,
		&output.Weapon.Label,
		&output.Weapon.WeaponType,
		&output.CollectionItem.Tier,
		&output.CollectionItem.XsynLocked,
		&output.CollectionItem.MarketLocked,
		&output.CollectionItem.Hash,
		&output.LastBid.ID,
		&output.LastBid.Username,
		&output.LastBid.FactionID,
		&output.LastBid.PublicAddress,
		&output.LastBid.Gid,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return output, nil
}

// MarketplaceItemKeycardSale gets a specific keycard item sale.
func MarketplaceItemKeycardSale(id uuid.UUID) (*server.MarketplaceSaleItem1155, error) {
	output := &server.MarketplaceSaleItem1155{}
	err := boiler.ItemKeycardSales(
		append(
			ItemKeycardSaleQueryMods,
			boiler.ItemKeycardSaleWhere.ID.EQ(id.String()),
		)...,
	).QueryRow(gamedb.StdConn).Scan(
		&output.ID,
		&output.FactionID,
		&output.ItemID,
		&output.ListingFeeTXID,
		&output.OwnerID,
		&output.BuyoutPrice,
		&output.EndAt,
		&output.SoldAt,
		&output.SoldFor,
		&output.SoldTXID,
		&output.SoldFeeTXID,
		&output.DeletedAt,
		&output.UpdatedAt,
		&output.CreatedAt,
		&output.Owner.ID,
		&output.Owner.FactionID,
		&output.Owner.Username,
		&output.Owner.PublicAddress,
		&output.Owner.Gid,
		&output.SoldTo.ID,
		&output.SoldTo.Username,
		&output.SoldTo.FactionID,
		&output.SoldTo.PublicAddress,
		&output.SoldTo.Gid,
		&output.Keycard.ID,
		&output.Keycard.Label,
		&output.Keycard.Description,
		&output.Keycard.Collection,
		&output.Keycard.KeycardTokenID,
		&output.Keycard.ImageURL,
		&output.Keycard.AnimationURL,
		&output.Keycard.KeycardGroup,
		&output.Keycard.Syndicate,
		&output.Keycard.CreatedAt,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return output, nil
}

type MarketplaceWeaponStatFilter struct {
	FilterStatAmmo                *WeaponStatFilterRange `json:"ammo"`
	FilterStatDamage              *WeaponStatFilterRange `json:"damage"`
	FilterStatDamageFalloff       *WeaponStatFilterRange `json:"damage_falloff"`
	FilterStatDamageFalloffRate   *WeaponStatFilterRange `json:"damage_falloff_rate"`
	FilterStatRadius              *WeaponStatFilterRange `json:"radius"`
	FilterStatRadiusDamageFalloff *WeaponStatFilterRange `json:"radius_damage_falloff"`
	FilterStatRateOfFire          *WeaponStatFilterRange `json:"rate_of_fire"`
	FilterStatEnergyCosts         *WeaponStatFilterRange `json:"energy_cost"`
	FilterStatProjectileSpeed     *WeaponStatFilterRange `json:"projectile_speed"`
	FilterStatSpread              *WeaponStatFilterRange `json:"spread"`
}

// MarketplaceItemSaleList returns a numeric paginated result of sales list.
func MarketplaceItemSaleList(
	itemType string,
	userID string,
	factionID string,
	search string,
	rarities []string,
	saleTypes []string,
	weaponTypes []string,
	weaponStats *MarketplaceWeaponStatFilter,
	ownedBy []string,
	sold bool,
	minPrice decimal.NullDecimal,
	maxPrice decimal.NullDecimal,
	offset int,
	pageSize int,
	sortBy string,
	sortDir SortByDir,
) (int64, []*server.MarketplaceSaleItem, error) {
	if !sortDir.IsValid() {
		return 0, nil, terror.Error(fmt.Errorf("invalid sort direction"))
	}

	queryMods := append(
		ItemSaleQueryMods,
		boiler.ItemSaleWhere.DeletedAt.IsNull(),
	)
	if factionID != "" {
		queryMods = append(queryMods, boiler.ItemSaleWhere.FactionID.EQ(factionID))
	}
	if itemType != "" {
		queryMods = append(queryMods, boiler.CollectionItemWhere.ItemType.EQ(itemType))
	}

	// Filters
	if len(rarities) > 0 {
		vals := []interface{}{}
		for _, v := range rarities {
			vals = append(vals, v)
		}
		queryMods = append(queryMods, qm.Expr(
			qm.Or2(qm.WhereIn(qm.Rels(mechSkinColTable, boiler.CollectionItemColumns.Tier)+" IN ?", vals...)),
			qm.Or2(qm.WhereIn(qm.Rels(weaponSkinColTable, boiler.CollectionItemColumns.Tier)+" IN ?", vals...)),
		))
	}
	if len(saleTypes) > 0 {
		saleTypeConditions := []qm.QueryMod{}
		for _, st := range saleTypes {
			switch st {
			case "BUY_NOW":
				saleTypeConditions = append(saleTypeConditions, qm.Or2(boiler.ItemSaleWhere.Buyout.EQ(true)))
			case "AUCTION":
				saleTypeConditions = append(saleTypeConditions, qm.Or2(boiler.ItemSaleWhere.Auction.EQ(true)))
			case "DUTCH_AUCTION":
				saleTypeConditions = append(saleTypeConditions, qm.Or2(boiler.ItemSaleWhere.DutchAuction.EQ(true)))
			}
		}
		queryMods = append(queryMods, qm.Expr(saleTypeConditions...))
	}
	if len(ownedBy) > 0 {
		isSelf := false
		isOthers := false
		for _, ownerType := range ownedBy {
			if ownerType == "self" {
				isSelf = true
			} else if ownerType == "others" {
				isOthers = true
			}
			if isSelf && isOthers {
				break
			}
		}
		if isSelf && !isOthers {
			queryMods = append(queryMods, boiler.ItemSaleWhere.OwnerID.EQ(userID))
		} else if !isSelf && isOthers {
			queryMods = append(queryMods, boiler.ItemSaleWhere.OwnerID.NEQ(userID))
		}
	}

	if len(weaponTypes) > 0 {
		queryMods = append(queryMods, boiler.BlueprintWeaponWhere.WeaponType.IN(weaponTypes))
	}
	if weaponStats != nil {
		if weaponStats.FilterStatAmmo != nil {
			queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.MaxAmmo, weaponStats.FilterStatAmmo)...)
		}
		if weaponStats.FilterStatDamage != nil {
			queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.Damage, weaponStats.FilterStatDamage)...)
		}
		if weaponStats.FilterStatDamageFalloff != nil {
			queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.DamageFalloff, weaponStats.FilterStatDamageFalloff)...)
		}
		if weaponStats.FilterStatDamageFalloffRate != nil {
			queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.DamageFalloffRate, weaponStats.FilterStatDamageFalloffRate)...)
		}
		if weaponStats.FilterStatRadius != nil {
			queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.Radius, weaponStats.FilterStatRadius)...)
		}
		if weaponStats.FilterStatRadiusDamageFalloff != nil {
			queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.RadiusDamageFalloff, weaponStats.FilterStatRadiusDamageFalloff)...)
		}
		if weaponStats.FilterStatRateOfFire != nil {
			queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.RateOfFire, weaponStats.FilterStatRateOfFire)...)
		}
		if weaponStats.FilterStatEnergyCosts != nil {
			queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.PowerCost, weaponStats.FilterStatEnergyCosts)...)
		}
		if weaponStats.FilterStatProjectileSpeed != nil {
			queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.ProjectileSpeed, weaponStats.FilterStatProjectileSpeed)...)
		}
		if weaponStats.FilterStatSpread != nil {
			queryMods = append(queryMods, GenerateWeaponStatFilterQueryMods(boiler.BlueprintWeaponColumns.Spread, weaponStats.FilterStatSpread)...)
		}
	}

	if minPrice.Valid {
		value := decimal.NewNullDecimal(minPrice.Decimal.Mul(decimal.New(1, 18)))
		queryMods = append(queryMods, qm.Expr(
			qm.Or2(
				qm.Expr(
					boiler.ItemSaleWhere.Buyout.EQ(true),
					boiler.ItemSaleWhere.BuyoutPrice.GTE(value),
				),
			),
			qm.Or2(
				qm.Expr(
					boiler.ItemSaleWhere.Auction.EQ(true),
					boiler.ItemSaleWhere.AuctionCurrentPrice.GTE(value),
				),
			),
		))
	}
	if maxPrice.Valid {
		value := decimal.NewNullDecimal(maxPrice.Decimal.Mul(decimal.New(1, 18)))
		queryMods = append(queryMods, qm.Expr(
			qm.Or2(
				qm.Expr(
					boiler.ItemSaleWhere.Buyout.EQ(true),
					boiler.ItemSaleWhere.BuyoutPrice.LTE(value),
				),
			),
			qm.Or2(
				qm.Expr(
					boiler.ItemSaleWhere.Auction.EQ(true),
					boiler.ItemSaleWhere.AuctionCurrentPrice.LTE(value),
				),
			),
		))
	}
	if sold {
		queryMods = append(queryMods, boiler.ItemSaleWhere.SoldAt.IsNotNull())
	} else {
		queryMods = append(queryMods,
			boiler.ItemSaleWhere.SoldAt.IsNull(),
			boiler.ItemSaleWhere.EndAt.GT(time.Now()),
			boiler.CollectionItemWhere.XsynLocked.EQ(false),
			boiler.CollectionItemWhere.MarketLocked.EQ(false),
		)
	}

	// Search
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			queryMods = append(queryMods, qm.And(
				fmt.Sprintf(
					`(
						(to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s::text) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
					)`,
					qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
					qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.Label),
					qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.WeaponType),
					qm.Rels(mechSkinColTable, boiler.CollectionItemColumns.Tier),
					qm.Rels(weaponSkinColTable, boiler.CollectionItemColumns.Tier),
					qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Username),
				),
				xsearch,
				xsearch,
				xsearch,
				xsearch,
				xsearch,
				xsearch,
				xsearch,
			))
		}
	}

	// Get total rows
	total, err := boiler.ItemSales(queryMods...).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if total == 0 {
		return 0, []*server.MarketplaceSaleItem{}, nil
	}

	// Sort
	var orderBy qm.QueryMod
	if sortBy == "alphabetical" {
		orderBy = qm.OrderBy(fmt.Sprintf("COALESCE(mechs.label, mechs.name, weapons.label) %s", sortDir))
	} else if sortBy == "time" {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.EndAt), sortDir))
	} else if sortBy == "rarity" {
		orderBy = GenerateTierSort(fmt.Sprintf("COALESCE(%s, %s)", qm.Rels(mechSkinColTable, boiler.CollectionItemColumns.Tier), qm.Rels(weaponSkinColTable, boiler.CollectionItemColumns.Tier)), sortDir)
	} else if sortBy == "price" && sold {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.SoldFor), sortDir))
	} else if sortBy == "price" && !sold {
		extractPriceFunc := "least"
		if sortDir == SortByDirDesc {
			extractPriceFunc = "greatest"
		}
		orderBy = qm.OrderBy(fmt.Sprintf(
			`%[1]s(
				%[2]s,
				CASE
					WHEN %[4]s = TRUE THEN %[3]s - (%[5]s * floor(extract(epoch FROM (least(%[6]s, now()) - %[7]s)) / 60))
					ELSE %[3]s
				END
			) %[8]s`,
			extractPriceFunc,
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.AuctionCurrentPrice),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.BuyoutPrice),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.DutchAuction),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.DutchAuctionDropRate),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.EndAt),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.CreatedAt),
			sortDir,
		))
	} else {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.CreatedAt), sortDir))
	}
	queryMods = append(queryMods, orderBy)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	records := []*server.MarketplaceSaleItem{}
	boil.DebugMode = true
	err = boiler.ItemSales(queryMods...).Bind(nil, gamedb.StdConn, &records)
	boil.DebugMode = false
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	return total, records, nil
}

// MarketplaceItemKeycardSaleList returns a numeric paginated result of keycard sales list.
func MarketplaceItemKeycardSaleList(
	userID string,
	factionID string,
	search string,
	filter *ListFilterRequest,
	ownedBy []string,
	minPrice decimal.NullDecimal,
	maxPrice decimal.NullDecimal,
	offset int,
	pageSize int,
	sortBy string,
	sortDir SortByDir,
	sold bool,
) (int64, []*server.MarketplaceSaleItem1155, error) {
	if !sortDir.IsValid() {
		return 0, nil, terror.Error(fmt.Errorf("invalid sort direction"))
	}

	queryMods := append(
		ItemKeycardSaleQueryMods,
		boiler.ItemKeycardSaleWhere.DeletedAt.IsNull(),
	)

	if factionID != "" {
		queryMods = append(queryMods, boiler.ItemKeycardSaleWhere.FactionID.EQ(factionID))
	}

	if sold {
		queryMods = append(queryMods, boiler.ItemKeycardSaleWhere.SoldAt.IsNotNull())
	} else {
		queryMods = append(queryMods, boiler.ItemKeycardSaleWhere.SoldTo.IsNull(), boiler.ItemKeycardSaleWhere.EndAt.GT(time.Now()))
	}

	// Filters
	if filter != nil {
		for i, f := range filter.Items {
			queryMod := GenerateListFilterQueryMod(*f, i, filter.LinkOperator)
			queryMods = append(queryMods, queryMod)
		}
	}
	if len(ownedBy) > 0 {
		isSelf := false
		isOthers := false
		for _, ownerType := range ownedBy {
			if ownerType == "self" {
				isSelf = true
			} else if ownerType == "others" {
				isOthers = true
			}
			if isSelf && isOthers {
				break
			}
		}
		if isSelf && !isOthers {
			queryMods = append(queryMods, boiler.ItemKeycardSaleWhere.OwnerID.EQ(userID))
		} else if !isSelf && isOthers {
			queryMods = append(queryMods, boiler.ItemKeycardSaleWhere.OwnerID.NEQ(userID))
		}
	}

	// Search
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			queryMods = append(queryMods, qm.And(
				fmt.Sprintf(
					`(
						(to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
					)`,
					qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.Label),
					qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.Description),
					qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Username),
				),
				xsearch,
				xsearch,
				xsearch,
			))
		}
	}

	if minPrice.Valid {
		value := minPrice.Decimal.Mul(decimal.New(1, 18))
		queryMods = append(queryMods, boiler.ItemKeycardSaleWhere.BuyoutPrice.GTE(value))
	}
	if maxPrice.Valid {
		value := maxPrice.Decimal.Mul(decimal.New(1, 18))
		queryMods = append(queryMods, boiler.ItemKeycardSaleWhere.BuyoutPrice.LTE(value))
	}

	// Get total rows
	total, err := boiler.ItemKeycardSales(queryMods...).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if total == 0 {
		return 0, []*server.MarketplaceSaleItem1155{}, nil
	}

	// Sort
	var orderBy qm.QueryMod
	if sortBy == "alphabetical" {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.Label), sortDir))
	} else if sortBy == "time" {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.EndAt), sortDir))
	} else if sortBy == "price" && sold {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.SoldFor), sortDir))
	} else if sortBy == "price" && !sold {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.BuyoutPrice), sortDir))
	} else {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.CreatedAt), sortDir))
	}
	queryMods = append(queryMods, orderBy)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	records := []*server.MarketplaceSaleItem1155{}
	err = boiler.ItemKeycardSales(queryMods...).Bind(nil, gamedb.StdConn, &records)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	return total, records, nil
}

// MarketplaceEventList lists all events involving the user.
func MarketplaceEventList(
	userID string,
	search string,
	eventType []string,
	offset int,
	pageSize int,
	sortBy string,
	sortDir SortByDir,
) (int64, []*server.MarketplaceEvent, error) {
	queryMods := []qm.QueryMod{
		// Item Sales
		qm.LeftOuterJoin(
			fmt.Sprintf(
				"%s ON %s = %s",
				boiler.TableNames.ItemSales,
				qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ID),
				qm.Rels(boiler.TableNames.MarketplaceEvents, boiler.MarketplaceEventColumns.RelatedSaleItemID),
			),
		),
		qm.LeftOuterJoin(
			fmt.Sprintf(
				"%s ON %s = %s",
				boiler.TableNames.CollectionItems,
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ID),
				qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.CollectionItemID),
			),
		),
		qm.LeftOuterJoin(
			fmt.Sprintf(
				"%s ON %s = %s AND %s = ?",
				boiler.TableNames.Mechs,
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
			),
			boiler.ItemTypeMech,
		),
		qm.LeftOuterJoin(
			fmt.Sprintf(
				"%s ON %s = %s",
				boiler.TableNames.MechSkin,
				qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.ID),
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
			),
		),
		qm.LeftOuterJoin(
			fmt.Sprintf(
				"%s ON %s = %s AND %s = ?",
				boiler.TableNames.MysteryCrate,
				qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.ID),
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
				qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
			),
			boiler.ItemTypeMysteryCrate,
		),

		// Keycards
		qm.LeftOuterJoin(
			fmt.Sprintf(
				"%s ON %s = %s",
				boiler.TableNames.ItemKeycardSales,
				qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.ID),
				qm.Rels(boiler.TableNames.MarketplaceEvents, boiler.MarketplaceEventColumns.RelatedSaleItemKeycardID),
			),
		),
		qm.LeftOuterJoin(
			fmt.Sprintf(
				"%s ON %s = %s",
				boiler.TableNames.PlayerKeycards,
				qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.ID),
				qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.ItemID),
			),
		),
		qm.LeftOuterJoin(
			fmt.Sprintf(
				"%s ON %s = %s",
				boiler.TableNames.BlueprintKeycards,
				qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.ID),
				qm.Rels(boiler.TableNames.PlayerKeycards, boiler.PlayerKeycardColumns.BlueprintKeycardID),
			),
		),

		// Item Seller owner
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s = COALESCE(%s, %s)",
				boiler.TableNames.Players,
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
				qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.OwnerID),
				qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.OwnerID),
			),
		),

		// Check if owner of any
		qm.Expr(
			boiler.MarketplaceEventWhere.UserID.EQ(userID),
			qm.Or2(qm.Expr(
				boiler.MarketplaceEventWhere.EventType.EQ(boiler.MarketplaceEventCancelled),
				boiler.MarketplaceEventWhere.RelatedSaleItemID.IsNotNull(),
				qm.And(fmt.Sprintf(
					`EXISTS (
						SELECT 1
						FROM %s _b
						WHERE _b.item_sale_id = %s
							AND _b.bidder_id = ?
						LIMIT 1
					)`,
					boiler.TableNames.ItemSalesBidHistory,
					qm.Rels(boiler.TableNames.MarketplaceEvents, boiler.MarketplaceEventColumns.RelatedSaleItemID),
				), userID),
			)),
		),
	}

	// Filters
	if len(eventType) > 0 {
		queryMods = append(queryMods, boiler.MarketplaceEventWhere.EventType.IN(eventType))
	}

	// Search
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			queryMods = append(queryMods, qm.And(
				fmt.Sprintf(
					`(
						(to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
					)`,
					qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
					qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.Label),
					qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Tier),
					qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Username),
				),
				xsearch,
				xsearch,
				xsearch,
				xsearch,
				xsearch,
			))
		}
	}

	// Get total rows
	total, err := boiler.MarketplaceEvents(queryMods...).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if total == 0 {
		return 0, []*server.MarketplaceEvent{}, nil
	}

	// Sort
	if !sortDir.IsValid() {
		sortDir = SortByDirDesc
	}
	if sortBy == "" {
		sortBy = "date"
	}

	if sortBy == boiler.MarketplaceEventColumns.CreatedAt {
		sortBy = qm.Rels(boiler.TableNames.MarketplaceEvents, boiler.MarketplaceEventColumns.CreatedAt)
	} else if sortBy == "alphabetical" {
		sortBy = fmt.Sprintf(
			`CASE
				WHEN %[1]s = '%[2]s' THEN %[4]s
				WHEN %[1]s = '%[5]s' THEN %[6]s
				ELSE %[7]s
			END`,
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
			boiler.ItemTypeMech,
			qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
			boiler.ItemTypeMysteryCrate,
			qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.Label),
			qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.Label),
		)
	} else {
		return 0, nil, terror.Error(fmt.Errorf("invalid sort column name"))
	}

	queryMods = append(
		queryMods,
		qm.OrderBy(sortBy+" "+string(sortDir)),
	)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	queryMods = append(queryMods,
		qm.Load(boiler.MarketplaceEventRels.RelatedSaleItem, qm.WithDeleted()),
		qm.Load(qm.Rels(
			boiler.MarketplaceEventRels.RelatedSaleItem,
			boiler.ItemSaleRels.CollectionItem,
		), qm.WithDeleted()),
		qm.Load(qm.Rels(
			boiler.MarketplaceEventRels.RelatedSaleItem,
			boiler.ItemSaleRels.SoldToPlayer,
		), qm.WithDeleted()),
		qm.Load(boiler.MarketplaceEventRels.RelatedSaleItemKeycard, qm.WithDeleted()),
		qm.Load(qm.Rels(
			boiler.MarketplaceEventRels.RelatedSaleItemKeycard,
			boiler.ItemKeycardSaleRels.Item,
			boiler.PlayerKeycardRels.BlueprintKeycard,
		), qm.WithDeleted()),
		qm.Load(qm.Rels(
			boiler.MarketplaceEventRels.RelatedSaleItemKeycard,
			boiler.ItemKeycardSaleRels.SoldToPlayer,
		), qm.WithDeleted()),
	)
	records, err := boiler.MarketplaceEvents(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	// Populate results
	collectionToMechID := map[string]string{}
	collectionToMysteryCrateID := map[string]string{}
	collectionToWeaponID := map[string]string{}
	mechIDs := []string{}
	mysteryCrateIDs := []string{}
	weaponIDs := []string{}

	output := []*server.MarketplaceEvent{}
	for _, r := range records {
		row := &server.MarketplaceEvent{
			ID:        r.ID,
			EventType: r.EventType,
			Amount:    r.Amount,
			CreatedAt: r.CreatedAt,
		}

		if r.R != nil {
			if r.R.RelatedSaleItem != nil {
				row.Item = &server.MarketplaceEventItem{
					ID:                   r.R.RelatedSaleItem.ID,
					FactionID:            r.R.RelatedSaleItem.FactionID,
					CollectionItemID:     r.R.RelatedSaleItem.CollectionItemID,
					CollectionItemType:   r.R.RelatedSaleItem.R.CollectionItem.ItemType,
					ListingFeeTXID:       r.R.RelatedSaleItem.ListingFeeTXID,
					OwnerID:              r.R.RelatedSaleItem.OwnerID,
					Auction:              r.R.RelatedSaleItem.Auction,
					AuctionCurrentPrice:  r.R.RelatedSaleItem.AuctionCurrentPrice,
					AuctionReservedPrice: r.R.RelatedSaleItem.AuctionReservedPrice,
					Buyout:               r.R.RelatedSaleItem.Buyout,
					BuyoutPrice:          r.R.RelatedSaleItem.BuyoutPrice,
					DutchAuction:         r.R.RelatedSaleItem.DutchAuction,
					DutchAuctionDropRate: r.R.RelatedSaleItem.DutchAuctionDropRate,
					EndAt:                r.R.RelatedSaleItem.EndAt,
					SoldAt:               r.R.RelatedSaleItem.SoldAt,
					SoldFor:              r.R.RelatedSaleItem.SoldFor,
					SoldTXID:             r.R.RelatedSaleItem.SoldTXID,
					SoldFeeTXID:          r.R.RelatedSaleItem.SoldFeeTXID,
					DeletedAt:            r.R.RelatedSaleItem.DeletedAt,
					UpdatedAt:            r.R.RelatedSaleItem.UpdatedAt,
					CreatedAt:            r.R.RelatedSaleItem.CreatedAt,
					CollectionItem: server.MarketplaceSaleCollectionItem{
						Hash:         r.R.RelatedSaleItem.R.CollectionItem.Hash,
						Tier:         null.StringFrom(r.R.RelatedSaleItem.R.CollectionItem.Tier),
						XsynLocked:   r.R.RelatedSaleItem.R.CollectionItem.XsynLocked,
						MarketLocked: r.R.RelatedSaleItem.R.CollectionItem.MarketLocked,
					},
				}
				if r.R.RelatedSaleItem.R.SoldToPlayer != nil {
					row.Item.SoldTo = server.MarketplaceUser{
						ID:            null.StringFrom(r.R.RelatedSaleItem.R.SoldToPlayer.ID),
						Username:      r.R.RelatedSaleItem.R.SoldToPlayer.Username,
						FactionID:     r.R.RelatedSaleItem.R.SoldToPlayer.FactionID,
						PublicAddress: r.R.RelatedSaleItem.R.SoldToPlayer.PublicAddress,
						Gid:           null.IntFrom(r.R.RelatedSaleItem.R.SoldToPlayer.Gid),
					}
				}
				switch r.R.RelatedSaleItem.R.CollectionItem.ItemType {
				case boiler.ItemTypeMech:
					mechIDs = append(mechIDs, r.R.RelatedSaleItem.R.CollectionItem.ItemID)
					collectionToMechID[row.Item.CollectionItemID] = r.R.RelatedSaleItem.R.CollectionItem.ItemID
				case boiler.ItemTypeMysteryCrate:
					mysteryCrateIDs = append(mysteryCrateIDs, r.R.RelatedSaleItem.R.CollectionItem.ItemID)
					collectionToMysteryCrateID[row.Item.CollectionItemID] = r.R.RelatedSaleItem.R.CollectionItem.ItemID
				case boiler.ItemTypeWeapon:
					weaponIDs = append(weaponIDs, r.R.RelatedSaleItem.R.CollectionItem.ItemID)
					collectionToWeaponID[row.Item.CollectionItemID] = r.R.RelatedSaleItem.R.CollectionItem.ItemID
				}
			} else if r.R.RelatedSaleItemKeycard != nil {
				row.Item = &server.MarketplaceEventItem{
					ID:                   r.R.RelatedSaleItemKeycard.ID,
					FactionID:            r.R.RelatedSaleItemKeycard.FactionID,
					CollectionItemID:     r.R.RelatedSaleItemKeycard.ItemID,
					CollectionItemType:   "keycard",
					ListingFeeTXID:       r.R.RelatedSaleItemKeycard.ListingFeeTXID,
					OwnerID:              r.R.RelatedSaleItemKeycard.OwnerID,
					Auction:              false,
					AuctionCurrentPrice:  decimal.NullDecimal{},
					AuctionReservedPrice: decimal.NullDecimal{},
					Buyout:               true,
					BuyoutPrice:          decimal.NewNullDecimal(r.R.RelatedSaleItemKeycard.BuyoutPrice),
					DutchAuction:         false,
					DutchAuctionDropRate: decimal.NullDecimal{},
					EndAt:                r.R.RelatedSaleItemKeycard.EndAt,
					SoldAt:               r.R.RelatedSaleItemKeycard.SoldAt,
					SoldFor:              r.R.RelatedSaleItemKeycard.SoldFor,
					SoldTXID:             r.R.RelatedSaleItemKeycard.SoldTXID,
					SoldFeeTXID:          r.R.RelatedSaleItemKeycard.SoldFeeTXID,
					DeletedAt:            r.R.RelatedSaleItemKeycard.DeletedAt,
					UpdatedAt:            r.R.RelatedSaleItemKeycard.UpdatedAt,
					CreatedAt:            r.R.RelatedSaleItemKeycard.CreatedAt,
					Keycard: server.AssetKeycardBlueprint{
						ID:             r.R.RelatedSaleItemKeycard.R.Item.R.BlueprintKeycard.ID,
						Label:          r.R.RelatedSaleItemKeycard.R.Item.R.BlueprintKeycard.Label,
						Description:    r.R.RelatedSaleItemKeycard.R.Item.R.BlueprintKeycard.Description,
						Collection:     r.R.RelatedSaleItemKeycard.R.Item.R.BlueprintKeycard.Collection,
						KeycardTokenID: r.R.RelatedSaleItemKeycard.R.Item.R.BlueprintKeycard.KeycardTokenID,
						ImageURL:       r.R.RelatedSaleItemKeycard.R.Item.R.BlueprintKeycard.ImageURL,
						AnimationURL:   r.R.RelatedSaleItemKeycard.R.Item.R.BlueprintKeycard.AnimationURL,
						KeycardGroup:   r.R.RelatedSaleItemKeycard.R.Item.R.BlueprintKeycard.KeycardGroup,
						Syndicate:      r.R.RelatedSaleItemKeycard.R.Item.R.BlueprintKeycard.Syndicate,
						CreatedAt:      r.R.RelatedSaleItemKeycard.R.Item.R.BlueprintKeycard.CreatedAt,
					},
				}
				if r.R.RelatedSaleItemKeycard.R.SoldToPlayer != nil {
					row.Item.SoldTo = server.MarketplaceUser{
						ID:            null.StringFrom(r.R.RelatedSaleItemKeycard.R.SoldToPlayer.ID),
						Username:      r.R.RelatedSaleItemKeycard.R.SoldToPlayer.Username,
						FactionID:     r.R.RelatedSaleItemKeycard.R.SoldToPlayer.FactionID,
						PublicAddress: r.R.RelatedSaleItemKeycard.R.SoldToPlayer.PublicAddress,
						Gid:           null.IntFrom(r.R.RelatedSaleItemKeycard.R.SoldToPlayer.Gid),
					}
				}
			}
		}
		output = append(output, row)
	}

	// Load in collection item details
	if len(mechIDs) > 0 {
		mechs := []*server.MarketplaceSaleItemMech{}
		err = boiler.Mechs(
			qm.Select(
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
				qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
				qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.Label),
			),
			qm.InnerJoin(
				fmt.Sprintf(
					"%s ON %s = %s",
					boiler.TableNames.MechSkin,
					qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.ID),
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
				),
			),
			qm.InnerJoin(
				fmt.Sprintf(
					"%s ON %s = %s",
					boiler.TableNames.BlueprintMechs,
					qm.Rels(boiler.TableNames.BlueprintMechs, boiler.BlueprintMechColumns.ID),
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.BlueprintID),
				),
			),
			boiler.MechWhere.ID.IN(mechIDs),
		).Bind(nil, gamedb.StdConn, &mechs)
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		for i := range output {
			itemID, ok := collectionToMechID[output[i].Item.CollectionItemID]
			if output[i].Item.CollectionItemType != boiler.ItemTypeMech || !ok {
				continue
			}
			for _, m := range mechs {
				if m.ID.String == itemID {
					output[i].Item.Mech = *m
					break
				}
			}
		}
	}
	if len(mysteryCrateIDs) > 0 {
		mysteryCrates := []*server.MarketplaceSaleItemMysteryCrate{}

		err = boiler.MysteryCrates(
			qm.Select(
				qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.ID),
				qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.Label),
				qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.Description),
				qm.Rels(boiler.TableNames.StorefrontMysteryCrates, boiler.StorefrontMysteryCrateColumns.ImageURL),
				qm.Rels(boiler.TableNames.StorefrontMysteryCrates, boiler.StorefrontMysteryCrateColumns.CardAnimationURL),
				qm.Rels(boiler.TableNames.StorefrontMysteryCrates, boiler.StorefrontMysteryCrateColumns.LargeImageURL),
				qm.Rels(boiler.TableNames.StorefrontMysteryCrates, boiler.StorefrontMysteryCrateColumns.AnimationURL),
			),
			qm.InnerJoin(fmt.Sprintf("%s ON %s = %s",
				boiler.TableNames.StorefrontMysteryCrates,
				qm.Rels(boiler.TableNames.StorefrontMysteryCrates, boiler.StorefrontMysteryCrateColumns.ID),
				qm.Rels(boiler.TableNames.MysteryCrate, boiler.MysteryCrateColumns.BlueprintID),
			)),
			boiler.MysteryCrateWhere.ID.IN(mysteryCrateIDs),
		).Bind(nil, gamedb.StdConn, &mysteryCrates)
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		for i := range output {
			itemID, ok := collectionToMysteryCrateID[output[i].Item.CollectionItemID]
			if output[i].Item.CollectionItemType != boiler.ItemTypeMysteryCrate || !ok {
				continue
			}
			for _, m := range mysteryCrates {
				if m.ID.String == itemID {
					output[i].Item.MysteryCrate = *m
					break
				}
			}
		}
	}
	if len(weaponIDs) > 0 {
		weapons := []*server.MarketplaceSaleItemWeapon{}
		err = boiler.Weapons(
			qm.Select(
				qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.ID),
				qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.Label),
				qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.WeaponType),
				fmt.Sprintf(`COALESCE(%s, %s)`,
					qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.AvatarURL),
					qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.ImageURL),
				),
			),
			qm.InnerJoin(
				fmt.Sprintf(
					"%s ON %s = %s",
					boiler.TableNames.WeaponSkin,
					qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.ID),
					qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.EquippedWeaponSkinID),
				),
			),
			qm.InnerJoin(
				fmt.Sprintf(
					"%s ON %s = %s",
					boiler.TableNames.BlueprintWeapons,
					qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID),
					qm.Rels(boiler.TableNames.Weapons, boiler.WeaponColumns.BlueprintID),
				),
			),
			qm.InnerJoin(
				fmt.Sprintf(
					"%s ON %s = %s",
					boiler.TableNames.BlueprintWeaponSkin,
					qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
					qm.Rels(boiler.TableNames.WeaponSkin, boiler.WeaponSkinColumns.BlueprintID),
				),
			),
			qm.InnerJoin(
				fmt.Sprintf(
					"%s ON %s = %s AND %s = %s",
					boiler.TableNames.WeaponModelSkinCompatibilities,
					qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.BlueprintWeaponSkinID),
					qm.Rels(boiler.TableNames.BlueprintWeaponSkin, boiler.BlueprintWeaponSkinColumns.ID),
					qm.Rels(boiler.TableNames.WeaponModelSkinCompatibilities, boiler.WeaponModelSkinCompatibilityColumns.WeaponModelID),
					qm.Rels(boiler.TableNames.BlueprintWeapons, boiler.BlueprintWeaponColumns.ID),
				),
			),
			boiler.WeaponWhere.ID.IN(weaponIDs),
		).Bind(context.Background(), gamedb.StdConn, &weapons)
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		for i := range output {
			itemID, ok := collectionToWeaponID[output[i].Item.CollectionItemID]
			if output[i].Item.CollectionItemType != boiler.ItemTypeWeapon || !ok {
				continue
			}
			for _, w := range weapons {
				if w.ID.String == itemID {
					output[i].Item.Weapon = *w
					break
				}
			}
		}
	}

	return total, output, nil
}

// MarketplaceSaleArchive archives as sale item.
func MarketplaceSaleArchive(conn boil.Executor, id uuid.UUID) error {
	obj := &boiler.ItemSale{
		ID:        id.String(),
		DeletedAt: null.TimeFrom(time.Now()),
	}
	_, err := obj.Update(conn, boil.Whitelist(boiler.ItemSaleColumns.DeletedAt))
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceSaleArchiveByItemID archives as sale item.
func MarketplaceSaleArchiveByItemID(conn boil.Executor, id uuid.UUID) error {
	asset, err := boiler.ItemSales(
		boiler.ItemSaleWhere.CollectionItemID.EQ(id.String()),
		boiler.ItemSaleWhere.EndAt.GT(time.Now()),
		boiler.ItemSaleWhere.DeletedAt.IsNull(),
	).One(conn)
	if err != nil {
		return terror.Error(err)
	}
	_, err = asset.Delete(conn, false)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// MarketplaceSaleItemUnlock removes the locked to marketplace on an archived item.
func MarketplaceSaleItemUnlock(conn boil.Executor, id uuid.UUID) error {
	q := `
		UPDATE collection_items
		SET locked_to_marketplace = false
		WHERE locked_to_marketplace = true AND id IN (
			SELECT _s.collection_item_id
			FROM item_sales _s
			WHERE _s.id = $1
				AND _s.deleted_at IS NOT NULL
		)`
	_, err := conn.Exec(q, id)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceKeycardSaleArchive archives as sale item.
func MarketplaceKeycardSaleArchive(conn boil.Executor, id uuid.UUID) error {
	obj := &boiler.ItemKeycardSale{
		ID:        id.String(),
		DeletedAt: null.TimeFrom(time.Now()),
	}
	_, err := obj.Update(conn, boil.Whitelist(boiler.ItemKeycardSaleColumns.DeletedAt))
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceSaleCreate inserts a new sale item.
func MarketplaceSaleCreate(
	conn boil.Executor,
	ownerID uuid.UUID,
	factionID uuid.UUID,
	listFeeTxnID null.String,
	endAt time.Time,
	collectionItemID uuid.UUID,
	hasBuyout bool,
	askingPrice decimal.NullDecimal,
	hasAuction bool,
	auctionReservedPrice decimal.NullDecimal,
	auctionCurrentPrice decimal.NullDecimal,
	hasDutchAuction bool,
	dutchAuctionDropRate decimal.NullDecimal,
) (*server.MarketplaceSaleItem, error) {
	obj := &boiler.ItemSale{
		OwnerID:          ownerID.String(),
		FactionID:        factionID.String(),
		ListingFeeTXID:   listFeeTxnID,
		CollectionItemID: collectionItemID.String(),
		EndAt:            endAt,
	}

	if hasBuyout {
		obj.Buyout = true
		obj.BuyoutPrice = decimal.NewNullDecimal(askingPrice.Decimal.Mul(decimal.New(1, 18)))
	}
	if hasAuction {
		obj.Auction = true
		if auctionCurrentPrice.Valid {
			obj.AuctionCurrentPrice = decimal.NewNullDecimal(auctionCurrentPrice.Decimal.Mul(decimal.New(1, 18)))
		} else {
			obj.AuctionCurrentPrice = decimal.NewNullDecimal(decimal.New(1, 18))
		}
		if auctionReservedPrice.Valid {
			obj.AuctionReservedPrice = decimal.NewNullDecimal(auctionReservedPrice.Decimal.Mul(decimal.New(1, 18)))
		}
	}
	if hasDutchAuction {
		obj.DutchAuction = true
		obj.BuyoutPrice = decimal.NewNullDecimal(askingPrice.Decimal.Mul(decimal.New(1, 18)))
		obj.DutchAuctionDropRate = decimal.NewNullDecimal(dutchAuctionDropRate.Decimal.Mul(decimal.New(1, 18)))
		if auctionReservedPrice.Valid {
			obj.AuctionReservedPrice = decimal.NewNullDecimal(auctionReservedPrice.Decimal.Mul(decimal.New(1, 18)))
		}
	}

	err := obj.Insert(conn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	output := &server.MarketplaceSaleItem{
		ID:                   obj.ID,
		FactionID:            obj.FactionID,
		CollectionItemID:     obj.CollectionItemID,
		ListingFeeTXID:       obj.ListingFeeTXID,
		OwnerID:              obj.OwnerID,
		Auction:              obj.Auction,
		AuctionCurrentPrice:  obj.AuctionCurrentPrice,
		AuctionReservedPrice: obj.AuctionReservedPrice,
		Buyout:               obj.Buyout,
		BuyoutPrice:          obj.BuyoutPrice,
		DutchAuction:         obj.DutchAuction,
		DutchAuctionDropRate: obj.DutchAuctionDropRate,
		EndAt:                obj.EndAt,
		SoldAt:               obj.SoldAt,
		SoldTXID:             obj.SoldTXID,
		DeletedAt:            obj.DeletedAt,
		UpdatedAt:            obj.UpdatedAt,
		CreatedAt:            obj.CreatedAt,
	}
	return output, nil
}

// CancelBidResponse contains the txid and amount on cancelled bids.
type CancelBidResponse struct {
	BidderID  string
	FactionID null.String
	TXID      string
	Amount    decimal.Decimal
}

// MarketplaceSaleCancelBids cancels all active bids and returns transaction ids needed to be retuned (ideally one).
func MarketplaceSaleCancelBids(conn boil.Executor, itemID uuid.UUID, msg string) ([]CancelBidResponse, error) {
	q := `
		UPDATE item_sales_bid_history b
		SET cancelled_at = NOW(),
			cancelled_reason = $2
		FROM players p
		WHERE b.item_sale_id = $1
			AND b.cancelled_at IS NULL
			AND p.id = b.bidder_id
		RETURNING bidder_id, faction_id, bid_tx_id, bid_price`
	rows, err := conn.Query(q, itemID, msg)
	if err != nil {
		return nil, terror.Error(err)
	}
	defer rows.Close()

	cancelBidList := []CancelBidResponse{}
	for rows.Next() {
		var outputItem CancelBidResponse
		err := rows.Scan(&outputItem.BidderID, &outputItem.FactionID, &outputItem.TXID, &outputItem.Amount)
		if err != nil {
			return nil, terror.Error(err)
		}
		cancelBidList = append(cancelBidList, outputItem)
	}
	return cancelBidList, nil
}

// MarketplaceSaleBidHistoryRefund adds in refund details to a specific bid.
func MarketplaceSaleBidHistoryRefund(conn boil.Executor, itemID uuid.UUID, txID, refundTxID string, cancelledAuction bool) error {
	cancelledAuctionSet := ""
	if cancelledAuction {
		cancelledAuctionSet = ", cancelled_at = NOW(), cancelled_reason = 'Auction Cancelled'"

	}
	q := fmt.Sprintf(`
		UPDATE item_sales_bid_history
		SET refund_bid_tx_id = $3
			%s
		WHERE item_sale_id = $1
			AND bid_tx_id = $2`,
		cancelledAuctionSet,
	)
	_, err := conn.Exec(q, itemID, txID, refundTxID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceSaleBidHistoryCreate inserts a new bid history record.
func MarketplaceSaleBidHistoryCreate(conn boil.Executor, id uuid.UUID, bidderUserID uuid.UUID, amount decimal.Decimal, txid string) (*boiler.ItemSalesBidHistory, error) {
	obj := &boiler.ItemSalesBidHistory{
		ItemSaleID: id.String(),
		BidderID:   bidderUserID.String(),
		BidTXID:    txid,
		BidPrice:   amount.Mul(decimal.New(1, 18)),
	}
	err := obj.Insert(conn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	return obj, nil
}

// MarketplaceSaleAuctionSync updates the current auction price based on the bid history.
func MarketplaceSaleAuctionSync(conn boil.Executor, id uuid.UUID) error {
	q := fmt.Sprintf(
		`UPDATE %s
		SET %s = (
			SELECT %s
			FROM %s
			WHERE %s = $1
				AND %s IS NULL 
			LIMIT 1
		)
		WHERE %s = $1`,
		boiler.TableNames.ItemSales,
		boiler.ItemSaleColumns.AuctionCurrentPrice,
		boiler.ItemSalesBidHistoryColumns.BidPrice,
		boiler.TableNames.ItemSalesBidHistory,
		boiler.ItemSalesBidHistoryColumns.ItemSaleID,
		boiler.ItemSalesBidHistoryColumns.CancelledAt,
		boiler.ItemSaleColumns.ID,
	)
	_, err := conn.Exec(q, id)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceCheckCollectionItem checks whether collection item is already in marketplace.
func MarketplaceCheckCollectionItem(collectionItemID uuid.UUID) (bool, error) {
	output, err := boiler.ItemSales(
		boiler.ItemSaleWhere.CollectionItemID.EQ(collectionItemID.String()),
		boiler.ItemSaleWhere.SoldAt.IsNull(),
		boiler.ItemSaleWhere.EndAt.GT(time.Now()),
	).Exists(gamedb.StdConn)
	if err != nil {
		return false, terror.Error(err)
	}
	return output, nil
}

// MarketplaceKeycardSaleCreate inserts a new sale item.
func MarketplaceKeycardSaleCreate(
	conn boil.Executor,
	ownerID uuid.UUID,
	factionID uuid.UUID,
	listFeeTxnID null.String,
	endAt time.Time,
	itemID uuid.UUID,
	askingPrice decimal.Decimal,
) (*server.MarketplaceSaleItem1155, error) {
	obj := &boiler.ItemKeycardSale{
		OwnerID:        ownerID.String(),
		FactionID:      factionID.String(),
		ListingFeeTXID: listFeeTxnID,
		ItemID:         itemID.String(),
		EndAt:          endAt,
		BuyoutPrice:    askingPrice.Mul(decimal.New(1, 18)),
	}
	err := obj.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	output := &server.MarketplaceSaleItem1155{
		ID:             obj.ID,
		FactionID:      obj.FactionID,
		ItemID:         obj.ItemID,
		ListingFeeTXID: obj.ListingFeeTXID,
		OwnerID:        obj.OwnerID,
		BuyoutPrice:    obj.BuyoutPrice,
		EndAt:          obj.EndAt,
		SoldAt:         obj.SoldAt,
		SoldFor:        obj.SoldFor,
		SoldTXID:       obj.SoldTXID,
		DeletedAt:      obj.DeletedAt,
		UpdatedAt:      obj.UpdatedAt,
		CreatedAt:      obj.CreatedAt,
	}
	return output, nil
}

// ChangeKeycardOwner changes a keycard from previous owner to new owner.
func ChangeKeycardOwner(conn boil.Executor, itemSaleID uuid.UUID) error {
	q := `
		INSERT INTO player_keycards (player_id, blueprint_keycard_id, count)

		SELECT iks.sold_to AS player_id, pk.blueprint_keycard_id, 1 AS count
		FROM item_keycard_sales iks
			INNER JOIN player_keycards pk ON pk.id = iks.item_id
		WHERE iks.id = $1 AND iks.sold_to IS NOT NULL
		ON CONFLICT (player_id, blueprint_keycard_id)
		DO UPDATE 
		SET count = player_keycards.count + 1`
	_, err := conn.Exec(q, itemSaleID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// DecrementPlayerKeycard deducts keycard count.
func DecrementPlayerKeycardCount(conn boil.Executor, playerKeycardID uuid.UUID) error {
	q := `
		UPDATE player_keycards 
		SET count = count - 1
		WHERE id = $1`
	_, err := conn.Exec(q, playerKeycardID.String())
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// IncrementPlayerKeycard deducts keycard count.
func IncrementPlayerKeycardCount(conn boil.Executor, playerKeycardID uuid.UUID) error {
	q := `
		UPDATE player_keycards 
		SET count = count + 1
		WHERE id = $1`
	_, err := conn.Exec(q, playerKeycardID.String())
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceAddEvent adds an event to marketplace logs.
func MarketplaceAddEvent(eventType string, userID string, amount decimal.NullDecimal, itemSaleID string, table string) error {
	obj := &boiler.MarketplaceEvent{
		UserID:    userID,
		EventType: eventType,
		Amount:    amount,
	}
	if table == boiler.TableNames.ItemKeycardSales {
		obj.RelatedSaleItemKeycardID = null.StringFrom(itemSaleID)
	} else if table == boiler.TableNames.ItemSales {
		obj.RelatedSaleItemID = null.StringFrom(itemSaleID)
	} else {
		return terror.Error(fmt.Errorf("invalid item sale table"))
	}
	err := obj.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceGetOtherAssets grabs a list of other assets found in item sale.
func MarketplaceGetOtherAssets(conn boil.Executor, itemSaleID string) ([]string, error) {
	var output types.StringArray
	err := conn.QueryRow(ItemSaleOtherAssetsSQL, itemSaleID).Scan(&output)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err)
	}
	return output, nil
}

// MarketplaceItemIsGenesisOrLimitedMech checks whether sale item is a genesis mech for sale.
func MarketplaceItemIsGenesisOrLimitedMech(conn boil.Executor, itemSaleID string) (bool, error) {
	mechRow, err := boiler.Mechs(
		qm.Select(qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID)),
		qm.InnerJoin(fmt.Sprintf(
			"%s ON %s = %s AND %s = ?",
			boiler.TableNames.CollectionItems,
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
		), boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMech),
		boiler.CollectionItemWhere.ItemID.EQ(itemSaleID),
	).One(conn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, terror.Error(err)
	}

	mech, err := Mech(conn, mechRow.ID)
	if err != nil {
		return false, terror.Error(err)
	}

	return mech.IsCompleteGenesis() || mech.IsCompleteLimited(), nil
}
