package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"server"
	"server/gamelog"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"

	"github.com/ninja-software/terror/v2"

	"golang.org/x/net/context"
)

// BattleStarted inserts a new battle into the DB
func BattleStarted(ctx context.Context, conn Conn, battle *server.Battle) error {
	gamelog.GameLog.Info().Str("fn", "BattleStarted").Msg("db func")
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
	gamelog.GameLog.Info().Str("fn", "BattleWarMachineAssign").Msg("db func")
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
	gamelog.GameLog.Info().Str("fn", "BattleEnded").Msg("db func")
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
	gamelog.GameLog.Info().Str("fn", "BattleWinnerWarMachinesSet").Msg("db func")
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
	gamelog.GameLog.Info().Str("fn", "BattleGet").Msg("db func")
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
	gamelog.GameLog.Info().Str("fn", "CreateBattleStateEvent").Msg("db func")
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
func BattleQueueInsert(ctx context.Context, conn Conn, warMachineMetadata *server.WarMachineMetadata, contractReward string, isInsured bool) error {
	gamelog.GameLog.Info().Str("fn", "BattleQueueInsert").Msg("db func")
	// marshal metadata
	jb, err := json.Marshal(warMachineMetadata)
	if err != nil {
		return terror.Error(err)
	}

	q := `
		INSERT INTO 
			battle_war_machine_queues (war_machine_metadata, contract_reward, is_insured)
		VALUES
			($1, $2, $3)
	`

	_, err = conn.Exec(ctx, q, jb, contractReward, isInsured)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func BattleQueueWarMachineUpdate(ctx context.Context, conn Conn, warMachineMetadata *server.WarMachineMetadata) error {
	gamelog.GameLog.Info().Str("fn", "BattleQueueWarMachineUpdate").Msg("db func")
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
		war_machine_metadata ->> 'hash' = $2 AND deleted_at ISNULL
	`

	_, err = conn.Exec(ctx, q, jb, warMachineMetadata.Hash)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func BattleQueueRemove(ctx context.Context, conn Conn, warMachineMetadata *server.WarMachineMetadata) error {
	gamelog.GameLog.Info().Str("fn", "BattleQueueRemove").Msg("db func")
	q := `
			UPDATE
				battle_war_machine_queues
			SET
				deleted_at = NOW()
			WHERE
				war_machine_metadata ->> 'hash' = $1 AND 
				deleted_at ISNULL
		`

	_, err := conn.Exec(ctx, q, warMachineMetadata.Hash)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func BattleQueueGetByFactionID(ctx context.Context, conn Conn, factionID server.FactionID) ([]*server.WarMachineMetadata, error) {
	gamelog.GameLog.Info().Str("fn", "BattleQueueGetByFactionID").Msg("db func")
	bqs := []*server.BattleQueueMetadata{}
	q := `
			SELECT
				war_machine_metadata, contract_reward
			FROM
				battle_war_machine_queues
			WHERE
				war_machine_metadata ->> 'factionID' = $1 AND deleted_at ISNULL
			ORDER BY
				queued_at asc
		`

	err := pgxscan.Select(ctx, conn, &bqs, q, factionID)
	if err != nil {
		return []*server.WarMachineMetadata{}, terror.Error(err)
	}

	wms := []*server.WarMachineMetadata{}
	for _, bq := range bqs {
		// insert contract reward in the mech
		bq.WarMachineMetadata.ContractReward = &bq.ContractReward

		wms = append(wms, bq.WarMachineMetadata)
	}

	return wms, nil
}

func AssetQueuingStat(ctx context.Context, conn Conn, hash string) (*server.BattleQueueMetadata, error) {
	gamelog.GameLog.Info().Str("fn", "AssetQueuingStat").Msg("db func")
	result := &server.BattleQueueMetadata{}
	q := `
		SELECT 
			*
		FROM
			battle_war_machine_queues
		WHERE
			war_machine_metadata ->> 'hash' = $1 AND deleted_at ISNULL
		limit 1
	`
	err := pgxscan.Get(ctx, conn, result, q, hash)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, terror.Error(err)
	}

	return result, nil
}

/*********************
* Asset Repair stuff *
*********************/
func AssetRepairInsert(ctx context.Context, conn Conn, assetRepairRecord *server.AssetRepairRecord) error {
	gamelog.GameLog.Info().Str("fn", "AssetRepairInsert").Msg("db func")
	q := `
		INSERT INTO
			asset_repair (hash, expect_completed_at, repair_mode)
		VALUES
			($1, $2, $3)
		RETURNING
			hash, expect_completed_at, repair_mode
	`

	err := pgxscan.Get(ctx, conn, assetRepairRecord, q,
		assetRepairRecord.Hash,
		assetRepairRecord.ExpectCompletedAt,
		assetRepairRecord.RepairMode,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func AssetRepairIncompleteGet(ctx context.Context, conn Conn, assetRepairRecord *server.AssetRepairRecord) error {
	gamelog.GameLog.Info().Str("fn", "AssetRepairIncompleteGet").Msg("db func")
	q := `
		SELECT * FROM asset_repair WHERE hash = $1 AND completed_at ISNULL
		limit 1;
	`

	err := pgxscan.Get(ctx, conn, assetRepairRecord, q, assetRepairRecord.Hash)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func AssetRepairPaidToComplete(ctx context.Context, conn Conn, assetRepairRecord *server.AssetRepairRecord) error {
	gamelog.GameLog.Info().Str("fn", "AssetRepairPaidToComplete").Msg("db func")
	q := `
		UPDATE
			asset_repair
		SET
			is_paidToComplete = TRUE,
			completedAt = NOW()
		WHERE
			hash = $1 AND completed_at ISNULL
		RETURNING
			hash, expect_completed_at, repair_mode, is_paid_to_complete, completed_at, created_at
	`
	err := pgxscan.Get(ctx, conn, assetRepairRecord, q, assetRepairRecord.Hash)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
