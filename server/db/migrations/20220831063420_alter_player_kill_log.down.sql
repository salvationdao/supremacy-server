DROP INDEX IF EXISTS idx_player_kill_log_team_kill_record_search;

ALTER TABLE player_kill_log
    DROP COLUMN IF EXISTS ability_offering_id,
    DROP COLUMN IF EXISTS game_ability_id,
    DROP COLUMN IF EXISTS is_verified,
    DROP COLUMN IF EXISTS related_play_ban_id;

ALTER TABLE battle_ability_triggers
    DROP COLUMN IF EXISTS on_mech_id,
    DROP COLUMN IF EXISTS trigger_type,
    DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE mech_ability_trigger_logs_old
    RENAME TO mech_ability_trigger_logs;