package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"server"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/helpers"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
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
	api.SubscribeCommand(HubKeyStreamCloseSubscribe, streamHub.StreamCloseSubscribeHandler)

	api.SubscribeCommand(HubKeyGlobalAnnouncementSubscribe, streamHub.GlobalAnnouncementSubscribe)

	return streamHub
}

const HubKeyStreamList hub.HubCommandKey = "STREAMLIST:SUBSCRIBE"

func (s *StreamsWS) StreamListSubscribeSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &StreamListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	if needProcess {
		streamList, err := db.GetStreamList(ctx, s.Conn)
		if err != nil {
			return req.TransactionID, "", terror.Error(err)
		}

		reply(streamList)
	}

	return req.TransactionID, messagebus.BusKey(HubKeyStreamList), nil
}

const HubKeyStreamCloseSubscribe hub.HubCommandKey = "STREAM:CLOSE:SUBSCRIBE"

//sets up subscription socket to push games left until stream closes
func (s *StreamsWS) StreamCloseSubscribeHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &StreamListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	//TODO ALEX: fix
	//gamesToClose := s.API.BattleArena.GetGamesToClose()
	//
	//reply(gamesToClose)
	return req.TransactionID, messagebus.BusKey(HubKeyStreamCloseSubscribe), nil
}

//creates api endpoint for manual override of games left until close and sends it via the subscription
func (api *API) CreateStreamCloseHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	gamesToCloseStruct := &server.GamesToCloseStream{}
	err := json.NewDecoder(r.Body).Decode(&gamesToCloseStruct)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	//TODO ALEX: fix
	//gamesToClose := gamesToCloseStruct.GamesToClose

	//api.BattleArena.PutGamesToClose(gamesToClose)

	//go api.messageBus.Send(messagebus.BusKey(HubKeyStreamCloseSubscribe), gamesToClose)

	return http.StatusOK, nil
}

func (api *API) GetStreamsHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	streams, err := db.GetStreamList(context.Background(), api.Conn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}
	go api.MessageBus.Send(messagebus.BusKey(HubKeyStreamList), streams)

	return helpers.EncodeJSON(w, streams)
}

func (api *API) CreateStreamHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	stream := &server.Stream{}

	err := json.NewDecoder(r.Body).Decode(stream)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	err = db.CreateStream(context.Background(), api.Conn, stream)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	_, err = db.GetStreamList(context.Background(), api.Conn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}
	// Alex: wtf??
	//go api.messageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), streamList)

	return http.StatusOK, nil
}

type StreamRequest struct {
	Host string `json:"host"`
}

func (api *API) DeleteStreamHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	stream := &StreamRequest{}
	err := json.NewDecoder(r.Body).Decode(stream)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	err = db.DeleteStream(context.Background(), api.Conn, stream.Host)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	_, err = db.GetStreamList(context.Background(), api.Conn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	// wtf??
	//go api.messageBus.Send(messagebus.BusKey(HubKeyVoteStageUpdated), streamList)

	return http.StatusOK, nil
}

// global announcements
const HubKeyGlobalAnnouncementSubscribe hub.HubCommandKey = "GLOBAL_ANNOUNCEMENT:SUBSCRIBE"

func (s *StreamsWS) GlobalAnnouncementSubscribe(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc, needProcess bool) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}

	if needProcess {
		// get announcement
		ga, err := boiler.GlobalAnnouncements().One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", "", terror.Error(err, "failed to get announcement")
		}

		currentBattle, err := boiler.Battles(qm.OrderBy("battle_number DESC")).One(gamedb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", "", terror.Error(err, "failed to get current battle")
		}

		// show if battle number is equal or in between the global announcement's to and from battle number
		if currentBattle != nil && ga != nil && currentBattle.BattleNumber >= ga.ShowFromBattleNumber.Int && currentBattle.BattleNumber <= ga.ShowUntilBattleNumber.Int {
			reply(ga)
		} else {
			reply(nil)
		}
	}

	return req.TransactionID, messagebus.BusKey(HubKeyGlobalAnnouncementSubscribe), nil
}
