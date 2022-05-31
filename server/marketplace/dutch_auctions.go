package marketplace

import (
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

type DutchAuctionManager struct {
	sync.RWMutex
	auctions map[string]*DutchAuctionItem
}

func NewDutchAuctionManager() *DutchAuctionManager {
	a := &DutchAuctionManager{
		auctions: make(map[string]*DutchAuctionItem),
	}
	return a
}

// Adds a new dutch auction to the manager queue.
func (a *DutchAuctionManager) AddItem(da *DutchAuctionItem) {
	da.AddEventHandler(a.handleClosedAuction)

	a.RLock()
	a.auctions[da.id.String()] = da
	a.RUnlock()
}

// Removes a dutch auction.
func (a *DutchAuctionManager) RemoveItem(da *DutchAuctionItem) {
	a.RLock()
	if _, ok := a.auctions[da.id.String()]; ok {
		delete(a.auctions, da.id.String())
	}
	a.RUnlock()
}

// Handles when an dutch auction ticker has ended.
func (a *DutchAuctionManager) handleClosedAuction(i *DutchAuctionItem) {
	a.RemoveItem(i)
}

/////////////////////
//  Auction Items  //
/////////////////////

type DutchAuctionClosedHandler func(*DutchAuctionItem)

// DutchAuctionItem manages the dutch auction item and updates the buyout price.
type DutchAuctionItem struct {
	sync.RWMutex
	id                   uuid.UUID
	startingPrice        decimal.Decimal
	currentPrice         decimal.Decimal
	reservedPrice        decimal.Decimal
	dropRatePrice        decimal.Decimal
	endAt                time.Time
	ticker               *time.Ticker
	closeAuctionHandlers []DutchAuctionClosedHandler
}

func NewDutchAuctionItem(id uuid.UUID, initialPrice, reservedPrice, dropRatePrice decimal.Decimal, endAt time.Time) *DutchAuctionItem {
	a := &DutchAuctionItem{
		id:                   id,
		startingPrice:        initialPrice,
		currentPrice:         initialPrice,
		dropRatePrice:        dropRatePrice,
		endAt:                endAt,
		closeAuctionHandlers: []DutchAuctionClosedHandler{},
	}
	return a
}

// Runs the dutch auction.
func (a *DutchAuctionItem) Run() {
	a.ticker = time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-a.ticker.C:
			a.currentPrice = decimal.Max(a.currentPrice.Sub(a.dropRatePrice), a.reservedPrice)
			if a.currentPrice.Equal(a.reservedPrice) || time.Now().After(a.endAt) {
				a.Stop()

				// Fire closed auction event
				a.RLock()
				for _, fn := range a.closeAuctionHandlers {
					fn(a)
				}
				a.RUnlock()
			}
		}
	}
}

// Stops the dutch auction tracker.
func (a *DutchAuctionItem) Stop() {
	if a.ticker == nil {
		return
	}
	a.ticker.Stop()
}

// AddClosedHandler attaches a callback when the current auction closes.
func (a *DutchAuctionItem) AddEventHandler(c DutchAuctionClosedHandler) {
	a.RLock()
	a.closeAuctionHandlers = append(a.closeAuctionHandlers, c)
	a.RUnlock()
}
