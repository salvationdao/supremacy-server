create table devices(
    id uuid primary key default gen_random_uuid(),
    player_id uuid not null references players(id),
    device_id text not null,
    name text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);