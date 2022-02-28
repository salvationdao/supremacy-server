package db

import (
	"context"
	"fmt"
	"log"
	"server"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

// WarMachineDestroyedEventCreate adds a battle log of BattleEventTypeWarMachineDestroyed
func WarMachineDestroyedEventCreate(ctx context.Context, conn Conn, battleID server.BattleID, warMachineDestroyedEvent *server.WarMachineDestroyedEvent) error {
	q := `
		WITH rows AS (
			INSERT INTO 
				battle_events (battle_id, event_type) 
			VALUES
				($1, $2)
			RETURNING
				id
		)
		INSERT INTO 
			battle_events_war_machine_destroyed (event_id, destroyed_war_machine_hash, kill_by_war_machine_hash, related_event_id)
		VALUES 
			((SELECT id FROM rows) ,$3, $4, $5)
		RETURNING 
			id, event_id, destroyed_war_machine_hash, kill_by_war_machine_hash, related_event_id;
	`

	err := pgxscan.Get(ctx, conn, warMachineDestroyedEvent, q,
		battleID,
		server.BattleEventTypeWarMachineDestroyed,
		warMachineDestroyedEvent.DestroyedWarMachineHash,
		warMachineDestroyedEvent.KillByWarMachineHash,
		warMachineDestroyedEvent.RelatedEventID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// WarMachineDestroyedEventAssistedWarMachineSet assign assisted war machine to a war machine destroyed event
func WarMachineDestroyedEventAssistedWarMachineSet(ctx context.Context, conn Conn, eventID server.WarMachineDestroyedEventID, warMachineHashs []string) error {
	q := `
		INSERT INTO
			battle_events_war_machine_destroyed_assisted_war_machines (war_machine_destroyed_event_id, war_machine_id)
		VALUES 
	`

	for i, warMachineID := range warMachineHashs {
		q += fmt.Sprintf("('%s','%s')", eventID, warMachineID)

		if i < len(warMachineHashs)-1 {
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

// GameAbilityEventCreate adds a battle log of BattleEvent
func GameAbilityEventCreate(ctx context.Context, conn Conn, battleID server.BattleID, gameAbilityEvent *server.GameAbilityEvent) error {
	q := `
		WITH rows AS (
			INSERT INTO 
				battle_events (battle_id, event_type) 
			VALUES
				($1, $2)
			RETURNING
				id
		)
		INSERT INTO 
			battle_events_game_ability (event_id, game_ability_id, ability_hash, is_triggered, triggered_by_user_id, triggered_on_cell_x, triggered_on_cell_y)
		VALUES 
			((SELECT id FROM rows), $3, $4, $5, $6, $7, $8)
		RETURNING 
			id, event_id, game_ability_id, ability_hash, is_triggered, triggered_by_user_id, triggered_on_cell_x, triggered_on_cell_y;
	`
	err := pgxscan.Get(ctx, conn, gameAbilityEvent, q,
		battleID,
		server.BattleEventTypeGameAbility,
		gameAbilityEvent.GameAbilityID,
		gameAbilityEvent.AbilityHash,
		gameAbilityEvent.IsTriggered,
		gameAbilityEvent.TriggeredByUserID,
		gameAbilityEvent.TriggeredOnCellX,
		gameAbilityEvent.TriggeredOnCellY,
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
			battle_events_war_machine_destroyed wmde
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
		case server.BattleEventTypeWarMachineDestroyed:
			eventObj := &server.WarMachineDestroyedEvent{}
			q = `SELECT * FROM battle_events_war_machine_destroyed WHERE event_id = $1`
			err := pgxscan.Get(ctx, conn, eventObj, q, evnt.ID)
			if err != nil {
				return nil, terror.Error(err)
			}
			evnt.Event = eventObj
		case server.BattleEventTypeGameAbility:
			eventObj := &server.GameAbilityEvent{}
			q = `SELECT * FROM battle_events_game_ability WHERE event_id = $1`
			err := pgxscan.Get(ctx, conn, eventObj, q, evnt.ID)
			if err != nil {
				return nil, terror.Error(err)
			}
			if eventObj == nil || eventObj.GameAbilityID == nil || eventObj.GameAbilityID.IsNil() {
				log.Println("missing game ability ID")
				continue
			}
			result, err := GameAbility(ctx, conn, *eventObj.GameAbilityID)
			if err != nil {
				return nil, terror.Error(err)
			}
			eventObj.GameAbility = result
			evnt.Event = eventObj
		case server.BattleEventTypeStateChange:
			eventObj := &server.BattleEventStateChange{}
			q = `SELECT * FROM battle_events_state WHERE event_id = $1`
			err := pgxscan.Get(ctx, conn, eventObj, q, evnt.ID)
			if err != nil {
				return nil, terror.Error(err)
			}
			evnt.Event = eventObj
		}
	}

	return events, nil
}
