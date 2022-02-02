package db

import (
	"context"
	"fmt"
	"server"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
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
			war_machine_destroyed_events (event_id, destroyed_war_machine_id, kill_by_war_machine_id, related_event_id)
		VALUES 
			((SELECT id FROM rows) ,$2, $3, $4)
		RETURNING 
			id, event_id, destroyed_war_machine_id, kill_by_war_machine_id, related_event_id;
	`

	err := pgxscan.Get(ctx, conn, warMachineDestroyedEvent, q,
		battleID,
		warMachineDestroyedEvent.DestroyedWarMachineID,
		warMachineDestroyedEvent.KillByWarMachineID,
		warMachineDestroyedEvent.RelatedEventID,
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
			faction_ability_events (event_id, faction_ability_id, is_triggered, triggered_by_user_id, triggered_on_cell_x, triggered_on_cell_y)
		VALUES 
			((SELECT id FROM rows), $2, $3, $4, $5, $6)
		RETURNING 
			id, event_id, faction_ability_id, is_triggered, triggered_by_user_id, triggered_on_cell_x, triggered_on_cell_y;
	`
	err := pgxscan.Get(ctx, conn, factionAbilityEvent, q,
		battleID,
		factionAbilityEvent.FactionAbilityID,
		factionAbilityEvent.IsTriggered,
		factionAbilityEvent.TriggeredByUserID,
		factionAbilityEvent.TriggeredOnCellX,
		factionAbilityEvent.TriggeredOnCellY,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// WarMachineDestroyedEventGetByBattleID return a list of war machine destroyed events by given battle id
func WarMachineDestroyedEventGetByBattleID(ctx context.Context, conn Conn, battleID server.BattleID) ([]*server.WarMachineDestroyedEvent, error) {
	events := []*server.WarMachineDestroyedEvent{}

	q := `
		SELECT wmde.* 
		FROM 
			war_machine_destroyed_events wmde
		INNER JOIN 
			battle_events be ON be.id = wmde. event_id AND be.battle_id = $1
	`

	err := pgxscan.Select(ctx, conn, &events, q, battleID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return events, nil
}

func GetEvents(ctx context.Context, conn Conn, since *time.Time) ([]*server.BattleEvent, error) {
	events := []*server.BattleEvent{}

	var args []interface{}
	fromQuery := ""
	limitQuery := "LIMIT 200"

	if since != nil && !since.IsZero() {
		args = append(args, since)
		fromQuery = "WHERE created_at > $1"
		limitQuery = "LIMIT 2000"
	}

	q := fmt.Sprintf(`SELECT * FROM battle_events %s ORDER BY created_at DESC  %s`, fromQuery, limitQuery)
	err := pgxscan.Select(ctx, conn, &events, q, args...)
	if err != nil {
		return nil, terror.Error(err)
	}

	for _, evnt := range events {
		switch evnt.EventType {
		case server.BattleEventWarMachineDestroyed:
			eventObj := &server.WarMachineDestroyedEvent{}
			q = `SELECT * FROM war_machine_destroyed_events WHERE event_id = $1`

			err := pgxscan.Get(ctx, conn, eventObj, q, evnt.ID)
			if err != nil {
				return nil, terror.Error(err)
			}
			evnt.Event = eventObj
		case server.BattleEventFactionAbility:
			eventObj := &server.FactionAbilityEvent{}
			q = `SELECT * FROM faction_ability_events WHERE event_id = $1`
			err := pgxscan.Get(ctx, conn, eventObj, q, evnt.ID)
			if err != nil {
				return nil, terror.Error(err)
			}
			evnt.Event = eventObj
		}

	}

	return events, nil
}
