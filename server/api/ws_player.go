package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server"
	"server/db"
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
	"github.com/volatiletech/sqlboiler/v4/types"
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

	api.SecureUserCommand(HubKeyPlayerUpdateSettings, pc.PlayerUpdateSettingsHandler)
	api.SecureUserCommand(HubKeyPlayerGetSettings, pc.PlayerGetSettingsHandler)
	api.SecureUserSubscribeCommand(HubKeyTelegramShortcodeRegistered, pc.PlayerGetTelegramShortcodeRegistered)

	// punish vote related
	api.SecureUserCommand(HubKeyPlayerPunishmentList, pc.PlayerPunishmentList)
	api.SecureUserCommand(HubKeyPlayerActiveCheck, pc.PlayerActiveCheckHandler)
	api.SecureUserFactionCommand(HubKeyFactionPlayerSearch, pc.FactionPlayerSearch)
	api.SecureUserFactionCommand(HubKeyPunishOptions, pc.PunishOptions)
	api.SecureUserFactionCommand(HubKeyPunishVote, pc.PunishVote)
	api.SecureUserFactionCommand(HubKeyIssuePunishVote, pc.IssuePunishVote)
	api.SecureUserFactionCommand(HubKeyPunishVotePriceQuote, pc.PunishVotePriceQuote)
	api.SecureUserFactionSubscribeCommand(HubKeyPunishVoteSubscribe, pc.PunishVoteSubscribeHandler)
	api.SecureUserFactionSubscribeCommand(HubKeyPunishVoteResultSubscribe, pc.PunishVoteResultSubscribeHandler)

	api.SecureUserFactionSubscribeCommand(HubKeyFactionActivePlayersSubscribe, pc.FactionActivePlayersSubscribeHandler)

	return pc
}

type PlayerUpdateSettingsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Key   string     `json:"key"`
		Value types.JSON `json:"value,omitempty"`
	} `json:"payload"`
}

const HubKeyPlayerUpdateSettings hub.HubCommandKey = "PLAYER:UPDATE_SETTINGS"

func (pc *PlayerController) PlayerUpdateSettingsHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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

const HubKeyPlayerGetSettings hub.HubCommandKey = "PLAYER:GET_SETTINGS"

//PlayerGetSettingsHandler gets settings based on key, sends settings value back as json
func (pc *PlayerController) PlayerGetSettingsHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
			// if there are no results, create entry in table and rreturn a null json- tells frontend to use default settings
			playerPrefs := &boiler.PlayerPreference{
				PlayerID:  wsc.Identifier(),
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

const HubKeyTelegramShortcodeRegistered hub.HubCommandKey = "USER:TELEGRAM_SHORTCODE_REGISTERED"

func (pc *PlayerController) PlayerGetTelegramShortcodeRegistered(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	if needProcess {
		reply(false)
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTelegramShortcodeRegistered, wsc.Identifier())), nil
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

type PlayerPunishment struct {
	*boiler.PunishedPlayer
	RelatedPunishVote *boiler.PunishVote   `json:"related_punish_vote"`
	PunishOption      *boiler.PunishOption `json:"punish_option"`
}

const HubKeyPlayerPunishmentList hub.HubCommandKey = "PLAYER:PUNISHMENT:LIST"

func (pc *PlayerController) PlayerPunishmentList(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	punishments, err := boiler.PunishedPlayers(
		boiler.PunishedPlayerWhere.PlayerID.EQ(wsc.Identifier()),
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		qm.Load(boiler.PunishedPlayerRels.PunishOption),
		qm.Load(boiler.PunishedPlayerRels.RelatedPunishVote),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("player id", wsc.Identifier()).Err(err).Msg("Failed to get player's punishment from db")
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

const HubKeyPlayerActiveCheck hub.HubCommandKey = "GOJI:BERRY:TEA"

func (pc *PlayerController) PlayerActiveCheckHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
	player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to get player from db")
	}

	if player.FactionID.Valid {
		fap, ok := pc.API.FactionActivePlayers[player.FactionID.String]
		if !ok {
			return nil
		}

		err = fap.Set(player.ID, isActive)
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

const HubKeyFactionPlayerSearch hub.HubCommandKey = "FACTION:PLAYER:SEARCH"

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
			boiler.PlayerColumns.Gid,
		),
		boiler.PlayerWhere.FactionID.EQ(player.FactionID),
		boiler.PlayerWhere.IsAi.EQ(false),
		boiler.PlayerWhere.ID.NEQ(wsc.Identifier()),
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

const HubKeyPunishVote hub.HubCommandKey = "PUNISH:VOTE"

func (pc *PlayerController) PunishVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &PunishVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// check player has at least 5 kills in the last 7 days
	killCount, err := db.GetPlayerAbilityKills(wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to get player last 7 days kill count from db")
	}

	if killCount < 5 {
		return terror.Error(terror.ErrForbidden, "Require at least 5 kills in last 7 days to vote")
	}

	// check player is available to be punished
	player, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to get current player from db")
	}

	// get player last 7 days kills count
	playerKill, err := db.GetPlayerAbilityKills(player.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("player_id", player.ID).Err(err).Msg("Failed to get player ability kills from db")
		return terror.Error(err, "Failed to get player ability kills from db")
	}

	if playerKill <= 0 {
		return terror.Error(fmt.Errorf("only players with positive ability kill count has the right"), "Does not meet the minimum ability kill count to do the punishment vote")
	}

	fpv, ok := pc.API.FactionPunishVote[player.FactionID.String]
	if !ok {
		return terror.Error(fmt.Errorf("player faction id does not exist"))
	}

	err = fpv.Vote(req.Payload.PunishVoteID, wsc.Identifier(), req.Payload.IsAgreed)
	if err != nil {
		return terror.Error(err, err.Error())
	}

	reply(true)

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

type PunishVotePriceQuoteRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		IntendToPunishPlayerID uuid.UUID `json:"intend_to_punish_player_id"`
	} `json:"payload"`
}

