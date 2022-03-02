package api

import (
	"context"
	"fmt"
	"math/big"
	"server"
	"server/battle_arena"
	"server/db"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/sasha-s/go-deadlock"
)

//type PricePool map[server.GameAbilityID]*GameAbilityPrice

type GameAbilityPrice struct {
	GameAbility *server.GameAbility

	deadlock.RWMutex
	isReached bool

	PriceRW        deadlock.RWMutex
	MaxTargetPrice server.BigInt
	TargetPrice    server.BigInt
	CurrentSups    server.BigInt

	TxMX   deadlock.Mutex
	TxRefs []string
}

type GameAbilityPoolTicker struct {
	TargetPriceUpdater     *tickle.Tickle
	TargetPriceBroadcaster *tickle.Tickle
	deadlock.RWMutex
}

func (api *API) StartGameAbilityPool(ctx context.Context, factionID server.FactionID, conn *pgxpool.Pool) {
	// initial game ability

	factionAbilitiesPool := &deadlock.Map{}

	go func() {
		for {
			if api.BattleArena.BattleActive() {
				api.abilityTargetPriceUpdater(factionID, conn)
				time.Sleep(10 * time.Second)
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}()

	go func() {
		for {
			if api.BattleArena.BattleActive() {
				api.abilityTargetPriceBroadcast(factionID)
				time.Sleep(500 * time.Millisecond)
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}()

	api.gameAbilityPool[factionID] = func(fn func(factionAbilitiesPool *deadlock.Map)) {
		fn(factionAbilitiesPool)
	}
}

func (api *API) abilityTargetPriceUpdater(factionID server.FactionID, conn *pgxpool.Pool) {
	minPrice := big.NewInt(1000000000000000000)
	// targetPriceChan := make(chan string)
	// errChan := make(chan error)

	// update ability target price
	api.gameAbilityPool[factionID](func(fap *deadlock.Map) {
		targetPriceList := []string{}
		fap.Range(func(key interface{}, gameAbilityPrice interface{}) bool {
			fa := gameAbilityPrice.(*GameAbilityPrice)

			// in order to reduce price by half after 5 minutes
			// reduce target price by 0.9772 on every tick
			fa.PriceRW.Lock()
			defer fa.PriceRW.Unlock()
			fa.TargetPrice.Mul(&fa.TargetPrice.Int, big.NewInt(9772))
			fa.TargetPrice.Div(&fa.TargetPrice.Int, big.NewInt(10000))

			if fa.TargetPrice.Cmp(minPrice) <= 0 {
				fa.TargetPrice = server.BigInt{Int: *minPrice}
				fap.Store(fa.GameAbility.Identity.String(), fa)
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
				fap.Store(fa.GameAbility.Identity.String(), fa)

				fa.TxMX.Lock()
				fa.TxRefs = []string{}
				fa.TxMX.Unlock()

				abilityTriggerEvent := &server.GameAbilityEvent{
					IsTriggered:         true,
					GameClientAbilityID: fa.GameAbility.GameClientAbilityID,
					ParticipantID:       fa.GameAbility.ParticipantID,
				}
				if fa.GameAbility.AbilityHash == "" {
					abilityTriggerEvent.GameAbilityID = &fa.GameAbility.ID
				} else {
					abilityTriggerEvent.AbilityHash = &fa.GameAbility.AbilityHash
				}

				// trigger battle arena function to handle game ability
				err := api.BattleArena.GameAbilityTrigger(abilityTriggerEvent)
				if err != nil {
					return false
				}

				ability := fa.GameAbility.Brief()
				if fa.GameAbility.AbilityHash == "" {
					go api.BroadcastGameNotificationAbility(context.Background(), GameNotificationTypeFactionAbility, &GameNotificationAbility{
						Ability: ability,
					})
				} else {
					warMachine := api.BattleArena.GetWarMachine(fa.GameAbility.WarMachineHash).Brief()
					// broadcast notification
					go api.BroadcastGameNotificationWarMachineAbility(context.Background(), &GameNotificationWarMachineAbility{
						Ability:    ability,
						WarMachine: warMachine,
					})
				}

				hasTriggered = 1
			}

			// record current price
			targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s_%d", fa.GameAbility.Identity, fa.TargetPrice.String(), fa.CurrentSups.String(), hasTriggered))

			return true
		})

		targetPrice := strings.Join(targetPriceList, "|")
		if targetPrice != "" {
			// prepare broadcast payload
			payload := []byte{}
			payload = append(payload, byte(battle_arena.NetMessageTypeAbilityTargetPriceTick))
			payload = append(payload, []byte(strings.Join(targetPriceList, "|"))...)

			// start broadcast
			api.NetMessageBus.Send(context.Background(), messagebus.NetBusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilityPriceUpdated, factionID)), payload)
		}
	})
}

func (api *API) abilityTargetPriceBroadcast(factionID server.FactionID) {
	// get current target price data
	api.gameAbilityPool[factionID](func(fap *deadlock.Map) {
		targetPriceList := []string{}
		fap.Range(func(key interface{}, gameAbilityPrice interface{}) bool {
			fa := gameAbilityPrice.(*GameAbilityPrice)
			fa.PriceRW.RLock()
			targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s_%d", fa.GameAbility.Identity, fa.TargetPrice.String(), fa.CurrentSups.String(), 0))
			fa.PriceRW.RUnlock()
			return true
		})
		targetPrice := strings.Join(targetPriceList, "|")

		if targetPrice != "" {
			// prepare broadcast payload
			payload := []byte{}
			payload = append(payload, byte(battle_arena.NetMessageTypeAbilityTargetPriceTick))
			payload = append(payload, []byte(targetPrice)...)

			// start broadcast
			api.NetMessageBus.Send(context.Background(), messagebus.NetBusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilityPriceUpdated, factionID)), payload)
		}
	})
}

