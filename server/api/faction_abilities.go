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
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"nhooyr.io/websocket"
)

type FactionAbilitiesPool map[server.FactionAbilityID]*FactionAbilityPrice

type FactionAbilityPrice struct {
	FactionAbility *server.FactionAbility
	MaxTargetPrice server.BigInt
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
	minPrice := big.NewInt(1000000000000000000)

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

				if fa.TargetPrice.Cmp(minPrice) <= 0 {
					fa.TargetPrice = server.BigInt{Int: *minPrice}
				}

				hasTriggered := 0
				if fa.TargetPrice.Cmp(&fa.CurrentSups.Int) <= 0 {

					// calc min target price (half of last max target price)
					minTargetPrice := server.BigInt{Int: *big.NewInt(0)}
					minTargetPrice.Add(&minTargetPrice.Int, &fa.MaxTargetPrice.Int)
					minTargetPrice.Div(&minTargetPrice.Int, big.NewInt(2))

					// calc current new target price (twice of current target price)
					newTargetPrice := server.BigInt{Int: *big.NewInt(0)}
					newTargetPrice.Add(&newTargetPrice.Int, &fa.TargetPrice.Int)
					newTargetPrice.Mul(&newTargetPrice.Int, big.NewInt(2))

					// reset target price and max target price
					fa.TargetPrice = server.BigInt{Int: *big.NewInt(0)}
					fa.MaxTargetPrice = server.BigInt{Int: *big.NewInt(0)}
					if newTargetPrice.Cmp(&minTargetPrice.Int) >= 0 {
						fa.TargetPrice.Add(&fa.TargetPrice.Int, &newTargetPrice.Int)
						fa.MaxTargetPrice.Add(&fa.MaxTargetPrice.Int, &newTargetPrice.Int)
					} else {
						fa.TargetPrice.Add(&fa.TargetPrice.Int, &minTargetPrice.Int)
						fa.MaxTargetPrice.Add(&fa.MaxTargetPrice.Int, &minTargetPrice.Int)
					}

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

					abilityTriggerEvent := &server.FactionAbilityEvent{
						IsTriggered:         true,
						GameClientAbilityID: fa.FactionAbility.GameClientAbilityID,
						ParticipantID:       fa.FactionAbility.ParticipantID,
					}
					if fa.FactionAbility.AbilityTokenID == 0 {
						abilityTriggerEvent.FactionAbilityID = &fa.FactionAbility.ID
					} else {
						abilityTriggerEvent.AbilityTokenID = &fa.FactionAbility.AbilityTokenID
					}

					// trigger battle arena function to handle faction ability
					err = api.BattleArena.FactionAbilityTrigger(abilityTriggerEvent)
					if err != nil {
						targetPriceChan <- ""
						errChan <- terror.Error(err)
						return
					}

					// broadcast notification
					go api.BroadcastGameNotification(GameNotificationTypeText, fmt.Sprintf(`Ability %s in %s had been triggered`, fa.FactionAbility.Label, api.factionMap[factionID].Label))

					hasTriggered = 1
				}

				// record current price
				targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s_%d", fa.FactionAbility.ID, fa.TargetPrice.String(), fa.CurrentSups.String(), hasTriggered))

				// store new target price to passport server, if the ability is nft
				if fa.FactionAbility.AbilityTokenID != 0 && fa.FactionAbility.WarMachineTokenID != 0 {
					api.Passport.AbilityUpdateTargetPrice(api.ctx, fa.FactionAbility.AbilityTokenID, fa.FactionAbility.WarMachineTokenID, fa.TargetPrice.String())
				} else {
					// update sups cost of the ability in db
					fa.FactionAbility.SupsCost = fa.TargetPrice.String()
					err := db.FactionExclusiveAbilitiesSupsCostUpdate(api.ctx, conn, fa.FactionAbility)
					if err != nil {
						targetPriceChan <- ""
						errChan <- terror.Error(err)
						return

					}
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
				targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s_%d", fa.FactionAbility.ID, fa.TargetPrice.String(), fa.CurrentSups.String(), 0))
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

func (api *API) startFactionAbilityPoolTicker(factionID server.FactionID, initialAbilities []*server.FactionAbility, introSecond int) {
	// start faction ability pool ticker after mech intro
	time.Sleep(time.Duration(introSecond) * time.Second)

	api.factionAbilityPool[factionID] <- func(fap FactionAbilitiesPool, fapt *FactionAbilityPoolTicker) {
		// set initial ability
		for _, ability := range initialAbilities {
			fap[ability.ID] = &FactionAbilityPrice{
				FactionAbility: ability,
				MaxTargetPrice: server.BigInt{Int: *big.NewInt(0)},
				TargetPrice:    server.BigInt{Int: *big.NewInt(0)},
				CurrentSups:    server.BigInt{Int: *big.NewInt(0)},
				TxRefs:         []server.TransactionReference{},
			}

			// calc target price
			initialTargetPrice := big.NewInt(0)
			initialTargetPrice, ok := initialTargetPrice.SetString(ability.SupsCost, 10)
			if !ok {
				api.Log.Err(fmt.Errorf("Failed to set initial target price"))
				return
			}

			fap[ability.ID].TargetPrice.Add(&fap[ability.ID].TargetPrice.Int, initialTargetPrice)
			fap[ability.ID].MaxTargetPrice.Add(&fap[ability.ID].MaxTargetPrice.Int, initialTargetPrice)
		}

		// broadcast ability
		api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilitiesUpdated, factionID)), initialAbilities)

		// start all the tickles
		fapt.TargetPriceUpdater.Start()
		fapt.TargetPriceBroadcaster.Start()
	}
}

func (api *API) stopFactionAbilityPoolTicker() {
	for factionID := range api.factionMap {
		api.factionAbilityPool[factionID] <- func(fap FactionAbilitiesPool, fapt *FactionAbilityPoolTicker) {
			// stop all the tickles
			if fapt.TargetPriceUpdater.NextTick != nil {
				fapt.TargetPriceUpdater.Stop()
			}
			if fapt.TargetPriceBroadcaster.NextTick != nil {
				fapt.TargetPriceBroadcaster.Stop()
			}

			// commit all the left over transactions
			txRefs := []server.TransactionReference{}
			for _, fa := range fap {
				txRefs = append(txRefs, fa.TxRefs...)
			}

			if len(txRefs) > 0 {
				_, err := api.Passport.ReleaseTransactions(context.Background(), txRefs)
				if err != nil {
					api.Log.Err(err)
					return
				}
			}
		}
	}
}
