package api

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"server"
	"server/battle_arena"
	"server/db"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
	"nhooyr.io/websocket"
)

type FactionAbilitiesPool map[server.FactionAbilityID]*FactionAbilityPrice

type FactionAbilityPrice struct {
	FactionAbility *server.FactionAbility
	TargetPrice    server.BigInt
	CurrentSups    server.BigInt
	TxRefs         []server.TransactionReference
}

type FactionAbilityPoolTicker struct {
	TargetPriceUpdater     *tickle.Tickle
	TargetPriceBroadcaster *tickle.Tickle
}

func (api *API) StartFactionAbilityPool(factionID server.FactionID, conn *pgxpool.Pool) {
	// initial faction ability
	factionAbilitiesPool := make(FactionAbilitiesPool)

	// get initial abilities
	initialAbilities, err := db.FactionExclusiveAbilitiesByFactionID(api.ctx, conn, factionID)
	if err != nil {
		api.Log.Err(err).Msg("Failed to query initial faction abilities")
		return
	}

	// set initial ability
	for _, ability := range initialAbilities {
		factionAbilitiesPool[ability.ID] = &FactionAbilityPrice{
			FactionAbility: ability,
			TargetPrice:    server.BigInt{Int: *big.NewInt(0)},
			CurrentSups:    server.BigInt{Int: *big.NewInt(0)},
			TxRefs:         []server.TransactionReference{},
		}

		// calc target price
		initialTarget := big.NewInt(0)
		initialTarget, ok := initialTarget.SetString(ability.SupsCost, 10)
		if !ok {
			api.Log.Err(fmt.Errorf("Failed to set initial target price"))
			return
		}

		factionAbilitiesPool[ability.ID].TargetPrice.Add(&factionAbilitiesPool[ability.ID].TargetPrice.Int, initialTarget)
	}

	// initialise target price ticker
	tickle.MinDurationOverride = true
	AbilityTargetPriceUpdaterLogger := log_helpers.NamedLogger(api.Log, "Ability target price Updater").Level(zerolog.Disabled)
	AbilityTargetPriceUpdater := tickle.New("Ability target price Updater", 10, api.abilityTargetPriceUpdaterFactory(factionID, conn))
	AbilityTargetPriceUpdater.Log = &AbilityTargetPriceUpdaterLogger

	// initialise target price broadcaster
	AbilityTargetPriceBroadcasterLogger := log_helpers.NamedLogger(api.Log, "Ability target price Broadcaster").Level(zerolog.Disabled)
	AbilityTargetPriceBroadcaster := tickle.New("Ability target price Broadcaster", 0.5, api.abilityTargetPriceBroadcasterFactory(factionID))
	AbilityTargetPriceBroadcaster.Log = &AbilityTargetPriceBroadcasterLogger

	ts := &FactionAbilityPoolTicker{
		TargetPriceUpdater:     AbilityTargetPriceUpdater,
		TargetPriceBroadcaster: AbilityTargetPriceBroadcaster,
	}

	for {
		for fn := range api.factionAbilityPool[factionID] {
			fn(factionAbilitiesPool, ts)
		}
	}
}

