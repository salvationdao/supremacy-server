package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"server"
	"server/battle"
	"server/db"
	"server/gamedb"
	"server/gamelog"
	"server/rpcclient"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/auth"
	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/pemistahl/lingua-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sasha-s/go-deadlock"
)

// WelcomePayload is the response sent when a client connects to the server
type WelcomePayload struct {
	Message string `json:"message"`
}

type BroadcastPayload struct {
	Key     hub.HubCommandKey `json:"key"`
	Payload interface{}       `json:"payload"`
}

type LiveVotingData struct {
	deadlock.Mutex
	TotalVote server.BigInt
}

type VotePriceSystem struct {
	VotePriceUpdater *tickle.Tickle

	GlobalVotePerTick []int64 // store last 100 tick total vote
	GlobalTotalVote   int64

	FactionVotePriceMap  map[server.FactionID]*FactionVotePrice
	FactionActivePlayers map[server.FactionID]*ActivePlayers
}

type FactionVotePrice struct {
	// priority lock
	OuterLock      deadlock.Mutex
	NextAccessLock deadlock.Mutex
	DataLock       deadlock.Mutex

	// price
	CurrentVotePriceSups server.BigInt
	CurrentVotePerTick   int64
}

// API server
type API struct {
	ctx    context.Context
	server *http.Server
	*auth.Auth
	Routes           chi.Router
	BattleArena      *battle.Arena
	HTMLSanitize     *bluemonday.Policy
	Hub              *hub.Hub
	MessageBus       *messagebus.MessageBus
	SMS              server.SMS
	Passport         *rpcclient.PassportXrpcClient
	Telegram         server.Telegram
	LanguageDetector lingua.LanguageDetector

	// ring check auth
	RingCheckAuthMap *RingCheckAuthMap

	// punish vote
	FactionPunishVote map[string]*PunishVoteTracker

	FactionActivePlayers map[string]*ActivePlayers

	// chatrooms
	GlobalChat      *Chatroom
	RedMountainChat *Chatroom
	BostonChat      *Chatroom
	ZaibatsuChat    *Chatroom

	Config *server.Config
}

// NewAPI registers routes
func NewAPI(
	ctx context.Context,
	battleArenaClient *battle.Arena,
	pp *rpcclient.PassportXrpcClient,
	HTMLSanitize *bluemonday.Policy,
	config *server.Config,
	messageBus *messagebus.MessageBus,
	gsHub *hub.Hub,
	sms server.SMS,
	telegram server.Telegram,
	languageDetector lingua.LanguageDetector,
) *API {

	// initialise api
	api := &API{
		Config:           config,
		ctx:              ctx,
		Routes:           chi.NewRouter(),
		MessageBus:       messageBus,
		HTMLSanitize:     HTMLSanitize,
		BattleArena:      battleArenaClient,
		Hub:              gsHub,
		RingCheckAuthMap: NewRingCheckMap(),
		Passport:         pp,
		SMS:              sms,
		Telegram:         telegram,
		LanguageDetector: languageDetector,

		FactionPunishVote:    make(map[string]*PunishVoteTracker),
		FactionActivePlayers: make(map[string]*ActivePlayers),

		// chatroom
		GlobalChat:      NewChatroom(nil),
		RedMountainChat: NewChatroom(&server.RedMountainFactionID),
		BostonChat:      NewChatroom(&server.BostonCyberneticsFactionID),
		ZaibatsuChat:    NewChatroom(&server.ZaibatsuFactionID),
	}

	battleArenaClient.SetMessageBus(messageBus)

	api.Routes.Use(middleware.RequestID)
	api.Routes.Use(middleware.RealIP)
	api.Routes.Use(gamelog.ChiLogger(zerolog.DebugLevel))
	api.Routes.Use(cors.New(cors.Options{AllowedOrigins: []string{"*"}}).Handler)
	api.Routes.Use(DatadogTracer.Middleware())

	api.Routes.Handle("/metrics", promhttp.Handler())
	api.Routes.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			sentryHandler := sentryhttp.New(sentryhttp.Options{})
			r.Use(sentryHandler.Handle)
		})
		r.Mount("/check", CheckRouter(battleArenaClient, telegram))
		r.Mount("/stat", AssetStatsRouter(api))
		r.Mount(fmt.Sprintf("/%s/Supremacy_game", server.SupremacyGameUserID), PassportWebhookRouter(config.PassportWebhookSecret, api))

		// Web sockets are long-lived, so we don't want the sentry performance tracer running for the life-time of the connection.
		// See roothub.ServeHTTP for the setup of sentry on this route.
		r.Handle("/ws", api.Hub)

		//TODO ALEX reimplement handlers

		r.Get("/blobs/{id}", WithError(api.IconDisplay))
		r.Post("/video_server", WithToken(config.ServerStreamKey, WithError(api.CreateStreamHandler)))
		r.Get("/video_server", WithToken(config.ServerStreamKey, WithError(api.GetStreamsHandler)))
		r.Delete("/video_server", WithToken(config.ServerStreamKey, WithError(api.DeleteStreamHandler)))
		r.Post("/close_stream", WithToken(config.ServerStreamKey, WithError(api.CreateStreamCloseHandler)))
		r.Get("/faction_data", WithError(api.GetFactionData))
		r.Get("/trigger/ability_file_upload", WithError(api.GetFactionData))

		r.Post("/global_announcement", WithToken(config.ServerStreamKey, WithError(api.GlobalAnnouncementSend)))
		r.Delete("/global_announcement", WithToken(config.ServerStreamKey, WithError(api.GlobalAnnouncementDelete)))
	})

	///////////////////////////
	//		 Controllers	 //
	///////////////////////////
	_ = NewCheckController(api)
	_ = NewUserController(api)
	_ = NewAuthController(api)
	_ = NewGameController(api)
	_ = NewStreamController(api)
	_ = NewPlayerController(api)
	_ = NewChatController(api)
	_ = NewBattleController(api)
	_ = NewPlayerAbilitiesController(api)

	// create a tickle that update faction mvp every day 00:00 am
	factionMvpUpdate := tickle.New("Calculate faction mvp player", 24*60*60, func() (int, error) {
		// set red mountain mvp player
		gamelog.L.Info().Str("faction_id", server.RedMountainFactionID.String()).Msg("Recalculate Red Mountain mvp player")
		err := db.FactionStatMVPUpdate(server.RedMountainFactionID.String())
		if err != nil {
			gamelog.L.Error().Str("faction_id", server.RedMountainFactionID.String()).Err(err).Msg("Failed to recalculate Red Mountain mvp player")
		}

		// set boston mvp player
		gamelog.L.Info().Str("faction_id", server.BostonCyberneticsFactionID.String()).Msg("Recalculate Boston mvp player")
		err = db.FactionStatMVPUpdate(server.BostonCyberneticsFactionID.String())
		if err != nil {
			gamelog.L.Error().Str("faction_id", server.BostonCyberneticsFactionID.String()).Err(err).Msg("Failed to recalculate Boston mvp player")
		}

		// set Zaibatsu mvp player
		gamelog.L.Info().Str("faction_id", server.ZaibatsuFactionID.String()).Msg("Recalculate Zaibatsu mvp player")
		err = db.FactionStatMVPUpdate(server.ZaibatsuFactionID.String())
		if err != nil {
			gamelog.L.Error().Str("faction_id", server.ZaibatsuFactionID.String()).Err(err).Msg("Failed to recalculate Zaibatsu mvp player")
		}

		return http.StatusOK, nil
	})
	factionMvpUpdate.Log = gamelog.L

	err := factionMvpUpdate.SetIntervalAt(24*time.Hour, 0, 0)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to set up faction mvp user update tickle")
	}

	// spin up a punish vote handlers for each faction
	err = api.PunishVoteTrackerSetup()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to setup punish vote tracker")
	}

	api.FactionActivePlayerSetup()

	return api
}

