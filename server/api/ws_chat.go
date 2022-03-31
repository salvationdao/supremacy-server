package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/sasha-s/go-deadlock"
	"github.com/volatiletech/null/v8"
	"html"
	"server"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/jackc/pgx/v4/pgxpool"
	leakybucket "github.com/kevinms/leakybucket-go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/log_helpers"
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

const PersistChatMessageLimit = 20

var profanityDetector = goaway.NewProfanityDetector().WithCustomDictionary(Profanities, []string{}, []string{})
var bm = bluemonday.StrictPolicy()

// ChatMessageSend contains chat message data to send.
type ChatMessageSend struct {
	Message           string      `json:"message"`
	MessageColor      string      `json:"message_color"`
	FromUserID        string      `json:"from_user_id"`
	FromUsername      string      `json:"from_username"`
	FromUserFactionID null.String `json:"from_user_faction_id"`
	SentAt            time.Time   `json:"sent_at"`

	TotalMultiplier   int64 `json:"total_multiplier"`
	TotalAbilityKills int64 `json:"total_ability_kills"`
}

// Chatroom holds a specific chat room
type Chatroom struct {
	deadlock.RWMutex
	factionID *server.FactionID
	messages  []*ChatMessageSend
}

func (c *Chatroom) AddMessage(message *ChatMessageSend) {
	c.Lock()
	c.messages = append(c.messages, message)
	if len(c.messages) >= PersistChatMessageLimit {
		c.messages = c.messages[1:]
	}
	c.Unlock()
}

func (c *Chatroom) Range(fn func(chatMessage *ChatMessageSend) bool) {
	c.RLock()
	for _, message := range c.messages {
		if !fn(message) {
			break
		}
	}
	c.RUnlock()
}

func NewChatroom(factionID *server.FactionID) *Chatroom {
	chatroom := &Chatroom{
		factionID: factionID,
		messages:  []*ChatMessageSend{},
	}
	return chatroom
}

// ChatController holds handlers for chat
type ChatController struct {
	Conn            *pgxpool.Pool
	Log             *zerolog.Logger
	API             *API
	GlobalChat      *Chatroom
	RedMountainChat *Chatroom
	BostonChat      *Chatroom
	ZaibatsuChat    *Chatroom
}

// NewChatController creates the role hub
func NewChatController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *ChatController {
	chatHub := &ChatController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "chat_hub"),
		API:  api,
		GlobalChat: &Chatroom{
			messages: []*ChatMessageSend{},
		},
		RedMountainChat: &Chatroom{
			factionID: &server.RedMountainFactionID,
			messages:  []*ChatMessageSend{},
		},
		BostonChat: &Chatroom{
			factionID: &server.BostonCyberneticsFactionID,
			messages:  []*ChatMessageSend{},
		},
		ZaibatsuChat: &Chatroom{
			factionID: &server.ZaibatsuFactionID,
			messages:  []*ChatMessageSend{},
		},
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

// rootHub.SecureCommand(HubKeyFactionChat, factionHub.ChatMessageHandler)
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

	player, err := boiler.FindPlayer(gamedb.StdConn, hubc.Identifier())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// check user is banned on chat
	isBanned, err := player.PunishedPlayers(
		boiler.PunishedPlayerWhere.PunishUntil.GT(time.Now()),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s on %s = %s and %s = 'limit_chat'",
				boiler.TableNames.PunishOptions,
				qm.Rels(boiler.TableNames.PunishOptions, boiler.PunishOptionColumns.ID),
				qm.Rels(boiler.TableNames.PunishedPlayers, boiler.PunishedPlayerColumns.PunishOptionID),
				qm.Rels(boiler.TableNames.PunishOptions, boiler.PunishOptionColumns.Key),
			),
		),
	).Exists(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to check player on the banned list")
		return terror.Error(err)
	}

	// if chat banned just return
	if isBanned {
		reply(true)
		return nil
	}

	// get faction primary colour from faction

	msg := html.UnescapeString(bm.Sanitize(req.Payload.Message))
	msg = profanityDetector.Censor(msg)
	if len(msg) > 280 {
		msg = firstN(msg, 280)
	}

	// check if the faction id is provided
	if !req.Payload.FactionID.IsNil() {
		if !player.FactionID.Valid || player.FactionID.String == "" {
			return terror.Error(terror.ErrInvalidInput, "Required to join a faction to send message, please enlist in a faction.")
		}

		if player.FactionID.String != req.Payload.FactionID.String() {
			return terror.Error(terror.ErrForbidden, "Users are not allow to join the faction chat which they are not belong to.")
		}

		chatMessage := &ChatMessageSend{
			Message:           msg,
			MessageColor:      req.Payload.MessageColor,
			FromUserID:        player.ID,
			FromUsername:      player.Username.String,
			FromUserFactionID: player.FactionID,
			SentAt:            time.Now(),
		}

		// Ability kills

		switch player.FactionID.String {
		case server.RedMountainFactionID.String():
			fc.RedMountainChat.AddMessage(chatMessage)
		case server.BostonCyberneticsFactionID.String():
			fc.BostonChat.AddMessage(chatMessage)
		case server.ZaibatsuFactionID.String():
			fc.ZaibatsuChat.AddMessage(chatMessage)
		default:
			// skip, if the faction id is not on the list
			return nil
		}

		// send message
		fc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionChatSubscribe, player.FactionID.String)), chatMessage)
		reply(true)
		return nil
	}

	// global message
	chatMessage := &ChatMessageSend{
		Message:           msg,
		MessageColor:      req.Payload.MessageColor,
		FromUserID:        player.ID,
		FromUsername:      player.Username.String,
		FromUserFactionID: player.FactionID,
		SentAt:            time.Now(),
	}
	fc.GlobalChat.AddMessage(chatMessage)
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

	resp := []*ChatMessageSend{}
	chatRangeHandler := func(message *ChatMessageSend) bool {
		resp = append(resp, message)
		return true
	}

	switch req.Payload.FactionID {
	case server.RedMountainFactionID:
		fc.RedMountainChat.Range(chatRangeHandler)
	case server.BostonCyberneticsFactionID:
		fc.BostonChat.Range(chatRangeHandler)
	case server.ZaibatsuFactionID:
		fc.ZaibatsuChat.Range(chatRangeHandler)
	default:
		fc.GlobalChat.Range(chatRangeHandler)
	}

	reply(resp)

	return nil
}

const HubKeyFactionChatSubscribe hub.HubCommandKey = "FACTION:CHAT:SUBSCRIBE"

func (fc *ChatController) FactionChatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionChatSubscribe, player.FactionID)), nil
}

const HubKeyGlobalChatSubscribe hub.HubCommandKey = "GLOBAL:CHAT:SUBSCRIBE"

func (fc *ChatController) GlobalChatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Could not subscribe to global chat updates, try again or contact support.")
	}
	return req.TransactionID, messagebus.BusKey(HubKeyGlobalChatSubscribe), nil
}
