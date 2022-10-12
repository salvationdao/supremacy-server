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
	"sort"
	"time"

	"github.com/sasha-s/go-deadlock"
	"golang.org/x/exp/slices"

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
	ChatMessageTypeModBan     ChatMessageType = "MOD_BAN"
	ChatMessageTypeNewBattle  ChatMessageType = "NEW_BATTLE"
)

type MessageText struct {
	ID           string           `json:"id"`
	Message      string           `json:"message"`
	MessageColor string           `json:"message_color"`
	FromUser     boiler.Player    `json:"from_user"`
	UserRank     string           `json:"user_rank"`
	FromUserStat *server.UserStat `json:"from_user_stat"`
	Lang         string           `json:"lang"`
	// TotalMultiplier string           `json:"total_multiplier"`
	// IsCitizen       bool             `json:"is_citizen"`
	BattleNumber int       `json:"battle_number"`
	Metadata     null.JSON `json:"metadata"`
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
	Likes    []string `json:"likes"`
	Dislikes []string `json:"dislikes"`
	Net      int      `json:"net"`
}

type TextMessageMetadata struct {
	TaggedUsersRead map[int]bool `json:"tagged_users_read"`
	Likes           *Likes       `json:"likes"`
	Reports         []string     `json:"reports"`
}

// Chatroom holds a specific chat room
type Chatroom struct {
	deadlock.RWMutex
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

func (c *Chatroom) ReadRange(fn func(chatMessage *ChatMessage) bool) {
	c.RLock()
	defer c.RUnlock()
	for _, message := range c.messages {
		if !fn(message) {
			break
		}
	}
}

func (c *Chatroom) WriteRange(fn func(chatMessage *ChatMessage) bool) {
	c.Lock()
	defer c.Unlock()
	for _, message := range c.messages {
		if !fn(message) {
			break
		}
	}
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
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				gamelog.L.Warn().Err(err).Interface("player.ID", player.ID).Msg("issue UserStatsGet")
				continue
			}

			if playerStat == nil {
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
				ID:           msg.ID,
				Message:      msg.Text,
				MessageColor: msg.MessageColor,
				FromUser:     *player,
				UserRank:     player.Rank,
				FromUserStat: stat,
				Metadata:     msg.Metadata,
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
	api.SecureUserCommand(HubKeyReactToMessage, chatHub.ReactToMessageHandler)
	api.SecureUserCommand(HubKeyChatReport, chatHub.ChatReportHandler)

	go api.MessageBroadcaster()

	return chatHub
}

const (
	RestrictionLocationSelect = "Select location"
	RestrictionAbilityTrigger = "Trigger abilities"
	RestrictionChatSend       = "Send chat"
	RestrictionChatView       = "Receive chat"
	RestrictionSupsContribute = "Contribute sups"
	RestrictionsMechQueuing   = "Mech queuing"
)

func (api *API) MessageBroadcaster() {
	for {
		select {
		case msg := <-api.ArenaManager.SystemBanManager.SystemBanMassageChan:
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
		case newBattleInfo := <-api.ArenaManager.NewBattleChan:
			err := api.BroadcastNewBattle(newBattleInfo.ID, newBattleInfo.BattleNumber)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("Could not broadcast battle info ", newBattleInfo).Msg("failed to broadcast new battle info")
				return
			}
		}
	}
}

