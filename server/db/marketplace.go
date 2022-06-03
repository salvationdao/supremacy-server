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
		item_sales.item_id AS item_id,
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
		item_sales.deleted_at AS deleted_at,
		item_sales.updated_at AS updated_at,
		item_sales.created_at AS created_at,
		players.id AS "players.id",
		players.username AS "players.username",
		players.public_address AS "players.public_address",
		players.gid AS "players.gid",
		collection_items.tier AS "collection_items.tier",
		mechs.id AS "mechs.id",
		mechs.name AS "mechs.name",
		mechs.label AS "mechs.label",
		mech_skin.avatar_url AS "mech_skin.avatar_url"`,
	),
	qm.InnerJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.CollectionItems,
			qm.Rels(boiler.TableNames.CollectionItems, boiler.CollectionItemColumns.ItemID),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ItemID),
		),
	),
	qm.LeftOuterJoin(
		fmt.Sprintf(
			"%s ON %s = %s",
			boiler.TableNames.Mechs,
			qm.Rels(boiler.TableNames.Mechs, boiler.MechColumns.ID),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.ItemID),
		),
	),
	qm.LeftOuterJoin(
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
			boiler.TableNames.Players,
			qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.ID),
			qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.OwnerID),
		),
	),
}

var itemKeycardSaleQueryMods = []qm.QueryMod{
	qm.Select(
		`item_keycard_sales.*,
		players.id AS "players.id",
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
		&output.ItemID,
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
		&output.DeletedAt,
		&output.UpdatedAt,
		&output.CreatedAt,
		&output.Owner.ID,
		&output.Owner.Username,
		&output.Owner.PublicAddress,
		&output.Owner.Gid,
		&output.Mech.ID,
		&output.Mech.Label,
		&output.Mech.Name,
		&output.Mech.Tier,
		&output.Mech.AvatarURL,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return output, nil
}

// MarketplaceItemKeycardSale gets a specific keycard item sale.
func MarketplaceItemKeycardSale(id uuid.UUID) (*server.MarketplaceKeycardSaleItem, error) {
	output := &server.MarketplaceKeycardSaleItem{}
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
		&output.DeletedAt,
		&output.UpdatedAt,
		&output.CreatedAt,
		&output.Owner.ID,
		&output.Owner.Username,
		&output.Owner.PublicAddress,
		&output.Owner.Gid,
		&output.Blueprints.ID,
		&output.Blueprints.Label,
		&output.Blueprints.Description,
		&output.Blueprints.Collection,
		&output.Blueprints.KeycardTokenID,
		&output.Blueprints.ImageURL,
		&output.Blueprints.AnimationURL,
		&output.Blueprints.KeycardGroup,
		&output.Blueprints.Syndicate,
		&output.Blueprints.CreatedAt,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return output, nil
}

// MarketplaceItemSaleList returns a numeric paginated result of sales list.
func MarketplaceItemSaleList(search string, filter *ListFilterRequest, rarities []string, excludeUserID string, offset int, pageSize int, sortBy string, sortDir SortByDir) (int64, []*server.MarketplaceSaleItem, error) {
	if !sortDir.IsValid() {
		return 0, nil, terror.Error(fmt.Errorf("invalid sort direction"))
	}

	queryMods := append(
		itemSaleQueryMods,
		boiler.ItemSaleWhere.OwnerID.NEQ(excludeUserID),
		boiler.ItemSaleWhere.SoldBy.IsNull(),
		boiler.ItemSaleWhere.EndAt.GT(time.Now()),
	)

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
	orderBy := qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.CreatedAt), sortDir))
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
	for _, r := range records {
		fmt.Println("Test", r.ID)
	}

	return total, records, nil
}

// MarketplaceItemKeycardSaleList returns a numeric paginated result of keycard sales list.
func MarketplaceItemKeycardSaleList(search string, filter *ListFilterRequest, excludeUserID string, offset int, pageSize int, sortBy string, sortDir SortByDir) (int64, []*server.MarketplaceKeycardSaleItem, error) {
	if !sortDir.IsValid() {
		return 0, nil, terror.Error(fmt.Errorf("invalid sort direction"))
	}

	queryMods := append(
		itemKeycardSaleQueryMods,
		boiler.ItemKeycardSaleWhere.OwnerID.NEQ(excludeUserID),
		boiler.ItemKeycardSaleWhere.SoldBy.IsNull(),
		boiler.ItemKeycardSaleWhere.EndAt.GT(time.Now()),
	)

	// Filters
	// if filter != nil {
	// 	for i, f := range filter.Items {
	// 		if f.Table != "" {
	// 			if f.Table == boiler.TableNames.BlueprintKeycards {
	// 				column := MechColumns(f.Column)
	// 				err := column.IsValid()
	// 				if err != nil {
	// 					return 0, nil, terror.Error(err)
	// 				}
	// 			}
	// 		}
	// 		queryMod := GenerateListFilterQueryMod(*f, i, filter.LinkOperator)
	// 		queryMods = append(queryMods, queryMod)
	// 	}
	// }

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
		return 0, []*server.MarketplaceKeycardSaleItem{}, nil
	}

	// Sort
	orderBy := qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.CreatedAt), sortDir))
	queryMods = append(queryMods, orderBy)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	records := []*server.MarketplaceKeycardSaleItem{}
	err = boiler.ItemKeycardSales(queryMods...).Bind(nil, gamedb.StdConn, &records)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	return total, records, nil
}

