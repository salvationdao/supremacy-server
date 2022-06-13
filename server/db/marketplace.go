package db

import (
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
)

var itemSaleQueryMods = []qm.QueryMod{
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
		item_sales.sold_by AS sold_by,
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

var itemKeycardSaleQueryMods = []qm.QueryMod{
	qm.Select(
		`item_keycard_sales.*,
		players.id AS "players.id",
		players.faction_id AS "players.faction_id",
		players.username AS "players.username",
		players.public_address AS "players.public_address",
		players.gid AS "players.gid",
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
}

// MarketplaceItemSale gets a specific item sale.
func MarketplaceItemSale(id uuid.UUID) (*server.MarketplaceSaleItem, error) {
	output := &server.MarketplaceSaleItem{}
	err := boiler.ItemSales(
		append(
			itemSaleQueryMods,
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
		&output.SoldBy,
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
			itemKeycardSaleQueryMods,
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
		&output.SoldBy,
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
	factionID string,
	search string,
	filter *ListFilterRequest,
	rarities []string,
	saleTypes []string,
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
		itemSaleQueryMods,
		boiler.ItemSaleWhere.SoldBy.IsNull(),
		boiler.ItemSaleWhere.EndAt.GT(time.Now()),
		boiler.ItemSaleWhere.DeletedAt.IsNull(),
		boiler.CollectionItemWhere.XsynLocked.EQ(false),
		boiler.CollectionItemWhere.MarketLocked.EQ(false),
	)

	if factionID != "" {
		queryMods = append(queryMods, boiler.ItemSaleWhere.FactionID.EQ(factionID))
	}

	// Filters
	if filter != nil {
		for i, f := range filter.Items {
			if f.Table != "" {
				if f.Table == boiler.TableNames.Mechs {
					column := MechColumns(f.Column)
					err := column.IsValid()
					if err != nil {
						return 0, nil, terror.Error(err)
					}
				}
			}
			queryMod := GenerateListFilterQueryMod(*f, i, filter.LinkOperator)
			queryMods = append(queryMods, queryMod)
		}
	}
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

	// Search
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			queryMods = append(queryMods, qm.And(
				fmt.Sprintf(
					"(to_tsvector('english', %s) @@ to_tsquery(?))",
					qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.Label),
				),
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
func MarketplaceItemKeycardSaleList(factionID string, search string, filter *ListFilterRequest, offset int, pageSize int, sortBy string, sortDir SortByDir) (int64, []*server.MarketplaceSaleItem1155, error) {
	if !sortDir.IsValid() {
		return 0, nil, terror.Error(fmt.Errorf("invalid sort direction"))
	}

	queryMods := append(
		itemKeycardSaleQueryMods,
		boiler.ItemKeycardSaleWhere.SoldBy.IsNull(),
		boiler.ItemKeycardSaleWhere.EndAt.GT(time.Now()),
		boiler.ItemKeycardSaleWhere.DeletedAt.IsNull(),
	)

	if factionID != "" {
		queryMods = append(queryMods, boiler.ItemKeycardSaleWhere.FactionID.EQ(factionID))
	}

	// Filters
	if filter != nil {
		for i, f := range filter.Items {
			queryMod := GenerateListFilterQueryMod(*f, i, filter.LinkOperator)
			queryMods = append(queryMods, queryMod)
		}
	}

	// Search
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			queryMods = append(queryMods, qm.And(
				fmt.Sprintf(
					"(to_tsvector('english', %s) @@ to_tsquery(?))",
					qm.Rels(boiler.TableNames.BlueprintKeycards, boiler.BlueprintKeycardColumns.Label),
				),
				xsearch,
			))
		}
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

// MarketplaceSaleCancelBids cancels all active bids and returns transaction ids needed to be retuned (ideally one).
func MarketplaceSaleCancelBids(conn boil.Executor, itemID uuid.UUID) ([]string, error) {
	q := `
		UPDATE item_sales_bid_history
		SET cancelled_at = NOW(),
			cancelled_reason = 'New Bid Placed'
		WHERE item_sale_id = $1 AND cancelled_at IS NULL
		RETURNING bid_tx_id`
	rows, err := conn.Query(q, itemID)
	if err != nil {
		return nil, terror.Error(err)
	}
	defer rows.Close()

	txidRefunds := []string{}
	for rows.Next() {
		var txid string
		err := rows.Scan(&txid)
		if err != nil {
			return nil, terror.Error(err)
		}
		txidRefunds = append(txidRefunds, txid)
	}
	return txidRefunds, nil
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
	q := `
		UPDATE collection_items AS ci
		SET owner_id = s.sold_by
		FROM item_sales s
			INNER JOIN collection_items ci_mech ON ci_mech.id = s.collection_item_id
				AND ci_mech.item_type = $2
			INNER JOIN mechs m ON m.id = ci_mech.item_id
			LEFT JOIN collection_items ci_mech_skin ON ci_mech_skin.item_id = m.chassis_skin_id
				AND ci_mech_skin.item_type = $3
			LEFT JOIN collection_items ci_power_core ON ci_power_core.item_id = m.power_core_id
				AND ci_power_core.item_type = $4
		WHERE s.id = $1
			AND s.sold_by IS NOT NULL
			AND ci.id in (ci_mech.id, ci_mech_skin.id, ci_power_core.id)`
	_, err := conn.Exec(q, itemSaleID, boiler.ItemTypeMech, boiler.ItemTypeMechSkin, boiler.ItemTypePowerCore)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ChangeMysteryCrateOwner transfers a collection item to a new owner.
func ChangeMysteryCrateOwner(conn boil.Executor, itemSaleID uuid.UUID) error {
	q := `
		UPDATE collection_items ci
		SET owner_id = s.sold_by
		FROM item_sales s
		WHERE s.id = $1
			AND ci.id = s.collection_item_id`
	_, err := conn.Exec(q, itemSaleID, boiler.ItemTypeMysteryCrate)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceKeycardSaleCreate inserts a new sale item.
func MarketplaceKeycardSaleCreate(
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

		SELECT iks.sold_by as player_id, pk.blueprint_keycard_id, 1 AS count
		FROM item_keycard_sales iks
			INNER JOIN player_keycards pk ON pk.id = iks.item_id
		WHERE iks.id = $1 AND iks.sold_by IS NOT NULL
		ON CONFLICT (player_id, blueprint_keycard_id)
		DO UPDATE 
		SET COUNT = excluded.count + 1`
	_, err := conn.Exec(q, itemSaleID)
	if err != nil {
		return terror.Error(err)
	}

	q = `
		UPDATE player_keycards AS pk
		SET count = count - 1
		FROM item_keycard_sales iks
		WHERE iks.id = $1
			AND pk.id = iks.item_id`
	_, err = conn.Exec(q, itemSaleID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
