package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"server"
	"server/battle"
	"server/db"
	"server/gamelog"
	"server/marketplace"
	"server/player_abilities"
	"server/profanities"
	"server/synctool"
	"server/syndicate"
	"server/xsyn_rpcclient"
	"sync"
	"time"

	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
	"github.com/pemistahl/lingua-go"
	"github.com/volatiletech/null/v8"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/meehow/securebytes"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/ws"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// WelcomePayload is the response sent when a client connects to the server
type WelcomePayload struct {
	Message string `json:"message"`
}

type LiveVotingData struct {
	sync.Mutex
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
	OuterLock      sync.Mutex
	NextAccessLock sync.Mutex
	DataLock       sync.Mutex

	// price
	CurrentVotePriceSups server.BigInt
	CurrentVotePerTick   int64
}

// API server
type API struct {
	ctx                      context.Context
	server                   *http.Server
	Routes                   chi.Router
	BattleArena              *battle.Arena
	HTMLSanitize             *bluemonday.Policy
	SMS                      server.SMS
	Passport                 *xsyn_rpcclient.XsynXrpcClient
	Telegram                 server.Telegram
	LanguageDetector         lingua.LanguageDetector
	Cookie                   *securebytes.SecureBytes
	IsCookieSecure           bool
	SalePlayerAbilityManager *player_abilities.SalePlayerAbilityManager
	Commander                *ws.Commander
	SecureUserCommander      *ws.Commander
	SecureFactionCommander   *ws.Commander

	// punish vote
	FactionPunishVote map[string]*PunishVoteTracker

	FactionActivePlayers map[string]*ActivePlayers

	// Marketplace
	MarketplaceController *marketplace.MarketplaceController

	// chatrooms
	GlobalChat       *Chatroom
	RedMountainChat  *Chatroom
	BostonChat       *Chatroom
	ZaibatsuChat     *Chatroom
	ProfanityManager *profanities.ProfanityManager

	SyndicateSystem *syndicate.System

	Config *server.Config

	SyncConfig *synctool.StaticSyncTool
}

