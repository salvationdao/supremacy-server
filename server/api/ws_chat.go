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
	"time"

	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/friendsofgo/errors"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/sasha-s/go-deadlock"

	goaway "github.com/TwiN/go-away"
	"github.com/jackc/pgx/v4/pgxpool"
	leakybucket "github.com/kevinms/leakybucket-go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var Profanities = []string{
	"fag",
	"fuck",
	"nigga",
	"nigger",
	"rape",
	"retard",
}

const PersistChatMessageLimit = 50

var profanityDetector = goaway.NewProfanityDetector().WithCustomDictionary(Profanities, []string{}, []string{})
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

func (c *Chatroom) Range(fn func(chatMessage *ChatMessage) bool) {
	c.RLock()
	for _, message := range c.messages {
		if !fn(message) {
			break
		}
	}
	c.RUnlock()
}

func NewChatroom(factionID *server.FactionID) *Chatroom {
	stream := "global"
	if factionID != nil {
		stream = factionID.String()
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
			player, _ = boiler.FindPlayer(gamedb.StdConn, msg.PlayerID)
			if player == nil {
				continue
			}
			playerStat, err := db.UserStatsGet(player.ID)
			if err != nil {
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

	chatroom := &Chatroom{
		factionID: factionID,
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

	api.Command(HubKeyChatPastMessages, chatHub.ChatPastMessagesHandler)
	api.SecureUserCommand(HubKeyChatMessage, chatHub.ChatMessageHandler)

	api.SubscribeCommand(HubKeyGlobalChatSubscribe, chatHub.GlobalChatUpdatedSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyFactionChatSubscribe, chatHub.FactionChatUpdatedSubscribeHandler)

	return chatHub
}

// FactionChatRequest sends chat message to specific faction.
type FactionChatRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID    server.FactionID `json:"faction_id"`
		MessageColor string           `json:"message_color"`
		Message      string           `json:"message"`
	} `json:"payload"`
}

const HubKeyChatMessage hub.HubCommandKey = "CHAT:MESSAGE"

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
func (fc *ChatController) ChatMessageHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Issue sending message in chat, try again or contact support."
	b1 := bucket.Add(hubc.Identifier(), 1)
	b2 := minuteBucket.Add(hubc.Identifier(), 1)

	if b1 == 0 || b2 == 0 {
		return terror.Error(fmt.Errorf("too many messages"), "Too many messages.")
	}

	req := &FactionChatRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	player, err := boiler.Players(
		qm.Select(
			boiler.PlayerColumns.ID,
			boiler.PlayerColumns.Username,
			boiler.PlayerColumns.Gid,
			boiler.PlayerColumns.FactionID,
			boiler.PlayerColumns.Rank,
			boiler.PlayerColumns.SentMessageCount,
		),
		boiler.PlayerWhere.ID.EQ(hubc.Identifier()),
	).One(gamedb.StdConn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// check user is banned on chat
	isBanned, err := player.PunishedPlayers(
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s on %s = %s and %s = ?",
				boiler.TableNames.PunishOptions,
				qm.Rels(boiler.TableNames.PunishOptions, boiler.PunishOptionColumns.ID),
				qm.Rels(boiler.TableNames.PunishedPlayers, boiler.PunishedPlayerColumns.PunishOptionID),
				qm.Rels(boiler.TableNames.PunishOptions, boiler.PunishOptionColumns.Key),
			),
			server.PunishmentOptionRestrictChat,
		),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to check player on the banned list")
		return terror.Error(err)
	}

	// if chat banned just return
	if isBanned {
		return terror.Error(fmt.Errorf("player is banned to chat"), "You are banned to chat")
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

	msg = profanityDetector.Censor(msg)
	if len(msg) > 280 {
		msg = firstN(msg, 280)
	}
	// get player current stat
	playerStat, err := db.UserStatsGet(player.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Unable to get player stat from db")
	}

	_, totalMultiplier, isCitizen := multipliers.GetPlayerMultipliersForBattle(player.ID, 10)
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
				FromUser:        *player,
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
		fc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionChatSubscribe, player.FactionID.String)), chatMessage)
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
			FromUser:        *player,
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

	fc.API.MessageBus.Send(messagebus.BusKey(HubKeyGlobalChatSubscribe), chatMessage)
	reply(true)

	return nil
}

// ChatPastMessagesRequest sends chat message to specific faction.
type ChatPastMessagesRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID server.FactionID `json:"faction_id"`
	} `json:"payload"`
}

const HubKeyChatPastMessages hub.HubCommandKey = "CHAT:PAST_MESSAGES"

func (fc *ChatController) ChatPastMessagesHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &ChatPastMessagesRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if !req.Payload.FactionID.IsNil() {
		uuidString := hubc.Identifier() // identifier gets set on auth by default, so no ident = not authed
		if uuidString == "" {
			return terror.Error(terror.ErrUnauthorised, "Unauthorised access.")
		}
	}

	resp := []*ChatMessage{}
	chatRangeHandler := func(message *ChatMessage) bool {
		resp = append(resp, message)
		return true
	}

	switch req.Payload.FactionID {
	case server.RedMountainFactionID:
		fc.API.RedMountainChat.Range(chatRangeHandler)
	case server.BostonCyberneticsFactionID:
		fc.API.BostonChat.Range(chatRangeHandler)
	case server.ZaibatsuFactionID:
		fc.API.ZaibatsuChat.Range(chatRangeHandler)
	default:
		fc.API.GlobalChat.Range(chatRangeHandler)
	}

	reply(resp)

	return nil
}

const HubKeyFactionChatSubscribe hub.HubCommandKey = "FACTION:CHAT:SUBSCRIBE"

func (fc *ChatController) FactionChatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	errMsg := "Could not subscribe to faction chat updates, try again or contact support."
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received.")
	}

	// get player in valid faction
	player, err := boiler.FindPlayer(gamedb.StdConn, client.Identifier())
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}
	if !player.FactionID.Valid || player.FactionID.String == uuid.Nil.String() {
		return "", "", terror.Error(terror.ErrInvalidInput, "Require to join faction to receive messages.")
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionChatSubscribe, player.FactionID.String)), nil
}

const HubKeyGlobalChatSubscribe hub.HubCommandKey = "GLOBAL:CHAT:SUBSCRIBE"

func (fc *ChatController) GlobalChatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Could not subscribe to global chat updates, try again or contact support.")
	}
	return req.TransactionID, messagebus.BusKey(HubKeyGlobalChatSubscribe), nil
}

func (api *API) AddFactionChatMessage(factionID string, msg *ChatMessage) {
	switch factionID {
	case server.RedMountainFactionID.String():
		api.RedMountainChat.AddMessage(msg)
	case server.BostonCyberneticsFactionID.String():
		api.BostonChat.AddMessage(msg)
	case server.ZaibatsuFactionID.String():
		api.ZaibatsuChat.AddMessage(msg)
	}
}
