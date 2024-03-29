package marketplace

import (
	"fmt"
	"math"
	"server"
	"server/asset"
	"server/benchmark"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type MarketplaceController struct {
	Passport *xsyn_rpcclient.XsynXrpcClient
}

type ItemSaleAuction struct {
	ID                   uuid.UUID           `boil:"id"`
	CollectionItemID     uuid.UUID           `boil:"collection_item_id"`
	ItemType             string              `boil:"item_type"`
	ItemLocked           bool                `boil:"item_locked"`
	Hash                 string              `boil:"hash"`
	OwnerID              uuid.UUID           `boil:"owner_id"`
	Auction              bool                `boil:"auction"`
	AuctionReservedPrice decimal.NullDecimal `boil:"auction_reserved_price"`
	BuyoutPrice          decimal.NullDecimal `boil:"buyout_price"`
	DutchAuction         bool                `boil:"dutch_auction"`
	DutchAuctionDropRate decimal.NullDecimal `boil:"dutch_auction_drop_rate"`
	AuctionBidPrice      decimal.Decimal     `boil:"auction_bid_price"`
	AuctionBidUserID     uuid.UUID           `boil:"auction_bid_user_id"`
	AuctionBidTXID       string              `boil:"auction_bid_tx_id"`
	FactionID            uuid.UUID           `boil:"faction_id"`
	CreatedAt            time.Time           `boil:"created_at"`
}

type AttributeInner struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

func NewMarketplaceController(pp *xsyn_rpcclient.XsynXrpcClient) *MarketplaceController {
	m := &MarketplaceController{pp}
	go m.Run()
	return m
}

func (m *MarketplaceController) Run() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the MarketplaceController!", r)
		}
	}()

	mainTicker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-mainTicker.C:
			bm := benchmark.New()

			bm.Start("finished_auctions")
			m.processFinishedAuctions()
			bm.End("finished_auctions")
			bm.Start("expired_keycards")
			m.processExpiredKeycardItemListings()
			bm.End("expired_keycards")
			bm.Start("unlock_in_marketplace_items")
			m.unlockCollectionItems()
			bm.End("unlock_in_marketplace_items")

			bm.Alert(60000)
		}
	}
}

// Unlocks all collection items that are no longer for sale.
// This function only processes expired listed items within 2 minutes.
func (m *MarketplaceController) unlockCollectionItems() {
	_, err := boiler.NewQuery(
		qm.SQL(`
			UPDATE collection_items
			SET locked_to_marketplace = false
			WHERE locked_to_marketplace = true AND id IN (
				SELECT _s.collection_item_id
				FROM item_sales _s
				WHERE _s.sold_to IS NULL
					AND (
						_s.end_at <= NOW()
						AND NOW() - _s.end_at < INTERVAL '2 MINUTE' 
					)
			) 
		`),
	).Exec(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "itemSaleUpdateLockToMarketplace").
			Err(err).Msg("unable to update all collection items lock to marketplace")
	}
}

