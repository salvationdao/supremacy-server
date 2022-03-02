package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"server"
	"server/battle_arena"
	"server/gamelog"
	"server/passport"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/sasha-s/go-deadlock"
	"nhooyr.io/websocket"
)

// FactionControllerWS holds handlers for checking server status
type FactionControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewFactionController creates the check hub
func NewFactionController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *FactionControllerWS {
	factionHub := &FactionControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "faction_hub"),
		API:  api,
	}

	api.SecureUserFactionCommand(HubKeGameAbilityContribute, factionHub.GameAbilityContribute)

	// subscription
	api.SecureUserFactionSubscribeCommand(HubKeyFactionAbilitiesUpdated, factionHub.FactionAbilitiesUpdateSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyWarMachineAbilitiesUpdated, factionHub.WarMachineAbilitiesUpdateSubscribeHandler)
	api.SecureUserSubscribeCommand(server.HubKeyFactionQueueJoin, factionHub.QueueSubscription)

	return factionHub
}

type GameAbilityContributeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		GameAbilityID server.GameAbilityID `json:"gameAbilityID"`
		Amount        server.BigInt        `json:"amount"`
	} `json:"payload"`
}

const HubKeGameAbilityContribute hub.HubCommandKey = "GAME:ABILITY:CONTRIBUTE"

