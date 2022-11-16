package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"server"
	"server/battle"
	"server/db"
	"server/discord"
	"server/fiat"
	"server/gamelog"
	"server/marketplace"
	"server/profanities"
	"server/quest"
	"server/sale_player_abilities"
	"server/synctool"
	"server/syndicate"
	"server/xsyn_rpcclient"
	"server/zendesk"
	"time"

	"github.com/ninja-software/tickle"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v72/client"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/meehow/securebytes"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-syndicate/ws"
	"github.com/pemistahl/lingua-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// API server
type API struct {
	ctx                      context.Context
	server                   *http.Server
	Routes                   chi.Router
	ArenaManager             *battle.ArenaManager
	HTMLSanitize             *bluemonday.Policy
	StripeClient             *client.API
	StripeWebhookSecret      string
	SMS                      server.SMS
	Passport                 *xsyn_rpcclient.XsynXrpcClient
	Telegram                 server.Telegram
	Discord                  *discord.DiscordSession
	Zendesk                  *zendesk.Zendesk
	LanguageDetector         lingua.LanguageDetector
	Cookie                   *securebytes.SecureBytes
	IsCookieSecure           bool
	SalePlayerAbilityManager *sale_player_abilities.SalePlayerAbilityManager
	Commander                *ws.Commander
	SecureUserCommander      *ws.Commander
	SecureFactionCommander   *ws.Commander

	// punish vote
	FactionPunishVote map[string]*PunishVoteTracker

	FactionActivePlayers map[string]*ActivePlayers

	VoiceChatListeners *VoiceChatListeners

	// marketplace
	MarketplaceController *marketplace.MarketplaceController

	// fiat
	FiatController *fiat.FiatController

	// chatrooms
	GlobalChat       *Chatroom
	RedMountainChat  *Chatroom
	BostonChat       *Chatroom
	ZaibatsuChat     *Chatroom
	ProfanityManager *profanities.ProfanityManager

	// captcha
	captcha *captcha

	SyndicateSystem *syndicate.System

	Config *server.Config

	SyncConfig *synctool.StaticSyncTool

	questManager *quest.System

	ViewerUpdateChan chan bool

	ChallengeFund decimal.Decimal
}

