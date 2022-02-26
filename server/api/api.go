package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"server"
	"server/battle_arena"
	"server/db"
	"server/passport"
	"sync"
	"time"

	"github.com/jpillora/backoff"

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

type LiveVotingData struct {
	sync.Mutex
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
	OuterLock      sync.Mutex
	NextAccessLock sync.Mutex
	DataLock       sync.Mutex

	// price
	CurrentVotePriceSups server.BigInt
	CurrentVotePerTick   int64
}

type BattleEndInfo struct {
	BattleID                     server.BattleID           `json:"battleID"`
	StartedAt                    time.Time                 `json:"startedAt"`
	EndedAt                      time.Time                 `json:"endedAt"`
	BattleIdentifier             int64                     `json:"battleIdentifier"`
	WinningCondition             string                    `json:"winningCondition"`
	WinningFaction               *server.FactionBrief      `json:"winningFaction"`
	WinningWarMachines           []*server.WarMachineBrief `json:"winningWarMachines"`
	TopSupsContributeFactions    []*server.FactionBrief    `json:"topSupsContributeFactions"`
	TopSupsContributors          []*server.UserBrief       `json:"topSupsContributors"`
	MostFrequentAbilityExecutors []*server.UserBrief       `json:"mostFrequentAbilityExecutors"`
}

