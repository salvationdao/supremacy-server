CREATE TABLE player_bans(
    id uuid primary key default gen_random_uuid(),
    -- ban info
    banned_player_id uuid not null references players (id),
    banned_by_id uuid not null references players (id),
    battle_number integer references battles(battle_number),
    reason text not null,
    banned_at timestamptz not null default now(),
    end_at timestamptz,

    -- unban mechanism
    manually_unban_by_id uuid not null,
    manually_unban_reason text,
    manually_unban_at timestamptz,

    -- ban option
    ban_sups_contribute bool not null default false,
    ban_location_select bool not null default false,
    ban_send_chat bool not null default false,
    ban_view_chat bool not null default false,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

