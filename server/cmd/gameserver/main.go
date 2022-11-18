package main

import (
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os/signal"
	"runtime"
	"server"
	"server/api"
	"server/asset"
	"server/battle"
	"server/comms"
	"server/db"
	"server/db/boiler"
	"server/discord"
	"server/gamedb"
	"server/gamelog"
	"server/profanities"
	"server/quest"
	"server/replay"
	"server/slack"
	"server/sms"
	"server/synctool"
	"server/telegram"
	"server/voice_chat"
	"server/xsyn_rpcclient"
	"server/zendesk"

	"github.com/volatiletech/null/v8"

	"github.com/ninja-software/terror/v2"

	"github.com/gofrs/uuid"
	"github.com/pemistahl/lingua-go"
	"github.com/stripe/stripe-go/v72/client"
	"github.com/urfave/cli/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/ninja-syndicate/ws"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	_ "net/http/pprof"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/log_helpers"
	"github.com/rs/zerolog"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"

	"context"
	"os"
)

// Variable passed in at compile time using `-ldflags`
var (
	Version          string // -X main.Version=$(git describe --tags --abbrev=0)
	GitHash          string // -X main.GitHash=$(git rev-parse HEAD)
	GitBranch        string // -X main.GitBranch=$(git rev-parse --abbrev-ref HEAD)
	BuildDate        string // -X main.BuildDate=$(date -u +%Y%m%d%H%M%S)
	UnCommittedFiles string // -X main.UnCommittedFiles=$(git status --porcelain | wc -l)"
)

const SentryReleasePrefix = "supremacy-gameserver"
const envPrefix = "GAMESERVER"

