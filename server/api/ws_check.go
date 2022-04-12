package api

import (
	"context"

	"github.com/ninja-syndicate/hub"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

// CheckControllerWS holds handlers for checking server status
type CheckControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewCheckController creates the check hub
func NewCheckController(api *API) *CheckControllerWS {
	checkHub := &CheckControllerWS{
		API: api,
	}

	api.Command(HubKeyCheck, checkHub.Handler)

	return checkHub
}

// HubKeyCheck is used to route to the  handler
const HubKeyCheck = hub.HubCommandKey("CHECK")

type CheckResponse struct {
	Check string `json:"check"`
}

func (ch *CheckControllerWS) Handler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	response := CheckResponse{Check: "ok"}
	err := check(ctx, ch.Conn)
	if err != nil {
		response.Check = err.Error()
	}
	reply(response)

	return nil
}