func (api *API) updateMessageMetadata(chatHistory *boiler.ChatHistory, jsonTextMsgMeta null.JSON, logger zerolog.Logger) error {
	logger = logger.With().Interface("updateMessageMetadata", chatHistory.ID).Logger()
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

	p, err := boiler.FindPlayer(gamedb.StdConn, chatHistory.PlayerID)
	if err != nil {
		logger.Error().Err(err).Msg("could not find chatHistory player")
		return err
	}

	player := boiler.Player{
		ID:               p.ID,
		Username:         p.Username,
		Gid:              p.Gid,
		FactionID:        p.FactionID,
		Rank:             p.Rank,
		SentMessageCount: p.SentMessageCount,
	}

	playerStat, err := db.UserStatsGet(player.ID)
	if err != nil {
		logger.Error().Err(err).Msg("could get player stats")
		return err
	}

	chatMessage := &ChatMessage{
		ID:     chatHistory.ID,
		Type:   boiler.ChatMSGTypeEnumTEXT,
		SentAt: chatHistory.CreatedAt,
		Data: &MessageText{
			ID:           chatHistory.ID,
			Message:      chatHistory.Text,
			MessageColor: chatHistory.MessageColor,
			FromUser:     player,
			UserRank:     player.Rank,
			FromUserStat: playerStat,
			Lang:         chatHistory.Lang,
			Metadata:     jsonTextMsgMeta,
		},
	}

	logger = logger.With().Interface("publishMetadata", chatMessage.Data).Logger()
	switch chatHistory.ChatStream {
	case server.RedMountainFactionID:
		api.RedMountainChat.WriteRange(fn)
		ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", server.RedMountainFactionID), HubKeyFactionChatSubscribe, []*ChatMessage{chatMessage})
	case server.BostonCyberneticsFactionID:
		api.BostonChat.WriteRange(fn)
		ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", server.BostonCyberneticsFactionID), HubKeyFactionChatSubscribe, []*ChatMessage{chatMessage})

	case server.ZaibatsuFactionID:
		api.ZaibatsuChat.WriteRange(fn)
		ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", server.ZaibatsuFactionID), HubKeyFactionChatSubscribe, []*ChatMessage{chatMessage})
	default:
		api.GlobalChat.WriteRange(fn)
		ws.PublishMessage("/public/global_chat", HubKeyGlobalChatSubscribe, []*ChatMessage{chatMessage})
	}

	return nil
}