// NewAPI registers routes
func NewAPI(
	ctx context.Context,
	arenaManager *battle.ArenaManager,
	pp *xsyn_rpcclient.XsynXrpcClient,
	HTMLSanitize *bluemonday.Policy,
	stripeClient *client.API,
	stripeWebhookSecret string,
	config *server.Config,
	sms server.SMS,
	telegram server.Telegram,
	discord *discord.DiscordSession,
	zendesk *zendesk.Zendesk,
	languageDetector lingua.LanguageDetector,
	pm *profanities.ProfanityManager,
	syncConfig *synctool.StaticSyncTool,
	questManager *quest.System,
	privateKeySignerHex string,
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
		ArenaManager:             arenaManager,
		Passport:                 pp,
		SMS:                      sms,
		Telegram:                 telegram,
		Discord:                  discord,
		StripeClient:             stripeClient,
		StripeWebhookSecret:      stripeWebhookSecret,
		Zendesk:                  zendesk,
		LanguageDetector:         languageDetector,
		IsCookieSecure:           config.CookieSecure,
		SalePlayerAbilityManager: sale_player_abilities.NewSalePlayerAbilitiesSystem(),
		Cookie: securebytes.New(
			[]byte(config.CookieKey),
			securebytes.ASN1Serializer{}),
		FactionPunishVote:    make(map[string]*PunishVoteTracker),
		FactionActivePlayers: make(map[string]*ActivePlayers),

		// marketplace
		MarketplaceController: marketplace.NewMarketplaceController(pp),

		// fiat
		FiatController: fiat.NewFiatController(pp, stripeClient),

		// chatroom
		GlobalChat:       NewChatroom(""),
		RedMountainChat:  NewChatroom(server.RedMountainFactionID),
		BostonChat:       NewChatroom(server.BostonCyberneticsFactionID),
		ZaibatsuChat:     NewChatroom(server.ZaibatsuFactionID),
		ProfanityManager: pm,
		SyndicateSystem:  ss,
		SyncConfig:       syncConfig,
		captcha: &captcha{
			secret:    config.CaptchaSecret,
			siteKey:   config.CaptchaSiteKey,
			verifyUrl: "https://hcaptcha.com/siteverify",
		},
		questManager: questManager,

		VoiceChatListeners: &VoiceChatListeners{},

		ViewerUpdateChan: make(chan bool),
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
	pac := NewAbilitiesController(api)
	pasc := NewPlayerAssetsController(api)
	_ = NewPlayerDevicesController(api)
	_ = NewHangarController(api)
	_ = NewCouponsController(api)
	NewSyndicateController(api)
	NewLeaderboardController(api)
	_ = NewSystemMessagesController(api)
	NewMechRepairController(api)
	fc := NewFiatController(api)
	_ = NewReplayController(api)
	NewVoiceStreamController(api)
	BattleQueueController(api)
	NewMarketplaceController(api)
	NewModToolsController(api)
	NewAdminController(api)
	NewModToolsController(api)

	api.Routes.Use(middleware.RequestID)
	api.Routes.Use(middleware.RealIP)
	api.Routes.Use(server.AddOriginToCtx())
	api.Routes.Use(gamelog.ChiLogger(zerolog.DebugLevel))
	api.Routes.Use(cors.New(
		cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		}).Handler,
	)

	api.Routes.Handle("/metrics", promhttp.Handler())
	api.Routes.Post("/stripe-webhook", WithError(fc.StripeWebhook))
	api.Routes.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			sentryHandler := sentryhttp.New(sentryhttp.Options{})
			r.Use(sentryHandler.Handle)
		})
		r.Mount("/check", CheckRouter(arenaManager, telegram, arenaManager.IsClientConnected))
		r.Mount("/stat", AssetStatsRouter(api))
		r.Mount(fmt.Sprintf("/%s/Supremacy_game", server.SupremacyGameUserID), PassportWebhookRouter(config.PassportWebhookSecret, api))

		r.Group(func(r chi.Router) {
			r.Use(server.RestDatadogTrace(config.Environment))

			r.Get("/max_weapon_stats", WithError(api.GetMaxWeaponStats))
			r.Mount("/battle_history", BattleHistoryRouter(privateKeySignerHex))
			r.Mount("/faction", FactionRouter(api))
			r.Mount("/feature", FeatureRouter(api))
			r.Mount("/auth", AuthRouter(api))
			r.Mount("/player_abilities", PlayerAbilitiesRouter(api))
			r.Mount("/replay", BattleReplayRouter(api))
			r.Mount("/sale_abilities", SaleAbilitiesRouter(api))
			r.Mount("/system_messages", SystemMessagesRouter(api))

			r.Mount("/battle", BattleRouter(api))
			r.Get("/telegram/shortcode_registered", WithToken(config.ServerStreamKey, WithError(api.PlayerGetTelegramShortcodeRegistered)))

			r.Mount("/syndicate", SyndicateRouter(api))
			r.Mount("/", AdminRoutes(api, config.ServerStreamKey))

			r.Post("/sync_data/{branch}", WithToken(config.ServerStreamKey, WithError(api.SyncStaticData)))
			r.Post("/profanities/add", WithToken(config.ServerStreamKey, WithError(api.AddPhraseToProfanityDictionary)))
		})

		r.Route("/ws", func(r chi.Router) {
			r.Use(ws.TrimPrefix("/api/ws"))

			// public route ws
			r.Mount("/public", ws.NewServer(func(s *ws.Server) {
				s.Mount("/commander", api.Commander)
				s.WS("/online", "", nil)
				s.WS("/global_chat", HubKeyGlobalChatSubscribe, cc.GlobalChatUpdatedSubscribeHandler)
				s.WS("/global_announcement", server.HubKeyGlobalAnnouncementSubscribe, sc.GlobalAnnouncementSubscribe)
				s.WS("/global_active_players", HubKeyGlobalActivePlayersSubscribe, pc.GlobalActivePlayersSubscribeHandler)

				s.WS("/challenge_fund", server.HubKeyChallengeFundSubscribe, api.ChallengeFundSubscribeHandler)

				s.WS("/arena_list", server.HubKeyBattleArenaListSubscribe, api.ArenaListSubscribeHandler)
				s.WS("/arena/{arena_id}/closed", server.HubKeyBattleArenaClosedSubscribe, api.ArenaClosedSubscribeHandler)

				// come from battle
				s.WS("/mech/{mech_id}/details", HubKeyPlayerAssetMechDetailPublic, pasc.PlayerAssetMechDetailPublic)
				s.WS("/mech/{mech_id}/is_staked", HubKeyMechIsStaked, api.MechIsStaked)
				s.WS("/custom_avatar/{avatar_id}/details", HubKeyPlayerCustomAvatarDetails, pc.ProfileCustomAvatarDetailsHandler)

				// battle related endpoint
				s.WS("/arena/{arena_id}/upcoming_battle", server.HubKeyNextBattleDetails, api.NextBattleDetails)
				s.WS("/arena/{arena_id}/notification", battle.HubKeyGameNotification, nil)
				s.WS("/arena/{arena_id}/game_settings", battle.HubKeyGameSettingsUpdated, api.ArenaManager.SendSettings)
				s.WS("/arena/{arena_id}/battle_end_result", battle.HubKeyBattleEndDetailUpdated, api.BattleEndDetail)
				s.WS("/arena/{arena_id}/battle_state", server.HubKeyBattleState, api.BattleState)

				s.WS("/live_viewer_count", HubKeyViewerLiveCountUpdated, api.LiveViewerCount)
			}))

			r.Mount("/secure", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthWS(false))
				s.WS("/sale_abilities", server.HubKeySaleAbilitiesListSubscribe, pac.SaleAbilitiesListSubscribeHandler)
				s.WS("/repair_offer/{offer_id}", server.HubKeyRepairOfferSubscribe, api.RepairOfferSubscribe)
				s.WS("/repair_offer/update", server.HubKeyRepairOfferUpdateSubscribe, api.RepairOfferList)
				s.WS("/mech/{mech_id}/repair_case", server.HubKeyMechRepairCase, api.MechRepairCaseSubscribe)
				s.WS("/mech/{mech_id}/active_repair_offer", server.HubKeyMechActiveRepairOffer, api.MechActiveRepairOfferSubscribe)
				s.WS("/battle_eta", server.HubKeyBattleETAUpdate, api.BattleETASubscribeHandler)
				s.WS("/game_map_list", HubKeyGameMapList, api.GameMapListSubscribeHandler)

				// user related
				s.WSTrack("/user/{user_id}", "user_id", server.HubKeyUserSubscribe, server.MustSecure(pc.PlayersSubscribeHandler), MustMatchUserID)
				s.WS("/user/{user_id}/owned_mechs", server.HubKeyPlayerOwnedMechs, server.MustSecure(api.PlayerMechs), MustMatchUserID)
				s.WS("/user/{user_id}/owned_weapons", server.HubKeyPlayerOwnedWeapons, server.MustSecure(api.PlayerWeapons), MustMatchUserID)
				s.WS("/user/{user_id}/stat", server.HubKeyUserStatSubscribe, server.MustSecure(pc.PlayersStatSubscribeHandler), MustMatchUserID)
				s.WS("/user/{user_id}/rank", server.HubKeyPlayerRankGet, server.MustSecure(pc.PlayerRankGet), MustMatchUserID)
				s.WS("/user/{user_id}/player_abilities", server.HubKeyPlayerAbilitiesList, server.MustSecure(pac.PlayerAbilitiesListHandler), MustMatchUserID)
				s.WS("/user/{user_id}/punishment_list", HubKeyPlayerPunishmentList, server.MustSecure(pc.PlayerPunishmentList), MustMatchUserID)
				s.WS("/user/{user_id}/system_messages", server.HubKeySystemMessageListUpdatedSubscribe, nil, MustMatchUserID)
				s.WS("/user/{user_id}/telegram_shortcode_register", server.HubKeyTelegramShortcodeRegistered, nil, MustMatchUserID)
				s.WS("/user/{user_id}/quest_stat", server.HubKeyPlayerQuestStats, server.MustSecure(pc.PlayerQuestStat), MustMatchUserID)
				s.WS("/user/{user_id}/quest_progression", server.HubKeyPlayerQuestProgressions, server.MustSecure(pc.PlayerQuestProgressions), MustMatchUserID)
				s.WS("/user/{user_id}/arena/{arena_id}", server.HubKeyVoiceStreams, server.MustSecure(api.VoiceStreamSubscribe), MustMatchUserID)
				s.WS("/user/{user_id}/arena/{arena_id}/listeners", server.HubKeyVoiceStreams, server.MustSecure(api.VoiceStreamListenersSubscribe), MustMatchUserID)

				s.WS("/user/{user_id}/queue_status", server.HubKeyPlayerQueueStatus, server.MustSecure(pc.PlayerQueueStatusHandler), MustMatchUserID)

				s.WS("/user/{user_id}/involved_battle_lobbies", server.HubKeyInvolvedBattleLobbyListUpdate, server.MustSecureFaction(api.PlayerInvolvedBattleLobbies))

				// fiat related
				s.WS("/user/{user_id}/shopping_cart_updated", server.HubKeyShoppingCartUpdated, server.MustSecure(fc.ShoppingCartUpdatedSubscriber), MustMatchUserID)
				s.WS("/user/{user_id}/shopping_cart_expired", server.HubKeyShoppingCartExpired, nil, MustMatchUserID)

				// user repair bay
				s.WS("/user/{user_id}/repair_bay", server.HubKeyMechRepairSlots, server.MustSecure(api.PlayerMechRepairSlots), MustMatchUserID)

				s.WS("/user/{user_id}/repair_agent/{repair_agent_id}/next_block", server.HubKeyNextRepairGameBlock, server.MustSecure(api.NextRepairBlock), MustMatchUserID)
			}))

			// secured user commander
			r.Mount("/user/{user_id}", ws.NewServer(func(s *ws.Server) {
				s.Use(api.AuthWS(true))
				s.Mount("/user_commander", api.SecureUserCommander)

				s.WS("/battle/{battle_id}/supporter_abilities", server.HubKeyPlayerSupportAbilities, server.MustSecure(pac.PlayerSupportAbilitiesHandler), MustMatchUserID)
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
				s.WS("/battle_lobbies", server.HubKeyBattleLobbyListUpdate, server.MustSecureFaction(api.BattleLobbyListUpdate))
				s.WS("/private_battle_lobby/{access_code}", server.HubKeyPrivateBattleLobbyUpdate, server.MustSecureFaction(api.PrivateBattleLobbyUpdate), MustHaveUrlParam("access_code"))

				s.WS("/mech/{mech_id}/details", HubKeyPlayerAssetMechDetail, server.MustSecureFaction(pasc.PlayerAssetMechDetail))
				s.WS("/mech/{mech_id}/brief_info", HubKeyPlayerAssetMechDetail, server.MustSecureFaction(pasc.PlayerAssetMechBriefInfo))
				s.WS("/utility/{utility_id}/details", HubKeyPlayerAssetUtilityDetail, server.MustSecureFaction(pasc.PlayerAssetUtilityDetail))
				s.WS("/weapon/{weapon_id}/details", HubKeyPlayerAssetWeaponDetail, server.MustSecureFaction(pasc.PlayerAssetWeaponDetail))
				s.WS("/power_core/{power_core_id}/details", HubKeyPlayerAssetPowerCoreDetail, server.MustSecureFaction(pasc.PlayerAssetPowerCoreDetail))

				s.WS("/crate/{crate_id}", server.HubKeyMysteryCrateSubscribe, server.MustSecureFaction(ssc.MysteryCrateSubscribeHandler))
				s.WS("/queue/{mech_id}", server.HubKeyPlayerAssetMechQueueSubscribe, server.MustSecureFaction(api.PlayerAssetMechQueueSubscribeHandler))

				s.WS("/staked_mechs", server.HubKeyFactionStakedMechs, server.MustSecureFaction(api.FactionStakedMechs))

				// subscription from battle
				s.WS("/arena/{arena_id}/mech/{slotNumber}/abilities", battle.HubKeyWarMachineAbilitiesUpdated, server.MustSecureFaction(api.ArenaManager.WarMachineAbilitiesUpdateSubscribeHandler))
				s.WS("/arena/{arena_id}/mech/{slotNumber}/abilities/{mech_ability_id}/cool_down_seconds", battle.HubKeyWarMachineAbilitySubscribe, server.MustSecureFaction(api.ArenaManager.WarMachineAbilitySubscribe))

				// syndicate related
				s.WS("/syndicate/{syndicate_id}", server.HubKeySyndicateGeneralDetailSubscribe, server.MustSecureFaction(api.SyndicateGeneralDetailSubscribeHandler), MustMatchSyndicate)
				s.WS("/syndicate/{syndicate_id}/directors", server.HubKeySyndicateDirectorsSubscribe, server.MustSecureFaction(api.SyndicateDirectorsSubscribeHandler), MustMatchSyndicate)
				s.WS("/syndicate/{syndicate_id}/committees", server.HubKeySyndicateCommitteesSubscribe, server.MustSecureFaction(api.SyndicateCommitteesSubscribeHandler), MustMatchSyndicate)
				s.WS("/syndicate/{syndicate_id}/ongoing_motions", server.HubKeySyndicateOngoingMotionSubscribe, server.MustSecureFaction(api.SyndicateOngoingMotionSubscribeHandler), MustMatchSyndicate)
				s.WS("/syndicate/{syndicate_id}/ongoing_election", server.HubKeySyndicateOngoingElectionSubscribe, server.MustSecureFaction(api.SyndicateOngoingElectionSubscribeHandler), MustMatchSyndicate)
			}))

			// mini map related
			r.Route("/mini_map/arena/{arena_id}", func(r chi.Router) {
				r.Mount("/public", ws.NewServer(func(s *ws.Server) {
					s.WSBinary("/minimap_events", api.ArenaManager.MinimapEventsSubscribeHandler)
					s.WSBinary("/mech_stats", api.ArenaManager.WarMachineStatsSubscribe)
					s.WS("/mini_map_ability_display_list", server.HubKeyMiniMapAbilityContentSubscribe, api.MiniMapAbilityDisplayList)
					s.WS("/minimap", server.HubKeyMiniMapUpdateSubscribe, api.ArenaManager.MinimapUpdatesSubscribeHandler)
				}))

				r.Mount("/faction/{faction_id}", ws.NewServer(func(s *ws.Server) {
					s.Use(api.AuthUserFactionWS(true))
					s.WS("/mech_command/{hash}", server.HubKeyMechCommandUpdateSubscribe, server.MustSecureFaction(api.ArenaManager.MechMoveCommandSubscriber))
					s.WS("/mech_commands", server.HubKeyFactionMechCommandUpdateSubscribe, server.MustSecureFaction(api.ArenaManager.MechCommandsSubscriber))
				}))
			})

		})
	})

	err = api.initialWSBroadcast()
	if err != nil {
		return nil, err
	}

	return api, nil
}