// Notifies Xsyn on Keycards that have recently expired.
func (m *MarketplaceController) processExpiredKeycardItemListings() {
	gamelog.L.Trace().Msg("processing expired keycard sale items started")

	// Query expired items
	queryMods := append(db.ItemKeycardSaleQueryMods,
		qm.Where(fmt.Sprintf(
			`(%[1]s <= NOW() AND NOW() - %[1]s <= INTERVAL '2 MINUTE')`,
			qm.Rels(boiler.TableNames.ItemKeycardSales, boiler.ItemKeycardSaleColumns.EndAt),
		)),
		boiler.ItemKeycardSaleWhere.SoldTo.IsNull(),
		boiler.ItemKeycardSaleWhere.DeletedAt.IsNull(),
	)

	records := []*server.MarketplaceSaleItem1155{}
	err := boiler.ItemKeycardSales(queryMods...).Bind(nil, gamedb.StdConn, &records)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "expireKeycardItemListings").
			Err(err).Msg("unable to get unprocessed expired keycard listings")
	}

	// Process expired items
	numProcessed := 0
	for _, expiredItem := range records {
		var assetJson types.JSON
		inner := &AttributeInner{
			TraitType: "Syndicate",
			Value:     "N/A",
		}
		err := assetJson.Marshal(inner)
		if err != nil {
			gamelog.L.Error().
				Str("item_id", expiredItem.ID).
				Str("user_id", expiredItem.OwnerID).
				Int("keycard_token_id", expiredItem.Keycard.KeycardTokenID).
				Err(err).
				Msg("unable to marshal XSYN asset count")
			continue
		}

		// Prepare to flag delete
		func() {
			tx, err := gamedb.StdConn.Begin()
			if err != nil {
				gamelog.L.Error().
					Str("item_id", expiredItem.ID).
					Str("user_id", expiredItem.OwnerID).
					Int("keycard_token_id", expiredItem.Keycard.KeycardTokenID).
					Err(err).
					Msg("unable to start db transaction (expired XSYN asset count)")
				return
			}

			// Restore Keycard Increment
			playerKeycardID, err := uuid.FromString(expiredItem.ItemID)
			if err != nil {
				gamelog.L.Error().
					Str("item_id", expiredItem.ID).
					Str("user_id", expiredItem.OwnerID).
					Int("keycard_token_id", expiredItem.Keycard.KeycardTokenID).
					Err(err).
					Msg("unable to parse player keycard id")
				return
			}

			err = db.IncrementPlayerKeycardCount(tx, playerKeycardID)
			if err != nil {
				gamelog.L.Error().
					Str("item_id", expiredItem.ID).
					Str("user_id", expiredItem.OwnerID).
					Int("keycard_token_id", expiredItem.Keycard.KeycardTokenID).
					Err(err).
					Msg("unable to restore increment player keycard count")
				return
			}

			// Archive Sale Item
			itemSaleKeycard := &boiler.ItemKeycardSale{
				ID:        expiredItem.ID,
				DeletedAt: null.TimeFrom(time.Now()),
			}
			_, err = itemSaleKeycard.Update(tx, boil.Whitelist(boiler.ItemKeycardSaleColumns.DeletedAt))
			if err != nil {
				gamelog.L.Error().
					Str("item_id", expiredItem.ID).
					Str("user_id", expiredItem.OwnerID).
					Int("keycard_token_id", expiredItem.Keycard.KeycardTokenID).
					Err(err).
					Msg("unable to start db transaction (expired XSYN asset count)")
				return
			}

			// Notify Xsyn about expired keycard
			keycardUpdate := &xsyn_rpcclient.Asset1155CountUpdateSupremacyReq{
				ApiKey:         m.Passport.ApiKey,
				TokenID:        expiredItem.Keycard.KeycardTokenID,
				Address:        expiredItem.Owner.PublicAddress.String,
				CollectionSlug: expiredItem.Keycard.Collection,
				Amount:         1,
				ImageURL:       expiredItem.Keycard.ImageURL,
				AnimationURL:   expiredItem.Keycard.AnimationURL,
				KeycardGroup:   expiredItem.Keycard.KeycardGroup,
				Attributes:     assetJson,
				IsAdd:          true,
			}
			_, err = m.Passport.UpdateKeycardCountXSYN(keycardUpdate)
			if err != nil {
				gamelog.L.Error().
					Str("item_id", expiredItem.ID).
					Str("user_id", expiredItem.OwnerID).
					Int("keycard_token_id", expiredItem.Keycard.KeycardTokenID).
					Err(err).
					Msg("unable to update XSYN asset count")
				return
			}

			// Commit Transaction
			err = tx.Commit()
			if err != nil {
				gamelog.L.Error().
					Str("item_id", expiredItem.ID).
					Str("user_id", expiredItem.OwnerID).
					Int("keycard_token_id", expiredItem.Keycard.KeycardTokenID).
					Err(err).
					Msg("unable to commit db transaction (expired XSYN asset count)")
				return
			}
			numProcessed++
		}()
	}
	gamelog.L.Trace().
		Int("num_processed", numProcessed).
		Int("num_failed", len(records)-numProcessed).
		Int("num_pending", len(records)).
		Msg("processing expired keycard sale items finished")
}

