package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	cli "github.com/urfave/cli/v2"
)

// Variable passed in at compile time using `-ldflags`
var (
	Version          string // -X main.Version=$(git describe --tags --abbrev=0)
	GitHash          string // -X main.GitHash=$(git rev-parse HEAD)
	GitBranch        string // -X main.GitBranch=$(git rev-parse --abbrev-ref HEAD)
	BuildDate        string // -X main.BuildDate=$(date -u +%Y%m%d%H%M%S)
	UnCommittedFiles string // -X main.UnCommittedFiles=$(git status --porcelain | wc -l)"
)

const SentryReleasePrefix = "supremacy-gameserver-seed"
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
				Name: "db",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "database_user", Value: "gameserver", EnvVars: []string{envPrefix + "_DATABASE_USER"}, Usage: "The database user"},
					&cli.StringFlag{Name: "database_pass", Value: "dev", EnvVars: []string{envPrefix + "_DATABASE_PASS"}, Usage: "The database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{envPrefix + "_DATABASE_HOST"}, Usage: "The database host"},
					&cli.StringFlag{Name: "database_port", Value: "5437", EnvVars: []string{envPrefix + "_DATABASE_PORT"}, Usage: "The database port"},
					&cli.StringFlag{Name: "database_name", Value: "gameserver", EnvVars: []string{envPrefix + "_DATABASE_NAME"}, Usage: "The database name"},
					&cli.StringFlag{Name: "database_application_name", Value: "API Server", EnvVars: []string{envPrefix + "_DATABASE_APPLICATION_NAME"}, Usage: "Postgres database name"},
					&cli.BoolFlag{Name: "assets", Value: false, Usage: "Whether to seed assets only"},

					//&cli.BoolFlag{Name: "seed", EnvVars: []string{"DB_SEED"}, Usage: "seed the database"},
				},
				Usage: "seed the database",
				Action: func(c *cli.Context) error {
					databaseUser := c.String("database_user")
					databasePass := c.String("database_pass")
					databaseHost := c.String("database_host")
					databasePort := c.String("database_port")
					databaseName := c.String("database_name")
					databaseAppName := c.String("database_application_name")

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
						return terror.Error(err)
					}

					seeder := NewSeeder(pgxconn)
					if c.Bool("assets") {
						return seeder.RunAssets()
					}
					return seeder.Run()
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
