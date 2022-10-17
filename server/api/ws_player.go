package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/xsyn_rpcclient"
	"strings"
	"time"
	"unicode"

	goaway "github.com/TwiN/go-away"
	"github.com/go-chi/chi/v5"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
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

	api.SecureUserCommand(HubKeyPlayerPreferencesGet, pc.PlayerPreferencesGetHandler)
	api.SecureUserCommand(HubKeyPlayerPreferencesUpdate, pc.PlayerPreferencesUpdateHandler)

	// punish vote related
	api.SecureUserCommand(HubKeyPlayerActiveCheck, pc.PlayerActiveCheckHandler)
	api.SecureUserCommand(HubKeyGetPlayerByGid, pc.GetPlayerByGidHandler)
	api.SecureUserFactionCommand(HubKeyFactionPlayerSearch, pc.FactionPlayerSearch)
	api.SecureUserFactionCommand(HubKeyInstantPassPunishVote, pc.PunishVoteInstantPassHandler)
	api.SecureUserFactionCommand(HubKeyPunishOptions, pc.PunishOptions)
	api.SecureUserFactionCommand(HubKeyPunishVote, pc.PunishVote)
	api.SecureUserFactionCommand(HubKeyIssuePunishVote, pc.IssuePunishVote)
	api.SecureUserFactionCommand(HubKeyPunishVotePriceQuote, pc.PunishVotePriceQuote)

	api.SecureUserCommand(HubKeyFactionEnlist, pc.PlayerFactionEnlistHandler)

	api.SecureUserCommand(HubKeyGameUserOnline, pc.UserOnline)

	api.SecureUserCommand(HubKeyPlayerQueueStatus, pc.PlayerQueueStatusHandler)

	// user profile commands
	api.Command(HubKeyPlayerProfileGet, pc.PlayerProfileGetHandler)
	api.SecureUserCommand(HubKeyPlayerUpdateUsername, pc.PlayerUpdateUsernameHandler)
	api.SecureUserCommand(HubKeyPlayerUpdateAboutMe, pc.PlayerUpdateAboutMeHandler)
	api.SecureUserCommand(HubKeyPlayerAvatarList, pc.ProfileAvatarListHandler)
	api.SecureUserCommand(HubKeyPlayerAvatarUpdate, pc.ProfileAvatarUpdateHandler)

	// custom avatar
	api.SecureUserCommand(HubKeyPlayerProfileLayersList, pc.PlayerProfileAvatarLayersListHandler)
	api.SecureUserCommand(HubKeyPlayerCustomAvatarList, pc.ProfileCustomAvatarListHandler)
	api.SecureUserCommand(HubKeyPlayerProfileCustomAvatarCreate, pc.PlayerProfileCustomAvatarCreate)
	api.SecureUserCommand(HubKeyPlayerProfileCustomAvatarUpdate, pc.PlayerProfileCustomAvatarUpdate)
	api.SecureUserCommand(HubKeyPlayerProfileCustomAvatarDelete, pc.PlayerProfileCustomAvatarDelete)

	api.SecureUserCommand(HubKeyGenOneTimeToken, pc.GenOneTimeToken)

	api.SecureUserFactionCommand(HubKeyPlayerSearch, pc.PlayerSearch)

	return pc
}

type PlayerQueueStatus struct {
	TotalQueued int64 `json:"total_queued"`
	QueueLimit  int64 `json:"queue_limit"`
}

const HubKeyPlayerQueueStatus = "PLAYER:QUEUE:STATUS"

func (pc *PlayerController) PlayerQueueStatusHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	queued, err := db.GetPlayerQueueCount(user.ID)
	if err != nil {
		return terror.Error(err, "Failed to get player queue status.")
	}

	reply(&PlayerQueueStatus{
		TotalQueued: queued,
		QueueLimit:  int64(db.GetIntWithDefault(db.KeyPlayerQueueLimit, 10)),
	})
	return nil
}

