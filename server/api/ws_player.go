package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"strings"
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
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	api.SecureUserFactionCommand(HubKeyPlayerSearch, pc.FactionPlayerSearch)
	api.SecureUserFactionCommand(HubKeyPunishVote, pc.PunishVote)
	api.SecureUserFactionCommand(HubKeyPunishOptions, pc.PunishOptions)
	api.SecureUserFactionCommand(HubKeyIssuePunishVote, pc.IssuePunishVote)
	api.SecureUserFactionSubscribeCommand(HubKeyPunishVoteSubscribe, pc.PunishVoteSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyPunishVoteResultSubscribe, pc.PunishVoteResultSubscribeHandler)

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

	go pc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPlayerPreferencesSubscribe, wsc.Identifier())), playerPrefs)
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

type PlayerSearchRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Search string `json:"search"`
	} `json:"payload"`
}

const HubKeyPlayerSearch hub.HubCommandKey = "PLAYER:SEARCH"

// FactionPlayerSearch return up to 5 players base on the given text
func (pc *PlayerController) FactionPlayerSearch(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &PlayerSearchRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	search := strings.TrimSpace(req.Payload.Search)
	if search == "" {
		return terror.Error(terror.ErrInvalidInput, "search key should not be empty")
	}

	player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to player from db")
	}

	ps, err := boiler.Players(
		qm.Select(
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.Username,
		),
		boiler.PlayerWhere.FactionID.EQ(player.FactionID),
		qm.Where(
			fmt.Sprintf("LOWER(%s) LIKE ?", qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Username)),
			"%"+strings.ToLower(search)+"%",
		),
		qm.Limit(5),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to search players from db")
	}

	reply(ps)
	return nil
}

type PunishVoteRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		PunishVoteID string `json:"punish_vote_id"`
		IsAgreed     bool   `json:"is_agreed"`
	} `json:"payload"`
}

const HubKeyPunishVote hub.HubCommandKey = "PUNISH:VOTE"

func (pc *PlayerController) PunishVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &PunishVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// check player is available to be punished
	player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to get current player from db")
	}

	playerStat, err := boiler.FindUserStat(gamedb.StdConn, player.ID)
	if err != nil {
		return terror.Error(err, "Failed to get user stat from db")
	}

	if playerStat.KillCount <= 0 {
		return terror.Error(fmt.Errorf("Only players with positive ability kill count has the right"))
	}

	if pc.API.FactionPunishVote[player.FactionID.String].Stage.Phase != PunishVotePhaseVoting && pc.API.FactionPunishVote[player.FactionID.String].PunishVoteID != req.Payload.PunishVoteID {
		return terror.Error(terror.ErrInvalidInput, "Incorrect vote phase or vote id")
	}

	// send vote into channel
	pc.API.FactionPunishVote[player.FactionID.String].VoteChan <- &PunishVote{
		PunishVoteID: req.Payload.PunishVoteID,
		playerID:     player.ID,
		IsAgreed:     req.Payload.IsAgreed,
	}

	return nil
}

const HubKeyPunishOptions hub.HubCommandKey = "PUNISH:OPTIONS"

func (pc *PlayerController) PunishOptions(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	bts, err := boiler.PunishOptions().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get punish options from db")
	}

	reply(bts)

	return nil
}

type IssuePunishVoteRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		IntendToPunishPlayerID uuid.UUID `json:"intend_to_punish_player_id"`
		PunishOptionID         string    `json:"punish_option_id"`
		Reason                 string    `json:"reason"`
	} `json:"payload"`
}

const HubKeyIssuePunishVote hub.HubCommandKey = "ISSUE:PUNISH:VOTE"