func main() {
	runtime.GOMAXPROCS(2)
	app := &cli.App{
		Compiled: time.Now(),
		Usage:    "Run the server server",
		Authors: []*cli.Author{
			{
				Name:  "Ninja Software",
				Email: "hello@ninjasoftware.com.au",
			},
		},
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			{
				// This is not using the built in version so ansible can more easily read the version
				Name: "version",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "full", Usage: "Prints full version and build info", Value: false},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("full") {
						fmt.Printf("Version=%s\n", Version)
						fmt.Printf("Commit=%s\n", GitHash)
						fmt.Printf("Branch=%s\n", GitBranch)
						fmt.Printf("BuildDate=%s\n", BuildDate)
						fmt.Printf("WorkingCopyState=%s uncommitted\n", UnCommittedFiles)
						return nil
					}
					fmt.Printf("%s\n", Version)
					return nil
				},
			},
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Flags: []cli.Flag{
					&cli.Uint64Flag{Name: "game_client_minimum_build_no", EnvVars: []string{envPrefix + "_GAMECLIENT_MINIMUM_BUILD_NO", "GAMECLIENT_MINIMUM_BUILD_NO"}, Usage: "The gameclient version the server is using."},

					&cli.StringFlag{Name: "database_user", Value: "gameserver", EnvVars: []string{envPrefix + "_DATABASE_USER", "DATABASE_USER"}, Usage: "The database user"},
					&cli.StringFlag{Name: "database_pass", Value: "dev", EnvVars: []string{envPrefix + "_DATABASE_PASS", "DATABASE_PASS"}, Usage: "The database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{envPrefix + "_DATABASE_HOST", "DATABASE_HOST"}, Usage: "The database host"},
					&cli.StringFlag{Name: "database_port", Value: "5437", EnvVars: []string{envPrefix + "_DATABASE_PORT", "DATABASE_PORT"}, Usage: "The database port"},
					&cli.StringFlag{Name: "database_name", Value: "gameserver", EnvVars: []string{envPrefix + "_DATABASE_NAME", "DATABASE_NAME"}, Usage: "The database name"},
					&cli.StringFlag{Name: "database_application_name", Value: "API Server", EnvVars: []string{envPrefix + "_DATABASE_APPLICATION_NAME"}, Usage: "Postgres database name"},

					&cli.StringFlag{Name: "log_level", Value: "DebugLevel", EnvVars: []string{envPrefix + "_LOG_LEVEL"}, Usage: "Set the log level for zerolog (Options: PanicLevel, FatalLevel, ErrorLevel, WarnLevel, InfoLevel, DebugLevel, TraceLevel"},
					&cli.StringFlag{Name: "environment", Value: "development", DefaultText: "development", EnvVars: []string{envPrefix + "_ENVIRONMENT", "ENVIRONMENT"}, Usage: "This program environment (development, testing, training, staging, production), it sets the log levels"},
					&cli.StringFlag{Name: "sentry_dsn_backend", Value: "", EnvVars: []string{envPrefix + "_SENTRY_DSN_BACKEND", "SENTRY_DSN_BACKEND"}, Usage: "Sends error to remote server. If set, it will send error."},
					&cli.StringFlag{Name: "sentry_server_name", Value: "dev-pc", EnvVars: []string{envPrefix + "_SENTRY_SERVER_NAME", "SENTRY_SERVER_NAME"}, Usage: "The machine name that this program is running on."},
					&cli.Float64Flag{Name: "sentry_sample_rate", Value: 1, EnvVars: []string{envPrefix + "_SENTRY_SAMPLE_RATE", "SENTRY_SAMPLE_RATE"}, Usage: "The percentage of trace sample to collect (0.0-1)"},

					&cli.StringFlag{Name: "battle_arena_addr", Value: ":8083", EnvVars: []string{envPrefix + "_BA_ADDR", "API_ADDR"}, Usage: ":port to run the battle arena server"},
					&cli.StringFlag{Name: "passport_addr", Value: "ws://localhost:8086/api/ws", EnvVars: []string{envPrefix + "_PASSPORT_ADDR", "PASSPORT_ADDR"}, Usage: " address of the passport server, inc protocol"},
					&cli.StringFlag{Name: "api_addr", Value: ":8084", EnvVars: []string{envPrefix + "_API_ADDR"}, Usage: ":port to run the API"},
					&cli.StringFlag{Name: "twitch_ui_web_host_url", Value: "http://localhost:8081", EnvVars: []string{"TWITCH_HOST_URL_FRONTEND"}, Usage: "Twitch url for CORS"},

					&cli.StringFlag{Name: "rootpath", Value: "../web/build", EnvVars: []string{envPrefix + "_ROOTPATH"}, Usage: "folder path of index.html"},
					&cli.StringFlag{Name: "userauth_jwtsecret", Value: "872ab3df-d7c7-4eb6-a052-4146d0f4dd15", EnvVars: []string{envPrefix + "_USERAUTH_JWTSECRET"}, Usage: "JWT secret"},

					&cli.BoolFlag{Name: "cookie_secure", Value: true, EnvVars: []string{envPrefix + "_COOKIE_SECURE", "COOKIE_SECURE"}, Usage: "set cookie secure"},

					&cli.StringFlag{Name: "cookie_key", Value: "asgk236tkj2kszaxfj.,.135j25khsafkahfgiu215hi2htkjahsgfih13kj56hkqhkahgbkashgk312ht5lk2qhafga", EnvVars: []string{envPrefix + "_COOKIE_KEY", "COOKIE_KEY"}, Usage: "cookie encryption key"},
					&cli.StringFlag{Name: "google_client_id", Value: "", EnvVars: []string{envPrefix + "_GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_ID"}, Usage: "Google Client ID for OAuth functionaility."},

					// SMS stuff
					&cli.StringFlag{Name: "twilio_sid", Value: "", EnvVars: []string{envPrefix + "_TWILIO_ACCOUNT_SID"}, Usage: "Twilio account sid"},
					&cli.StringFlag{Name: "twilio_api_key", Value: "", EnvVars: []string{envPrefix + "_TWILIO_API_KEY"}, Usage: "Twilio api key"},
					&cli.StringFlag{Name: "twilio_api_secret", Value: "", EnvVars: []string{envPrefix + "_TWILIO_API_SECRET"}, Usage: "Twilio api secret"},
					&cli.StringFlag{Name: "sms_from_number", Value: "", EnvVars: []string{envPrefix + "_SMS_FROM_NUMBER"}, Usage: "Number to send SMS from"},

					// telegram bot token
					&cli.StringFlag{Name: "telegram_bot_token", Value: "", EnvVars: []string{envPrefix + "_TELEGRAM_BOT_TOKEN"}, Usage: "telegram bot token"},

					// stripe stuff
					&cli.StringFlag{Name: "stripe_webhook_secret", Value: "", EnvVars: []string{envPrefix + "_STRIPE_WEBHOOK_SECRET"}, Usage: "stripe payment webhook secret key"},
					&cli.StringFlag{Name: "stripe_secret_key", Value: "", EnvVars: []string{envPrefix + "_STRIPE_SECRET_KEY"}, Usage: "stripe payment api secret key"},

					// TODO: clear up token
					&cli.BoolFlag{Name: "jwt_encrypt", Value: true, EnvVars: []string{envPrefix + "_JWT_ENCRYPT", "JWT_ENCRYPT"}, Usage: "set if to encrypt jwt tokens or not"},
					&cli.StringFlag{Name: "jwt_encrypt_key", Value: "ITF1vauAxvJlF0PLNY9btOO9ZzbUmc6X", EnvVars: []string{envPrefix + "_JWT_KEY", "JWT_KEY"}, Usage: "supports key sizes of 16, 24 or 32 bytes"},
					&cli.IntFlag{Name: "jwt_expiry_days", Value: 1, EnvVars: []string{envPrefix + "_JWT_EXPIRY_DAYS", "JWT_EXPIRY_DAYS"}, Usage: "expiry days for auth tokens"},
					&cli.StringFlag{Name: "jwt_key", Value: "9a5b8421bbe14e5a904cfd150a9951d3", EnvVars: []string{"STREAM_SITE_JWT_KEY"}, Usage: "JWT Key for signing token on stream site"},

					&cli.StringFlag{Name: "passport_server_token", Value: "e79422b7-7bfe-4463-897b-a1d22bf2e0bc", EnvVars: []string{envPrefix + "_PASSPORT_TOKEN"}, Usage: "Token to auth to passport server"},
					&cli.StringFlag{Name: "server_stream_key", Value: "6c7b4a82-7797-4847-836e-978399830878", EnvVars: []string{envPrefix + "_SERVER_STREAM_KEY"}, Usage: "Authorization key to crud servers"},
					&cli.StringFlag{Name: "passport_webhook_secret", Value: "e1BD3FF270804c6a9edJDzzDks87a8a4fde15c7=", EnvVars: []string{"PASSPORT_WEBHOOK_SECRET"}, Usage: "Authorization key to passport webhook"},

					&cli.IntFlag{Name: "database_max_idle_conns", Value: 40, EnvVars: []string{envPrefix + "_DATABASE_MAX_IDLE_CONNS"}, Usage: "Database max idle conns"},
					&cli.IntFlag{Name: "database_max_open_conns", Value: 50, EnvVars: []string{envPrefix + "_DATABASE_MAX_OPEN_CONNS"}, Usage: "Database max open conns"},

					&cli.BoolFlag{Name: "pprof_datadog", Value: true, EnvVars: []string{envPrefix + "_PPROF_DATADOG"}, Usage: "Use datadog pprof to collect debug info"},
					&cli.StringSliceFlag{Name: "pprof_datadog_profiles", Value: cli.NewStringSlice("cpu", "heap"), EnvVars: []string{envPrefix + "_PPROF_DATADOG_PROFILES"}, Usage: "Comma seprated list of profiles to collect. Options: cpu,heap,block,mutex,goroutine,metrics"},
					&cli.DurationFlag{Name: "pprof_datadog_interval_sec", Value: 60, EnvVars: []string{envPrefix + "_PPROF_DATADOG_INTERVAL_SEC"}, Usage: "Specifies the period at which profiles will be collected"},
					&cli.DurationFlag{Name: "pprof_datadog_duration_sec", Value: 60, EnvVars: []string{envPrefix + "_PPROF_DATADOG_DURATION_SEC"}, Usage: "Specifies the length of the CPU profile snapshot"},

					&cli.StringFlag{Name: "auth_callback_url", Value: "https://play.supremacygame.io/login-redirect", EnvVars: []string{envPrefix + "_AUTH_CALLBACK_URL"}, Usage: "The url for gameserver to redirect after completing the auth flow"},
					&cli.StringFlag{Name: "auth_hangar_callback_url", Value: "https://hangar.supremacygame.io", EnvVars: []string{envPrefix + "_AUTH_HANGAR_CALLBACK_URL"}, Usage: "The url for gameserver to redirect after completing the auth flow"},

					&cli.BoolFlag{Name: "sync_keycards", Value: false, EnvVars: []string{envPrefix + "_SYNC_KEYCARDS"}, Usage: "Sync keycard data from .csv file"},
					&cli.StringFlag{Name: "keycard_csv_path", Value: "", EnvVars: []string{envPrefix + "_KEYCARD_CSV_PATH"}, Usage: "File path for csv to sync keycards"},

					&cli.StringFlag{Name: "github_token", Value: "", EnvVars: []string{envPrefix + "_GITHUB_ACCESS_TOKEN", "GITHUB_PAT"}, Usage: "Github token for access to private repo"},

					&cli.StringFlag{Name: "captcha_site_key", Value: "", EnvVars: []string{envPrefix + "_CAPTCHA_SITE_KEY", "CAPTCHA_SITE_KEY"}, Usage: "Captcha site key"},
					&cli.StringFlag{Name: "captcha_secret", Value: "", EnvVars: []string{envPrefix + "_CAPTCHA_SECRET", "CAPTCHA_SECRET"}, Usage: "Captcha secret"},

					&cli.StringFlag{Name: "zendesk_token", Value: "", EnvVars: []string{envPrefix + "_ZENDESK_TOKEN"}, Usage: "Zendesk token to write tickets/requests"},
					&cli.StringFlag{Name: "zendesk_email", Value: "", EnvVars: []string{envPrefix + "_ZENDESK_EMAIL"}, Usage: "Zendesk email to write tickets/requests"},
					&cli.StringFlag{Name: "zendesk_url", Value: "", EnvVars: []string{envPrefix + "_ZENDESK_URL"}, Usage: "Zendesk url to write tickets/requests"},

					&cli.StringFlag{Name: "ovenmedia_auth_key", Value: "test", EnvVars: []string{envPrefix + "_OVENMEDIA_AUTH_KEY"}, Usage: "Auth key for ovenmedia"},
					&cli.StringFlag{Name: "slack_auth_token", EnvVars: []string{envPrefix + "_SLACK_AUTH_TOKEN"}, Usage: "Slack app token for mod tools"},

					// Crypto signatures for battle histories
					&cli.StringFlag{Name: "private_key_signer_hex", Value: "0x5f3b57101caf01c3d91e50809e70d84fcc404dd108aa8a9aa3e1a6c482267f48", EnvVars: []string{envPrefix + "_PRIVATE_KEY_SIGNER_HEX"}, Usage: "Private key for signing battle records (default is testnet dev private key)"},
					&cli.StringFlag{Name: "ovenmedia_signed_key", Value: "aKq#1kj", EnvVars: []string{envPrefix + "_OVENMEDIA_SIGNED_KEY"}, Usage: "Ovenmedia secret sign key"},

					&cli.StringFlag{Name: "discord_auth_token", Value: "", EnvVars: []string{envPrefix + "_DISCORD_AUTH_TOKEN"}, Usage: "Discord bot auth token"},
					&cli.StringFlag{Name: "discord_app_id", Value: "", EnvVars: []string{envPrefix + "_DISCORD_APP_ID"}, Usage: "Discord bot app id"},
				},
				Usage: "run server",
				Action: func(c *cli.Context) error {
					start := time.Now()

					gameClientMinimumBuildNo := c.Uint64("game_client_minimum_build_no")

					databaseMaxIdleConns := c.Int("database_max_idle_conns")
					databaseMaxOpenConns := c.Int("database_max_open_conns")

					databaseUser := c.String("database_user")
					databasePass := c.String("database_pass")
					databaseHost := c.String("database_host")
					databasePort := c.String("database_port")
					databaseName := c.String("database_name")
					databaseAppName := c.String("database_application_name")

					twilioSid := c.String("twilio_sid")
					twilioApiKey := c.String("twilio_api_key")
					twilioApiSecrete := c.String("twilio_api_secret")
					smsFromNumber := c.String("sms_from_number")
					githubToken := c.String("github_token")

					zendeskToken := c.String("zendesk_token")
					zendeskEmail := c.String("zendesk_email")
					zendeskUrl := c.String("zendesk_url")

					telegramBotToken := c.String("telegram_bot_token")

					stripeSecretKey := c.String("stripe_secret_key")

					passportAddr := c.String("passport_addr")
					passportClientToken := c.String("passport_server_token")

					syncKeycard := c.Bool("sync_keycards")
					keycardCSVPath := c.String("keycard_csv_path")

					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()
					environment := c.String("environment")

					discordAuthToken := c.String("discord_auth_token")
					discordAppID := c.String("discord_app_id")

					replay.OvenMediaAuthKey = c.String("ovenmedia_auth_key")
					voice_chat.VoiceChatSecretKey = c.String("ovenmedia_signed_key")
					slack.ModToolsAppToken = c.String("slack_auth_token")

					server.SetEnv(environment)

					battleArenaAddr := c.String("battle_arena_addr")
					level := c.String("log_level")
					gamelog.New(environment, level)

					// initialise ws package
					ws.Init(&ws.Config{
						Logger:        gamelog.L,
						SkipRateLimit: environment == "staging" || environment == "development",
					})

					tracer.Start(
						tracer.WithEnv(environment),
						tracer.WithService(envPrefix),
						tracer.WithServiceVersion(Version),
						tracer.WithLogger(gamelog.DatadogLog{L: gamelog.L}), // configure before profiler so profiler will use this logger
					)
					defer tracer.Stop()

					// Datadog Tracing an profiling
					if c.Bool("pprof_datadog") {
						// Decode Profile types
						active := c.StringSlice("pprof_datadog_profiles")
						profilers := []profiler.ProfileType{}
						for _, act := range active {
							switch act {
							case profiler.CPUProfile.String():
								gamelog.L.Debug().Msgf("Adding Datadog profiler: %s", profiler.CPUProfile)
								profilers = append(profilers, profiler.CPUProfile)
							case profiler.HeapProfile.String():
								gamelog.L.Debug().Msgf("Adding Datadog profiler: %s", profiler.HeapProfile)
								profilers = append(profilers, profiler.HeapProfile)
							case profiler.BlockProfile.String():
								gamelog.L.Debug().Msgf("Adding Datadog profiler: %s", profiler.BlockProfile)
								profilers = append(profilers, profiler.BlockProfile)
							case profiler.MutexProfile.String():
								gamelog.L.Debug().Msgf("Adding Datadog profiler: %s", profiler.MutexProfile)
								profilers = append(profilers, profiler.MutexProfile)
							case profiler.GoroutineProfile.String():
								gamelog.L.Debug().Msgf("Adding Datadog profiler: %s", profiler.GoroutineProfile)
								profilers = append(profilers, profiler.GoroutineProfile)
							case profiler.MetricsProfile.String():
								gamelog.L.Debug().Msgf("Adding Datadog profiler: %s", profiler.MetricsProfile)
								profilers = append(profilers, profiler.MetricsProfile)
							}
						}
						if environment != "development" {
							err := profiler.Start(
								// Service configuration
								profiler.WithService(envPrefix),
								profiler.WithVersion(Version),
								profiler.WithEnv(environment),
								// This doesn't have a WithLogger option but it can use the tracer logger if tracer is configured first.
								// Profiler configuration
								profiler.WithPeriod(c.Duration("pprof_datadog_interval_sec")*time.Second),
								profiler.CPUDuration(c.Duration("pprof_datadog_duration_sec")*time.Second),
								profiler.WithProfileTypes(
									profilers...,
								),
							)
							if err != nil {
								gamelog.L.Error().Err(err).Msg("Failed to start Datadog Profiler")
							}
							gamelog.L.Info().Strs("with", active).Msg("Starting datadog profiler")
							defer profiler.Stop()
						}

					}

					if gameClientMinimumBuildNo == 0 {
						gamelog.L.Panic().Msg("game_client_minimum_build_no not set or zero value")
					}

					sqlconn, err := gamedb.SqlConnect(
						databaseUser,
						databasePass,
						databaseHost,
						databasePort,
						databaseName,
						databaseAppName,
						Version,
						databaseMaxIdleConns,
						databaseMaxOpenConns,
					)
					if err != nil {
						return terror.Panic(err)
					}
					err = gamedb.New(sqlconn)
					if err != nil {
						return terror.Panic(err)
					}

					u, err := url.Parse(passportAddr)
					if err != nil {
						return terror.Panic(err)
					}

					gamelog.L.Info().Msg("start rpc client")
					rpcClient := xsyn_rpcclient.NewXsynXrpcClient(passportClientToken, u.Hostname(), 10001, 34)

					gamelog.L.Info().Msg("start rpc server")
					rpcServer := &comms.XrpcServer{}

					err = rpcServer.Listen(rpcClient, 11001, 34)
					if err != nil {
						return err
					}

					gamelog.L.Info().Msg("Setting twilio client")
					// initialise smser
					twilio, err := sms.NewTwilio(twilioSid, twilioApiKey, twilioApiSecrete, smsFromNumber, environment)
					if err != nil {
						return terror.Error(err, "SMS init failed")
					}
					gamelog.L.Info().Msgf("twilio took %s", time.Since(start))
					start = time.Now()

					gamelog.L.Info().Msg("Setting up telegram bot")
					// initialise telegram bot
					telebot, err := telegram.NewTelegram(telegramBotToken, environment, func(owner string, success bool) {
						ws.PublishMessage(fmt.Sprintf("/secure/user/%s/telegram_shortcode_register", owner), server.HubKeyTelegramShortcodeRegistered, success)
					})
					if err != nil {
						return terror.Error(err, "Telegram init failed")
					}

					gamelog.L.Info().Msgf("Telegram took %s", time.Since(start))

					// initialise discord bot
					discordBot, err := discord.NewDiscordBot(discordAuthToken, discordAppID, !server.IsDevelopmentEnv())
					if err != nil {
						return terror.Error(err, "Discord init failed")
					}

					start = time.Now()
					// initialise stripe
					stripeClient := &client.API{}
					stripeClient.Init(stripeSecretKey, nil)
					// initialise lingua language detector
					languages := []lingua.Language{
						lingua.English,
						lingua.Tagalog,
					}
					gamelog.L.Info().Msg("Setting new NewLanguageDetectorBuilder")

					if environment != "development" {
						languages = append(languages,
							[]lingua.Language{
								lingua.French,
								lingua.German,
								lingua.Spanish,
								lingua.Italian,
								lingua.Vietnamese,
								lingua.Japanese,
								lingua.Chinese,
								lingua.Russian,
								lingua.Indonesian,
								lingua.Hindi,
								lingua.Portuguese,
								lingua.Dutch,
								lingua.Croatian,
							}...)
					}

					detector := lingua.NewLanguageDetectorBuilder().FromLanguages(languages...).WithPreloadedLanguageModels().Build()
					gamelog.L.Info().Msgf("NewLanguageDetectorBuilder took %s", time.Since(start))

					start = time.Now()
					// initialise profanity manager
					gamelog.L.Info().Msg("Setting up profanity manager")
					pm, err := profanities.NewProfanityManager()
					if err != nil {
						return terror.Error(err, "Profanity manager init failed")
					}
					gamelog.L.Info().Msgf("Profanity manager took %s", time.Since(start))

					start = time.Now()
					// initialise quest manager
					qm, err := quest.New()
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					gamelog.L.Info().Msgf("Quest manager took %s", time.Since(start))

					start = time.Now()
					// initialise battle arena
					gamelog.L.Info().Str("battle_arena_addr", battleArenaAddr).Msg("Setting up battle arena")

					arenaManager, err := battle.NewArenaManager(&battle.Opts{
						Addr:                     battleArenaAddr,
						RPCClient:                rpcClient,
						SMS:                      twilio,
						Telegram:                 telebot,
						GameClientMinimumBuildNo: gameClientMinimumBuildNo,
						QuestManager:             qm,
					})
					if err != nil {
						return terror.Error(err, "Arena Manager init failed")
					}

					gamelog.L.Info().Msgf("Battle arena took %s", time.Since(start))

					start = time.Now()
					staticDataURL := fmt.Sprintf("https://%s@raw.githubusercontent.com/ninja-syndicate/supremacy-static-data", githubToken)

					gamelog.L.Info().Msg("Setting up Zendesk")
					zendesk, err := zendesk.NewZendesk(zendeskToken, zendeskEmail, zendeskUrl, environment)
					if err != nil {
						return terror.Error(err, "Zendesk init failed")
					}
					gamelog.L.Info().Msgf("Zendesk took %s", time.Since(start))

					gamelog.L.Info().Msg("Setting up API")
					api, err := SetupAPI(c, ctx, log_helpers.NamedLogger(gamelog.L, "API"), arenaManager, rpcClient, twilio, telebot, discordBot, zendesk, detector, pm, stripeClient, staticDataURL, qm)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					gamelog.L.Info().Msgf("API took %s", time.Since(start))

					if environment == "production" || environment == "staging" {
						gamelog.L.Info().Msg("Running telegram bot")
						go telebot.RunTelegram(telebot.Bot)
					}

					// we need to update some IDs on passport server, just the once,
					// TODO: After deploying composable migration, talk to vinnie about removing this
					gamelog.L.Info().Msg("Running one off funcs")
					asset.RegisterAllNewAssets(rpcClient)
					gamelog.L.Info().Msgf("RegisterAllNewAssets took %s", time.Since(start))
					start = time.Now()
					UpdateXsynStoreItemTemplates(rpcClient)
					gamelog.L.Info().Msgf("UpdateXsynStoreItemTemplates took %s", time.Since(start))
					start = time.Now()

					if syncKeycard { // TODO: Remove after syncing keycards
						UpdateKeycard(api, rpcClient, keycardCSVPath)
						gamelog.L.Info().Msgf("UpdateKeycard took %s", time.Since(start))
						start = time.Now()
					}
					gamelog.L.Info().Msg("One off funcs finished")

					gamelog.L.Info().Msg("Running asset transfers")
					asset.SyncAssetOwners(rpcClient)
					gamelog.L.Info().Msgf("Asset transfers took %s", time.Since(start))

					// stops all battle replay recordings when server goes down
					go func() {
						stop := make(chan os.Signal)
						signal.Notify(stop, os.Interrupt)
						<-stop
						err = replay.StopAllActiveRecording()
						if err != nil {
							gamelog.L.Error().Err(err).Msg("Failed to stop all active recordings")
						}
						os.Exit(2)
					}()

					gamelog.L.Info().Msg("Running API")
					err = api.Run(ctx)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					log_helpers.TerrorEcho(ctx, err, gamelog.L)
					return nil
				},
			},
			{
				Name:    "sync",
				Aliases: []string{"sy"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "database_user", Value: "gameserver", EnvVars: []string{envPrefix + "_DATABASE_USER", "DATABASE_USER"}, Usage: "The database user"},
					&cli.StringFlag{Name: "database_pass", Value: "dev", EnvVars: []string{envPrefix + "_DATABASE_PASS", "DATABASE_PASS"}, Usage: "The database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{envPrefix + "_DATABASE_HOST", "DATABASE_HOST"}, Usage: "The database host"},
					&cli.StringFlag{Name: "database_port", Value: "5432", EnvVars: []string{envPrefix + "_DATABASE_PORT", "DATABASE_PORT"}, Usage: "The database port"},
					&cli.StringFlag{Name: "database_name", Value: "gameserver", EnvVars: []string{envPrefix + "_DATABASE_NAME", "DATABASE_NAME"}, Usage: "The database name"},
					&cli.StringFlag{Name: "database_application_name", Value: "API Sync", EnvVars: []string{envPrefix + "_DATABASE_APPLICATION_NAME"}, Usage: "Postgres database name"},
					&cli.StringFlag{Name: "static_path", Value: "./synctool/temp-sync/supremacy-static-data/", EnvVars: []string{envPrefix + "_STATIC_PATH"}, Usage: "Static path to file"},
					&cli.IntFlag{Name: "database_max_idle_conns", Value: 40, EnvVars: []string{envPrefix + "_DATABASE_MAX_IDLE_CONNS"}, Usage: "Database max idle conns"},
					&cli.IntFlag{Name: "database_max_open_conns", Value: 50, EnvVars: []string{envPrefix + "_DATABASE_MAX_OPEN_CONNS"}, Usage: "Database max open conns"},
				},
				Usage: "sync static data",
				Action: func(c *cli.Context) error {
					fmt.Println("Running Sync")
					databaseUser := c.String("database_user")
					databasePass := c.String("database_pass")
					databaseHost := c.String("database_host")
					databasePort := c.String("database_port")
					databaseName := c.String("database_name")
					databaseAppName := c.String("database_application_name")
					databaseMaxIdleConns := c.Int("database_max_idle_conns")
					databaseMaxOpenConns := c.Int("database_max_open_conns")

					filePath := c.String("static_path")

					sqlconn, err := gamedb.SqlConnect(
						databaseUser,
						databasePass,
						databaseHost,
						databasePort,
						databaseName,
						databaseAppName,
						Version,
						databaseMaxIdleConns,
						databaseMaxOpenConns,
					)
					if err != nil {
						return terror.Panic(err)
					}

					err = sqlconn.Ping()
					if err != nil {
						return terror.Panic(err, "Failed to ping to DB")
					}

					dt := &synctool.StaticSyncTool{
						DB:       sqlconn,
						FilePath: filePath,
					}

					err = synctool.SyncTool(dt)
					if err != nil {
						return err
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1) // so ci knows it no good
	}
}