type UserUpdatedRequest struct {
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

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

	// give user default profile avatar images
	err = db.GiveDefaultAvatars(user.ID, user.FactionID.String)
	if err != nil {
		return err
	}

	// update user faction in passport
	err = pc.API.Passport.UserFactionEnlist(user.ID, user.FactionID.String)
	if err != nil {
		return terror.Error(err, "Failed to sync passport db")
	}

	if !server.IsProductionEnv() {
		// assign mechs base on player's faction
		labelList := []string{}
		switch user.FactionID.String {
		case server.RedMountainFactionID:
			labelList = []string{
				"Red Mountain Olympus Mons LY07 Villain Chassis",
				"Red Mountain Olympus Mons LY07 Evo Chassis",
				"Red Mountain Olympus Mons LY07 Red Blue Chassis",
			}
		case server.BostonCyberneticsFactionID:
			labelList = []string{
				"Boston Cybernetics Law Enforcer X-1000 White Blue Chassis",
				"Boston Cybernetics Law Enforcer X-1000 BioHazard Chassis",
				"Boston Cybernetics Law Enforcer X-1000 Crystal Blue Chassis",
			}

		case server.ZaibatsuFactionID:
			labelList = []string{
				"Zaibatsu Tenshi Mk1 White Neon Chassis",
				"Zaibatsu Tenshi Mk1 Destroyer Chassis",
				"Zaibatsu Tenshi Mk1 Evangelica Chassis",
			}
		}

		templateIDS := []string{}
		templates, err := boiler.Templates(
			boiler.TemplateWhere.Label.IN(labelList),
		).All(tx)
		if err != nil {
			return terror.Error(err, "Failed to sync passport db")
		}

		for i := 0; i < 3; i++ {
			for _, tmpl := range templates {
				templateIDS = append(templateIDS, tmpl.ID)
			}
		}

		err = pc.API.Passport.AssignTemplateToUser(&xsyn_rpcclient.AssignTemplateReq{
			TemplateIDs: templateIDS,
			UserID:      user.ID,
		})
		if err != nil {
			return terror.Error(err, "Failed to sync passport db")
		}
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, "Failed to commit db transaction")
	}

	err = user.L.LoadRole(gamedb.StdConn, true, user, nil)
	if err != nil {
		return terror.Error(err, "Failed to load role")
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s", user.ID), server.HubKeyUserSubscribe, server.PlayerFromBoiler(user))

	reply(true)

	return nil
}

type PlayerUpdateSettingsRequest struct {
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

func (api *API) PlayerGetTelegramShortcodeRegistered(w http.ResponseWriter, r *http.Request) (int, error) {
	return helpers.EncodeJSON(w, false)
}

type PlayerPunishment struct {
	*boiler.PlayerBan
	RelatedPunishVote *boiler.PunishVote `json:"related_punish_vote"`
	Restrictions      []string           `json:"restrictions"`
	BanByUser         *boiler.Player     `json:"ban_by_user"`
	IsPermanent       bool               `json:"is_permanent"`
}

const HubKeyPlayerPunishmentList = "PLAYER:PUNISHMENT:LIST"

func (pc *PlayerController) PlayerPunishmentList(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	// get current player's punishment
	punishments, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(user.ID),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
		qm.Load(boiler.PlayerBanRels.RelatedPunishVote),
		qm.Load(boiler.PlayerBanRels.BannedBy, qm.Select(boiler.PlayerColumns.ID, boiler.PlayerColumns.Username, boiler.PlayerColumns.Gid)),
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
			PlayerBan:         punishment,
			RelatedPunishVote: punishment.R.RelatedPunishVote,
			Restrictions:      PlayerBanRestrictions(punishment),
			BanByUser:         punishment.R.BannedBy,
			IsPermanent:       punishment.EndAt.After(time.Now().AddDate(0, 1, 0)),
		})
	}

	reply(playerPunishments)

	return nil
}

