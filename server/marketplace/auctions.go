package marketplace

import (
	"fmt"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/xsyn_rpcclient"
	"time"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AuctionController struct {
	Passport *xsyn_rpcclient.XsynXrpcClient
}

type ItemSaleAuction struct {
	ID               uuid.UUID `boil:"id"`
	ItemID           uuid.UUID `boil:"item_id"`
	OwnerID          uuid.UUID `boil:"owner_id"`
	AuctionBidPrice  string    `boil:"auction_bid_price"`
	AuctionBidUserID uuid.UUID `boil:"auction_bid_user_id"`
	FactionID        uuid.UUID `boil:"faction_id"`
}

func NewAuctionController(pp *xsyn_rpcclient.XsynXrpcClient) *AuctionController {
	a := &AuctionController{pp}
	go a.Run()
	return a
}

func (a *AuctionController) Run() {
	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("panic! panic! panic! Panic at the AuctionController!", r)
		}
	}()

	mainTicker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-mainTicker.C:
			// Scan all completed auctions that haven't went through payment process
			gamelog.L.Info().Msg("processing completed auction items started")

			completedAuctions := []*ItemSaleAuction{}
			err := boiler.NewQuery(
				qm.SQL(`
					SELECT item_sales.id AS id,
						item_sales.item_id,
						item_sales.owner_id,
						item_sales_bid_history.bid_price AS auction_bid_price,
						item_sales_bid_history.bidder_id AS auction_bid_user_id,
						players.faction_id
					FROM item_sales 
						INNER JOIN item_sales_bid_history ON item_sales_bid_history.item_sale_id = item_sales.id
							AND item_sales_bid_history.cancelled_at IS NULL
							AND item_sales_bid_history.refund_bid_tx_id IS NULL
						INNER JOIN players ON players.id = item_sales.owner_id 
					WHERE item_sales.auction = TRUE
						AND item_sales.sold_by IS NULL
						AND item_sales.end_at <= NOW()
						AND item_sales_bid_history.bid_price >= item_sales.auction_reserved_price
				`),
			).Bind(nil, gamedb.StdConn, &completedAuctions)
			if err != nil {
				gamelog.L.Error().
					Str("db func", "itemSales").
					Err(err).Msg("unable to retrieve completed auctions on marketplace")
			}

			numProcessed := 0
			for _, auctionItem := range completedAuctions {
				// Get Faction Account sending bid amount to
				factionAccountID, ok := server.FactionUsers[auctionItem.FactionID.String()]
				if !ok {
					err = fmt.Errorf("failed to get hard coded syndicate player id")
					gamelog.L.Error().
						Str("owner_id", auctionItem.OwnerID.String()).
						Str("faction_id", auctionItem.FactionID.String()).
						Err(err).
						Msg("unable to get hard coded syndicate player ID from faction ID")
					continue
				}

				// Transfer Sups to Owner
				// TODO: Deal with sales cut
				txid, err := a.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
					FromUserID:           uuid.Must(uuid.FromString(factionAccountID)),
					ToUserID:             uuid.Must(uuid.FromString(auctionItem.OwnerID.String())),
					Amount:               auctionItem.AuctionBidPrice,
					TransactionReference: server.TransactionReference(fmt.Sprintf("marketplace_buy_auction_item:%s|%d", auctionItem.ID.String(), time.Now().UnixNano())),
					Group:                string(server.TransactionGroupMarketplace),
					SubGroup:             "SUPREMACY",
					Description:          fmt.Sprintf("marketplace buy auction item: %s", auctionItem.ID),
					NotSafe:              true,
				})
				if err != nil {
					gamelog.L.Error().
						Str("item_id", auctionItem.ID.String()).
						Str("user_id", auctionItem.AuctionBidUserID.String()).
						Str("cost", auctionItem.AuctionBidPrice).
						Err(err).
						Msg("Failed to send sups to item seller.")
					continue
				}

				// Begin Transaction
				tx, err := gamedb.StdConn.Begin()
				if err != nil {
					gamelog.L.Error().
						Str("item_id", auctionItem.ID.String()).
						Str("user_id", auctionItem.AuctionBidUserID.String()).
						Str("cost", auctionItem.AuctionBidPrice).
						Err(err).
						Msg("Failed to start db transaction.")
					continue
				}
				defer tx.Rollback()

				// Update Item Sale Record
				saleItemRecord := &boiler.ItemSale{
					ID:       auctionItem.ID.String(),
					SoldAt:   null.TimeFrom(time.Now()),
					SoldFor:  null.StringFrom(auctionItem.AuctionBidPrice),
					SoldTXID: null.StringFrom(txid),
					SoldBy:   null.StringFrom(auctionItem.AuctionBidUserID.String()),
				}
				_, err = saleItemRecord.Update(tx, boil.Whitelist(
					boiler.ItemSaleColumns.SoldAt,
					boiler.ItemSaleColumns.SoldFor,
					boiler.ItemSaleColumns.SoldTXID,
					boiler.ItemSaleColumns.SoldBy,
				))
				if err != nil {
					a.Passport.RefundSupsMessage(txid)
					err = fmt.Errorf("failed to complete payment transaction")
					gamelog.L.Error().
						Str("item_id", auctionItem.ID.String()).
						Str("user_id", auctionItem.AuctionBidUserID.String()).
						Str("cost", auctionItem.AuctionBidPrice).
						Err(err).
						Msg("Failed to process transaction for Purchase Sale Item.")
					continue
				}

				// Transfer ownership of asset
				err = db.ChangeMechOwner(tx, auctionItem.ID)
				if err != nil {
					a.Passport.RefundSupsMessage(txid)
					gamelog.L.Error().
						Str("item_id", auctionItem.ID.String()).
						Str("user_id", auctionItem.AuctionBidUserID.String()).
						Str("cost", auctionItem.AuctionBidPrice).
						Err(err).
						Msg("Failed to Transfer Mech to New Owner")
					continue
				}

				// Commit Transaction
				err = tx.Commit()
				if err != nil {
					a.Passport.RefundSupsMessage(txid)
					gamelog.L.Error().
						Str("item_id", auctionItem.ID.String()).
						Str("user_id", auctionItem.AuctionBidUserID.String()).
						Str("cost", auctionItem.AuctionBidPrice).
						Err(err).
						Msg("Failed to commit db transaction")
					continue
				}

				numProcessed++
			}

			gamelog.L.Info().
				Int("num_processed", numProcessed).
				Int("num_failed", len(completedAuctions)-numProcessed).
				Int("num_pending", len(completedAuctions)).
				Msg("processing completed auction items completed")
		}
	}
}
