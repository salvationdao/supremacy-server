package main

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/url"
	"server"
	"server/api"
	"server/battle"
	"server/comms"
	"server/db"
	"server/db/boiler"
	"server/gamedb"
	"server/gamelog"
	"server/sms"
	"server/telegram"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"server/rpcclient"

	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
	zerologger "github.com/ninja-syndicate/hub/ext/zerolog"
	"github.com/pemistahl/lingua-go"
	"nhooyr.io/websocket"

	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/log_helpers"

	_ "net/http/pprof"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"

	"context"
	"os"

	"github.com/urfave/cli/v2"
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
					&cli.StringFlag{Name: "google_client_id", Value: "", EnvVars: []string{envPrefix + "_GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_ID"}, Usage: "Google Client ID for OAuth functionaility."},

					// SMS stuff
					&cli.StringFlag{Name: "twilio_sid", Value: "", EnvVars: []string{envPrefix + "_TWILIO_ACCOUNT_SID"}, Usage: "Twilio account sid"},
					&cli.StringFlag{Name: "twilio_api_key", Value: "", EnvVars: []string{envPrefix + "_TWILIO_API_KEY"}, Usage: "Twilio api key"},
					&cli.StringFlag{Name: "twilio_api_secret", Value: "", EnvVars: []string{envPrefix + "_TWILIO_API_SECRET"}, Usage: "Twilio api secret"},
					&cli.StringFlag{Name: "sms_from_number", Value: "", EnvVars: []string{envPrefix + "_SMS_FROM_NUMBER"}, Usage: "Number to send SMS from"},

					// telegram bot token
					&cli.StringFlag{Name: "telegram_bot_token", Value: "", EnvVars: []string{envPrefix + "_TELEGRAM_BOT_TOKEN"}, Usage: "telegram bot token"},

					// TODO: clear up token
					&cli.BoolFlag{Name: "jwt_encrypt", Value: true, EnvVars: []string{envPrefix + "_JWT_ENCRYPT", "JWT_ENCRYPT"}, Usage: "set if to encrypt jwt tokens or not"},
					&cli.StringFlag{Name: "jwt_encrypt_key", Value: "ITF1vauAxvJlF0PLNY9btOO9ZzbUmc6X", EnvVars: []string{envPrefix + "_JWT_KEY", "JWT_KEY"}, Usage: "supports key sizes of 16, 24 or 32 bytes"},
					&cli.IntFlag{Name: "jwt_expiry_days", Value: 1, EnvVars: []string{envPrefix + "_JWT_EXPIRY_DAYS", "JWT_EXPIRY_DAYS"}, Usage: "expiry days for auth tokens"},
					&cli.StringFlag{Name: "jwt_key", Value: "9a5b8421bbe14e5a904cfd150a9951d3", EnvVars: []string{"STREAM_SITE_JWT_KEY"}, Usage: "JWT Key for signing token on stream site"},

					&cli.StringFlag{Name: "passport_server_token", Value: "e79422b7-7bfe-4463-897b-a1d22bf2e0bc", EnvVars: []string{envPrefix + "_PASSPORT_TOKEN"}, Usage: "Token to auth to passport server"},
					&cli.StringFlag{Name: "server_stream_key", Value: "6c7b4a82-7797-4847-836e-978399830878", EnvVars: []string{envPrefix + "_SERVER_STREAM_KEY"}, Usage: "Authorization key to crud servers"},
					&cli.StringFlag{Name: "passport_webhook_secret", Value: "e1BD3FF270804c6a9edJDzzDks87a8a4fde15c7=", EnvVars: []string{"PASSPORT_WEBHOOK_SECRET"}, Usage: "Authorization key to passport webhook"},

					&cli.IntFlag{Name: "database_max_pool_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_POOL_CONNS"}, Usage: "Database max pool conns"},
					&cli.IntFlag{Name: "database_max_idle_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_IDLE_CONNS"}, Usage: "Database max idle conns"},
					&cli.IntFlag{Name: "database_max_open_conns", Value: 2000, EnvVars: []string{envPrefix + "_DATABASE_MAX_OPEN_CONNS"}, Usage: "Database max open conns"},

					&cli.BoolFlag{Name: "pprof_datadog", Value: true, EnvVars: []string{envPrefix + "_PPROF_DATADOG"}, Usage: "Use datadog pprof to collect debug info"},
					&cli.StringSliceFlag{Name: "pprof_datadog_profiles", Value: cli.NewStringSlice("cpu", "heap"), EnvVars: []string{envPrefix + "_PPROF_DATADOG_PROFILES"}, Usage: "Comma seprated list of profiles to collect. Options: cpu,heap,block,mutex,goroutine,metrics"},
					&cli.DurationFlag{Name: "pprof_datadog_interval_sec", Value: 60, EnvVars: []string{envPrefix + "_PPROF_DATADOG_INTERVAL_SEC"}, Usage: "Specifies the period at which profiles will be collected"},
					&cli.DurationFlag{Name: "pprof_datadog_duration_sec", Value: 60, EnvVars: []string{envPrefix + "_PPROF_DATADOG_DURATION_SEC"}, Usage: "Specifies the length of the CPU profile snapshot"},
				},
				Usage: "run server",
				Action: func(c *cli.Context) error {
					gameClientMinimumBuildNo := c.Uint64("game_client_minimum_build_no")

					databaseMaxPoolConns := c.Int("database_max_pool_conns")
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

					telegramBotToken := c.String("telegram_bot_token")

					passportAddr := c.String("passport_addr")
					passportClientToken := c.String("passport_server_token")

					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()
					environment := c.String("environment")

					server.SetEnv(environment)

					battleArenaAddr := c.String("battle_arena_addr")
					level := c.String("log_level")
					gamelog.New(environment, level)

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

					pgxconn, err := pgxconnect(
						databaseUser,
						databasePass,
						databaseHost,
						databasePort,
						databaseName,
						databaseAppName,
						Version,
						databaseMaxPoolConns,
					)
					if err != nil {
						return terror.Panic(err)
					}
					sqlconn, err := sqlConnect(
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
					err = gamedb.New(pgxconn, sqlconn)
					if err != nil {
						return terror.Panic(err)
					}

					u, err := url.Parse(passportAddr)
					if err != nil {
						return terror.Panic(err)
					}
					hostname := u.Hostname()
					rpcAddrs := []string{
						fmt.Sprintf("%s:10001", hostname),
						fmt.Sprintf("%s:10002", hostname),
						fmt.Sprintf("%s:10003", hostname),
						fmt.Sprintf("%s:10004", hostname),
						fmt.Sprintf("%s:10005", hostname),
						fmt.Sprintf("%s:10006", hostname),
						fmt.Sprintf("%s:10007", hostname),
						fmt.Sprintf("%s:10008", hostname),
						fmt.Sprintf("%s:10009", hostname),
						fmt.Sprintf("%s:10010", hostname),
						fmt.Sprintf("%s:10011", hostname),
						fmt.Sprintf("%s:10012", hostname),
						fmt.Sprintf("%s:10013", hostname),
						fmt.Sprintf("%s:10014", hostname),
						fmt.Sprintf("%s:10015", hostname),
						fmt.Sprintf("%s:10016", hostname),
						fmt.Sprintf("%s:10017", hostname),
						fmt.Sprintf("%s:10018", hostname),
						fmt.Sprintf("%s:10019", hostname),
						fmt.Sprintf("%s:10020", hostname),
						fmt.Sprintf("%s:10021", hostname),
						fmt.Sprintf("%s:10022", hostname),
						fmt.Sprintf("%s:10023", hostname),
						fmt.Sprintf("%s:10024", hostname),
						fmt.Sprintf("%s:10025", hostname),
						fmt.Sprintf("%s:10026", hostname),
						fmt.Sprintf("%s:10027", hostname),
						fmt.Sprintf("%s:10028", hostname),
						fmt.Sprintf("%s:10029", hostname),
						fmt.Sprintf("%s:10030", hostname),
						fmt.Sprintf("%s:10031", hostname),
						fmt.Sprintf("%s:10032", hostname),
						fmt.Sprintf("%s:10033", hostname),
						fmt.Sprintf("%s:10034", hostname),
						fmt.Sprintf("%s:10035", hostname),
					}
					gamelog.L.Info().Msg("start rpc client")
					rpcClient := rpcclient.NewPassportXrpcClient(passportClientToken, rpcAddrs)

					gamelog.L.Info().Msg("start rpc server")
					rpcServer := &comms.XrpcServer{}

					err = rpcServer.Listen(
						rpcClient,
						":11001",
						":11002",
						":11003",
						":11004",
						":11005",
						":11006",
						":11007",
						":11008",
						":11009",
						":11010",
						":11011",
						":11012",
						":11013",
						":11014",
						":11015",
						":11016",
						":11017",
						":11018",
						":11019",
						":11020",
						":11021",
						":11022",
						":11023",
						":11024",
						":11025",
						":11026",
						":11027",
						":11028",
						":11029",
						":11030",
						":11031",
						":11032",
						":11033",
						":11034",
						":11035",
					)
					if err != nil {
						return terror.Error(err)
					}

					gamelog.L.Info().Str("battle_arena_addr", battleArenaAddr).Msg("Setting up battle arena client")

					// initialise smser
					twilio, err := sms.NewTwilio(twilioSid, twilioApiKey, twilioApiSecrete, smsFromNumber, environment)
					if err != nil {
						return terror.Error(err, "SMS init failed")
					}

					// initialise message bus
					messageBus := messagebus.NewMessageBus(log_helpers.NamedLogger(gamelog.L, "message_bus"))
					gsHub := hub.New(&hub.Config{
						Log:            zerologger.New(*log_helpers.NamedLogger(gamelog.L, "hub library")),
						LoggingEnabled: false,
						WelcomeMsg: &hub.WelcomeMsg{
							Key:     "WELCOME",
							Payload: nil,
						},
						AcceptOptions: &websocket.AcceptOptions{
							InsecureSkipVerify: true, // TODO: set this depending on environment
							OriginPatterns:     []string{"*"},
						},
						ClientOfflineFn: func(cl *hub.Client) {
							messageBus.UnsubAll(cl)
						},
						Tracer: DatadogTracer.New(),
					})

					// initialise telegram bot
					telebot, err := telegram.NewTelegram(telegramBotToken, environment, func(owner string, success bool) {
						go messageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", telegram.HubKeyTelegramShortcodeRegistered, owner)), success)
					})
					if err != nil {
						return terror.Error(err, "Telegram init failed")
					}

					//initialize lingua language detector
					languages := []lingua.Language{
						lingua.English,
						lingua.French,
						lingua.German,
						lingua.Spanish,
						lingua.Italian,
						lingua.Tagalog,
						lingua.Vietnamese,
						lingua.Japanese,
						lingua.Chinese,
						lingua.Russian,
						lingua.Indonesian,
						lingua.Hindi,
						lingua.Portuguese,
						lingua.Dutch,
						lingua.Croatian,
					}
					detector := lingua.NewLanguageDetectorBuilder().FromLanguages(languages...).WithPreloadedLanguageModels().Build()

					gamelog.L.Info().Str("battle_arena_addr", battleArenaAddr).Msg("Set up hub")

					ba := battle.NewArena(&battle.Opts{
						Addr:                     battleArenaAddr,
						Conn:                     pgxconn,
						MessageBus:               messageBus,
						Hub:                      gsHub,
						RPCClient:                rpcClient,
						SMS:                      twilio,
						Telegram:                 telebot,
						GameClientMinimumBuildNo: gameClientMinimumBuildNo,
					})
					gamelog.L.Info().Str("battle_arena_addr", battleArenaAddr).Msg("set up arena")
					gamelog.L.Info().Msg("Setting up webhook rest API")
					api, err := SetupAPI(c, ctx, log_helpers.NamedLogger(gamelog.L, "API"), ba, pgxconn, rpcClient, messageBus, gsHub, twilio, telebot, detector)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					if environment == "production" || environment == "staging" {
						gamelog.L.Info().Msg("Running telegram bot")
						go telebot.RunTelegram(telebot.Bot)
					}

					// we need to update some IDs on passport server, just the once,
					// TODO: After deploying composable migration, talk to vinnie about removing this
					UpdatePurchaseItems(rpcClient)

					gamelog.L.Info().Msg("Running webhook rest API")
					err = api.Run(ctx)
					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
					log_helpers.TerrorEcho(ctx, err, gamelog.L)
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

func UpdatePurchaseItems(pp *rpcclient.PassportXrpcClient) {
	updated := db.GetBoolWithDefault("UPDATED_PURCHASED_ITEMS_IDS", false)
	if !updated {
		var list []*rpcclient.UpdateAssetIDReq
		// get all the old ids and new ids
		err := boiler.NewQuery(
			qm.SQL(`SELECT mo.chassis_id as asset_ID, mo.id as old_asset_ID FROM mechs_old mo`),
		).Bind(nil, gamedb.StdConn, &list)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("issue getting mech ids")
			return
		}

		err = pp.UpdateAssetsID(list)
		if err != nil {
			gamelog.L.Error().Err(err).Msg("issue updating getting mech ids")
			return
		}
		db.PutBool("UPDATED_PURCHASED_ITEMS_IDS", true)
	}
}

func SetupAPI(ctxCLI *cli.Context, ctx context.Context, log *zerolog.Logger, battleArenaClient *battle.Arena, conn *pgxpool.Pool, passport *rpcclient.PassportXrpcClient, messageBus *messagebus.MessageBus, gsHub *hub.Hub, sms server.SMS, telegram server.Telegram, languageDetector lingua.LanguageDetector) (*api.API, error) {
	environment := ctxCLI.String("environment")
	sentryDSNBackend := ctxCLI.String("sentry_dsn_backend")
	sentryServerName := ctxCLI.String("sentry_server_name")
	sentryTraceRate := ctxCLI.Float64("sentry_sample_rate")
	sentryRelease := fmt.Sprintf("%s@%s", SentryReleasePrefix, Version)
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
			return nil, terror.Error(err)
		}
	}

	jwtKey := ctxCLI.String("jwt_key")
	jwtKeyByteArray, err := base64.StdEncoding.DecodeString(jwtKey)
	if err != nil {
		return nil, terror.Error(err)
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
		JwtKey:                jwtKeyByteArray,
		Environment:           environment,
		Address:               apiAddr,
	}

	// HTML Sanitizer
	HTMLSanitizePolicy := bluemonday.UGCPolicy()
	HTMLSanitizePolicy.AllowAttrs("class").OnElements("img", "table", "tr", "td", "p")

	// API Server
	serverAPI := api.NewAPI(ctx, battleArenaClient, passport, HTMLSanitizePolicy, config, messageBus, gsHub, sms, telegram, languageDetector)
	return serverAPI, nil
}