// FactionChatRequest sends chat message to specific faction.
type FactionChatRequest struct {
	Payload struct {
		Id              string           `json:"id"`
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

var bucket = leakybucket.NewCollector(4, 1, true) // 4 msg per second

// ChatMessageHandler sends chat message from player
func (fc *ChatController) ChatMessageHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	b1 := bucket.Add(user.ID, 1)

	if b1 == 0 {
		return terror.Warn(fmt.Errorf("too many messages"), "Too many messages.")
	}

	req := &FactionChatRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if !user.FactionID.Valid {
		return terror.Error(terror.ErrForbidden, "You must be enrolled in a faction to chat.")
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
	).One(gamedb.StdConn)
	if err == nil {
		// if chat banned just return
		hours := int(time.Until(isBanned.EndAt).Hours())
		expiresIn := fmt.Sprintf("%d hour(s)", hours)
		if hours < 1 {
			expiresIn = fmt.Sprintf("%d minute(s)", int(time.Until(isBanned.EndAt).Minutes()))
		}
		return terror.Error(fmt.Errorf("player is banned to chat"), fmt.Sprintf("You are banned from chatting. Your ban ends in %s.", expiresIn))

	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		gamelog.L.Error().Err(err).Msg("Failed to check player on the banned list")
		return err
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

	// disabled for alex's request 25/08/2022
	//msg = fc.API.ProfanityManager.Detector.Censor(msg)
	if len(msg) > 280 {
		msg = firstN(msg, 280)
	}
	// get player current stat
	playerStat, err := db.UserStatsGet(player.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Unable to get player stat from db")
	}

	taggedUsersGid := make(map[int]bool)
	for _, gid := range req.Payload.TaggedUsersGids {
		taggedUsersGid[gid] = false
	}

	textMsgMetadata := &TextMessageMetadata{
		Likes:           &Likes{[]string{}, []string{}, 0},
		TaggedUsersRead: taggedUsersGid,
		Reports:         []string{},
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
			ID:              req.Payload.Id,
			FactionID:       player.FactionID.String,
			PlayerID:        player.ID,
			MessageColor:    req.Payload.MessageColor,
			BattleID:        null.String{},
			MSGType:         boiler.ChatMSGTypeEnumTEXT,
			UserRank:        player.Rank,
			TotalMultiplier: "",
			KillCount:       fmt.Sprintf("%d", playerStat.AbilityKillCount),
			Text:            msg,
			ChatStream:      player.FactionID.String,
			IsCitizen:       false,
			Lang:            language,
			Metadata:        jsonTextMsgMeta,
		}

		err = cm.Insert(gamedb.StdConn, boil.Infer())
		if err != nil {
			gamelog.L.Error().Err(err).Msg("unable to insert msg into chat history")
		}

		// check player quest reward
		fc.API.questManager.ChatMessageQuestCheck(user.ID)

		chatMessage := &ChatMessage{
			ID:     cm.ID,
			Type:   boiler.ChatMSGTypeEnumTEXT,
			SentAt: cm.CreatedAt,
			Data: &MessageText{
				ID:           cm.ID,
				Message:      msg,
				MessageColor: req.Payload.MessageColor,
				FromUser:     player,
				UserRank:     player.Rank,
				FromUserStat: playerStat,
				Lang:         language,
				Metadata:     jsonTextMsgMeta,
			},
		}

		// add message to chatroom
		fc.API.AddFactionChatMessage(player.FactionID.String, chatMessage)

		// send message
		ws.PublishMessage(fmt.Sprintf("/faction/%s/faction_chat", player.FactionID.String), HubKeyFactionChatSubscribe, []*ChatMessage{chatMessage})
		reply(true)
		return nil
	}

	// global message
	cm := boiler.ChatHistory{
		ID:              req.Payload.Id,
		FactionID:       player.FactionID.String,
		PlayerID:        player.ID,
		MessageColor:    req.Payload.MessageColor,
		BattleID:        null.String{},
		MSGType:         boiler.ChatMSGTypeEnumTEXT,
		UserRank:        player.Rank,
		TotalMultiplier: "",
		KillCount:       fmt.Sprintf("%d", playerStat.AbilityKillCount),
		Text:            msg,
		ChatStream:      "global",
		IsCitizen:       false,
		Lang:            language,
		Metadata:        jsonTextMsgMeta,
	}

	err = cm.Insert(gamedb.StdConn, boil.Infer())
	if err != nil {
		gamelog.L.Error().Err(err).Msg("unable to insert msg into chat history")
	}

	// check player quest reward
	fc.API.questManager.ChatMessageQuestCheck(user.ID)

	chatMessage := &ChatMessage{
		ID:     cm.ID,
		Type:   boiler.ChatMSGTypeEnumTEXT,
		SentAt: cm.CreatedAt,
		Data: &MessageText{
			ID:           cm.ID,
			Message:      msg,
			MessageColor: req.Payload.MessageColor,
			FromUser:     player,
			UserRank:     player.Rank,
			FromUserStat: playerStat,
			Lang:         language,
			Metadata:     jsonTextMsgMeta,
		},
	}

	fc.API.GlobalChat.AddMessage(chatMessage)
	ws.PublishMessage("/public/global_chat", HubKeyGlobalChatSubscribe, []*ChatMessage{chatMessage})
	reply(chatMessage)

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
	err = fc.API.updateMessageMetadata(chatHistory, jsonTextMsgMeta, l)
	if err != nil {
		l.Error().Err(err).Msg("unable to update and publish metadata")
		return terror.Error(err, genericErrorMessage)
	}

	reply(true)
	return nil
}

type ReactToMessageRequest struct {
	Payload struct {
		ChatHistoryID string `json:"chat_history_id"`
		Reaction      string `json:"reaction"`
	} `json:"payload"`
}

const HubKeyReactToMessage = "REACT:MESSAGE"

func (fc *ChatController) ReactToMessageHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	l := gamelog.L.With().Str("func", "ReactToMessageHandler").Str("user_id", user.ID).Logger()
	genericErrorMessage := "Unable to react to message, try again or contact support."

	req := &ReactToMessageRequest{}
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

	if chatHistory.PlayerID == user.ID {
		return terror.Error(fmt.Errorf("cannot react to user's own message"), "Cannot react to your own message.")
	}

	metadata := &TextMessageMetadata{}

	l = l.With().Interface("UnmarshalMetadata", chatHistory.Metadata).Logger()
	err = chatHistory.Metadata.Unmarshal(metadata)
	if err != nil {
		l.Error().Err(err).Msg("unable to unmarshal chat history metadata.")
		return terror.Error(err, genericErrorMessage)
	}

	switch req.Payload.Reaction {
	//handle likes
	case "like":

		//check if user id is in likes then "unlike"- take user name out of likes array
		i := slices.Index(metadata.Likes.Likes, user.ID)
		if i != -1 {
			metadata.Likes.Likes = slices.Delete(metadata.Likes.Likes, i, i+1)
			break
		}
		//check if user id is in dislikes then take username out of dislikes and put into likes array
		i = slices.Index(metadata.Likes.Dislikes, user.ID)
		if i != -1 {
			metadata.Likes.Dislikes = slices.Delete(metadata.Likes.Dislikes, i, i+1)
		}
		//else put into likes array
		metadata.Likes.Likes = append(metadata.Likes.Likes, user.ID)
		break
		//handle dislikes
	case "dislike":
		//check if user id is in dislikes the "undislike" - take username out of dislikes array
		i := slices.Index(metadata.Likes.Dislikes, user.ID)
		if i != -1 {
			metadata.Likes.Dislikes = slices.Delete(metadata.Likes.Dislikes, i, i+1)
			break
		}
		//check if user id is in likes then take username out of likes and put into dislikes array
		i = slices.Index(metadata.Likes.Likes, user.ID)
		if i != -1 {
			metadata.Likes.Likes = slices.Delete(metadata.Likes.Likes, i, i+1)
		}
		//else put into dislikes array
		metadata.Likes.Dislikes = append(metadata.Likes.Dislikes, user.ID)
		break
	}

	metadata.Likes.Net = len(metadata.Likes.Likes) - len(metadata.Likes.Dislikes)

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
		l.Error().Err(err).Msg("unable to update message's reactions.")
		return terror.Error(err, genericErrorMessage)
	}

	l = l.With().Interface("UpdateCachedMessages", chatHistory).Logger()
	// change metadata of a specific message
	err = fc.API.updateMessageMetadata(chatHistory, jsonTextMsgMeta, l)
	if err != nil {
		l.Error().Err(err).Msg("unable to update and publish metadata")
		return terror.Error(err, genericErrorMessage)
	}

	reply(metadata.Likes)
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
		fc.API.RedMountainChat.ReadRange(chatRangeHandler)
	case server.BostonCyberneticsFactionID:
		fc.API.BostonChat.ReadRange(chatRangeHandler)
	case server.ZaibatsuFactionID:
		fc.API.ZaibatsuChat.ReadRange(chatRangeHandler)
	default:
		return terror.Error(terror.ErrInvalidInput, "Invalid faction id")
	}

	reply(resp)

	return nil
}

