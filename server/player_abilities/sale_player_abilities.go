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
	salePlayerAbilities map[uuid.UUID]*boiler.SalePlayerAbility // map[ability_id]*Ability
	userPurchaseLimits  map[uuid.UUID]map[string]int            // map[player_id]map[sale_ability_id]purchase count for the current sale period
	nextSalePeriod      time.Time                               // timestamp of when the next sale period will begin

	// KVs
	UserPurchaseLimit               int
	PriceTickerIntervalSeconds      int
	SalePeriodTickerIntervalSeconds int
	TimeBetweenRefreshSeconds       int
	ReductionPercentage             decimal.Decimal
	InflationPercentage             decimal.Decimal
	FloorPrice                      decimal.Decimal
	Limit                           int

	// on sale ability purchase
	Purchase chan *Purchase

	closed *atomic.Bool
	sync.RWMutex
}

func NewSalePlayerAbilitiesSystem() *SalePlayerAbilitiesSystem {
	saleAbilities, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.AvailableUntil.GT(null.TimeFrom(time.Now()))).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to populate salePlayerAbilities map with existing abilities from db")
	}
	salePlayerAbilities := map[uuid.UUID]*boiler.SalePlayerAbility{}
	for _, s := range saleAbilities {
		sID := uuid.FromStringOrNil(s.ID)
		salePlayerAbilities[sID] = s
	}

	pas := &SalePlayerAbilitiesSystem{
		salePlayerAbilities:             salePlayerAbilities,
		userPurchaseLimits:              make(map[uuid.UUID]map[string]int),
		nextSalePeriod:                  time.Now(),
		UserPurchaseLimit:               db.GetIntWithDefault(db.KeySaleAbilityPurchaseLimit, 1),                                 // default 1 purchase per user per ability
		PriceTickerIntervalSeconds:      db.GetIntWithDefault(db.SaleAbilityPriceTickerIntervalSeconds, 5),                       // default 5 seconds
		SalePeriodTickerIntervalSeconds: db.GetIntWithDefault(db.SaleAbilitySalePeriodTickerIntervalSeconds, 600),                // default 10 minutes (600 seconds)
		TimeBetweenRefreshSeconds:       db.GetIntWithDefault(db.SaleAbilityTimeBetweenRefreshSeconds, 3600),                     // default 1 hour (3600 seconds)
		ReductionPercentage:             db.GetDecimalWithDefault(db.SaleAbilityReductionPercentage, decimal.NewFromFloat(1.0)),  // default 1%
		InflationPercentage:             db.GetDecimalWithDefault(db.SaleAbilityInflationPercentage, decimal.NewFromFloat(20.0)), // default 20%
		FloorPrice:                      db.GetDecimalWithDefault(db.SaleAbilityFloorPrice, decimal.New(10, 18)),                 // default 10 sups
		Limit:                           db.GetIntWithDefault(db.SaleAbilityLimit, 3),                                            // default 3
		Purchase:                        make(chan *Purchase),
		closed:                          atomic.NewBool(false),
	}

	go pas.SalePlayerAbilitiesUpdater()

	return pas
}

func (pas *SalePlayerAbilitiesSystem) NextSalePeriod() time.Time {
	pas.RLock()
	defer pas.RUnlock()

	return pas.nextSalePeriod
}

func (pas *SalePlayerAbilitiesSystem) ResetUserPurchaseCounts() {
	pas.Lock()
	defer pas.Unlock()

	// Reset map
	pas.userPurchaseLimits = make(map[uuid.UUID]map[string]int)

	// Update sale period
	pas.nextSalePeriod = time.Now().Add(time.Duration(pas.SalePeriodTickerIntervalSeconds) * time.Second)
}

func (pas *SalePlayerAbilitiesSystem) AddToUserPurchaseCount(userID uuid.UUID, saleAbilityID string) error {
	pas.Lock()
	defer pas.Unlock()

	abilitiesMap, ok := pas.userPurchaseLimits[userID]
	if !ok {
		abilitiesMap = map[string]int{}
		pas.userPurchaseLimits[userID] = abilitiesMap
	}

	count, ok := abilitiesMap[saleAbilityID]
	if !ok {
		abilitiesMap[saleAbilityID] = 0
	} else if count == pas.UserPurchaseLimit {
		_, minutes, _ := pas.nextSalePeriod.Clock()
		return fmt.Errorf("User has hit their purchase limit of %d for this ability. Please try again in %d minutes", pas.UserPurchaseLimit, minutes)
	}

	abilitiesMap[saleAbilityID] = count + 1
	pas.userPurchaseLimits[userID] = abilitiesMap

	return nil
}

