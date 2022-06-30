package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"server"
	"server/battle"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/marketplace"
	"server/player_abilities"
	"server/profanities"
	"server/xsyn_rpcclient"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
	"github.com/pemistahl/lingua-go"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/meehow/securebytes"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/terror/v2"
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
	ctx                       context.Context
	server                    *http.Server
	Routes                    chi.Router
	BattleArena               *battle.Arena
	HTMLSanitize              *bluemonday.Policy
	SMS                       server.SMS
	Passport                  *xsyn_rpcclient.XsynXrpcClient
	Telegram                  server.Telegram
	LanguageDetector          lingua.LanguageDetector
	Cookie                    *securebytes.SecureBytes
	IsCookieSecure            bool
	SalePlayerAbilitiesSystem *player_abilities.SalePlayerAbilitiesSystem
	Commander                 *ws.Commander
	SecureUserCommander       *ws.Commander
	SecureFactionCommander    *ws.Commander

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

	Config *server.Config
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
) *API {
	// initialise api
	api := &API{
		Config:                    config,
		ctx:                       ctx,
		Routes:                    chi.NewRouter(),
		HTMLSanitize:              HTMLSanitize,
		BattleArena:               battleArenaClient,
		Passport:                  pp,
		SMS:                       sms,
		Telegram:                  telegram,
		LanguageDetector:          languageDetector,
		IsCookieSecure:            config.CookieSecure,
		SalePlayerAbilitiesSystem: player_abilities.NewSalePlayerAbilitiesSystem(),
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
	_ = NewPlayerAssetsController(api)
	_ = NewHangarController(api)
	_ = NewCouponsController(api)

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

			r.Get("/give_crates/{public_address}", WithError(WithDev(api.DevGiveCrates)))

			r.Post("/video_server", WithToken(config.ServerStreamKey, WithError(api.CreateStreamHandler)))
			r.Get("/video_server", WithError(api.GetStreamsHandler))
			r.Delete("/video_server", WithToken(config.ServerStreamKey, WithError(api.DeleteStreamHandler)))
			r.Post("/close_stream", WithToken(config.ServerStreamKey, WithError(api.CreateStreamCloseHandler)))
			r.Mount("/faction", FactionRouter(api))
			r.Mount("/auth", AuthRouter(api))

			r.Mount("/battle", BattleRouter(battleArenaClient))
			r.Post("/global_announcement", WithToken(config.ServerStreamKey, WithError(api.GlobalAnnouncementSend)))
			r.Delete("/global_announcement", WithToken(config.ServerStreamKey, WithError(api.GlobalAnnouncementDelete)))

			r.Get("/telegram/shortcode_registered", WithToken(config.ServerStreamKey, WithError(api.PlayerGetTelegramShortcodeRegistered)))

			r.Post("/chat_shadowban", WithToken(config.ServerStreamKey, WithError(api.ShadowbanChatPlayer)))
			r.Post("/chat_shadowban/remove", WithToken(config.ServerStreamKey, WithError(api.ShadowbanChatPlayerRemove)))
			r.Get("/chat_shadowban/list", WithToken(config.ServerStreamKey, WithError(api.ShadowbanChatPlayerList)))

		})

		r.Post("/profanities/add", WithToken(config.ServerStreamKey, WithError(api.AddPhraseToProfanityDictionary)))

		r.Route("/ws", func(r chi.Router) {
			r.Use(ws.TrimPrefix("/api/ws"))

			// public route ws
			r.Mount("/public", ws.NewServer(func(s *ws.Server) {
				s.Mount("/commander", api.Commander)
				s.WS("/global_chat", HubKeyGlobalChatSubscribe, cc.GlobalChatUpdatedSubscribeHandler)
				s.WS("/global_announcement", server.HubKeyGlobalAnnouncementSubscribe, sc.GlobalAnnouncementSubscribe)

				// endpoint for demoing battle ability showcase to non-login player
				s.WS("/battle_ability", battle.HubKeyBattleAbilityUpdated, api.BattleArena.PublicBattleAbilityUpdateSubscribeHandler)

				// come from battle
				s.WS("/notification", battle.HubKeyGameNotification, nil)
				s.WSBatch("/mech/{slotNumber}", "/public/mech", battle.HubKeyWarMachineStatUpdated, battleArenaClient.WarMachineStatUpdatedSubscribe)
			}))

			r.Mount("/secure_public", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthWS(true, false))
				s.WS("/sale_abilities", server.HubKeySaleAbilitiesList, server.MustSecure(pac.SaleAbilitiesListHandler))
			}))

			// battle arena route ws
			r.Mount("/battle", ws.NewServer(func(s *ws.Server) {
				s.WS("/*", battle.HubKeyGameSettingsUpdated, battleArenaClient.SendSettings)
				s.WS("/bribe_stage", battle.HubKeyBribeStageUpdateSubscribe, battleArenaClient.BribeStageSubscribe)
				s.WS("/live_data", "", nil)
			}))

			// secured user route ws
			r.Mount("/user/{user_id}", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthWS(true, true))
				s.Mount("/user_commander", api.SecureUserCommander)
				s.WS("/*", HubKeyUserSubscribe, server.MustSecure(pc.PlayersSubscribeHandler))
				s.WS("/multipliers", battle.HubKeyMultiplierSubscribe, server.MustSecure(battleArenaClient.MultiplierUpdate))
				s.WS("/mystery_crates", HubKeyMysteryCrateOwnershipSubscribe, server.MustSecure(ssc.MysteryCrateOwnershipSubscribeHandler))
				s.WS("/player_abilities", server.HubKeyPlayerAbilitiesList, server.MustSecure(pac.PlayerAbilitiesListHandler))
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

				// subscription from battle
				s.WS("/queue", battle.WSQueueStatusSubscribe, server.MustSecureFaction(battleArenaClient.QueueStatusSubscribeHandler))
				s.WS("/queue/{mech_id}", battle.WSPlayerAssetMechQueueSubscribe, server.MustSecureFaction(battleArenaClient.PlayerAssetMechQueueSubscribeHandler))
				s.WS("/queue-update", battle.WSPlayerAssetMechQueueUpdateSubscribe, nil)
				s.WS("/crate/{crate_id}", HubKeyMysteryCrateSubscribe, server.MustSecureFaction(ssc.MysteryCrateSubscribeHandler))

				s.WS("/mech_command/{hash}", battle.HubKeyMechMoveCommandSubscribe, server.MustSecureFaction(api.BattleArena.MechMoveCommandSubscriber))
				s.WS("/mech_commands", battle.HubKeyMechCommandsSubscribe, server.MustSecureFaction(api.BattleArena.MechCommandsSubscriber))
				s.WS("/mech_command_notification", battle.HubKeyGameNotification, nil)
			}))

			// handle abilities ws
			r.Mount("/ability/{faction_id}", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthUserFactionWS(true))
				s.WS("/*", battle.HubKeyBattleAbilityUpdated, server.MustSecureFaction(battleArenaClient.BattleAbilityUpdateSubscribeHandler))
				s.WS("/faction", battle.HubKeyFactionUniqueAbilitiesUpdated, server.MustSecureFaction(battleArenaClient.FactionAbilitiesUpdateSubscribeHandler))
				s.WS("/mech/{slotNumber}", battle.HubKeyWarMachineAbilitiesUpdated, server.MustSecureFaction(battleArenaClient.WarMachineAbilitiesUpdateSubscribeHandler))
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

