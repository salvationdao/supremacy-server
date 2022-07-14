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

    ceo_player_id uuid references players(id),
    admin_id uuid references players(id),

    seat_count int NOT NULL DEFAULT 10,

    -- general detail
    name text not null UNIQUE,
    symbol TEXT NOT NULL UNIQUE,
    logo_id uuid references blobs(id),

    -- payment detail
    join_fee numeric(28) NOT NULL default 0,
    exit_fee numeric(28) NOT NULL default 0,

    monthly_dues numeric(28) NOT NULL DEFAULT 0,

    -- battle win columns
    deploying_member_cut_percentage decimal NOT NULL DEFAULT 0,
    member_assist_cut_percentage decimal NOT NULL DEFAULT 0,
    mech_owner_cut_percentage decimal NOT NULL DEFAULT 0,
    syndicate_cut_percentage decimal NOT NULL DEFAULT 0,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE TABLE syndicate_directors(
    syndicate_id uuid not null references syndicates(id),
    player_id uuid not null references players(id),
    PRIMARY KEY (syndicate_id, player_id),
    created_at timestamptz not null default NOW()
);

CREATE INDEX IF NOT EXISTS idx_syndicate_director_syndicate on syndicate_directors(syndicate_id);
CREATE INDEX IF NOT EXISTS idx_syndicate_director_player on syndicate_directors(player_id);
CREATE INDEX IF NOT EXISTS idx_syndicate_director_search on syndicate_directors(syndicate_id, player_id);

CREATE TABLE syndicate_committees(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    syndicate_id uuid not null references syndicates(id),
    player_id uuid not null references players(id),
    created_at timestamptz not null default NOW()
);

CREATE INDEX IF NOT EXISTS idx_syndicate_committees_syndicate on syndicate_committees(syndicate_id);
CREATE INDEX IF NOT EXISTS idx_syndicate_committees_player on syndicate_committees(player_id);
CREATE INDEX IF NOT EXISTS idx_syndicate_committees_search on syndicate_committees(syndicate_id, player_id);

ALTER TABLE players
    ADD COLUMN IF NOT EXISTS syndicate_id uuid references syndicates(id);

CREATE INDEX IF NOT EXISTS idx_player_syndicate on players(syndicate_id);

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
    'CHANGE_ENTRY_FEE',
    'CHANGE_MONTHLY_DUES',
    'CHANGE_BATTLE_WIN_CUT',
    'ADD_RULE',
    'REMOVE_RULE',
    'CHANGE_RULE',
    'REMOVE_MEMBER',

    'APPOINT_COMMITTEE',
    'REMOVE_COMMITTEE',

    'ADMIN_ELECTION', -- das exclusive
    'DEPOSE_ADMIN',

    -- boarder director exclusive
    'APPOINT_DIRECTOR',
    'REMOVE_DIRECTOR',
    'DEPOSE_CEO',
    'CEO_ELECTION'
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
    old_symbol text,
    new_symbol text,

    old_syndicate_name text,
    new_syndicate_name text,

    old_logo_id uuid references blobs(id),
    new_logo_id uuid references blobs(id),

    -- payment change
    old_join_fee numeric(28),
    new_join_fee numeric(28),

    old_exit_fee numeric(28),
    new_exit_fee numeric(28),

    old_monthly_dues numeric(28),
    new_monthly_dues numeric(28),

    old_deploying_member_cut_percentage decimal,
    new_deploying_member_cut_percentage decimal,

    old_member_assist_cut_percentage decimal,
    new_member_assist_cut_percentage decimal,

    old_mech_owner_cut_percentage decimal,
    new_mech_owner_cut_percentage decimal,

    old_syndicate_cut_percentage decimal,
    new_syndicate_cut_percentage decimal,

    -- add/remove/change rule
    rule_id uuid references syndicate_rules(id),
    old_rule_number int,
    new_rule_number int,
    old_rule_content text,
    new_rule_content text,

    -- appoint/remove member/committee/director
    member_id uuid references players(id),

    result SYNDICATE_MOTION_RESULT,
    note text,

    ended_at timestamptz not null,
    actual_ended_at timestamptz,
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_syndicate_motion_created_at_descending on syndicate_motions(created_at desc);
CREATE INDEX IF NOT EXISTS idx_syndicate_motion_type on syndicate_motions(type);
CREATE INDEX IF NOT EXISTS idx_syndicate_motion_syndicate_id on syndicate_motions(syndicate_id);

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

DROP TYPE IF EXISTS SYNDICATE_ELECTION_TYPE;
CREATE TYPE SYNDICATE_ELECTION_TYPE AS ENUM (
    'ADMIN', -- das exclusive
    'CEO' -- board of directors exclusive
);

CREATE TABLE syndicate_elections(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    syndicate_id uuid not null references syndicates(id),
    type SYNDICATE_ELECTION_TYPE not null,
    parent_election_id uuid references syndicate_elections(id), -- if tie result, auto start another election
    winner_id uuid references players(id),
    started_at timestamptz not null,
    end_at timestamptz not null,
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE TABLE syndicate_election_candidates(
    syndicate_election_id uuid not null references syndicate_elections(id),
    candidate_id uuid not null references players(id),
    PRIMARY KEY (syndicate_election_id, candidate_id),

    syndicate_id uuid not null references syndicates(id),
    resigned_at timestamptz,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE TABLE syndicate_election_votes(
    syndicate_election_id uuid not null references syndicate_elections(id),
    voter_id uuid not null references players(id),
    PRIMARY KEY (syndicate_election_id, voter_id),

    voted_for_candidate_id uuid not null references players(id),
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

ALTER TABLE blobs
    ADD COLUMN IF NOT EXISTS is_remote boolean NOT NULL DEFAULT FALSE,
    ALTER "file" DROP NOT NULL;
