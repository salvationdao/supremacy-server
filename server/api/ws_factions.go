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

	api.SecureUserFactionCommand(HubKeyFactionAbilityContribute, factionHub.FactionAbilityContribute)

	// subscription
	api.SecureUserFactionSubscribeCommand(HubKeyFactionAbilitiesUpdated, factionHub.FactionAbilitiesUpdateSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyFactionWarMachineQueueUpdated, factionHub.FactionWarMachineQueueUpdateSubscribeHandler)

	return factionHub
}

type FactionAbilityContributeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionAbilityID server.FactionAbilityID `json:"factionAbilityID"`
		Amount           server.BigInt           `json:"amount"`
	} `json:"payload"`
}

const HubKeyFactionAbilityContribute hub.HubCommandKey = "FACTION:ABILITY:CONTRIBUTE"

func (fc *FactionControllerWS) FactionAbilityContribute(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &FactionAbilityContributeRequest{}
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
	fc.API.factionAbilityPool[factionID] <- func(fap FactionAbilitiesPool, fapt *FactionAbilityPoolTicker) {
		// find ability
		fa, ok := fap[req.Payload.FactionAbilityID]
		if !ok {
			targetPriceChan <- ""
			errChan <- terror.Error(terror.ErrInvalidInput, "Target ability does not exists")
			return
		}

		// check sups
		reason := fmt.Sprintf("battle:%s|faction_ability_contribution:%s", fc.API.BattleArena.CurrentBattleID(), req.Payload.FactionAbilityID)
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
		fa.TxRefs = []server.TransactionReference{}

		// double the target price
		fa.TargetPrice.Mul(&fa.TargetPrice.Int, big.NewInt(2))

		// reset current sups to zero
		fa.CurrentSups = server.BigInt{Int: *big.NewInt(0)}

		// update sups cost of the ability in db
		fa.FactionAbility.SupsCost = fa.TargetPrice.String()
		err = db.FactionExclusiveAbilitiesSupsCostUpdate(ctx, fc.Conn, fa.FactionAbility)
		if err != nil {
			targetPriceChan <- ""
			errChan <- terror.Error(err)
			return
		}

		// trigger battle arena function to handle faction ability
		userIDStr := userID.String()
		err = fc.API.BattleArena.FactionAbilityTrigger(&battle_arena.AbilityTriggerRequest{
			FactionID:           fa.FactionAbility.FactionID,
			FactionAbilityID:    fa.FactionAbility.ID,
			IsSuccess:           true,
			TriggeredByUserID:   &userIDStr,
			TriggeredByUsername: &hcd.Username,
			GameClientAbilityID: fa.FactionAbility.GameClientAbilityID,
		})
		if err != nil {
			targetPriceChan <- ""
			errChan <- terror.Error(err)
			return
		}

		// broadcast notification
		go fc.API.BroadcastGameNotification(GameNotificationTypeText, fmt.Sprintf("User %s from %s has trigger %s", hcd.Username, fc.API.factionMap[hcd.FactionID].Label, fa.FactionAbility.Label))

		// prepare broadcast data
		targetPriceList := []string{}
		for abilityID, fa := range fap {
			hasTriggered := 0
			if abilityID == req.Payload.FactionAbilityID {
				hasTriggered = 1
			}
			targetPriceList = append(targetPriceList, fmt.Sprintf("%s_%s_%s_%d", fa.FactionAbility.ID, fa.TargetPrice.String(), fa.CurrentSups.String(), hasTriggered))
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
					err = c.SendWithMessageType(payload, websocket.MessageBinary)
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

	fc.API.factionAbilityPool[hcd.FactionID] <- func(fap FactionAbilitiesPool, fapt *FactionAbilityPoolTicker) {
		abilities := []*server.FactionAbility{}
		for _, fa := range fap {
			fa.FactionAbility.CurrentSups = fa.CurrentSups.String()
			abilities = append(abilities, fa.FactionAbility)
		}
		reply(abilities)
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionAbilitiesUpdated, hcd.FactionID))

	return req.TransactionID, busKey, nil

}

const HubKeyFactionWarMachineQueueUpdated hub.HubCommandKey = "FACTION:WAR:MACHINE:QUEUE:UPDATED"

// FactionWarMachineQueueUpdateSubscribeHandler
func (fc *FactionControllerWS) FactionWarMachineQueueUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	// get hub client
	hcd, err := fc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return "", "", terror.Error(err)
	}

	if battleQueue, ok := fc.API.BattleArena.BattleQueueMap[hcd.FactionID]; ok {
		battleQueue <- func(wmql *battle_arena.WarMachineQueuingList) {
			maxLength := 5
			if len(wmql.WarMachines) < maxLength {
				maxLength = len(wmql.WarMachines)
			}

			reply(wmql.WarMachines[:maxLength])
		}
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionWarMachineQueueUpdated, hcd.FactionID))

	return req.TransactionID, busKey, nil
}
