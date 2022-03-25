package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type PlayerController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewPlayerController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *PlayerController {
	pc := &PlayerController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "player_controller"),
		API:  api,
	}

	api.SecureUserCommand(HubKeyPlayerBattleQueueNotifications, pc.PlayerBattleQueueNotificationsHandler)
	api.SecureUserSubscribeCommand(HubKeyPlayerPreferencesSubscribe, pc.PlayerPreferencesSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyPlayerBattleQueueBrowserSubscribe, pc.PlayerBattleQueueBrowserSubscribeHandler)

	// faction lose select privilege
	api.SecureUserFactionCommand(HubKeyIssueBanVote, pc.IssueBanVote)
	api.SecureUserFactionCommand(HubKeyBanVote, pc.BanVote)
	api.SecureUserFactionSubscribeCommand(HubKeyBanVoteSubscribe, pc.BanVoteSubscribeHandler)

	return pc
}

const HubKeyPlayerPreferencesSubscribe hub.HubCommandKey = "PLAYER:PREFERENCES_SUBSCRIBE"

func (pc *PlayerController) PlayerPreferencesSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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

func (pc *PlayerController) PlayerBattleQueueNotificationsHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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

	go pc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPlayerPreferencesSubscribe, wsc.Identifier())), playerPrefs)
	reply(true)
	return nil
}

const HubKeyPlayerBattleQueueBrowserSubscribe hub.HubCommandKey = "PLAYER:BROWSER_NOFTICATION_SUBSCRIBE"

func (pc *PlayerController) PlayerBattleQueueBrowserSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPlayerBattleQueueBrowserSubscribe, wsc.Identifier())), nil
}

type BanVoteRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		BanVoteID string `json:"ban_vote_id"`
		IsAgreed  bool   `json:"is_agreed"`
	} `json:"payload"`
}

const HubKeyBanVote hub.HubCommandKey = "BAN:VOTE"

func (pc *PlayerController) BanVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &BanVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// check player is available to be banned
	_, err = boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to get current player from db")
	}

	return nil
}

type BanType string

const (
	BanTypeTeamKill = "TEAM_KILL"
	// will add more banning type in the future
)

type IssueBanVoteRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		IntendToBanPlayerID uuid.UUID `json:"intend_to_ban_player_id"`
		BanType             BanType   `json:"ban_type"`
		Reason              string    `json:"reason"`
	} `json:"payload"`
}

const HubKeyIssueBanVote hub.HubCommandKey = "ISSUE:BAN:VOTE"

func (pc *PlayerController) IssueBanVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &IssueBanVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// check player is available to be banned
	// get players
	currentPlayer, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to get current player from db")
	}

	intendToBenPlayer, err := boiler.FindPlayer(gamedb.StdConn, req.Payload.IntendToBanPlayerID.String())
	if err != nil {
		return terror.Error(err, "Failed to get intend to ban player from db")
	}

	if !intendToBenPlayer.FactionID.Valid || intendToBenPlayer.FactionID.String != currentPlayer.FactionID.String {
		return terror.Error(fmt.Errorf("unable to ban player who is not in your faction"), "Unable to ban player who is not in your faction")
	}

	// if the player is already in ban period
	bannedPlayer, err := boiler.BannedPlayers(
		boiler.BannedPlayerWhere.ID.EQ(req.Payload.IntendToBanPlayerID.String()),
		boiler.BannedPlayerWhere.BanUntil.GT(time.Now()),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to get the banned player from db")
	}

	if bannedPlayer != nil {
		return terror.Error(fmt.Errorf("Player is already banned"), fmt.Sprintf("The player is already banned for %s", bannedPlayer.BanType))
	}

	// check the player is reported
	banVote, err := boiler.BanVotes(
		boiler.BanVoteWhere.ReportedPlayerID.EQ(req.Payload.IntendToBanPlayerID.String()),
		boiler.BanVoteWhere.Status.EQ(BanVoteStatusPending),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to check ban vote from db")
	}

	if banVote != nil {
		return terror.Error(fmt.Errorf("Player is already reported"), fmt.Sprintf("The player has a pending banning report issued by %s", banVote.IssuedByUsername))
	}

	// get the highest price
	price := currentPlayer.IssueBanFee
	// if the reported cost is higher than issue fee, change price to report cost
	if intendToBenPlayer.ReportedCost.GreaterThan(price) {
		price = intendToBenPlayer.ReportedCost
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, "Failed to start a db transaction")
	}

	defer func() {
		err = tx.Rollback()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("Failed to rollback db")
		}
	}()

	// issue a ban vote
	banVote = &boiler.BanVote{
		Type:                   string(req.Payload.BanType),
		Reason:                 req.Payload.Reason,
		FactionID:              currentPlayer.FactionID.String,
		IssuedByID:             currentPlayer.ID,
		IssuedByUsername:       currentPlayer.Username.String,
		ReportedPlayerID:       intendToBenPlayer.ID,
		ReportedPlayerUsername: intendToBenPlayer.Username.String,
		Status:                 BanVoteStatusPending,
	}
	err = banVote.Insert(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to issue a ban vote")
	}

	// double the issue fee of current user
	currentPlayer.IssueBanFee = currentPlayer.IssueBanFee.Mul(decimal.NewFromInt(2))

	_, err = currentPlayer.Update(tx, boil.Whitelist(boiler.PlayerColumns.IssueBanFee))
	if err != nil {
		return terror.Error(err, "Failed to update issue ban fee")
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed to commit db transaction")
	}

	return nil
}

type BanVoteStatus string

const (
	BanVoteStatusPassed  = "PASSED"
	BanVoteStatusFailed  = "FAILED"
	BanVoteStatusPending = "PENDING"
)

const HubKeyBanVoteSubscribe hub.HubCommandKey = "BAN:VOTE:SUBSCRIBE"

func (pc *PlayerController) BanVoteSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	// get player
	player, err := boiler.FindPlayer(gamedb.StdConn, client.Identifier())
	if err != nil {
		return "", "", terror.Error(err, "Failed to get player from db")
	}

	if !player.FactionID.Valid {
		return "", "", terror.Error(fmt.Errorf("player should join faction to subscribe on ban vote"), "Player should join a faction to subscribe on ban vote")
	}

	// get current ongoing ban vote
	// pending status with started_at and ended_at is set
	bv, err := boiler.BanVotes(
		boiler.BanVoteWhere.FactionID.EQ(player.FactionID.String),
		boiler.BanVoteWhere.Status.EQ(BanVoteStatusPending),
		boiler.BanVoteWhere.StartedAt.IsNotNull(),
		boiler.BanVoteWhere.EndedAt.IsNotNull(),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", "", terror.Error(err, "Failed to get ongoing ban vote from db")
	}

	if bv != nil {
		reply(bv)
	}
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBanVoteSubscribe, player.FactionID.String)), nil
}
