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
	// player abilities
	salePlayerAbilities map[uuid.UUID]*boiler.SalePlayerAbility // map[ability_id]*Ability

	// ability purchase
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
		salePlayerAbilities: salePlayerAbilities,
		Purchase:            make(chan *Purchase),
		closed:              atomic.NewBool(false),
	}

	go pas.SalePlayerAbilitiesUpdater()

	return pas
}

func (pas *SalePlayerAbilitiesSystem) SalePlayerAbilitiesUpdater() {
	priceTickerInterval := db.GetIntWithDefault(db.SaleAbilityPriceTickerIntervalSeconds, 5) // default 5 seconds
	priceTicker := time.NewTicker(time.Duration(priceTickerInterval) * time.Second)
	reductionPercentage := db.GetDecimalWithDefault(db.SaleAbilityReductionPercentage, decimal.NewFromFloat(1.0))  // default 1%
	floorPrice := db.GetDecimalWithDefault(db.SaleAbilityFloorPrice, decimal.New(10, 18))                          // default 10 sups
	timeBetweenRefreshInSeconds := db.GetIntWithDefault(db.SaleAbilityTimeBetweenRefreshSeconds, 3600)             // default 1 hour (3600 seconds)
	limit := db.GetIntWithDefault(db.SaleAbilityLimit, 3)                                                          // default 3
	inflationPercentage := db.GetDecimalWithDefault(db.SaleAbilityInflationPercentage, decimal.NewFromFloat(20.0)) // default 20%

	defer func() {
		priceTicker.Stop()
		pas.closed.Store(true)
	}()

	oneHundred := decimal.NewFromFloat(100.0)
	for {
		select {
		case <-priceTicker.C:

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
					allSaleAbilities, err := boiler.SalePlayerAbilities(
						qm.Load(boiler.SalePlayerAbilityRels.Blueprint),
					).All(gamedb.StdConn)
					if err != nil {
						gamelog.L.Error().Err(err).Msg(fmt.Sprintf("failed to get %d random sale abilities", limit))
						break
					}
					if len(allSaleAbilities) == 0 {
						gamelog.L.Warn().Msg("no sale abilities could be found in the db")
						break
					}

					oneHourFromNow := time.Now().Add(time.Duration(timeBetweenRefreshInSeconds) * time.Second)
					rand.Seed(time.Now().UnixNano())
					randomIndexes := rand.Perm(len(allSaleAbilities))
					for _, i := range randomIndexes[:limit] {
						allSaleAbilities[i].AvailableUntil = null.TimeFrom(oneHourFromNow)
						allSaleAbilities[i].AmountSold = 0 // reset amount sold
						saleAbilities = append(saleAbilities, allSaleAbilities[i])
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
				s.CurrentPrice = s.CurrentPrice.Mul(oneHundred.Sub(reductionPercentage).Div(oneHundred))
				if s.CurrentPrice.LessThan(floorPrice) {
					s.CurrentPrice = floorPrice
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
		case purchase := <-pas.Purchase:
			if saleAbility, ok := pas.salePlayerAbilities[purchase.AbilityID]; ok {
				saleAbility.CurrentPrice = saleAbility.CurrentPrice.Mul(oneHundred.Add(inflationPercentage).Div(oneHundred))
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
