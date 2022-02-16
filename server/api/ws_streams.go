package api

import (
	"context"
	"encoding/json"
	"server/db"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

type StreamsWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

type StreamListRequest struct {
	*hub.HubCommandRequest
}

func NewStreamController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *StreamsWS {
	streamHub := &StreamsWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "game_hub"),
		API:  api,
	}

	api.SubscribeCommand(HubKeyStreamList, streamHub.StreamListSubscribeSubscribeHandler)

	return streamHub
}

const HubKeyStreamList hub.HubCommandKey = "STREAMLIST:SUBSCRIBE"

func (s *StreamsWS) StreamListSubscribeSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &StreamListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	streamList, err := db.GetStreamList(ctx, s.Conn)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	reply(streamList)

	return req.TransactionID, messagebus.BusKey(HubKeyStreamList), nil
}
