package player_abilities

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"sync"
	"time"

	"github.com/ninja-syndicate/ws"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/atomic"
)

type Purchase struct {
	PlayerID  uuid.UUID
	AbilityID uuid.UUID // sale ability id
}

type SaleAbilityPriceResponse struct {
	ID           string `json:"id"`
	CurrentPrice string `json:"current_price"`
}

type SaleAbilityAmountResponse struct {
	ID         string `json:"id"`
	AmountSold int    `json:"amount_sold"`
}

// Used for sale abilities
type SalePlayerAbilitiesSystem struct {
	// sale player abilities
	salePlayerAbilities     map[uuid.UUID]*boiler.SalePlayerAbility // map[ability_id]*Ability
	salePlayerAbilitiesPool []*boiler.SalePlayerAbility
	userPurchaseLimits      map[uuid.UUID]int // map[player_id]purchase count for the current sale period
	nextRefresh             time.Time         // timestamp of when the next sale period will begin

	// KVs
	UserPurchaseLimit          int
	PriceTickerIntervalSeconds int
	TimeBetweenRefreshSeconds  int
	ReductionPercentage        decimal.Decimal
	InflationPercentage        decimal.Decimal
	FloorPrice                 decimal.Decimal
	Limit                      int

	// on sale ability purchase
	Purchase chan *Purchase

	closed *atomic.Bool
	sync.RWMutex
}

func NewSalePlayerAbilitiesSystem() *SalePlayerAbilitiesSystem {
	saleAbilities, err := boiler.SalePlayerAbilities(
		boiler.SalePlayerAbilityWhere.AvailableUntil.GT(null.TimeFrom(time.Now())),
		boiler.SalePlayerAbilityWhere.RarityWeight.GT(0),
		boiler.SalePlayerAbilityWhere.DeletedAt.IsNull(),
		qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to populate salePlayerAbilities map with existing abilities from db")
	}
	salePlayerAbilities := map[uuid.UUID]*boiler.SalePlayerAbility{}
	for _, s := range saleAbilities {
		sID := uuid.FromStringOrNil(s.ID)
		salePlayerAbilities[sID] = s
	}

	saAvailable, err := boiler.SalePlayerAbilities(
		boiler.SalePlayerAbilityWhere.RarityWeight.GT(0),
		boiler.SalePlayerAbilityWhere.DeletedAt.IsNull(),
		qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to initialise pool of sale abilities from db")
	}
	saPool := []*boiler.SalePlayerAbility{}
	for _, sa := range saAvailable {
		for i := 0; i < sa.RarityWeight; i++ {
			saPool = append(saPool, sa)
		}
	}
	gamelog.L.Debug().Msg(fmt.Sprintf("initialised pool of sale abilities with %d entries", len(saPool)))

	timeBetweenRefreshSeconds := db.GetIntWithDefault(db.KeySaleAbilityTimeBetweenRefreshSeconds, 600) // default 10 minutes (600 seconds)
	pas := &SalePlayerAbilitiesSystem{
		salePlayerAbilities:        salePlayerAbilities,
		salePlayerAbilitiesPool:    saPool,
		userPurchaseLimits:         make(map[uuid.UUID]int),
		nextRefresh:                time.Now().Add(time.Duration(timeBetweenRefreshSeconds) * time.Second),
		UserPurchaseLimit:          db.GetIntWithDefault(db.KeySaleAbilityPurchaseLimit, 1),              // default 1 purchase per user per ability
		PriceTickerIntervalSeconds: db.GetIntWithDefault(db.KeySaleAbilityPriceTickerIntervalSeconds, 5), // default 5 seconds
		TimeBetweenRefreshSeconds:  timeBetweenRefreshSeconds,
		ReductionPercentage:        db.GetDecimalWithDefault(db.KeySaleAbilityReductionPercentage, decimal.NewFromFloat(1.0)),  // default 1%
		InflationPercentage:        db.GetDecimalWithDefault(db.KeySaleAbilityInflationPercentage, decimal.NewFromFloat(20.0)), // default 20%
		FloorPrice:                 db.GetDecimalWithDefault(db.KeySaleAbilityFloorPrice, decimal.New(10, 18)),                 // default 10 sups
		Limit:                      db.GetIntWithDefault(db.KeySaleAbilityLimit, 3),                                            // default 3
		Purchase:                   make(chan *Purchase),
		closed:                     atomic.NewBool(false),
	}

	go pas.SalePlayerAbilitiesUpdater()

	return pas
}

