package api

import (
	"context"
	"encoding/json"
	"fmt"
	"server"
	"server/passport"

	"github.com/ninja-software/hub/v2/ext/messagebus"
)

type PassportUserOnlineStatusRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		UserID server.UserID `json:"userID"`
		Status bool          `json:"status"`
	} `json:"payload"`
}

func (api *API) PassportUserOnlineStatusHandler(ctx context.Context, payload []byte) {
	req := &PassportUserOnlineStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport user online handler request")
	}

	// TODO: maybe add a difference between passport online and gameserver online
	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, req.Payload.UserID)), req.Payload.Status)
}

type PassportUserUpdatedRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		User *server.User `json:"user"`
	} `json:"payload"`
}

func (api *API) PassportUserUpdatedHandler(ctx context.Context, payload []byte) {
	req := &PassportUserUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport user updated handler request")
	}

	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, req.Payload.User.ID)), req.Payload.User)
}

type PassportUserSupsUpdatedRequest struct {
	Key     passport.Event `json:"key"`
	Payload struct {
		UserID server.UserID `json:"userID"`
		Sups   server.BigInt `json:"sups"`
	} `json:"payload"`
}

func (api *API) PassportUserSupsUpdatedHandler(ctx context.Context, payload []byte) {
	req := &PassportUserSupsUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		api.Log.Err(err).Msg("error unmarshalling passport user sups updated request")
	}

	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsUpdated, req.Payload.UserID)), req.Payload.Sups.String())
}
