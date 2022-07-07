CREATE TABLE syndicates(
    id uuid primary key DEFAULT gen_random_uuid(),
    faction_id uuid not null references factions (id),
    founded_by_id uuid not null references players (id),
    name text not null UNIQUE,
    avatar_url text,
    seat_count int NOT NULL DEFAULT 10,
    join_fee numeric(28) NOT NULL,
    exit_fee numeric(28) NOT NULL,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

ALTER TABLE players
    ADD COLUMN IF NOT EXISTS syndicate_id uuid references syndicates(id);

DROP TYPE IF EXISTS SYNDICATE_EVENT_TYPE;
CREATE TYPE SYNDICATE_EVENT_TYPE AS ENUM (
    'MEMBER_JOIN',
    'MEMBER_LEAVE',
    'UPDATE_PROFILE'
);

CREATE TABLE syndicate_event_log(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    involved_player_id uuid not null references players (id),
    type SYNDICATE_EVENT_TYPE NOT NULL,
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE TABLE syndicate_win_distributions(
    id uuid primary key references syndicates(id),
    deploying_user_percentage decimal NOT NULL DEFAULT 0,
    ability_kill_percentage decimal NOT NULL DEFAULT 0,
    mech_owner_percentage decimal NOT NULL DEFAULT 0,
    syndicate_cut_percentage decimal NOT NULL DEFAULT 0,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);