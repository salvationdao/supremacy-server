package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
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
	api.Command(HubKeyTwitchJWTAuth, authHub.JWTAuth)

	return authHub
}

const HubKeyAuthSessionIDGet = hub.HubCommandKey("AUTH:SESSION:ID:GET")

// GetHubSessionID return hub client's session id for ring check authentication
func (ac *AuthControllerWS) GetHubSessionID(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	ac.API.ringCheckAuthChan <- func(rcam RingCheckAuthMap) {
		rcam[string(wsc.SessionID)] = wsc
	}

	reply(wsc.SessionID)
	return nil
}

// TwitchAuthRequest authenticate a twitch user
type TwitchAuthRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TwitchToken string `json:"twitchToken"`
	} `json:"payload"`
}

const HubKeyTwitchJWTAuth = hub.HubCommandKey("TWITCH:JWT:AUTH")

func (ac *AuthControllerWS) JWTAuth(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &TwitchAuthRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	ac.API.ringCheckAuthChan <- func(rca RingCheckAuthMap) {
		rca[req.Payload.TwitchToken] = wsc
	}

	// distroy the token in 30 second
	go func() {
		time.Sleep(600 * time.Second)

		ac.API.ringCheckAuthChan <- func(rca RingCheckAuthMap) {
			_, ok := rca[req.Payload.TwitchToken]
			if ok {
				delete(rca, req.Payload.TwitchToken)
			}
		}
	}()

	reply(true)

	return nil
}