// initialWSBroadcast include all the initial go routines that trigger ws broadcast
// IMPORTANT: All the initial broadcast functions need to be triggered AFTER the ws tree is built.
// otherwise, the server will panic!!!
func (api *API) initialWSBroadcast() error {
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

	// spin up a punishment vote handlers for each faction
	err = api.PunishVoteTrackerSetup()
	if err != nil {
		gamelog.L.Error().Err(err).Msg("Failed to setup punish vote tracker")
	}

	api.FactionActivePlayerSetup()
	go api.ChallengeFundDebounceBroadcast()

	// set user online debounce
	go api.debounceSendingViewerCount()

	// start player rank updater
	api.ArenaManager.PlayerRankUpdater()

	// check default battle lobbies
	err = api.ArenaManager.SetDefaultPublicBattleLobbies()
	if err != nil {
		return err
	}

	// start repair offer cleaner
	go api.ArenaManager.RepairOfferCleaner()

	// start debounce lobby update sender
	go api.ArenaManager.DebounceSendBattleLobbiesUpdate()

	return nil
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

	api.ArenaManager.Serve()

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

func (api *API) ChallengeFundDebounceBroadcast() {
	// initialise challenge fund
	api.ChallengeFund = decimal.Zero
	interval := 500 * time.Millisecond

	timer := time.NewTimer(interval)
	for {
		select {
		case <-api.ArenaManager.ChallengeFundUpdateChan:
			timer.Reset(interval)
		case <-timer.C:
			api.ChallengeFund = api.Passport.UserBalanceGet(uuid.FromStringOrNil(server.SupremacyChallengeFundUserID))
			ws.PublishMessage("/public/challenge_fund", server.HubKeyChallengeFundSubscribe, api.ChallengeFund)
		}
	}

}
