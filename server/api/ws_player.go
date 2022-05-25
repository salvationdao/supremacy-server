package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/rpcclient"
	"strings"
	"time"

	"github.com/ninja-syndicate/ws"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type PlayerController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewPlayerController(api *API) *PlayerController {
	pc := &PlayerController{
		API: api,
	}

	api.SecureUserCommand(HubKeyPlayerUpdateSettings, pc.PlayerUpdateSettingsHandler)
	api.SecureUserCommand(HubKeyPlayerGetSettings, pc.PlayerGetSettingsHandler)

	api.SecureUserCommand(HubKeyPlayerPreferencesGet, pc.PlayerPreferencesGetHandler)
	api.SecureUserCommand(HubKeyPlayerPreferencesUpdate, pc.PlayerPreferencesUpdateHandler)

	// punish vote related
	api.SecureUserCommand(HubKeyPlayerPunishmentList, pc.PlayerPunishmentList)
	api.SecureUserCommand(HubKeyPlayerActiveCheck, pc.PlayerActiveCheckHandler)
	api.SecureUserFactionCommand(HubKeyFactionPlayerSearch, pc.FactionPlayerSearch)
	api.SecureUserFactionCommand(HubKeyPunishOptions, pc.PunishOptions)
	api.SecureUserFactionCommand(HubKeyPunishVote, pc.PunishVote)
	api.SecureUserFactionCommand(HubKeyIssuePunishVote, pc.IssuePunishVote)
	api.SecureUserFactionCommand(HubKeyPunishVotePriceQuote, pc.PunishVotePriceQuote)

	api.SecureUserCommand(HubKeyFactionEnlist, pc.PlayerFactionEnlistHandler)

	api.SecureUserCommand(HubKeyPlayerRankGet, pc.PlayerRankGet)

	api.SecureUserCommand(HubKeyGameUserOnline, pc.UserOnline)

	return pc
}

type UserUpdatedRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

const HubKeyUserSubscribe = "USER:SUBSCRIBE"

// FactionEnlistRequest enlist a faction
type FactionEnlistRequest struct {
	Payload struct {
		FactionID string `json:"faction_id"`
	} `json:"payload"`
}

const HubKeyFactionEnlist = "FACTION:ENLIST"

func (pc *PlayerController) PlayerFactionEnlistHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &FactionEnlistRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// check player faction
	if user.FactionID.Valid {
		return terror.Error(fmt.Errorf("player already enlist faction"), "User has already enlisted a faction")
	}

	if req.Payload.FactionID == "" {
		return terror.Error(fmt.Errorf("faction id is empty"), "Faction id is missing")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, "Failed to start db transaction")
	}

	defer tx.Rollback()
	user.FactionID = null.StringFrom(req.Payload.FactionID)
	_, err = user.Update(tx, boil.Whitelist(boiler.PlayerColumns.FactionID))
	if err != nil {
		return terror.Error(err, "Failed to update faction in db")
	}

	// update user faction in passport
	err = pc.API.Passport.UserFactionEnlist(user.ID, user.FactionID.String)
	if err != nil {
		return terror.Error(err, "Failed to sync passport db")
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed to commit db transaction")
	}

	ws.PublishMessage(fmt.Sprintf("/user/%s", user.ID), HubKeyUserSubscribe, user)

	reply(true)

	return nil
}

type PlayerUpdateSettingsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Key   string     `json:"key"`
		Value types.JSON `json:"value,omitempty"`
	} `json:"payload"`
}

const HubKeyPlayerUpdateSettings = "PLAYER:UPDATE_SETTINGS"

func (pc *PlayerController) PlayerUpdateSettingsHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue updating settings, try again or contact support."
	req := &PlayerUpdateSettingsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	//getting user's notification settings from the database
	userSettings, err := boiler.FindPlayerPreference(gamedb.StdConn, user.ID, req.Payload.Key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// if there are no results, make an entry for the user with settings values sent from frontend
			playerPrefs := &boiler.PlayerPreference{
				PlayerID:  user.ID,
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

type PlayerNotificationPreferences struct {
	SMSNotifications      bool `json:"sms_notifications"`
	PushNotifications     bool `json:"push_notifications"`
	TelegramNotifications bool `json:"telegram_notifications"`
}

type PlayerGetSettingsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Key string `json:"key"`
	} `json:"payload"`
}

const HubKeyPlayerGetSettings = "PLAYER:GET_SETTINGS"

