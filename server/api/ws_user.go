package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/passport"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/messagebus"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

type UserControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewUserController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *UserControllerWS {
	uch := &UserControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "twitch_hub"),
		API:  api,
	}

	// TODO: delete this when passport finish user faction assign function
	uch.API.SecureUserCommand(HubKeyUserRandomFactionUpdate, uch.UserRandomUpdateFaction)

	uch.API.SecureUserSubscribeCommand(HubKeyUserSubscribe, uch.UserSubscribeHandler)
	uch.API.SecureUserSubscribeCommand(HubKeyUserOnlineStatus, uch.OnlineStatusSubscribeHandler)
	uch.API.SecureUserSubscribeCommand(HubKeyUserSupsUpdated, uch.SupsUpdateSubscribeHandler)

	return uch
}

const HubKeyUserRandomFactionUpdate hub.HubCommandKey = "USER:RANDOM:FACTION:UPDATE"

func (uc *UserControllerWS) UserRandomUpdateFaction(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	clientDetail, err := uc.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return terror.Error(err)
	}

	if !clientDetail.FactionID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "unable to reassign new faction")
	}

	randomFaction := passport.RandomFaction(uc.API.factionMap)
	err = uc.API.Passport.UserFactionUpdate(ctx, clientDetail.ID, randomFaction.ID, req.TransactionID)
	if err != nil {
		return terror.Error(err)
	}

	reply(true)
	return nil
}

const HubKeyUserSubscribe hub.HubCommandKey = "USER:SUBSCRIBE"

// UserSubscribeHandler to subscribe to a user
func (ctrlr *UserControllerWS) UserSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &HubKeyUserOnlineStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	user, err := ctrlr.API.Passport.UserGetByID(ctx, req.Payload.ID, req.TransactionID)
	if err != nil {
		return "", "", err
	}

	reply(user)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID)), nil
}

// HubKeyUserOnlineStatus subscribes to a user's online status (returns boolean)
const HubKeyUserOnlineStatus hub.HubCommandKey = "USER:ONLINE_STATUS"

// HubKeyUserOnlineStatusRequest to subscribe to user online status changes
type HubKeyUserOnlineStatusRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID       server.UserID `json:"id"`
		Username string        `json:"username"` // Optional username instead of id
	} `json:"payload"`
}

// OnlineStatusSubscribeHandler to subscribe to user online status changes
func (ctrlr *UserControllerWS) OnlineStatusSubscribeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &HubKeyUserOnlineStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	userID := req.Payload.ID
	if userID.IsNil() && req.Payload.Username == "" {
		return req.TransactionID, "", terror.Error(terror.ErrInvalidInput, "User ID or username is required")
	}
	if userID.IsNil() {
		user, err := ctrlr.API.Passport.UserGetByUsername(ctx, req.Payload.Username, req.TransactionID)
		if err != nil {
			return req.TransactionID, "", terror.Error(err, "Unable to load current user")
		}
		userID = user.ID
	}

	if userID.IsNil() {
		return req.TransactionID, "", terror.Error(fmt.Errorf("userId is still nil for %s %s", req.Payload.ID, req.Payload.Username), "Unable to load current user")
	}

	// get gameserver online status
	online := false
	ctrlr.API.Hub.Clients(func(clients hub.ClientsList) {
		for cl := range clients {
			if cl.Identifier() == userID.String() {
				online = true
				break
			}
		}
	})

	// TODO: get passport online status?

	reply(online)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, userID.String())), nil
}

const HubKeyUserSupsUpdated = hub.HubCommandKey("USER:SUPS:UPDATED")

func (uc *UserControllerWS) SupsUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsUpdated, wsc.Identifier()))
	return req.TransactionID, busKey, nil
}
