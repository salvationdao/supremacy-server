package api

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"server"
	"server/battle_arena"
	"server/db"
	"server/passport"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"nhooyr.io/websocket"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/log_helpers"
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

type VotePriceSystem struct {
	VotePriceUpdater    *tickle.Tickle
	VotePriceForecaster *tickle.Tickle

	GlobalVotePerTick []int64 // store last 100 tick total vote
	GlobalTotalVote   int64

	FactionVotePriceMap map[server.FactionID]*FactionVotePrice
}

type FactionVotePrice struct {
	// priority lock
	OuterLock      sync.Mutex
	NextAccessLock sync.Mutex
	DataLock       sync.Mutex

	// price
	CurrentVotePriceSups server.BigInt
	CurrentVotePerTick   int64
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

	// voting channels
	liveSupsSpend map[server.FactionID]chan func(*LiveVotingData)

	// client channels
	hubClientDetail map[*hub.Client]chan func(*HubClientDetail)
	onlineClientMap chan *ClientUpdate

	// ring check auth
	ringCheckAuthChan chan func(RingCheckAuthMap)

	// voting channels
	votePhaseChecker *VotePhaseChecker
	votingCycle      chan func(*VoteStage, *VoteAbility, FactionUserVoteMap, *FactionTotalVote, *VoteWinner, *VotingCycleTicker)
	votePriceSystem  *VotePriceSystem

	// faction abilities
	factionAbilityPool map[server.FactionID]chan func(FactionAbilitiesPool, *FactionAbilityPoolTicker)
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
		// factionVoteCycle: make(map[server.FactionID]chan func(*server.Faction, *VoteStage, FirstVoteState, *FirstVoteResult, *secondVoteResult, *FactionVotingTicker)),
		liveSupsSpend: make(map[server.FactionID]chan func(*LiveVotingData)),

		votingCycle: make(chan func(*VoteStage, *VoteAbility, FactionUserVoteMap, *FactionTotalVote, *VoteWinner, *VotingCycleTicker)),

		// channel for handling hub client
		hubClientDetail: make(map[*hub.Client]chan func(*HubClientDetail)),
		onlineClientMap: make(chan *ClientUpdate),

		// ring check auth
		ringCheckAuthChan: make(chan func(RingCheckAuthMap)),

		factionAbilityPool: make(map[server.FactionID]chan func(FactionAbilitiesPool, *FactionAbilityPoolTicker)),
	}

	// start twitch jwt auth listener
	go api.startAuthRignCheckListener()

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
		r.Get("/events", WithError(api.BattleArena.GetEvents))
	})

	///////////////////////////
	//		 Controllers	 //
	///////////////////////////
	_ = NewCheckController(log, conn, api)
	_ = NewUserController(log, conn, api)
	_ = NewAuthController(log, conn, api)
	_ = NewVoteController(log, conn, api)
	_ = NewFactionController(log, conn, api)

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

	go api.SetupAfterConnections(conn)

	return api
}