const HubKeyGlobalChatSubscribe = "GLOBAL:CHAT:SUBSCRIBE"

func (fc *ChatController) GlobalChatUpdatedSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	resp := []*ChatMessage{}
	fc.API.GlobalChat.ReadRange(func(message *ChatMessage) bool {
		resp = append(resp, message)
		return true
	})
	reply(resp)
	return nil
}

func (api *API) BroadcastNewBattle(id string, battleNumber int) error {
	factions, err := boiler.Factions().All(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, "Could not get all factions, try again or contact support.")
	}

	cm := &ChatMessage{
		ID:     id,
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

	exists, err := boiler.PlayerExists(gamedb.StdConn, req.Payload.PlayerID)
	if err != nil {
		l.Error().Err(err).Msg("could not check if player associated with player ID exists")
		return terror.Error(err, "Something went wrong while trying to ban this player. Please try again.")
	}

	if !exists {
		l.Warn().Msg("tried to ban player that doesn't exist")
		return terror.Error(fmt.Errorf("attempted to chat ban a player that doesnt exist"))
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
		return terror.Error(terror.ErrForbidden, fmt.Sprintf("Player is already chat banned. Reason: '%s'. Expires in: %s.", isAlreadyBanned.Reason, expiresIn))
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

type ChatReportRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		MessageID        string `json:"message_id"`
		Reason           string `json:"reason"`
		OtherDescription string `json:"other_description,omitempty"`
		Description      string `json:"description"`
	} `json:"payload"`
}

const HubKeyChatReport = "CHAT:REPORT:MESSAGE"

//leak 1 request per user per 30 sec
var chatReportBucket = leakybucket.NewCollector(1, 30, true)