func (pas *SalePlayerAbilitiesSystem) NextRefresh() time.Time {
	pas.RLock()
	defer pas.RUnlock()

	return pas.nextRefresh
}

func (pas *SalePlayerAbilitiesSystem) ResetUserPurchaseCounts() {
	pas.Lock()
	defer pas.Unlock()

	// Reset map
	pas.userPurchaseLimits = make(map[uuid.UUID]int)

	// Update sale period
	pas.nextRefresh = time.Now().Add(time.Duration(pas.TimeBetweenRefreshSeconds) * time.Second)
}

func (pas *SalePlayerAbilitiesSystem) CanUserPurchase(userID uuid.UUID) bool {
	pas.RLock()
	defer pas.RUnlock()

	count, ok := pas.userPurchaseLimits[userID]
	if !ok {
		return true
	}

	return count < pas.UserPurchaseLimit
}

func (pas *SalePlayerAbilitiesSystem) AddToUserPurchaseCount(userID uuid.UUID) error {
	pas.Lock()
	defer pas.Unlock()

	count, ok := pas.userPurchaseLimits[userID]
	if !ok {
		count = 0
	}

	if count == pas.UserPurchaseLimit {
		minutes := int(time.Until(pas.nextRefresh).Minutes())
		msg := fmt.Sprintf("Please try again in %d minutes.", minutes)
		if minutes < 1 {
			msg = fmt.Sprintf("Please try again in %d seconds.", int(time.Until(pas.nextRefresh).Seconds()))
		}
		return fmt.Errorf("You have hit your purchase limit of %d during this sale period. %s", pas.UserPurchaseLimit, msg)
	}

	pas.userPurchaseLimits[userID] = count + 1

	return nil
}