func (fc *FactionControllerWS) GameAbilityContribute(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	gamelog.GameLog.Info().Str("fn", "GameAbilityContribute").RawJSON("req", payload).Msg("ws handler")
	req := &GameAbilityContributeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get user detail
	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return terror.Error(errors.New("user ID is nil"), "There was a problem getting your user")
	}

	// check whether the battle is started
	if fc.API.BattleArena.GetCurrentState().State != server.StateMatchStart {
		return terror.Error(fmt.Errorf("wrong game state: current state %s, match state %s", fc.API.BattleArena.GetCurrentState().State, server.StateMatchStart), "The battle hasn't started yet")
	}

	fc.API.votePhaseChecker.RLock()
	if fc.API.votePhaseChecker.Phase == VotePhaseWaitMechIntro {
		fc.API.votePhaseChecker.RUnlock()
		return terror.Error(terror.ErrForbidden, "Ability Contribute are available after intro")
	}
	fc.API.votePhaseChecker.RUnlock()

	if req.Payload.Amount.Cmp(big.NewInt(0)) <= 0 {
		return terror.Error(fmt.Errorf("bad amount: %s", req.Payload.Amount.String()), "Invalid contribute amount")
	}

	// get client detail
	hcd := fc.API.UserMap.GetUserDetail(wsc)
	if hcd == nil {
		return terror.Error(errors.New("nil hcd"), "User not found")
	}

	factionID := hcd.FactionID

	// calculate how many sups worth
	oneSups := big.NewInt(0)
	oneSups, ok := oneSups.SetString("1000000000000000000", 10)
	if !ok {
		return terror.Error(fmt.Errorf("unable to convert 1000000000000000000 to big int"), "There was an issue calculating the SUPS")
	}
	req.Payload.Amount.Mul(&req.Payload.Amount.Int, oneSups)

	fc.API.gameAbilityPool[factionID](func(fap *deadlock.Map) {
		// find ability

		faIface, ok := fap.Load(req.Payload.GameAbilityID.String())
		if !ok {
			fc.Log.Err(errors.New("could not load")).Msg("fap load")
			return
		}

		fa := faIface.(*GameAbilityPrice)

		fa.RLock()
		if fa.isReached {
			fa.RUnlock()
			return
		}
		fa.RUnlock()

		exceedFund := big.NewInt(0)
		exceedFund.Add(exceedFund, &fa.CurrentSups.Int)
		exceedFund.Add(exceedFund, &req.Payload.Amount.Int)

		isReached := false
		if exceedFund.Cmp(&fa.MaxTargetPrice.Int) >= 0 {
			fa.Lock()
			fa.isReached = true
			isReached = true
			fa.Unlock()
		}

		reduceAmount := server.BigInt{Int: *big.NewInt(0)}
		if !isReached {
			reduceAmount.Add(&reduceAmount.Int, &req.Payload.Amount.Int)
		} else {
			reduceAmount.Add(&reduceAmount.Int, &fa.MaxTargetPrice.Int)
			reduceAmount.Sub(&reduceAmount.Int, &fa.CurrentSups.Int)
		}

		// check sups
		reason := fmt.Sprintf("battle:%s|game_ability_contribution:%s", fc.API.BattleArena.CurrentBattleID(), req.Payload.GameAbilityID)

		go func() {
			fc.API.Passport.SpendSupMessage(passport.SpendSupsReq{
				FromUserID:           userID,
				Amount:               reduceAmount.String(),
				TransactionReference: server.TransactionReference(fmt.Sprintf("%s|%s", reason, uuid.Must(uuid.NewV4()))),
				GroupID:              "Battle",
			}, func(transaction string) {
				faIface, ok := fap.Load(req.Payload.GameAbilityID.String())
				if !ok {
					fc.Log.Err(errors.New("could not load")).Msg("fap load")
					return
				}

				fa := faIface.(*GameAbilityPrice)

				fa.PriceRW.Lock()
				fa.TxMX.Lock()
				fa.TxRefs = append(fa.TxRefs, transaction)

				fc.API.liveSupsSpend[hcd.FactionID].Lock()
				fc.API.liveSupsSpend[hcd.FactionID].TotalVote.Add(&fc.API.liveSupsSpend[hcd.FactionID].TotalVote.Int, &req.Payload.Amount.Int)
				fc.API.liveSupsSpend[hcd.FactionID].Unlock()

				// fc.API.ClientVoted(wsc)

				fc.API.UserMultiplier.Voted(userID)

				// append transaction ref

				if !isReached {
					// increase current sups and return
					fa.CurrentSups.Add(&fa.CurrentSups.Int, &req.Payload.Amount.Int)
					fap.Store(fa.GameAbility.Identity.String(), fa)
					fa.TxMX.Unlock()
					fa.PriceRW.Unlock()

					data := fmt.Sprintf("%s_%s_%s_%d", fa.GameAbility.Identity, fa.TargetPrice.String(), fa.CurrentSups.String(), 0)
					payload := []byte{}
					payload = append(payload, byte(battle_arena.NetMessageTypeAbilityTargetPriceTick))
					payload = append(payload, []byte(data)...)
					go wsc.SendWithMessageType(payload, websocket.MessageBinary)
					return
				}

				// otherwise, clear transaction and bump the price
				fa.TxRefs = []string{}
				fa.TxMX.Unlock()
				fa.PriceRW.Unlock()

				fa.Lock()
				defer fa.Unlock()
				fa.isReached = false

				fa.PriceRW.Lock()
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

				// update sups cost of the ability in db
				fa.GameAbility.SupsCost = fa.TargetPrice.String()

				fap.Store(fa.GameAbility.Identity.String(), fa)
				fa.PriceRW.Unlock()

				// trigger battle arena function to handle game ability
				abilityTriggerEvent := &server.GameAbilityEvent{
					IsTriggered:         true,
					TriggeredByUserID:   &userID,
					TriggeredByUsername: &hcd.Username,
					GameClientAbilityID: fa.GameAbility.GameClientAbilityID,
					WarMachineHash:      &fa.GameAbility.WarMachineHash,
				}

				if fa.GameAbility.AbilityHash == "" {
					abilityTriggerEvent.GameAbilityID = &fa.GameAbility.ID
				} else {
					abilityTriggerEvent.AbilityHash = &fa.GameAbility.AbilityHash
				}

				err = fc.API.BattleArena.GameAbilityTrigger(abilityTriggerEvent)
				if err != nil {
					fc.Log.Err(err).Msg("")
					return
				}

				triggeredBy := hcd.Brief()
				ability := fa.GameAbility.Brief()
				// broadcast notification
				if fa.GameAbility.AbilityHash == "" {
					go fc.API.BroadcastGameNotificationAbility(ctx, GameNotificationTypeFactionAbility, &GameNotificationAbility{
						User:    triggeredBy,
						Ability: ability,
					})
				} else {
					warMachine := fc.API.BattleArena.GetWarMachine(fa.GameAbility.WarMachineHash).Brief()
					// broadcast notification
					go fc.API.BroadcastGameNotificationWarMachineAbility(ctx, &GameNotificationWarMachineAbility{
						User:       hcd.Brief(),
						Ability:    fa.GameAbility.Brief(),
						WarMachine: warMachine,
					})
				}
				// prepare broadcast data
				targetPriceList := []string{}

				fap.Range(func(key interface{}, gameAbilityPrice interface{}) bool {
					fa := gameAbilityPrice.(*GameAbilityPrice)
					hasTriggered := 0
					if server.GameAbilityID(fa.GameAbility.Identity) == req.Payload.GameAbilityID {
						hasTriggered = 1
					}
					fa.PriceRW.RLock()
					targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s_%d", fa.GameAbility.Identity, fa.TargetPrice.String(), fa.CurrentSups.String(), hasTriggered))
					fa.PriceRW.RUnlock()
					return true
				})

				// broadcast if target price is updated
				if strings.Join(targetPriceList, "|") != "" {
					// prepare broadcast payload
					payload := []byte{}
					payload = append(payload, byte(battle_arena.NetMessageTypeAbilityTargetPriceTick))
					payload = append(payload, []byte(strings.Join(targetPriceList, "|"))...)
					go fc.API.NetMessageBus.Send(ctx, messagebus.NetBusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilityPriceUpdated, factionID)), payload)
				}

			})
		}()
	})

	reply(true)

	return nil
}

