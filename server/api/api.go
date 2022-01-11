package api

import (
	"context"
	"encoding/json"
	"net/http"
	"server"
	"server/battle_arena"
	"server/passport"

	"github.com/ninja-software/hub/v2/ext/messagebus"
	"nhooyr.io/websocket"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/auth"
	zerologger "github.com/ninja-software/hub/v2/ext/zerolog"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/tickle"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// WelcomePayload is the response sent when a client connects to the server
type WelcomePayload struct {
	Message string `json:"message"`
}

type BroadcastPayload struct {
	Key     hub.HubCommandKey `json:"key"`
	Payload interface{}       `json:"payload"`
}

// API server
type API struct {
	*auth.Auth
	Log          *zerolog.Logger
	Routes       chi.Router
	Addr         string
	BattleArena  *battle_arena.BattleArena
	HTMLSanitize *bluemonday.Policy
	Hub          *hub.Hub
	MessageBus   *messagebus.MessageBus
	Passport     *passport.Passport

	factionMap map[server.FactionID]*server.Faction

	// map routines
	factionVoteCycle map[server.FactionID]chan func(*server.Faction, *VoteStage, FirstVoteState, *FirstVoteResult, *secondVoteResult, *FactionVotingTicker)

	hubClientDetail map[*hub.Client]chan func(*HubClientDetail)
	onlineClientMap map[server.UserID]chan func(ClientInstanceMap, *SupremacyTokenState, *tickle.Tickle)
}

// NewAPI registers routes
func NewAPI(
	log *zerolog.Logger,
	battleArenaClient *battle_arena.BattleArena,
	pp *passport.Passport,
	factionMap map[server.FactionID]*server.Faction,
	cancelOnPanic context.CancelFunc,
	addr string,
	HTMLSanitize *bluemonday.Policy,
	conn *pgxpool.Pool,
	config *server.Config,
) *API {
	// initialise message bus
	messageBus, offlineFunc := messagebus.NewMessageBus(log_helpers.NamedLogger(log, "message_bus"))

	// initialise api
	api := &API{
		Log:          log_helpers.NamedLogger(log, "api"),
		Routes:       chi.NewRouter(),
		Passport:     pp,
		Addr:         addr,
		MessageBus:   messageBus,
		HTMLSanitize: HTMLSanitize,
		BattleArena:  battleArenaClient,
		factionMap:   factionMap,
		Hub: hub.New(&hub.Config{
			Log: zerologger.New(*log_helpers.NamedLogger(log, "hub library")),
			WelcomeMsg: &hub.WelcomeMsg{
				Key:     "WELCOME",
				Payload: nil,
			},
			AcceptOptions: &websocket.AcceptOptions{
				InsecureSkipVerify: true,
			},
			ClientOfflineFn: offlineFunc,
		}),

		// channel for faction voting system
		factionVoteCycle: make(map[server.FactionID]chan func(*server.Faction, *VoteStage, FirstVoteState, *FirstVoteResult, *secondVoteResult, *FactionVotingTicker)),

		// channel for handling hub client
		hubClientDetail: make(map[*hub.Client]chan func(*HubClientDetail)),
		onlineClientMap: make(map[server.UserID](chan func(ClientInstanceMap, *SupremacyTokenState, *tickle.Tickle))),
	}

	// start the default online client map
	defaultHubClientUUID := server.UserID(uuid.Nil)
	api.onlineClientMap[defaultHubClientUUID] = make(chan func(ClientInstanceMap, *SupremacyTokenState, *tickle.Tickle))
	go api.startOnlineClientTracker(defaultHubClientUUID, 0)

	// get all the faction list from passport server and create channel
	for _, faction := range factionMap {
		api.factionVoteCycle[faction.ID] = make(chan func(*server.Faction, *VoteStage, FirstVoteState, *FirstVoteResult, *secondVoteResult, *FactionVotingTicker))
		go api.startFactionVoteCycle(faction)
	}

	// add online/offline event handlers
	api.Hub.Events.AddEventHandler(hub.EventOnline, api.onlineEventHandler)
	api.Hub.Events.AddEventHandler(hub.EventOffline, api.offlineEventHandler)

	api.Routes.Use(middleware.RequestID)
	api.Routes.Use(middleware.RealIP)
	api.Routes.Use(cors.New(cors.Options{}).Handler)

	// Commented out by vinnie 22/12/21 -- Looks like we don't need the auth extension atm since using a different flow

	//var err error
	//api.Auth, err = auth.New(api.Hub, &auth.Config{
	//	CookieSecure: config.CookieSecure,
	//	UserGetter: &UserGetter{
	//		Log:  log_helpers.NamedLogger(log, "user getter"),
	//		Conn: conn,
	//	},
	//	Tokens: &Tokens{
	//		Conn:                conn,
	//		tokenExpirationDays: config.TokenExpirationDays,
	//		encryptToken:        config.EncryptTokens,
	//		encryptTokenKey:     config.EncryptTokensKey,
	//	},
	//	//Whitelist           bool
	//})
	//if err != nil {
	//	log.Fatal().Msgf("failed to init hub auther: %s", err.Error())
	//}

	api.Routes.Handle("/metrics", promhttp.Handler())
	api.Routes.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			sentryHandler := sentryhttp.New(sentryhttp.Options{})
			r.Use(sentryHandler.Handle)
		})
		// Web sockets are long-lived, so we don't want the sentry performance tracer running for the life-time of the connection.
		// See roothub.ServeHTTP for the setup of sentry on this route.
		r.Handle("/ws", api.Hub)
		r.Get("/game_settings", WithError(api.GetGameSettings))
		r.Get("/second_votes", WithError(api.GetSecondVotes))
		// r.HandleFunc("/temp-random-faction", api.GetRandomFaction)
		r.HandleFunc("/start", api.Start) // TODO: will be removed at a later date
	})

	_ = NewCheckController(log, conn, api)
	_ = NewTwitchController(log, conn, api)
	_ = NewUserController(log, conn, api)

	///////////////////////////
	//		Battle Arena	 //
	///////////////////////////

	api.BattleArena.Events.AddEventHandler(battle_arena.EventGameStart, api.BattleStartSignal)
	api.BattleArena.Events.AddEventHandler(battle_arena.EventGameEnd, api.BattleEndSignal)
	api.BattleArena.Events.AddEventHandler(battle_arena.EventWarMachinePositionChanged, api.UpdateWarMachinePosition)

	///////////////////////////
	//			Passport	 //
	///////////////////////////

	api.Passport.Events.AddEventHandler(passport.EventUserOnlineStatus, api.PassportUserOnlineStatusHandler)
	api.Passport.Events.AddEventHandler(passport.EventUserUpdated, api.PassportUserUpdatedHandler)

	return api
}

