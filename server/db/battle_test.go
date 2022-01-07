package db_test

import (
	"context"
	"gameserver"
	"gameserver/db"
	"testing"

	"github.com/gofrs/uuid"
)

//var conn *pgxpool.Pool

func TestIntakes(t *testing.T) {
	ctx := context.Background()

	battleUUID, err := uuid.NewV4()
	if err != nil {
		t.Fatal(err)
	}
	warMachines := []*gameserver.WarMachine{}

	// create `10 war machines
	for i := 0; i < 10; i++ {
		newUUID, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}
		warMachines = append(warMachines, &gameserver.WarMachine{ID: gameserver.WarMachineID(newUUID)})
	}

	t.Run("insert_battle", func(t *testing.T) {
		err = db.BattleStarted(ctx, conn, gameserver.BattleID(battleUUID), warMachines)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("battle_get", func(t *testing.T) {

		battle, err := db.BattleGet(ctx, conn, gameserver.BattleID(battleUUID))
		if err != nil {
			t.Fatal(err)
		}

		if battle.ID.String() != battleUUID.String() {
			t.Fatalf("wanted battleID: %s, got battleID: %s", battleUUID.String(), battle.ID.String())
		}

		if len(battle.WarMachines) != len(warMachines) {
			t.Fatalf("wrong number of war machines, wanted: %d, got: %d", len(warMachines), len(battle.WarMachines))
		}

	outer:
		for _, wm := range battle.WarMachines {
			for _, wwm := range warMachines {
				if wm.ID.String() == wwm.ID.String() {
					continue outer
				}
			}
			t.Fatalf("unable to find match for warmachine: %s", wm.ID.String())
		}
	})
	t.Run("insert_battle_event_war_machine_destroyed", func(t *testing.T) {
		err := db.WarMachineDestroyed(ctx, conn, gameserver.BattleID(battleUUID), gameserver.WarMachineDestroyed{
			DestroyedWarMachineID: warMachines[0].ID,
			KillerWarMachineID:    &warMachines[1].ID,
			KilledBy:              "Laser Cannon",
		})
		if err != nil {
			t.Fatal(err)
		}
	})

}
