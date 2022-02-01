package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"server"
	"server/battle_arena"
	"server/passport"
	"strconv"
	"sync"
	"time"

	"github.com/ninja-syndicate/hub/ext/messagebus"
	"nhooyr.io/websocket"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/auth"
	zerologger "github.com/ninja-syndicate/hub/ext/zerolog"
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
	ctx    context.Context
	server *http.Server
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

	// map channels
	factionVoteCycle map[server.FactionID]chan func(*server.Faction, *VoteStage, FirstVoteState, *FirstVoteResult, *secondVoteResult, *FactionVotingTicker)
	liveVotingData   map[server.FactionID]chan func(*LiveVotingData)

	// client channels
	hubClientDetail map[*hub.Client]chan func(*HubClientDetail)
	onlineClientMap chan *ClientUpdate

	// battle queue channels
	battleQueueMap map[server.FactionID]chan func(*warMachineQueuingList)

	ringCheckAuthChan chan func(RingCheckAuthMap)
}

// NewAPI registers routes
func NewAPI(
	ctx context.Context,
	log *zerolog.Logger,
	battleArenaClient *battle_arena.BattleArena,
	pp *passport.Passport,
	addr string,
	HTMLSanitize *bluemonday.Policy,
	conn *pgxpool.Pool,
	config *server.Config,
) *API {

	// initialise message bus
	messageBus, offlineFunc := messagebus.NewMessageBus(log_helpers.NamedLogger(log, "message_bus"))
	// initialise api
	api := &API{
		ctx:          ctx,
		Log:          log_helpers.NamedLogger(log, "api"),
		Routes:       chi.NewRouter(),
		Passport:     pp,
		Addr:         addr,
		MessageBus:   messageBus,
		HTMLSanitize: HTMLSanitize,
		BattleArena:  battleArenaClient,
		Hub: hub.New(&hub.Config{
			Log: zerologger.New(*log_helpers.NamedLogger(log, "hub library")),
			WelcomeMsg: &hub.WelcomeMsg{
				Key:     "WELCOME",
				Payload: nil,
			},
			AcceptOptions: &websocket.AcceptOptions{
				InsecureSkipVerify: true, // TODO: set this depending on environment
				OriginPatterns:     []string{config.TwitchUIHostURL},
			},
			ClientOfflineFn: offlineFunc,
		}),
		// channel for faction voting system
		factionVoteCycle: make(map[server.FactionID]chan func(*server.Faction, *VoteStage, FirstVoteState, *FirstVoteResult, *secondVoteResult, *FactionVotingTicker)),
		liveVotingData:   make(map[server.FactionID]chan func(*LiveVotingData)),

		// channel for handling hub client
		hubClientDetail: make(map[*hub.Client]chan func(*HubClientDetail)),
		onlineClientMap: make(chan *ClientUpdate),
		// channel for battle queue
		battleQueueMap: make(map[server.FactionID]chan func(*warMachineQueuingList)),

		ringCheckAuthChan: make(chan func(RingCheckAuthMap)),
	}

	// start twitch jwt auth listener
	go api.startTwitchJWTAuthListener()

	api.Routes.Use(middleware.RequestID)
	api.Routes.Use(middleware.RealIP)
	api.Routes.Use(cors.New(cors.Options{AllowedOrigins: []string{config.TwitchUIHostURL}}).Handler)

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
	})
	///////////////////////////
	//		 Controllers	 //
	///////////////////////////
	_ = NewCheckController(log, conn, api)
	_ = NewTwitchController(log, conn, api)
	_ = NewUserController(log, conn, api)
	_ = NewAuthController(log, conn, api)

	///////////////////////////
	//		 Hub Events		 //
	///////////////////////////
	api.Hub.Events.AddEventHandler(hub.EventOnline, api.onlineEventHandler)
	api.Hub.Events.AddEventHandler(hub.EventOffline, api.offlineEventHandler)

	///////////////////////////
	//	Battle Arena Events	 //
	///////////////////////////
	api.BattleArena.Events.AddEventHandler(battle_arena.EventGameStart, api.BattleStartSignal)
	api.BattleArena.Events.AddEventHandler(battle_arena.EventGameEnd, api.BattleEndSignal)
	api.BattleArena.Events.AddEventHandler(battle_arena.EventWarMachinePositionChanged, api.UpdateWarMachinePosition)

	///////////////////////////
	//	 Passport Events	 //
	///////////////////////////
	// api.Passport.Events.AddEventHandler(passport.EventUserOnlineStatus, api.PassportUserOnlineStatusHandler)
	api.Passport.Events.AddEventHandler(passport.EventUserUpdated, api.PassportUserUpdatedHandler)
	api.Passport.Events.AddEventHandler(passport.EventUserEnlistFaction, api.PassportUserEnlistFactionHandler)
	api.Passport.Events.AddEventHandler(passport.EventBattleQueueJoin, api.PassportBattleQueueJoinHandler)
	api.Passport.Events.AddEventHandler(passport.EventBattleQueueLeave, api.PassportBattleQueueReleaseHandler)
	api.Passport.Events.AddEventHandler(passport.EventWarMachineQueuePositionGet, api.PassportWarMachineQueuePositionHandler)
	api.Passport.Events.AddEventHandler(passport.EventAuthRingCheck, api.AuthRingCheckHandler)

	// listen to the client online and action channel
	go api.ClientListener()

	go api.SetupAfterConnections()

	return api
}

