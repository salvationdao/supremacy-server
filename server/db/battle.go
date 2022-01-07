package db

import (
	"server"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/ninja-software/terror/v2"

	"golang.org/x/net/context"
)

// BattleStarted inserts a new battle into the DB
func BattleStarted(ctx context.Context, conn Conn, battleID server.BattleID, warMachines []*server.WarMachine) error {
	q := `INSERT INTO battle (id, war_machines)
		VALUES ($1, $2);`

	_, err := conn.Exec(ctx, q, battleID, warMachines)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// BattleEnded sets a battle as ended
func BattleEnded(ctx context.Context, conn Conn, battleID server.BattleID, winningWarMachines []*server.WarMachineID, winningCondition server.BattleWinCondition) error {
	q := `
	UPDATE battle
	SET winning_war_machines = $2, winning_condition = $3, ended_at = NOW()
	WHERE id = $1;`

	_, err := conn.Exec(ctx, q, battleID, winningWarMachines, winningCondition)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// BattleGet gets a battle via battle uuid
func BattleGet(ctx context.Context, conn Conn, battleID server.BattleID) (*server.Battle, error) {
	result := &server.Battle{}

	q := `SELECT * FROM battle WHERE id = $1;`

	err := pgxscan.Get(ctx, conn, result, q, battleID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return result, nil
}

// WarMachineDestroyed adds a battle log of BattleEventWarMachineDestroyed
func WarMachineDestroyed(ctx context.Context, conn Conn, battleID server.BattleID, warMachineDestroyedEvent server.WarMachineDestroyed) error {
	q := `INSERT INTO battle_event (battle_id, event_type, event)
		VALUES ($1, $2, $3);`
	_, err := conn.Exec(ctx, q, battleID, server.BattleEventWarMachineDestroyed, warMachineDestroyedEvent)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// FactionActionTriggered adds a battle log of BattleEvent
func FactionActionTriggered(ctx context.Context, conn Conn, battleID server.BattleID, factionAbilityEvent server.FactionAbility) error {
	q := `INSERT INTO battle_event (battle_id, event_type, event)
		VALUES ($1, $2, $3);`
	_, err := conn.Exec(ctx, q, battleID, server.BattleEventFactionAbility, factionAbilityEvent)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}