//PlayerGetSettingsHandler gets settings based on key, sends settings value back as json
func (pc *PlayerController) PlayerGetSettingsHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue getting settings, try again or contact support."
	req := &PlayerGetSettingsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	//getting user's notification settings from the database
	userSettings, err := boiler.FindPlayerPreference(gamedb.StdConn, user.ID, req.Payload.Key)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// if there are no results, create entry in table and rreturn a null json- tells frontend to use default settings
			playerPrefs := &boiler.PlayerPreference{
				PlayerID:  user.ID,
				Key:       req.Payload.Key,
				CreatedAt: time.Now()}

			playerPrefs.Value.Marshal(PlayerNotificationPreferences{
				SMSNotifications:      false,
				PushNotifications:     false,
				TelegramNotifications: false,
			})

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

	//send back userSettings
	reply(userSettings.Value)
	return nil
}

//const HubKeyTelegramShortcodeRegistered = "USER:TELEGRAM_SHORTCODE_REGISTERED"

func (api *API) PlayerGetTelegramShortcodeRegistered(w http.ResponseWriter, r *http.Request) (int, error) {
	return helpers.EncodeJSON(w, false)
}

const HubKeyPlayerBattleQueueBrowserSubscribe = "PLAYER:BROWSER_NOTIFICATION_SUBSCRIBE"

func (pc *PlayerController) PlayerBattleQueueBrowserSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPlayerBattleQueueBrowserSubscribe, wsc.Identifier())), nil
}

type PlayerPunishment struct {
	*boiler.PunishedPlayer
	RelatedPunishVote *boiler.PunishVote   `json:"related_punish_vote"`
	PunishOption      *boiler.PunishOption `json:"punish_option"`
}

const HubKeyPlayerPunishmentList = "PLAYER:PUNISHMENT:LIST"

func (pc *PlayerController) PlayerPunishmentList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	punishments, err := boiler.PunishedPlayers(
		boiler.PunishedPlayerWhere.PlayerID.EQ(user.ID),
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		qm.Load(boiler.PunishedPlayerRels.PunishOption),
		qm.Load(boiler.PunishedPlayerRels.RelatedPunishVote),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to get player's punishment from db")
		return terror.Error(err, "Failed to get player's punishment from db")
	}

	if punishments == nil || len(punishments) == 0 {
		reply([]*PlayerPunishment{})
		return nil
	}

	playerPunishments := []*PlayerPunishment{}
	for _, punishment := range punishments {
		playerPunishments = append(playerPunishments, &PlayerPunishment{
			PunishedPlayer:    punishment,
			RelatedPunishVote: punishment.R.RelatedPunishVote,
			PunishOption:      punishment.R.PunishOption,
		})
	}

	reply(playerPunishments)

	return nil
}

type PlayerActiveCheckRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Fruit string `json:"fruit"`
	} `json:"payload"`
}

const HubKeyPlayerActiveCheck = "GOJI:BERRY:TEA"

func (pc *PlayerController) PlayerActiveCheckHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerActiveCheckRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	isActive := false
	switch req.Payload.Fruit {
	case "APPLE":
		isActive = true
	case "BANANA":
		isActive = false
	default:
		return terror.Error(terror.ErrInvalidInput, "Invalid active stat")
	}

	// get player

	if user.FactionID.Valid {
		fap, ok := pc.API.FactionActivePlayers[user.FactionID.String]
		if !ok {
			return nil
		}

		err = fap.Set(user.ID, isActive)
		if err != nil {
			return terror.Error(err, "Failed to update player active stat")
		}

		// debounce broadcast active player
		fap.ActivePlayerListChan <- &ActivePlayerBroadcast{
			Players: fap.CurrentFactionActivePlayer(),
		}
	}

	return nil
}

type PlayerSearchRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Search string `json:"search"`
	} `json:"payload"`
}

const HubKeyFactionPlayerSearch = "FACTION:PLAYER:SEARCH"

