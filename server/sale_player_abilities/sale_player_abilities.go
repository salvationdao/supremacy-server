package sale_player_abilities

import (
	"fmt"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/sasha-s/go-deadlock"

	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/atomic"
)

type Claim struct {
	SaleID string // sale ability id
}

type Purchase struct {
	SaleID string // sale ability id
}

type SaleAbilityPriceResponse struct {
	ID           string `json:"id"`
	CurrentPrice string `json:"current_price"`
}

// Used for sale abilities
type SalePlayerAbilityManager struct {
	// sale player abilities
	salePlayerAbilities          map[string]*boiler.SalePlayerAbility // map[sale_id]*Ability
	salePlayerAbilitiesWithDupes []*db.SaleAbilityDetailed
	salePlayerAbilitiesPool      []*boiler.SalePlayerAbility
	totalSaleAbilities           int
	nextRefresh                  RefreshTime    // timestamp of when the next sale period will begin
	userPurchaseLimits           map[string]int // map[player_id]purchase count for the current sale period

	// KVs
	UserPurchaseLimit          int
	PriceTickerIntervalSeconds int
	TimeBetweenRefreshSeconds  int
	ReductionPercentage        decimal.Decimal
	InflationPercentage        decimal.Decimal
	FloorPrice                 decimal.Decimal
	DisplayLimit               int

	// on sale ability or purchase
	Purchase chan *Purchase

	closed *atomic.Bool
	deadlock.RWMutex
}

type RefreshTime struct {
	Server time.Time
	Client time.Time
}

func NewSalePlayerAbilitiesSystem() *SalePlayerAbilityManager {
	priceTickerIntervalSeconds := db.GetIntWithDefault(db.KeySaleAbilityPriceTickerIntervalSeconds, 5)
	pas := &SalePlayerAbilityManager{
		salePlayerAbilities:          map[string]*boiler.SalePlayerAbility{},
		salePlayerAbilitiesWithDupes: []*db.SaleAbilityDetailed{},
		salePlayerAbilitiesPool:      []*boiler.SalePlayerAbility{},
		totalSaleAbilities:           0,
		userPurchaseLimits:           make(map[string]int),
		nextRefresh: RefreshTime{
			Server: time.Now(),
			Client: time.Now().Add(time.Duration(priceTickerIntervalSeconds+1) * time.Second),
		},
		UserPurchaseLimit:          db.GetIntWithDefault(db.KeySaleAbilityPurchaseLimit, 1),                                    // default 1 purchase per user per sale period
		PriceTickerIntervalSeconds: priceTickerIntervalSeconds,                                                                 // default 5 seconds
		TimeBetweenRefreshSeconds:  db.GetIntWithDefault(db.KeySaleAbilityTimeBetweenRefreshSeconds, 600),                      // default 10 minutes (600 seconds),
		ReductionPercentage:        db.GetDecimalWithDefault(db.KeySaleAbilityReductionPercentage, decimal.NewFromFloat(1.0)),  // default 1%
		InflationPercentage:        db.GetDecimalWithDefault(db.KeySaleAbilityInflationPercentage, decimal.NewFromFloat(20.0)), // default 20%
		FloorPrice:                 db.GetDecimalWithDefault(db.KeySaleAbilityFloorPrice, decimal.New(10, 18)),                 // default 10 sups
		DisplayLimit:               db.GetIntWithDefault(db.KeySaleAbilityLimit, 3),                                            // default 3 abilities displayed per sale period
		Purchase:                   make(chan *Purchase),
		closed:                     atomic.NewBool(false),
	}

	pas.RehydratePool()

	go pas.SalePlayerAbilitiesUpdater()

	return pas
}

func (pas *SalePlayerAbilityManager) CurrentSaleList() []*db.SaleAbilityDetailed {
	pas.RLock()
	defer pas.RUnlock()

	return pas.salePlayerAbilitiesWithDupes
}