func UpdateXsynStoreItemTemplates(pp *xsyn_rpcclient.XsynXrpcClient) {
	updated := db.GetBoolWithDefault("UPDATED_TEMPLATE_ITEMS_IDS", false)
	if !updated {
		var assets []*xsyn_rpcclient.TemplatesToUpdate
		query := `
				SELECT tpo.id AS old_template_id, tpbp.template_id AS new_template_id
				FROM templates_old tpo
				INNER JOIN template_blueprints tpbp ON tpo.blueprint_chassis_id =  tpbp.blueprint_id_old; `
		err := boiler.NewQuery(qm.SQL(query)).Bind(nil, gamedb.StdConn, &assets)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("issue getting template ids")
			return
		}

		err = pp.UpdateStoreItemIDs(assets)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("issue updating template ids on passport")
			return
		}
		gamelog.L.Info().Msg("Successfully updated xsyn store template items")
		db.PutBool("UPDATED_TEMPLATE_ITEMS_IDS", true)
	}

}

type KeyCardUpdate struct {
	PublicAddress string
	BlueprintID   string
}

func UpdateKeycard(api *api.API, pp *xsyn_rpcclient.XsynXrpcClient, filePath string) {
	gamelog.L.Info().Msg("Syncing Keycards with Passport")
	updated := db.GetBoolWithDefault("UPDATED_KEYCARD_ITEMS", false)
	if !updated {
		f, err := os.OpenFile(filePath, os.O_RDONLY, 0755)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("issue updating keycards")
			return
		}

		defer f.Close()

		r := csv.NewReader(f)

		if _, err := r.Read(); err != nil {
			return
		}

		records, err := r.ReadAll()
		if err != nil {
			gamelog.L.Error().Err(err).Msg("issue reading csv")
			return
		}

		var KeyCardUpdates []KeyCardUpdate
		for _, record := range records {
			keyCardUpdate := &KeyCardUpdate{
				PublicAddress: record[0],
				BlueprintID:   record[1],
			}

			KeyCardUpdates = append(KeyCardUpdates, *keyCardUpdate)
		}

		failed := 0
		success := 0

		var keycardAssets xsyn_rpcclient.UpdateUser1155AssetReq
		var keyCardData []xsyn_rpcclient.Supremacy1155Asset
		for i, KeyCardUpdate := range KeyCardUpdates {
			keycard, err := boiler.BlueprintKeycards(boiler.BlueprintKeycardWhere.ID.EQ(KeyCardUpdate.BlueprintID)).One(gamedb.StdConn)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("failed to get keycard blueprint")
				continue
			}

			if i == 0 {
				keycardAssets.PublicAddress = KeyCardUpdate.PublicAddress

				attrValue := "N/A"
				if keycard.Syndicate.Valid {
					attrValue = keycard.Syndicate.String
				}

				keyCardData = append(keyCardData, xsyn_rpcclient.Supremacy1155Asset{
					BlueprintID:    keycard.ID,
					Label:          keycard.Label,
					Description:    keycard.Description,
					CollectionSlug: "supremacy-achievements",
					TokenID:        keycard.KeycardTokenID,
					Count:          1,
					ImageURL:       keycard.ImageURL,
					AnimationURL:   keycard.AnimationURL.String,
					KeycardGroup:   keycard.KeycardGroup,
					Attributes: []xsyn_rpcclient.SupremacyKeycardAttribute{
						xsyn_rpcclient.SupremacyKeycardAttribute{
							TraitType: "Syndicate",
							Value:     attrValue,
						},
					},
				})
				continue
			}

			if KeyCardUpdate.PublicAddress == KeyCardUpdates[i-1].PublicAddress {
				attrValue := "N/A"
				if keycard.Syndicate.Valid {
					attrValue = keycard.Syndicate.String
				}

				keyCardData = append(keyCardData, xsyn_rpcclient.Supremacy1155Asset{
					BlueprintID:    keycard.ID,
					Label:          keycard.Label,
					Description:    keycard.Description,
					CollectionSlug: "supremacy-achievements",
					TokenID:        keycard.KeycardTokenID,
					Count:          1,
					ImageURL:       keycard.ImageURL,
					AnimationURL:   keycard.AnimationURL.String,
					KeycardGroup:   keycard.KeycardGroup,
					Attributes: []xsyn_rpcclient.SupremacyKeycardAttribute{
						xsyn_rpcclient.SupremacyKeycardAttribute{
							TraitType: "Syndicate",
							Value:     attrValue,
						},
					},
				})
				continue
			}

			keycardAssets.AssetData = keyCardData
			resp, err := pp.UpdateKeycardItem(&keycardAssets)
			if err != nil {
				gamelog.L.Error().Err(err).Msg("Failed to update key card item from passport server")
				failed++
				for _, assetData := range keycardAssets.AssetData {
					failedSync := &boiler.FailedPlayerKeycardsSync{
						PublicAddress:      keycardAssets.PublicAddress,
						BlueprintKeycardID: assetData.BlueprintID,
						Count:              assetData.Count,
						Reason:             "Passport RPC Error",
					}

					if err := failedSync.Insert(gamedb.StdConn, boil.Infer()); err != nil {
						gamelog.L.Error().Str("public_address", keycardAssets.PublicAddress).Str("blueprint_id", assetData.BlueprintID).Msg("Failed to insert failed sync item")
						continue
					}
				}
				continue
			}
			factionID := uuid.Nil
			if resp.FactionID.Valid {
				factionID = uuid.Must(uuid.FromString(resp.FactionID.String))
			}

			err = api.UpsertPlayer(resp.UserID, null.StringFrom(resp.Username), resp.PublicAddress, null.StringFrom(factionID.String()), nil, null.Bool{})
			if err != nil {
				gamelog.L.Error().Err(err).Str("public_address", keycardAssets.PublicAddress).Str("factionID", factionID.String()).Str("resp.Username", resp.Username).Str("resp.UserID", resp.UserID).Msg("failed to register player")
			}

			for _, assetData := range keyCardData {
				playerKeycard := boiler.PlayerKeycard{
					PlayerID:           resp.UserID,
					BlueprintKeycardID: assetData.BlueprintID,
					Count:              assetData.Count,
				}

				err := playerKeycard.Insert(gamedb.StdConn, boil.Infer())
				if err != nil {
					failed++
					gamelog.L.Error().Interface("PlayerKeycards", playerKeycard).Err(err).Msg("failed to insert new player keycard")
					failedSync := &boiler.FailedPlayerKeycardsSync{
						PublicAddress:      keycardAssets.PublicAddress,
						BlueprintKeycardID: assetData.BlueprintID,
						Count:              assetData.Count,
						Reason:             fmt.Sprintf("Gameserver Insert Error: %s", err.Error()),
					}

					if failedSync.Insert(gamedb.StdConn, boil.Infer()) != nil {
						gamelog.L.Error().Str("public_address", keycardAssets.PublicAddress).Str("blueprint_id", assetData.BlueprintID).Msg("Failed to insert failed sync item")
						continue
					}
					continue
				}
				success++
			}

			keyCardData = nil

			keycardAssets.PublicAddress = KeyCardUpdate.PublicAddress
			attrValue := "N/A"
			if keycard.Syndicate.Valid {
				attrValue = keycard.Syndicate.String
			}

			keyCardData = append(keyCardData, xsyn_rpcclient.Supremacy1155Asset{
				BlueprintID:    keycard.ID,
				Label:          keycard.Label,
				Description:    keycard.Description,
				CollectionSlug: "supremacy-achievements",
				TokenID:        keycard.KeycardTokenID,
				Count:          1,
				ImageURL:       keycard.ImageURL,
				AnimationURL:   keycard.AnimationURL.String,
				KeycardGroup:   keycard.KeycardGroup,
				Attributes: []xsyn_rpcclient.SupremacyKeycardAttribute{
					xsyn_rpcclient.SupremacyKeycardAttribute{
						TraitType: "Syndicate",
						Value:     attrValue,
					},
				},
			})

		}

		db.PutBool("UPDATED_KEYCARD_ITEMS", true)

		gamelog.L.Info().Int("Success", success).Int("Failed", failed).Msg("Completed importing text game non-minted assets")
	}

}

