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

	"github.com/ninja-syndicate/hub"
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
	ID     string          `json:"id"`
	Type   ChatMessageType `json:"type"`
	SentAt time.Time       `json:"sent_at"`
	Data   interface{}     `json:"data"`
}

type ChatMessageType string

const (
	ChatMessageTypeText       ChatMessageType = "TEXT"
	ChatMessageTypePunishVote ChatMessageType = "PUNISH_VOTE"
	ChatMessageTypeSystemBan  ChatMessageType = "SYSTEM_BAN"
	ChatMessageTypeNewBattle  ChatMessageType = "NEW_BATTLE"
)

type MessageText struct {
	ID              string           `json:"id"`
	Message         string           `json:"message"`
	MessageColor    string           `json:"message_color"`
	FromUser        boiler.Player    `json:"from_user"`
	UserRank        string           `json:"user_rank"`
	FromUserStat    *server.UserStat `json:"from_user_stat"`
	Lang            string           `json:"lang"`
	TotalMultiplier string           `json:"total_multiplier"`
	IsCitizen       bool             `json:"is_citizen"`
	BattleNumber    int              `json:"battle_number"`
	Metadata        null.JSON        `json:"metadata"`
}

type MessagePunishVote struct {
	ID           string        `json:"id"`
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

type MessageSystemBan struct {
	ID           string         `json:"id"`
	BannedByUser *boiler.Player `json:"banned_by_user"`
	BannedUser   *boiler.Player `json:"banned_user"`

	FactionID    null.String `json:"faction_id"`
	BattleNumber null.Int    `json:"battle_number"`

	Reason      string `json:"reason"`
	BanDuration string `json:"ban_duration"`

	IsPermanentBan bool     `json:"is_permanent_ban"`
	Restrictions   []string `json:"restrictions"`
}

type MessageNewBattle struct {
	BattleNumber int `json:"battle_number"`
}

type Likes struct {
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`
	Net      int `json:"net"`
}

type TextMessageMetadata struct {
	//gid:true(read/unread)
	TaggedUsersRead map[int]bool `json:"tagged_users_read"`
	Likes           *Likes       `json:"likes"`
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
	cmstoSend := []*ChatMessage{}
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

		if msg.MSGType == boiler.ChatMSGTypeEnumNEW_BATTLE {
			cm := &ChatMessage{}
			err := msg.Metadata.Unmarshal(cm)
			if err != nil {
				continue
			}

			cms[i] = cm
			cmstoSend = append(cmstoSend, cms[i])
			continue
		}

		cms[i] = &ChatMessage{
			ID:     msg.ID,
			Type:   ChatMessageType(msg.MSGType),
			SentAt: msg.CreatedAt,
			Data: &MessageText{
				ID:              msg.ID,
				Message:         msg.Text,
				MessageColor:    msg.MessageColor,
				FromUser:        *player,
				UserRank:        player.Rank,
				FromUserStat:    stat,
				TotalMultiplier: msg.TotalMultiplier,
				IsCitizen:       msg.IsCitizen,
				Metadata:        msg.Metadata,
			},
		}
		cmstoSend = append(cmstoSend, cms[i])
	}

	// sort the messages to the correct order
	sort.Slice(cmstoSend, func(i, j int) bool {
		return cmstoSend[i].SentAt.Before(cmstoSend[j].SentAt)
	})

	factionUUID := server.FactionID(uuid.FromStringOrNil(factionID))
	chatroom := &Chatroom{
		factionID: &factionUUID,
		messages:  cmstoSend,
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
	api.SecureUserCommand(HubKeyReadTaggedMessage, chatHub.ReadTaggedMessageHandler)
	api.SecureUserCommand(HubKeyChatBanPlayer, chatHub.ChatBanPlayerHandler)

	go api.MessageBroadcaster()

	return chatHub
}

const (
	RestrictionLocationSelect = "Select location"
	RestrictionAbilityTrigger = "Trigger abilities"
	RestrictionChatSend       = "Send chat"
	RestrictionChatView       = "Receive chat"
	RestrictionSupsContribute = "Contribute sups"
)

func (api *API) MessageBroadcaster() {
	for {
		select {
		case msg := <-api.BattleArena.SystemBanManager.SystemBanMassageChan:

			banMessage := &MessageSystemBan{
				ID:             uuid.Must(uuid.NewV4()).String(),
				BannedByUser:   msg.SystemPlayer,
				BannedUser:     msg.BannedPlayer,
				FactionID:      msg.FactionID,
				BattleNumber:   msg.PlayerBan.BattleNumber,
				Reason:         msg.PlayerBan.Reason,
				BanDuration:    msg.BanDuration,
				IsPermanentBan: msg.PlayerBan.EndAt.After(time.Now().AddDate(0, 1, 0)),
				Restrictions:   PlayerBanRestrictions(msg.PlayerBan),
			}

			cm := &ChatMessage{
				ID:     banMessage.ID,
				Type:   ChatMessageTypeSystemBan,
				SentAt: time.Now(),
				Data:   banMessage,
			}

			switch msg.FactionID.String {
			case server.RedMountainFactionID:
				api.RedMountainChat.AddMessage(cm)
				ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", msg.FactionID.String), HubKeyFactionChatSubscribe, []*ChatMessage{cm})

			case server.BostonCyberneticsFactionID:
				api.BostonChat.AddMessage(cm)
				ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", msg.FactionID.String), HubKeyFactionChatSubscribe, []*ChatMessage{cm})

			case server.ZaibatsuFactionID:
				api.ZaibatsuChat.AddMessage(cm)
				ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", msg.FactionID.String), HubKeyFactionChatSubscribe, []*ChatMessage{cm})

			default:
				api.GlobalChat.AddMessage(cm)
				ws.PublishMessage("/public/global_chat", HubKeyGlobalChatSubscribe, []*ChatMessage{cm})
			}
		case newBattleInfo := <-api.BattleArena.NewBattleChan:
			err := api.BroadcastNewBattle(newBattleInfo.BattleNumber)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("Could not broadcast battle info ", newBattleInfo).Msg("failed to broadcast new battle info")
				return
			}
		}
	}
}

// FactionChatRequest sends chat message to specific faction.
type FactionChatRequest struct {
	Payload struct {
		FactionID       server.FactionID `json:"faction_id"`
		MessageColor    string           `json:"message_color"`
		Message         string           `json:"message"`
		TaggedUsersGids []int            `json:"tagged_users_gids"`
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

var bucket = leakybucket.NewCollector(2, 2, true)
var minuteBucket = leakybucket.NewCollector(0.5, 1, true)

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
		boiler.PlayerBanWhere.ManuallyUnbanByID.IsNull(),
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

	lastBattleNum := 0
	lastBattle, err := boiler.Battles(
		qm.Select(boiler.BattleColumns.BattleNumber),
		qm.OrderBy(fmt.Sprintf("%s %s", boiler.BattleColumns.BattleNumber, "DESC")),
		boiler.BattleWhere.EndedAt.IsNotNull()).One(gamedb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Unable to get last battle for chat")
	}

	if lastBattle != nil {
		lastBattleNum = lastBattle.BattleNumber
	}

	_, totalMultiplier, isCitizen := multipliers.GetPlayerMultipliersForBattle(player.ID, lastBattleNum)

	taggedUsersGid := make(map[int]bool)
	for _, gid := range req.Payload.TaggedUsersGids {
		taggedUsersGid[gid] = false
	}

	textMsgMetadata := &TextMessageMetadata{
		Likes:           &Likes{0, 0, 0},
		TaggedUsersRead: taggedUsersGid,
	}

	var jsonTextMsgMeta null.JSON
	err = jsonTextMsgMeta.Marshal(textMsgMetadata)
	if err != nil {
		return terror.Error(err, "Could not marshal json")
	}
	// check if the faction id is provided
	if !req.Payload.FactionID.IsNil() {
		if !player.FactionID.Valid || player.FactionID.String == "" {
			return terror.Error(terror.ErrInvalidInput, "Required to join a faction to send message, please enlist in a faction.")
		}

		if player.FactionID.String != req.Payload.FactionID.String() {
			return terror.Error(terror.ErrForbidden, "Users are not allow to join the faction chat which they are not belong to.")
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
			Metadata:        jsonTextMsgMeta,
		}

		err = cm.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable to insert msg into chat history")
		}

		chatMessage := &ChatMessage{
			ID:     cm.ID,
			Type:   boiler.ChatMSGTypeEnumTEXT,
			SentAt: cm.CreatedAt,
			Data: &MessageText{
				ID:              cm.ID,
				Message:         msg,
				MessageColor:    req.Payload.MessageColor,
				FromUser:        player,
				UserRank:        player.Rank,
				FromUserStat:    playerStat,
				TotalMultiplier: multipliers.FriendlyFormatMultiplier(totalMultiplier),
				IsCitizen:       isCitizen,
				Lang:            language,
				Metadata:        jsonTextMsgMeta,
			},
		}

		// Ability kills
		fc.API.AddFactionChatMessage(player.FactionID.String, chatMessage)

		// send message
		ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", player.FactionID.String), HubKeyFactionChatSubscribe, []*ChatMessage{chatMessage})
		reply(true)
		return nil
	}

	// global message
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
		Metadata:        jsonTextMsgMeta,
	}

	err = cm.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to insert msg into chat history")
	}

	chatMessage := &ChatMessage{
		ID:     cm.ID,
		Type:   boiler.ChatMSGTypeEnumTEXT,
		SentAt: cm.CreatedAt,
		Data: &MessageText{
			ID:              cm.ID,
			Message:         msg,
			MessageColor:    req.Payload.MessageColor,
			FromUser:        player,
			UserRank:        player.Rank,
			FromUserStat:    playerStat,
			TotalMultiplier: multipliers.FriendlyFormatMultiplier(totalMultiplier),
			IsCitizen:       isCitizen,
			Lang:            language,
			Metadata:        jsonTextMsgMeta,
		},
	}

	fc.API.GlobalChat.AddMessage(chatMessage)
	ws.PublishMessage("/public/global_chat", HubKeyGlobalChatSubscribe, []*ChatMessage{chatMessage})
	reply(true)

	return nil
}

type ReadTaggedMessageRequest struct {
	Payload struct {
		ChatHistoryID string `json:"chat_history_id"`
	} `json:"payload"`
}

const HubKeyReadTaggedMessage = "READ:TAGGED:MESSAGE"

func (fc *ChatController) ReadTaggedMessageHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "ReadTaggedMessageHandler").Str("user_id", user.ID).Logger()
	genericErrorMessage := "Unable to mark message as read, try again or contact support."

	req := &ReadTaggedMessageRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		l.Error().Err(err).Msg("json unmarshal error")
		return terror.Error(err, "Invalid request received.")
	}

	l = l.With().Interface("ChatHistoryID", req.Payload.ChatHistoryID).Logger()
	chatHistory, err := boiler.FindChatHistory(gamedb.StdConn, req.Payload.ChatHistoryID)
	if err != nil {
		l.Error().Err(err).Msg("unable to retrieve chat history message from ID.")
		return terror.Error(err, genericErrorMessage)
	}

	metadata := &TextMessageMetadata{}

	l = l.With().Interface("UnmarshalMetadata", chatHistory.Metadata).Logger()
	err = chatHistory.Metadata.Unmarshal(metadata)
	if err != nil {
		l.Error().Err(err).Msg("unable to unmarshal chat history metadata.")
		return terror.Error(err, genericErrorMessage)
	}

	metadata.TaggedUsersRead[user.Gid] = true

	l = l.With().Interface("MarshalMetadata", metadata).Logger()
	var jsonTextMsgMeta null.JSON
	err = jsonTextMsgMeta.Marshal(metadata)
	if err != nil {
		l.Error().Err(err).Msg("unable to marshal updated metadata.")
		return terror.Error(err, genericErrorMessage)
	}

	l = l.With().Interface("UpdateMetadata", jsonTextMsgMeta).Logger()
	chatHistory.Metadata = jsonTextMsgMeta
	_, err = chatHistory.Update(gamedb.StdConn, boil.Whitelist(boiler.ChatHistoryColumns.Metadata))
	if err != nil {
		l.Error().Err(err).Msg("unable to update chat history to mark as read.")
		return terror.Error(err, genericErrorMessage)
	}

	l = l.With().Interface("UpdateCachedMessages", chatHistory).Logger()
	// change metadata of a specific message
	fn := func(chatMessage *ChatMessage) bool {
		if chatMessage.ID != chatHistory.ID {
			return true
		}

		mt, ok := chatMessage.Data.(*MessageText)
		if ok {
			mt.Metadata = chatHistory.Metadata
		}

		return false
	}

	switch chatHistory.ChatStream {
	case server.RedMountainFactionID:
		fc.API.RedMountainChat.Range(fn)
	case server.BostonCyberneticsFactionID:
		fc.API.BostonChat.Range(fn)
	case server.ZaibatsuFactionID:
		fc.API.ZaibatsuChat.Range(fn)
	default:
		fc.API.GlobalChat.Range(fn)
	}

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

func (api *API) BroadcastNewBattle(battleNumber int) error {
	factions, err := boiler.Factions().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Could not get all factions, try again or contact support.")
	}

	cm := &ChatMessage{
		Type:   ChatMessageTypeNewBattle,
		SentAt: time.Now(),
		Data:   MessageNewBattle{BattleNumber: battleNumber},
	}

	var jsonMeta null.JSON
	err = jsonMeta.Marshal(cm)
	if err != nil {
		return err
	}

	for _, faction := range factions {
		ch := &boiler.ChatHistory{
			FactionID:       faction.ID,
			PlayerID:        server.SupremacyBattleUserID,
			MessageColor:    "",
			Text:            "",
			MSGType:         boiler.ChatMSGTypeEnumNEW_BATTLE,
			ChatStream:      faction.ID,
			UserRank:        "",
			TotalMultiplier: "",
			KillCount:       "",
			IsCitizen:       false,
			Lang:            "",
			Metadata:        jsonMeta,
		}
		err = ch.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			return terror.Error(err, "Could not create NEW_BATTLE message in chat history.")
		}
	}

	ch := &boiler.ChatHistory{
		FactionID:       server.RedMountainFactionID,
		PlayerID:        server.SupremacyBattleUserID,
		MessageColor:    "",
		Text:            "",
		MSGType:         boiler.ChatMSGTypeEnumNEW_BATTLE,
		ChatStream:      "global",
		UserRank:        "",
		TotalMultiplier: "",
		KillCount:       "",
		IsCitizen:       false,
		Lang:            "",
		Metadata:        jsonMeta,
	}
	err = ch.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Could not create NEW_BATTLE message in chat history.")
	}

	api.RedMountainChat.AddMessage(cm)
	ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", server.RedMountainFactionID), HubKeyFactionChatSubscribe, []*ChatMessage{cm})

	api.BostonChat.AddMessage(cm)
	ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", server.BostonCyberneticsFactionID), HubKeyFactionChatSubscribe, []*ChatMessage{cm})

	api.ZaibatsuChat.AddMessage(cm)
	ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", server.ZaibatsuFactionID), HubKeyFactionChatSubscribe, []*ChatMessage{cm})

	api.GlobalChat.AddMessage(cm)
	ws.PublishMessage("/public/global_chat", HubKeyGlobalChatSubscribe, []*ChatMessage{cm})

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

type ChatBanPlayerRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		PlayerID        string `json:"player_id"`
		Reason          string `json:"reason"`
		DurationMinutes int    `json:"duration_minutes"`
	} `json:"payload"`
}

const HubKeyChatBanPlayer = "CHAT:BAN:PLAYER"

func (fc *ChatController) ChatBanPlayerHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "ChatBanUserHandler").Str("user_id", user.ID).Logger()

	req := &ChatBanPlayerRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		l.Error().Err(err).Msg("json unmarshal error")
		return terror.Error(err, "Invalid request received.")
	}
	l = l.With().Interface("payload", req).Logger()

	hasPermission, err := boiler.PlayersFeatures(
		boiler.PlayersFeatureWhere.PlayerID.EQ(user.ID),
		boiler.PlayersFeatureWhere.FeatureName.EQ(boiler.FeatureNameCHAT_BAN),
	).Exists(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("failed to check if user has permission to execute chat ban")
		return terror.Error(err)
	}

	if !hasPermission {
		return terror.Error(terror.ErrUnauthorised)
	}

	if req.Payload.PlayerID == "" {
		l.Warn().Msg("player id was not given when attempting to chat ban")
		return terror.Error(terror.ErrInvalidInput, "Player must be specified.")
	}

	if req.Payload.Reason == "" {
		return terror.Error(terror.ErrInvalidInput, "A reason must be provided.")
	}

	if req.Payload.DurationMinutes == 0 {
		return terror.Error(terror.ErrInvalidInput, "A duration must be provided.")
	}

	if req.Payload.PlayerID == user.ID {
		return terror.Error(terror.ErrForbidden, "You cannot ban yourself.")
	}

	player, err := boiler.FindPlayer(gamedb.StdConn, req.Payload.PlayerID)
	if err != nil {
		l.Error().Err(err).Msg("could not find player associated with player ID")
		return terror.Error(err, "Something went wrong while trying to ban this player. Please try again.")
	}

	isAlreadyBanned, err := boiler.PlayerBans(
		boiler.PlayerBanWhere.BannedPlayerID.EQ(req.Payload.PlayerID),
		boiler.PlayerBanWhere.BanSendChat.EQ(true),
		boiler.PlayerBanWhere.EndAt.GT(time.Now()),
	).One(gamedb.StdConn)
	if err == nil {
		l.Warn().Interface("existingPlayerBan", isAlreadyBanned).Msg("player is already chat banned, skipping")
		hours := int(time.Until(isAlreadyBanned.EndAt).Hours())
		expiresIn := fmt.Sprintf("%d hour(s)", hours)
		if hours < 1 {
			expiresIn = fmt.Sprintf("%d minute(s)", int(time.Until(isAlreadyBanned.EndAt).Minutes()))
		}
		return terror.Error(terror.ErrForbidden, fmt.Sprintf("Player %s is already chat banned. Reason: '%s'. Expires in: %s.", player.Username.String, isAlreadyBanned.Reason, expiresIn))
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		l.Error().Err(err).Msg("failed to check if player is already chat banned")
		return terror.Error(err, "Something went wrong while trying to ban this player. Please try again.")
	}

	duration := time.Duration(time.Minute * time.Duration(req.Payload.DurationMinutes))
	pb := &boiler.PlayerBan{
		BanFrom:        boiler.BanFromTypeADMIN,
		BannedPlayerID: req.Payload.PlayerID,
		BannedByID:     user.ID,
		Reason:         req.Payload.Reason,
		BannedAt:       time.Now(),
		EndAt:          time.Now().Add(duration),
		BanSendChat:    true,
	}
	l = l.With().Interface("playerBan", pb).Logger()
	err = pb.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		l.Error().Err(err).Msg("failed to create new player_ban entry in db")
		return terror.Error(err, "Something went wrong while trying to ban this player. Please try again or contact")
	}

	reply(true)

	return nil
}