// NewAPI registers routes
func NewAPI(
	ctx context.Context,
	battleArenaClient *battle.Arena,
	pp *xsyn_rpcclient.XsynXrpcClient,
	HTMLSanitize *bluemonday.Policy,
	config *server.Config,
	sms server.SMS,
	telegram server.Telegram,
	languageDetector lingua.LanguageDetector,
	pm *profanities.ProfanityManager,
	syncConfig *synctool.StaticSyncTool,
) (*API, error) {
	// spin up syndicate system
	ss, err := syndicate.NewSystem(pp, pm)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to spin up syndicate system")
		return nil, err
	}

	// initialise api
	api := &API{
		Config:                   config,
		ctx:                      ctx,
		Routes:                   chi.NewRouter(),
		HTMLSanitize:             HTMLSanitize,
		BattleArena:              battleArenaClient,
		Passport:                 pp,
		SMS:                      sms,
		Telegram:                 telegram,
		LanguageDetector:         languageDetector,
		IsCookieSecure:           config.CookieSecure,
		SalePlayerAbilityManager: player_abilities.NewSalePlayerAbilitiesSystem(),
		Cookie: securebytes.New(
			[]byte(config.CookieKey),
			securebytes.ASN1Serializer{}),
		FactionPunishVote:    make(map[string]*PunishVoteTracker),
		FactionActivePlayers: make(map[string]*ActivePlayers),

		// marketplace
		MarketplaceController: marketplace.NewMarketplaceController(pp),

		// chatroom
		GlobalChat:       NewChatroom(""),
		RedMountainChat:  NewChatroom(server.RedMountainFactionID),
		BostonChat:       NewChatroom(server.BostonCyberneticsFactionID),
		ZaibatsuChat:     NewChatroom(server.ZaibatsuFactionID),
		ProfanityManager: pm,
		SyndicateSystem:  ss,
		SyncConfig:       syncConfig,
	}

	api.Commander = ws.NewCommander(func(c *ws.Commander) {
		c.RestBridge("/rest")
	})
	api.SecureUserCommander = ws.NewCommander(func(c *ws.Commander) {
		c.RestBridge("/rest")
	})
	api.SecureFactionCommander = ws.NewCommander(func(c *ws.Commander) {
		c.RestBridge("/rest")
	})

	///////////////////////////
	//		 Controllers	 //
	///////////////////////////
	_ = NewCheckController(api)
	//_ = NewUserController(api)
	sc := NewStreamController(api)
	pc := NewPlayerController(api)
	cc := NewChatController(api)
	ssc := NewStoreController(api)
	_ = NewBattleController(api)
	mc := NewMarketplaceController(api)
	pac := NewPlayerAbilitiesController(api)
	pasc := NewPlayerAssetsController(api)
	_ = NewPlayerDevicesController(api)
	_ = NewHangarController(api)
	_ = NewCouponsController(api)
	NewSyndicateController(api)
	_ = NewLeaderboardController(api)
	_ = NewSystemMessagesController(api)
	NewWSWarMachineController(api)

	api.Routes.Use(middleware.RequestID)
	api.Routes.Use(middleware.RealIP)
	api.Routes.Use(gamelog.ChiLogger(zerolog.DebugLevel))
	api.Routes.Use(cors.New(
		cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}).Handler,
	)

	api.Routes.Handle("/metrics", promhttp.Handler())
	api.Routes.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			sentryHandler := sentryhttp.New(sentryhttp.Options{})
			r.Use(sentryHandler.Handle)
		})
		r.Mount("/check", CheckRouter(battleArenaClient, telegram, battleArenaClient.IsClientConnected))
		r.Mount("/stat", AssetStatsRouter(api))
		r.Mount(fmt.Sprintf("/%s/Supremacy_game", server.SupremacyGameUserID), PassportWebhookRouter(config.PassportWebhookSecret, api))

		// Web sockets are long-lived, so we don't want the sentry performance tracer running for the life-time of the connection.
		// See roothub.ServeHTTP for the setup of sentry on this route.
		//TODO ALEX reimplement handlers

		r.Group(func(r chi.Router) {
			if config.Environment != "development" {
				// TODO: Create new tracer not using HUB
				r.Use(DatadogTracer.Middleware())

			}

			if config.Environment == "development" {
				r.Get("/give_crates/{crate_type}/{public_address}", WithError(WithDev(api.DevGiveCrates)))
			}

			r.Get("/max_weapon_stats", WithError(api.GetMaxWeaponStats))
			r.Mount("/battle_history", BattleHistoryRouter())
			r.Mount("/faction", FactionRouter(api))
			r.Mount("/feature", FeatureRouter(api))
			r.Mount("/auth", AuthRouter(api))
			r.Mount("/player_abilities", PlayerAbilitiesRouter(api))
			r.Mount("/sale_abilities", SaleAbilitiesRouter(api))
			r.Mount("/system_messages", SystemMessagesRouter(api))

			r.Mount("/battle", BattleRouter(battleArenaClient))
			r.Get("/telegram/shortcode_registered", WithToken(config.ServerStreamKey, WithError(api.PlayerGetTelegramShortcodeRegistered)))

			r.Mount("/syndicate", SyndicateRouter(api))
		})

		r.Mount("/", AdminRoutes(api, config.ServerStreamKey))

		r.Post("/sync_data/{branch}", WithToken(config.ServerStreamKey, WithError(api.SyncStaticData)))
		r.Post("/profanities/add", WithToken(config.ServerStreamKey, WithError(api.AddPhraseToProfanityDictionary)))

		r.Route("/ws", func(r chi.Router) {
			r.Use(ws.TrimPrefix("/api/ws"))

			// public route ws
			r.Mount("/public", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthWS(false, false, "/sale_abilities"))

				s.Mount("/commander", api.Commander)
				s.WS("/online", "", nil)
				s.WS("/global_chat", HubKeyGlobalChatSubscribe, cc.GlobalChatUpdatedSubscribeHandler)
				s.WS("/global_announcement", server.HubKeyGlobalAnnouncementSubscribe, sc.GlobalAnnouncementSubscribe)

				// endpoint for demoing battle ability showcase to non-login player
				s.WS("/battle_ability", battle.HubKeyBattleAbilityUpdated, api.BattleArena.PublicBattleAbilityUpdateSubscribeHandler)

				s.WS("/minimap", battle.HubKeyMinimapUpdatesSubscribe, api.BattleArena.MinimapUpdatesSubscribeHandler)
				s.WS("/sale_abilities", server.HubKeySaleAbilitiesList, server.MustSecure(pac.SaleAbilitiesListHandler), MustLogin)

				// come from battle
				s.WS("/notification", battle.HubKeyGameNotification, nil)
				s.WS("/mech/{mech_id}/details", HubKeyPlayerAssetMechDetailPublic, pasc.PlayerAssetMechDetailPublic)

				s.WS("/game_settings", battle.HubKeyGameSettingsUpdated, battleArenaClient.SendSettings)
				s.WSBatch("/mech/{slotNumber}", "/public/mech", battle.HubKeyWarMachineStatUpdated, battleArenaClient.WarMachineStatSubscribe)
				s.WS("/bribe_stage", battle.HubKeyBribeStageUpdateSubscribe, battleArenaClient.BribeStageSubscribe)
				s.WS("/live_data", "", nil)
				s.WS("/global_active_players", HubKeyGlobalActivePlayersSubscribe, pc.GlobalActivePlayersSubscribeHandler)
			}))

			// secured user route ws
			r.Mount("/user/{user_id}", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthWS(true, true))
				s.Mount("/user_commander", api.SecureUserCommander)
				s.WSTrack("/*", "user_id", server.HubKeyUserSubscribe, server.MustSecure(pc.PlayersSubscribeHandler))
				s.WS("/player_abilities", server.HubKeyPlayerAbilitiesList, server.MustSecure(pac.PlayerAbilitiesListHandler))
				s.WS("/punishment_list", HubKeyPlayerPunishmentList, server.MustSecure(pc.PlayerPunishmentList))
				s.WS("/player_weapons", server.HubKeyPlayerWeaponsList, server.MustSecure(pasc.PlayerWeaponsListHandler))
				s.WS("/battle_ability/check_opt_in", battle.HubKeyBattleAbilityOptInCheck, server.MustSecure(battleArenaClient.BattleAbilityOptInSubscribeHandler), MustHaveFaction)
				s.WS("/system_messages", server.HubKeySystemMessageListUpdatedSubscribe, nil)
			}))

			// secured faction route ws
			r.Mount("/faction/{faction_id}", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthUserFactionWS(true))
				s.WS("/*", HubKeyFactionActivePlayersSubscribe, server.MustSecureFaction(pc.FactionActivePlayersSubscribeHandler))
				s.Mount("/faction_commander", api.SecureFactionCommander)
				s.WS("/punish_vote", HubKeyPunishVoteSubscribe, server.MustSecureFaction(pc.PunishVoteSubscribeHandler))
				s.WS("/punish_vote/{punish_vote_id}/command_override", HubKeyPunishVoteCommandOverrideCountSubscribe, server.MustSecureFaction(pc.PunishVoteCommandOverrideCountSubscribeHandler))
				s.WS("/faction_chat", HubKeyFactionChatSubscribe, server.MustSecureFaction(cc.FactionChatUpdatedSubscribeHandler))
				s.WS("/marketplace/{id}", HubKeyMarketplaceSalesItemUpdate, server.MustSecureFaction(mc.SalesItemUpdateSubscriber))

				s.WS("/mech/{mech_id}/details", HubKeyPlayerAssetMechDetail, server.MustSecureFaction(pasc.PlayerAssetMechDetail))
				s.WS("/weapon/{weapon_id}/details", HubKeyPlayerAssetWeaponDetail, server.MustSecureFaction(pasc.PlayerAssetWeaponDetail))

				// subscription from battle
				s.WS("/queue", battle.WSQueueStatusSubscribe, server.MustSecureFaction(battleArenaClient.QueueStatusSubscribeHandler))
				s.WS("/queue/{mech_id}", battle.WSPlayerAssetMechQueueSubscribe, server.MustSecureFaction(battleArenaClient.PlayerAssetMechQueueSubscribeHandler))
				s.WS("/queue-update", battle.WSPlayerAssetMechQueueUpdateSubscribe, nil)
				s.WS("/crate/{crate_id}", HubKeyMysteryCrateSubscribe, server.MustSecureFaction(ssc.MysteryCrateSubscribeHandler))

				s.WS("/mech_command/{hash}", battle.HubKeyMechMoveCommandSubscribe, server.MustSecureFaction(api.BattleArena.MechMoveCommandSubscriber))
				s.WS("/mech_commands", battle.HubKeyMechCommandsSubscribe, server.MustSecureFaction(api.BattleArena.MechCommandsSubscriber))
				s.WS("/mech_command_notification", battle.HubKeyGameNotification, nil)

				s.WS("/mech/{mech_id}/repair_status", server.WarMachineRepairStatusSubscribe, server.MustSecureFaction(api.WarMachineRepairStatusSubscribeHandler))
				s.WS("/mech/{mech_id}/repair-update", battle.WSPlayerAssetMechQueueUpdateSubscribe, nil)

				s.WS("/mech/{slotNumber}/abilities", battle.HubKeyWarMachineAbilitiesUpdated, server.MustSecureFaction(battleArenaClient.WarMachineAbilitiesUpdateSubscribeHandler))
				s.WS("/mech/{slotNumber}/abilities/{mech_ability_id}/cool_down_seconds", battle.HubKeyWarMachineAbilitySubscribe, server.MustSecureFaction(battleArenaClient.WarMachineAbilitySubscribe))

				// syndicate related
				s.WS("/syndicate/{syndicate_id}", server.HubKeySyndicateGeneralDetailSubscribe, server.MustSecureFaction(api.SyndicateGeneralDetailSubscribeHandler), MustMatchSyndicate)
				s.WS("/syndicate/{syndicate_id}/directors", server.HubKeySyndicateDirectorsSubscribe, server.MustSecureFaction(api.SyndicateDirectorsSubscribeHandler), MustMatchSyndicate)
				s.WS("/syndicate/{syndicate_id}/committees", server.HubKeySyndicateCommitteesSubscribe, server.MustSecureFaction(api.SyndicateCommitteesSubscribeHandler), MustMatchSyndicate)
				s.WS("/syndicate/{syndicate_id}/ongoing_motions", server.HubKeySyndicateOngoingMotionSubscribe, server.MustSecureFaction(api.SyndicateOngoingMotionSubscribeHandler), MustMatchSyndicate)
				s.WS("/syndicate/{syndicate_id}/ongoing_election", server.HubKeySyndicateOngoingElectionSubscribe, server.MustSecureFaction(api.SyndicateOngoingElectionSubscribeHandler), MustMatchSyndicate)
			}))
		})
	})

	// create a tickle that update faction mvp every day 00:00 am
	factionMvpUpdate := tickle.New("Calculate faction mvp player", 24*60*60, func() (int, error) {
		// set red mountain mvp player
		gamelog.L.Info().Str("faction_id", server.RedMountainFactionID).Msg("Recalculate Red Mountain mvp player")
		err := db.FactionStatMVPUpdate(server.RedMountainFactionID)
		if err != nil {
			gamelog.L.Error().Str("faction_id", server.RedMountainFactionID).Err(err).Msg("Failed to recalculate Red Mountain mvp player")
		}

		// set boston mvp player
		gamelog.L.Info().Str("faction_id", server.BostonCyberneticsFactionID).Msg("Recalculate Boston mvp player")
		err = db.FactionStatMVPUpdate(server.BostonCyberneticsFactionID)
		if err != nil {
			gamelog.L.Error().Str("faction_id", server.BostonCyberneticsFactionID).Err(err).Msg("Failed to recalculate Boston mvp player")
		}

		// set Zaibatsu mvp player
		gamelog.L.Info().Str("faction_id", server.ZaibatsuFactionID).Msg("Recalculate Zaibatsu mvp player")
		err = db.FactionStatMVPUpdate(server.ZaibatsuFactionID)
		if err != nil {
			gamelog.L.Error().Str("faction_id", server.ZaibatsuFactionID).Err(err).Msg("Failed to recalculate Zaibatsu mvp player")
		}

		return http.StatusOK, nil
	})
	factionMvpUpdate.Log = gamelog.L

	err = factionMvpUpdate.SetIntervalAt(24*time.Hour, 0, 0)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to set up faction mvp user update tickle")
	}

	// spin up a punish vote handlers for each faction
	err = api.PunishVoteTrackerSetup()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to setup punish vote tracker")
	}

	api.FactionActivePlayerSetup()

	return api, nil
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

	api.BattleArena.Serve()

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

