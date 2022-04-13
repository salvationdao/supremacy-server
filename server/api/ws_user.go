package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
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

func NewUserController(api *API) *UserControllerWS {
	uch := &UserControllerWS{
		API: api,
	}

	api.SecureUserSubscribeCommand(HubKeyUserSubscribe, uch.UserSubscribeHandler)

	return uch
}

const HubKeyUserOnlineStatus hub.HubCommandKey = hub.HubCommandKey("USER:ONLINE_STATUS")
const HubKeyUserRingCheck hub.HubCommandKey = hub.HubCommandKey("RING:CHECK")

type UserUpdatedRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID string `json:"id"`
	} `json:"payload"`
}

const HubKeyUserSubscribe hub.HubCommandKey = hub.HubCommandKey("USER:SUBSCRIBE")

func (uch *UserControllerWS) UserSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &UserUpdatedRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, req.Payload.ID)), nil
}