// Run the API service
func (api *API) Run(ctx context.Context) error {
	api.server = &http.Server{
		Addr:    api.Config.Address,
		Handler: api.Routes,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	gamelog.L.Info().Msgf("Starting API Server on %v", api.server.Addr)

	go func() {
		<-ctx.Done()
		api.Close()
	}()

	return api.server.ListenAndServe()
}

func (api *API) Close() {
	ctx, cancel := context.WithTimeout(api.ctx, 5*time.Second)
	defer cancel()
	gamelog.L.Info().Msg("Stopping API")
	err := api.server.Shutdown(ctx)
	if err != nil {
		gamelog.L.Warn().Err(err).Msg("")
	}
}

func (api *API) IconDisplay(w http.ResponseWriter, r *http.Request) (int, error) {
	defer r.Body.Close()

	// Get blob id
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return http.StatusBadRequest, terror.Error(terror.ErrInvalidInput, "no id provided")
	}
	id, err := uuid.FromString(idStr)
	blobID := server.BlobID(id)
	if err != nil {
		return http.StatusBadRequest, terror.Error(terror.ErrInvalidInput, "invalid id provided")
	}

	var blob server.Blob

	// Get blob
	err = db.FindBlob(context.Background(), gamedb.Conn, &blob, blobID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return http.StatusNotFound, terror.Error(err, "attachment not found")
		}
		return http.StatusInternalServerError, terror.Error(err, "could not get attachment")
	}

	// Get disposition
	disposition := "attachment"
	isViewDisposition := r.URL.Query().Get("view")
	if isViewDisposition == "true" {
		disposition = "inline"
	}

	// tell the browser the returned content should be downloaded/inline
	if blob.MimeType != "" && blob.MimeType != "unknown" {
		w.Header().Add("Content-Type", blob.MimeType)
	}
	w.Header().Add("Content-Disposition", fmt.Sprintf("%s;filename=%s", disposition, blob.FileName))
	_, err = w.Write(blob.File)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

/**********************
* Auth Ring Check Map *
**********************/

type RingCheckAuthMap struct {
	deadlock.Map
}

func NewRingCheckMap() *RingCheckAuthMap {
	return &RingCheckAuthMap{
		deadlock.Map{},
	}
}

func (rcm *RingCheckAuthMap) Record(key string, cl *hub.Client) {
	rcm.Store(key, cl)
}

func (rcm *RingCheckAuthMap) Remove(key string) {
	rcm.Delete(key)
}

func (rcm *RingCheckAuthMap) Check(key string) (*hub.Client, error) {
	value, ok := rcm.Load(key)
	if !ok {
		gamelog.L.Error().Str("key", key).Msg("hub client not found")
		return nil, terror.Error(fmt.Errorf("hub client not found"))
	}

	hubc, ok := value.(*hub.Client)
	if !ok {
		return nil, terror.Error(fmt.Errorf("hub client not found"))
	}

	return hubc, nil
}
