package api

import (
	"context"
	"encoding/json"
	"fmt"
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
	BattleID                    server.BattleID `json:"battleID"`
	TopSupsContributor          *server.User    `json:"topSupsContributor"`
	TopSupsContributeFaction    *server.Faction `json:"topSupsContributeFaction"`
	TopApplauseContributor      *server.User    `json:"topApplauseContributor"`
	MostFrequentAbilityExecutor *server.User    `json:"mostFrequentAbilityExecutor"`
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
	Conn         *pgxpool.Pool
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
	votingCycle      chan func(*VoteStage, *VoteAbility, FactionUserVoteMap, *FactionTotalVote, *VoteWinner, *VotingCycleTicker, UserVoteMap)
	votePriceSystem  *VotePriceSystem

	// faction abilities
	gameAbilityPool map[server.FactionID]chan func(GameAbilitiesPool, *GameAbilityPoolTicker)

	// viewer live count
	viewerLiveCount chan func(ViewerLiveCount, ViewerIDMap)

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
		Conn:         conn,
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
		votingCycle:   make(chan func(*VoteStage, *VoteAbility, FactionUserVoteMap, *FactionTotalVote, *VoteWinner, *VotingCycleTicker, UserVoteMap)),
		liveSupsSpend: make(map[server.FactionID]chan func(*LiveVotingData)),

		// channel for handling hub client
		hubClientDetail: make(map[*hub.Client]chan func(*HubClientDetail)),
		onlineClientMap: make(chan *ClientUpdate),

		// ring check auth
		ringCheckAuthChan: make(chan func(RingCheckAuthMap)),

		// game ability pool
		gameAbilityPool: make(map[server.FactionID]chan func(GameAbilitiesPool, *GameAbilityPoolTicker)),

		// faction viewer count
		viewerLiveCount: make(chan func(ViewerLiveCount, ViewerIDMap)),

		battleEndInfo: &BattleEndInfo{},
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
	_ = NewGameController(log, conn, api)

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

	// listen to the client online and action channel
	go api.ClientListener()

	go api.SetupAfterConnections(ctx, conn)

	return api
}

func (api *API) SetupAfterConnections(ctx context.Context, conn *pgxpool.Pool) {
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

	go api.initialiseViewerLiveCount(ctx, factions)
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
		api.liveSupsSpend[faction.ID] = make(chan func(*LiveVotingData))
		go api.startLiveVotingDataTicker(faction.ID)

		// game ability pool
		api.gameAbilityPool[faction.ID] = make(chan func(GameAbilitiesPool, *GameAbilityPoolTicker))
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
					err := c.SendWithMessageType(ctx, payload, websocket.MessageBinary)
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
		go api.viewerLiveCountAdd(server.FactionID(uuid.Nil))
	}

	// broadcast current game state
	go func() {
		ba := api.BattleArena.GetCurrentState()
		// delay 2 second to wait frontend setup key map
		time.Sleep(3 * time.Second)

		// marshal payload
		gameSettingsData, err := json.Marshal(&BroadcastPayload{
			Key: HubKeyGameSettingsUpdated,
			Payload: &GameSettingsResponse{
				GameMap:     ba.GameMap,
				WarMachines: ba.WarMachines,
				// WarMachineLocation: ba.BattleHistory[0],
			},
		})
		if err != nil {
			api.Log.Err(err).Msg("failed to marshal data")
			return
		}

		err = wsc.Send(ctx, gameSettingsData)
		if err != nil {
			api.Log.Err(err).Msg("failed to send broadcast")
		}
	}()
}

func (api *API) offlineEventHandler(ctx context.Context, wsc *hub.Client, clients hub.ClientsList, ch hub.TriggerChan) {
	currentUser, err := api.getClientDetailFromChannel(wsc)
	if err != nil {
		api.Log.Err(err).Msg("failed to get client detail")
	}
	go api.viewerLiveCountRemove(currentUser.FactionID)

	// set client offline
	noClientLeft := api.ClientOffline(wsc)

	// check vote if there is not client instances of the offline user
	if noClientLeft && api.votePhaseChecker.Phase == VotePhaseLocationSelect {
		// check the user is selecting ability location
		api.votingCycle <- func(vs *VoteStage, va *VoteAbility, fuvm FactionUserVoteMap, ftv *FactionTotalVote, vw *VoteWinner, vct *VotingCycleTicker, uvm UserVoteMap) {
			if len(vw.List) > 0 && vw.List[0].String() == wsc.Identifier() {
				if err != nil {
					api.Log.Err(err).Msg("failed to get user")
				}
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
						Type: LocationSelectTypeCancelledDisconnect,
						Ability: &AbilityBrief{
							Label:    va.BattleAbility.Label,
							ImageUrl: va.BattleAbility.ImageUrl,
							Colour:   va.BattleAbility.Colour,
						},
					})

					// get random ability collection set
					battleAbility, factionAbilityMap, err := api.BattleArena.RandomBattleAbility()
					if err != nil {
						api.Log.Err(err)
					}

					api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteBattleAbilityUpdated), battleAbility)

					// initialise new ability collection
					va.BattleAbility = battleAbility

					// initialise new game ability map
					for fid, ability := range factionAbilityMap {
						va.FactionAbilityMap[fid] = ability
					}

					// voting phase change
					api.votePhaseChecker.Phase = VotePhaseVoteCooldown
					vs.Phase = VotePhaseVoteCooldown
					vs.EndTime = time.Now().Add(time.Duration(va.BattleAbility.CooldownDurationSecond) * time.Second)

					// broadcast current stage to faction users
					api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), vs)

					return
				}

				// otherwise, choose next winner
				api.votePhaseChecker.Phase = VotePhaseLocationSelect
				vs.Phase = VotePhaseLocationSelect
				vs.EndTime = time.Now().Add(LocationSelectDurationSecond * time.Second)

				// otherwise announce another winner
				api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyVoteWinnerAnnouncement, winnerClientID)), &WinnerSelectAbilityLocation{
					GameAbility: va.FactionAbilityMap[nextUser.FactionID],
					EndTime:     vs.EndTime,
				})

				// broadcast winner select location
				go api.BroadcastGameNotificationLocationSelect(ctx, &GameNotificationLocationSelect{
					Type: LocationSelectTypeFailedDisconnect,
					Ability: &AbilityBrief{
						Label:    va.BattleAbility.Label,
						ImageUrl: va.BattleAbility.ImageUrl,
						Colour:   va.BattleAbility.Colour,
					},
					CurrentUser: &UserBrief{
						Username: currentUser.Username,
						AvatarID: currentUser.avatarID,
						Faction: &FactionBrief{
							Label:      api.factionMap[currentUser.FactionID].Label,
							Theme:      api.factionMap[currentUser.FactionID].Theme,
							LogoBlobID: api.factionMap[currentUser.FactionID].LogoBlobID,
						},
					},
					NextUser: &UserBrief{
						Username: nextUser.Username,
						AvatarID: nextUser.avatarID,
						Faction: &FactionBrief{
							Label:      api.factionMap[nextUser.FactionID].Label,
							Theme:      api.factionMap[nextUser.FactionID].Theme,
							LogoBlobID: api.factionMap[nextUser.FactionID].LogoBlobID,
						},
					},
				})

				// broadcast current stage to faction users
				api.MessageBus.Send(ctx, messagebus.BusKey(HubKeyVoteStageUpdated), vs)
			}
		}
	}

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
