package db

import (
	"fmt"
	"server"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"golang.org/x/net/context"
)

// WarMachineDestroyedEventCreate adds a battle log of BattleEventWarMachineDestroyed
func WarMachineDestroyedEventCreate(ctx context.Context, conn Conn, battleID server.BattleID, warMachineDestroyedEvent *server.WarMachineDestroyedEvent) error {
	q := `
		WITH rows AS (
			INSERT INTO 
				battle_events (battle_id) 
			VALUES
				($1)
			RETURNING
				id
		)
		INSERT INTO 
			war_machine_destroyed_events (event_id, destroyed_war_machine_id, kill_by_war_machine_id, kill_by_faction_ability_id)
		VALUES 
			((SELECT id FROM rows) ,$2, $3, $4)
		RETURNING 
			id, event_id, destroyed_war_machine_id, kill_by_war_machine_id, kill_by_faction_ability_id;
	`
	err := pgxscan.Get(ctx, conn, warMachineDestroyedEvent, q,
		battleID,
		warMachineDestroyedEvent.DestroyedWarMachineID,
		warMachineDestroyedEvent.KillByWarMachineID,
		warMachineDestroyedEvent.KillByFactionAbilityID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// WarMachineDestroyedEventAssistedWarMachineSet assign assisted war machine to a war machine destroyed event
func WarMachineDestroyedEventAssistedWarMachineSet(ctx context.Context, conn Conn, eventID server.WarMachineDestroyedEventID, warMachineIDs []uint64) error {
	q := `
		INSERT INTO
			war_machine_destroyed_events_assisted_war_machines (war_machine_destroyed_event_id, war_machine_id)
		VALUES 
	`

	for i, warMachineID := range warMachineIDs {
		q += fmt.Sprintf("('%s','%d')", eventID, warMachineID)

		if i < len(warMachineIDs)-1 {
			q += ","
			continue
		}

		q += ";"
	}

	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// FactionAbilityEventCreate adds a battle log of BattleEvent
func FactionAbilityEventCreate(ctx context.Context, conn Conn, battleID server.BattleID, factionAbilityEvent *server.FactionAbilityEvent) error {
	q := `
		WITH rows AS (
			INSERT INTO 
				battle_events (battle_id) 
			VALUES
				($1)
			RETURNING
				id
		)
		INSERT INTO 
			faction_ability_events (event_id, faction_ability_id, is_triggered, triggered_by_user, triggered_on_cell_x, triggered_on_cell_y)
		VALUES 
			((SELECT id FROM rows), $2, $3, $4, $5, $6)
		RETURNING 
			id, event_id, faction_ability_id, is_triggered, triggered_by_user, triggered_on_cell_x, triggered_on_cell_y;
	`
	err := pgxscan.Get(ctx, conn, factionAbilityEvent, q,
		battleID,
		factionAbilityEvent.FactionAbilityID,
		factionAbilityEvent.IsTriggered,
		factionAbilityEvent.TriggeredByUser,
		factionAbilityEvent.TriggeredOnCellX,
		factionAbilityEvent.TriggeredOnCellY,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}