// API server
type API struct {
	ctx    context.Context
	server *http.Server
	*auth.Auth
	Log           *zerolog.Logger
	Routes        chi.Router
	Addr          string
	BattleArena   *battle_arena.BattleArena
	HTMLSanitize  *bluemonday.Policy
	Hub           *hub.Hub
	Conn          *pgxpool.Pool
	MessageBus    *messagebus.MessageBus
	NetMessageBus *messagebus.NetBus
	Passport      *passport.Passport

	factionMap map[server.FactionID]*server.Faction

	// voting channels
	liveSupsSpend map[server.FactionID]*LiveVotingData

	// client channels
	// onlineClientMap chan *ClientUpdate

	UserMultiplier *UserMultiplier
	// client detail
	UserMap *UserMap
	// ring check auth
	RingCheckAuthMap *RingCheckAuthMap

	// voting channels
	votePhaseChecker *VotePhaseChecker
	votingCycle      chan func(*VoteAbility, FactionUserVoteMap, *FactionTransactions, *FactionTotalVote, *VoteWinner, *VotingCycleTicker, UserVoteMap)
	votePriceSystem  *VotePriceSystem

	// faction abilities
	gameAbilityPool map[server.FactionID]func(func(*sync.Map))

	// viewer live count
	viewerLiveCount *ViewerLiveCount

	battleEndInfo *BattleEndInfo
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

	netMessageBus := messagebus.NewNetBus(log_helpers.NamedLogger(log, "net_message_bus"))

	// initialise message bus
	messageBus := messagebus.NewMessageBus(log_helpers.NamedLogger(log, "message_bus"))
	// initialise api
	api := &API{
		ctx:           ctx,
		Log:           log_helpers.NamedLogger(log, "api"),
		Routes:        chi.NewRouter(),
		Passport:      pp,
		Addr:          addr,
		MessageBus:    messageBus,
		NetMessageBus: netMessageBus,
		HTMLSanitize:  HTMLSanitize,
		BattleArena:   battleArenaClient,
		Conn:          conn,
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
			ClientOfflineFn: func(cl *hub.Client) {
				netMessageBus.UnsubAll(cl)
				messageBus.UnsubAll(cl)
			},
		}),
		// channel for faction voting system
		votingCycle:   make(chan func(*VoteAbility, FactionUserVoteMap, *FactionTransactions, *FactionTotalVote, *VoteWinner, *VotingCycleTicker, UserVoteMap)),
		liveSupsSpend: make(map[server.FactionID]*LiveVotingData),

		// channel for handling hub client
		// onlineClientMap: make(chan *ClientUpdate),

		// ring check auth
		RingCheckAuthMap: NewRingCheckMap(),

		// game ability pool
		gameAbilityPool: make(map[server.FactionID]func(func(*sync.Map))),

		// faction viewer count
		battleEndInfo: &BattleEndInfo{},
	}

	api.Routes.Use(middleware.RequestID)
	api.Routes.Use(middleware.RealIP)
	api.Routes.Use(cors.New(cors.Options{AllowedOrigins: []string{config.TwitchUIHostURL}}).Handler)

	api.Routes.Handle("/metrics", promhttp.Handler())
	api.Routes.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			sentryHandler := sentryhttp.New(sentryhttp.Options{})
			r.Use(sentryHandler.Handle)
		})
		r.Mount("/check", CheckRouter(log_helpers.NamedLogger(log, "check router"), conn))

		// Web sockets are long-lived, so we don't want the sentry performance tracer running for the life-time of the connection.
		// See roothub.ServeHTTP for the setup of sentry on this route.
		r.Handle("/ws", api.Hub)
		r.Get("/battlequeue", WithError(api.BattleArena.GetBattleQueue))
		r.Get("/events", WithError(api.BattleArena.GetEvents))
		r.Get("/faction_stats", WithError(api.BattleArena.FactionStats))
		r.Get("/user_stats", WithError(api.BattleArena.UserStats))
		r.Get("/abilities", WithError(api.BattleArena.GetAbility))
		r.Get("/blobs/{id}", WithError(api.BattleArena.GetBlob))
		r.Post("/video_server", WithToken(config.ServerStreamKey, WithError((api.CreateStreamHandler))))
		r.Get("/video_server", WithToken(config.ServerStreamKey, WithError((api.GetStreamsHandler))))
		r.Delete("/video_server", WithToken(config.ServerStreamKey, WithError((api.DeleteStreamHandler))))
		r.Get("/faction_data", WithError(api.GetFactionData))
		r.Get("/trigger/ability_file_upload", WithError(api.GetFactionData))
	})

	// set viewer live count
	api.viewerLiveCount = NewViewerLiveCount(api.NetMessageBus)
	api.UserMap = NewUserMap(api.viewerLiveCount)
	api.UserMultiplier = NewUserMultiplier(api.UserMap, api.Passport, api.BattleArena)

	///////////////////////////
	//		 Controllers	 //
	///////////////////////////
	_ = NewCheckController(log, conn, api)
	_ = NewUserController(log, conn, api)
	_ = NewAuthController(log, conn, api)
	_ = NewVoteController(log, conn, api)
	_ = NewFactionController(log, conn, api)
	_ = NewGameController(log, conn, api)
	_ = NewStreamController(log, conn, api)

	///////////////////////////
	//		 Hub Events		 //
	///////////////////////////
	api.Hub.Events.AddEventHandler(hub.EventOnline, api.onlineEventHandler, func(e error) {})
	api.Hub.Events.AddEventHandler(hub.EventOffline, api.offlineEventHandler, func(e error) {})

	///////////////////////////
	//	Battle Arena Events	 //
	///////////////////////////
	api.BattleArena.Events.AddEventHandler(battle_arena.EventGameInit, api.BattleInitSignal)
	api.BattleArena.Events.AddEventHandler(battle_arena.EventGameStart, api.BattleStartSignal)
	api.BattleArena.Events.AddEventHandler(battle_arena.EventGameEnd, api.BattleEndSignal)
	api.BattleArena.Events.AddEventHandler(battle_arena.EventWarMachineDestroyed, api.WarMachineDestroyedBroadcast)
	api.BattleArena.Events.AddEventHandler(battle_arena.EventWarMachinePositionChanged, api.UpdateWarMachinePosition)
	api.BattleArena.Events.AddEventHandler(battle_arena.EventWarMachineQueueUpdated, api.UpdateWarMachineQueue)

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
	api.Passport.Events.AddEventHandler(passport.EventAssetInsurancePay, api.PassportAssetInsurancePayHandler)
	api.Passport.Events.AddEventHandler(passport.EventFactionStatGet, api.PassportFactionStatGetHandler)
	api.Passport.Events.AddEventHandler(passport.EventUserSupsMultiplierGet, api.PassportUserSupsMultiplierGetHandler)
	api.Passport.Events.AddEventHandler(passport.EventUserStatGet, api.PassportUserStatGetHandler)

	api.SetupAfterConnections(ctx, conn)

	return api
}