func (pas *SalePlayerAbilitiesSystem) SalePlayerAbilitiesUpdater() {
	priceTicker := time.NewTicker(time.Duration(pas.PriceTickerIntervalSeconds) * time.Second)
	salePeriodTicker := time.NewTicker(time.Duration(pas.SalePeriodTickerIntervalSeconds) * time.Second)

	defer func() {
		priceTicker.Stop()
		pas.closed.Store(true)
	}()

	oneHundred := decimal.NewFromFloat(100.0)
	for {
		select {
		case <-priceTicker.C:
			// Price ticker ticks every 5 seconds, updates prices of abilities and refreshes the sale ability list when all abilities on sale have expired
			// Check each ability that is on sale, remove them if expired or if their sale limit has been reached
			for _, s := range pas.salePlayerAbilities {
				if s.AvailableUntil.Time.After(time.Now()) {
					continue
				}
				sID := uuid.FromStringOrNil(s.ID)
				delete(pas.salePlayerAbilities, sID)
			}

			if len(pas.salePlayerAbilities) < 1 {
				gamelog.L.Debug().Msg("repopulating sale abilities since there aren't any more")
				// If no abilities are on sale, refill sale abilities
				saleAbilities, err := boiler.SalePlayerAbilities(
					boiler.SalePlayerAbilityWhere.AvailableUntil.GT(null.TimeFrom(time.Now())),
					qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
				).All(gamedb.StdConn)
				if errors.Is(err, sql.ErrNoRows) || len(saleAbilities) == 0 {
					gamelog.L.Debug().Msg("refreshing sale abilities in db")
					// If no sale abilities, get 3 random sale abilities and update their time to an hour from now
					q := fmt.Sprintf(
						`
						with cte as (
							select random() * (
								select sum(rarity_weight)
								from sale_player_abilities
							) R
						)
						select 
							Q.id,
							Q.blueprint_id,
							Q.current_price,
							Q.available_until,
							Q.amount_sold,
							Q.sale_limit
						from (
							select id, blueprint_id, current_price, available_until, amount_sold, sale_limit, sum(rarity_weight) over (order by id) S, R
							from sale_player_abilities spa
							cross join cte
						) Q
						where S >= R
						order by Q.id
						limit 1;
					`,
					)

					// Find 3 random weighted abilities
					weightedSaleAbilities := []*boiler.SalePlayerAbility{}
					for {
						w := &boiler.SalePlayerAbility{}
						err := boiler.NewQuery(
							qm.SQL(q),
							qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
						).Bind(nil, gamedb.StdConn, w)
						if err != nil {
							gamelog.L.Error().Err(err).Msg(fmt.Sprintf("failed to get %d random weighted sale abilities", pas.Limit))
							return
						}

						isDuplicate := false
						for _, wsa := range weightedSaleAbilities {
							if wsa.ID == w.ID {
								isDuplicate = true
								break
							}
						}
						if isDuplicate {
							continue
						}

						weightedSaleAbilities = append(weightedSaleAbilities, w)

						if len(weightedSaleAbilities) == pas.Limit {
							break
						}
					}
					if len(weightedSaleAbilities) == 0 {
						gamelog.L.Warn().Msg("no sale abilities could be found in the db")
						break
					}

					oneHourFromNow := time.Now().Add(time.Duration(pas.TimeBetweenRefreshSeconds) * time.Second)
					rand.Seed(time.Now().UnixNano())
					randomIndexes := rand.Perm(len(weightedSaleAbilities))
					for _, i := range randomIndexes[:pas.Limit] {
						weightedSaleAbilities[i].AvailableUntil = null.TimeFrom(oneHourFromNow)
						weightedSaleAbilities[i].AmountSold = 0 // reset amount sold
						saleAbilities = append(saleAbilities, weightedSaleAbilities[i])
					}

					_, err = saleAbilities.UpdateAll(gamedb.StdConn, boiler.M{
						"available_until": oneHourFromNow,
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
					ws.PublishMessage("/secure_public/sale_abilities", server.HubKeySaleAbilitiesList, struct {
						NextRefreshTime *time.Time                `json:"next_refresh_time"`
						SaleAbilities   []*db.SaleAbilityDetailed `json:"sale_abilities,omitempty"`
					}{
						NextRefreshTime: &oneHourFromNow,
						SaleAbilities:   detailedSaleAbilities,
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
				ws.PublishMessage("/secure_public/sale_abilities", server.HubKeySaleAbilitiesPriceSubscribe, SaleAbilityPriceResponse{
					ID:           s.ID,
					CurrentPrice: s.CurrentPrice.StringFixed(0),
				})
			}
			break
		case <-salePeriodTicker.C:
			// Sale period ticker ticks every 10 minutes, resets the user ability purchase counts
			// Reset user ability purchase counts
			pas.ResetUserPurchaseCounts()
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
				ws.PublishMessage("/secure_public/sale_abilities", server.HubKeySaleAbilitiesPriceSubscribe, SaleAbilityPriceResponse{
					ID:           saleAbility.ID,
					CurrentPrice: saleAbility.CurrentPrice.StringFixed(0),
				})

				// Broadcast updated sale ability sold amount
				ws.PublishMessage("/secure_public/sale_abilities", server.HubKeySaleAbilitiesAmountSubscribe, SaleAbilityAmountResponse{
					ID:         saleAbility.ID,
					AmountSold: saleAbility.AmountSold,
				})
			}
			break
		}
	}
}