// FactionPlayerSearch return up to 5 players base on the given text
func (pc *PlayerController) FactionPlayerSearch(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerSearchRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	search := strings.TrimSpace(req.Payload.Search)
	if search == "" {
		return terror.Error(terror.ErrInvalidInput, "search key should not be empty")
	}

	ps, err := boiler.Players(
		qm.Select(
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.Username,
			boiler.PlayerColumns.Gid,
		),
		boiler.PlayerWhere.FactionID.EQ(user.FactionID),
		boiler.PlayerWhere.IsAi.EQ(false),
		boiler.PlayerWhere.ID.NEQ(user.ID),
		qm.Where(
			fmt.Sprintf("LOWER(%s||'#'||%s::TEXT) LIKE ?",
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Username),
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Gid),
			),
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

const HubKeyPunishVote = "PUNISH:VOTE"

func (pc *PlayerController) PunishVote(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PunishVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	us, err := db.UserStatsGet(user.ID)
	if err != nil {
		return terror.Error(err, "Failed to get user stat from db")
	}

	if us.LastSevenDaysKills < 5 && us.AbilityKillCount < 100 {
		return terror.Error(terror.ErrForbidden, "Require at least 5 kills in last 7 days or 100 kills in lifetime to vote")
	}

	// check player is available to be punished
	fpv, ok := pc.API.FactionPunishVote[user.FactionID.String]
	if !ok {
		return terror.Error(fmt.Errorf("player faction id does not exist"))
	}

	err = fpv.Vote(req.Payload.PunishVoteID, user.ID, req.Payload.IsAgreed)
	if err != nil {
		return terror.Error(err, err.Error())
	}

	reply(true)

	return nil
}

const HubKeyPunishOptions = "PUNISH:OPTIONS"

func (pc *PlayerController) PunishOptions(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	bts, err := boiler.PunishOptions().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get punish options from db")
	}

	reply(bts)

	return nil
}

type PunishVotePriceQuoteRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		IntendToPunishPlayerID uuid.UUID `json:"intend_to_punish_player_id"`
	} `json:"payload"`
}

const HubKeyPunishVotePriceQuote = "PUNISH:VOTE:PRICE:QUOTE"

