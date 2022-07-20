package player_abilities

import (
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

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/atomic"
)

type Claim struct {
	AbilityID string // sale ability id
}

// Used for sale abilities
type SalePlayerAbilitiesSystem struct {
	// sale player abilities
	salePlayerAbilities          map[string]*boiler.SalePlayerAbility // map[sale_id]*Ability
	salePlayerAbilitiesWithDupes []*db.SaleAbilityDetailed
	salePlayerAbilitiesPool      []*boiler.SalePlayerAbility
	totalSaleAbilities           int
	nextRefresh                  time.Time      // timestamp of when the next sale period will begin
	userClaimLimits              map[string]int // map[player_id]purchase count for the current sale period

	// KVs
	UserClaimLimit             int
	PriceTickerIntervalSeconds int
	TimeBetweenRefreshSeconds  int
	DisplayLimit               int

	// on sale ability purchase
	Claim chan *Claim

	closed *atomic.Bool
	sync.RWMutex
}

func NewSalePlayerAbilitiesSystem() *SalePlayerAbilitiesSystem {
	timeBetweenRefreshSeconds := db.GetIntWithDefault(db.KeySaleAbilityTimeBetweenRefreshSeconds, 600) // default 10 minutes (600 seconds)
	pas := &SalePlayerAbilitiesSystem{
		salePlayerAbilities:          map[string]*boiler.SalePlayerAbility{},
		salePlayerAbilitiesWithDupes: []*db.SaleAbilityDetailed{},
		salePlayerAbilitiesPool:      []*boiler.SalePlayerAbility{},
		totalSaleAbilities:           0,
		userClaimLimits:              make(map[string]int),
		nextRefresh:                  time.Now(),
		UserClaimLimit:               db.GetIntWithDefault(db.KeySaleAbilityPurchaseLimit, 1),              // default 1 purchase per user per ability
		PriceTickerIntervalSeconds:   db.GetIntWithDefault(db.KeySaleAbilityPriceTickerIntervalSeconds, 5), // default 5 seconds
		TimeBetweenRefreshSeconds:    timeBetweenRefreshSeconds,
		DisplayLimit:                 db.GetIntWithDefault(db.KeySaleAbilityLimit, 3), // default 3
		Claim:                        make(chan *Claim),
		closed:                       atomic.NewBool(false),
	}

	pas.RehydratePool()

	go pas.SalePlayerAbilitiesUpdater()

	return pas
}

func (pas *SalePlayerAbilitiesSystem) CurrentSaleList() []*db.SaleAbilityDetailed {
	pas.RLock()
	defer pas.RUnlock()

	return pas.salePlayerAbilitiesWithDupes
}

func (pas *SalePlayerAbilitiesSystem) RehydratePool() {
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

func (pas *SalePlayerAbilitiesSystem) NextRefresh() time.Time {
	pas.RLock()
	defer pas.RUnlock()

	return pas.nextRefresh
}

func (pas *SalePlayerAbilitiesSystem) Refresh() {
	pas.Lock()
	defer pas.Unlock()

	// Reset map
	pas.userClaimLimits = make(map[string]int)

	// Update sale period
	pas.nextRefresh = time.Now().Add(time.Duration(pas.TimeBetweenRefreshSeconds) * time.Second)
}

func (pas *SalePlayerAbilitiesSystem) IsAbilityAvailable(saleID string) bool {
	_, ok := pas.salePlayerAbilities[saleID]

	return ok
}

func (pas *SalePlayerAbilitiesSystem) CanUserClaim(userID string) bool {
	pas.RLock()
	defer pas.RUnlock()

	count, ok := pas.userClaimLimits[userID]
	if !ok {
		return true
	}

	return count < pas.UserClaimLimit
}

func (pas *SalePlayerAbilitiesSystem) AddToUserClaimCount(userID string) error {
	pas.Lock()
	defer pas.Unlock()

	count, ok := pas.userClaimLimits[userID]
	if !ok {
		count = 0
	}

	if count == pas.UserClaimLimit {
		minutes := int(time.Until(pas.nextRefresh).Minutes())
		msg := fmt.Sprintf("Please try again in %d minutes.", minutes)
		if minutes < 1 {
			msg = fmt.Sprintf("Please try again in %d seconds.", int(time.Until(pas.nextRefresh).Seconds()))
		}
		return fmt.Errorf("You have hit your claim limit of %d during this sale period. %s", pas.UserClaimLimit, msg)
	}

	pas.userClaimLimits[userID] = count + 1

	return nil
}

func (pas *SalePlayerAbilitiesSystem) SalePlayerAbilitiesUpdater() {
	priceTicker := time.NewTicker(time.Duration(pas.PriceTickerIntervalSeconds) * time.Second)

	defer func() {
		priceTicker.Stop()
		pas.closed.Store(true)
	}()

	for {
		select {
		case <-priceTicker.C:
			if len(pas.salePlayerAbilitiesPool) == 0 {
				gamelog.L.Debug().Msg("populating sale player abilities pool because it was empty")
				pas.RehydratePool()
			}

			// Update prices of abilities and refresh the sale ability list when period has ended
			if time.Now().After(pas.NextRefresh()) {
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
				ws.PublishMessage("/public/sale_abilities", server.HubKeySaleAbilitiesList, struct {
					NextRefreshTime              *time.Time                `json:"next_refresh_time"`
					RefreshPeriodDurationSeconds int                       `json:"refresh_period_duration_seconds"`
					SaleAbilities                []*db.SaleAbilityDetailed `json:"sale_abilities,omitempty"`
				}{
					NextRefreshTime:              &pas.nextRefresh,
					RefreshPeriodDurationSeconds: pas.TimeBetweenRefreshSeconds,
					SaleAbilities:                detailedSalePlayerAbilities,
				})

				pas.salePlayerAbilitiesWithDupes = []*db.SaleAbilityDetailed{}
				for _, s := range detailedSalePlayerAbilities {
					pas.salePlayerAbilities[s.ID] = s.SalePlayerAbility
					pas.salePlayerAbilitiesWithDupes = append(pas.salePlayerAbilitiesWithDupes, s)
				}
			}
		case purchase := <-pas.Claim:
			if saleAbility, ok := pas.salePlayerAbilities[purchase.AbilityID]; ok {
				saleAbility.AmountSold = saleAbility.AmountSold + 1
				_, err := saleAbility.Update(gamedb.StdConn, boil.Whitelist(
					boiler.SalePlayerAbilityColumns.AmountSold,
				))
				if err != nil {
					gamelog.L.Error().Err(err).Str("salePlayerAbilityID", saleAbility.ID).Interface("sale ability", saleAbility).Msg("failed to update sale ability amount sold")
					break
				}
			}
		}
	}
}