// MarketplaceSaleCreate inserts a new sale item.
func MarketplaceSaleCreate(
	ownerID uuid.UUID,
	factionID uuid.UUID,
	listFeeTxnID string,
	endAt time.Time,
	itemID uuid.UUID,
	hasBuyout bool,
	askingPrice *decimal.Decimal,
	hasAuction bool,
	auctionReservedPrice *decimal.Decimal,
	hasDutchAuction bool,
	dutchAuctionDropRate *decimal.Decimal,
) (*server.MarketplaceSaleItem, error) {
	obj := &boiler.ItemSale{
		OwnerID:        ownerID.String(),
		FactionID:      factionID.String(),
		ListingFeeTXID: listFeeTxnID,
		ItemID:         itemID.String(),
		EndAt:          endAt,
	}

	if hasBuyout {
		obj.Buyout = true
		obj.BuyoutPrice = null.StringFrom(askingPrice.String())
	}
	if hasAuction {
		obj.Auction = true
		obj.AuctionCurrentPrice = null.StringFrom(auctionReservedPrice.String())
		obj.AuctionReservedPrice = null.StringFrom(auctionReservedPrice.String())
	}
	if hasDutchAuction {
		obj.DutchAuction = true
		obj.DutchAuctionDropRate = null.StringFrom(dutchAuctionDropRate.String())
	}

	err := obj.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	output := &server.MarketplaceSaleItem{
		ID:                   obj.ID,
		FactionID:            obj.FactionID,
		ItemID:               obj.ItemID,
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
func MarketplaceSaleCancelBids(itemID uuid.UUID) ([]string, error) {
	q := `
		UPDATE item_sale_bid_history
		SET canceled_at = NOW()
		WHERE item_sale_id = $1 AND canceled_at IS NULL
		RETURNING bid_tx_id`
	rows, err := gamedb.StdConn.Query(q, itemID)
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
func MarketplaceSaleBidHistoryRefund(itemID uuid.UUID, txID, refundTxID string) error {
	q := `
		UPDATE item_sale_bid_history
		SET refund_bid_tx_id = $3
		WHERE item_sale_id = $1
			AND bid_tx_id = $2`
	_, err := gamedb.StdConn.Exec(q, itemID, txID, refundTxID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// MarketplaceSaleBidHistoryCreate inserts a new bid history record.
func MarketplaceSaleBidHistoryCreate(id uuid.UUID, bidderUserID uuid.UUID, amount decimal.Decimal, txid string) (*boiler.ItemSalesBidHistory, error) {
	obj := &boiler.ItemSalesBidHistory{
		ItemSaleID: id.String(),
		BidderID:   bidderUserID.String(),
		BidTXID:    txid,
		BidPrice:   amount.String(),
	}
	err := obj.Insert(gamedb.StdConn, boil.Infer())
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
func MarketplaceSaleAuctionSync(id uuid.UUID) error {
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
	_, err := gamedb.StdConn.Exec(q, id)
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
func MarketplaceCheckCollectionItem(mechID uuid.UUID) (bool, error) {
	output, err := boiler.ItemSales(
		boiler.ItemSaleWhere.ItemID.EQ(mechID.String()),
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
// TODO: Subject to change...
func ChangeMechOwner(conn boil.Executor, itemSaleID uuid.UUID) error {
	q := `
		UPDATE collection_items AS ci
		SET owner_id = s.sold_by
		FROM item_sales s
		WHERE s.id = $1
			AND ci.id IN (
				SELECT _ci.id
				FROM mechs _m
					INNER JOIN collection_items _ci ON _ci.item_id = _m.id
						OR _ci.item_id = _m.chassis_skin_id
						OR _ci.item_id = _m.power_core_id
				WHERE _m.id = s.item_id
			)`
	_, err := conn.Exec(q, itemSaleID)
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
) (*server.MarketplaceKeycardSaleItem, error) {
	obj := &boiler.ItemKeycardSale{
		OwnerID:        ownerID.String(),
		FactionID:      factionID.String(),
		ListingFeeTXID: listFeeTxnID,
		ItemID:         itemID.String(),
		EndAt:          endAt,
		BuyoutPrice:    askingPrice.String(),
	}
	err := obj.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	output := &server.MarketplaceKeycardSaleItem{
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