func (api *API) abilityTargetPriceUpdaterFactory(factionID server.FactionID, conn *pgxpool.Pool) func() (int, error) {
	return func() (int, error) {
		targetPriceChan := make(chan string)
		errChan := make(chan error)

		// update ability target price
		api.factionAbilityPool[factionID] <- func(fap FactionAbilitiesPool, fapt *FactionAbilityPoolTicker) {
			targetPriceList := []string{}
			for _, fa := range fap {
				// in order to reduce price by half after 5 minutes
				// reduce target price by 0.9772 on every tick
				fa.TargetPrice.Mul(&fa.TargetPrice.Int, big.NewInt(9772))
				fa.TargetPrice.Div(&fa.TargetPrice.Int, big.NewInt(10000))

				if fa.TargetPrice.Cmp(&fa.CurrentSups.Int) <= 0 {
					//double the target price
					fa.TargetPrice.Mul(&fa.TargetPrice.Int, big.NewInt(2))

					// reset current sups to zero
					fa.CurrentSups = server.BigInt{Int: *big.NewInt(0)}

					// commit all the transactions
					_, err := api.Passport.CommitTransactions(context.Background(), fa.TxRefs)
					if err != nil {
						targetPriceChan <- ""
						errChan <- terror.Error(err)
						return
					}
					fa.TxRefs = []server.TransactionReference{}

					// trigger battle arena function to handle faction ability
					err = api.BattleArena.FactionAbilityTrigger(&battle_arena.AbilityTriggerRequest{
						FactionID:           fa.FactionAbility.FactionID,
						FactionAbilityID:    fa.FactionAbility.ID,
						IsSuccess:           true,
						GameClientAbilityID: fa.FactionAbility.GameClientAbilityID,
					})
					if err != nil {
						targetPriceChan <- ""
						errChan <- terror.Error(err)
						return
					}

					// broadcast notification
					go api.BroadcastGameNotification(GameNotificationTypeText, fmt.Sprintf(`Ability %s in %s had been triggered`, fa.FactionAbility.Label, api.factionMap[factionID].Label))

				}

				// record current price
				targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s", fa.FactionAbility.ID, fa.TargetPrice.String(), fa.CurrentSups.String()))

				// update sups cost of the ability in db
				fa.FactionAbility.SupsCost = fa.TargetPrice.String()
				err := db.FactionExclusiveAbilitiesSupsCostUpdate(api.ctx, conn, fa.FactionAbility)
				if err != nil {
					targetPriceChan <- ""
					errChan <- terror.Error(err)
					return

				}
			}
			targetPriceChan <- strings.Join(targetPriceList, "|")
			errChan <- nil
		}

		// wait for target price change
		targetPrice := <-targetPriceChan

		// wait for error check
		err := <-errChan
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err)
		}

		if targetPrice != "" {
			// prepare broadcast payload
			payload := []byte{}
			payload = append(payload, byte(battle_arena.NetMessageTypeAbilityTargetPriceTick))
			payload = append(payload, []byte(targetPrice)...)
			// start broadcast
			api.Hub.Clients(func(clients hub.ClientsList) {
				for client, ok := range clients {
					if !ok {
						continue
					}
					go func(c *hub.Client) {
						// get user faction id
						hcd, err := api.getClientDetailFromChannel(c)
						// skip, if error or not current faction player
						if err != nil || hcd.FactionID != factionID {
							return
						}

						// broadcast vote price forecast
						err = c.SendWithMessageType(payload, websocket.MessageBinary)
						if err != nil {
							api.Log.Err(err).Msg("failed to send broadcast")
						}
					}(client)
				}
			})
		}

		return http.StatusOK, nil
	}
}

func (api *API) abilityTargetPriceBroadcasterFactory(factionID server.FactionID) func() (int, error) {
	return func() (int, error) {
		// get current target price data
		targetPriceChan := make(chan string)
		api.factionAbilityPool[factionID] <- func(fap FactionAbilitiesPool, fapt *FactionAbilityPoolTicker) {
			targetPriceList := []string{}
			for _, fa := range fap {
				targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s", fa.FactionAbility.ID, fa.TargetPrice.String(), fa.CurrentSups.String()))
			}
			targetPriceChan <- strings.Join(targetPriceList, "|")
		}
		targetPrice := <-targetPriceChan

		if targetPrice != "" {
			// prepare broadcast payload
			payload := []byte{}
			payload = append(payload, byte(battle_arena.NetMessageTypeAbilityTargetPriceTick))
			payload = append(payload, []byte(targetPrice)...)
			// start broadcast
			api.Hub.Clients(func(clients hub.ClientsList) {
				for client, ok := range clients {
					if !ok {
						continue
					}
					go func(c *hub.Client) {
						// get user faction id
						hcd, err := api.getClientDetailFromChannel(c)
						// skip, if error or not current faction player
						if err != nil || hcd.FactionID != factionID {
							return
						}

						// broadcast vote price forecast
						err = c.SendWithMessageType(payload, websocket.MessageBinary)
						if err != nil {
							api.Log.Err(err).Msg("failed to send broadcast")
						}
					}(client)
				}
			})
		}

		return http.StatusOK, nil
	}
}

func (api *API) startFactionAbilityPoolTicker() {
	for factionID := range api.factionMap {
		api.factionAbilityPool[factionID] <- func(fap FactionAbilitiesPool, fapt *FactionAbilityPoolTicker) {
			// start all the tickles
			fapt.TargetPriceUpdater.Start()
			fapt.TargetPriceBroadcaster.Start()
		}
	}
}

func (api *API) stopFactionAbilityPoolTicker() {
	for factionID := range api.factionMap {
		api.factionAbilityPool[factionID] <- func(fap FactionAbilitiesPool, fapt *FactionAbilityPoolTicker) {
			// stop all the tickles
			fapt.TargetPriceUpdater.Stop()
			fapt.TargetPriceBroadcaster.Stop()

			// commit all the left over transactions
			txRefs := []server.TransactionReference{}
			for _, fa := range fap {
				txRefs = append(txRefs, fa.TxRefs...)
			}

			if len(txRefs) > 0 {
				_, err := api.Passport.CommitTransactions(context.Background(), txRefs)
				if err != nil {
					api.Log.Err(err)
					return
				}
			}
		}
	}
}
