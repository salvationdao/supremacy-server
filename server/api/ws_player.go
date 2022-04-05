package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
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

	api.SecureUserCommand(HubKeyPlayerUpdateSettings, pctrlr.PlayerUpdateSettingsHandler)
	api.SecureUserCommand(HubKeyPlayerGetSettings, pctrlr.PlayerGetSettingsHandler)
	api.SecureUserSubscribeCommand(HubKeyTelegramShortcodeRegistered, pctrlr.PlayerGetTelegramShortcodeRegistered)

	return pctrlr
}

type PlayerUpdateSettingsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Key   string     `json:"key"`
		Value types.JSON `json:"value,omitempty"`
	} `json:"payload"`
}

const HubKeyPlayerUpdateSettings hub.HubCommandKey = "PLAYER:UPDATE_SETTINGS"

func (ctrlr *PlayerController) PlayerUpdateSettingsHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Issue updating settings, try again or contact support."
	req := &PlayerUpdateSettingsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	//getting user's notification settings from the database
	userSettings, err := boiler.FindPlayerPreference(gamedb.StdConn, player.ID, req.Payload.Key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// if there are no results, make an entry for the user with settings values sent from frontend
			playerPrefs := &boiler.PlayerPreference{
				PlayerID:  wsc.Identifier(),
				Key:       req.Payload.Key,
				Value:     req.Payload.Value,
				CreatedAt: time.Now()}

			err := playerPrefs.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				return terror.Error(err, errMsg)
			}

			reply(playerPrefs.Value)
			return nil
		} else {
			return terror.Error(err, errMsg)
		}
	}

	payloadStr := req.Payload.Value.String()
	dbStr := strings.ReplaceAll(userSettings.Value.String(), " ", "")

	//if the payload includes a new value, update it in the db
	if payloadStr != dbStr {
		userSettings.Value = req.Payload.Value
		_, err := userSettings.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerPreferenceColumns.Value))
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}

	//send back userSettings values
	reply(userSettings.Value)
	return nil
}

type PlayerGetSettingsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Key string `json:"key"`
	} `json:"payload"`
}

const HubKeyPlayerGetSettings hub.HubCommandKey = "PLAYER:GET_SETTINGS"

//gets settings based on key, sends settings value back as json
func (ctrlr *PlayerController) PlayerGetSettingsHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Issue getting settings, try again or contact support."
	req := &PlayerGetSettingsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	//getting user's notification settings from the database
	userSettings, err := boiler.FindPlayerPreference(gamedb.StdConn, player.ID, req.Payload.Key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// if there are no results, return a null json- tells frontend to use default settings
			reply(null.JSON{})
			return nil
		} else {
			return terror.Error(err, errMsg)
		}
	}

	//send back userSettings
	reply(userSettings.Value)
	reply(true)
	return nil
}

const HubKeyTelegramShortcodeRegistered hub.HubCommandKey = "USER:TELEGRAM_SHORTCODE_REGISTERED"

func (ctrlr *PlayerController) PlayerGetTelegramShortcodeRegistered(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	reply(nil)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTelegramShortcodeRegistered, wsc.Identifier())), nil

}