func (api *API) SetupAfterConnections(ctx context.Context, conn *pgxpool.Pool) {
	var factions []*server.Faction

	for !api.Passport.Connected {
		time.Sleep(5 * time.Second)
	}

	b := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    30 * time.Second,
		Factor: 2,
	}
	// S1008: should use 'return atomic.LoadInt32(&(b.flag)) != 0' instead of 'if atomic.LoadInt32(&(b.flag)) != 0 { return true }; return false' (gosimple)
	//        if atomic.LoadInt32(&(b.flag)) != 0 {

	// get factions from passport, retrying every 10 seconds until we ge them.
	for len(factions) <= 0 {
		if !api.Passport.Connected {
			time.Sleep(b.Duration())
			continue
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		api.Passport.FactionAll(func(msg []byte) {
			defer wg.Done()
			resp := &passport.FactionAllResponse{}
			err := json.Unmarshal(msg, resp)
			if err != nil {
				return
			}
			if err != nil {
				api.Passport.Log.Err(err).Msg("unable to get factions")
			}

			factions = resp.Factions
		})
		wg.Wait()

		if len(factions) > 0 {
			break
		}
		time.Sleep(b.Duration())
	}

	// listen to the client online and action channel
	// go api.ClientListener()

	go api.startSpoilOfWarBroadcaster(ctx)

	// build faction map for main server
	api.factionMap = make(map[server.FactionID]*server.Faction)
	for _, faction := range factions {
		err := db.FactionVotePriceGet(context.Background(), conn, faction)
		if err != nil {
			api.Log.Err(err).Msg("unable to get faction vote price")
		}

		api.factionMap[faction.ID] = faction

		// start live voting ticker
		api.liveSupsSpend[faction.ID] = &LiveVotingData{sync.Mutex{}, server.BigInt{Int: *big.NewInt(0)}}

		// game ability pool

		go api.StartGameAbilityPool(ctx, faction.ID, conn)
	}

	// initialise vote price system
	go api.startVotePriceSystem(ctx, factions, conn)

	// initialise voting cycle
	go api.StartVotingCycle(ctx, factions)

	// set faction map for battle arena server
	api.BattleArena.SetFactionMap(api.factionMap)

	// declare live sups spend broadcaster
	tickle.MinDurationOverride = true
	liveVotingBroadcasterLogger := log_helpers.NamedLogger(api.Log, "Live Sups spend Broadcaster").Level(zerolog.Disabled)
	liveVotingBroadcaster := tickle.New("Live Sups spend Broadcaster", 0.2, func() (int, error) {
		totalVote := server.BigInt{Int: *big.NewInt(0)}
		totalVoteMutex := sync.Mutex{}
		for _, faction := range factions {
			voteCount := big.NewInt(0)
			api.liveSupsSpend[faction.ID].Lock()
			voteCount.Add(voteCount, &api.liveSupsSpend[faction.ID].TotalVote.Int)
			api.liveSupsSpend[faction.ID].TotalVote = server.BigInt{Int: *big.NewInt(0)}
			api.liveSupsSpend[faction.ID].Unlock()

			// protect total vote
			totalVoteMutex.Lock()
			totalVote.Add(&totalVote.Int, voteCount)
			totalVoteMutex.Unlock()

		}

		// prepare payload
		payload := []byte{}
		payload = append(payload, byte(battle_arena.NetMessageTypeLiveVotingTick))
		payload = append(payload, []byte(totalVote.Int.String())...)

		api.NetMessageBus.Send(ctx, messagebus.NetBusKey(HubKeyLiveVoteUpdated), payload)

		return http.StatusOK, nil
	})
	liveVotingBroadcaster.Log = &liveVotingBroadcasterLogger

	// start live voting broadcaster
	liveVotingBroadcaster.Start()
}

