package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"server"
	"server/battle_arena"
	"server/db"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
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

	api.SecureUserFactionCommand(HubKeGameyAbilityContribute, factionHub.GameAbilityContribute)

	// subscription
	api.SecureUserFactionSubscribeCommand(HubKeyFactionAbilitiesUpdated, factionHub.FactionAbilitiesUpdateSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyWarMachineAbilitiesUpdated, factionHub.WarMachineAbilitiesUpdateSubscribeHandler)
	return factionHub
}

type GameAbilityContributeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		GameAbilityID server.GameAbilityID `json:"gameAbilityID"`
		Amount        server.BigInt        `json:"amount"`
	} `json:"payload"`
}

const HubKeGameyAbilityContribute hub.HubCommandKey = "GAME:ABILITY:CONTRIBUTE"

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

	// get client detail
	hcd, err := fc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(err)
	}

	factionID := hcd.FactionID

	// check whether the battle is started
	if fc.API.votePhaseChecker.Phase == VotePhaseHold {
		return terror.Error(terror.ErrInvalidInput, "The battle hasn't started yet")
	}

	// calculate how many sups worth
	oneSups := big.NewInt(0)
	oneSups, ok := oneSups.SetString("1000000000000000000", 10)
	if !ok {
		return terror.Error(fmt.Errorf("Unable to convert 1000000000000000000 to big int"))
	}
	req.Payload.Amount.Mul(&req.Payload.Amount.Int, oneSups)

	targetPriceChan := make(chan string)
	errChan := make(chan error)
	fc.API.gameAbilityPool[factionID] <- func(fap GameAbilitiesPool, fapt *GameAbilityPoolTicker) {
		// find ability
		fa, ok := fap[req.Payload.GameAbilityID]
		if !ok {
			targetPriceChan <- ""
			errChan <- terror.Error(terror.ErrInvalidInput, "Target ability does not exists")
			return
		}

		// check sups
		reason := fmt.Sprintf("battle:%s|game_ability_contribution:%s", fc.API.BattleArena.CurrentBattleID(), req.Payload.GameAbilityID)
		supTransactionReference, err := fc.API.Passport.SendHoldSupsMessage(context.Background(), userID, req.Payload.Amount, req.TransactionID, reason)
		if err != nil {
			targetPriceChan <- ""
			errChan <- terror.Error(err)
			return
		}

		// append transaction ref
		fa.TxRefs = append(fa.TxRefs, supTransactionReference)

		// increase current sups
		fa.CurrentSups.Add(&fa.CurrentSups.Int, &req.Payload.Amount.Int)

		// skip, if current sups is less than target price
		if fa.CurrentSups.Cmp(&fa.TargetPrice.Int) < 0 {
			targetPriceChan <- ""
			errChan <- nil
			return
		}

		// commit all the transactions
		_, err = fc.API.Passport.CommitTransactions(ctx, fa.TxRefs)
		if err != nil {
			targetPriceChan <- ""
			errChan <- terror.Error(err)
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
				targetPriceChan <- ""
				errChan <- terror.Error(err)
				return
			}
		}

		// trigger battle arena function to handle game ability
		abilityTriggerEvent := &server.GameAbilityEvent{
			IsTriggered:         true,
			TriggeredByUserID:   &userID,
			TriggeredByUsername: &hcd.Username,
			GameClientAbilityID: fa.GameAbility.GameClientAbilityID,
			ParticipantID:       fa.GameAbility.ParticipantID,
		}
		if fa.GameAbility.AbilityTokenID == 0 {
			abilityTriggerEvent.GameAbilityID = &fa.GameAbility.ID
		} else {
			abilityTriggerEvent.AbilityTokenID = &fa.GameAbility.AbilityTokenID
		}
		err = fc.API.BattleArena.GameAbilityTrigger(abilityTriggerEvent)
		if err != nil {
			targetPriceChan <- ""
			errChan <- terror.Error(err)
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
			// record ability triggered event for battle end content
			fc.API.battleEndInfo.BattleEvents = append(fc.API.battleEndInfo.BattleEvents, &BattleEventRecord{
				Type:      server.BattleEventTypeGameAbility,
				CreatedAt: time.Now(),
				Event: &BattleAbilityEventRecord{
					TriggeredByUser: triggeredBy,
					Ability:         ability,
				},
			})
		} else {
			warMachine := fc.API.BattleArena.GetWarMachine(fa.GameAbility.WarMachineTokenID).Brief()
			// broadcast notification
			go fc.API.BroadcastGameNotificationWarMachineAbility(ctx, &GameNotificationWarMachineAbility{
				User:       hcd.Brief(),
				Ability:    fa.GameAbility.Brief(),
				WarMachine: fc.API.BattleArena.GetWarMachine(fa.GameAbility.WarMachineTokenID).Brief(),
			})

			// record ability triggered event for battle end content
			fc.API.battleEndInfo.BattleEvents = append(fc.API.battleEndInfo.BattleEvents, &BattleEventRecord{
				Type:      server.BattleEventTypeGameAbility,
				CreatedAt: time.Now(),
				Event: &BattleAbilityEventRecord{
					TriggeredByUser:       triggeredBy,
					Ability:               ability,
					TriggeredOnWarMachine: warMachine,
				},
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

		targetPriceChan <- strings.Join(targetPriceList, "|")
		errChan <- nil
	}

	// wait for target price change
	targetPrice := <-targetPriceChan

	// wait for error check
	err = <-errChan
	if err != nil {
		return terror.Error(err)
	}

	// store vote amount to live voting data after vote success
	fc.API.liveSupsSpend[hcd.FactionID] <- func(lvd *LiveVotingData) {
		lvd.TotalVote.Add(&lvd.TotalVote.Int, &req.Payload.Amount.Int)
	}

	// broadcast if target price is updated
	if targetPrice != "" {
		// prepare broadcast payload
		payload := []byte{}
		payload = append(payload, byte(battle_arena.NetMessageTypeAbilityTargetPriceTick))
		payload = append(payload, []byte(targetPrice)...)
		// start broadcast
		fc.API.Hub.Clients(func(clients hub.ClientsList) {
			for client, ok := range clients {
				if !ok {
					continue
				}
				go func(c *hub.Client) {
					// get user faction id
					hcd, err := fc.API.getClientDetailFromChannel(c)
					// skip, if error or not current faction player
					if err != nil || hcd.FactionID != factionID {
						return
					}

					// broadcast vote price forecast
					err = c.SendWithMessageType(ctx, payload, websocket.MessageBinary)
					if err != nil {
						fc.API.Log.Err(err).Msg("failed to send broadcast")
					}
				}(client)
			}
		})
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

	fc.API.gameAbilityPool[hcd.FactionID] <- func(fap GameAbilitiesPool, fapt *GameAbilityPoolTicker) {
		abilities := []*server.GameAbility{}
		for _, fa := range fap {
			if fa.GameAbility.AbilityTokenID > 0 {
				continue
			}
			fa.GameAbility.CurrentSups = fa.CurrentSups.String()
			abilities = append(abilities, fa.GameAbility)
		}
		reply(abilities)
	}

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
	req := &WarMachineAbilitiesUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	hcd, err := fc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return "", "", terror.Error(err)
	}

	fc.API.gameAbilityPool[hcd.FactionID] <- func(fap GameAbilitiesPool, fapt *GameAbilityPoolTicker) {
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
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s:%x", HubKeyWarMachineAbilitiesUpdated, hcd.FactionID, req.Payload.ParticipantID))

	return req.TransactionID, busKey, nil
}
