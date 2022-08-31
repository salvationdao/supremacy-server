ALTER TABLE player_kill_log
    ADD COLUMN IF NOT EXISTS ability_offering_id UUID,
    ADD COLUMN IF NOT EXISTS game_ability_id     UUID REFERENCES game_abilities (id),
    ADD COLUMN IF NOT EXISTS is_verified         BOOL NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS related_play_ban_id UUID REFERENCES player_bans (id);

-- set previous player kill logs is verified
UPDATE
    player_kill_log
SET is_verified = TRUE;

CREATE INDEX IF NOT EXISTS idx_player_kill_log_team_kill_record_search ON player_kill_log (player_id, game_ability_id, is_team_kill, is_verified);

DROP TYPE IF EXISTS ABILITY_TRIGGER_TYPE;
CREATE TYPE ABILITY_TRIGGER_TYPE AS ENUM ('BATTLE_ABILITY','MECH_ABILITY','PLAYER_ABILITY');

ALTER TABLE battle_ability_triggers
    ADD COLUMN IF NOT EXISTS on_mech_id uuid REFERENCES mechs (id),
    ADD COLUMN IF NOT EXISTS trigger_type ABILITY_TRIGGER_TYPE NOT NULL DEFAULT 'BATTLE_ABILITY',
    ADD COLUMN IF NOT EXISTS deleted_at timestamptz;

ALTER TABLE mech_ability_trigger_logs
    RENAME TO mech_ability_trigger_logs_old;