package api

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-syndicate/hub"
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

	return uch
}

const HubKeyUserOnlineStatus hub.HubCommandKey = hub.HubCommandKey("USER:ONLINE_STATUS")
const HubKeyUserSubscribe hub.HubCommandKey = hub.HubCommandKey("USER:SUBSCRIBE")
