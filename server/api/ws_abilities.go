package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"server"
	"server/battle_arena"
	"server/db"
	"server/passport"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
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
	api.SecureUserFactionSubscribeCommand(HubKeyUserWarMachineQueueUpdated, factionHub.UserWarMachineQueueUpdatedSubscribeHandler)
	return factionHub
}

type UserWarMachineQueueUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID server.FactionID `json:"factionID"`
		UserID    server.UserID    `json:"userID"`
	} `json:"payload"`
}

const HubKeyUserWarMachineQueueUpdated hub.HubCommandKey = "USER:WAR:MACHINE:QUEUE:UPDATED"

// UserWarMachineQueueUpdatedSubscribeHandler subscribes a user to a list of their queued mechs
func (fc *FactionControllerWS) UserWarMachineQueueUpdatedSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	hcd, err := fc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return "", "", terror.Error(err)
	}

	if hcd.FactionID.IsNil() || hcd.FactionID.String() == "" {
		return "", "", terror.Error(fmt.Errorf("no faction ID provided"))
	}

	inBattleWarMachines := fc.API.BattleArena.InGameWarMachines()

	fc.API.BattleArena.BattleQueueMap[hcd.FactionID] <- func(wmq *battle_arena.WarMachineQueuingList) {
		warMachineQueuePosition := []*passport.WarMachineQueuePosition{}
		for i, wm := range wmq.WarMachines {
			if wm.OwnedByID != hcd.ID {
				continue
			}
			warMachineQueuePosition = append(warMachineQueuePosition, &passport.WarMachineQueuePosition{
				WarMachineMetadata: wm,
				Position:           i,
			})
		}
		// get in game war machine
		for _, wm := range inBattleWarMachines {
			if wm.OwnedByID != hcd.ID {
				continue
			}
			warMachineQueuePosition = append(warMachineQueuePosition, &passport.WarMachineQueuePosition{
				WarMachineMetadata: wm,
				Position:           -1,
			})
		}
		reply(warMachineQueuePosition)
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueueUpdated, hcd.ID))
	return req.TransactionID, busKey, nil
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
	req := &GameAbilityContributeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get user detail
	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	// check whether the battle is started
	if fc.API.BattleArena.GetCurrentState().State != server.StateMatchStart {
		return terror.Error(terror.ErrInvalidInput, "The battle hasn't started yet")
	}

	if fc.API.votePhaseChecker.Phase == VotePhaseWaitMechIntro {
		return terror.Error(terror.ErrInvalidInput, "Ability Contribute are available after intro")
	}

	if req.Payload.Amount.Cmp(big.NewInt(0)) <= 0 {
		return terror.Error(terror.ErrInvalidInput, "Invalid contribute amount")
	}

	// get client detail
	hcd, err := fc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(err)
	}

	factionID := hcd.FactionID

	// calculate how many sups worth
	oneSups := big.NewInt(0)
	oneSups, ok := oneSups.SetString("1000000000000000000", 10)
	if !ok {
		return terror.Error(fmt.Errorf("unable to convert 1000000000000000000 to big int"))
	}
	req.Payload.Amount.Mul(&req.Payload.Amount.Int, oneSups)

	fc.Log.Info().Msg("STARTING fc.API.gameAbilityPool[factionID]")
	go func() {
		select {
		case fc.API.gameAbilityPool[factionID] <- func(fap GameAbilitiesPool, fapt *GameAbilityPoolTicker) {
			// find ability
			fa, ok := fap[req.Payload.GameAbilityID]
			if !ok {
				fc.Log.Err(fmt.Errorf("error doesn't exist"))
				return
			}

			// check sups
			reason := fmt.Sprintf("battle:%s|game_ability_contribution:%s", fc.API.BattleArena.CurrentBattleID(), req.Payload.GameAbilityID)

			supTransactionReference, err := fc.API.Passport.SendHoldSupsMessage(ctx, userID, req.Payload.Amount, reason)
			if err != nil {
				fc.Log.Err(err).Msg("")
				return
			}

			// append transaction ref
			fa.TxRefs = append(fa.TxRefs, supTransactionReference)

			// increase current sups
			fa.CurrentSups.Add(&fa.CurrentSups.Int, &req.Payload.Amount.Int)

			// skip, if current sups is less than target price
			if fa.CurrentSups.Cmp(&fa.TargetPrice.Int) < 0 {
				return
			} else if fa.CurrentSups.Cmp(&fa.TargetPrice.Int) > 0 {
				// get dif
				reason := fmt.Sprintf("battle:%s|game_ability_contribution:%s:overpay_refund", fc.API.BattleArena.CurrentBattleID(), req.Payload.GameAbilityID)

				supReverseTransactionReference, err := fc.API.Passport.SendHoldSupsMessage(ctx, userID, req.Payload.Amount, reason)
				if err != nil {
					fc.Log.Err(err).Msg("")
				}
				if err == nil {
					// append transaction ref
					fa.TxRefs = append(fa.TxRefs, supReverseTransactionReference)
				}
			}

			// commit all the transactions
			results, err := fc.API.Passport.CommitTransactions(ctx, fa.TxRefs)
			if err != nil {
				fc.Log.Err(err).Msg("")
				return
			}

			// clear transaction reference
			fa.TxRefs = []server.TransactionReference{}

			for _, result := range results {
				if result.Status == server.TransactionFailed {
					// decrease amount
					fa.CurrentSups.Sub(&fa.CurrentSups.Int, &result.Amount.Int)
				}
			}

			// skip, if current sups is less than target price
			if fa.CurrentSups.Cmp(&fa.TargetPrice.Int) < 0 {
				return
			}

			// clear transaction reference
			fa.TxRefs = []server.TransactionReference{}

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

			// store new target price to passport server, if the ability is nft
			if fa.GameAbility.AbilityTokenID != 0 && fa.GameAbility.WarMachineTokenID != 0 {
				fc.API.Passport.AbilityUpdateTargetPrice(fc.API.ctx, fa.GameAbility.AbilityTokenID, fa.GameAbility.WarMachineTokenID, fa.TargetPrice.String())
			} else {
				err = db.FactionExclusiveAbilitiesSupsCostUpdate(ctx, fc.Conn, fa.GameAbility)
				if err != nil {
					fc.Log.Err(err).Msg("")
					return
				}
			}

			// trigger battle arena function to handle game ability
			abilityTriggerEvent := &server.GameAbilityEvent{
				IsTriggered:         true,
				TriggeredByUserID:   &userID,
				TriggeredByUsername: &hcd.Username,
				GameClientAbilityID: fa.GameAbility.GameClientAbilityID,
				WarMachineTokenID:   &fa.GameAbility.WarMachineTokenID,
			}

			if fa.GameAbility.AbilityTokenID == 0 {
				abilityTriggerEvent.GameAbilityID = &fa.GameAbility.ID
			} else {
				abilityTriggerEvent.AbilityTokenID = &fa.GameAbility.AbilityTokenID
			}
			err = fc.API.BattleArena.GameAbilityTrigger(abilityTriggerEvent)
			if err != nil {
				fc.Log.Err(err).Msg("")
				return
			}

			triggeredBy := hcd.Brief()
			ability := fa.GameAbility.Brief()
			// broadcast notification
			if fa.GameAbility.AbilityTokenID == 0 {
				go fc.API.BroadcastGameNotificationAbility(ctx, GameNotificationTypeFactionAbility, &GameNotificationAbility{
					User:    triggeredBy,
					Ability: ability,
				})
			} else {
				warMachine := fc.API.BattleArena.GetWarMachine(fa.GameAbility.WarMachineTokenID).Brief()
				// broadcast notification
				go fc.API.BroadcastGameNotificationWarMachineAbility(ctx, &GameNotificationWarMachineAbility{
					User:       hcd.Brief(),
					Ability:    fa.GameAbility.Brief(),
					WarMachine: warMachine,
				})
			}
			// prepare broadcast data
			targetPriceList := []string{}
			for abilityID, fa := range fap {
				hasTriggered := 0
				if abilityID == req.Payload.GameAbilityID {
					hasTriggered = 1
				}
				targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s_%d", fa.GameAbility.ID, fa.TargetPrice.String(), fa.CurrentSups.String(), hasTriggered))
			}

			// broadcast if target price is updated
			if strings.Join(targetPriceList, "|") != "" {
				// prepare broadcast payload
				payload := []byte{}
				payload = append(payload, byte(battle_arena.NetMessageTypeAbilityTargetPriceTick))
				payload = append(payload, []byte(strings.Join(targetPriceList, "|"))...)
				fc.API.NetMessageBus.Send(ctx, messagebus.NetBusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilityPriceUpdated, factionID)), payload)
			}
		}:

		case <-time.After(30 * time.Second):
			fc.API.Log.Err(errors.New("timeout on channel send exceeded"))
			panic("Game Ability Contribute")
		}

	}()

	reply(true)

	select {
	case fc.API.liveSupsSpend[hcd.FactionID] <- func(lvd *LiveVotingData) {
		lvd.TotalVote.Add(&lvd.TotalVote.Int, &req.Payload.Amount.Int)
	}:

	case <-time.After(10 * time.Second):
		fc.API.Log.Err(errors.New("timeout on channel send exceeded"))
		panic("store vote amount to live voting data after vote success")
	}

	return nil
}

