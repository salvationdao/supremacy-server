package db_test

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/ory/dockertest/v3/docker"

	"github.com/ninja-software/terror/v2"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/ory/dockertest/v3"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

var conn *pgxpool.Pool

//go:embed migrations
var migrations embed.FS

func TestMain(m *testing.M) {
	fmt.Println("Spinning up docker container for postgres...")

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	user := "test"
	password := "dev"
	dbName := "test"

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13-alpine",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	resource.Expire(60) // Tell docker to hard kill the container in 60 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		ctx := context.Background()
		connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			user,
			password,
			"localhost",
			resource.GetPort("5432/tcp"),
			dbName,
		)

		pgxPoolConfig, err := pgxpool.ParseConfig(connString)
		if err != nil {
			return terror.Error(err, "")
		}

		pgxPoolConfig.ConnConfig.LogLevel = pgx.LogLevelTrace

		conn, err = pgxpool.ConnectConfig(ctx, pgxPoolConfig)
		if err != nil {
			return terror.Error(err, "")
		}

		fmt.Println("Running Migration...")

		source, err := httpfs.New(http.FS(migrations), "migrations")
		if err != nil {
			log.Fatal(err)
		}

		mig, err := migrate.NewWithSourceInstance("embed", source, connString)
		if err != nil {
			log.Fatal(err)
		}
		if err := mig.Up(); err != nil {
			log.Fatal(err)
		}
		source.Close()

		fmt.Println("Postgres Ready.")

		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	fmt.Println("Running tests...")
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

// dbFlush clears all tables rows without deleting the schema
func dbFlush(ctx context.Context) {
	q := `
    do
    $$
      declare
        l_stmt text;
      begin
        select 'truncate ' || string_agg(format('%I.%I', schemaname, tablename), ',')
          into l_stmt
        from pg_tables
        where schemaname in ('public');
      
        execute l_stmt;
      end;
    $$`
	conn.Exec(ctx, q)
}
