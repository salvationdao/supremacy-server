package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/battle_arena"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

// GameControllerWS holds handlers for checking server status
type GameControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewGameController creates the check hub
func NewGameController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *GameControllerWS {
	gameHub := &GameControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "game_hub"),
		API:  api,
	}

	api.Command(HubKeyFactionColour, gameHub.FactionColour)
	api.SubscribeCommand(HubKeyWarMachineDestroyedUpdated, gameHub.WarMachineDestroyedUpdateSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyFactionWarMachineQueueUpdated, gameHub.FactionWarMachineQueueUpdateSubscribeHandler)
	api.SubscribeCommand(HubKeyBattleEndDetailUpdated, gameHub.BattleEndDetailUpdateSubscribeHandler)

	return gameHub
}

const HubKeyFactionColour hub.HubCommandKey = "FACTION:COLOUR"

type FactionColourRespose struct {
	RedMountain string `json:"redMountain"`
	Boston      string `json:"boston"`
	Zaibatsu    string `json:"zaibatsu"`
}

func (gc *GameControllerWS) FactionColour(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	if gc.API.factionMap == nil {
		return terror.Error(terror.ErrForbidden, "faction data not ready yet")
	}

	reply(&FactionColourRespose{
		RedMountain: gc.API.factionMap[server.RedMountainFactionID].Theme.Primary,
		Boston:      gc.API.factionMap[server.BostonCyberneticsFactionID].Theme.Primary,
		Zaibatsu:    gc.API.factionMap[server.ZaibatsuFactionID].Theme.Primary,
	})

	return nil
}

type WarMachineDestroyedRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ParticipantID byte `json:"participantID"`
	} `json:"payload"`
}

const HubKeyWarMachineDestroyedUpdated hub.HubCommandKey = "WAR:MACHINE:DESTROYED:UPDATED"

// WarMachineDestroyedUpdateSubscribeHandler to subscribe on war machine destroyed
func (gc *GameControllerWS) WarMachineDestroyedUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &WarMachineDestroyedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	record := gc.API.BattleArena.WarMachineDestroyedRecord(req.Payload.ParticipantID)
	if record != nil {
		reply(record)
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%x", HubKeyWarMachineDestroyedUpdated, req.Payload.ParticipantID))
	return req.TransactionID, busKey, nil
}

const HubKeyFactionWarMachineQueueUpdated hub.HubCommandKey = "FACTION:WAR:MACHINE:QUEUE:UPDATED"

// FactionWarMachineQueueUpdateSubscribeHandler subscribe on war machine queue position change
func (gc *GameControllerWS) FactionWarMachineQueueUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	// get hub client
	hcd := gc.API.UserMap.GetUserDetail(wsc)
	if hcd == nil {
		return "", "", terror.Error(err)
	}

	if battleQueue, ok := gc.API.BattleArena.BattleQueueMap[hcd.FactionID]; ok {
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

type BattleEventRecord struct {
	Type      server.BattleEventType `json:"type"`
	CreatedAt time.Time              `json:"createdAt"`
	Event     interface{}            `json:"event"`
}

type BattleAbilityEventRecord struct {
	Ability               *server.AbilityBrief    `json:"ability"`
	TriggeredByUser       *server.UserBrief       `json:"triggeredByUser,omitempty"`
	TriggeredOnCellX      *int                    `json:"x,omitempty"`
	TriggeredOnCellY      *int                    `json:"y,omitempty"`
	TriggeredOnWarMachine *server.WarMachineBrief `json:"triggeredOnWarMachine,omitempty"`
}

type WarMachineDestroyedEventRecord struct {
	DestroyedWarMachine *server.WarMachineBrief `json:"destroyedWarMachine"`
	KilledByWarMachine  *server.WarMachineBrief `json:"killedByWarMachineID,omitempty"`
	KilledBy            string                  `json:"killedBy"`
}

const HubKeyBattleEndDetailUpdated hub.HubCommandKey = "BATTLE:END:DETAIL:UPDATED"

func (gc *GameControllerWS) BattleEndDetailUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	if gc.API.BattleArena.GetCurrentState().EndedAt != nil {
		reply(gc.API.battleEndInfo)
	}

	return req.TransactionID, messagebus.BusKey(HubKeyBattleEndDetailUpdated), nil
}

/**********************
*  Game Notification  *
**********************/
type GameNotificationType string

const (
	GameNotificationTypeText              GameNotificationType = "TEXT"
	GameNotificationTypeLocationSelect    GameNotificationType = "LOCATION_SELECT"
	GameNotificationTypeBattleAbility     GameNotificationType = "BATTLE_ABILITY"
	GameNotificationTypeFactionAbility    GameNotificationType = "FACTION_ABILITY"
	GameNotificationTypeWarMachineAbility GameNotificationType = "WAR_MACHINE_ABILITY"
)