func (pc *PlayerController) PunishVotePriceQuote(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PunishVotePriceQuoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// check player is available to be punished
	intendToBenPlayer, err := boiler.FindPlayer(gamedb.StdConn, req.Payload.IntendToPunishPlayerID.String())
	if err != nil {
		return terror.Error(err, "Failed to get intend to punish player from db")
	}

	if !intendToBenPlayer.FactionID.Valid || intendToBenPlayer.FactionID.String != factionID {
		return terror.Error(fmt.Errorf("unable to punish player who is not in your faction"), "Unable to quote the price of punish vote with a player in other faction")
	}

	// get the highest price
	price := user.IssuePunishFee
	// if the reported cost is higher than issue fee, change price to report cost
	if intendToBenPlayer.ReportedCost.GreaterThan(price) {
		price = intendToBenPlayer.ReportedCost
	}

	reply(price)

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

const HubKeyIssuePunishVote = "ISSUE:PUNISH:VOTE"

func (pc *PlayerController) IssuePunishVote(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &IssuePunishVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	us, err := db.UserStatsGet(user.ID)
	if err != nil {
		return terror.Error(err, "Failed to get user stat from db")
	}

	if us.LastSevenDaysKills < 5 && us.AbilityKillCount < 100 {
		return terror.Error(terror.ErrForbidden, "Require at least 5 kills in last 7 days or 100 kills in lifetime to vote")
	}

	// check player is available to be punished
	intendToBenPlayer, err := boiler.FindPlayer(gamedb.StdConn, req.Payload.IntendToPunishPlayerID.String())
	if err != nil {
		return terror.Error(err, "Failed to get intend to punish player from db")
	}

	if !intendToBenPlayer.FactionID.Valid || intendToBenPlayer.FactionID.String != factionID {
		return terror.Error(fmt.Errorf("unable to punish player who is not in your faction"), "Unable to punish player who is not in your faction")
	}

	if req.Payload.Reason == "" {
		return terror.Error(terror.ErrInvalidInput, "Reason is required")
	}

	// get punish type
	punishOption, err := boiler.FindPunishOption(gamedb.StdConn, req.Payload.PunishOptionID)
	if err != nil {
		return terror.Error(err, "Failed to get punish type from db")
	}

	if _, ok := pc.API.FactionPunishVote[factionID]; !ok {
		gamelog.L.Error().Str("faction id", user.FactionID.String).Err(fmt.Errorf("faction punish vote not found")).Msg("Faction punish vote not found")
		return terror.Error(fmt.Errorf("faction punish vote not found"), "Faction punish vote not found")
	}

	// ensure punish vote is issued synchroniously in faction
	pc.API.FactionPunishVote[factionID].Lock()
	defer pc.API.FactionPunishVote[factionID].Unlock()

	// check player is currently punished with the same option
	punishedPlayer, err := boiler.PunishedPlayers(
		boiler.PunishedPlayerWhere.PlayerID.EQ(req.Payload.IntendToPunishPlayerID.String()),
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		boiler.PunishedPlayerWhere.PunishOptionID.EQ(punishOption.ID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to get the punished player from db")
	}

	if punishedPlayer != nil {
		return terror.Error(fmt.Errorf("player is already punished"), fmt.Sprintf("The player is already punished for %s", punishOption.Key))
	}

	// check player has a pending punish vote with the same option
	punishVote, err := boiler.PunishVotes(
		boiler.PunishVoteWhere.ReportedPlayerID.EQ(req.Payload.IntendToPunishPlayerID.String()),
		boiler.PunishVoteWhere.Status.EQ(string(PunishVoteStatusPending)),
		boiler.PunishVoteWhere.PunishOptionID.EQ(punishOption.ID),
	).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to check punish vote from db")
	}

	if punishVote != nil {
		return terror.Error(fmt.Errorf("player is already reported"), fmt.Sprintf("The player has a pending punishing report issued by %s", punishVote.IssuedByUsername))
	}

	// get the highest price
	price := user.IssuePunishFee
	// if the reported cost is higher than issue fee, change price to report cost
	if intendToBenPlayer.ReportedCost.GreaterThan(price) {
		price = intendToBenPlayer.ReportedCost
	}

	// pay sups to syndicate
	userID := uuid.FromStringOrNil(user.ID)

	factionAccountID, ok := server.FactionUsers[factionID]
	if !ok {
		gamelog.L.Error().
			Str("player id", user.ID).
			Str("faction ID", user.FactionID.String).
			Err(fmt.Errorf("failed to get hard coded syndicate player id")).
			Msg("unable to get hard coded syndicate player ID from faction ID")
	}

	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, "Failed to start a db transaction")
	}

	defer tx.Rollback()

	// issue punish vote
	punishVote = &boiler.PunishVote{
		PunishOptionID:         punishOption.ID,
		Reason:                 req.Payload.Reason,
		FactionID:              factionID,
		IssuedByID:             user.ID,
		IssuedByGid:            user.Gid,
		IssuedByUsername:       user.Username.String,
		ReportedPlayerID:       intendToBenPlayer.ID,
		ReportedPlayerUsername: intendToBenPlayer.Username.String,
		ReportedPlayerGid:      intendToBenPlayer.Gid,
		Status:                 string(PunishVoteStatusPending),
	}
	err = punishVote.Insert(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to issue a punish vote")
	}

	// double the issue fee of current user
	user.IssuePunishFee = user.IssuePunishFee.Mul(decimal.NewFromInt(2))

	_, err = user.Update(tx, boil.Whitelist(boiler.PlayerColumns.IssuePunishFee))
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
		Description:          "issue vote for punishing player",
		NotSafe:              true,
	})
	if err != nil {
		gamelog.L.Error().Str("player_id", user.ID).Str("amount", price.Mul(decimal.New(1, 18)).String()).Err(err).Msg("Failed to pay sups for issuing player punish vote")
		return err
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed to commit db transaction")
	}

	reply(true)

	return nil
}

type PunishVoteStatus string

const (
	PunishVoteStatusPassed  PunishVoteStatus = "PASSED"
	PunishVoteStatusFailed  PunishVoteStatus = "FAILED"
	PunishVoteStatusPending PunishVoteStatus = "PENDING"
)

type PunishVoteResponse struct {
	*boiler.PunishVote
	PunishOption *boiler.PunishOption `json:"punish_option"`
	Decision     *PunishVoteDecision  `json:"decision,omitempty"`
}

type PunishVoteDecision struct {
	IsAgreed bool `json:"is_agreed"`
}

const HubKeyPunishVoteSubscribe = "PUNISH:VOTE:SUBSCRIBE"

