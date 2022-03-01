package api

import (
	"context"
	"encoding/json"
	"net/http"
	"server"
	"server/db"
	"server/helpers"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

type StreamsWS struct {
	Conn            *pgxpool.Pool
	Log             *zerolog.Logger
	API             *API
	ServerStreamKey string
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
	api.SubscribeCommand(HubKeyGlobalAnnouncementSubscribe, streamHub.GlobalMessageSubscribe)

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

type StreamRequest struct {
	Host string `json:"host"`
}

func (api *API) GetStreamsHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	streams, err := db.GetStreamList(r.Context(), api.Conn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}
	go api.MessageBus.Send(r.Context(), messagebus.BusKey(HubKeyStreamList), streams)

	return helpers.EncodeJSON(w, streams)
}

func (api *API) CreateStreamHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	stream := &server.Stream{}

	err := json.NewDecoder(r.Body).Decode(stream)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	err = db.CreateStream(r.Context(), api.Conn, stream)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	streamList, err := db.GetStreamList(r.Context(), api.Conn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	go api.MessageBus.Send(r.Context(), messagebus.BusKey(HubKeyVoteStageUpdated), streamList)

	return http.StatusOK, nil
}

func (api *API) DeleteStreamHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	stream := &StreamRequest{}
	err := json.NewDecoder(r.Body).Decode(stream)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	err = db.DeleteStream(r.Context(), api.Conn, stream.Host)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	streamList, err := db.GetStreamList(r.Context(), api.Conn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	go api.MessageBus.Send(r.Context(), messagebus.BusKey(HubKeyVoteStageUpdated), streamList)

	return http.StatusOK, nil
}

const HubKeyGlobalAnnouncementSubscribe hub.HubCommandKey = "GLOBAL_ANNOUNCEMENT:SUBSCRIBE"

func (s *StreamsWS) GlobalMessageSubscribe(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}
	reply(s.API.GlobalAnnouncement)
	return req.TransactionID, messagebus.BusKey(HubKeyGlobalAnnouncementSubscribe), nil
}
