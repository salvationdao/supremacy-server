package battle

import (
	"database/sql"
	"errors"
	"fmt"
	"server"
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
	Amount    decimal.Decimal
}

type PlayerAbilitiesSystem struct {
	// player abilities
	salePlayerAbilities map[string]*boiler.SalePlayerAbility // map[ability_id]*Ability

	// ability purchase
	purchase chan *Purchase

	_battle *Battle
	closed  *atomic.Bool
	sync.RWMutex
}

func (as *PlayerAbilitiesSystem) battle() *Battle {
	as.RLock()
	defer as.RUnlock()
	return as._battle
}

func NewPlayerAbilitiesSystem(battle *Battle) *PlayerAbilitiesSystem {
	salePlayerAbilities := map[string]*boiler.SalePlayerAbility{}

	pas := &PlayerAbilitiesSystem{
		salePlayerAbilities: salePlayerAbilities,
		purchase:            make(chan *Purchase),
		closed:              atomic.NewBool(false),
	}

	go pas.SalePlayerAbilitiesUpdater()

	return pas
}

func (pas *PlayerAbilitiesSystem) SalePlayerAbilitiesUpdater() {
	priceTicker := time.NewTicker(1 * time.Second)
	saleTicker := time.NewTicker(1 * time.Minute)

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
			pas.Lock()
			for _, s := range pas.salePlayerAbilities {
				if s.AvailableUntil.Time.Before(time.Now()) {
					continue
				}

				// todo: make this multiplier a kv value
				s.CurrentPrice = s.CurrentPrice.Mul(decimal.NewFromFloat(0.99)) // decrease by 1%
				// todo: make this check a kv value
				if s.CurrentPrice.LessThan(decimal.NewFromInt(10)) {
					s.CurrentPrice = decimal.NewFromInt(10)
				}

				_, err := s.Update(gamedb.StdConn, boil.Infer())
				if err != nil {
					gamelog.L.Error().Str("salePlayerAbilityID", s.ID).Str("new currentPrice", s.CurrentPrice.String()).Err(err).Msg("failed to update sale ability price")
					continue
				}

				// Broadcast updated sale ability
				pas.battle().arena.messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeySaleAbilityPriceSubscribe, s.ID)), s.CurrentPrice)
			}
			pas.Unlock()
			break
		case <-saleTicker.C:
			pas.Lock()
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
				if errors.Is(err, sql.ErrNoRows) {
					// If no sale abilities, get 3 random sale abilities and update their time to an hour from now
					// todo: change to kv value
					limit := 3
					saleAbilities, err = boiler.SalePlayerAbilities(qm.OrderBy("random()"), qm.Limit(limit)).All(gamedb.StdConn)
					if err != nil {
						gamelog.L.Error().Err(err).Msg(fmt.Sprintf("failed to get %d random sale abilities", limit))
					}
					for _, s := range saleAbilities {
						s.AvailableUntil = null.TimeFrom(s.AvailableUntil.Time.Add(time.Hour))
					}
				} else if err != nil {
					gamelog.L.Error().Err(err).Msg("failed to fill sale player abilities map with new sale abilities")
					break
				}
				for _, s := range saleAbilities {
					pas.salePlayerAbilities[s.ID] = s
				}
				break
			}
			pas.Unlock()
		}
	}
}
