package db

import (
	"gameserver"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/ninja-software/terror/v2"

	"golang.org/x/net/context"
)

// BattleStarted inserts a new battle into the DB
func BattleStarted(ctx context.Context, conn Conn, battleID gameserver.BattleID, warMachines []*gameserver.WarMachine) error {
	q := `INSERT INTO battle (id, war_machines)
		VALUES ($1, $2);`

	_, err := conn.Exec(ctx, q, battleID, warMachines)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// BattleEnded sets a battle as ended
func BattleEnded(ctx context.Context, conn Conn, battleID gameserver.BattleID, winningWarMachines []*gameserver.WarMachineID, winningCondition gameserver.BattleWinCondition) error {
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
func BattleGet(ctx context.Context, conn Conn, battleID gameserver.BattleID) (*gameserver.Battle, error) {
	result := &gameserver.Battle{}

	q := `SELECT * FROM battle WHERE id = $1;`

	err := pgxscan.Get(ctx, conn, result, q, battleID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return result, nil
}

// WarMachineDestroyed adds a battle log of BattleEventWarMachineDestroyed
func WarMachineDestroyed(ctx context.Context, conn Conn, battleID gameserver.BattleID, warMachineDestroyedEvent gameserver.WarMachineDestroyed) error {
	q := `INSERT INTO battle_event (battle_id, event_type, event)
		VALUES ($1, $2, $3);`
	_, err := conn.Exec(ctx, q, battleID, gameserver.BattleEventWarMachineDestroyed, warMachineDestroyedEvent)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// FactionActionTriggered adds a battle log of BattleEvent
func FactionActionTriggered(ctx context.Context, conn Conn, battleID gameserver.BattleID, factionAbilityEvent gameserver.FactionAbility) error {
	q := `INSERT INTO battle_event (battle_id, event_type, event)
		VALUES ($1, $2, $3);`
	_, err := conn.Exec(ctx, q, battleID, gameserver.BattleEventFactionAbility, factionAbilityEvent)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}