func PlayerBanRestrictions(pb *boiler.PlayerBan) []string {
	restrictions := []string{}
	if pb.BanLocationSelect {
		restrictions = append(restrictions, RestrictionLocationSelect, RestrictionAbilityTrigger)
	}
	if pb.BanSendChat {
		restrictions = append(restrictions, RestrictionChatSend)
	}
	if pb.BanViewChat {
		restrictions = append(restrictions, RestrictionChatView)
	}
	if pb.BanSupsContribute {
		restrictions = append(restrictions, RestrictionSupsContribute)
	}
	if pb.BanMechQueue {
		restrictions = append(restrictions, RestrictionsMechQueuing)
	}
	return restrictions
}

type PlayerActiveCheckRequest struct {
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

	fap, ok := pc.API.FactionActivePlayers["GLOBAL"]
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

	return nil
}

type PlayerSearchRequest struct {
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

type GetPlayerByGidRequest struct {
	Payload struct {
		Gid int `json:"gid"`
	} `json:"payload"`
}

const HubKeyGetPlayerByGid = "GET:PLAYER:GID"

func (pc *PlayerController) GetPlayerByGidHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "GetPlayerByGidHandler").Str("user_id", user.ID).Logger()

	req := &GetPlayerByGidRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		l.Error().Err(err).Msg("json unmarshal error")
		return terror.Error(err, "Invalid request received")
	}

	l = l.With().Interface("GID", req.Payload.Gid).Logger()
	p, err := boiler.Players(
		boiler.PlayerWhere.Gid.EQ(req.Payload.Gid),
	).One(gamedb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			l.Error().Err(err).Msg("player with gid does not exist.")
			return nil
		}
		l.Error().Err(err).Msg("unable to retrieve player by GID.")
		return terror.Error(err, "Unable to find player, try again or contact support.")
	}

	pp := &server.PublicPlayer{
		ID:        p.ID,
		Username:  p.Username,
		Gid:       p.Gid,
		FactionID: p.FactionID,
		AboutMe:   p.AboutMe,
		Rank:      p.Rank,
		CreatedAt: p.CreatedAt,
	}

	reply(pp)
	return nil
}

type PunishVoteInstantPassRequest struct {
	Payload struct {
		PunishVoteID string `json:"punish_vote_id"`
	} `json:"payload"`
}

const HubKeyPlayerSearch = "PLAYER:SEARCH"

// PlayerSearch return up to 10 players base on the given text
func (pc *PlayerController) PlayerSearch(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PlayerSearchRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if user.RoleID == server.UserRolePlayer.String() || user.RoleID == "" {
		return terror.Error(err, "User is not an admin")
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
		boiler.PlayerWhere.IsAi.EQ(false),
		boiler.PlayerWhere.ID.NEQ(user.ID),
		qm.Where(
			fmt.Sprintf("LOWER(%s||'#'||%s::TEXT) LIKE ?",
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Username),
				qm.Rels(boiler.TableNames.Players, boiler.PlayerColumns.Gid),
			),
			"%"+strings.ToLower(search)+"%",
		),
		qm.Limit(10),
	).All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to search players from db")
	}

	reply(ps)
	return nil
}

const HubKeyInstantPassPunishVote = "PUNISH:VOTE:INSTANT:PASS"

