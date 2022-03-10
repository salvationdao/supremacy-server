package battle

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/gamelog"

	"github.com/gofrs/uuid"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

type WarMachineDestroyedEventRecord struct {
	DestroyedWarMachine *server.WarMachineBrief `json:"destroyedWarMachine"`
	KilledByWarMachine  *server.WarMachineBrief `json:"killedByWarMachineID,omitempty"`
	KilledBy            string                  `json:"killedBy"`
}

/**********************
*  Game Notification  *
**********************/
type GameNotificationType string

const (
	GameNotificationTypeText                GameNotificationType = "TEXT"
	GameNotificationTypeLocationSelect      GameNotificationType = "LOCATION_SELECT"
	GameNotificationTypeBattleAbility       GameNotificationType = "BATTLE_ABILITY"
	GameNotificationTypeFactionAbility      GameNotificationType = "FACTION_ABILITY"
	GameNotificationTypeWarMachineAbility   GameNotificationType = "WAR_MACHINE_ABILITY"
	GameNotificationTypeWarMachineDestroyed GameNotificationType = "WAR_MACHINE_DESTROYED"
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

const HubKeyMultiplierUpdate hub.HubCommandKey = "USER:SUPS:MULTIPLIER:SUBSCRIBE"

func (arena *Arena) HubKeyMultiplierUpdate(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err)
	}

	if arena.currentBattle == nil {
		return "", "", fmt.Errorf("no active battle")
	}

	id, err := uuid.FromString(wsc.Identifier())
	if err != nil {
		gamelog.L.Warn().Err(err).Str("id", wsc.Identifier()).Msg("unable to create uuid from websocket client identifier id")
		return "", "", fmt.Errorf("no active battle")
	}

	m, total := arena.currentBattle.multipliers.PlayerMultipliers(id)

	reply(&MultiplierUpdate{
		UserMultipliers:  m,
		TotalMultipliers: fmt.Sprintf("%dx", total),
	})

	return req.TransactionID, messagebus.BusKey(HubKeyMultiplierUpdate), nil
}

const HubKeyViewerLiveCountUpdated = hub.HubCommandKey("VIEWER:LIVE:COUNT:UPDATED")

func (arena *Arena) ViewerLiveCountUpdateSubscribeHandler(tx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err)
	}

	userID := server.UserID(uuid.FromStringOrNil(wsc.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}

	return req.TransactionID, messagebus.BusKey(HubKeyViewerLiveCountUpdated), nil
}

const HubKeyGameNotification hub.HubCommandKey = "GAME:NOTIFICATION"

// WinnerAnnouncementSubscribeHandler subscribe on vote winner to pick location
func (arena *Arena) GameNotificationSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err)
	}

	return req.TransactionID, messagebus.BusKey(HubKeyGameNotification), nil
}

// BroadcastGameNotificationText broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationText(data string) {
	arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeyGameNotification), &GameNotification{
		Type: GameNotificationTypeText,
		Data: data,
	})
}

// BroadcastGameNotificationLocationSelect broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationLocationSelect(data *GameNotificationLocationSelect) {
	arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeyGameNotification), &GameNotification{
		Type: GameNotificationTypeLocationSelect,
		Data: data,
	})
}

// BroadcastGameNotificationAbility broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationAbility(notificationType GameNotificationType, data *GameNotificationAbility) {
	arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeyGameNotification), &GameNotification{
		Type: notificationType,
		Data: data,
	})
}

// BroadcastGameNotificationWarMachineAbility broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationWarMachineAbility(data *GameNotificationWarMachineAbility) {
	arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeyGameNotification), &GameNotification{
		Type: GameNotificationTypeWarMachineAbility,
		Data: data,
	})
}

// BroadcastGameNotificationWarMachineDestroyed broadcast game notification to client
func (arena *Arena) BroadcastGameNotificationWarMachineDestroyed(data *WarMachineDestroyedEventRecord) {
	arena.messageBus.Send(context.Background(), messagebus.BusKey(HubKeyGameNotification), &GameNotification{
		Type: GameNotificationTypeWarMachineDestroyed,
		Data: data,
	})
}
