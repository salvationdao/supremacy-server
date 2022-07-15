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
	"strings"
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
	userPurchaseLimits  map[uuid.UUID]int                       // map[player_id]purchase count for the current sale period
	nextRefresh         time.Time                               // timestamp of when the next sale period will begin

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
		boiler.SalePlayerAbilityWhere.RarityWeight.GTE(0),
		boiler.SalePlayerAbilityWhere.DeletedAt.IsNull(),
	).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to populate salePlayerAbilities map with existing abilities from db")
	}
	salePlayerAbilities := map[uuid.UUID]*boiler.SalePlayerAbility{}
	for _, s := range saleAbilities {
		sID := uuid.FromStringOrNil(s.ID)
		salePlayerAbilities[sID] = s
	}

	timeBetweenRefreshSeconds := db.GetIntWithDefault(db.KeySaleAbilityTimeBetweenRefreshSeconds, 600) // default 10 minutes (600 seconds)
	pas := &SalePlayerAbilitiesSystem{
		salePlayerAbilities:        salePlayerAbilities,
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
ticker:
	for {
		select {
		case <-priceTicker.C:
			// Price ticker ticks every 5 seconds, updates prices of abilities and refreshes the sale ability list when all abilities on sale have expired
			// Check each ability that is on sale, remove them if expired or if their sale limit has been reached
			for _, s := range pas.salePlayerAbilities {
				if s.AvailableUntil.Time.After(time.Now()) && s.RarityWeight >= 0 {
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
					boiler.SalePlayerAbilityWhere.DeletedAt.IsNull(),
					qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
				).All(gamedb.StdConn)
				if errors.Is(err, sql.ErrNoRows) || len(saleAbilities) == 0 {
					gamelog.L.Debug().Msg("refreshing sale abilities in db")
					// If no sale abilities, get 3 random sale abilities and update their time to an hour from now
					// Find 3 random weighted abilities
					weightedSaleAbilities := []*boiler.SalePlayerAbility{}
					weightedSaleAbilityIDs := []string{}
					attempts := 0
					for {
						attempts++
						if attempts > 15 {
							gamelog.L.Error().Err(err).Msg(fmt.Sprintf("failed to get %d random weighted sale abilities in under 15 attempts, aborting", pas.Limit))
							continue ticker
						}
						notIn := ""
						if len(weightedSaleAbilityIDs) > 0 {
							notIn += "and Q.id not in (" + strings.Join(weightedSaleAbilityIDs, ",") + ")"
						}

						q := fmt.Sprintf(
							`
							with cte as (
								select random() * (
									select sum(rarity_weight)
									from sale_player_abilities spa
									where spa.deleted_at is null and spa.rarity_weight >= 0
								) R
							)
							select 
								Q.id,
								Q.blueprint_id,
								Q.current_price,
								Q.available_until,
								Q.amount_sold,
								Q.sale_limit,
								Q.rarity_weight,
								Q.deleted_at
							from (
								select id, blueprint_id, current_price, available_until, amount_sold, sale_limit, rarity_weight, deleted_at, sum(rarity_weight) over (order by id) S, R
								from sale_player_abilities spa
								cross join cte
								where spa.deleted_at is null and spa.rarity_weight >= 0
							) Q
							where S >= R %s
							order by Q.id
							limit 1;
						`,
							notIn,
						)

						w := &boiler.SalePlayerAbility{}
						err := boiler.NewQuery(
							qm.SQL(q),
							qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
						).Bind(nil, gamedb.StdConn, w)
						if errors.Is(err, sql.ErrNoRows) {
							gamelog.L.Debug().Err(err).Msg(fmt.Sprintf("couldn't find a random weighted sale ability, retrying. attempts: %d", attempts))
							continue
						} else if err != nil {
							gamelog.L.Error().Err(err).Msg(fmt.Sprintf("failed to get %d random weighted sale abilities", pas.Limit))
							continue ticker
						}

						weightedSaleAbilities = append(weightedSaleAbilities, w)
						weightedSaleAbilityIDs = append(weightedSaleAbilityIDs, fmt.Sprintf("'%s'", w.ID))

						if len(weightedSaleAbilities) == pas.Limit {
							break
						}
					}
					if len(weightedSaleAbilities) == 0 {
						gamelog.L.Warn().Msg("no sale abilities could be found in the db")
						break
					}

					tenMinutesFromNow := time.Now().Add(time.Duration(pas.TimeBetweenRefreshSeconds) * time.Second)
					rand.Seed(time.Now().UnixNano())
					randomIndexes := rand.Perm(len(weightedSaleAbilities))
					for _, i := range randomIndexes[:pas.Limit] {
						weightedSaleAbilities[i].AvailableUntil = null.TimeFrom(tenMinutesFromNow)
						weightedSaleAbilities[i].AmountSold = 0 // reset amount sold
						saleAbilities = append(saleAbilities, weightedSaleAbilities[i])
					}

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