// Scan all completed auctions that haven't went through payment process.
// This also processes and bids that exceed the dutch auction drop rate.
func (m *MarketplaceController) processFinishedAuctions() {
	gamelog.L.Trace().Msg("processing completed auction items started")

	auctions := []*ItemSaleAuction{}
	err := boiler.NewQuery(
		qm.SQL(`
			SELECT item_sales.id AS id,
				item_sales.collection_item_id,
				collection_items.item_type,
				collection_items.hash,
				item_sales.owner_id,
				item_sales.auction,
				item_sales.auction_reserved_price,
				item_sales.buyout_price,
				item_sales.dutch_auction,
				item_sales.dutch_auction_drop_rate,
				item_sales.created_at,
				(collection_items.xsyn_locked OR collection_items.market_locked) AS item_locked,
				item_sales_bid_history.bid_price AS auction_bid_price,
				item_sales_bid_history.bidder_id AS auction_bid_user_id,
				item_sales_bid_history.bid_tx_id AS auction_bid_tx_id,
				players.faction_id
			FROM item_sales 
				INNER JOIN item_sales_bid_history ON item_sales_bid_history.item_sale_id = item_sales.id
					AND item_sales_bid_history.cancelled_at IS NULL
					AND item_sales_bid_history.refund_bid_tx_id IS NULL
				INNER JOIN players ON players.id = item_sales.owner_id 
				INNER JOIN collection_items ON collection_items.id = item_sales.collection_item_id
			WHERE item_sales.auction = TRUE
				AND item_sales.sold_to IS NULL
				AND item_sales_bid_history.bid_price > 0
				AND item_sales.deleted_at IS NULL
				AND (
					item_sales.end_at <= NOW()
					OR (
						item_sales.dutch_auction = TRUE
						AND item_sales.buyout_price IS NOT NULL
						AND item_sales.dutch_auction_drop_rate IS NOT NULL
					)
					OR collection_items.xsyn_locked = true
					OR collection_items.market_locked = true
				)`),
	).Bind(nil, gamedb.StdConn, &auctions)
	if err != nil {
		gamelog.L.Error().
			Str("db func", "itemSales").
			Err(err).Msg("unable to retrieve completed auctions on marketplace")
	}

	numProcessed := 0
	for _, auctionItem := range auctions {
		func() {
			l := gamelog.L.With().
				Str("item_id", auctionItem.ID.String()).
				Str("user_id", auctionItem.AuctionBidUserID.String()).
				Str("cost", auctionItem.AuctionBidPrice.String()).
				Logger()

			// Check if current bid is below reserved price and issue refunds.
			if auctionItem.ItemLocked || (auctionItem.Auction && auctionItem.AuctionReservedPrice.Valid && auctionItem.AuctionReservedPrice.Decimal.GreaterThan(auctionItem.AuctionBidPrice)) {
				factionAccountID, ok := server.FactionUsers[auctionItem.FactionID.String()]
				if !ok {
					l.Error().Err(err).Str("bidTID", auctionItem.AuctionBidTXID).Msg("unable to get find faction account")
					return
				}
				factID := uuid.Must(uuid.FromString(factionAccountID))
				syndicateBalance := m.Passport.UserBalanceGet(factID)
				if syndicateBalance.LessThanOrEqual(auctionItem.AuctionBidPrice) {
					txid, err := m.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
						FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
						ToUserID:             factID,
						Amount:               auctionItem.AuctionBidPrice.StringFixed(0),
						TransactionReference: server.TransactionReference(fmt.Sprintf("bid_refunds|%s|%d", auctionItem.AuctionBidUserID, time.Now().UnixNano())),
						Group:                string(server.TransactionGroupSupremacy),
						SubGroup:             string(server.TransactionGroupMarketplace),
						Description:          fmt.Sprintf("Bid Refund for Player: %s (item sale: %s)", auctionItem.AuctionBidUserID, auctionItem.ID),
					})
					if err != nil {
						l.Error().
							Str("Faction ID", factionAccountID).
							Str("Amount", auctionItem.AuctionBidPrice.StringFixed(0)).
							Err(err).
							Msg("Could not transfer money from treasury into syndicate account!!")
						return
					}
					l.Warn().
						Str("Faction ID", factionAccountID).
						Str("Amount", auctionItem.AuctionBidPrice.StringFixed(0)).
						Str("TXID", txid).
						Err(err).
						Msg("Had to transfer funds to the syndicate account")
				}

				rtxid, err := m.Passport.RefundSupsMessage(auctionItem.AuctionBidTXID)
				if err != nil {
					l.Error().
						Str("bid_tx_id", auctionItem.AuctionBidTXID).
						Err(err).
						Msg("unable to refund cancelled auction bid transaction")
					return
				}
				err = db.MarketplaceSaleBidHistoryRefund(gamedb.StdConn, auctionItem.ID, auctionItem.AuctionBidTXID, rtxid, true)
				if err != nil {
					l.Error().
						Str("bid_tx_id", auctionItem.AuctionBidTXID).
						Str("refund_tx_id", rtxid).
						Err(err).
						Msg("unable to update refund tx id on bid record")
					return
				}
				err = db.MarketplaceAddEvent(boiler.MarketplaceEventBidRefund, auctionItem.AuctionBidUserID.String(), decimal.NewNullDecimal(auctionItem.AuctionBidPrice), auctionItem.ID.String(), boiler.TableNames.ItemSales)
				if err != nil {
					l.Error().
						Str("bid_tx_id", auctionItem.AuctionBidTXID).
						Str("refund_tx_id", rtxid).
						Err(err).
						Msg("Failed to log bid refund event.")
				}
				numProcessed++
				return
			}

			// Check if Dutch Auction and is below the bidder's price, bidder wins
			if auctionItem.DutchAuction {
				minutesLapse := decimal.NewFromFloat(math.Floor(time.Now().Sub(auctionItem.CreatedAt).Minutes()))
				dutchAuctionAmount := auctionItem.BuyoutPrice.Decimal.Sub(auctionItem.DutchAuctionDropRate.Decimal.Mul(minutesLapse))

				if auctionItem.AuctionReservedPrice.Valid {
					if dutchAuctionAmount.LessThan(auctionItem.AuctionReservedPrice.Decimal) {
						dutchAuctionAmount = auctionItem.AuctionReservedPrice.Decimal
					}
				} else {
					if dutchAuctionAmount.LessThanOrEqual(decimal.Zero) {
						dutchAuctionAmount = decimal.New(1, 18)
					}
				}

				if dutchAuctionAmount.GreaterThan(auctionItem.AuctionBidPrice) {
					numProcessed++
					return
				}
			}

			// Get Faction Account sending bid amount to
			factionAccountID, ok := server.FactionUsers[auctionItem.FactionID.String()]
			if !ok {
				err = fmt.Errorf("failed to get hard coded syndicate player id")
				l.Error().
					Str("owner_id", auctionItem.OwnerID.String()).
					Str("faction_id", auctionItem.FactionID.String()).
					Err(err).
					Msg("unable to get hard coded syndicate player ID from faction ID")
				return
			}

			// Transfer Sups to Owner
			salesCutPercentageFee := db.GetDecimalWithDefault(db.KeyMarketplaceSaleCutPercentageFee, decimal.NewFromFloat(0.1))
			salesCutAmount := auctionItem.AuctionBidPrice.Mul(decimal.NewFromInt(1).Sub(salesCutPercentageFee))
			factionAccountUUID := uuid.Must(uuid.FromString(factionAccountID))

			syndicateBalance := m.Passport.UserBalanceGet(factionAccountUUID)
			if syndicateBalance.LessThanOrEqual(salesCutAmount) {
				txid, err := m.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
					FromUserID:           uuid.UUID(server.XsynTreasuryUserID),
					ToUserID:             factionAccountUUID,
					Amount:               salesCutAmount.StringFixed(0),
					TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item|auction|%s|%d", auctionItem.ID, time.Now().UnixNano())),
					Group:                string(server.TransactionGroupSupremacy),
					SubGroup:             string(server.TransactionGroupMarketplace),
					Description:          fmt.Sprintf("Marketplace Buy Item Payment (%d%% cut): %s", salesCutPercentageFee.Mul(decimal.NewFromInt(100)).IntPart(), auctionItem.ID),
				})
				if err != nil {
					l.Error().
						Str("Faction ID", factionAccountID).
						Str("Amount", salesCutAmount.StringFixed(0)).
						Err(err).
						Msg("Could not transfer money from treasury into syndicate account!!")
					return
				}
				l.Warn().
					Str("Faction ID", factionAccountID).
					Str("Amount", salesCutAmount.StringFixed(0)).
					Str("TXID", txid).
					Err(err).
					Msg("Had to transfer funds to the syndicate account")
			}
			txid, err := m.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
				FromUserID:           factionAccountUUID,
				ToUserID:             uuid.Must(uuid.FromString(auctionItem.OwnerID.String())),
				Amount:               salesCutAmount.String(),
				TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_item|auction|%s|%d", auctionItem.ID.String(), time.Now().UnixNano())),
				Group:                string(server.TransactionGroupSupremacy),
				SubGroup:             string(server.TransactionGroupMarketplace),
				Description:          fmt.Sprintf("Marketplace Buy Item Payment (%d%% cut): %s", salesCutPercentageFee.Mul(decimal.NewFromInt(100)).IntPart(), auctionItem.ID),
			})
			if err != nil {
				l.Error().Err(err).Msg("Failed to send sups to item seller.")
				return
			}

			// Begin Transaction
			tx, err := gamedb.StdConn.Begin()
			if err != nil {
				l.Error().Err(err).Msg("Failed to start db transaction.")
				return
			}
			defer tx.Rollback()

			// Update Item Sale Record
			saleItemRecord := &boiler.ItemSale{
				ID:        auctionItem.ID.String(),
				SoldAt:    null.TimeFrom(time.Now()),
				SoldFor:   decimal.NewNullDecimal(auctionItem.AuctionBidPrice),
				SoldTXID:  null.StringFrom(txid),
				SoldTo:    null.StringFrom(auctionItem.AuctionBidUserID.String()),
				UpdatedAt: time.Now(),
			}
			_, err = saleItemRecord.Update(tx, boil.Whitelist(
				boiler.ItemSaleColumns.SoldAt,
				boiler.ItemSaleColumns.SoldFor,
				boiler.ItemSaleColumns.SoldTXID,
				boiler.ItemSaleColumns.SoldTo,
				boiler.ItemSaleColumns.UpdatedAt,
			))
			if err != nil {
				m.Passport.RefundSupsMessage(txid)
				err = fmt.Errorf("failed to complete payment transaction")
				l.Error().
					Err(err).
					Msg("Failed to process transaction for Purchase Sale Item.")
				return
			}

			rpcAssetTransferRollback, err := TransferAssetsToXsyn(gamedb.StdConn, m.Passport, auctionItem.OwnerID.String(), auctionItem.AuctionBidUserID.String(), txid, auctionItem.Hash, auctionItem.ID.String())
			if err != nil {
				m.Passport.RefundSupsMessage(txid)
				l.Error().
					Err(err).
					Msg("Failed to process transaction for Purchase Sale Item m.Passport.TransferAsset.")
				return
			}

			err = HandleMarketplaceAssetTransfer(tx, m.Passport, auctionItem.ID.String())
			if err != nil {
				m.Passport.RefundSupsMessage(txid)
				rpcAssetTransferRollback()
				l.Error().Err(err).Msg("Failed to transfer item to new owner")
				return
			}

			// Unlock Listed Item
			collectionItem := boiler.CollectionItem{
				ID:                  auctionItem.CollectionItemID.String(),
				LockedToMarketplace: false,
			}

			_, err = collectionItem.Update(tx, boil.Whitelist(
				boiler.CollectionItemColumns.ID,
				boiler.CollectionItemColumns.LockedToMarketplace,
			))
			if err != nil {
				m.Passport.RefundSupsMessage(txid)
				rpcAssetTransferRollback()
				l.Error().Err(err).Msg("Failed to unlock marketplace listed collection item.")
				return
			}

			// Commit Transaction
			err = tx.Commit()
			if err != nil {
				m.Passport.RefundSupsMessage(txid)
				rpcAssetTransferRollback()
				l.Error().Err(err).Msg("Failed to commit db transaction")
				return
			}

			// Log Event
			err = db.MarketplaceAddEvent(boiler.MarketplaceEventPurchase, auctionItem.AuctionBidUserID.String(), decimal.NewNullDecimal(auctionItem.AuctionBidPrice), auctionItem.ID.String(), boiler.TableNames.ItemSales)
			if err != nil {
				l.Error().Err(err).Msg("Failed to log purchase event.")
			}
			err = db.MarketplaceAddEvent(boiler.MarketplaceEventSold, auctionItem.OwnerID.String(), decimal.NewNullDecimal(auctionItem.AuctionBidPrice), auctionItem.ID.String(), boiler.TableNames.ItemSales)
			if err != nil {
				l.Error().Err(err).Msg("Failed to log sold event.")
			}

			numProcessed++
		}()
	}

	gamelog.L.Debug().
		Int("num_processed", numProcessed).
		Int("num_failed", len(auctions)-numProcessed).
		Int("num_pending", len(auctions)).
		Msg("processing completed auction items completed")
}