func (api *API) startGameAbilityPoolTicker(ctx context.Context, factionID server.FactionID, initialAbilities []*server.GameAbility, introSecond int) {
	// start game ability pool ticker after mech intro
	time.Sleep(time.Duration(introSecond) * time.Second)

	api.gameAbilityPool[factionID](func(fap *deadlock.Map) {
		// clean up ability incase battle didn't end properly
		fap.Range(func(key, value interface{}) bool {
			fap.Delete(key)
			return true
		})

		// start filling initial abilities
		factionAbilities := []*server.GameAbility{}
		warMachineAbilities := make(map[byte][]*server.GameAbility)

		for _, ability := range initialAbilities {

			fa := &GameAbilityPrice{
				GameAbility:    ability,
				MaxTargetPrice: server.BigInt{Int: *big.NewInt(0)},
				TargetPrice:    server.BigInt{Int: *big.NewInt(0)},
				CurrentSups:    server.BigInt{Int: *big.NewInt(0)},
				TxRefs:         []string{},
			}

			if factionID != server.ZaibatsuFactionID {
				if ability.AbilityHash == "" {
					factionAbilities = append(factionAbilities, ability)
				} else {
					if _, ok := warMachineAbilities[*ability.ParticipantID]; !ok {
						warMachineAbilities[*ability.ParticipantID] = []*server.GameAbility{}
					}
					warMachineAbilities[*ability.ParticipantID] = append(warMachineAbilities[*ability.ParticipantID], ability)
				}
			} else {
				if ability.AbilityHash == "" && ability.GameClientAbilityID != 11 {
					factionAbilities = append(factionAbilities, ability)
				} else {
					if _, ok := warMachineAbilities[*ability.ParticipantID]; !ok {
						warMachineAbilities[*ability.ParticipantID] = []*server.GameAbility{}
					}
					warMachineAbilities[*ability.ParticipantID] = append(warMachineAbilities[*ability.ParticipantID], ability)
				}
			}

			// calc target price
			initialTargetPrice := big.NewInt(0)
			initialTargetPrice, ok := initialTargetPrice.SetString(ability.SupsCost, 10)
			if !ok {
				api.Log.Err(fmt.Errorf("Failed to set initial target price"))
				return
			}

			fa.TargetPrice.Add(&fa.TargetPrice.Int, initialTargetPrice)
			fa.MaxTargetPrice.Add(&fa.MaxTargetPrice.Int, initialTargetPrice)
			fap.Store(ability.Identity.String(), fa)
		}
		// broadcast abilities
		if len(factionAbilities) > 0 {
			go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilitiesUpdated, factionID)), factionAbilities)
		}

		// // broadcast war machine ability
		// for participantID, abilities := range warMachineAbilities {
		// 	go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s:%x", HubKeyWarMachineAbilitiesUpdated, factionID, participantID)), abilities)
		// }
	})

}

func (api *API) stopGameAbilityPoolTicker() {
	txRefs := []string{}
	for factionID := range api.factionMap {

		// treat
		if factionID == server.ZaibatsuFactionID {
			api.gameAbilityPool[factionID](func(fap *deadlock.Map) {
				// commit all the left over transactions

				// do calculate for zaibatsu faction ability
				zaiTargetPrice := big.NewInt(0)
				dividedBy := 0

				fap.Range(func(key interface{}, value interface{}) bool {
					// read data
					fa := value.(*GameAbilityPrice)

					// directly target zaibatsu faction id
					if fa.GameAbility.GameClientAbilityID == 11 {
						zaiTargetPrice.Add(zaiTargetPrice, &fa.TargetPrice.Int)
						dividedBy += 1
					}

					// store left over transaction
					txRefs = append(txRefs, fa.TxRefs...)

					fap.Delete(key)
					return true
				})

				// calculate zaibatsu faction ability
				if dividedBy > 0 {
					actuallyPrice := big.NewInt(0)
					actuallyPrice.Add(actuallyPrice, zaiTargetPrice) // total remain price divide by total war machine
					actuallyPrice.Div(actuallyPrice, big.NewInt(int64(dividedBy)))

					// store new price back to db
					err := db.ZaibatsuFactionAbilityUpdate(context.Background(), api.Conn, actuallyPrice.String())
					if err != nil {
						api.Log.Err(err).Msg("failed to update zaibatsu faction ability")
					}

				}

			})
		} else {
			api.gameAbilityPool[factionID](func(fap *deadlock.Map) {
				fap.Range(func(key interface{}, value interface{}) bool {
					// read data
					fa := value.(*GameAbilityPrice)
					fa.GameAbility.SupsCost = fa.TargetPrice.String()

					// store new price back to db
					err := db.FactionExclusiveAbilitiesSupsCostUpdate(context.Background(), api.Conn, fa.GameAbility)
					if err != nil {
						api.Log.Err(err).Msg("failed to update zaibatsu faction ability")
					}

					// get left over transaction
					txRefs = append(txRefs, fa.TxRefs...)
					fap.Delete(key)
					return true
				})

			})
		}

	}
	api.Passport.ReleaseTransactions(txRefs)
}