// Event handlers
func (api *API) onlineEventHandler(ctx context.Context, wsc *hub.Client) error {
	// initialise a client detail channel if not on the list
	api.viewerLiveCount.Add(server.FactionID(uuid.Nil))

	// broadcast current game state
	go func() {
		ba := api.BattleArena.GetCurrentState()
		// delay 2 second to wait frontend setup key map
		time.Sleep(3 * time.Second)

		// marshal payload
		gsr := &GameSettingsResponse{
			GameMap:     ba.GameMap,
			WarMachines: ba.WarMachines,
		}
		if ba.BattleHistory != nil && len(ba.BattleHistory) > 0 {
			gsr.WarMachineLocation = ba.BattleHistory[0]
		}
		gameSettingsData, err := json.Marshal(&BroadcastPayload{
			Key:     HubKeyGameSettingsUpdated,
			Payload: gsr,
		})

		if err != nil {
			api.Log.Err(err).Msg("failed to marshal data")
			return
		}

		go wsc.Send(gameSettingsData)
	}()
	return nil
}

func (api *API) offlineEventHandler(ctx context.Context, wsc *hub.Client) error {
	currentUser := api.UserMap.GetUserDetail(wsc)

	noClientLeft := false
	if currentUser != nil {
		// remove client multipliers
		api.viewerLiveCount.Sub(currentUser.FactionID)
		api.UserMultiplier.Offline(currentUser.ID)
		// clean up the client detail map
		noClientLeft = api.UserMap.Remove(wsc)
	} else {
		api.viewerLiveCount.Sub(server.FactionID(uuid.Nil))
	}

	// set client offline
	// noClientLeft := api.ClientOffline(wsc)

	// check vote if there is not client instances of the offline user
	if noClientLeft && currentUser != nil && api.votePhaseChecker.Phase == VotePhaseLocationSelect {
		// check the user is selecting ability location
		api.votingCycle <- func(va *VoteAbility, fuvm FactionUserVoteMap, fts *FactionTransactions, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker, uvm UserVoteMap) {
			if len(vw.List) > 0 && vw.List[0].String() == currentUser.ID.String() {
				// pop out the first user of the list
				if len(vw.List) > 1 {
					vw.List = vw.List[1:]
				} else {
					vw.List = []server.UserID{}
				}

				// get next winner
				nextUser, winnerClientID := api.getNextWinnerDetail(vw)
				if nextUser == nil {
					// if no winner left, enter cooldown phase
					go api.BroadcastGameNotificationLocationSelect(ctx, &GameNotificationLocationSelect{
						Type:    LocationSelectTypeCancelledDisconnect,
						Ability: va.BattleAbility.Brief(),
					})

					// get random ability collection set
					battleAbility, factionAbilityMap, err := api.BattleArena.RandomBattleAbility()
					if err != nil {
						api.Log.Err(err)
					}

					go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteBattleAbilityUpdated), battleAbility)

					// initialise new ability collection
					va.BattleAbility = battleAbility

					// initialise new game ability map
					for fid, ability := range factionAbilityMap {
						va.FactionAbilityMap[fid] = ability
					}

					// voting phase change
					api.votePhaseChecker.Lock()
					api.votePhaseChecker.Phase = VotePhaseVoteCooldown
					api.votePhaseChecker.EndTime = time.Now().Add(time.Duration(va.BattleAbility.CooldownDurationSecond) * time.Second)
					api.votePhaseChecker.Unlock()

					// broadcast current stage to faction users
					go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)

					return
				}

				// otherwise, choose next winner
				api.votePhaseChecker.Lock()
				endTime := time.Now().Add(LocationSelectDurationSecond * time.Second)
				api.votePhaseChecker.Phase = VotePhaseLocationSelect
				api.votePhaseChecker.EndTime = endTime
				api.votePhaseChecker.Unlock()

				// otherwise announce another winner
				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyVoteWinnerAnnouncement, winnerClientID)), &WinnerSelectAbilityLocation{
					GameAbility: va.FactionAbilityMap[nextUser.FactionID],
					EndTime:     endTime,
				})

				// broadcast winner select location
				go api.BroadcastGameNotificationLocationSelect(ctx, &GameNotificationLocationSelect{
					Type:        LocationSelectTypeFailedDisconnect,
					Ability:     va.BattleAbility.Brief(),
					CurrentUser: currentUser.Brief(),
					NextUser:    nextUser.Brief(),
				})

				// broadcast current stage to faction users
				go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), api.votePhaseChecker)
			}
		}
	}
	return nil
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