// Event handlers
func (api *API) onlineEventHandler(ctx context.Context, wsc *hub.Client, clients hub.ClientsList, ch hub.TriggerChan) {
	_, ok := api.hubClientDetail[wsc]
	if !ok {
		// initialise a client detail channel if not on the list
		api.hubClientDetail[wsc] = make(chan func(*HubClientDetail))
		go api.startClientTracker(wsc)
	}

	api.onlineClientMap[server.UserID(uuid.Nil)] <- func(cim ClientInstanceMap, cps *SupremacyTokenState, t *tickle.Tickle) {
		// register the client instance if not exists
		if _, ok := cim[wsc]; !ok {
			cim[wsc] = true
		}
	}
}

func (api *API) offlineEventHandler(ctx context.Context, wsc *hub.Client, clients hub.ClientsList, ch hub.TriggerChan) {
	// remove client detail from
	_, ok := api.hubClientDetail[wsc]
	if !ok {
		return
	}

	hubClientDetail, err := api.getClientDetailFromChannel(wsc)
	if err != nil {
		api.Log.Err(err).Msg("User not found")
		return
	}

	// clean up the map
	delete(api.hubClientDetail, wsc)

	shouldDeleteChan := make(chan bool)
	// delete the online instance from the map
	api.onlineClientMap[hubClientDetail.ID] <- func(cim ClientInstanceMap, cps *SupremacyTokenState, t *tickle.Tickle) {
		delete(cim, wsc)

		if len(cim) == 0 && !hubClientDetail.ID.IsNil() {
			// stop channel point tickle
			if t.NextTick != nil {
				t.Stop()
			}
			// TODO: store current channel point back to passport
			api.Log.Info().Msgf("Store the connect point of user %s back to passport", hubClientDetail.ID)

			shouldDeleteChan <- true
			return
		}
		shouldDeleteChan <- false
	}

	// delete map instance if required
	if <-shouldDeleteChan {

		delete(api.onlineClientMap, hubClientDetail.ID)
	}
}

// Run the API service
func (api *API) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    api.Addr,
		Handler: api.Routes,
	}

	api.Log.Info().Msgf("Starting API Server on %v", server.Addr)

	go func() {
		select {
		case <-ctx.Done():
			api.Log.Info().Msg("Stopping API")
			err := server.Shutdown(ctx)
			if err != nil {
				api.Log.Warn().Err(err).Msg("")
			}
		}
	}()

	return server.ListenAndServe()
}

type GameSettingsResponse struct {
	GameMap     *server.GameMap      `json:"gameMap"`
	WarMachines []*server.WarMachine `json:"warMachines"`
}

const HubKeyGameSettingsUpdated hub.HubCommandKey = hub.HubCommandKey("GAME:SETTINGS:UPDATED")

// BattleStartSignal start all the voting cycle
func (api *API) BattleStartSignal(ctx context.Context, ed *battle_arena.EventData) {

	// marshal payload
	gameSettingsData, err := json.Marshal(&BroadcastPayload{
		Key: HubKeyGameSettingsUpdated,
		Payload: &GameSettingsResponse{
			GameMap:     ed.BattleArena.GameMap,
			WarMachines: ed.BattleArena.WarMachines,
		},
	})
	if err != nil {
		return
	}

	// broadcast game settings to all the connected clients
	api.Hub.Clients(func(clients hub.ClientsList) {
		for client, ok := range clients {
			if !ok {
				continue
			}
			go client.Send(gameSettingsData)
		}
	})

	// start battle voting phase for all the factions
	for factionID := range api.factionVoteCycle {
		go api.startVotingCycle(factionID)
	}
}

// BattleEndSignal terminate all the voting cycle
func (api *API) BattleEndSignal(ctx context.Context, ed *battle_arena.EventData) {
	// start battle voting phase for all the factions
	for factionID := range api.factionVoteCycle {
		go api.pauseVotingCycle(factionID)
	}
}

// // GetRandomFaction just a dummy end point to give a random faction to a user
// func (api *API) GetRandomFaction(w http.ResponseWriter, r *http.Request) {
// 	randomFaction := passport.RandomFaction(api.factions)

// 	code := r.URL.Query().Get("twitchID")
// 	user := api.Passport.FakeUserLoginWithoutFaction(code)
// 	// This will normally be saved on passport
// 	user.Faction = randomFaction

// }

// Start starts the battle flow
func (api *API) Start(w http.ResponseWriter, r *http.Request) {
	go func() {
		err := api.BattleArena.InitNextBattle()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}()

	w.WriteHeader(http.StatusOK)
}
