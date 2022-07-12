CREATE TABLE symbols(
    id uuid primary key default gen_random_uuid(),
    image_url text NOT NULL,
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

DROP TYPE IF EXISTS SYNDICATE_TYPE;
CREATE TYPE SYNDICATE_TYPE AS ENUM (
    'CORPORATION',
    'DECENTRALISED'
);

CREATE TABLE syndicates(
    id uuid primary key DEFAULT gen_random_uuid(),
    type SYNDICATE_TYPE not null,
    faction_id uuid not null references factions (id),
    founded_by_id uuid not null references players (id),
    honorary_founder bool not null default false,
    name text not null UNIQUE,
    symbol_id uuid NOT NULL REFERENCES symbols (id),
    naming_convention text,
    seat_count int NOT NULL DEFAULT 10,

    -- payment detail
    join_fee numeric(28) NOT NULL default 0,
    exit_fee numeric(28) NOT NULL default 0,

    -- battle win columns
    deploying_member_cut_percentage decimal NOT NULL DEFAULT 0,
    member_assist_cut_percentage decimal NOT NULL DEFAULT 0,
    mech_owner_cut_percentage decimal NOT NULL DEFAULT 0,
    syndicate_cut_percentage decimal NOT NULL DEFAULT 0,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

ALTER TABLE players
    ADD COLUMN IF NOT EXISTS syndicate_id uuid references syndicates(id),
    ADD COLUMN IF NOT EXISTS director_of_syndicate_id uuid REFERENCES syndicates(id);

CREATE INDEX IF NOT EXISTS idx_player_syndicate on players(syndicate_id);
CREATE INDEX IF NOT EXISTS idx_player_syndicate_director on players(director_of_syndicate_id);

DROP TYPE IF EXISTS SYNDICATE_EVENT_TYPE;
CREATE TYPE SYNDICATE_EVENT_TYPE AS ENUM (
    'MEMBER_JOIN',
    'MEMBER_LEAVE',
    'UPDATE_PROFILE',
    'CONTRIBUTE_FUND'
);

CREATE TABLE syndicate_event_log(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    involved_player_id uuid not null references players (id),
    type SYNDICATE_EVENT_TYPE NOT NULL,
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE TABLE syndicate_rules(
    id uuid primary key default gen_random_uuid(),
    syndicate_id uuid not null references syndicates(id),
    number int not null,
    content text not null,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_syndicate_rule_syndicate on syndicate_rules(syndicate_id);

DROP TYPE IF EXISTS SYNDICATE_MOTION_TYPE;
CREATE TYPE SYNDICATE_MOTION_TYPE AS ENUM (
    'CHANGE_GENERAL_DETAIL',
    'CHANGE_PAYMENT_SETTING',
    'ADD_RULE',
    'REMOVE_RULE',
    'CHANGE_RULE',
    'APPOINT_DIRECTOR',
    'REMOVE_DIRECTOR',
    'REMOVE_FOUNDER'
);

DROP TYPE IF EXISTS SYNDICATE_MOTION_RESULT;
CREATE TYPE SYNDICATE_MOTION_RESULT AS ENUM (
    'PASSED',
    'FAILED',
    'FORCE_CLOSED'
);

CREATE TABLE syndicate_motions(
    id uuid primary key default gen_random_uuid(),
    syndicate_id uuid not null references syndicates(id),
    type SYNDICATE_MOTION_TYPE not null,
    issued_by_id uuid not null references players(id),
    reason text not null,

    -- content
    new_symbol_id uuid references symbols(id),
    new_name text,
    new_naming_convention text,

    -- payment change
    new_join_fee numeric(28),
    new_exit_fee numeric(28),
    new_deploying_member_cut_percentage decimal,
    new_member_assist_cut_percentage decimal,
    new_mech_owner_cut_percentage decimal,
    new_syndicate_cut_percentage decimal,

    -- add/remove/change rule
    rule_id uuid references syndicate_rules(id),
    new_rule_number int,
    new_rule_content text,

    -- appoint/remove director
    director_id uuid references players(id),

    result SYNDICATE_MOTION_RESULT,
    ended_at timestamptz not null,
    actual_ended_at timestamptz,
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE TABLE syndicate_motion_votes(
    id uuid primary key default gen_random_uuid(),
    motion_id uuid not null references syndicate_motions(id),
    vote_by_id uuid not null references players(id),
    is_agreed bool not null,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_motion_vote_motion_id on syndicate_motion_votes(motion_id);