func (pas *SalePlayerAbilityManager) RehydratePool() {
	pas.Lock()
	defer pas.Unlock()

	saAvailable, err := boiler.SalePlayerAbilities(
		boiler.SalePlayerAbilityWhere.RarityWeight.GT(0),
		boiler.SalePlayerAbilityWhere.DeletedAt.IsNull(),
		qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to refresh pool of sale abilities from db")
	}
	saPool := []*boiler.SalePlayerAbility{}
	for _, sa := range saAvailable {
		for i := 0; i < sa.RarityWeight; i++ {
			saPool = append(saPool, sa)
		}
	}

	pas.totalSaleAbilities = len(saAvailable)
	pas.salePlayerAbilitiesPool = saPool

	gamelog.L.Debug().Msg(fmt.Sprintf("refreshed pool of sale abilities with %d entries", len(saPool)))
}

func (pas *SalePlayerAbilityManager) NextRefresh() RefreshTime {
	pas.RLock()
	defer pas.RUnlock()

	return pas.nextRefresh
}

func (pas *SalePlayerAbilityManager) NextRefreshInSeconds() int {
	pas.RLock()
	defer pas.RUnlock()

	return int(pas.nextRefresh.Client.Sub(time.Now()).Seconds())
}

func (pas *SalePlayerAbilityManager) Refresh() {
	pas.Lock()
	defer pas.Unlock()

	// Reset and purchase limits
	pas.userPurchaseLimits = make(map[string]int)

	// Update sale period
	pas.nextRefresh = RefreshTime{
		Server: time.Now().Add(time.Duration(pas.TimeBetweenRefreshSeconds) * time.Second),
		Client: time.Now().Add(time.Duration(pas.TimeBetweenRefreshSeconds+pas.PriceTickerIntervalSeconds+1) * time.Second),
	}
}

func (pas *SalePlayerAbilityManager) IsAbilityAvailable(saleID string) bool {
	pas.RLock()
	defer pas.RUnlock()
	_, ok := pas.salePlayerAbilities[saleID]

	return ok
}

func (pas *SalePlayerAbilityManager) CanUserPurchase(userID string) bool {
	pas.RLock()
	defer pas.RUnlock()

	count, ok := pas.userPurchaseLimits[userID]
	if !ok {
		return true
	}

	return count < pas.UserPurchaseLimit
}

func (pas *SalePlayerAbilityManager) AddToUserPurchaseCount(userID string) error {
	pas.Lock()
	defer pas.Unlock()

	count, ok := pas.userPurchaseLimits[userID]
	if !ok {
		count = 0
	}

	if count == pas.UserPurchaseLimit {
		minutes := int(time.Until(pas.nextRefresh.Client).Minutes())
		msg := fmt.Sprintf("Please try again in %d minutes.", minutes)
		if minutes < 1 {
			msg = fmt.Sprintf("Please try again in %d seconds.", int(time.Until(pas.nextRefresh.Client).Seconds()))
		}
		return fmt.Errorf("You have hit your purchase limit of %d during this sale period. %s", pas.UserPurchaseLimit, msg)
	}

	pas.userPurchaseLimits[userID] = count + 1

	return nil
}