func (pas *SalePlayerAbilitiesSystem) SalePlayerAbilitiesUpdater() {
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
				saAvailable, err := boiler.SalePlayerAbilities(
					boiler.SalePlayerAbilityWhere.RarityWeight.GT(0),
					boiler.SalePlayerAbilityWhere.DeletedAt.IsNull(),
					qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
				).All(gamedb.StdConn)
				if err != nil {
					gamelog.L.Error().Err(err).Msg("failed to initialise pool of sale abilities from db")
				}
				saPool := []*boiler.SalePlayerAbility{}
				for _, sa := range saAvailable {
					for i := 0; i < sa.RarityWeight; i++ {
						saPool = append(saPool, sa)
					}
				}
				gamelog.L.Debug().Msg(fmt.Sprintf("initialised pool of sale abilities with %d entries", len(saPool)))
			}

			// Price ticker ticks every 5 seconds, updates prices of abilities and refreshes the sale ability list when all abilities on sale have expired
			// Check each ability that is on sale, remove them if expired or if their sale limit has been reached
			for _, s := range pas.salePlayerAbilities {
				if s.AvailableUntil.Time.After(time.Now()) && s.RarityWeight >= 0 {
					continue
				}
				sID := uuid.FromStringOrNil(s.ID)
				delete(pas.salePlayerAbilities, sID)
			}

			if len(pas.salePlayerAbilities) == 0 {
				gamelog.L.Debug().Msg("repopulating sale abilities since there aren't any more")
				// If no abilities are on sale, refill sale abilities
				saleAbilities, err := boiler.SalePlayerAbilities(
					boiler.SalePlayerAbilityWhere.AvailableUntil.GT(null.TimeFrom(time.Now())),
					boiler.SalePlayerAbilityWhere.DeletedAt.IsNull(),
					qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
				).All(gamedb.StdConn)
				if errors.Is(err, sql.ErrNoRows) || len(saleAbilities) == 0 {
					gamelog.L.Debug().Msg("refreshing sale abilities in db")
					// If no sale abilities, get 3 random sale abilities and update their time to an hour from now
					// Find 3 random weighted abilities
					selected := map[string]*boiler.SalePlayerAbility{}
					rand.Seed(time.Now().Unix())
					for {
						s := pas.salePlayerAbilitiesPool[rand.Intn(len(pas.salePlayerAbilitiesPool))]

						_, ok := selected[s.ID]
						if ok {
							// Is duplicate
							continue
						}
						selected[s.ID] = s
						saleAbilities = append(saleAbilities, s)

						if len(selected) == pas.Limit {
							break
						}
					}

					if len(selected) == 0 {
						gamelog.L.Warn().Msg("no sale abilities could be found")
						break
					}

					tenMinutesFromNow := time.Now().Add(time.Duration(pas.TimeBetweenRefreshSeconds) * time.Second)
					_, err = saleAbilities.UpdateAll(gamedb.StdConn, boiler.M{
						"available_until": tenMinutesFromNow,
						"amount_sold":     0,
					})
					if err != nil {
						gamelog.L.Error().Err(err).Msg("failed to update sale ability with new expiration date")
						continue
					}

					detailedSaleAbilities := []*db.SaleAbilityDetailed{}
					for _, s := range saleAbilities {
						detailedSaleAbilities = append(detailedSaleAbilities, &db.SaleAbilityDetailed{
							SalePlayerAbility: s,
							Ability:           s.R.Blueprint,
						})
					}

					// Reset user purchase counts
					pas.ResetUserPurchaseCounts()

					// Broadcast trigger of sale abilities list update
					ws.PublishMessage("/public/sale_abilities", server.HubKeySaleAbilitiesList, struct {
						NextRefreshTime              *time.Time                `json:"next_refresh_time"`
						RefreshPeriodDurationSeconds int                       `json:"refresh_period_duration_seconds"`
						SaleAbilities                []*db.SaleAbilityDetailed `json:"sale_abilities,omitempty"`
					}{
						NextRefreshTime:              &pas.nextRefresh,
						RefreshPeriodDurationSeconds: pas.TimeBetweenRefreshSeconds,
						SaleAbilities:                detailedSaleAbilities,
					})
				} else if err != nil {
					gamelog.L.Error().Err(err).Msg("failed to fill sale player abilities map with new sale abilities")
					break
				}
				for _, s := range saleAbilities {
					sID := uuid.FromStringOrNil(s.ID)
					pas.salePlayerAbilities[sID] = s
				}
			}

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

				// Broadcast updated sale ability
				ws.PublishMessage("/public/sale_abilities", server.HubKeySaleAbilitiesPriceSubscribe, SaleAbilityPriceResponse{
					ID:           s.ID,
					CurrentPrice: s.CurrentPrice.StringFixed(0),
				})
			}
			break
		case purchase := <-pas.Purchase:
			if saleAbility, ok := pas.salePlayerAbilities[purchase.AbilityID]; ok {
				saleAbility.CurrentPrice = saleAbility.CurrentPrice.Mul(oneHundred.Add(pas.InflationPercentage).Div(oneHundred))
				saleAbility.AmountSold = saleAbility.AmountSold + 1
				_, err := saleAbility.Update(gamedb.StdConn, boil.Whitelist(
					boiler.SalePlayerAbilityColumns.CurrentPrice,
					boiler.SalePlayerAbilityColumns.AmountSold,
				))
				if err != nil {
					gamelog.L.Error().Err(err).Str("salePlayerAbilityID", saleAbility.ID).Str("new price", saleAbility.CurrentPrice.String()).Interface("sale ability", saleAbility).Msg("failed to update sale ability price")
					break
				}

				// Broadcast updated sale ability price
				ws.PublishMessage("/public/sale_abilities", server.HubKeySaleAbilitiesPriceSubscribe, SaleAbilityPriceResponse{
					ID:           saleAbility.ID,
					CurrentPrice: saleAbility.CurrentPrice.StringFixed(0),
				})

				// Broadcast updated sale ability sold amount
				ws.PublishMessage("/public/sale_abilities", server.HubKeySaleAbilitiesAmountSubscribe, SaleAbilityAmountResponse{
					ID:         saleAbility.ID,
					AmountSold: saleAbility.AmountSold,
				})
			}
			break
		}
	}
}