func (pc *PlayerController) PunishVoteSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	// only pass down vote, if there is an ongoing vote
	if fpv, ok := pc.API.FactionPunishVote[factionID]; ok {
		fpv.RLock()
		defer fpv.RUnlock()
		if fpv.CurrentPunishVote != nil && fpv.Stage.Phase == PunishVotePhaseVoting {
			bv, err := boiler.PunishVotes(
				boiler.PunishVoteWhere.ID.EQ(fpv.CurrentPunishVote.ID),
				qm.Load(boiler.PunishVoteRels.PunishOption),
			).One(gamedb.StdConn)
			if err != nil {
				return terror.Error(err, "Failed to get punish vote from db")
			}

			pvr := &PunishVoteResponse{
				PunishVote:   bv,
				PunishOption: bv.R.PunishOption,
			}

			// check user has voted
			decision, err := boiler.PlayersPunishVotes(
				boiler.PlayersPunishVoteWhere.PunishVoteID.EQ(bv.ID),
				boiler.PlayersPunishVoteWhere.PlayerID.EQ(user.ID),
			).One(gamedb.StdConn)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return terror.Error(err, "Failed to check player had voted")
			}

			if decision != nil {
				pvr.Decision = &PunishVoteDecision{
					IsAgreed: decision.IsAgreed,
				}
			}

			reply(pvr)
		}
	}

	return nil
}

type PunishVoteResult struct {
	PunishVoteID          string `json:"punish_vote_id"`
	TotalPlayerNumber     int    `json:"total_player_number"`
	AgreedPlayerNumber    int    `json:"agreed_player_number"`
	DisagreedPlayerNumber int    `json:"disagreed_player_number"`
}

const HubKeyPunishVoteResultSubscribe = "PUNISH:VOTE:RESULT:SUBSCRIBE"

const HubKeyFactionActivePlayersSubscribe = "FACTION:ACTIVE:PLAYER:SUBSCRIBE"

func (pc *PlayerController) FactionActivePlayersSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	player, err := boiler.FindPlayer(gamedb.StdConn, user.ID)
	if err != nil {
		return terror.Error(err, "Failed to get player from db")
	}

	fap, ok := pc.API.FactionActivePlayers[player.FactionID.String]
	if !ok {
		return terror.Error(terror.ErrForbidden, "Faction does not exist in the list")
	}

	reply(fap.CurrentFactionActivePlayer())

	return nil
}

func (pc *PlayerController) PlayersSubscribeHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	// broadcast player
	reply(user)

	// broadcast player stat
	us, err := db.UserStatsGet(user.ID)
	if err != nil {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to get player stat")
	}

	if us != nil {
		ws.PublishMessage(fmt.Sprintf("/user/%s", user.ID), battle.HubKeyUserStatSubscribe, us)
	}

	// broadcast player punishment list
	// get current player's punishment
	punishments, err := boiler.PunishedPlayers(
		boiler.PunishedPlayerWhere.PlayerID.EQ(user.ID),
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		qm.Load(boiler.PunishedPlayerRels.PunishOption),
		qm.Load(boiler.PunishedPlayerRels.RelatedPunishVote),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to get player's punishment from db")
		return nil
	}

	if punishments == nil || len(punishments) == 0 {
		return nil
	}

	playerPunishments := []*PlayerPunishment{}
	for _, punishment := range punishments {
		playerPunishments = append(playerPunishments, &PlayerPunishment{
			PunishedPlayer:    punishment,
			RelatedPunishVote: punishment.R.RelatedPunishVote,
			PunishOption:      punishment.R.PunishOption,
		})
	}

	// send to the player
	ws.PublishMessage(fmt.Sprintf("/user/%s", user.ID), HubKeyPlayerPunishmentList, playerPunishments)

	return nil
}

const HubKeyPlayerRankGet = "PLAYER:RANK:GET"

func (pc *PlayerController) PlayerRankGet(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	player, err := boiler.Players(
		qm.Select(
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.Rank,
		),
		boiler.PlayerWhere.ID.EQ(user.ID),
	).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to get player rank from db")
		return terror.Error(err, "Failed to get player rank from db")
	}

	reply(player.Rank)

	return nil
}

const HubKeyGameUserOnline = "GAME:ONLINE"

func (pc *PlayerController) UserOnline(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	if pc.API.BattleArena.CurrentBattle() == nil {
		return nil
	}
	uID, err := uuid.FromString(user.ID)
	if uID.IsNil() || err != nil {
		gamelog.L.Error().Str("uuid", user.ID).Err(err).Msg("invalid input data")
		return fmt.Errorf("unable to construct user uuid")
	}
	userID := server.UserID(uID)

	// TODO: handle faction swap from non-faction to faction
	if !user.FactionID.Valid {
		return nil
	}

	battleUser := &battle.BattleUser{
		ID:        uuid.FromStringOrNil(userID.String()),
		Username:  user.Username.String,
		FactionID: user.FactionID.String,
	}

	reply(pc.API.BattleArena.CurrentBattle().UserOnline(battleUser))

	return nil
}