type GameNotificationKill struct {
	DestroyedWarMachine *server.WarMachineBrief `json:"DestroyedWarMachine"`
	KillerWarMachine    *server.WarMachineBrief `json:"killerWarMachine,omitempty"`
	KilledByAbility     *server.AbilityBrief    `json:"killedByAbility,omitempty"`
}

type LocationSelectType string

const (
	LocationSelectTypeFailedDisconnect    = "FAILED_DISCONNECT"
	LocationSelectTypeFailedTimeout       = "FAILED_TIMEOUT"
	LocationSelectTypeCancelledNoPlayer   = "CANCELLED_NO_PLAYER"
	LocationSelectTypeCancelledDisconnect = "CANCELLED_DISCONNECT"
	LocationSelectTypeTrigger             = "TRIGGER"
)

type GameNotificationLocationSelect struct {
	Type        LocationSelectType   `json:"type"`
	X           *int                 `json:"x,omitempty"`
	Y           *int                 `json:"y,omitempty"`
	CurrentUser *server.UserBrief    `json:"currentUser,omitempty"`
	NextUser    *server.UserBrief    `json:"nextUser,omitempty"`
	Ability     *server.AbilityBrief `json:"ability,omitempty"`
}

type GameNotificationAbility struct {
	User    *server.UserBrief    `json:"user,omitempty"`
	Ability *server.AbilityBrief `json:"ability,omitempty"`
}

type GameNotificationWarMachineAbility struct {
	User       *server.UserBrief       `json:"user,omitempty"`
	Ability    *server.AbilityBrief    `json:"ability,omitempty"`
	WarMachine *server.WarMachineBrief `json:"warMachine,omitempty"`
}

type GameNotification struct {
	Type GameNotificationType `json:"type"`
	Data interface{}          `json:"data"`
}

const HubKeyGameNotification hub.HubCommandKey = "GAME:NOTIFICATION"

// BroadcastGameNotificationText broadcast game notification to client
func (api *API) BroadcastGameNotificationText(ctx context.Context, data string) {
	// broadcast countered notification
	broadcastData, err := json.Marshal(&BroadcastPayload{
		Key: HubKeyGameNotification,
		Payload: &GameNotification{
			Type: GameNotificationTypeText,
			Data: data,
		},
	})
	if err != nil {
		api.Log.Err(err).Msg("marshal broadcast payload")
		return
	}

	api.clientBroadcast(ctx, broadcastData)
}

// BroadcastGameNotificationLocationSelect broadcast game notification to client
func (api *API) BroadcastGameNotificationLocationSelect(ctx context.Context, data *GameNotificationLocationSelect) {
	// broadcast countered notification
	broadcastData, err := json.Marshal(&BroadcastPayload{
		Key: HubKeyGameNotification,
		Payload: &GameNotification{
			Type: GameNotificationTypeLocationSelect,
			Data: data,
		},
	})
	if err != nil {
		api.Log.Err(err).Msg("marshal broadcast payload")
		return
	}
	api.clientBroadcast(ctx, broadcastData)
}

// BroadcastGameNotificationAbility broadcast game notification to client
func (api *API) BroadcastGameNotificationAbility(ctx context.Context, notificationType GameNotificationType, data *GameNotificationAbility) {
	// broadcast countered notification
	broadcastData, err := json.Marshal(&BroadcastPayload{
		Key: HubKeyGameNotification,
		Payload: &GameNotification{
			Type: notificationType,
			Data: data,
		},
	})
	if err != nil {
		api.Log.Err(err).Msg("marshal broadcast payload")
		return
	}
	api.clientBroadcast(ctx, broadcastData)
}

// BroadcastGameNotificationWarMachineAbility broadcast game notification to client
func (api *API) BroadcastGameNotificationWarMachineAbility(ctx context.Context, data *GameNotificationWarMachineAbility) {
	// broadcast countered notification
	broadcastData, err := json.Marshal(&BroadcastPayload{
		Key: HubKeyGameNotification,
		Payload: &GameNotification{
			Type: GameNotificationTypeWarMachineAbility,
			Data: data,
		},
	})
	if err != nil {
		api.Log.Err(err).Msg("marshal broadcast payload")
		return
	}
	api.clientBroadcast(ctx, broadcastData)
}

func (api *API) clientBroadcast(ctx context.Context, data []byte) {
	api.Hub.Clients(func(sessionID hub.SessionID, client *hub.Client) bool {
		go client.Send(data)
		return true
	})
}
