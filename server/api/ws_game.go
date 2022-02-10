package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"

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

type GameNotificationType string

const (
	GameNotificationTypeText              GameNotificationType = "TEXT"
	GameNotificationTypeLocationSelect    GameNotificationType = "LOCATION_SELECT"
	GameNotificationTypeBattleAbility     GameNotificationType = "BATTLE_ABILITY"
	GameNotificationTypeFactionAbility    GameNotificationType = "FACTION_ABILITY"
	GameNotificationTypeWarMachineAbility GameNotificationType = "WAR_MACHINE_ABILITY"
)

type WarMachineBrief struct {
	ImageUrl string        `json:"image"`
	Name     string        `json:"name"`
	Faction  *FactionBrief `json:"faction"`
}

type AbilityBrief struct {
	Label    string `json:"label"`
	ImageUrl string `json:"imageUrl"`
	Colour   string `json:"colour"`
}

type UserBrief struct {
	Username string         `json:"username"`
	AvatarID *server.BlobID `json:"avatarID,omitempty"`
	Faction  *FactionBrief  `json:"faction"`
}

type FactionBrief struct {
	Label      string               `json:"label"`
	LogoBlobID server.BlobID        `json:"logoBlobID,omitempty"`
	Theme      *server.FactionTheme `json:"theme"`
}

type GameNotificationKill struct {
	DestroyedWarMachine *WarMachineBrief `json:"DestroyedWarMachine"`
	KillerWarMachine    *WarMachineBrief `json:"killerWarMachine,omitempty"`
	KilledByAbility     *AbilityBrief    `json:"killedByAbility,omitempty"`
}

type LocationSelectType string

const (
	LocationSelectTypeFailed    = "FAILED"
	LocationSelectTypeCancelled = "CANCELLED"
	LocationSelectTypeTrigger   = "TRIGGER"
)

type GameNotificationLocationSelect struct {
	Type        LocationSelectType `json:"type"`
	X           *int               `json:"x,omitempty"`
	Y           *int               `json:"y,omitempty"`
	CurrentUser *UserBrief         `json:"CurrentUser,omitempty"`
	NextUser    *UserBrief         `json:"nextUser,omitempty"`
	Ability     *AbilityBrief      `json:"ability,omitempty"`
	Reason      string             `json:"reason"` // announce failed to pick reason
}

type GameNotificationAbility struct {
	User    *UserBrief    `json:"user"`
	Ability *AbilityBrief `json:"ability,omitempty"`
}

type GameNotificationWarMachineAbility struct {
	User       *UserBrief       `json:"user"`
	Ability    *AbilityBrief    `json:"ability,omitempty"`
	WarMachine *WarMachineBrief `json:"WarMachine"`
}

type GameNotification struct {
	Type GameNotificationType `json:"type"`
	Data interface{}          `json:"data"`
}

const HubKeyGameNotification hub.HubCommandKey = "GAME:NOTIFICATION"

// BroadcastGameNotificationText broadcast game notification to client
func (api *API) BroadcastGameNotificationText(data string) {
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

	api.clientBroadcast(broadcastData)
}

// BroadcastGameNotificationLocationSelect broadcast game notification to client
func (api *API) BroadcastGameNotificationLocationSelect(data *GameNotificationLocationSelect) {
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
	api.clientBroadcast(broadcastData)
}

// BroadcastGameNotificationAbility broadcast game notification to client
func (api *API) BroadcastGameNotificationAbility(notificationType GameNotificationType, data *GameNotificationAbility) {
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
	api.clientBroadcast(broadcastData)
}

// BroadcastGameNotificationWarMachineAbility broadcast game notification to client
func (api *API) BroadcastGameNotificationWarMachineAbility(data *GameNotificationWarMachineAbility) {
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
	api.clientBroadcast(broadcastData)
}

func (api *API) clientBroadcast(data []byte) {
	api.Hub.Clients(func(clients hub.ClientsList) {
		for client, ok := range clients {
			if !ok {
				continue
			}
			go func(c *hub.Client) {
				err := c.Send(data)
				if err != nil {
					api.Log.Err(err).Msg("failed to send broadcast")
				}
			}(client)
		}
	})
}