func (api *API) AuthWS(required bool, userIDMustMatch bool) func(next http.Handler) http.Handler {
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
					if required {
						gamelog.L.Error().Err(err).Msg("decrypting cookie error")
						return
					}
					next.ServeHTTP(w, r)
					return
				}
			}

			user, err := api.TokenLogin(token)
			if err != nil {
				if required {
					gamelog.L.Error().Err(err).Msg("authentication error")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			if userIDMustMatch {
				userID := chi.URLParam(r, "user_id")
				if userID == "" || userID != user.ID {
					gamelog.L.Error().Err(fmt.Errorf("user id check failed")).
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
func (api *API) TokenLogin(tokenBase64 string) (*boiler.Player, error) {
	userResp, err := api.Passport.TokenLogin(tokenBase64)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to login with token")
		return nil, err
	}

	err = api.UpsertPlayer(userResp.ID, null.StringFrom(userResp.Username), userResp.PublicAddress, userResp.FactionID, nil)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to update player detail")
		return nil, err
	}

	return boiler.FindPlayer(gamedb.StdConn, userResp.ID)
}

func WithDev(next func(w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		if os.Getenv("PASSPORT_ENVIRONMENT") != "development" {
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}
		devPass := r.Header.Get("X-Authorization")
		if devPass != "NinjaDojo_!" {
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}

		return next(w, r)
	}
	return fn
}

func (api *API) DevGiveCrates(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := common.HexToAddress(chi.URLParam(r, "public_address"))
	user, err := boiler.Players(boiler.PlayerWhere.PublicAddress.EQ(null.StringFrom(publicAddress.String()))).One(gamedb.StdConn)
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to get player by pub address")

		return http.StatusInternalServerError, err
	}

	// purchase crate
	tx, err := gamedb.StdConn.Begin()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("start tx2: %w", err)
	}
	defer tx.Rollback()

	storeCrate, err := boiler.StorefrontMysteryCrates(
		boiler.StorefrontMysteryCrateWhere.MysteryCrateType.EQ("WEAPON"),
		boiler.StorefrontMysteryCrateWhere.FactionID.EQ(user.FactionID.String),
		qm.Load(boiler.StorefrontMysteryCrateRels.Faction),
	).One(gamedb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get crate for purchase, please try again or contact support.")
	}

	assignedCrate, err := assignAndRegisterPurchasedCrate(user.ID, storeCrate, tx, api)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to purchase mystery crate, please try again or contact support.")
	}

	err = tx.Commit()
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Could not open mystery crate, please try again or contact support.")
	}

	// open crate
	tx2, err := gamedb.StdConn.Begin()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("start tx2: %w", err)
	}
	defer tx2.Rollback()
	collectionItem, err := boiler.CollectionItems(
		boiler.CollectionItemWhere.ItemID.EQ(assignedCrate.ItemID),
		boiler.CollectionItemWhere.ItemType.EQ(boiler.ItemTypeMysteryCrate),
	).One(tx2)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Could not find collection item, try again or contact support.")
	}

	// update mystery crate
	crate := boiler.MysteryCrate{}
	q := `
	UPDATE mystery_crate
	SET opened = TRUE
	WHERE id = $1 AND opened = FALSE 
	RETURNING id, type, faction_id, label, opened, locked_until, purchased, deleted_at, updated_at, created_at, description`
	err = gamedb.StdConn.
		QueryRow(q, collectionItem.ItemID).
		Scan(
			&crate.ID,
			&crate.Type,
			&crate.FactionID,
			&crate.Label,
			&crate.Opened,
			&crate.LockedUntil,
			&crate.Purchased,
			&crate.DeletedAt,
			&crate.UpdatedAt,
			&crate.CreatedAt,
			&crate.Description,
		)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Could not find crate, try again or contact support.")
	}

	items := OpenCrateResponse{}

	blueprintItems, err := crate.MysteryCrateBlueprints().All(tx2)
	if err != nil {
		gamelog.L.Error().Err(err).Msg(fmt.Sprintf("failed to get blueprint relationships from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
		return http.StatusInternalServerError, terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
	}

	for _, blueprintItem := range blueprintItems {
		switch blueprintItem.BlueprintType {
		case boiler.TemplateItemTypeMECH:
			bp, err := db.BlueprintMech(blueprintItem.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get mech blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
			}

			mech, err := db.InsertNewMech(tx2, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new mech from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get mech during crate opening, try again or contact support.")
			}
			items.Mech = mech
		case boiler.TemplateItemTypeWEAPON:
			bp, err := db.BlueprintWeapon(blueprintItem.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get weapon blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get weapon blueprint during crate opening, try again or contact support.")
			}

			weapon, err := db.InsertNewWeapon(tx2, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new weapon from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get weapon during crate opening, try again or contact support.")
			}
			items.Weapons = append(items.Weapons, weapon)
		case boiler.TemplateItemTypeMECH_SKIN:
			bp, err := db.BlueprintMechSkinSkin(blueprintItem.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get mech skin blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get mech skin blueprint during crate opening, try again or contact support.")
			}

			mechSkin, err := db.InsertNewMechSkin(tx2, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new mech skin from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get mech skin during crate opening, try again or contact support.")
			}
			items.MechSkin = mechSkin
		case boiler.TemplateItemTypeWEAPON_SKIN:
			bp, err := db.BlueprintWeaponSkin(blueprintItem.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get weapon skin blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get weapon skin blueprint during crate opening, try again or contact support.")
			}
			weaponSkin, err := db.InsertNewWeaponSkin(tx2, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new weapon skin from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get weapon skin during crate opening, try again or contact support.")
			}
			items.WeaponSkin = weaponSkin
		case boiler.TemplateItemTypePOWER_CORE:
			bp, err := db.BlueprintPowerCore(blueprintItem.BlueprintID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to get powercore blueprint from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get powercore blueprint during crate opening, try again or contact support.")
			}

			powerCore, err := db.InsertNewPowerCore(tx2, uuid.FromStringOrNil(user.ID), bp)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Interface("crate blueprint", blueprintItem).Msg(fmt.Sprintf("failed to insert new powercore from crate: %s, for user: %s, CRATE:OPEN", crate.ID, user.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not get powercore during crate opening, try again or contact support.")
			}
			items.PowerCore = powerCore
		}
	}

	if crate.Type == boiler.CrateTypeMECH {
		//attach mech_skin to mech - mech
		err = db.AttachMechSkinToMech(tx2, user.ID, items.Mech.ID, items.MechSkin.ID, false)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach mech skin to mech during CRATE:OPEN crate: %s", crate.ID))
			return http.StatusInternalServerError, terror.Error(err, "Could not open crate, try again or contact support.")
		}

		err = db.AttachPowerCoreToMech(tx2, user.ID, items.Mech.ID, items.PowerCore.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach powercore to mech during CRATE:OPEN crate: %s", crate.ID))
			return http.StatusInternalServerError, terror.Error(err, "Could not open crate, try again or contact support.")
		}

		//attach weapons to mech -mech
		for _, weapon := range items.Weapons {
			err = db.AttachWeaponToMech(tx2, user.ID, items.Mech.ID, weapon.ID)
			if err != nil {
				gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach weapons to mech during CRATE:OPEN crate: %s", crate.ID))
				return http.StatusInternalServerError, terror.Error(err, "Could not open crate, try again or contact support.")
			}
		}

		mech, err := db.Mech(tx2, items.Mech.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to get final mech during CRATE:OPEN crate: %s", crate.ID))
			return http.StatusInternalServerError, terror.Error(err, "Could not open crate, try again or contact support.")
		}
		items.Mech = mech
	}

	if crate.Type == boiler.CrateTypeWEAPON {
		//attach weapon_skin to weapon -weapon
		if len(items.Weapons) != 1 {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("too many weapons in crate: %s", crate.ID))
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("too many weapons in weapon crate"), "Could not open crate, try again or contact support.")
		}
		err = db.AttachWeaponSkinToWeapon(tx2, user.ID, items.Weapons[0].ID, items.WeaponSkin.ID)
		if err != nil {
			gamelog.L.Error().Err(err).Interface("crate", crate).Msg(fmt.Sprintf("failed to attach weapon skin to weapon during CRATE:OPEN crate: %s", crate.ID))
			return http.StatusInternalServerError, terror.Error(err, "Could not open crate, try again or contact support.")
		}
	}

	err = tx2.Commit()
	if err != nil {
		// gamelog.L.Error().Err(err).Interface("crate", crate).Msg("failed to open mystery crate")
		return http.StatusInternalServerError, terror.Error(err, "Could not open mystery crate, please try again or contact support.")
	}

	return http.StatusOK, nil
}