func (api *API) AuthUserFactionWS(factionIDMustMatch bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var token string
			var ok bool

			cookie, err := r.Cookie("xsyn-token")
			if err != nil {
				token = r.URL.Query().Get("token")
				if token == "" {
					token, ok = r.Context().Value("token").(string)
					if !ok || token == "" {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
				}
			} else {
				if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
					gamelog.L.Error().Err(err).Msg("decrypting cookie error")
					return
				}
			}

			user, err := api.TokenLogin(token)
			if err != nil {
				fmt.Fprintf(w, "authentication error: %v", err)
				return
			}

			// get ip
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ipaddr, _, _ := net.SplitHostPort(r.RemoteAddr)
				userIP := net.ParseIP(ipaddr)
				if userIP == nil {
					ip = ipaddr
				} else {
					ip = userIP.String()
				}
			}

			// upsert player ip logs
			err = db.PlayerIPUpsert(user.ID, ip)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to log player ip")
				fmt.Fprintf(w, "invalid ip address")
				return
			}

			if !user.FactionID.Valid {
				fmt.Fprintf(w, "authentication error: user has not enlisted in one of the factions")
				return
			}

			if factionIDMustMatch {
				factionID := chi.URLParam(r, "faction_id")
				if factionID == "" || factionID != user.FactionID.String {
					fmt.Fprintf(w, "faction id check failed... url faction id: %s, user faction id: %s, url:%s", factionID, user.FactionID.String, r.URL.Path)
					return
				}
			}

			ctxWithUserID := context.WithValue(r.Context(), "user_id", user.ID)
			ctx := context.WithValue(ctxWithUserID, "faction_id", user.FactionID.String)
			*r = *r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func (api *API) AuthWS(required bool, userIDMustMatch bool, onlyAuthPaths ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var token string
			var ok bool

			cookie, err := r.Cookie("xsyn-token")
			if err != nil {
				token = r.URL.Query().Get("token")
				if token == "" {
					token, ok = r.Context().Value("token").(string)
					if !ok || token == "" {
						if required {
							gamelog.L.Debug().Err(err).Msg("missing token and cookie")
							http.Error(w, "Unauthorized", http.StatusUnauthorized)
							return
						}
						next.ServeHTTP(w, r)
						return
					}
				}
			} else {
				if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
					if required {
						gamelog.L.Debug().Err(err).Msg("decrypting cookie error")
						return
					}
					next.ServeHTTP(w, r)
					return
				}
			}

			if !required {
				path := r.URL.Path

				exists := false
				for _, p := range onlyAuthPaths {
					if p == path {
						exists = true
						break
					}
				}

				if !exists {
					next.ServeHTTP(w, r)
					return
				}
			}

			user, err := api.TokenLogin(token)
			if err != nil {
				gamelog.L.Debug().Err(err).Msg("authentication error")
				return
			}

			// get ip
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ipaddr, _, _ := net.SplitHostPort(r.RemoteAddr)
				userIP := net.ParseIP(ipaddr)
				if userIP == nil {
					ip = ipaddr
				} else {
					ip = userIP.String()
				}
			}

			// upsert player ip logs
			err = db.PlayerIPUpsert(user.ID, ip)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to log player ip")
				return
			}

			if userIDMustMatch {
				userID := chi.URLParam(r, "user_id")
				if userID == "" || userID != user.ID {
					gamelog.L.Debug().Err(fmt.Errorf("user id check failed")).
						Str("userID", userID).
						Str("user.ID", user.ID).
						Str("r.URL.Path", r.URL.Path).
						Msg("user id check failed")
					return
				}
			}

			ctx := context.WithValue(r.Context(), "user_id", user.ID)
			*r = *r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		return http.HandlerFunc(fn)
	}
}

// TokenLogin gets a user from the token
func (api *API) TokenLogin(tokenBase64 string, ignoreErr ...bool) (*server.Player, error) {
	ignoreError := len(ignoreErr) > 0 && ignoreErr[0] == true

	userResp, err := api.Passport.TokenLogin(tokenBase64)
	if err != nil {
		if !ignoreError {
			if err.Error() == "session is expired" {
				gamelog.L.Debug().Err(err).Msg("Failed to login with token")
			}
			gamelog.L.Error().Err(err).Msg("Failed to login with token")
		}
		return nil, err
	}

	err = api.UpsertPlayer(userResp.ID, null.StringFrom(userResp.Username), userResp.PublicAddress, userResp.FactionID, nil)
	if err != nil {
		if !ignoreError {
			gamelog.L.Error().Err(err).Msg("Failed to update player detail")
		}
		return nil, err
	}

	serverPlayer, err := db.GetPlayer(userResp.ID)
	if err != nil {
		if !ignoreError {
			gamelog.L.Error().Err(err).Msg("Failed to get player by ID")
		}
		return nil, err
	}

	return serverPlayer, nil
}
