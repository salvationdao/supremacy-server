package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/multipliers"
	"sort"
	"sync"
	"time"

	"github.com/gofrs/uuid"

	"github.com/ninja-syndicate/ws"

	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/friendsofgo/errors"

	"github.com/ninja-software/terror/v2"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kevinms/leakybucket-go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

const PersistChatMessageLimit = 50

var bm = bluemonday.StrictPolicy()

// ChatMessage contains chat message data to send.
type ChatMessage struct {
	Type   ChatMessageType `json:"type"`
	SentAt time.Time       `json:"sent_at"`
	Data   interface{}     `json:"data"`
}

type ChatMessageType string

const (
	ChatMessageTypeText       ChatMessageType = "TEXT"
	ChatMessageTypePunishVote ChatMessageType = "PUNISH_VOTE"
)

type MessageText struct {
	Message         string           `json:"message"`
	MessageColor    string           `json:"message_color"`
	FromUser        boiler.Player    `json:"from_user"`
	UserRank        string           `json:"user_rank"`
	FromUserStat    *server.UserStat `json:"from_user_stat"`
	Lang            string           `json:"lang"`
	TotalMultiplier string           `json:"total_multiplier"`
	IsCitizen       bool             `json:"is_citizen"`
}

type MessagePunishVote struct {
	IssuedByUser boiler.Player `json:"issued_by_user"`
	ReportedUser boiler.Player `json:"reported_user"`

	// vote result
	IsPassed              bool                `json:"is_passed"`
	TotalPlayerNumber     int                 `json:"total_player_number"`
	AgreedPlayerNumber    int                 `json:"agreed_player_number"`
	DisagreedPlayerNumber int                 `json:"disagreed_player_number"`
	PunishOption          boiler.PunishOption `json:"punish_option"`
	PunishReason          string              `json:"punish_reason"`
	InstantPassByUsers    []*boiler.Player    `json:"instant_pass_by_users"`
}

// Chatroom holds a specific chat room
type Chatroom struct {
	sync.RWMutex
	factionID *server.FactionID
	messages  []*ChatMessage
}

func (c *Chatroom) AddMessage(message *ChatMessage) {
	c.Lock()
	c.messages = append(c.messages, message)
	if len(c.messages) >= PersistChatMessageLimit {
		c.messages = c.messages[1:]
	}
	c.Unlock()
}

func (c *Chatroom) Range(fn func(chatMessage *ChatMessage) bool) {
	c.RLock()
	for _, message := range c.messages {
		if !fn(message) {
			break
		}
	}
	c.RUnlock()
}

func isFingerPrintBanned(playerID string) bool {
	// get fingerprints from player
	fps, err := boiler.PlayerFingerprints(boiler.PlayerFingerprintWhere.PlayerID.EQ(playerID)).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Warn().Err(err).Interface("msg.PlayerID", playerID).Msg("issue finding player fingerprints")
		return false
	}
	if errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Warn().Err(err).Interface("msg.PlayerID", playerID).Msg("player has no fingerprints")
		return false
	}

	ids := []string{}
	for _, f := range fps {
		ids = append(ids, f.FingerprintID)
	}
	// check if any of the players fingerprints are banned
	bannedFingerprints, err := boiler.ChatBannedFingerprints(boiler.ChatBannedFingerprintWhere.FingerprintID.IN(ids)).All(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Warn().Err(err).Interface("msg.PlayerID", playerID).Msg("issue checking if player is banned")
		return false
	}
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false
	}
	return len(bannedFingerprints) > 0
}

func NewChatroom(factionID string) *Chatroom {
	stream := "global"
	if factionID != "" {
		stream = factionID
	}
	msgs, _ := boiler.ChatHistories(
		boiler.ChatHistoryWhere.ChatStream.EQ(stream),
		qm.OrderBy(fmt.Sprintf("%s %s", boiler.ChatHistoryColumns.CreatedAt, "DESC")),
		qm.Limit(PersistChatMessageLimit),
	).All(gamedb.StdConn)

	players := map[string]*boiler.Player{}
	stats := map[string]*server.UserStat{}

	cms := make([]*ChatMessage, len(msgs))
	for i, msg := range msgs {

		player, ok := players[msg.PlayerID]
		if !ok {
			var err error
			player, err = boiler.Players(
				qm.Select(
					boiler.PlayerColumns.ID,
					boiler.PlayerColumns.Username,
					boiler.PlayerColumns.Gid,
					boiler.PlayerColumns.FactionID,
					boiler.PlayerColumns.Rank,
					boiler.PlayerColumns.SentMessageCount,
				),
				boiler.PlayerWhere.ID.EQ(msg.PlayerID),
			).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Warn().Err(err).Interface("msg.PlayerID", msg.PlayerID).Msg("issue finding player")
				continue
			}
			playerStat, err := db.UserStatsGet(player.ID)
			if err != nil {
				gamelog.L.Warn().Err(err).Interface("player.ID", player.ID).Msg("issue UserStatsGet")
				continue
			}
			stats[player.ID] = playerStat
		}
		stat := stats[player.ID]

		cms[i] = &ChatMessage{
			Type:   ChatMessageType(msg.MSGType),
			SentAt: msg.CreatedAt,
			Data: &MessageText{
				Message:         msg.Text,
				MessageColor:    msg.MessageColor,
				FromUser:        *player,
				UserRank:        player.Rank,
				FromUserStat:    stat,
				TotalMultiplier: msg.TotalMultiplier,
				IsCitizen:       msg.IsCitizen,
			},
		}
	}

	// sort the messages to the correct order
	sort.Slice(cms, func(i, j int) bool {
		return cms[i].SentAt.Before(cms[j].SentAt)
	})

	factionUUID := server.FactionID(uuid.FromStringOrNil(factionID))
	chatroom := &Chatroom{
		factionID: &factionUUID,
		messages:  cms,
	}
	return chatroom
}

