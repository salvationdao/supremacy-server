package api

import (
	"context"
	"encoding/json"
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

	//api.SubscribeCommand(HubKeyBattleEndDetailUpdated, gameHub.BattleEndDetailUpdateSubscribeHandler)
	api.SubscribeCommand(HubKeyAISpawned, gameHub.AISpawnedSubscribeHandler)

	// api.SubscribeCommand(HubKeyGameNotification, gameHub.GameNotificationSubscribeHandler)
	return gameHub
}

const HubKeyAISpawned hub.HubCommandKey = "AI:SPAWNED"

// AISpawnedSubscribeHandler to subscribe on war machine destroyed
func (gc *GameControllerWS) AISpawnedSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &struct {
		*hub.HubCommandRequest
	}{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	busKey := messagebus.BusKey(HubKeyAISpawned)
	return req.TransactionID, busKey, nil
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