const HubKeyPlayerPreferencesGet = "PLAYER:PREFERENCES_GET"

// PlayerPreferencesGetHandler gets player's preferences
func (pc *PlayerController) PlayerPreferencesGetHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue getting player preferences, try again or contact support."
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request")

	}

	// try get player's preferences
	prefs, err := boiler.PlayerSettingsPreferences(boiler.PlayerSettingsPreferenceWhere.PlayerID.EQ(user.ID)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, errMsg)
	}

	// if there are no results, create new player preferences
	if errors.Is(err, sql.ErrNoRows) {
		_prefs := &boiler.PlayerSettingsPreference{
			PlayerID: user.ID,
		}

		err := _prefs.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, errMsg)
		}
		reply(_prefs)
		return nil
	}

	reply(prefs)
	return nil

}

const HubKeyPlayerPreferencesUpdate = "PLAYER:PREFERENCES_UPDATE"

type PlayerPreferencesUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		EnableTelegramNotifications bool   `json:"enable_telegram_notifications"`
		EnableSMSNotifications      bool   `json:"enable_sms_notifications"`
		EnablePushNotifications     bool   `json:"enable_push_notifications"`
		MobileNumber                string `json:"mobile_number"`
	} `json:"payload"`
}

func (pc *PlayerController) PlayerPreferencesUpdateHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue updating settings, try again or contact support."
	req := &PlayerPreferencesUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// getting player's preferences
	prefs, err := boiler.PlayerSettingsPreferences(boiler.PlayerSettingsPreferenceWhere.PlayerID.EQ(user.ID)).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, errMsg)
	}

	// if player doesnt have preferences saved, create a new one
	if errors.Is(err, sql.ErrNoRows) {
		_prefs := &boiler.PlayerSettingsPreference{
			PlayerID:                    user.ID,
			EnableTelegramNotifications: req.Payload.EnableTelegramNotifications,
			EnableSMSNotifications:      req.Payload.EnableSMSNotifications,
			EnablePushNotifications:     req.Payload.EnablePushNotifications,
		}

		// check mobile number
		if req.Payload.MobileNumber != "" && req.Payload.EnableSMSNotifications {
			mobileNumber, err := pc.API.SMS.Lookup(req.Payload.MobileNumber)
			if err != nil {
				gamelog.L.Warn().Err(err).Str("mobile number", req.Payload.MobileNumber).Msg("Failed to lookup mobile number through twilio api")
				return terror.Error(err, "Invalid phone number")
			}

			// set the verified mobile number
			_prefs.MobileNumber = null.StringFrom(mobileNumber)
		}

		err = _prefs.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, errMsg)
		}

		// if new preferences and has telegram notifications enabled, must register to telebot
		if _prefs.EnableTelegramNotifications {
			_, err = pc.API.Telegram.PreferencesUpdate(user.ID)
			if err != nil {
				return terror.Error(err, errMsg)
			}
		}
		reply(_prefs)

		return nil
	}

	// update preferences
	prefs.EnableTelegramNotifications = req.Payload.EnableTelegramNotifications
	prefs.EnableSMSNotifications = req.Payload.EnableSMSNotifications
	prefs.EnablePushNotifications = req.Payload.EnablePushNotifications
	if !prefs.EnableTelegramNotifications {
		prefs.Shortcode = ""
	}

	if req.Payload.EnableSMSNotifications && req.Payload.MobileNumber != "" {
		// check mobile number
		mobileNumber, err := pc.API.SMS.Lookup(req.Payload.MobileNumber)
		if err != nil {
			gamelog.L.Warn().Err(err).Str("mobile number", req.Payload.MobileNumber).Msg("Failed to lookup mobile number through twilio api")
			return terror.Error(err, "Invalid phone number")
		}

		// set the verified mobile number
		prefs.MobileNumber = null.StringFrom(mobileNumber)
	}

	if req.Payload.MobileNumber == "" {
		prefs.MobileNumber = null.String{}
	}

	_, err = prefs.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// if telegram enabled but is not registered
	if prefs.EnableTelegramNotifications && (!prefs.TelegramID.Valid && prefs.Shortcode == "") {
		prefs, err = pc.API.Telegram.PreferencesUpdate(user.ID)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}

	reply(prefs)
	return nil
}