func (pas *SalePlayerAbilityManager) SalePlayerAbilitiesUpdater() {
	priceTicker := time.NewTicker(time.Duration(pas.PriceTickerIntervalSeconds) * time.Second)

	defer func() {
		priceTicker.Stop()
		pas.closed.Store(true)
	}()

	oneHundred := decimal.NewFromFloat(100.0)
	for {
		select {
		case <-priceTicker.C:
			if len(pas.salePlayerAbilitiesPool) == 0 {
				gamelog.L.Debug().Msg("populating sale player abilities pool because it was empty")
				pas.RehydratePool()
			}

			// Update prices of abilities and refresh the sale ability list when period has ended
			if time.Now().After(pas.NextRefresh().Server) {
				gamelog.L.Debug().Msg("refreshing sale abilities in db")
				// If no sale abilities, get 3 random sale abilities and update their time to an hour from now
				// Find 3 random weighted abilities
				selected := map[string]struct{}{}
				rand.Seed(time.Now().Unix())
				attempts := 0
				saleAbilities := boiler.SalePlayerAbilitySlice{}
				for {
					attempts++
					s := pas.salePlayerAbilitiesPool[rand.Intn(len(pas.salePlayerAbilitiesPool))]

					_, ok := selected[s.ID]
					if ok {
						// Is duplicate
						if pas.DisplayLimit <= pas.totalSaleAbilities || len(saleAbilities) < pas.totalSaleAbilities {
							continue
						}
					}
					selected[s.ID] = struct{}{}
					saleAbilities = append(saleAbilities, s)

					if len(saleAbilities) == pas.DisplayLimit {
						break
					}
				}

				if len(saleAbilities) == 0 {
					gamelog.L.Warn().Msg("no sale abilities could be found")
					break
				}

				// Reset user purchase counts
				pas.Refresh()

				detailedSalePlayerAbilities := []*db.SaleAbilityDetailed{}
				for _, sa := range saleAbilities {
					detailedSalePlayerAbilities = append(detailedSalePlayerAbilities, &db.SaleAbilityDetailed{
						SalePlayerAbility: sa,
						Ability:           sa.R.Blueprint,
					})
				}

				// Broadcast trigger of sale abilities list update
				nextRefresh := pas.NextRefresh().Client
				ws.PublishMessage("/secure/sale_abilities", server.HubKeySaleAbilitiesListSubscribe, struct {
					NextRefreshTime *time.Time                `json:"next_refresh_time"`
					TimeLeftSeconds int                       `json:"time_left_seconds"`
					SaleAbilities   []*db.SaleAbilityDetailed `json:"sale_abilities"`
				}{
					NextRefreshTime: &nextRefresh,
					TimeLeftSeconds: pas.NextRefreshInSeconds(),
					SaleAbilities:   detailedSalePlayerAbilities,
				})

				pas.salePlayerAbilitiesWithDupes = []*db.SaleAbilityDetailed{}
				for _, s := range detailedSalePlayerAbilities {
					pas.salePlayerAbilities[s.ID] = s.SalePlayerAbility
					pas.salePlayerAbilitiesWithDupes = append(pas.salePlayerAbilitiesWithDupes, s)
				}
			}

			updatedPrices := []*SaleAbilityPriceResponse{}
			for _, s := range pas.salePlayerAbilities {
				s.CurrentPrice = s.CurrentPrice.Mul(oneHundred.Sub(pas.ReductionPercentage).Div(oneHundred))
				if s.CurrentPrice.LessThan(pas.FloorPrice) {
					s.CurrentPrice = pas.FloorPrice
				}

				_, err := s.Update(gamedb.StdConn, boil.Whitelist(boiler.SalePlayerAbilityColumns.CurrentPrice))
				if err != nil {
					gamelog.L.Error().Err(err).Str("salePlayerAbilityID", s.ID).Str("new price", s.CurrentPrice.String()).Interface("sale ability", s).Msg("failed to update sale ability price")
					continue
				}

				updatedPrices = append(updatedPrices, &SaleAbilityPriceResponse{
					ID:           s.ID,
					CurrentPrice: s.CurrentPrice.StringFixed(0),
				})
			}

			// Broadcast updated sale ability prices
			ws.PublishMessage("/secure/sale_abilities", server.HubKeySaleAbilitiesPriceSubscribe, updatedPrices)
		case purchase := <-pas.Purchase:
			if saleAbility, ok := pas.salePlayerAbilities[purchase.SaleID]; ok {
				saleAbility.CurrentPrice = saleAbility.CurrentPrice.Mul(oneHundred.Add(pas.InflationPercentage).Div(oneHundred))
				saleAbility.AmountSold = saleAbility.AmountSold + 1
				_, err := saleAbility.Update(gamedb.StdConn, boil.Whitelist(
					boiler.SalePlayerAbilityColumns.CurrentPrice,
					boiler.SalePlayerAbilityColumns.AmountSold,
				))
				if err != nil {
					gamelog.L.Error().Err(err).Str("salePlayerAbilityID", saleAbility.ID).Interface("sale ability", saleAbility).Msg("failed to update sale ability amount sold and current price")
					break
				}

				updatedPrices := []*SaleAbilityPriceResponse{}
				for _, s := range pas.salePlayerAbilities {
					updatedPrices = append(updatedPrices, &SaleAbilityPriceResponse{
						ID:           s.ID,
						CurrentPrice: s.CurrentPrice.StringFixed(0),
					})
				}

				// Broadcast updated sale ability price
				ws.PublishMessage("/secure/sale_abilities", server.HubKeySaleAbilitiesPriceSubscribe, updatedPrices)
			}
		}
	}
}
