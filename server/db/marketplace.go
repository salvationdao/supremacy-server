package db

import (
	"database/sql"
	"errors"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var ItemSaleQueryMods = []qm.QueryMod{
	qm.Select(
		`item_sales.id AS id,
		item_sales.faction_id AS faction_id,
		item_sales.collection_item_id AS collection_item_id,
		item_sales.listing_fee_tx_id AS listing_fee_tx_id,
		item_sales.owner_id AS owner_id,
		item_sales.auction AS auction,
		item_sales.auction_current_price AS auction_current_price,
		item_sales.auction_reserved_price AS auction_reserved_price,
		item_sales.buyout AS buyout,
		item_sales.buyout_price AS buyout_price,
		item_sales.dutch_auction AS dutch_auction,
		item_sales.dutch_auction_drop_rate AS dutch_auction_drop_rate,
		item_sales.end_at AS end_at,
		item_sales.sold_at AS sold_at,
		item_sales.sold_for AS sold_for,
		st.id AS "sold_to.id",
		st.username AS "sold_to.username",
		st.faction_id AS "sold_to.faction_id",
		st.public_address AS "sold_to.public_address",
		st.gid AS "sold_to.gid",
		item_sales.sold_tx_id AS sold_tx_id,
		item_sales.sold_fee_tx_id AS sold_fee_tx_id,
		item_sales.deleted_at AS deleted_at,
		item_sales.updated_at AS updated_at,
		item_sales.created_at AS created_at,
		collection_items.item_type AS collection_item_type,
		(SELECT COUNT(*) FROM item_sales_bid_history _isbh WHERE item_sale_id = item_sales.id) AS total_bids,
		players.id AS "players.id",
		players.username AS "players.username",
		players.faction_id AS "players.faction_id",
		players.public_address AS "players.public_address",
		players.gid AS "players.gid",
		mechs.id AS "mechs.id",
		mechs.name AS "mechs.name",
		mechs.label AS "mechs.label",
		mech_skin.avatar_url AS "mech_skin.avatar_url",
		mystery_crate.id AS "mystery_crate.id",
		mystery_crate.label AS "mystery_crate.label",
		mystery_crate.description AS "mystery_crate.description",
		collection_items.tier AS "collection_items.tier",
		collection_items.image_url AS "collection_items.image_url",
		collection_items.card_animation_url AS "collection_items.card_animation_url",
		collection_items.avatar_url AS "collection_items.avatar_url",
		collection_items.large_image_url AS "collection_items.large_image_url",
		collection_items.background_color AS "collection_items.background_color",
		collection_items.youtube_url AS "collection_items.youtube_url",
		collection_items.xsyn_locked AS "collection_items.xsyn_locked",
		collection_items.market_locked AS "collection_items.market_locked",
		collection_items.hash AS "collection_items.hash",
		bidder.id AS "bidder.id",
		bidder.username AS "bidder.username",
		bidder.faction_id AS "bidder.faction_id",
		bidder.public_address AS "bidder.public_address",
		bidder.gid AS "bidder.gid"
		`,
	),
	qm.InnerJoin(
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
	qm.InnerJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.Players,
			qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.OwnerID),
		),
	),
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s AS st ON %s = %s",
			boiler.TableNames.Players,
			qm.Rels("st", boiler.PlayerColumns.ID),
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
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s AS bidder ON %s = %s",
			boiler.TableNames.Players,
			qm.Rels("bidder", boiler.PlayerColumns.ID),
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
		&output.Mech.AvatarURL,
		&output.MysteryCrate.ID,
		&output.MysteryCrate.Label,
		&output.MysteryCrate.Description,
		&output.CollectionItem.Tier,
		&output.CollectionItem.ImageURL,
		&output.CollectionItem.CardAnimationURL,
		&output.CollectionItem.AvatarURL,
		&output.CollectionItem.LargeImageURL,
		&output.CollectionItem.BackgroundColor,
		&output.CollectionItem.YoutubeURL,
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