func (pc *PlayerController) IssuePunishVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &IssuePunishVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// check player is available to be punished
	// get players
	currentPlayer, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to get current player from db")
	}

	intendToBenPlayer, err := boiler.FindPlayer(gamedb.StdConn, req.Payload.IntendToPunishPlayerID.String())
	if err != nil {
		return terror.Error(err, "Failed to get intend to punish player from db")
	}

	if !intendToBenPlayer.FactionID.Valid || intendToBenPlayer.FactionID.String != currentPlayer.FactionID.String {
		return terror.Error(fmt.Errorf("unable to punish player who is not in your faction"), "Unable to punish player who is not in your faction")
	}

	// ensure punish vote is issued synchroniously in faction
	pc.API.FactionPunishVote[currentPlayer.FactionID.String].Lock()
	defer pc.API.FactionPunishVote[currentPlayer.FactionID.String].Unlock()

	// if the player is already in punish period
	punishedPlayer, err := boiler.PunishedPlayers(
		boiler.PunishedPlayerWhere.ID.EQ(req.Payload.IntendToPunishPlayerID.String()),
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		qm.Load(boiler.PunishedPlayerRels.PunishOption),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to get the punished player from db")
	}

	if punishedPlayer != nil {
		return terror.Error(fmt.Errorf("Player is already punished"), fmt.Sprintf("The player is already punished for %s", punishedPlayer.R.PunishOption.Key))
	}

	// get punish type
	punishOption, err := boiler.FindPunishOption(gamedb.StdConn, req.Payload.PunishOptionID)
	if err != nil {
		return terror.Error(err, "Failed to get punish type from db")
	}

	// check the player is reported
	punishVote, err := boiler.PunishVotes(
		boiler.PunishVoteWhere.ReportedPlayerID.EQ(req.Payload.IntendToPunishPlayerID.String()),
		boiler.PunishVoteWhere.Status.EQ(string(PunishVoteStatusPending)),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to check punish vote from db")
	}

	if punishVote != nil {
		return terror.Error(fmt.Errorf("Player is already reported"), fmt.Sprintf("The player has a pending punishning report issued by %s", punishVote.IssuedByUsername))
	}

	// get the highest price
	price := currentPlayer.IssuePunishFee
	// if the reported cost is higher than issue fee, change price to report cost
	if intendToBenPlayer.ReportedCost.GreaterThan(price) {
		price = intendToBenPlayer.ReportedCost
	}

	// pay sups to syndicate
	userID := uuid.FromStringOrNil(currentPlayer.ID)

	factionAccountID, ok := server.FactionUsers[currentPlayer.FactionID.String]
	if !ok {
		gamelog.L.Error().
			Str("player id", currentPlayer.ID).
			Str("faction ID", currentPlayer.FactionID.String).
			Err(fmt.Errorf("Failed to get hard coded syndicate player id")).
			Msg("unable to get hard coded syndicate player ID from faction ID")
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

	// issue a punish vote
	punishVote = &boiler.PunishVote{
		PunishOptionID:         punishOption.ID,
		Reason:                 req.Payload.Reason,
		FactionID:              currentPlayer.FactionID.String,
		IssuedByID:             currentPlayer.ID,
		IssuedByUsername:       currentPlayer.Username.String,
		ReportedPlayerID:       intendToBenPlayer.ID,
		ReportedPlayerUsername: intendToBenPlayer.Username.String,
		Status:                 string(PunishVoteStatusPending),
	}
	err = punishVote.Insert(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to issue a punish vote")
	}

	// double the issue fee of current user
	currentPlayer.IssuePunishFee = currentPlayer.IssuePunishFee.Mul(decimal.NewFromInt(2))

	_, err = currentPlayer.Update(tx, boil.Whitelist(boiler.PlayerColumns.IssuePunishFee))
	if err != nil {
		return terror.Error(err, "Failed to update issue punish fee")
	}

	// pay fee to syndicate
	_, err = pc.API.Passport.SpendSupMessage(rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               price.Mul(decimal.New(1, 18)).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("issue_punish_vote|%s|%d", punishVote.ID, time.Now().UnixNano())),
		Group:                "issue punish vote",
		SubGroup:             string(punishOption.Key),
		Description:          "issue vote for punishning player",
		NotSafe:              true,
	})
	if err != nil {
		gamelog.L.Error().Str("player_id", currentPlayer.ID).Str("amount", price.Mul(decimal.New(1, 18)).String()).Err(err).Msg("Failed to pay sups for issuing player punish vote")
		return terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed to commit db transaction")
	}

	return nil
}

type PunishVoteStatus string

const (
	PunishVoteStatusPassed  PunishVoteStatus = "PASSED"
	PunishVoteStatusFailed  PunishVoteStatus = "FAILED"
	PunishVoteStatusPending PunishVoteStatus = "PENDING"
)

const HubKeyPunishVoteSubscribe hub.HubCommandKey = "PUNISH:VOTE:SUBSCRIBE"

func (pc *PlayerController) PunishVoteSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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
		return "", "", terror.Error(fmt.Errorf("player should join faction to subscribe on punish vote"), "Player should join a faction to subscribe on punish vote")
	}

	// only pass down vote, if there is an ongoing vote
	if fbv, ok := pc.API.FactionPunishVote[player.FactionID.String]; ok && fbv.Stage.Phase == PunishVotePhaseVoting {
		bv, err := boiler.PunishVotes(
			boiler.PunishVoteWhere.ID.EQ(fbv.PunishVoteID),
			qm.Load(boiler.PunishVoteRels.PunishOption),
		).One(gamedb.StdConn)
		if err != nil {
			return "", "", terror.Error(err, "Failed to get punish vote from db")
		}
		reply(&PunishVoteInstance{
			PunishVote:   bv,
			PunishOption: bv.R.PunishOption,
		})
	}
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPunishVoteSubscribe, player.FactionID.String)), nil
}

type PunishVoteResult struct {
	PunishVoteID          string `json:"punish_vote_id"`
	AgreedPlayerNumber    int    `json:"agreed_player_number"`
	DisagreedPlayerNumber int    `json:"disagreed_player_number"`
}

const HubKeyPunishVoteResultSubscribe hub.HubCommandKey = "PUNISH:VOTE:RESULT:SUBSCRIBE"

func (pc *PlayerController) PunishVoteResultSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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
		return "", "", terror.Error(fmt.Errorf("player should join faction to subscribe on punish vote"), "Player should join a faction to subscribe on punish vote")
	}

	// only pass down vote result, if there is an ongoing punish vote
	if fbv, ok := pc.API.FactionPunishVote[player.FactionID.String]; ok && fbv.Stage.Phase == PunishVotePhaseVoting {
		reply(&PunishVoteResult{
			PunishVoteID:          fbv.PunishVoteID,
			AgreedPlayerNumber:    len(fbv.AgreedPlayerIDs),
			DisagreedPlayerNumber: len(fbv.DisagreedPlayerIDs),
		})
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPunishVoteResultSubscribe, player.FactionID.String)), nil
}
