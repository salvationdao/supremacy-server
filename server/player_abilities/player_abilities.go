package player_abilities

import (
	"database/sql"
	"errors"
	"fmt"
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
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/atomic"
)

type Purchase struct {
	PlayerID  uuid.UUID
	AbilityID string
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
	salePlayerAbilities := map[string]*boiler.SalePlayerAbility{}

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
	priceTicker := time.NewTicker(1 * time.Second)
	saleTicker := time.NewTicker(5 * time.Second)

	defer func() {
		if r := recover(); r != nil {
			gamelog.LogPanicRecovery("Panic! Panic! Panic! Panic at the FactionUniqueAbilityUpdater!", r)

			// re-run ability updater if ability system has not been cleaned up yet
			if pas != nil {
				pas.SalePlayerAbilitiesUpdater()
			}
		}
	}()

	defer func() {
		priceTicker.Stop()
		saleTicker.Stop()
		pas.closed.Store(true)
	}()

	for {
		select {
		case <-priceTicker.C:
			oneHundred := decimal.NewFromFloat(100.0)
			reductionPercentage := db.GetDecimalWithDefault("sale_ability_reduction_percentage", decimal.NewFromFloat(1.0)) // default 1%
			floorPrice := db.GetDecimalWithDefault("sale_ability_floor_price", decimal.NewFromFloat(10))                    // default 10 sups

			for _, s := range pas.salePlayerAbilities {
				if s.AvailableUntil.Time.Before(time.Now()) {
					continue
				}

				s.CurrentPrice = s.CurrentPrice.Mul(oneHundred.Sub(reductionPercentage).Div(oneHundred))
				if s.CurrentPrice.LessThan(floorPrice) {
					s.CurrentPrice = floorPrice
				}

				_, err := s.Update(gamedb.StdConn, boil.Infer())
				if err != nil {
					gamelog.L.Error().Str("salePlayerAbilityID", s.ID).Str("new price", s.CurrentPrice.String()).Err(err).Msg("failed to update sale ability price")
					continue
				}

				// Broadcast updated sale ability
				pas.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeySaleAbilityPriceSubscribe, s.ID)), s.CurrentPrice)
			}
			break
		case <-saleTicker.C:
			// Check each ability that is on sale, remove them if expired
			for _, s := range pas.salePlayerAbilities {
				if !s.AvailableUntil.Time.After(time.Now()) {
					continue
				}
				delete(pas.salePlayerAbilities, s.ID)
			}

			if len(pas.salePlayerAbilities) < 1 {
				// If no abilities are on sale, refill sale abilities
				saleAbilities, err := boiler.SalePlayerAbilities(boiler.SalePlayerAbilityWhere.AvailableUntil.GT(null.TimeFrom(time.Now()))).All(gamedb.StdConn)
				if errors.Is(err, sql.ErrNoRows) || len(saleAbilities) == 0 {
					// If no sale abilities, get 3 random sale abilities and update their time to an hour from now
					// todo: change to kv value
					limit := 3
					saleAbilities, err = boiler.SalePlayerAbilities(qm.OrderBy("random()"), qm.Limit(limit)).All(gamedb.StdConn)
					if err != nil {
						gamelog.L.Error().Err(err).Msg(fmt.Sprintf("failed to get %d random sale abilities", limit))
						break
					}
					_, err = saleAbilities.UpdateAll(gamedb.StdConn, boiler.M{
						"available_until": time.Now().Add(time.Hour),
					})
					if err != nil {
						gamelog.L.Error().Err(err).Msg("failed to update sale ability with new expiration date")
						continue
					}
				} else if err != nil {
					gamelog.L.Error().Err(err).Msg("failed to fill sale player abilities map with new sale abilities")
					break
				}
				for _, s := range saleAbilities {
					pas.salePlayerAbilities[s.ID] = s
				}
				// Broadcast trigger of sale abilities list update
				pas.messageBus.Send(messagebus.BusKey(server.HubKeySaleAbilitiesListUpdated), true)
			}
			break
		case purchase := <-pas.Purchase:
			if saleAbility, ok := pas.salePlayerAbilities[purchase.AbilityID]; ok {
				oneHundred := decimal.NewFromFloat(100.0)
				inflationPercentage := db.GetDecimalWithDefault("sale_ability_inflation_percentage", decimal.NewFromFloat(20.0)) // default 20%
				saleAbility.CurrentPrice = saleAbility.CurrentPrice.Mul(oneHundred.Add(inflationPercentage).Div(oneHundred))
				_, err := saleAbility.Update(gamedb.StdConn, boil.Infer())
				if err != nil {
					gamelog.L.Error().Str("salePlayerAbilityID", saleAbility.ID).Str("new price", saleAbility.CurrentPrice.String()).Err(err).Msg("failed to update sale ability price")
					break
				}
				pas.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeySaleAbilityPriceSubscribe, saleAbility.ID)), saleAbility.CurrentPrice)
			}
			break
		}
	}
}