const HubKeyPunishVotePriceQuote hub.HubCommandKey = "PUNISH:VOTE:PRICE:QUOTE"

func (pc *PlayerController) PunishVotePriceQuote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &PunishVotePriceQuoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// check player is available to be punished
	currentPlayer, err := boiler.FindPlayer(gamedb.StdConn, wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to get current player from db")
	}

	intendToBenPlayer, err := boiler.FindPlayer(gamedb.StdConn, req.Payload.IntendToPunishPlayerID.String())
	if err != nil {
		return terror.Error(err, "Failed to get intend to punish player from db")
	}

	if !intendToBenPlayer.FactionID.Valid || intendToBenPlayer.FactionID.String != currentPlayer.FactionID.String {
		return terror.Error(fmt.Errorf("unable to punish player who is not in your faction"), "Unable to quote the price of punish vote with a player in other faction")
	}

	// get the highest price
	price := currentPlayer.IssuePunishFee
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

const HubKeyIssuePunishVote hub.HubCommandKey = "ISSUE:PUNISH:VOTE"

func (pc *PlayerController) IssuePunishVote(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &IssuePunishVoteRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// check player has at least 5 kills in the last 7 days
	killCount, err := db.GetPlayerAbilityKills(wsc.Identifier())
	if err != nil {
		return terror.Error(err, "Failed to get player last 7 days kill count from db")
	}

	if killCount < 5 {
		return terror.Error(terror.ErrForbidden, "Require at least 5 kills in last 7 days to issue vote")
	}

	// check player is available to be punished
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

	if req.Payload.Reason == "" {
		return terror.Error(terror.ErrInvalidInput, "Reason is required")
	}

	// get punish type
	punishOption, err := boiler.FindPunishOption(gamedb.StdConn, req.Payload.PunishOptionID)
	if err != nil {
		return terror.Error(err, "Failed to get punish type from db")
	}

	if _, ok := pc.API.FactionPunishVote[currentPlayer.FactionID.String]; !ok {
		gamelog.L.Error().Str("faction id", currentPlayer.FactionID.String).Err(fmt.Errorf("faction punish vote not found")).Msg("Faction punish vote not found")
		return terror.Error(fmt.Errorf("faction punish vote not found"), "Faction punish vote not found")
	}

	// ensure punish vote is issued synchroniously in faction
	pc.API.FactionPunishVote[currentPlayer.FactionID.String].Lock()
	defer pc.API.FactionPunishVote[currentPlayer.FactionID.String].Unlock()

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
		FactionID:              currentPlayer.FactionID.String,
		IssuedByID:             currentPlayer.ID,
		IssuedByGid:            currentPlayer.Gid,
		IssuedByUsername:       currentPlayer.Username.String,
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
		Description:          "issue vote for punishing player",
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

const HubKeyPunishVoteSubscribe hub.HubCommandKey = "PUNISH:VOTE:SUBSCRIBE"

func (pc *PlayerController) PunishVoteSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
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

	if needProcess {
		// only pass down vote, if there is an ongoing vote
		if fpv, ok := pc.API.FactionPunishVote[player.FactionID.String]; ok {
			fpv.RLock()
			defer fpv.RUnlock()
			if fpv.CurrentPunishVote != nil && fpv.Stage.Phase == PunishVotePhaseVoting {
				bv, err := boiler.PunishVotes(
					boiler.PunishVoteWhere.ID.EQ(fpv.CurrentPunishVote.ID),
					qm.Load(boiler.PunishVoteRels.PunishOption),
				).One(gamedb.StdConn)
				if err != nil {
					return "", "", terror.Error(err, "Failed to get punish vote from db")
				}

				pvr := &PunishVoteResponse{
					PunishVote:   bv,
					PunishOption: bv.R.PunishOption,
				}

				// check user has voted
				decision, err := boiler.PlayersPunishVotes(
					boiler.PlayersPunishVoteWhere.PunishVoteID.EQ(bv.ID),
					boiler.PlayersPunishVoteWhere.PlayerID.EQ(client.Identifier()),
				).One(gamedb.StdConn)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return "", "", terror.Error(err, "Failed to check player had voted")
				}

				if decision != nil {
					pvr.Decision = &PunishVoteDecision{
						IsAgreed: decision.IsAgreed,
					}
				}

				reply(pvr)
			}
		}
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPunishVoteSubscribe, player.FactionID.String)), nil
}