const HubKeyFactionAbilitiesUpdated hub.HubCommandKey = "FACTION:ABILITIES:UPDATED"

func (fc *FactionControllerWS) FactionAbilitiesUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	gamelog.GameLog.Info().Str("fn", "FactionAbilitiesUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	hcd := fc.API.UserMap.GetUserDetail(wsc)
	if hcd == nil {
		return "", "", terror.Error(fmt.Errorf("hcd is nil"), "User not found")
	}

	// skip, if faction is zaibatsu
	if hcd.FactionID == server.ZaibatsuFactionID {
		return "", "", nil
	}

	fc.API.gameAbilityPool[hcd.FactionID](func(fap *deadlock.Map) {
		abilities := []*server.GameAbility{}

		fap.Range(func(key interface{}, gameAbilityPrice interface{}) bool {
			fa := gameAbilityPrice.(*GameAbilityPrice)
			if fa.GameAbility.AbilityHash > "" {
				return true
			}
			fa.GameAbility.CurrentSups = fa.CurrentSups.String()
			abilities = append(abilities, fa.GameAbility)

			return true
		})

		reply(abilities)
	})
	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilitiesUpdated, hcd.FactionID))
	return req.TransactionID, busKey, nil
}

const HubKeyWarMachineAbilitiesUpdated hub.HubCommandKey = "WAR:MACHINE:ABILITIES:UPDATED"

type WarMachineAbilitiesUpdatedRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ParticipantID byte `json:"participantID"`
	} `json:"payload"`
}

// WarMachineAbilitiesUpdateSubscribeHandler subscribe on war machine abilities
func (fc *FactionControllerWS) WarMachineAbilitiesUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	gamelog.GameLog.Info().Str("fn", "WarMachineAbilitiesUpdateSubscribeHandler").RawJSON("req", payload).Msg("ws handler")
	req := &WarMachineAbilitiesUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	hcd := fc.API.UserMap.GetUserDetail(wsc)
	if hcd == nil {
		return "", "", terror.Error(fmt.Errorf("hcd is nil"), "User not found")
	}

	fc.API.gameAbilityPool[hcd.FactionID](func(fap *deadlock.Map) {
		abilities := []*server.GameAbility{}
		fap.Range(func(key interface{}, gameAbilityPrice interface{}) bool {
			fa := gameAbilityPrice.(*GameAbilityPrice)
			if hcd.FactionID == server.ZaibatsuFactionID {
				if fa.GameAbility.ParticipantID == nil ||
					*fa.GameAbility.ParticipantID != req.Payload.ParticipantID {
					return true
				}
			} else {
				if fa.GameAbility.AbilityHash == "" ||
					fa.GameAbility.ParticipantID == nil ||
					*fa.GameAbility.ParticipantID != req.Payload.ParticipantID {
					return true
				}
			}

			fa.GameAbility.CurrentSups = fa.CurrentSups.String()
			abilities = append(abilities, fa.GameAbility)
			fap.Store(fa.GameAbility.Identity.String(), fa)

			return true
		})

		reply(abilities)
	})
	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s:%x", HubKeyWarMachineAbilitiesUpdated, hcd.FactionID, req.Payload.ParticipantID))
	return req.TransactionID, busKey, nil
}

func (fc *FactionControllerWS) QueueSubscription(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	gamelog.GameLog.Info().Str("fn", "QueueSubscription").RawJSON("req", payload).Msg("ws handler")
	req := &WarMachineAbilitiesUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	hcd := fc.API.UserMap.GetUserDetail(wsc)
	if hcd == nil {
		return "", "", terror.Error(fmt.Errorf("hcd is nil"), "User not found")
	}

	if hcd.Faction == nil {
		return "", "", nil
	}

	if fc.API.BattleArena.WarMachineQueue == nil {
		return "", "", nil
	}

	switch hcd.Faction.Label {
	case "Red Mountain Offworld Mining Corporation":
		if fc.API.BattleArena.WarMachineQueue.RedMountain == nil {
			reply(0)
		}
		reply(fc.API.BattleArena.WarMachineQueue.RedMountain.QueuingLength())
	case "Boston Cybernetics":
		if fc.API.BattleArena.WarMachineQueue.RedMountain == nil {
			reply(0)
		}
		reply(fc.API.BattleArena.WarMachineQueue.Boston.QueuingLength())
	case "Zaibatsu Heavy Industries":
		if fc.API.BattleArena.WarMachineQueue.RedMountain == nil {
			reply(0)
		}
		reply(fc.API.BattleArena.WarMachineQueue.Zaibatsu.QueuingLength())
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", server.HubKeyFactionQueueJoin, hcd.FactionID.String()))
	return req.TransactionID, busKey, nil
}