// ChatController holds handlers for chat
type ChatController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewChatController creates the role hub
func NewChatController(api *API) *ChatController {
	chatHub := &ChatController{
		API: api,
	}

	api.SecureUserCommand(HubKeyChatMessage, chatHub.ChatMessageHandler)

	return chatHub
}

// FactionChatRequest sends chat message to specific faction.
type FactionChatRequest struct {
	Payload struct {
		FactionID    server.FactionID `json:"faction_id"`
		MessageColor string           `json:"message_color"`
		Message      string           `json:"message"`
	} `json:"payload"`
}

const HubKeyChatMessage = "CHAT:MESSAGE"

func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}

var bucket = leakybucket.NewCollector(2, 10, true)
var minuteBucket = leakybucket.NewCollector(0.5, 30, true)

// ChatMessageHandler sends chat message from player
func (fc *ChatController) ChatMessageHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	b1 := bucket.Add(user.ID, 1)
	b2 := minuteBucket.Add(user.ID, 1)

	if b1 == 0 || b2 == 0 {
		return terror.Error(fmt.Errorf("too many messages"), "Too many messages.")
	}

	req := &FactionChatRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// omit unused player detail
	player := boiler.Player{
		ID:               user.ID,
		Username:         user.Username,
		Gid:              user.Gid,
		FactionID:        user.FactionID,
		Rank:             user.Rank,
		SentMessageCount: user.SentMessageCount,
	}

	// check user is banned on chat
	isBanned, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(user.ID),
		boiler.PlayerBanWhere.BanSendChat.EQ(true),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to check player on the banned list")
		return err
	}

	// if chat banned just return
	if isBanned {
		return terror.Error(fmt.Errorf("player is banned to chat"), "You are banned to chat")
	}

	// user's fingerprint banned (shadow ban)
	fingerprintBanned := isFingerPrintBanned(user.ID)
	if fingerprintBanned {
		reply(true)
		return nil
	}

	// update player sent message count
	player.SentMessageCount += 1
	_, err = player.Update(gamedb.StdConn, boil.Whitelist(boiler.PlayerColumns.SentMessageCount))
	if err != nil {
		return terror.Error(err, "Failed to update player sent message count")
	}

	msg := html.UnescapeString(bm.Sanitize(req.Payload.Message))

	linguaLanguage, exists := fc.API.LanguageDetector.DetectLanguageOf(msg)
	language := linguaLanguage.String()
	if language == "Unknown" {
		language = db.GetUserLanguage(player.ID)
	}

	func() {
		if exists && language != "English" {
			dbLanguageExists, err := boiler.Languages(boiler.LanguageWhere.Name.EQ(language)).Exists(gamedb.StdConn)
			if err != nil {
				gamelog.L.Warn().Err(err).Msg("can't find language")
				return
			}
			if !dbLanguageExists {
				//insert into language db
				languageStruct := &boiler.Language{
					Name: language,
				}
				languageStruct.Insert(gamedb.StdConn, boil.Infer())
			}

			dbLanguage, err := boiler.Languages(boiler.LanguageWhere.Name.EQ(language)).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Warn().Err(err).Msg("can't find language")
				return
			}

			playerLanguageStruct := &boiler.PlayerLanguage{
				PlayerID:       player.ID,
				LanguageID:     dbLanguage.ID,
				TextIdentified: msg,
				FactionID:      player.FactionID.String,
			}
			playerLanguageStruct.Insert(gamedb.StdConn, boil.Infer())
		}
	}()

	msg = fc.API.ProfanityManager.Detector.Censor(msg)
	if len(msg) > 280 {
		msg = firstN(msg, 280)
	}
	// get player current stat
	playerStat, err := db.UserStatsGet(player.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Unable to get player stat from db")
	}

	battleNum := 0
	lastBattle, err := boiler.Battles(
		qm.Select(boiler.BattleColumns.BattleNumber),
		qm.OrderBy(fmt.Sprintf("%s %s", boiler.BattleColumns.BattleNumber, "DESC")),
		boiler.BattleWhere.EndedAt.IsNotNull()).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Unable to get last battle for chat")
	}

	if lastBattle != nil {
		battleNum = lastBattle.BattleNumber
	}

	_, totalMultiplier, isCitizen := multipliers.GetPlayerMultipliersForBattle(player.ID, battleNum)
	// check if the faction id is provided
	if !req.Payload.FactionID.IsNil() {
		if !player.FactionID.Valid || player.FactionID.String == "" {
			return terror.Error(terror.ErrInvalidInput, "Required to join a faction to send message, please enlist in a faction.")
		}

		if player.FactionID.String != req.Payload.FactionID.String() {
			return terror.Error(terror.ErrForbidden, "Users are not allow to join the faction chat which they are not belong to.")
		}

		chatMessage := &ChatMessage{
			Type:   ChatMessageTypeText,
			SentAt: time.Now(),
			Data: MessageText{
				Message:         msg,
				MessageColor:    req.Payload.MessageColor,
				FromUser:        player,
				UserRank:        player.Rank,
				FromUserStat:    playerStat,
				TotalMultiplier: multipliers.FriendlyFormatMultiplier(totalMultiplier),
				IsCitizen:       isCitizen,
				Lang:            language,
			},
		}

		cm := boiler.ChatHistory{
			FactionID:       player.FactionID.String,
			PlayerID:        player.ID,
			MessageColor:    req.Payload.MessageColor,
			BattleID:        null.String{},
			MSGType:         boiler.ChatMSGTypeEnumTEXT,
			UserRank:        player.Rank,
			TotalMultiplier: multipliers.FriendlyFormatMultiplier(totalMultiplier),
			KillCount:       fmt.Sprintf("%d", playerStat.AbilityKillCount),
			Text:            msg,
			ChatStream:      player.FactionID.String,
			IsCitizen:       isCitizen,
			Lang:            language,
		}

		err = cm.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable to insert msg into chat history")
		}

		// Ability kills
		fc.API.AddFactionChatMessage(player.FactionID.String, chatMessage)

		// send message
		ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", player.FactionID.String), HubKeyFactionChatSubscribe, []*ChatMessage{chatMessage})
		reply(true)
		return nil
	}

	// global message
	chatMessage := &ChatMessage{
		Type:   ChatMessageTypeText,
		SentAt: time.Now(),
		Data: MessageText{
			Message:         msg,
			MessageColor:    req.Payload.MessageColor,
			FromUser:        player,
			UserRank:        player.Rank,
			FromUserStat:    playerStat,
			TotalMultiplier: multipliers.FriendlyFormatMultiplier(totalMultiplier),
			IsCitizen:       isCitizen,
			Lang:            language,
		},
	}

	cm := boiler.ChatHistory{
		FactionID:       player.FactionID.String,
		PlayerID:        player.ID,
		MessageColor:    req.Payload.MessageColor,
		BattleID:        null.String{},
		MSGType:         boiler.ChatMSGTypeEnumTEXT,
		UserRank:        player.Rank,
		TotalMultiplier: multipliers.FriendlyFormatMultiplier(totalMultiplier),
		KillCount:       fmt.Sprintf("%d", playerStat.AbilityKillCount),
		Text:            msg,
		ChatStream:      "global",
		IsCitizen:       isCitizen,
		Lang:            language,
	}

	err = cm.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to insert msg into chat history")
	}

	fc.API.GlobalChat.AddMessage(chatMessage)
	ws.PublishMessage("/public/global_chat", HubKeyGlobalChatSubscribe, []*ChatMessage{chatMessage})
	reply(true)

	return nil
}

