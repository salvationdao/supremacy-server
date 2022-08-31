DROP INDEX IF EXISTS idx_player_kill_log_team_kill_record_search;
DROP TYPE IF EXISTS ABILITY_TRIGGER_TYPE;

ALTER TABLE player_kill_log
    DROP COLUMN IF EXISTS ability_offering_id,
    DROP COLUMN IF EXISTS game_ability_id,
    DROP COLUMN IF EXISTS is_verified,
    DROP COLUMN IF EXISTS related_play_ban_id;

ALTER TABLE battle_ability_triggers
    DROP COLUMN IF EXISTS on_mech_id,
    DROP COLUMN IF EXISTS trigger_type,
    DROP COLUMN IF EXISTS deleted_at;

CREATE TABLE mech_ability_trigger_logs
(
    id              uuid primary key     default gen_random_uuid(),
    triggered_by_id uuid        not null references players (id),
    mech_id         uuid        not null references mechs (id),
    game_ability_id uuid        not null references game_abilities (id),
    battle_number   integer     not null default 0,
    created_at      timestamptz not null default now(),
    updated_at      timestamptz not null default now(),
    deleted_at      timestamptz
);