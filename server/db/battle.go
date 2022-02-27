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
		RETURNING id, identifier, game_map_id, started_at;
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
			battle_id = $1 AND war_machine_stat->>'hash' IN (
	`
	for i, warMachine := range warMachines {
		q += fmt.Sprintf("'%s'", warMachine.Hash)
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
func CreateBattleStateEvent(ctx context.Context, conn Conn, battleID server.BattleID, state string, detail []byte) (*server.BattleEventStateChange, error) {
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
			battle_events_state (event_id, state, detail)
		VALUES 
			((SELECT id FROM rows), $3, $4)
		RETURNING 
			id, event_id, state;
	`
	err := pgxscan.Get(ctx, conn, event, q,
		battleID,
		server.BattleEventTypeStateChange,
		state,
		detail,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return event, nil
}

/*********************
* Battle Queue stuff *
*********************/
func BattleQueueInsert(ctx context.Context, conn Conn, warMachineMetadata *server.WarMachineMetadata) error {
	// marshal metadata
	jb, err := json.Marshal(warMachineMetadata)
	if err != nil {
		return terror.Error(err)
	}

	q := `
		INSERT INTO 
			battle_war_machine_queues (war_machine_metadata)
		VALUES
			($1)
	`

	_, err = conn.Exec(ctx, q, jb)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func BattleQueueWarMachineUpdate(ctx context.Context, conn Conn, warMachineMetadata *server.WarMachineMetadata) error {
	// marshal metadata
	jb, err := json.Marshal(warMachineMetadata)
	if err != nil {
		return terror.Error(err)
	}

	q := `
	UPDATE
		battle_war_machine_queues
	SET
		war_machine_metadata = $1
	WHERE
		war_machine_metadata ->> 'hash' = $2 AND released_at ISNULL
	`

	_, err = conn.Exec(ctx, q, jb, warMachineMetadata.Hash)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func BattleQueueRemove(ctx context.Context, conn Conn, warMachineMetadata *server.WarMachineMetadata) error {
	q := `
			UPDATE
				battle_war_machine_queues
			SET
				released_at = NOW()
			WHERE
				war_machine_metadata ->> 'hash' = $1 AND 
				war_machine_metadata ->> 'factionID' = $2 AND 
				released_at ISNULL
		`

	_, err := conn.Exec(ctx, q, warMachineMetadata.Hash, warMachineMetadata.FactionID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func BattleQueueRead(ctx context.Context, conn Conn, factionID server.FactionID) ([]*server.WarMachineMetadata, error) {
	bqs := []*server.BattleQueueMetadata{}
	q := `
			SELECT
				war_machine_metadata
			FROM
				battle_war_machine_queues
			WHERE
				war_machine_metadata ->> 'factionID' = $1 AND released_at ISNULL
			ORDER BY
				queued_at asc
		`

	err := pgxscan.Select(ctx, conn, &bqs, q, factionID)
	if err != nil {
		return []*server.WarMachineMetadata{}, terror.Error(err)
	}

	wms := []*server.WarMachineMetadata{}
	for _, bq := range bqs {
		wms = append(wms, bq.WarMachineMetadata)
	}

	return wms, nil
}

func BattleQueueingHashesGet(ctx context.Context, conn Conn) ([]string, error) {
	bqh := []string{}
	q := `
		SELECT
			war_machine_metadata ->> 'hash' as hash
		FROM
			battle_war_machine_queues
		where 
			released_at ISNULL
	`

	err := pgxscan.Select(ctx, conn, &bqh, q)
	if err != nil {
		return []string{}, terror.Error(err)
	}

	return bqh, nil
}