func (pc *PlayerController) PunishVoteInstantPassHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PunishVoteInstantPassRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// check player is available vote
	player, err := boiler.FindPlayer(gamedb.StdConn, user.ID)
	if err != nil {
		return terror.Error(err, "Failed to get current player from db")
	}

	if player.Rank != boiler.PlayerRankEnumGENERAL {
		return terror.Error(terror.ErrInvalidInput, "Only players with rank 'GENERAL' can instantly pass a punish vote.")
	}

	// check punish vote is finalised
	fpv, ok := pc.API.FactionPunishVote[player.FactionID.String]
	if !ok {
		return terror.Error(fmt.Errorf("player faction id does not exist"))
	}

	err = fpv.InstantPass(pc.API.Passport, req.Payload.PunishVoteID, user.ID)
	if err != nil {
		return terror.Error(err, err.Error())
	}

	reply(true)

	// update instant vote count
	requiredAmount := db.GetIntWithDefault(db.KeyInstantPassRequiredAmount, 2)

	count, err := boiler.PunishVoteInstantPassRecords(
		boiler.PunishVoteInstantPassRecordWhere.PunishVoteID.EQ(req.Payload.PunishVoteID),
	).Count(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get punish vote count")
	}

	ws.PublishMessage(fmt.Sprintf("/faction/%s/punish_vote/%s/command_override", factionID, req.Payload.PunishVoteID), HubKeyPunishVoteCommandOverrideCountSubscribe, fmt.Sprintf("%d/%d", count, requiredAmount))

	return nil
}

type PunishVoteRequest struct {
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
	queries := []qm.QueryMod{
		boiler.PlayerBanWhere.BannedPlayerID.EQ(req.Payload.IntendToPunishPlayerID.String()),
	}

	switch punishOption.Key {
	case "restrict_sups_contribution":
		queries = append(queries, boiler.PlayerBanWhere.BanSupsContribute.EQ(true))
	case "restrict_location_select":
		queries = append(queries, boiler.PlayerBanWhere.BanLocationSelect.EQ(true))

		// skip, if the player is in the team kill courtroom
		if pc.API.ArenaManager.SystemBanManager.HasOngoingTeamKillCases(intendToBenPlayer.ID) {
			return terror.Error(fmt.Errorf("player is listed on system ban"), "The player is already listed on system ban list")
		}
	case "restrict_chat":
		queries = append(queries, boiler.PlayerBanWhere.BanSendChat.EQ(true))
	}

	queries = append(queries,
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
	)

	punishedPlayer, err := boiler.PlayerBans(queries...).One(gamedb.StdConn)
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
		InstantPassFee:         price.Mul(decimal.New(1, 18)),
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
	txid, err := pc.API.Passport.SpendSupMessage(xsyn_rpcclient.SpendSupsReq{
		FromUserID:           userID,
		ToUserID:             uuid.Must(uuid.FromString(factionAccountID)),
		Amount:               price.Mul(decimal.New(1, 18)).String(),
		TransactionReference: server.TransactionReference(fmt.Sprintf("issue_punish_vote|%s|%d", punishVote.ID, time.Now().UnixNano())),
		Group:                "issue punish vote",
		SubGroup:             string(punishOption.Key),
		Description:          "issue vote for punishing player",
	})
	if err != nil {
		gamelog.L.Error().Str("player_id", user.ID).Str("amount", price.Mul(decimal.New(1, 18)).String()).Err(err).Msg("Failed to pay sups for issuing player punish vote")
		return err
	}

	err = tx.Commit()
	if err != nil {
		pc.API.Passport.RefundSupsMessage(txid)
		return terror.Error(err, "Failed to commit db transaction")
	}

	reply(true)

	return nil
}

const HubKeyPunishVoteCommandOverrideCountSubscribe = "PUNISH:VOTE:COMMAND:OVERRIDE:COUNT:SUBSCRIBE"

