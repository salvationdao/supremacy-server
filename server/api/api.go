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
	"server/passport"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/auth"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
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

	FactionVotePriceMap map[server.FactionID]*FactionVotePrice
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

type BattleEndInfo struct {
	BattleID                     server.BattleID           `json:"battle_id"`
	StartedAt                    time.Time                 `json:"started_at"`
	EndedAt                      time.Time                 `json:"ended_at"`
	BattleIdentifier             int64                     `json:"battle_identifier"`
	WinningCondition             string                    `json:"winning_condition"`
	WinningFaction               *server.FactionBrief      `json:"winning_faction"`
	WinningWarMachines           []*server.WarMachineBrief `json:"winning_war_machines"`
	TopSupsContributeFactions    []*server.FactionBrief    `json:"top_sups_contribute_factions"`
	TopSupsContributors          []*server.UserBrief       `json:"top_sups_contributors"`
	MostFrequentAbilityExecutors []*server.UserBrief       `json:"most_frequent_ability_executors"`
}

// API server
type API struct {
	ctx    context.Context
	server *http.Server
	*auth.Auth
	Log           *zerolog.Logger
	Routes        chi.Router
	Addr          string
	BattleArena   *battle.Arena
	HTMLSanitize  *bluemonday.Policy
	Hub           *hub.Hub
	Conn          *pgxpool.Pool
	MessageBus    *messagebus.MessageBus
	NetMessageBus *messagebus.NetBus
	Passport      *passport.Passport

	// client detail
	UserMap *UserMap
	// ring check auth
	RingCheckAuthMap *RingCheckAuthMap
}

const SupremacyGameUserID = "4fae8fdf-584f-46bb-9cb9-bb32ae20177e"

// NewAPI registers routes
func NewAPI(
	ctx context.Context,
	log *zerolog.Logger,
	battleArenaClient *battle.Arena,
	pp *passport.Passport,
	addr string,
	HTMLSanitize *bluemonday.Policy,
	conn *pgxpool.Pool,
	config *server.Config,
	messageBus *messagebus.MessageBus,
	netMessageBus *messagebus.NetBus,
	gsHub *hub.Hub,
) *API {

	// initialise api
	api := &API{
		ctx:              ctx,
		Log:              log_helpers.NamedLogger(log, "api"),
		Routes:           chi.NewRouter(),
		Passport:         pp,
		Addr:             addr,
		MessageBus:       messageBus,
		NetMessageBus:    netMessageBus,
		HTMLSanitize:     HTMLSanitize,
		BattleArena:      battleArenaClient,
		Conn:             conn,
		Hub:              gsHub,
		RingCheckAuthMap: NewRingCheckMap(),
	}

	battleArenaClient.SetMessageBus(messageBus, netMessageBus)

	api.Routes.Use(middleware.RequestID)
	api.Routes.Use(middleware.RealIP)
	api.Routes.Use(middleware.Logger)
	api.Routes.Use(cors.New(cors.Options{AllowedOrigins: []string{"*"}}).Handler)

	api.Routes.Handle("/metrics", promhttp.Handler())
	api.Routes.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			sentryHandler := sentryhttp.New(sentryhttp.Options{})
			r.Use(sentryHandler.Handle)
		})
		r.Mount("/check", CheckRouter(log_helpers.NamedLogger(log, "check router"), conn))
		r.Mount(fmt.Sprintf("/%s/Supremacy_game", SupremacyGameUserID), PassportWebhookRouter(log, conn, config.PassportWebhookSecret, api))

		// Web sockets are long-lived, so we don't want the sentry performance tracer running for the life-time of the connection.
		// See roothub.ServeHTTP for the setup of sentry on this route.
		r.Handle("/ws", api.Hub)

		//TODO ALEX reimplement handlers

		//r.Get("/battlequeue", WithError(api.BattleArena.GetBattleQueue))
		//r.Get("/events", WithError(api.BattleArena.GetEvents))
		//r.Get("/faction_stats", WithError(api.BattleArena.FactionStats))
		//r.Get("/user_stats", WithError(api.BattleArena.UserStats))
		//r.Get("/abilities", WithError(api.BattleArena.GetAbility))

		r.Get("/blobs/{id}", WithError(api.IconDisplay))
		r.Post("/video_server", WithToken(config.ServerStreamKey, WithError((api.CreateStreamHandler))))
		r.Get("/video_server", WithToken(config.ServerStreamKey, WithError((api.GetStreamsHandler))))
		r.Delete("/video_server", WithToken(config.ServerStreamKey, WithError((api.DeleteStreamHandler))))
		r.Post("/close_stream", WithToken(config.ServerStreamKey, WithError(api.CreateStreamCloseHandler)))
		r.Get("/faction_data", WithError(api.GetFactionData))
		r.Get("/trigger/ability_file_upload", WithError(api.GetFactionData))
	})

	///////////////////////////
	//		 Controllers	 //
	///////////////////////////
	_ = NewCheckController(log, conn, api)
	_ = NewUserController(log, conn, api)
	_ = NewAuthController(log, conn, api)
	_ = NewVoteController(log, conn, api)
	// _ = NewFactionController(log, conn, api)
	_ = NewGameController(log, conn, api)
	_ = NewStreamController(log, conn, api)

	return api
}

// Run the API service
func (api *API) Run(ctx context.Context) error {
	api.server = &http.Server{
		Addr:    api.Addr,
		Handler: api.Routes,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	api.Log.Info().Msgf("Starting API Server on %v", api.server.Addr)

	go func() {
		<-ctx.Done()
		api.Close()
	}()

	return api.server.ListenAndServe()
}

func (api *API) Close() {
	ctx, cancel := context.WithTimeout(api.ctx, 5*time.Second)
	defer cancel()
	api.Log.Info().Msg("Stopping API")
	err := api.server.Shutdown(ctx)
	if err != nil {
		api.Log.Warn().Err(err).Msg("")
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
	err = db.FindBlob(context.Background(), api.Conn, &blob, blobID)
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