const HubKeyFactionAbilitiesUpdated hub.HubCommandKey = "FACTION:ABILITIES:UPDATED"

func (fc *FactionControllerWS) FactionAbilitiesUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	hcd, err := fc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return "", "", terror.Error(err)
	}

	select {
	case fc.API.gameAbilityPool[hcd.FactionID] <- func(fap GameAbilitiesPool, fapt *GameAbilityPoolTicker) {
		abilities := []*server.GameAbility{}
		for _, fa := range fap {
			if fa.GameAbility.AbilityTokenID > 0 {
				continue
			}
			fa.GameAbility.CurrentSups = fa.CurrentSups.String()
			abilities = append(abilities, fa.GameAbility)
		}
		reply(abilities)
	}:
		busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilitiesUpdated, hcd.FactionID))
		return req.TransactionID, busKey, nil

	case <-time.After(10 * time.Second):
		fc.API.Log.Err(errors.New("timeout on channel send exceeded"))
		panic("Client Battle Reward Update")
	}
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
	req := &WarMachineAbilitiesUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	hcd, err := fc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return "", "", terror.Error(err)
	}

	select {
	case fc.API.gameAbilityPool[hcd.FactionID] <- func(fap GameAbilitiesPool, fapt *GameAbilityPoolTicker) {
		abilities := []*server.GameAbility{}
		for _, fa := range fap {
			if fa.GameAbility.AbilityTokenID == 0 ||
				fa.GameAbility.ParticipantID == nil ||
				*fa.GameAbility.ParticipantID != req.Payload.ParticipantID {
				continue
			}
			fa.GameAbility.CurrentSups = fa.CurrentSups.String()
			abilities = append(abilities, fa.GameAbility)
		}
		reply(abilities)
	}:

		busKey := messagebus.BusKey(fmt.Sprintf("%s:%s:%x", HubKeyWarMachineAbilitiesUpdated, hcd.FactionID, req.Payload.ParticipantID))
		return req.TransactionID, busKey, nil

	case <-time.After(10 * time.Second):
		fc.API.Log.Err(errors.New("timeout on channel send exceeded"))
		panic("War Machine Abilities Update Subscribe Handler")
	}
}
