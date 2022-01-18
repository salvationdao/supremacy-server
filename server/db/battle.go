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
func BattleWarMachineAssign(ctx context.Context, conn Conn, battleID server.BattleID, warMachineNFTs []*server.WarMachineNFT) error {
	q := `
		INSERT INTO 
			battles_war_machines (battle_id, war_machine_id, join_as_faction_id, war_machine_stat)
		VALUES

	`

	var args []interface{}
	for i, warMachineNFT := range warMachineNFTs {

		b, err := json.Marshal(warMachineNFT)
		if err != nil {
			return terror.Error(err)
		}

		args = append(args, b)

		q += fmt.Sprintf("('%s', %d, '%s', $%d)", battleID, warMachineNFT.TokenID, warMachineNFT.FactionID, len(args))

		if i < len(warMachineNFTs)-1 {
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
func BattleWinnerWarMachinesSet(ctx context.Context, conn Conn, battleID server.BattleID, warMachineIDs []uint64) error {
	q := `
		UPDATE
			battles_war_machines
		SET
			is_winner = true
		WHERE 
			battle_id = $1 AND war_machine_id IN (
	`
	for i, warMachineID := range warMachineIDs {
		q += fmt.Sprintf("'%d'", warMachineID)
		if i < len(warMachineIDs)-1 {
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
