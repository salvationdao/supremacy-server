package marketplace

import (
	"fmt"
	"server/api"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AuctionController struct {
	API *api.API
}

type ItemSaleAuction struct {
	ID               string `boil:"id"`
	AuctionBidUserID string `boil:"auction_bid_user_id"`
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
						MAX(item_sales_bid_history.bid_price) AS auction_bid_price,
						item_sales_bid_history.bidder_id AS auction_bid_user_id
					FROM item_sales 
						INNER JOIN item_sales_bid_history ON item_sales_bid_history.item_sale_id = item_sales.id
							AND item_sales_bid_history.cancelled_at is null
					WHERE item_sales.auction = TRUE
						AND item_sales.sold_by IS NOT NULL
						AND item_sales.end_at <= NOW()
					GROUP BY item_sales.id, item_sales_bid_history.bidder_id
					HAVING MAX(item_sales_bid_history.bid_price) >= auction_reserved_price ;
				`),
			).Bind(nil, gamedb.StdConn, &completedAuctions)
			if err != nil {
				gamelog.L.Error().
					Str("db func", "itemSales").
					Err(err).Msg("unable to retrieve completed auctions on marketplace")
			} else {
				gamelog.L.Info().Int("num_pending", len(completedAuctions)).Msg("# completed auctions pending")
			}

			// Take Payment for Auction Items
			for _, auctionItem := range completedAuctions {
				fmt.Println("TODO", auctionItem)
				// bidderUserID, err := uuid.FromString(auctionItem.AuctionBidUserID)
				// if err != nil {
				// 	gamelog.L.Error().
				// 		Str("user_id", auctionItem.AuctionBidUserID).
				// 		Str("item_id", auctionItem.ID).
				// 		Err(err).
				// 		Msg("Unable to get winning auction bid price.")
				// 	continue
				// }

				// saleItemCost, err := decimal.NewFromString(auctionItem.AuctionBidPrice)
				// if err != nil {
				// 	gamelog.L.Error().
				// 		Str("user_id", auctionItem.AuctionBidUserID).
				// 		Str("item_id", auctionItem.ID).
				// 		Err(err).
				// 		Msg("Unable to get winning auction bid price.")
				// 	continue
				// }

				// userBalance
			}

			gamelog.L.Info().Msg("processing completed auction items completed")
		}
	}
}
