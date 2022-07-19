CREATE TABLE battle_ability_opt_in_logs(
    id uuid primary key default gen_random_uuid(),
    battle_id uuid not null references battles(id),

    player_id uuid not null references players(id),
    battle_ability_offering_id uuid not null, -- this is generate from go
    CONSTRAINT opt_per_player_per_ability UNIQUE (player_id, battle_ability_offering_id),

    faction_id uuid not null references factions (id),
    battle_ability_id uuid not null references battle_abilities(id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

create index battle_ability_opt_in_logs_created_at_descending ON battle_ability_opt_in_logs(created_at DESC);

create index idx_player_active_log_active_at_descending on player_active_logs(active_at desc);