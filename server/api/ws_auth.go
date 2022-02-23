package api

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
)

// AuthControllerWS holds handlers for checking server status
type AuthControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewAuthController creates the check hub
func NewAuthController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *AuthControllerWS {
	authHub := &AuthControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "auth_hub"),
		API:  api,
	}

	api.Command(HubKeyAuthSessionIDGet, authHub.GetHubSessionID)

	return authHub
}

const HubKeyAuthSessionIDGet = hub.HubCommandKey("AUTH:SESSION:ID:GET")

// GetHubSessionID return hub client's session id for ring check authentication
func (ac *AuthControllerWS) GetHubSessionID(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	ac.API.ringCheckAuthChan <- func(rcam RingCheckAuthMap) {
		rcam[string(wsc.SessionID)] = wsc
		reply(wsc.SessionID)
	}
	return nil
}