func (api *API) SetupAfterConnections(conn *pgxpool.Pool) {
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

	// build faction map for main server
	api.factionMap = make(map[server.FactionID]*server.Faction)
	for _, faction := range factions {
		err := db.FactionVotePriceGet(context.Background(), conn, faction)
		if err != nil {
			api.Log.Err(err).Msg("unable to get faction vote price")
		}

		api.factionMap[faction.ID] = faction

		// start live voting ticker
		api.liveSupsSpend[faction.ID] = make(chan func(*LiveVotingData))
		go api.startLiveVotingDataTicker(faction.ID)

		// faction ability pool
		api.factionAbilityPool[faction.ID] = make(chan func(FactionAbilitiesPool, *FactionAbilityPoolTicker))
		go api.StartFactionAbilityPool(faction.ID, conn)
	}

	// initialise vote price system
	go api.startVotePriceSystem(factions, conn)

	// initialise voting cycle
	go api.StartVotingCycle(factions)

	// set faction map for battle arena server
	api.BattleArena.SetFactionMap(api.factionMap)

	// declare live sups spend broadcaster
	tickle.MinDurationOverride = true
	liveVotingBroadcasterLogger := log_helpers.NamedLogger(api.Log, "Live Sups spend Broadcaster").Level(zerolog.Disabled)
	liveVotingBroadcaster := tickle.New("Live Sups spend Broadcaster", 0.2, func() (int, error) {
		totalVote := server.BigInt{Int: *big.NewInt(0)}
		totalVoteMutex := sync.Mutex{}
		for _, faction := range factions {
			voteCountChan := make(chan server.BigInt)
			api.liveSupsSpend[faction.ID] <- func(lvd *LiveVotingData) {
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

		// prepare payload
		payload := []byte{}
		payload = append(payload, byte(battle_arena.NetMessageTypeLiveVotingTick))
		payload = append(payload, []byte(totalVote.Int.String())...)

		api.Hub.Clients(func(clients hub.ClientsList) {
			for client, ok := range clients {
				if !ok {
					continue
				}
				go func(c *hub.Client) {
					err := c.SendWithMessageType(payload, websocket.MessageBinary)
					if err != nil {
						api.Log.Err(err).Msg("failed to send broadcast")
					}
				}(client)
			}
		})

		return http.StatusOK, nil
	})
	liveVotingBroadcaster.Log = &liveVotingBroadcasterLogger

	// start live voting broadcaster
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

	ba := api.BattleArena.GetCurrentState()

	go func() {
		// delay 2 second to wait frontend setup key map
		time.Sleep(2 * time.Second)

		// marshal payload
		gameSettingsData, err := json.Marshal(&BroadcastPayload{
			Key: HubKeyGameSettingsUpdated,
			Payload: &GameSettingsResponse{
				GameMap:     ba.GameMap,
				WarMachines: ba.WarMachines,
				//WarMachineLocation: ba.BattleHistory[0],
			},
		})
		if err != nil {
			api.Log.Err(err).Msg("failed to marshal data")
			return
		}

		err = wsc.Send(gameSettingsData)
		if err != nil {
			api.Log.Err(err).Msg("failed to send broadcast")
		}
	}()

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
	// WarMachineLocation []byte                  `json:"warMachineLocation"`
}

const HubKeyGameSettingsUpdated = hub.HubCommandKey("GAME:SETTINGS:UPDATED")

// BattleStartSignal start all the voting cycle
func (api *API) BattleStartSignal(ctx context.Context, ed *battle_arena.EventData) {
	// build faction detail to battle start
	warMachines := ed.BattleArena.WarMachines
	for _, wm := range warMachines {
		wm.Faction = ed.BattleArena.FactionMap[wm.FactionID]
	}

	// marshal payload
	gameSettingsData, err := json.Marshal(&BroadcastPayload{
		Key: HubKeyGameSettingsUpdated,
		Payload: &GameSettingsResponse{
			GameMap:     ed.BattleArena.GameMap,
			WarMachines: ed.BattleArena.WarMachines,
			// WarMachineLocation: ed.BattleArena.BattleHistory[0],
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

	// start voting cycle, initial intro time equal: (mech_count * 3 + 7) seconds
	introSecond := len(warMachines)*3 + 7

	go api.startVotingCycle(introSecond)

	for factionID := range api.factionMap {
		// get initial abilities
		initialAbilities, err := db.FactionExclusiveAbilitiesByFactionID(api.ctx, api.BattleArena.Conn, factionID)
		if err != nil {
			api.Log.Err(err).Msg("Failed to query initial faction abilities")
			return
		}
		for _, ab := range initialAbilities {
			ab.Title = "FACTION_WIDE"
			ab.CurrentSups = "0"
		}

		for _, wm := range ed.BattleArena.WarMachines {
			if wm.FactionID != factionID || len(wm.Abilities) == 0 {
				continue
			}

			for _, ability := range wm.Abilities {
				initialAbilities = append(initialAbilities, &server.FactionAbility{
					ID:                  server.FactionAbilityID(uuid.Must(uuid.NewV4())), // generate a uuid for frontend to track sups contribution
					GameClientAbilityID: byte(ability.GameClientID),
					ImageUrl:            ability.Image,
					FactionID:           factionID,
					Label:               ability.Name,
					SupsCost:            ability.SupsCost,
					CurrentSups:         "0",
					AbilityTokenID:      ability.TokenID,
					WarMachineTokenID:   wm.TokenID,
					ParticipantID:       &wm.ParticipantID,
					Title:               wm.Name,
				})
			}
		}

		go api.startFactionAbilityPoolTicker(factionID, initialAbilities, introSecond)
	}

}

// BattleEndSignal terminate all the voting cycle
func (api *API) BattleEndSignal(ctx context.Context, ed *battle_arena.EventData) {
	// stop all the tickles in voting cycle
	go api.stopVotingCycle()
	go api.stopFactionAbilityPoolTicker()

	// parse battle reward list
	api.Hub.Clients(func(clients hub.ClientsList) {
		for c := range clients {
			go func(c *hub.Client) {
				userID := server.UserID(uuid.FromStringOrNil(c.Identifier()))
				if userID.IsNil() {
					return
				}
				hcd, err := api.getClientDetailFromChannel(c)
				if err != nil || hcd.FactionID.IsNil() {
					return
				}

				brs := []BattleRewardType{}
				// check reward
				if hcd.FactionID == ed.BattleRewardList.WinnerFactionID {
					brs = append(brs, BattleRewardTypeFaction)
				}

				if _, ok := ed.BattleRewardList.WinningWarMachineOwnerIDs[userID]; ok {
					brs = append(brs, BattleRewardTypeWinner)
				}

				if _, ok := ed.BattleRewardList.ExecuteKillWarMachineOwnerIDs[userID]; ok {
					brs = append(brs, BattleRewardTypeKill)
				}

				if len(brs) == 0 {
					return
				}

				api.ClientBattleRewardUpdate(c, &ClientBattleReward{
					BattleID: ed.BattleRewardList.BattleID,
					Rewards:  brs,
				})
			}(c)
		}
	})
}