type PunishVoteResult struct {
	PunishVoteID          string `json:"punish_vote_id"`
	TotalPlayerNumber     int    `json:"total_player_number"`
	AgreedPlayerNumber    int    `json:"agreed_player_number"`
	DisagreedPlayerNumber int    `json:"disagreed_player_number"`
}

const HubKeyPunishVoteResultSubscribe hub.HubCommandKey = "PUNISH:VOTE:RESULT:SUBSCRIBE"

func (pc *PlayerController) PunishVoteResultSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
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

	if needProcess {
		// only pass down vote result, if there is an ongoing punish vote
		if fpv, ok := pc.API.FactionPunishVote[player.FactionID.String]; ok && fpv.Stage.Phase == PunishVotePhaseVoting {
			fpv.RLock()
			defer fpv.RUnlock()
			if fpv.CurrentPunishVote != nil {
				reply(&PunishVoteResult{
					PunishVoteID:          fpv.CurrentPunishVote.ID,
					TotalPlayerNumber:     len(fpv.CurrentPunishVote.PlayerPool),
					AgreedPlayerNumber:    len(fpv.CurrentPunishVote.AgreedPlayerIDs),
					DisagreedPlayerNumber: len(fpv.CurrentPunishVote.DisagreedPlayerIDs),
				})
			}
		}
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyPunishVoteResultSubscribe, player.FactionID.String)), nil
}

const HubKeyFactionActivePlayersSubscribe hub.HubCommandKey = "FACTION:ACTIVE:PLAYER:SUBSCRIBE"

func (pc *PlayerController) FactionActivePlayersSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	player, err := boiler.FindPlayer(gamedb.StdConn, client.Identifier())
	if err != nil {
		return "", "", terror.Error(err, "Failed to get player from db")
	}

	fap, ok := pc.API.FactionActivePlayers[player.FactionID.String]
	if !ok {
		return "", "", terror.Error(terror.ErrForbidden, "Faction does not exist in the list")
	}

	if needProcess {
		reply(fap.CurrentFactionActivePlayer())
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionActivePlayersSubscribe, player.FactionID.String)), nil
}
