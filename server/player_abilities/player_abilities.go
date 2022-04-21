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

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"go.uber.org/atomic"
)

type Purchase struct {
	PlayerID  uuid.UUID
	AbilityID string // sale ability id
}

type PlayerAbilitiesSystem struct {
	// player abilities
	salePlayerAbilities map[string]*boiler.SalePlayerAbility // map[ability_id]*Ability

	// ability purchase
	Purchase chan *Purchase

	messageBus *messagebus.MessageBus
	closed     *atomic.Bool
	sync.RWMutex
}

func NewPlayerAbilitiesSystem(messagebus *messagebus.MessageBus) *PlayerAbilitiesSystem {

	saleAbilities, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.AvailableUntil.GT(null.TimeFrom(time.Now()))).All(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("failed to populate salePlayerAbilities map with existing abilities from db")
	}
	salePlayerAbilities := map[string]*boiler.SalePlayerAbility{}
	for _, s := range saleAbilities {
		salePlayerAbilities[s.ID] = s
	}

	pas := &PlayerAbilitiesSystem{
		salePlayerAbilities: salePlayerAbilities,
		Purchase:            make(chan *Purchase),
		closed:              atomic.NewBool(false),
		messageBus:          messagebus,
	}

	go pas.SalePlayerAbilitiesUpdater()

	return pas
}

func (pas *PlayerAbilitiesSystem) SalePlayerAbilitiesUpdater() {
	priceTickerInterval := db.GetIntWithDefault("sale_ability_price_ticker_interval_seconds", 5) // default 5 seconds
	priceTicker := time.NewTicker(time.Duration(priceTickerInterval) * time.Second)

	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("Panic! Panic! Panic! Panic at the SalePlayerAbilitiesUpdater!", r)

			// re-run ability updater if ability system has not been cleaned up yet
			if pas != nil {
				pas.SalePlayerAbilitiesUpdater()
			}
		}
	}()

	defer func() {
		priceTicker.Stop()
		pas.closed.Store(true)
	}()

	oneHundred := decimal.NewFromFloat(100.0)
	for {
		select {
		case <-priceTicker.C:
			reductionPercentage := db.GetDecimalWithDefault("sale_ability_reduction_percentage", decimal.NewFromFloat(1.0)) // default 1%
			floorPrice := db.GetDecimalWithDefault("sale_ability_floor_price", decimal.New(10, 18))                         // default 10 sups

			// Check each ability that is on sale, remove them if expired
			for _, s := range pas.salePlayerAbilities {
				if s.AvailableUntil.Time.After(time.Now()) {
					continue
				}
				delete(pas.salePlayerAbilities, s.ID)
			}

			if len(pas.salePlayerAbilities) < 1 {
				gamelog.L.Debug().Msg("repopulating sale abilities since there aren't any more")
				// If no abilities are on sale, refill sale abilities
				saleAbilities, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.AvailableUntil.GT(null.TimeFrom(time.Now()))).All(gamedb.StdConn)
				if errors.Is(err, sql.ErrNoRows) || len(saleAbilities) == 0 {
					gamelog.L.Debug().Msg("refreshing sale abilities in db")
					// If no sale abilities, get 3 random sale abilities and update their time to an hour from now
					limit := db.GetIntWithDefault("sale_ability_limit", 3) // default 3
					allSaleAbilities, err := boiler.SalePlayerAbilities().All(gamedb.StdConn)
					if err != nil {
						gamelog.L.Error().Err(err).Msg(fmt.Sprintf("failed to get %d random sale abilities", limit))
						break
					}

					oneHourFromNow := time.Now().Add(time.Hour)
					rand.Seed(time.Now().UnixNano())
					randomIndexes := rand.Perm(len(allSaleAbilities))
					for _, i := range randomIndexes[:limit] {
						allSaleAbilities[i].AvailableUntil = null.TimeFrom(oneHourFromNow)
						saleAbilities = append(saleAbilities, allSaleAbilities[i])
					}

					_, err = saleAbilities.UpdateAll(gamedb.StdConn, boiler.M{
						"available_until": oneHourFromNow,
					})
					if err != nil {
						gamelog.L.Error().Err(err).Msg("failed to update sale ability with new expiration date")
						continue
					}
					// Broadcast trigger of sale abilities list update
					pas.messageBus.Send(messagebus.BusKey(server.HubKeySaleAbilitiesListUpdated), true)
				} else if err != nil {
					gamelog.L.Error().Err(err).Msg("failed to fill sale player abilities map with new sale abilities")
					break
				}
				for _, s := range saleAbilities {
					pas.salePlayerAbilities[s.ID] = s
				}
			}

			for _, s := range pas.salePlayerAbilities {
				s.CurrentPrice = s.CurrentPrice.Mul(oneHundred.Sub(reductionPercentage).Div(oneHundred))
				if s.CurrentPrice.LessThan(floorPrice) {
					s.CurrentPrice = floorPrice
				}

				_, err := s.Update(gamedb.StdConn, boil.Infer())
				if err != nil {
					gamelog.L.Error().Err(err).Str("salePlayerAbilityID", s.ID).Str("new price", s.CurrentPrice.String()).Interface("sale ability", s).Msg("failed to update sale ability price")
					continue
				}

				// Broadcast updated sale ability
				pas.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeySaleAbilityPriceSubscribe, s.ID)), s.CurrentPrice)
			}
			break
		case purchase := <-pas.Purchase:
			if saleAbility, ok := pas.salePlayerAbilities[purchase.AbilityID]; ok {
				inflationPercentage := db.GetDecimalWithDefault("sale_ability_inflation_percentage", decimal.NewFromFloat(20.0)) // default 20%
				saleAbility.CurrentPrice = saleAbility.CurrentPrice.Mul(oneHundred.Add(inflationPercentage).Div(oneHundred))
				_, err := saleAbility.Update(gamedb.StdConn, boil.Infer())
				if err != nil {
					gamelog.L.Error().Err(err).Str("salePlayerAbilityID", saleAbility.ID).Str("new price", saleAbility.CurrentPrice.String()).Interface("sale ability", saleAbility).Msg("failed to update sale ability price")
					break
				}
				pas.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeySaleAbilityPriceSubscribe, saleAbility.ID)), saleAbility.CurrentPrice)
			}
			break
		}
	}
}
