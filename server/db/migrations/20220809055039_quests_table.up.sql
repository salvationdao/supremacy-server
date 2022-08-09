CREATE TABLE quests(
    id uuid not null default gen_random_uuid(),
    name text not null,
    description text not null,
    -- requirement
    required_number int not null,

    -- repeatable
    ended_at timestamptz not null,
    interval_duration_hours int,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);