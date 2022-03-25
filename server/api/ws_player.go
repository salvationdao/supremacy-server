package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
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
	api.SecureUserSubscribeCommand(HubKeyPlayerPreferencesSubscribe, pctrlr.PlayerPreferencesSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyPlayerBattleQueueBrowserSubscribe, pctrlr.PlayerBattleQueueBrowserSubscribeHandler)

	return pctrlr
}

const HubKeyPlayerPreferencesSubscribe hub.HubCommandKey = "PLAYER:PREFERENCES_SUBSCRIBE"

func (ctrlr *PlayerController) PlayerPreferencesSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	playerPrefs, err := boiler.PlayerPreferences(boiler.PlayerPreferenceWhere.PlayerID.EQ(wsc.Identifier())).One(gamedb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			playerPrefs := &boiler.PlayerPreference{PlayerID: wsc.Identifier()}
			err := playerPrefs.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				return "", "", terror.Error(err, "Issue getting settings, try again or contact support.")
			}
		} else {
			return "", "", terror.Error(err, "Issue getting setting, try again or contact support.")
		}
	}
	reply(playerPrefs)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPlayerPreferencesSubscribe, wsc.Identifier())), nil
}

type PlayerBattleQueueNotificationsReq struct {
	*hub.HubCommandRequest
	Payload struct {
		BattleQueueSMS     bool `json:"battle_queue_sms"`
		BattleQueueBrowser bool `json:"battle_queue_browser"`
	} `json:"payload"`
}

const HubKeyPlayerBattleQueueNotifications hub.HubCommandKey = "PLAYER:TOGGLE_BATTLE_QUEUE_NOTIFICATIONS"

func (ctrlr *PlayerController) PlayerBattleQueueNotificationsHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &PlayerBattleQueueNotificationsReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to updated preference, try again or contact support.")
	}

	playerPrefs, err := boiler.PlayerPreferences(boiler.PlayerPreferenceWhere.PlayerID.EQ(wsc.Identifier())).One(gamedb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			playerPrefs := &boiler.PlayerPreference{PlayerID: wsc.Identifier()}
			err := playerPrefs.Insert(gamedb.StdConn, boil.Infer())
			if err != nil {
				return terror.Error(err, "Issue updating setting, try again or contact support.")
			}
		} else {
			return terror.Error(err, "Issue updating setting, try again or contact support.")
		}
	}

	updateFields := []string{}
	if playerPrefs.NotificationsBattleQueueSMS != req.Payload.BattleQueueSMS {
		if !player.MobileNumber.Valid {
			return terror.Warn(fmt.Errorf("no mobile set"), "Set your mobile number on Passport to enable this feature.")
		}
		playerPrefs.NotificationsBattleQueueSMS = req.Payload.BattleQueueSMS
		updateFields = append(updateFields, boiler.PlayerPreferenceColumns.NotificationsBattleQueueSMS)
	}
	if playerPrefs.NotificationsBattleQueueBrowser != req.Payload.BattleQueueBrowser {
		playerPrefs.NotificationsBattleQueueBrowser = req.Payload.BattleQueueBrowser
		updateFields = append(updateFields, boiler.PlayerPreferenceColumns.NotificationsBattleQueueBrowser)
	}

	if len(updateFields) > 0 {
		_, err = playerPrefs.Update(gamedb.StdConn, boil.Whitelist(updateFields...))
		if err != nil {
			return terror.Error(err, "Issue updating setting, try again or contact support.")
		}
	}

	go ctrlr.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPlayerPreferencesSubscribe, wsc.Identifier())), playerPrefs)
	reply(true)
	return nil
}

const HubKeyPlayerBattleQueueBrowserSubscribe hub.HubCommandKey = "PLAYER:BROWSER_NOFTICATION_SUBSCRIBE"

func (ctrlr *PlayerController) PlayerBattleQueueBrowserSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPlayerBattleQueueBrowserSubscribe, wsc.Identifier())), nil
}
