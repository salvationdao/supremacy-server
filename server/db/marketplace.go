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
	qm.Load(boiler.ItemSaleRels.Owner),
}

var itemKeycardSaleQueryMods = []qm.QueryMod{
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
			qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.ItemID),
		),
	),
	qm.Load(qm.Rels(boiler.ItemKeycardSaleRels.Item, boiler.PlayerKeycardRels.BlueprintKeycard)),
}

// MarketplaceItemSale gets a specific item sale.
func MarketplaceItemSale(id uuid.UUID) (*server.MarketplaceSaleItem, error) {
	item, err := boiler.ItemSales(
		append(
			itemSaleQueryMods,
			boiler.ItemSaleWhere.ID.EQ(id.String()),
		)...,
	).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}

	// Get sale item details and the item for sale
	output := &server.MarketplaceSaleItem{
		ItemSale: item,
		Owner:    item.R.Owner,
	}
	mech, err := Mech(item.ItemID)
	if err != nil {
		return nil, terror.Error(err)
	}
	output.Mech = mech
	return output, nil
}

// MarketplaceItemKeycardSale gets a specific keycard item sale.
func MarketplaceItemKeycardSale(id uuid.UUID) (*server.MarketplaceKeycardSaleItem, error) {
	item, err := boiler.ItemKeycardSales(
		append(
			itemSaleQueryMods,
			boiler.ItemKeycardSaleWhere.ID.EQ(id.String()),
		)...,
	).One(gamedb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}

	// Get sale item details and the item for sale
	output := &server.MarketplaceKeycardSaleItem{
		ItemKeycardSale: item,
		Owner:           item.R.Owner,
	}
	if item.R != nil && item.R.Item != nil && item.R.Item.R != nil && item.R.Item.R.BlueprintKeycard != nil {
		output.Blueprints = &server.AssetKeycardBlueprint{
			ID:             item.R.Item.R.BlueprintKeycard.ID,
			Label:          item.R.Item.R.BlueprintKeycard.Label,
			Description:    item.R.Item.R.BlueprintKeycard.Description,
			Collection:     item.R.Item.R.BlueprintKeycard.Collection,
			KeycardTokenID: item.R.Item.R.BlueprintKeycard.KeycardTokenID,
			ImageURL:       item.R.Item.R.BlueprintKeycard.ImageURL,
			AnimationURL:   item.R.Item.R.BlueprintKeycard.AnimationURL,
			KeycardGroup:   item.R.Item.R.BlueprintKeycard.KeycardGroup,
			Syndicate:      item.R.Item.R.BlueprintKeycard.Syndicate,
			CreatedAt:      item.R.Item.R.BlueprintKeycard.CreatedAt,
		}
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

	itemSales, err := boiler.ItemSales(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	// Load in related items
	records := []*server.MarketplaceSaleItem{}
	itemIDs := []string{}
	mechIDs := []string{}
	for _, row := range itemSales {
		// if row.ItemType == boiler.ItemTypeMech {
		mechIDs = append(mechIDs, row.ItemID)
		itemIDs = append(itemIDs, row.ItemID)
		// }
		records = append(records, &server.MarketplaceSaleItem{
			ItemSale: row,
			Owner:    row.R.Owner,
		})
	}
	if len(itemIDs) > 0 {
		collectionItems, err := boiler.CollectionItems(
			boiler.CollectionItemWhere.ItemID.IN(itemIDs),
		).All(gamedb.StdConn)
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		for i, row := range records {
			for _, collection := range collectionItems {
				// if row.ItemType == boiler.ItemTypeMech && row.ItemID == mech.ID {
				if row.ItemID == collection.ItemID {
					records[i].Collection = collection
					break
				}
			}
		}
	}
	if len(mechIDs) > 0 {
		mechs, err := Mechs(mechIDs...)
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		for i, row := range records {
			for _, mech := range mechs {
				// if row.ItemType == boiler.ItemTypeMech && row.ItemID == mech.ID {
				if row.ItemID == mech.ID {
					records[i].Mech = mech
					break
				}
			}
		}
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
	total, err := boiler.ItemSales(queryMods...).Count(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if total == 0 {
		return 0, []*server.MarketplaceKeycardSaleItem{}, nil
	}

	// Sort
	orderBy := qm.OrderBy(fmt.Sprintf("%s %s", qm.Rels(boiler.TableNames.ItemSales, boiler.ItemSaleColumns.CreatedAt), sortDir))
	queryMods = append(queryMods, orderBy)

	// Limit/Offset
	if pageSize > 0 {
		queryMods = append(queryMods, qm.Limit(pageSize), qm.Offset(offset))
	}

	itemSales, err := boiler.ItemKeycardSales(queryMods...).All(gamedb.StdConn)
	if err != nil {
		return 0, nil, terror.Error(err)
	}

	// Load in related items
	records := []*server.MarketplaceKeycardSaleItem{}
	for _, row := range itemSales {
		item := &server.MarketplaceKeycardSaleItem{
			ItemKeycardSale: row,
			Owner:           row.R.Owner,
		}
		if item.R != nil && item.R.Item != nil && item.R.Item.R != nil && item.R.Item.R.BlueprintKeycard != nil {
			item.Blueprints = &server.AssetKeycardBlueprint{
				ID:             item.R.Item.R.BlueprintKeycard.ID,
				Label:          item.R.Item.R.BlueprintKeycard.Label,
				Description:    item.R.Item.R.BlueprintKeycard.Description,
				Collection:     item.R.Item.R.BlueprintKeycard.Collection,
				KeycardTokenID: item.R.Item.R.BlueprintKeycard.KeycardTokenID,
				ImageURL:       item.R.Item.R.BlueprintKeycard.ImageURL,
				AnimationURL:   item.R.Item.R.BlueprintKeycard.AnimationURL,
				KeycardGroup:   item.R.Item.R.BlueprintKeycard.KeycardGroup,
				Syndicate:      item.R.Item.R.BlueprintKeycard.Syndicate,
				CreatedAt:      item.R.Item.R.BlueprintKeycard.CreatedAt,
			}
		}

		records = append(records, item)
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
		obj.DutchActionDropRate = null.StringFrom(dutchAuctionDropRate.String())
	}

	err := obj.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	output := &server.MarketplaceSaleItem{
		ItemSale: obj,
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

// MarketplaceCheckMech checks whether mech is already in marketplace.
func MarketplaceCheckMech(mechID uuid.UUID) (bool, error) {
	output, err := boiler.ItemSales(
		boiler.ItemSaleWhere.ItemID.EQ(mechID.String()),
	).Exists(gamedb.StdConn)
	if err != nil {
		return false, terror.Error(err)
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
	_, err := gamedb.StdConn.Exec(q, itemSaleID)
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
	askingPrice *decimal.Decimal,
) (*server.MarketplaceKeycardSaleItem, error) {
	obj := &boiler.ItemKeycardSale{
		OwnerID:        ownerID.String(),
		FactionID:      factionID.String(),
		ListingFeeTXID: listFeeTxnID,
		ItemID:         itemID.String(),
		EndAt:          endAt,
		BuyoutPrice:    null.StringFrom(askingPrice.String()),
	}
	err := obj.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	output := &server.MarketplaceKeycardSaleItem{
		ItemKeycardSale: obj,
	}
	return output, nil
}

// ChangeKeycardOwner changes a keycard from previous owner to new owner.
func ChangeKeycardOwner(itemSaleID uuid.UUID) error {
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, "Failed to update player.")
	}

	q := `
		INSERT INTO player_keycards (player_id, blueprint_keycard_id, count)

		SELECT iks.sold_by as player_id, pk.blueprint_keycard_id, 1 AS count
		FROM item_keycard_sales iks
			INNER JOIN player_keycards pk ON pk.id = iks.item_id
		WHERE iks.id = $1 AND iks.sold_by IS NOT NULL
		ON CONFLICT (player_id, blueprint_keycard_id)
		DO UPDATE 
		SET COUNT = excluded.count + 1`
	_, err = tx.Exec(q, itemSaleID)
	if err != nil {
		return terror.Error(err)
	}

	q = `
		UPDATE player_keycards AS pk
		SET count = count - 1
		FROM item_keycard_sales iks
		WHERE iks.id = $1
			AND pk.id = iks.item_id`
	_, err = tx.Exec(q, itemSaleID)
	if err != nil {
		return terror.Error(err)
	}

	q = `
		DELETE FROM player_keycards 
		WHERE count = 0`
	_, err = tx.Exec(q, itemSaleID)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
