package db

import (
	"encoding/json"
	"fmt"
	"server"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/ninja-software/terror/v2"

	"golang.org/x/net/context"
)

// BattleStarted inserts a new battle into the DB
func BattleStarted(ctx context.Context, conn Conn, battle *server.Battle) error {
	q := `
		INSERT INTO 
			battles (id, game_map_id)
		VALUES 
			($1, $2)
		RETURNING id, game_map_id, started_at;
	`

	err := pgxscan.Get(ctx, conn, battle, q, battle.ID, battle.GameMapID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// BattleWarMachineAssign assign war machines into a battle
func BattleWarMachineAssign(ctx context.Context, conn Conn, battleID server.BattleID, warMachineMetadatas []*server.WarMachineMetadata) error {
	q := `
		INSERT INTO 
			battles_war_machines (battle_id, war_machine_stat)
		VALUES

	`

	var args []interface{}
	for i, warMachineMetadata := range warMachineMetadatas {

		b, err := json.Marshal(warMachineMetadata)
		if err != nil {
			return terror.Error(err)
		}

		args = append(args, b)

		q += fmt.Sprintf("('%s', $%d)", battleID, len(args))

		if i < len(warMachineMetadatas)-1 {
			q += ","
			continue
		}

		q += ";"
	}
	_, err := conn.Exec(ctx, q, args...)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// BattleEnded sets a battle as ended
func BattleEnded(ctx context.Context, conn Conn, battleID server.BattleID, winningCondition server.BattleWinCondition) error {
	q := `
		UPDATE 
			battles
		SET 
			winning_condition = $2, ended_at = NOW()
		WHERE 
			id = $1;
	`

	_, err := conn.Exec(ctx, q, battleID, winningCondition)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// BattleWinnerWarMachinesSet set war machine as winner
func BattleWinnerWarMachinesSet(ctx context.Context, conn Conn, battleID server.BattleID, warMachines []*server.WarMachineMetadata) error {
	q := `
		UPDATE
			battles_war_machines
		SET
			is_winner = true
		WHERE 
			battle_id = $1 AND war_machine_stat->>'tokenID' IN (
	`
	for i, warMachine := range warMachines {
		q += fmt.Sprintf("'%d'", warMachine.TokenID)
		if i < len(warMachines)-1 {
			q += ","
			continue
		}
		q += ")"
	}

	_, err := conn.Exec(ctx, q, battleID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// BattleGet gets a battle via battle uuid
func BattleGet(ctx context.Context, conn Conn, battleID server.BattleID) (*server.Battle, error) {
	result := &server.Battle{}

	q := `SELECT * FROM battles WHERE id = $1;`

	err := pgxscan.Get(ctx, conn, result, q, battleID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return result, nil
}

// CreateBattleStateEvent adds a battle log of BattleEvent
func CreateBattleStateEvent(ctx context.Context, conn Conn, battleID server.BattleID, state string) (*server.BattleEventStateChange, error) {
	event := &server.BattleEventStateChange{}
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
			battle_events_state (event_id, state)
		VALUES 
			((SELECT id FROM rows), $3)
		RETURNING 
			id, event_id, state;
	`
	err := pgxscan.Get(ctx, conn, event, q,
		battleID,
		server.BattleEventTypeStateChange,
		state,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return event, nil
}
