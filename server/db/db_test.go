package db_test

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"server"
	"server/db"
	"testing"

	"github.com/gofrs/uuid"
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

func TestDatabase(t *testing.T) {
	ctx := context.Background()

	// test game map
	gameMap := &server.GameMap{
		Name:          "test",
		ImageUrl:      "url",
		Width:         1,
		Height:        1,
		CellsX:        1,
		CellsY:        1,
		TopPixels:     1,
		LeftPixels:    1,
		Scale:         1,
		DisabledCells: []int{1},
	}

	t.Run("Create game map", func(t *testing.T) {
		err := db.GameMapCreate(ctx, conn, gameMap)
		if err != nil {
			t.Errorf("fail to create game map\n")
			t.Fatal()
			return
		}
	})

	t.Run("Get game map by id", func(t *testing.T) {
		_, err := db.GameMapGet(ctx, conn, gameMap.ID)
		if err != nil {
			t.Errorf("fail to get game map by id\n")
			t.Fatal()
			return
		}
	})

	t.Run("Get random game map", func(t *testing.T) {
		_, err := db.GameMapGetRandom(ctx, conn)
		if err != nil {
			t.Errorf("fail to get random game map\n")
			t.Fatal()
			return
		}
	})

	// Test battle functions
	battle := &server.Battle{
		GameMapID: gameMap.ID,
	}

	t.Run("Start a new battle", func(t *testing.T) {
		err := db.BattleStarted(ctx, conn, battle)
		if err != nil {
			t.Errorf("fail to start a new battle\n")
			t.Fatal()
			return
		}
	})

	t.Run("Get a battle", func(t *testing.T) {
		_, err := db.BattleGet(ctx, conn, battle.ID)
		if err != nil {
			fmt.Println(err)
			t.Errorf("fail to get a battle\n")
			t.Fatal()
			return
		}
	})

	t.Run("End a battle", func(t *testing.T) {
		err := db.BattleEnded(ctx, conn, battle.ID, server.WinConditionLastAlive)
		if err != nil {
			t.Errorf("fail to end a battle\n")
			t.Fatal()
			return
		}
	})

	warMachine := &server.WarMachine{
		ID:              server.WarMachineID(uuid.Must(uuid.NewV4())),
		Name:            "test machine",
		BaseHealthPoint: 1,
		BaseShieldPoint: 1,
	}
	t.Run("Create war machine", func(t *testing.T) {
		err := db.WarMachineCreate(ctx, conn, warMachine)
		if err != nil {
			t.Errorf("fail to create war machine\n")
			t.Fatal()
			return
		}
	})

	t.Run("Assign war machines to the battle", func(t *testing.T) {
		err := db.BattleWarMachineAssign(ctx, conn, battle.ID, []server.WarMachineID{warMachine.ID})
		if err != nil {
			t.Errorf("fail to assign war machines to the battle\n")
			t.Fatal()
			return
		}
	})

	// faction := &server.Faction{
	// 	Label:  "test faction",
	// 	Colour: "test",
	// }
	// t.Run("Create new faction", func(t *testing.T) {
	// 	err := db.FactionCreate(ctx, conn, faction)
	// 	if err != nil {
	// 		t.Errorf("fail to create new faction\n")
	// 		t.Fatal()
	// 		return
	// 	}
	// })

	// t.Run("Get faction", func(t *testing.T) {
	// 	_, err := db.FactionGet(ctx, conn, faction.ID)
	// 	if err != nil {
	// 		t.Errorf("fail to get faction\n")
	// 		t.Fatal()
	// 		return
	// 	}
	// })

	factionAbility := &server.FactionAbility{
		FactionID:              server.FactionID(uuid.Must(uuid.NewV4())),
		Label:                  "test action",
		Type:                   "test",
		Colour:                 "test",
		SupsCost:               100,
		CooldownDurationSecond: 100,
	}

	t.Run("Create new faction action", func(t *testing.T) {
		err := db.FactionAbilityCreate(ctx, conn, factionAbility)
		if err != nil {
			t.Errorf("fail to create new faction action\n")
			t.Fatal()
			return
		}
	})

	warMachineDestroyedEvent := &server.WarMachineDestroyedEvent{
		DestroyedWarMachineID: warMachine.ID,
		KillByWarMachineID:    &warMachine.ID,
	}

	// add battle event
	t.Run("Log war machine destroyed event", func(t *testing.T) {
		err := db.WarMachineDestroyedEventCreate(ctx, conn, battle.ID, warMachineDestroyedEvent)
		if err != nil {
			fmt.Println(err)
			t.Errorf("fail to create war machine destroyed event\n")
			t.Fatal()
			return
		}
	})

	t.Run("Assign assisted war machines to a destroyed event", func(t *testing.T) {
		err := db.WarMachineDestroyedEventAssistedWarMachineSet(ctx, conn, warMachineDestroyedEvent.ID, []server.WarMachineID{warMachine.ID})
		if err != nil {
			fmt.Println(err)
			t.Errorf("fail to assign assisted war machines to war machine destroyed event\n")
			t.Fatal()
			return
		}
	})

	factionAbilityEvent := &server.FactionAbilityEvent{
		FactionAbilityID: factionAbility.ID,
		IsTriggered:      false,
	}

	// add battle event
	t.Run("Log faction action event", func(t *testing.T) {
		err := db.FactionAbilityEventCreate(ctx, conn, battle.ID, factionAbilityEvent)
		if err != nil {
			fmt.Println(err)
			t.Errorf("fail to log faction action event\n")
			t.Fatal()
			return
		}
	})

}
