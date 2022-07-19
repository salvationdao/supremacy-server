UPDATE
    game_abilities
SET
    description = 'Consume your remaining shield for an explosive defence mechanism.',
    deleted_at = null
WHERE
    label = 'OVERCHARGE';

UPDATE
    game_abilities
SET
    deleted_at = now()
WHERE
    game_client_ability_id = 3 or game_client_ability_id = 4 OR game_client_ability_id = 7;

INSERT INTO
    game_abilities (game_client_ability_id, faction_id, label, colour, image_url, description, text_colour, level, sups_cost)
VALUES
    (5, '98bf7bb3-1a7c-4f21-8843-458d62884060', 'OVERCHARGE','#FFFFFF','https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-overcharge.jpg', 'Consume your remaining shield for an explosive defence mechanism.','#000000','MECH','100000000000000000000'),
    (5, '7c6dde21-b067-46cf-9e56-155c88a520e2', 'OVERCHARGE','#FFFFFF','https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-overcharge.jpg', 'Consume your remaining shield for an explosive defence mechanism.','#000000','MECH','100000000000000000000');

UPDATE
    game_abilities
SET
    level = 'PLAYER'
WHERE
    game_client_ability_id = 5 or game_client_ability_id = 2;

CREATE TABLE mech_ability_trigger_logs(
    id uuid primary key default gen_random_uuid(),
    triggered_by_id uuid not null references players(id),
    mech_id uuid not null references mechs(id),
    game_ability_id uuid not null references game_abilities(id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

CREATE INDEX idx_mech_ability_trigger_log_search ON mech_ability_trigger_logs(mech_id, game_ability_id, created_at DESC, deleted_at);
CREATE INDEX idx_mech_ability_trigger_log_created_at_descending ON mech_move_command_logs(created_at DESC);