func (pc *PlayerController) PunishVoteCommandOverrideCountSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	cctx := chi.RouteContext(ctx)
	punishVoteID := cctx.URLParam("punish_vote_id")

	requiredAmount := db.GetIntWithDefault(db.KeyInstantPassRequiredAmount, 2)

	count, err := boiler.PunishVoteInstantPassRecords(
		boiler.PunishVoteInstantPassRecordWhere.PunishVoteID.EQ(punishVoteID),
	).Count(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get punish vote count")
	}

	reply(fmt.Sprintf("%d/%d", count, requiredAmount))

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
	PunishOption       *boiler.PunishOption `json:"punish_option"`
	Decision           *PunishVoteDecision  `json:"decision,omitempty"`
	InstantPassUserIDs []string             `json:"instant_pass_user_ids"`
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
				qm.Load(boiler.PunishVoteRels.PunishVoteInstantPassRecords),
			).One(gamedb.StdConn)
			if err != nil {
				return terror.Error(err, "Failed to get punish vote from db")
			}

			pvr := &PunishVoteResponse{
				PunishVote:         bv,
				PunishOption:       bv.R.PunishOption,
				InstantPassUserIDs: []string{},
			}

			for _, ipr := range bv.R.PunishVoteInstantPassRecords {
				pvr.InstantPassUserIDs = append(pvr.InstantPassUserIDs, ipr.VoteByPlayerID)
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

const HubKeyGlobalActivePlayersSubscribe = "GLOBAL:ACTIVE:PLAYER:SUBSCRIBE"

func (pc *PlayerController) GlobalActivePlayersSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	fap, ok := pc.API.FactionActivePlayers["GLOBAL"]
	if !ok {
		return terror.Error(terror.ErrForbidden, "Could not subscribe to active players in global chat, try again or contact support.")
	}

	reply(fap.CurrentFactionActivePlayer())

	return nil
}

func (pc *PlayerController) PlayersSubscribeHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	// pass viewer update
	pc.API.ViewerUpdateChan <- true

	// broadcast player
	features, err := db.GetPlayerFeaturesByID(user.ID)
	if err != nil {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to get player feature")
	}

	err = user.L.LoadRole(gamedb.StdConn, true, user, nil)
	if err != nil {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to get player role")
	}

	reply(server.PlayerFromBoiler(user, features))

	// broadcast player stat
	us, err := db.UserStatsGet(user.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to get player stat")
		}
	}

	if us != nil {
		ws.PublishMessage(fmt.Sprintf("/secure/user/%s/stat", user.ID), server.HubKeyUserStatSubscribe, us)
	}

	// broadcast player punishment list
	// get current player's punishment
	punishments, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(user.ID),
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
		qm.Load(boiler.PlayerBanRels.RelatedPunishVote),
		qm.Load(boiler.PlayerBanRels.BannedBy, qm.Select(boiler.PlayerColumns.ID, boiler.PlayerColumns.Username, boiler.PlayerColumns.Gid)),
	).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to get player's punishment from db")
		return terror.Error(err, "Failed to get player's punishment from db")
	}

	if punishments == nil || len(punishments) == 0 {
		return nil
	}

	playerPunishments := []*PlayerPunishment{}
	for _, punishment := range punishments {
		playerPunishments = append(playerPunishments, &PlayerPunishment{
			PlayerBan:         punishment,
			RelatedPunishVote: punishment.R.RelatedPunishVote,
			Restrictions:      PlayerBanRestrictions(punishment),
			BanByUser:         punishment.R.BannedBy,
			IsPermanent:       punishment.EndAt.After(time.Now().AddDate(0, 1, 0)),
		})
	}

	// send to the player
	ws.PublishMessage(fmt.Sprintf("/secure/user/%s/punishment_list", user.ID), HubKeyPlayerPunishmentList, playerPunishments)

	return nil
}

func (pc *PlayerController) PlayersStatSubscribeHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	us, err := db.UserStatsGet(user.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Str("player id", user.ID).Err(err).Msg("Failed to get player stat")
		return terror.Error(err, "Failed to load player stat")
	}

	if us != nil {
		reply(us)
	}

	return nil
}

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

type UserOnlineRequest struct {
	Payload struct {
		ArenaID string `json:"arena_id"`
	} `json:"payload"`
}

func (pc *PlayerController) UserOnline(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &UserOnlineRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	arena, err := pc.API.ArenaManager.GetArena(req.Payload.ArenaID)
	if err != nil {
		return err
	}

	btl := arena.CurrentBattle()
	if btl == nil {
		return nil
	}

	err = db.BattleViewerUpsert(btl.ID, user.ID)
	if err != nil {
		gamelog.L.Error().
			Str("battle_id", btl.ID).
			Str("player_id", user.ID).
			Err(err).
			Msg("could not upsert battle viewer")
		return terror.Error(err, "Failed to record battle viewer.")
	}

	return nil
}

