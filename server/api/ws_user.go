package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
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

	// uch.API.SecureUserSubscribeCommand(HubKeyUserOnlineStatus, uch.OnlineStatusSubscribeHandler)
	uch.API.SecureUserSubscribeCommand(HubKeyUserSupsUpdated, uch.SupsUpdateSubscribeHandler)

	return uch
}

// // HubKeyUserOnlineStatus subscribes to a user's online status (returns boolean)

// // HubKeyUserOnlineStatusRequest to subscribe to user online status changes
// type HubKeyUserOnlineStatusRequest struct {
// 	*hub.HubCommandRequest
// 	Payload struct {
// 		ID       server.UserID `json:"id"`
// 		Username string        `json:"username"` // Optional username instead of id
// 	} `json:"payload"`
// }

// // OnlineStatusSubscribeHandler to subscribe to user online status changes
// func (ctrlr *UserControllerWS) OnlineStatusSubscribeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
// 	req := &HubKeyUserOnlineStatusRequest{}
// 	err := json.Unmarshal(payload, req)
// 	if err != nil {
// 		return req.TransactionID, "", terror.Error(err, "Invalid request received")
// 	}

// 	userID := req.Payload.ID
// 	if userID.IsNil() && req.Payload.Username == "" {
// 		return req.TransactionID, "", terror.Error(terror.ErrInvalidInput, "User ID or username is required")
// 	}
// 	if userID.IsNil() {
// 		user, err := ctrlr.API.Passport.UserGetByUsername(ctx, req.Payload.Username, req.TransactionID)
// 		if err != nil {
// 			return req.TransactionID, "", terror.Error(err, "Unable to load current user")
// 		}
// 		userID = user.ID
// 	}

// 	if userID.IsNil() {
// 		return req.TransactionID, "", terror.Error(fmt.Errorf("userID is still nil for %s %s", req.Payload.ID, req.Payload.Username), "Unable to load current user")
// 	}

// 	// get gameserver online status
// 	online := false
// 	ctrlr.API.Hub.Clients(func(clients hub.ClientsList) {
// 		for cl := range clients {
// 			if cl.Identifier() == userID.String() {
// 				online = true
// 				break
// 			}
// 		}
// 	})

// 	// TODO: get passport online status?

// 	reply(online)
// 	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, userID.String())), nil
// }

const HubKeyUserOnlineStatus hub.HubCommandKey = hub.HubCommandKey("USER:ONLINE_STATUS")
const HubKeyUserSubscribe hub.HubCommandKey = hub.HubCommandKey("USER:SUBSCRIBE")
const HubKeyUserSupsUpdated hub.HubCommandKey = hub.HubCommandKey("USER:SUPS:UPDATED")

func (uc *UserControllerWS) SupsUpdateSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsUpdated, wsc.Identifier()))
	return req.TransactionID, busKey, nil
}
