package api

import (
	"context"
	"encoding/json"
	"fmt"

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

	uch.API.SubscribeCommand(HubKeyUser, uch.UserSubscribeHandler)

	return uch
}

const HubKeyUser hub.HubCommandKey = "USER"

// UserSubscribeHandler to subscribe to a user
func (ctrlr *UserControllerWS) UserSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {

	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	// get hub client
	hubClientDetail, err := ctrlr.API.getClientDetailFromChannel(wsc)
	if err != nil {
		return "", "", terror.Error(err)
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUser, hubClientDetail.ID))
	return req.TransactionID, busKey, nil
}