const HubKeyPlayerPreferencesGet = "PLAYER:PREFERENCES_GET"

// PlayerPreferencesGetHandler gets player's preferences
func (pc *PlayerController) PlayerPreferencesGetHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue getting player preferences, try again or contact support."

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

type PlayerUpdateUsernameRequest struct {
	Payload struct {
		PlayerID    string `json:"player_id"`
		NewUsername string `json:"new_username"`
	} `json:"payload"`
}

const HubKeyPlayerUpdateUsername = "PLAYER:UPDATE:USERNAME"

func (pc *PlayerController) PlayerUpdateUsernameHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue updating username, try again or contact support."
	req := &PlayerUpdateUsernameRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	// check if user
	if req.Payload.PlayerID != user.ID {
		return terror.Error(err, "You do not have permission to update this section")
	}

	// check profanity/ check if valid username
	err = IsValidUsername(req.Payload.NewUsername)
	if err != nil {
		return terror.Error(err, "Invalid username, must be between 3 - 15 characters long, cannot contain profanities.")
	}
	user.Username = null.StringFrom(req.Payload.NewUsername)
	user.UpdatedAt = time.Now()

	_, err = user.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// update in xsyn
	err = pc.API.Passport.UserUpdateUsername(user.ID, req.Payload.NewUsername)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	reply(user.Username.String)
	err = user.L.LoadRole(gamedb.StdConn, true, user, nil)
	if err != nil {
		return terror.Error(err, "Failed to update player's marketing preferences.")
	}

	ws.PublishMessage(fmt.Sprintf("/secure/user/%s", user.ID), server.HubKeyUserSubscribe, server.PlayerFromBoiler(user))

	return nil
}

type PlayerUpdateAboutMeRequest struct {
	Payload struct {
		PlayerID string `json:"player_id"`
		AboutMe  string `json:"about_me"`
	} `json:"payload"`
}

const HubKeyPlayerUpdateAboutMe = "PLAYER:UPDATE:ABOUT_ME"

func (pc *PlayerController) PlayerUpdateAboutMeHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue updating about me, try again or contact support."
	req := &PlayerUpdateAboutMeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	// check if user
	if req.Payload.PlayerID != user.ID {
		return terror.Error(err, "You do not have permission to update this section")
	}
	// check profanity/ check if valid about me
	err = IsValidAboutMe(req.Payload.AboutMe)
	if err != nil {
		return terror.Error(err, "Invalid about me, must be between 3 - 400 characters long, cannot contain profanities.")

	}
	_, err = user.Update(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, errMsg)
	}
	user.AboutMe = null.StringFrom(req.Payload.AboutMe)
	user.UpdatedAt = time.Now()
	_, err = user.Update(gamedb.StdConn, boil.Whitelist(
		boiler.PlayerColumns.AboutMe,
		boiler.PlayerColumns.UpdatedAt,
	))
	if err != nil {
		return terror.Error(err, errMsg)
	}
	resp := &server.PublicPlayer{
		ID:        user.ID,
		Username:  user.Username,
		Gid:       user.Gid,
		FactionID: user.FactionID,
		AboutMe:   user.AboutMe,
		Rank:      user.Rank,
		CreatedAt: user.CreatedAt,
	}
	reply(resp)
	return nil
}

