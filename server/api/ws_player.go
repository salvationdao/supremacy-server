package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"server/db/boiler"
	"server/gamedb"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/types"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type PlayerController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewPlayerController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *PlayerController {
	pctrlr := &PlayerController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "player_ctrlr"),
		API:  api,
	}

	api.SecureUserCommand(HubKeyPlayerBattleQueueNotifications, pctrlr.PlayerBattleQueueNotificationsHandler)

	return pctrlr
}

type PlayerBattleQueueNotificationsReq struct {
	*hub.HubCommandRequest
	Payload struct {
		Key   string     `json:"key"`
		Value types.JSON `json:"value,omitempty"`
	} `json:"payload"`
}

const HubKeyPlayerBattleQueueNotifications hub.HubCommandKey = "PLAYER:TOGGLE_BATTLE_QUEUE_NOTIFICATIONS"

func (ctrlr *PlayerController) PlayerBattleQueueNotificationsHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Issue getting settings, try again or contact support."
	req := &PlayerBattleQueueNotificationsReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// getting user's notification settings from the database
	userSettings, err := boiler.FindPlayerPreference(gamedb.StdConn, player.ID, req.Payload.Key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// if there are no results, make an entry for the user with base settings
			playerPrefs := &boiler.PlayerPreference{
				PlayerID:  wsc.Identifier(),
				Key:       req.Payload.Key,
				Value:     []byte{},
				CreatedAt: time.Now()}
			err := playerPrefs.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				return terror.Error(err, errMsg)
			}
		} else {
			return terror.Error(err, errMsg)
		}
	}

	payloadStr := req.Payload.Value.String()
	dbStr := strings.ReplaceAll(userSettings.Value.String(), " ", "")

	// if the payload includes a new value, update it in the db
	if payloadStr != dbStr {
		userSettings.Value = []byte(req.Payload.Value)
		_, err := userSettings.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerPreferenceColumns.Value))
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}

	reply(userSettings)
	return nil
}