func (api *API) SetupAfterConnections() {
	var factions []*server.Faction
	var err error

	// get factions from passport, retrying every 10 seconds until we ge them.
	for {
		// since the passport spins up concurrently the passport connection may not be setup right away, so we check every second for the connection
		for api.Passport == nil || api.Passport.Conn == nil || !api.Passport.Conn.Connected {
			time.Sleep(1 * time.Second)
		}

		factions, err = api.Passport.FactionAll(api.ctx, "faction all")
		if err != nil {
			api.Log.Err(err).Msg("unable to get factions")
		}

		if len(factions) > 0 {
			break
		}

		time.Sleep(5 * time.Second)
	}

	api.factionMap = make(map[server.FactionID]*server.Faction)
	for _, faction := range factions {
		api.factionMap[faction.ID] = faction
	}

	// get all the faction list from passport server
	for _, faction := range api.factionMap {
		// start live voting ticker
		api.liveVotingData[faction.ID] = make(chan func(*LiveVotingData))
		go api.startLiveVotingDataTicker(faction.ID)

		// start voting cycle
		api.factionVoteCycle[faction.ID] = make(chan func(*server.Faction, *VoteStage, FirstVoteState, *FirstVoteResult, *secondVoteResult, *FactionVotingTicker))
		go api.startFactionVoteCycle(faction)

		// start battle queue
		api.battleQueueMap[faction.ID] = make(chan func(*warMachineQueuingList))
		go api.startBattleQueue(faction.ID)
	}

	// start live voting broadcaster
	tickle.MinDurationOverride = true
	liveVotingBroadcasterLogger := log_helpers.NamedLogger(api.Log, "Live Voting Broadcaster").Level(zerolog.TraceLevel)
	liveVotingBroadcaster := tickle.New("Live Voting Broadcaster", 0.2, func() (int, error) {
		totalVote := server.BigInt{Int: *big.NewInt(0)}
		totalVoteMutex := sync.Mutex{}
		for _, faction := range factions {
			voteCountChan := make(chan server.BigInt)
			api.liveVotingData[faction.ID] <- func(lvd *LiveVotingData) {
				// pass value back
				voteCountChan <- lvd.TotalVote

				// clear current value
				lvd.TotalVote = server.BigInt{Int: *big.NewInt(0)}
			}

			voteCount := <-voteCountChan

			// protect total vote
			totalVoteMutex.Lock()
			totalVote.Add(&totalVote.Int, &voteCount.Int)
			totalVoteMutex.Unlock()
		}

		// broadcast notification to all the connected clients
		broadcastData, err := json.Marshal(&BroadcastPayload{
			Key:     HubKeyTwitchLiveVotingDataUpdated,
			Payload: totalVote,
		})
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err)
		}

		api.Hub.Clients(func(clients hub.ClientsList) {

			for client, ok := range clients {
				if !ok {
					continue
				}
				go func(c *hub.Client) {
					err := c.Send(broadcastData)
					if err != nil {
						api.Log.Err(err).Msg("failed to send broadcast")
					}
				}(client)
			}
		})

		return http.StatusOK, nil
	})
	liveVotingBroadcaster.Log = &liveVotingBroadcasterLogger
	// liveVotingBroadcaster.DisableLogging = true

	liveVotingBroadcaster.Start()
}

