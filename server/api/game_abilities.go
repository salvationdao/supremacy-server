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

type GameAbilitiesPool map[server.GameAbilityID]*GameAbilityPrice

type GameAbilityPrice struct {
	GameAbility    *server.GameAbility
	MaxTargetPrice server.BigInt
	TargetPrice    server.BigInt
	CurrentSups    server.BigInt
	TxRefs         []server.TransactionReference
}

type GameAbilityPoolTicker struct {
	TargetPriceUpdater     *tickle.Tickle
	TargetPriceBroadcaster *tickle.Tickle
}

func (api *API) StartGameAbilityPool(factionID server.FactionID, conn *pgxpool.Pool) {
	// initial game ability
	factionAbilitiesPool := make(GameAbilitiesPool)

	// initialise target price ticker
	tickle.MinDurationOverride = true
	AbilityTargetPriceUpdaterLogger := log_helpers.NamedLogger(api.Log, "Ability target price Updater").Level(zerolog.Disabled)
	AbilityTargetPriceUpdater := tickle.New("Ability target price Updater", 10, api.abilityTargetPriceUpdaterFactory(factionID, conn))
	AbilityTargetPriceUpdater.Log = &AbilityTargetPriceUpdaterLogger

	// initialise target price broadcaster
	AbilityTargetPriceBroadcasterLogger := log_helpers.NamedLogger(api.Log, "Ability target price Broadcaster").Level(zerolog.Disabled)
	AbilityTargetPriceBroadcaster := tickle.New("Ability target price Broadcaster", 0.5, api.abilityTargetPriceBroadcasterFactory(factionID))
	AbilityTargetPriceBroadcaster.Log = &AbilityTargetPriceBroadcasterLogger

	ts := &GameAbilityPoolTicker{
		TargetPriceUpdater:     AbilityTargetPriceUpdater,
		TargetPriceBroadcaster: AbilityTargetPriceBroadcaster,
	}

	for {
		for fn := range api.gameAbilityPool[factionID] {
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
		api.gameAbilityPool[factionID] <- func(fap GameAbilitiesPool, fapt *GameAbilityPoolTicker) {
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

					abilityTriggerEvent := &server.GameAbilityEvent{
						IsTriggered:         true,
						GameClientAbilityID: fa.GameAbility.GameClientAbilityID,
						ParticipantID:       fa.GameAbility.ParticipantID,
					}
					if fa.GameAbility.AbilityTokenID == 0 {
						abilityTriggerEvent.GameAbilityID = &fa.GameAbility.ID
					} else {
						abilityTriggerEvent.AbilityTokenID = &fa.GameAbility.AbilityTokenID
					}

					// trigger battle arena function to handle game ability
					err = api.BattleArena.GameAbilityTrigger(abilityTriggerEvent)
					if err != nil {
						targetPriceChan <- ""
						errChan <- terror.Error(err)
						return
					}

					if fa.GameAbility.AbilityTokenID == 0 {
						go api.BroadcastGameNotificationAbility(GameNotificationTypeFactionAbility, &GameNotificationAbility{
							Ability: &AbilityBrief{
								Label:    fa.GameAbility.Label,
								ImageUrl: fa.GameAbility.ImageUrl,
								Colour:   fa.GameAbility.Colour,
							},
						})
					} else {
						// broadcast notification
						go api.BroadcastGameNotificationWarMachineAbility(&GameNotificationWarMachineAbility{
							Ability: &AbilityBrief{
								Label:    fa.GameAbility.Label,
								ImageUrl: fa.GameAbility.ImageUrl,
								Colour:   fa.GameAbility.Colour,
							},
							WarMachine: &WarMachineBrief{
								Name:     fa.GameAbility.WarMachineName,
								ImageUrl: fa.GameAbility.WarMachineImage,
								Faction: &FactionBrief{
									Label:      api.factionMap[fa.GameAbility.FactionID].Label,
									Theme:      api.factionMap[fa.GameAbility.FactionID].Theme,
									LogoBlobID: api.factionMap[fa.GameAbility.FactionID].LogoBlobID,
								},
							},
						})
					}

					hasTriggered = 1
				}

				// record current price
				targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s_%d", fa.GameAbility.ID, fa.TargetPrice.String(), fa.CurrentSups.String(), hasTriggered))

				// store new target price to passport server, if the ability is nft
				if fa.GameAbility.AbilityTokenID != 0 && fa.GameAbility.WarMachineTokenID != 0 {
					api.Passport.AbilityUpdateTargetPrice(api.ctx, fa.GameAbility.AbilityTokenID, fa.GameAbility.WarMachineTokenID, fa.TargetPrice.String())
				} else {
					// update sups cost of the ability in db
					fa.GameAbility.SupsCost = fa.TargetPrice.String()
					err := db.FactionExclusiveAbilitiesSupsCostUpdate(api.ctx, conn, fa.GameAbility)
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
		api.gameAbilityPool[factionID] <- func(fap GameAbilitiesPool, fapt *GameAbilityPoolTicker) {
			targetPriceList := []string{}
			for _, fa := range fap {
				targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s_%d", fa.GameAbility.ID, fa.TargetPrice.String(), fa.CurrentSups.String(), 0))
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

func (api *API) startGameAbilityPoolTicker(factionID server.FactionID, initialAbilities []*server.GameAbility, introSecond int) {
	// start game ability pool ticker after mech intro
	time.Sleep(time.Duration(introSecond) * time.Second)

	api.gameAbilityPool[factionID] <- func(fap GameAbilitiesPool, fapt *GameAbilityPoolTicker) {
		// set initial ability
		factionAbilities := []*server.GameAbility{}
		warMachineAbilities := make(map[byte][]*server.GameAbility)
		for _, ability := range initialAbilities {
			fap[ability.ID] = &GameAbilityPrice{
				GameAbility:    ability,
				MaxTargetPrice: server.BigInt{Int: *big.NewInt(0)},
				TargetPrice:    server.BigInt{Int: *big.NewInt(0)},
				CurrentSups:    server.BigInt{Int: *big.NewInt(0)},
				TxRefs:         []server.TransactionReference{},
			}

			if ability.AbilityTokenID == 0 {
				factionAbilities = append(factionAbilities, ability)
			} else {
				if _, ok := warMachineAbilities[*ability.ParticipantID]; !ok {
					warMachineAbilities[*ability.ParticipantID] = []*server.GameAbility{}
				}
				warMachineAbilities[*ability.ParticipantID] = append(warMachineAbilities[*ability.ParticipantID], ability)
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

		// broadcast abilities
		if len(factionAbilities) > 0 {
			api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilitiesUpdated, factionID)), factionAbilities)
		}

		// broadcast war machine ability
		for participantID, abilities := range warMachineAbilities {
			api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s:%x", HubKeyWarMachineAbilitiesUpdated, factionID, participantID)), abilities)
		}

		// start all the tickles
		fapt.TargetPriceUpdater.Start()
		fapt.TargetPriceBroadcaster.Start()
	}
}

func (api *API) stopGameAbilityPoolTicker() {
	for factionID := range api.factionMap {
		api.gameAbilityPool[factionID] <- func(fap GameAbilitiesPool, fapt *GameAbilityPoolTicker) {
			// stop all the tickles
			if fapt.TargetPriceUpdater.NextTick != nil {
				fapt.TargetPriceUpdater.Stop()
			}
			if fapt.TargetPriceBroadcaster.NextTick != nil {
				fapt.TargetPriceBroadcaster.Stop()
			}

			// clean up pool
			for key := range fap {
				delete(fap, key)
			}

			// commit all the left over transactions
			txRefs := []server.TransactionReference{}
			for _, fa := range fap {
				txRefs = append(txRefs, fa.TxRefs...)
			}

			api.Passport.ReleaseTransactions(context.Background(), txRefs)
		}
	}
}
