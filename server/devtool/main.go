package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"log"
	"net/url"
	"os"
)

type DevTool struct {
	db *sql.DB
}

func main() {
	if os.Getenv("GAMESERVER_ENVIRONMENT") == "production" {
		log.Fatal("Only works in dev and staging environment")
	}

	syncMech := flag.Bool("sync_mech", false, "Sync mech skins and models with staging data")

	params := url.Values{}
	params.Add("sslmode", "disable")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		"gameserver",
		"dev",
		"localhost",
		"5437",
		"gameserver",
		params.Encode(),
	)
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		log.Fatal(err)
	}
	conn := stdlib.OpenDB(*cfg)
	if err != nil {
		log.Fatal(err)
	}

	newDevTool := DevTool{db: conn}

	if syncMech != nil && *syncMech {
		newDevTool.SyncMechs()
	}

}