func IsValidUsername(username string) error {
	// Must contain at least 3 characters
	// Cannot contain more than 15 characters
	// Cannot contain profanity
	// Can only contain the following symbols: _
	hasDisallowedSymbol := false
	if UsernameRegExp.Match([]byte(username)) {
		hasDisallowedSymbol = true
	}

	if TrimUsername(username) == "" {
		return terror.Error(fmt.Errorf("username cannot be empty"), "Invalid username. Your username cannot be empty.")
	}
	if PrintableLen(TrimUsername(username)) < 3 {
		return terror.Error(fmt.Errorf("username must be at least characters long"), "Invalid username. Your username must be at least 3 characters long.")
	}
	if PrintableLen(TrimUsername(username)) > 30 {
		return terror.Error(fmt.Errorf("username cannot be more than 30 characters long"), "Invalid username. Your username cannot be more than 30 characters long.")
	}
	if hasDisallowedSymbol {
		return terror.Error(fmt.Errorf("username cannot contain disallowed symbols"), "Invalid username. Your username contains a disallowed symbol.")
	}

	profanityDetector := goaway.NewProfanityDetector()
	profanityDetector = profanityDetector.WithSanitizeLeetSpeak(false)

	if profanityDetector.IsProfane(username) {
		return terror.Error(fmt.Errorf("username contains profanity"), "Invalid username. Your username contains profanity.")
	}

	return nil
}

func IsValidAboutMe(aboutMe string) error {
	// Must contain at least 3 characters
	// Cannot contain more than 400 characters
	// Cannot contain profanity

	if TrimUsername(aboutMe) == "" {
		return terror.Error(fmt.Errorf("about me cannot be empty"), "Invalid about me. Your about me cannot be empty.")
	}
	if PrintableLen(TrimUsername(aboutMe)) < 3 {
		return terror.Error(fmt.Errorf("about me must be at least 3 characters long"), "Invalid about me. Your about me must be at least 3 characters long.")
	}
	if PrintableLen(TrimUsername(aboutMe)) > 400 {
		return terror.Error(fmt.Errorf("about me cannot be more than 400 characters long"), "Invalid about me. Your about me cannot be more than 30 characters long.")
	}

	profanityDetector := goaway.NewProfanityDetector()
	profanityDetector = profanityDetector.WithSanitizeLeetSpeak(false)

	if profanityDetector.IsProfane(aboutMe) {
		return terror.Error(fmt.Errorf("about me contains profanity"), "Invalid about me. Your about me contains profanity.")
	}

	return nil
}

// TrimUsername removes misuse of invisible characters.
func TrimUsername(username string) string {
	// Check if entire string is nothing not non-printable characters
	isEmpty := true
	runes := []rune(username)
	for _, r := range runes {
		if unicode.IsPrint(r) && !unicode.IsSpace(r) {
			isEmpty = false
			break
		}
	}
	if isEmpty {
		return ""
	}

	// Remove Spaces like characters Around String (keep mark ones)
	output := strings.Trim(username, " \u00A0\u180E\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u200B\u202F\u205F\u3000\uFEFF\u2423\u2422\u2420")

	// Enforce one Space like characters between words
	output = strings.Join(strings.Fields(output), " ")

	return output
}

const HubKeyGenOneTimeToken = "GEN:ONE:TIME:TOKEN"

// GenOneTimeToken Generates a token used to create a QR code to log a player into the supremacy companion app
func (pc *PlayerController) GenOneTimeToken(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "GenOneTimeToken").Str("user id", user.ID).Logger()

	resp, err := pc.API.Passport.GenOneTimeToken(user.ID)
	if err != nil {
		l.Error().Err(err).Msg("Failed to generate QR code token")
		return terror.Error(err, "Failed to get login token")
	}

	reply(resp)
	return nil
}

// PlayerQuestStat return current player quest progression
func (pc *PlayerController) PlayerQuestStat(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	pqs, err := db.PlayerQuestStatGet(user.ID)
	if err != nil {
		return err
	}

	reply(pqs)

	return nil
}

// PlayerQuestProgressions return current player quest progression
func (pc *PlayerController) PlayerQuestProgressions(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("player id", user.ID).Str("func name", "PlayerQuestProgressions").Logger()
	result, err := db.PlayerQuestProgressions(user.ID)
	if err != nil {
		l.Error().Err(err).Msg("Failed to load player quest progressions.")
		return err
	}
	reply(result)

	return nil
}