func (fc *ChatController) ChatReportHandler(ctx context.Context, user *boiler.Player, key string, payload []byte, reply ws.ReplyFunc) error {
	reportBucket := chatReportBucket.Add(user.ID, 30)

	//only allow request to go through if it has been 30 seconds since user's last report
	if reportBucket < 30 {
		return terror.Error(fmt.Errorf("too many reports"), "Too many report requests, your original request would have contained the context of the message.")
	}

	l := gamelog.L.With().Str("func", "ChatReportHandler").Str("user_id", user.ID).Logger()

	genericErrorMessage := "Could not report message, try again or contact support."
	req := &ChatReportRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		l.Error().Err(err).Msg("json unmarshal error")
		return terror.Error(err, "Invalid request received.")
	}

	l = l.With().Interface("payload", req).Interface("ChatHistoryID", req.Payload.MessageID).Logger()
	chatHistory, err := boiler.FindChatHistory(gamedb.StdConn, req.Payload.MessageID)
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

	//check if user has already reported, if so return error
	i := slices.Index(metadata.Reports, user.ID)
	if i != -1 {
		return terror.Error(fmt.Errorf("user attempted to report message more than once"), "Cannot report message more than once, support will act on this ticket as soon as possible.")
	}

	//get user who sent offending msg
	l = l.With().Interface("GetReportedPlayer", chatHistory.PlayerID).Logger()
	reportedPlayer, err := boiler.FindPlayer(gamedb.StdConn, chatHistory.PlayerID)
	if err != nil {
		l.Error().Err(err).Msg("unable to get reported player")
		return terror.Error(err, genericErrorMessage)
	}

	//get 5 mins before and 5 mins after specific to chat stream
	l = l.With().Interface("GetContext", chatHistory.ID).Logger()
	msgs, err := boiler.ChatHistories(
		boiler.ChatHistoryWhere.ChatStream.EQ(chatHistory.ChatStream),
		boiler.ChatHistoryWhere.CreatedAt.GT(chatHistory.CreatedAt.Add(time.Minute*(-5))),
		boiler.ChatHistoryWhere.CreatedAt.LT(chatHistory.CreatedAt.Add(time.Minute*5)),
		qm.OrderBy("created_at"),
	).All(gamedb.StdConn)
	if err != nil {
		l.Error().Err(err).Msg("unable to unmarshal chat history metadata.")
		return terror.Error(err, genericErrorMessage)
	}

	l = l.With().Interface("CreateContextString", chatHistory.ID).Logger()
	reportContext := ""
	for _, msg := range msgs {
		p, err := boiler.FindPlayer(gamedb.StdConn, msg.PlayerID)
		if err != nil {
			l.Error().Err(err).Msg("unable to find player id.")
			return terror.Error(err, genericErrorMessage)
		}

		if msg.ID == chatHistory.ID {
			reportContext = reportContext + fmt.Sprintf(" \n ")
		}
		reportContext = reportContext + fmt.Sprintf("[%s] %s(%s): %s \n", msg.CreatedAt, p.Username.String, p.ID, msg.Text)
		if msg.ID == chatHistory.ID {
			reportContext = reportContext + fmt.Sprintf(" \n ")
		}
	}
	reason := req.Payload.Reason
	if reason == "Other" {
		reason = req.Payload.Reason + "- " + req.Payload.OtherDescription
	}

	subject := fmt.Sprintf("Reported Player - %s(%s): %s", reportedPlayer.Username.String, reportedPlayer.ID, reason)
	comment := fmt.Sprintf("Messager/Offender: %s(%s) \n Reported By: %s(%s) \n \n Message ID: %s \n Message: %s \n Reporter Comment: %s \n \n Context: \n %s", reportedPlayer.Username.String, reportedPlayer.ID, user.Username.String, user.ID, chatHistory.ID, chatHistory.Text, req.Payload.Description, reportContext)

	//send through to zendesk

	l = l.With().Interface("NewZendeskRequest", chatHistory.ID).Logger()
	_, err = fc.API.Zendesk.NewRequest(user.Username.String, user.ID, subject, comment, "Chat Report")
	if err != nil {
		l.Error().Err(err).Msg("unable send zendesk request.")
		return terror.Error(err, genericErrorMessage)
	}

	//add user id to report metadata (cant report again)
	metadata.Reports = append(metadata.Reports, user.ID)

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
		l.Error().Err(err).Msg("unable to update chat history to update reported.")
		return terror.Error(err, genericErrorMessage)
	}

	err = fc.API.updateMessageMetadata(chatHistory, jsonTextMsgMeta, l)
	if err != nil {
		l.Error().Err(err).Msg("unable to update and publish metadata")
		return terror.Error(err, genericErrorMessage)
	}

	reply(true)

	return nil
}
