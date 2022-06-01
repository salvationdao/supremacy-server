package marketplace

import (
	"fmt"
	"server/api"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AuctionController struct {
	API *api.API
}

type ItemSaleAuction struct {
	boiler.ItemSale  `boil:",bind"`
	AuctionBidUserID string `boil:"auction_bid_user_id"`
	AuctionBidTxID   string `boil:"auction_bid_tx_id"`
	AuctionBidPrice  string `boil:"auction_bid_price"`
}

func NewAuctionController() *AuctionController {
	a := &AuctionController{}
	return a
}

func (a *AuctionController) Run() {
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
						item_sales.
						item_sales_bid_history.bid_price AS auction_bid_price,
						item_sales_bid_history.bid_tx_id AS auction_bid_tx_id,
						item_sales_bid_history.bidder_id AS auction_bid_user_id
					FROM item_sales 
						INNER JOIN item_sales_bid_history ON item_sales_bid_history.item_sale_id = item_sales.id
							AND item_sales_bid_history.cancelled_at IS NULL
							AND item_sales_bid_history.refund_bid_tx_id IS NOT NULL
					WHERE item_sales.auction = TRUE
						AND item_sales.sold_by IS NOT NULL
						AND item_sales.end_at <= NOW()
						AND item_sales_bid_history.bid_price >= item_sales.auction_reserved_price
				`),
			).Bind(nil, gamedb.StdConn, &completedAuctions)
			if err != nil {
				gamelog.L.Error().
					Str("db func", "itemSales").
					Err(err).Msg("unable to retrieve completed auctions on marketplace")
			} else {
				gamelog.L.Info().Int("num_pending", len(completedAuctions)).Msg("# completed auctions pending")
			}

			for _, auctionItem := range completedAuctions {
				itemSaleID, err := uuid.FromString(auctionItem.ID)
				if err != nil {
					gamelog.L.Error().
						Str("item_sale_id", auctionItem.ID).
						Err(err).
						Msg("Failed to parse Item Sale ID.")
					continue
				}

				// Update Item Sale Record
				saleItemRecord := auctionItem.ItemSale
				saleItemRecord.SoldAt = null.TimeFrom(time.Now())
				saleItemRecord.SoldFor = null.StringFrom(auctionItem.AuctionBidPrice)
				saleItemRecord.SoldTXID = null.StringFrom(auctionItem.AuctionBidTxID)
				saleItemRecord.SoldBy = null.StringFrom(auctionItem.AuctionBidUserID)

				_, err = saleItemRecord.Update(gamedb.StdConn, boil.Infer())
				if err != nil {
					err = fmt.Errorf("failed to complete payment transaction")
					gamelog.L.Error().
						Str("item_id", auctionItem.ID).
						Str("user_id", auctionItem.AuctionBidUserID).
						Str("cost", auctionItem.AuctionBidPrice).
						Err(err).
						Msg("Failed to process transaction for Purchase Sale Item.")
					continue
				}

				// Transfer ownership of asset
				err = db.ChangeMechOwner(itemSaleID)
				if err != nil {
					gamelog.L.Error().
						Str("item_id", auctionItem.ID).
						Str("user_id", auctionItem.AuctionBidUserID).
						Str("cost", auctionItem.AuctionBidPrice).
						Err(err).
						Msg("Failed to Transfer Mech to New Owner")
					continue
				}
			}

			gamelog.L.Info().Msg("processing completed auction items completed")
		}
	}
}