func SetupAPI(
	ctxCLI *cli.Context,
	ctx context.Context,
	log *zerolog.Logger,
	arenaManager *battle.ArenaManager,
	passport *xsyn_rpcclient.XsynXrpcClient,
	sms server.SMS,
	telegram server.Telegram,
	discord *discord.DiscordSession,
	zendesk *zendesk.Zendesk,
	languageDetector lingua.LanguageDetector,
	pm *profanities.ProfanityManager,
	stripeClient *client.API,
	staticSyncURL string,
	questManager *quest.System,
) (*api.API, error) {
	environment := ctxCLI.String("environment")
	sentryDSNBackend := ctxCLI.String("sentry_dsn_backend")
	sentryServerName := ctxCLI.String("sentry_server_name")
	sentryTraceRate := ctxCLI.Float64("sentry_sample_rate")
	sentryRelease := fmt.Sprintf("%s@%s", SentryReleasePrefix, Version)
	stripeWebhookSecret := ctxCLI.String("stripe_webhook_secret")
	err := log_helpers.SentryInit(sentryDSNBackend, sentryServerName, sentryRelease, environment, sentryTraceRate, log)
	switch errors.Unwrap(err) {
	case log_helpers.ErrSentryInitEnvironment:
		return nil, terror.Error(err, fmt.Sprintf("got environment %s", environment))
	case log_helpers.ErrSentryInitDSN, log_helpers.ErrSentryInitVersion:
		if terror.GetLevel(err) == terror.ErrLevelPanic {
			// if the level is panic then in a prod environment
			// so keep panicing
			return nil, terror.Panic(err)
		}
	default:
		if err != nil {
			return nil, err
		}
	}

	jwtKey := ctxCLI.String("jwt_key")
	jwtKeyByteArray, err := base64.StdEncoding.DecodeString(jwtKey)
	if err != nil {
		return nil, err
	}

	apiAddr := ctxCLI.String("api_addr")

	config := &server.Config{
		CookieSecure:          ctxCLI.Bool("cookie_secure"),
		EncryptTokens:         ctxCLI.Bool("jwt_encrypt"),
		EncryptTokensKey:      ctxCLI.String("jwt_encrypt_key"),
		TokenExpirationDays:   ctxCLI.Int("jwt_expiry_days"),
		TwitchUIHostURL:       ctxCLI.String("twitch_ui_web_host_url"),
		ServerStreamKey:       ctxCLI.String("server_stream_key"),
		PassportWebhookSecret: ctxCLI.String("passport_webhook_secret"),
		CookieKey:             ctxCLI.String("cookie_key"),
		JwtKey:                jwtKeyByteArray,
		Environment:           environment,
		Address:               apiAddr,
		AuthCallbackURL:       ctxCLI.String("auth_callback_url"),
		AuthHangarCallbackURL: ctxCLI.String("auth_hangar_callback_url"),
		CaptchaSiteKey:        ctxCLI.String("captcha_site_key"),
		CaptchaSecret:         ctxCLI.String("captcha_secret"),
	}

	syncConfig := &synctool.StaticSyncTool{
		FilePath: staticSyncURL,
	}

	// HTML Sanitizer
	HTMLSanitizePolicy := bluemonday.UGCPolicy()
	HTMLSanitizePolicy.AllowAttrs("class").OnElements("img", "table", "tr", "td", "p")

	// API Server
	privateKeySignerHex := ctxCLI.String("private_key_signer_hex")
	serverAPI, err := api.NewAPI(ctx, arenaManager, passport, HTMLSanitizePolicy, stripeClient, stripeWebhookSecret, config, sms, telegram, discord, zendesk, languageDetector, pm, syncConfig, questManager, privateKeySignerHex)
	if err != nil {
		return nil, err
	}
	return serverAPI, nil
}
