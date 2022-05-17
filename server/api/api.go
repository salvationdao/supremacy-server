package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/helpers"
	"server/player_abilities"
	"server/rpcclient"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/meehow/securebytes"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/auth"
	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/ninja-syndicate/ws"
	"github.com/pemistahl/lingua-go"
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
	Routes                    chi.Router
	BattleArena               *battle.Arena
	HTMLSanitize              *bluemonday.Policy
	Hub                       *hub.Hub
	MessageBus                *messagebus.MessageBus
	SMS                       server.SMS
	Passport                  *rpcclient.PassportXrpcClient
	Telegram                  server.Telegram
	LanguageDetector          lingua.LanguageDetector
	Cookie                    *securebytes.SecureBytes
	IsCookieSecure            bool
	SalePlayerAbilitiesSystem *player_abilities.SalePlayerAbilitiesSystem
	Commander                 *ws.Commander
	SecureUserCommander       *ws.Commander
	SecureFactionCommander    *ws.Commander
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
		Config:                    config,
		ctx:                       ctx,
		Routes:                    chi.NewRouter(),
		MessageBus:                messageBus,
		HTMLSanitize:              HTMLSanitize,
		BattleArena:               battleArenaClient,
		Hub:                       gsHub,
		RingCheckAuthMap:          NewRingCheckMap(),
		Passport:                  pp,
		SMS:                       sms,
		Telegram:                  telegram,
		LanguageDetector:          languageDetector,
		IsCookieSecure:            config.CookieSecure,
		SalePlayerAbilitiesSystem: player_abilities.NewSalePlayerAbilitiesSystem(messageBus),
		Cookie: securebytes.New(
			[]byte(config.CookieKey),
			securebytes.ASN1Serializer{}),
		FactionPunishVote:    make(map[string]*PunishVoteTracker),
		FactionActivePlayers: make(map[string]*ActivePlayers),

		// chatroom
		GlobalChat:      NewChatroom(""),
		RedMountainChat: NewChatroom(server.RedMountainFactionID),
		BostonChat:      NewChatroom(server.BostonCyberneticsFactionID),
		ZaibatsuChat:    NewChatroom(server.ZaibatsuFactionID),
	}

	battleArenaClient.SetMessageBus(messageBus)

	api.Commander = ws.NewCommander(func(c *ws.Commander) {
		c.RestBridge("/rest")
	})
	api.SecureUserCommander = ws.NewCommander(func(c *ws.Commander) {
		c.RestBridge("/rest")
	}, "user commander")
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
	_ = NewBattleController(api)
	_ = NewPlayerAbilitiesController(api)

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
	api.Routes.Use(DatadogTracer.Middleware())

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

		r.Post("/video_server", WithToken(config.ServerStreamKey, WithError(api.CreateStreamHandler)))
		r.Get("/video_server", WithCookie(api, WithError(api.GetStreamsHandler)))
		r.Delete("/video_server", WithToken(config.ServerStreamKey, WithError(api.DeleteStreamHandler)))
		r.Post("/close_stream", WithToken(config.ServerStreamKey, WithError(api.CreateStreamCloseHandler)))
		r.Mount("/faction", FactionRouter(api))
		r.Get("/auth/xsyn", api.XSYNAuth)
		r.Get("/auth/check", WithError(api.AuthCheckHandler))

		r.Mount("/battle", BattleRouter(battleArenaClient))
		r.Post("/global_announcement", WithToken(config.ServerStreamKey, WithError(api.GlobalAnnouncementSend)))
		r.Delete("/global_announcement", WithToken(config.ServerStreamKey, WithError(api.GlobalAnnouncementDelete)))
		r.Get("/telegram/shortcode_registered", WithToken(config.ServerStreamKey, WithError(api.PlayerGetTelegramShortcodeRegistered)))

		r.Route("/ws", func(r chi.Router) {
			r.Use(ws.TrimPrefix("/api/ws"))
			r.Mount("/public", ws.NewServer(func(s *ws.Server) {
				s.Mount("/commander", api.Commander)
				s.WS("/*", "", nil)
				s.WS("/global_chat", HubKeyGlobalChatSubscribe, cc.GlobalChatUpdatedSubscribeHandler)
				s.WS("/global_announcement", server.HubKeyGlobalAnnouncementSubscribe, sc.GlobalAnnouncementSubscribe)
				s.WS("/live_data", server.HubKeySaleAbilityPriceSubscribe, nil)
			}))

			// battle arena route
			r.Mount("/battle", ws.NewServer(battleArenaClient.Route(api.AuthWS, api.AuthUserFactionWS)))

			// secured user route
			r.Mount("/user/{user_id}", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthWS(true, true))
				s.Mount("/user_commander", api.SecureUserCommander)
				s.WS("/*", HubKeyUserSubscribe, server.MustSecure(pc.PlayersSubscribeHandler))
				s.WS("/multipliers", battle.HubKeyMultiplierSubscribe, server.MustSecure(api.BattleArena.MultiplierUpdate))
			}))

			// secured faction route
			r.Mount("/faction/{faction_id}", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthUserFactionWS(true))
				s.WS("/*", HubKeyFactionActivePlayersSubscribe, server.MustSecureFaction(pc.FactionActivePlayersSubscribeHandler))
				s.Mount("/faction_commander", api.SecureFactionCommander)
				s.WS("/punish_vote", HubKeyPunishVoteSubscribe, server.MustSecureFaction(pc.PunishVoteSubscribeHandler))
				s.WS("/faction_chat", HubKeyFactionChatSubscribe, server.MustSecureFaction(cc.FactionChatUpdatedSubscribeHandler))
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

