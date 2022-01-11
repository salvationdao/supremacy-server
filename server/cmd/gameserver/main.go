package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/url"
	"server"
	"server/api"
	"server/battle_arena"
	"server/passport"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"

	"github.com/microcosm-cc/bluemonday"

	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"

	"context"
	"os"

	"github.com/oklog/run"
	"github.com/urfave/cli/v2"
)

// Version build Version
const Version = "v0.1.0"

// SentryVersion passed in using
//   ```sh
//   go build -ldflags "-X main.SentryVersion=" main.go
//   ```
var SentryVersion string

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
				Name:    "serve",
				Aliases: []string{"s"},
				Flags: []cli.Flag{
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
					&cli.StringFlag{Name: "passport_client_id", Value: "gameserver", EnvVars: []string{envPrefix + "_PASSPORT_CLIENT_ID_ADDR", "PASSPORT_CLIENT_ID_ADDR"}, Usage: "game server client ID to auth with passport"},
					&cli.StringFlag{Name: "passport_client_secret", Value: "noidea", EnvVars: []string{envPrefix + "_PASSPORT_CLIENT_SECRET_ADDR", "PASSPORT_CLIENT_SECRET_ADDR"}, Usage: "game server client secret to auth with passport"},

					&cli.StringFlag{Name: "api_addr", Value: ":8084", EnvVars: []string{envPrefix + "_API_ADDR"}, Usage: ":port to run the API"},
					&cli.StringFlag{Name: "rootpath", Value: "../web/build", EnvVars: []string{envPrefix + "_ROOTPATH"}, Usage: "folder path of index.html"},
					&cli.StringFlag{Name: "userauth_jwtsecret", Value: "872ab3df-d7c7-4eb6-a052-4146d0f4dd15", EnvVars: []string{envPrefix + "_USERAUTH_JWTSECRET"}, Usage: "JWT secret"},

					&cli.BoolFlag{Name: "cookie_secure", Value: true, EnvVars: []string{envPrefix + "_COOKIE_SECURE", "COOKIE_SECURE"}, Usage: "set cookie secure"},
					&cli.StringFlag{Name: "google_client_id", Value: "", EnvVars: []string{envPrefix + "_GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_ID"}, Usage: "Google Client ID for OAuth functionaility."},

					// Twitch server stuff
					&cli.StringFlag{Name: "twitch_extension_secret", Value: "", EnvVars: []string{envPrefix + "_TWITCH_EXTENSION_SECRET", "_TWITCH_EXTENSION_SECRET"}, Usage: "Twitch Extension Secret for verifying tokens sent with requests"},

					&cli.BoolFlag{Name: "jwt_encrypt", Value: true, EnvVars: []string{envPrefix + "_JWT_ENCRYPT", "JWT_ENCRYPT"}, Usage: "set if to encrypt jwt tokens or not"},
					&cli.StringFlag{Name: "jwt_encrypt_key", Value: "ITF1vauAxvJlF0PLNY9btOO9ZzbUmc6X", EnvVars: []string{envPrefix + "_JWT_KEY", "JWT_KEY"}, Usage: "supports key sizes of 16, 24 or 32 bytes"},
					&cli.IntFlag{Name: "jwt_expiry_days", Value: 1, EnvVars: []string{envPrefix + "_JWT_EXPIRY_DAYS", "JWT_EXPIRY_DAYS"}, Usage: "expiry days for auth tokens"},
				},
				Usage: "run server",
				Action: func(c *cli.Context) error {
					databaseUser := c.String("database_user")
					databasePass := c.String("database_pass")
					databaseHost := c.String("database_host")
					databasePort := c.String("database_port")
					databaseName := c.String("database_name")
					databaseAppName := c.String("database_application_name")

					passportAddr := c.String("passport_addr")
					passportClientID := c.String("passport_client_id")
					passportClientSecret := c.String("passport_client_secret")

					ctx, cancel := context.WithCancel(c.Context)
					environment := c.String("environment")
					battleArenaAddr := c.String("battle_arena_addr")
					level := c.String("log_level")
					logger := log_helpers.LoggerInitZero(environment, level)
					logger.Info().Msg("zerolog initialised")

					pgxconn, err := pgxconnect(
						databaseUser,
						databasePass,
						databaseHost,
						databasePort,
						databaseName,
						databaseAppName,
						Version,
					)
					if err != nil {
						cancel()
						return terror.Panic(err)
					}

					pp, err := passport.NewPassport(ctx, log_helpers.NamedLogger(logger, "passport"),
						passportAddr,
						passportClientID,
						passportClientSecret,
					)
					if err != nil {
						logger.Err(err).Msgf("failed to create passport connection")
						//cancel()
						//return terror.Panic(err)
					}

					battleArenaClient := battle_arena.NewBattleArenaClient(ctx, log_helpers.NamedLogger(logger, "BattleArena"), pgxconn, pp, battleArenaAddr)

					g := &run.Group{}
					// Listen for os.interrupt
					g.Add(run.SignalHandler(ctx, os.Interrupt))

					// Connect to passport
					g.Add(func() error { return pp.Connect(ctx) }, func(err error) {
						cancel()
						panic(err)
					})

					// Start Gameserver - Gameclient server
					g.Add(func() error { return battleArenaClient.Serve(ctx) }, func(err error) {
						cancel()
						panic(err)
					})

					// Start API/Client server
					g.Add(func() error {
						return ServeFunc(c, ctx, log_helpers.NamedLogger(logger, "API"), battleArenaClient, pgxconn, pp)
					}, func(err error) {
						cancel()
						panic(err)
					})

					err = g.Run()
					if errors.Is(err, run.SignalError{Signal: os.Interrupt}) {
						err = terror.Warn(err)
					}
					log_helpers.TerrorEcho(ctx, err, logger)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func ServeFunc(ctxCLI *cli.Context, ctx context.Context, log *zerolog.Logger, battleArenaClient *battle_arena.BattleArena, conn *pgxpool.Pool, passport *passport.Passport) error {
	environment := ctxCLI.String("environment")
	sentryDSNBackend := ctxCLI.String("sentry_dsn_backend")
	sentryServerName := ctxCLI.String("sentry_server_name")
	sentryTraceRate := ctxCLI.Float64("sentry_sample_rate")
	err := log_helpers.SentryInit(sentryDSNBackend, sentryServerName, SentryVersion, environment, sentryTraceRate, log)
	switch errors.Unwrap(err) {
	case log_helpers.ErrSentryInitEnvironment:
		return terror.Error(err, fmt.Sprintf("got environment %s", environment))
	case log_helpers.ErrSentryInitDSN, log_helpers.ErrSentryInitVersion:
		if terror.GetLevel(err) == terror.ErrLevelPanic {
			// if the level is panic then in a prod environment
			// so keep panicing
			return terror.Panic(err)
		}
	default:
		if err != nil {
			return terror.Error(err)
		}
	}

	apiAddr := ctxCLI.String("api_addr")

	config := &server.Config{
		CookieSecure:        ctxCLI.Bool("cookie_secure"),
		EncryptTokens:       ctxCLI.Bool("jwt_encrypt"),
		EncryptTokensKey:    ctxCLI.String("jwt_encrypt_key"),
		TokenExpirationDays: ctxCLI.Int("jwt_expiry_days"),
	}

	twitchExtensionSecret := ctxCLI.String("twitch_extension_secret")
	if twitchExtensionSecret == "" {
		return fmt.Errorf("missing twitch extension secret")
	}
	secret, err := base64.StdEncoding.DecodeString(twitchExtensionSecret)
	if err != nil {
		return terror.Error(err, "Failed to decode twitch extension secret")
	}

	// HTML Sanitizer
	HTMLSanitizePolicy := bluemonday.UGCPolicy()
	HTMLSanitizePolicy.AllowAttrs("class").OnElements("img", "table", "tr", "td", "p")

	// API Server
	ctx, cancelOnPanic := context.WithCancel(ctx)
	serverAPI := api.NewAPI(log, battleArenaClient, passport, cancelOnPanic, apiAddr, HTMLSanitizePolicy, conn, secret, config)
	return serverAPI.Run(ctx)
}

func pgxconnect(
	DatabaseUser string,
	DatabasePass string,
	DatabaseHost string,
	DatabasePort string,
	DatabaseName string,
	DatabaseApplicationName string,
	APIVersion string,
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

	ctx := context.Background()
	conn, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, terror.Panic(err, "could not initialise database")
	}

	return conn, nil
}
