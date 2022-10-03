DROP INDEX IF EXISTS idx_battle_bounties_available_check;
DROP TABLE IF EXISTS battle_bounties;

DROP INDEX IF EXISTS idx_battle_lobbies_mechs_queue_battle_check;
DROP INDEX IF EXISTS idx_battle_lobbies_mechs_queue_check;
DROP INDEX IF EXISTS idx_battle_lobbies_mechs_lobby_queue_check;
DROP TABLE IF EXISTS battle_lobbies_mechs;

DROP INDEX IF EXISTS idx_battle_lobby_complete_check;
DROP INDEX IF EXISTS idx_battle_lobby_scheduled_at;
DROP INDEX IF EXISTS idx_battle_lobby_queue_available_check;
DROP INDEX IF EXISTS idx_battle_lobby_ready_available_check;
DROP INDEX IF EXISTS idx_battle_lobby_queue_position_check;
DROP TABLE IF EXISTS battle_lobbies;

DROP INDEX IF EXISTS idx_player_kill_log_offering_id;

ALTER TABLE battle_map_queue_old RENAME TO battle_map_queue;
ALTER TABLE battle_war_machine_queues_old RENAME TO battle_war_machine_queues;
ALTER TABLE battle_queue_fees_old RENAME TO battle_queue_fees;
ALTER TABLE battle_queue_old RENAME TO battle_queue;
ALTER TABLE repair_cases DROP COLUMN IF EXISTS paused_at;

ALTER TABLE battles RENAME COLUMN ended_battle_seconds_old TO ended_battle_seconds;
ALTER TABLE battles RENAME COLUMN started_battle_seconds_old TO started_battle_seconds;