const HubKeyFactionChatSubscribe = "FACTION:CHAT:SUBSCRIBE"

func (fc *ChatController) FactionChatUpdatedSubscribeHandler(ctx context.Context, user *boiler.Player, factionID string, key string, payload []byte, reply ws.ReplyFunc) error {
	resp := []*ChatMessage{}
	chatRangeHandler := func(message *ChatMessage) bool {
		resp = append(resp, message)
		return true
	}
	switch factionID {
	case server.RedMountainFactionID:
		fc.API.RedMountainChat.Range(chatRangeHandler)
	case server.BostonCyberneticsFactionID:
		fc.API.BostonChat.Range(chatRangeHandler)
	case server.ZaibatsuFactionID:
		fc.API.ZaibatsuChat.Range(chatRangeHandler)
	default:
		return terror.Error(terror.ErrInvalidInput, "Invalid faction id")
	}

	reply(resp)

	return nil
}

const HubKeyGlobalChatSubscribe = "GLOBAL:CHAT:SUBSCRIBE"

func (fc *ChatController) GlobalChatUpdatedSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	resp := []*ChatMessage{}
	fc.API.GlobalChat.Range(func(message *ChatMessage) bool {
		resp = append(resp, message)
		return true
	})
	reply(resp)
	return nil
}

func (api *API) AddFactionChatMessage(factionID string, msg *ChatMessage) {
	switch factionID {
	case server.RedMountainFactionID:
		api.RedMountainChat.AddMessage(msg)
	case server.BostonCyberneticsFactionID:
		api.BostonChat.AddMessage(msg)
	case server.ZaibatsuFactionID:
		api.ZaibatsuChat.AddMessage(msg)
	}
}