func pgxconnect(
	DatabaseUser string,
	DatabasePass string,
	DatabaseHost string,
	DatabasePort string,
	DatabaseName string,
	DatabaseApplicationName string,
	APIVersion string,
	maxPoolConns int,
) (*pgxpool.Pool, error) {
	params := url.Values{}
	params.Add("sslmode", "disable")
	if DatabaseApplicationName != "" {
		params.Add("application_name", fmt.Sprintf("%s %s", DatabaseApplicationName, APIVersion))
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		DatabaseUser,
		DatabasePass,
		DatabaseHost,
		DatabasePort,
		DatabaseName,
		params.Encode(),
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, terror.Panic(err, "could not initialise database")
	}

	poolConfig.ConnConfig.LogLevel = pgx.LogLevelTrace

	poolConfig.MaxConns = int32(maxPoolConns)

	ctx := context.Background()
	conn, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, terror.Panic(err, "could not initialise database")
	}

	return conn, nil
}

func sqlConnect(
	databaseTxUser string,
	databaseTxPass string,
	databaseHost string,
	databasePort string,
	databaseName string,
	DatabaseApplicationName string,
	APIVersion string,
	maxIdle int,
	maxOpen int,
) (*sql.DB, error) {
	params := url.Values{}
	params.Add("sslmode", "disable")
	if DatabaseApplicationName != "" {
		params.Add("application_name", fmt.Sprintf("%s %s", DatabaseApplicationName, APIVersion))
	}
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		databaseTxUser,
		databaseTxPass,
		databaseHost,
		databasePort,
		databaseName,
		params.Encode(),
	)
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	conn := stdlib.OpenDB(*cfg)
	if err != nil {
		return nil, err
	}
	conn.SetMaxIdleConns(maxIdle)
	conn.SetMaxOpenConns(maxOpen)
	return conn, nil

}