// MarketplaceItemSaleList returns a numeric paginated result of sales list.
func MarketplaceItemSaleList(
	itemType string,
	userID string,
	factionID string,
	search string,
	rarities []string,
	saleTypes []string,
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
		queryMods = append(queryMods, boiler.CollectionItemWhere.Tier.IN(rarities))
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
						OR (to_tsvector('english', %s) @@ to_tsquery(?))
					)`,
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Label),
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
					qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.Tier),
					qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Username),
				),
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
		orderBy = qm.OrderBy(fmt.Sprintf("COALESCE(mechs.label, mechs.name) %s", sortDir))
	} else if sortBy == "time" {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.EndAt), sortDir))
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
	err = boiler.ItemSales(queryMods...).Bind(nil, gamedb.StdConn, &records)
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
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemSaleColumns.EndAt), sortDir))
	} else if sortBy == "price" && sold {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemSaleColumns.SoldFor), sortDir))
	} else if sortBy == "price" && !sold {
		orderBy = qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemSaleColumns.BuyoutPrice), sortDir))
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
	eventType string,
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
		qm.Where(
			fmt.Sprintf(
				`(
					%s = ?
					OR %s = ?
					OR EXISTS (
						SELECT 1
						FROM item_sales_bid_history _b
						WHERE _b.item_sale_id = marketplace_events.related_sale_item_id
							AND _b.bidder_id = ?
						LIMIT 1
					)
				)`,
				qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.OwnerID),
				qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.OwnerID),
			),
			userID,
			userID,
			userID,
		),
	}

	// Filters
	if eventType != "" {
		queryMods = append(queryMods, boiler.MarketplaceEventWhere.EventType.EQ(eventType))
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
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Label),
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
				WHEN %[1]s = '%[2]s' THEN COALESCE(%[3]s, %[4]s)
				WHEN %[1]s = '%[5]s' THEN %[6]s
				ELSE %[7]s
			END`,
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemType),
			boiler.ItemTypeMech,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Label),
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
		qm.Load(qm.Rels(
			boiler.MarketplaceEventRels.RelatedSaleItem,
			boiler.ItemSaleRels.CollectionItem,
		)),
		qm.Load(qm.Rels(
			boiler.MarketplaceEventRels.RelatedSaleItem,
			boiler.ItemSaleRels.SoldToPlayer,
		)),
		qm.Load(qm.Rels(
			boiler.MarketplaceEventRels.RelatedSaleItemKeycard,
			boiler.ItemKeycardSaleRels.Item,
			boiler.PlayerKeycardRels.BlueprintKeycard,
		)),
		qm.Load(qm.Rels(
			boiler.MarketplaceEventRels.RelatedSaleItemKeycard,
			boiler.ItemKeycardSaleRels.SoldToPlayer,
		)),
	)
	records, err := boiler.MarketplaceEvents(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	// Populate reasons
	mechIDs := []string{}
	mysteryCrateIDs := []string{}

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
					mechIDs = append(mechIDs, r.R.RelatedSaleItem.CollectionItemID)
				case boiler.ItemTypeMysteryCrate:
					mysteryCrateIDs = append(mysteryCrateIDs, r.R.RelatedSaleItem.CollectionItemID)
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

		// Load in collection item details
		if len(mechIDs) > 0 {
			mechs := []*server.MarketplaceSaleItemMech{}
			err = boiler.Mechs(
				qm.Select(
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Name),
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Label),
					qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.AvatarURL),
				),
				qm.InnerJoin(
					fmt.Sprintf(
						"%s ON %s = %s",
						boiler.TableNames.MechSkin,
						qm.Rels(boiler.TableNames.MechSkin, boiler.MechSkinColumns.ID),
						qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ChassisSkinID),
					),
				),
				boiler.MechWhere.ID.IN(mechIDs),
			).Bind(nil, gamedb.StdConn, &mechs)
			if err != nil {
				return 0, nil, terror.Error(err)
			}
			for i := range output {
				if output[i].Item.CollectionItemType != boiler.ItemTypeMech {
					continue
				}
				for _, m := range mechs {
					if m.ID.String == output[i].Item.CollectionItemID {
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
				),
				boiler.MysteryCrateWhere.ID.IN(mysteryCrateIDs),
			).Bind(nil, gamedb.StdConn, &mysteryCrates)
			if err != nil {
				return 0, nil, terror.Error(err)
			}
			for i := range output {
				if output[i].Item.CollectionItemType != boiler.ItemTypeMysteryCrate {
					continue
				}
				for _, m := range mysteryCrates {
					if m.ID.String == output[i].Item.CollectionItemID {
						output[i].Item.MysteryCrate = *m
						break
					}
				}
			}
		}
		output = append(output, row)
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
	listFeeTxnID string,
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
	TXID   string
	Amount decimal.Decimal
}

