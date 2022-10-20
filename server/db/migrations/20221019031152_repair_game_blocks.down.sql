ALTER TABLE repair_agent_logs_old
    RENAME TO repair_agent_logs;

DROP INDEX IF EXISTS idx_repair_game_block_log_repair_agent_id;

DROP TABLE repair_game_block_logs;
DROP TABLE repair_game_blocks;

DROP TYPE IF EXISTS REPAIR_GAME_BLOCK_TRIGGER_KEY;
DROP TYPE IF EXISTS REPAIR_GAME_BLOCK_TYPE;
