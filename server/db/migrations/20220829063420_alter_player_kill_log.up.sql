ALTER TABLE player_kill_log
    ADD COLUMN IF NOT EXISTS ability_offering_id UUID,
    ADD COLUMN IF NOT EXISTS game_ability_id     UUID REFERENCES game_abilities (id),
    ADD COLUMN IF NOT EXISTS is_verified         BOOL NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS related_play_ban_id UUID REFERENCES player_bans (id);

-- set previous player kill logs is verified
UPDATE
    player_kill_log
SET
    is_verified = true;

CREATE INDEX IF NOT EXISTS idx_player_kill_log_team_kill_record_search ON player_kill_log (game_ability_id, is_team_kill, is_verified, ability_offering_id);