func HandleMarketplaceAssetTransfer(conn boil.Executor, rpcClient *xsyn_rpcclient.XsynXrpcClient, itemSaleID string) error {
	l := gamelog.L.With().Interface("itemSaleID", itemSaleID).Str("func", "HandleMarketplaceAuctionAssetTransfer").Logger()
	attachedHashes := []string{}

	itemSale, err := boiler.FindItemSale(conn, itemSaleID)
	if err != nil {
		l.Error().Err(err).Msg("failed to find item sale")
		return err
	}
	colItem, err := boiler.FindCollectionItem(conn, itemSale.CollectionItemID)
	if err != nil {
		l.Error().Err(err).Msg("failed to find collection Item")
		return err
	}

	switch colItem.ItemType {
	case boiler.ItemTypeWeapon:
		attachedColItems, err := asset.TransferWeaponToNewOwner(conn, colItem.ItemID, itemSale.SoldTo.String, colItem.XsynLocked, null.NewString("", false))
		if err != nil {
			l.Error().Err(err).Msg("failed to transfer mech to new owner")
			return err
		}
		for _, colItem := range attachedColItems {
			attachedHashes = append(attachedHashes, colItem.Hash)
		}
	case boiler.ItemTypeMech:
		attachedColItems, err := asset.TransferMechToNewOwner(conn, colItem.ItemID, itemSale.SoldTo.String, colItem.XsynLocked, null.NewString("", false))
		if err != nil {
			l.Error().Err(err).Msg("failed to transfer mech to new owner")
			return err
		}
		for _, colItem := range attachedColItems {
			attachedHashes = append(attachedHashes, colItem.Hash)
		}
	case boiler.ItemTypeUtility,
		boiler.ItemTypeMechSkin,
		boiler.ItemTypeMechAnimation,
		boiler.ItemTypePowerCore,
		boiler.ItemTypeMysteryCrate,
		boiler.ItemTypeWeaponSkin:
		colItem.OwnerID = itemSale.SoldTo.String
		_, err = colItem.Update(conn, boil.Infer())
		if err != nil {
			l.Error().Err(err).Msg("failed to transfer mech to new owner")
			return err
		}
	default:
		return fmt.Errorf("unhandled item type")
	}
	for _, hash := range attachedHashes {
		err := rpcClient.TransferAsset(
			colItem.OwnerID,
			itemSale.SoldTo.String,
			hash,
			itemSale.SoldTXID,
			func(rpcClient *xsyn_rpcclient.XsynXrpcClient, eventID int64) {
				asset.UpdateLatestHandledTransferEvent(rpcClient, eventID)
			},
		)
		if err != nil && strings.Contains(err.Error(), "asset not exist") {
			l.Warn().Err(err).Msg("failed to transfer attached asset on xsyn TransferWeaponToNewOwner")
		} else if err != nil {
			l.Error().Err(err).Msg("failed to transfer attached asset on xsyn TransferWeaponToNewOwner")
			return err
		}
	}

	return nil
}