func (api *API) AuthCheckHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	cookie, err := r.Cookie("xsyn-token")
	if err != nil {
		// check whether token is attached
		gamelog.L.Debug().Msg("Cookie not found")

		token := r.URL.Query().Get("token")
		if token == "" {
			return http.StatusBadRequest, terror.Warn(fmt.Errorf("no cookie and token are provided"), "Player are not signed in.")
		}

		// check user from token
		player, err := api.TokenLogin(token)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to authentication")
		}

		// write cookie
		err = api.WriteCookie(w, token)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Failed to write cookie")
		}

		return helpers.EncodeJSON(w, player)
	}

	var token string
	if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to decrypt token")
	}

	// check user from token
	player, err := api.TokenLogin(token)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to authentication")
	}

	return helpers.EncodeJSON(w, player)
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

func (api *API) AuthUserFactionWS(factionIDMustMatch bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var token string

			cookie, err := r.Cookie("xsyn-token")
			if err != nil {
				fmt.Fprintf(w, "cookie not found: %v", err)
				return
			}

			if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
				fmt.Fprintf(w, "decryption error: %v", err)
				return
			}

			user, err := api.TokenLogin(token)
			if err != nil {
				fmt.Fprintf(w, "authentication error: %v", err)
				return
			}

			if !user.FactionID.Valid {
				fmt.Fprintf(w, "authentication error: user has not enlisted in one of the faction")
				return
			}

			if factionIDMustMatch {
				factionID := chi.URLParam(r, "faction_id")
				if factionID == "" || factionID != user.FactionID.String {
					fmt.Fprintf(w, "authentication location error: %v", err)
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

func (api *API) AuthWS(required bool, userIDMustMatch bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var token string

			cookie, err := r.Cookie("xsyn-token")
			if err != nil {
				fmt.Fprintf(w, "cookie not found: %v", err)
				token = r.URL.Query().Get("token")
				if token == "" {
					return
				}
			} else {
				if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
					fmt.Fprintf(w, "decryption error: %v", err)
					if required {
						return
					}
					next.ServeHTTP(w, r)
					return
				}
			}

			user, err := api.TokenLogin(token)
			if err != nil {
				fmt.Fprintf(w, "authentication error: %v", err)
				if required {
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			if userIDMustMatch {
				userID := chi.URLParam(r, "user_id")
				if userID == "" || userID != user.ID {
					fmt.Fprintf(w, "authentication location error: %v", err)
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
func (api *API) TokenLogin(tokenBase64 string) (*boiler.Player, error) {
	useResp, err := api.Passport.TokenLogin(tokenBase64)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to login with token")

		return nil, err
	}

	return boiler.FindPlayer(gamedb.StdConn, useResp.ID)
}
