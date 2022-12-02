UPDATE repair_game_blocks SET deleted_at = now() WHERE type = 'SHRINK';

ALTER TABLE repair_game_block_logs
DROP COLUMN IF EXISTS size_multiplier;

ALTER TABLE repair_game_blocks
DROP COLUMN IF EXISTS min_size_multiplier,
DROP COLUMN IF EXISTS max_size_multiplier;