// Event handlers
func (api *API) onlineEventHandler(ctx context.Context, wsc *hub.Client, clients hub.ClientsList, ch hub.TriggerChan) {
	_, ok := api.hubClientDetail[wsc]
	if !ok {
		// initialise a client detail channel if not on the list
		api.hubClientDetail[wsc] = make(chan func(*HubClientDetail))
		go api.startClientTracker(wsc)
	}

}

func (api *API) offlineEventHandler(ctx context.Context, wsc *hub.Client, clients hub.ClientsList, ch hub.TriggerChan) {
	api.ClientOffline(wsc)
	// clean up the map
	delete(api.hubClientDetail, wsc)
}

// Run the API service
func (api *API) Run(ctx context.Context) error {
	api.server = &http.Server{
		Addr:    api.Addr,
		Handler: api.Routes,
	}

	api.Log.Info().Msgf("Starting API Server on %v", api.server.Addr)

	go func() {
		<-ctx.Done()
		api.Close()
	}()

	return api.server.ListenAndServe()
}

func (api *API) Close() {
	ctx := context.Background()
	api.Log.Info().Msg("Stopping API")
	err := api.server.Shutdown(ctx)
	if err != nil {
		api.Log.Warn().Err(err).Msg("")
	}
}

type GameSettingsResponse struct {
	GameMap     *server.GameMap         `json:"gameMap"`
	WarMachines []*server.WarMachineNFT `json:"warMachines"`
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
			go func(c *hub.Client) {
				err := c.Send(gameSettingsData)
				if err != nil {
					api.Log.Err(err).Msg("failed to send broadcast")
				}
			}(client)
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

	// release war machine
	for _, warMachine := range ed.BattleArena.WarMachines {
		warMachine.Durability = 100 * warMachine.RemainHitPoint / warMachine.MaxHitPoint
	}

	// release war machine
	if len(ed.BattleArena.WarMachines) > 0 {
		api.Passport.AssetRelease(
			context.Background(),
			"release_asset",
			ed.BattleArena.WarMachines,
		)
	}

	// start a new battle after 5 second
	go func() {

		for i := 5; i > 0; i-- {
			fmt.Println("Countdown ", strconv.Itoa(i), " second")
			time.Sleep(1 * time.Second)
		}

		fmt.Println("Init new game")

		// get NFT
		WarMachineList := []*server.WarMachineNFT{}
		for factionID := range api.battleQueueMap {
			WarMachineList = append(WarMachineList, api.GetBattleWarMachineFromQueue(factionID)...)
		}

		if len(WarMachineList) > 0 {
			tokenIDs := []uint64{}
			for _, warMachine := range WarMachineList {
				tokenIDs = append(tokenIDs, warMachine.TokenID)
			}

			// set war machine lock request
			err := api.Passport.AssetLock(ctx, "asset_lock", tokenIDs)
			if err != nil {
				api.Log.Err(err).Msg("Failed to lock assets")
				return
			}
		}

		// start another battle
		err := api.BattleArena.InitNextBattle(WarMachineList)
		if err != nil {
			api.Log.Err(err).Msg("Failed to initialise next battle")
			return
		}
	}()
}