// MarketplaceSaleCancelBids cancels all active bids and returns transaction ids needed to be retuned (ideally one).
func MarketplaceSaleCancelBids(conn boil.Executor, itemID uuid.UUID, msg string) ([]CancelBidResponse, error) {
	q := `
		UPDATE item_sales_bid_history
		SET cancelled_at = NOW(),
			cancelled_reason = $2
		WHERE item_sale_id = $1 AND cancelled_at IS NULL
		RETURNING bid_tx_id, bid_price`
	rows, err := conn.Query(q, itemID, msg)
	if err != nil {
		return nil, terror.Error(err)
	}
	defer rows.Close()

	cancelBidList := []CancelBidResponse{}
	for rows.Next() {
		var outputItem CancelBidResponse
		err := rows.Scan(&outputItem.TXID, &outputItem.Amount)
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

// MarketplaceLastSaleBid gets the last sale bid.
func MarketplaceLastSaleBid(itemSaleID uuid.UUID) (*boiler.ItemSalesBidHistory, error) {
	obj, err := boiler.ItemSalesBidHistories(
		boiler.ItemSalesBidHistoryWhere.ItemSaleID.EQ(itemSaleID.String()),
		boiler.ItemSalesBidHistoryWhere.CancelledAt.IsNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
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

// MarketplaceSaleItemExists checks whether given sales item exists.
func MarketplaceSaleItemExists(id uuid.UUID) (bool, error) {
	output, err := boiler.ItemSaleExists(gamedb.StdConn, id.String())
	if err != nil {
		return false, terror.Error(err)
	}
	return output, nil
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

// MarketplaceCountKeycards counts number of player's keycard for sale.
// This is used to check whether player can sell more.
func MarketplaceCountKeycards(playerKeycardID uuid.UUID) (int64, error) {
	output, err := boiler.ItemKeycardSales(
		boiler.ItemKeycardSaleWhere.ItemID.EQ(playerKeycardID.String()),
		boiler.ItemKeycardSaleWhere.SoldAt.IsNull(),
		boiler.ItemKeycardSaleWhere.EndAt.GT(time.Now()),
	).Count(gamedb.StdConn)
	if err != nil {
		return 0, terror.Error(err)
	}
	return output, nil
}

// ChangeMechOwner transfers a collection item to a new owner.
func ChangeMechOwner(conn boil.Executor, itemSaleID uuid.UUID) error {
	itemSale, err := boiler.FindItemSale(conn, itemSaleID.String())
	if err != nil {
		gamelog.L.Error().
			Err(err).
			Str("itemSaleID", itemSaleID.String()).
			Msg("ChangeMechOwner")
		return err
	}
	colItem, err := boiler.FindCollectionItem(conn, itemSale.CollectionItemID)
	if err != nil {
		gamelog.L.Error().
			Err(err).
			Str("itemSaleID", itemSaleID.String()).
			Msg("ChangeMechOwner")
		return err
	}

	mech, err := Mech(colItem.ItemID)
	if err != nil {
		gamelog.L.Error().
			Err(err).
			Str("itemSaleID", itemSaleID.String()).
			Msg("ChangeMechOwner")
		return err
	}

	_, err = boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(colItem.ItemID),
	).UpdateAll(conn, boiler.M{
		"owner_id": itemSale.SoldTo,
	})
	if err != nil {
		gamelog.L.Error().
			Err(err).
			Str("itemSaleID", itemSaleID.String()).
			Msg("ChangeMechOwner")
		return err
	}

	if mech.ChassisSkin != nil {
		_, err = boiler.CollectionItems(
			boiler.CollectionItemWhere.ItemID.EQ(mech.ChassisSkin.ID),
		).UpdateAll(conn, boiler.M{
			"owner_id": itemSale.SoldTo,
		})
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Str("itemSaleID", itemSaleID.String()).
				Msg("ChangeMechOwner")
			return err
		}
	}

	if mech.PowerCoreID.Valid {
		_, err = boiler.CollectionItems(
			boiler.CollectionItemWhere.ItemID.EQ(mech.PowerCoreID.String),
		).UpdateAll(conn, boiler.M{
			"owner_id": itemSale.SoldTo,
		})
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Str("itemSaleID", itemSaleID.String()).
				Msg("ChangeMechOwner")
			return err
		}
	}

	for _, w := range mech.Weapons {
		_, err = boiler.CollectionItems(
			boiler.CollectionItemWhere.ItemID.EQ(w.ID),
		).UpdateAll(conn, boiler.M{
			"owner_id": itemSale.SoldTo,
		})
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Str("itemSaleID", itemSaleID.String()).
				Msg("ChangeMechOwner")
			return err
		}
	}

	for _, u := range mech.Utility {
		_, err = boiler.CollectionItems(
			boiler.CollectionItemWhere.ItemID.EQ(u.ID),
		).UpdateAll(conn, boiler.M{
			"owner_id": itemSale.SoldTo,
		})
		if err != nil {
			gamelog.L.Error().
				Err(err).
				Str("itemSaleID", itemSaleID.String()).
				Msg("ChangeMechOwner")
			return err
		}
	}
	return nil
}

// ChangeMysteryCrateOwner transfers a collection item to a new owner.
func ChangeMysteryCrateOwner(conn boil.Executor, crateCollectionItemID string, newOwnerID string) error {
	_, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ID.EQ(crateCollectionItemID),
	).UpdateAll(conn,
		boiler.M{
			"owner_id": newOwnerID,
		})
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceKeycardSaleCreate inserts a new sale item.
func MarketplaceKeycardSaleCreate(
	conn boil.Executor,
	ownerID uuid.UUID,
	factionID uuid.UUID,
	listFeeTxnID string,
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
		SET count = excluded.count + 1`
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
func MarketplaceAddEvent(eventType string, amount decimal.NullDecimal, itemSaleID string, table string) error {
	obj := &boiler.MarketplaceEvent{
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
