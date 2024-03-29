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
    member_monthly_dues numeric(28) NOT NULL DEFAULT 0,

    -- battle win columns
    deploying_member_cut_percentage decimal NOT NULL DEFAULT 0,
    member_assist_cut_percentage decimal NOT NULL DEFAULT 0,
    mech_owner_cut_percentage decimal NOT NULL DEFAULT 0,

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

    'DEPOSE_ADMIN', -- das exclusive

    -- boarder director exclusive
    'APPOINT_DIRECTOR',
    'REMOVE_DIRECTOR',
    'DEPOSE_CEO'
);

DROP TYPE IF EXISTS SYNDICATE_MOTION_RESULT;
CREATE TYPE SYNDICATE_MOTION_RESULT AS ENUM (
    'PASSED',
    'FAILED',
    'TERMINATED',
    'LEADER_ACCEPTED',
    'LEADER_REJECTED'
);

CREATE TABLE syndicate_motions(
    id uuid primary key default gen_random_uuid(),
    syndicate_id uuid not null references syndicates(id),
    type SYNDICATE_MOTION_TYPE not null,
    issued_by_id uuid not null references players(id),
    end_at timestamptz not null,
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
    finalised_at timestamptz,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_syndicate_motion_created_at_descending on syndicate_motions(created_at desc);
CREATE INDEX IF NOT EXISTS idx_syndicate_motion_type on syndicate_motions(type);
CREATE INDEX IF NOT EXISTS idx_syndicate_motion_syndicate_id on syndicate_motions(syndicate_id);

CREATE TABLE syndicate_motion_votes(
    motion_id uuid not null references syndicate_motions(id),
    vote_by_id uuid not null references players(id),
    PRIMARY KEY (motion_id, vote_by_id),

    is_agreed bool not null,
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_motion_vote_motion_id on syndicate_motion_votes(motion_id);

-- this table is for corporation syndicate only, any passed motion should be accepted by ceo or admin
CREATE TABLE syndicate_pending_motions(
    id uuid primary key default gen_random_uuid(),
    syndicate_id uuid not null references syndicates(id),
    motion_id uuid not null references syndicate_motions(id),
    final_decision text,
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);


DROP TYPE IF EXISTS SYNDICATE_ELECTION_TYPE;
CREATE TYPE SYNDICATE_ELECTION_TYPE AS ENUM (
    'ADMIN', -- das exclusive
    'CEO' -- board of directors exclusive
);

DROP TYPE IF EXISTS SYNDICATE_ELECTION_RESULT;
CREATE TYPE SYNDICATE_ELECTION_RESULT AS ENUM (
    'WINNER_APPEAR',
    'TIE',
    'TIE_SECOND_TIME',
    'NO_VOTE',
    'NO_CANDIDATE',
    'TERMINATED'
);

CREATE TABLE syndicate_elections(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    syndicate_id uuid not null references syndicates(id),
    type SYNDICATE_ELECTION_TYPE not null,
    parent_election_id uuid references syndicate_elections(id), -- if tie result, auto start another election
    winner_id uuid references players(id),
    started_at timestamptz not null,
    candidate_register_close_at timestamptz not null,
    end_at timestamptz not null,
    finalised_at timestamptz,
    result SYNDICATE_ELECTION_RESULT,
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_syndicate_election_created_at_descending on syndicate_elections(created_at desc);
CREATE INDEX IF NOT EXISTS idx_syndicate_election_syndicate_id on syndicate_elections(syndicate_id);
CREATE INDEX IF NOT EXISTS idx_syndicate_election_search on syndicate_elections(syndicate_id, finalised_at);

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
    voted_by_id uuid not null references players(id),
    PRIMARY KEY (syndicate_election_id, voted_by_id),

    voted_for_candidate_id uuid not null references players(id),
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

DROP TYPE IF EXISTS QUESTIONNAIRE_TYPE;
CREATE TYPE QUESTIONNAIRE_TYPE AS ENUM (
    'TEXT',
    'SINGLE_SELECT',
    'MULTI_SELECT'
);

DROP TYPE IF EXISTS QUESTIONNAIRE_USAGE;
CREATE TYPE QUESTIONNAIRE_USAGE AS ENUM (
    'JOIN_REQUEST'
);

CREATE TABLE syndicate_questionnaires(
    id uuid primary key default gen_random_uuid(),
    syndicate_id uuid not null references syndicates(id),
    usage QUESTIONNAIRE_USAGE NOT NULL,

    number int not null,
    must_answer bool not null,
    question text not null,
    type QUESTIONNAIRE_TYPE not null,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE TABLE questionnaire_options(
    id uuid primary key default gen_random_uuid(),
    questionnaire_id uuid not null references syndicate_questionnaires(id),
    content text not null,
    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

DROP TYPE IF EXISTS SYNDICATE_JOIN_APPLICATION_RESULT;
CREATE TYPE SYNDICATE_JOIN_APPLICATION_RESULT AS ENUM (
    'ACCEPTED',
    'REJECTED',
    'TERMINATED'
);

CREATE TABLE syndicate_join_applications(
    id uuid primary key default gen_random_uuid(),
    syndicate_id uuid not null references syndicates(id),
    applicant_id uuid not null references players(id),
    expire_at timestamptz not null,
    paid_amount numeric(28) not null, -- record the amount of sups player paid upfront
    tx_id text,
    refund_tx_id text,

    result SYNDICATE_JOIN_APPLICATION_RESULT,
    note text,
    finalised_at timestamptz,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE TABLE questionnaire_answer(
    id uuid primary key default gen_random_uuid(),
    syndicate_join_application_id uuid references syndicate_join_applications(id),

    -- record question and answer players submitted
    question text not null,
    answer text,
    selections text[],

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

CREATE TABLE application_votes(
    application_id uuid not null references syndicate_join_applications(id),
    voted_by_id uuid not null references players(id),
    PRIMARY KEY (application_id, voted_by_id),
    is_agreed bool not null,

    created_at timestamptz not null default NOW(),
    updated_at timestamptz not null default NOW(),
    deleted_at timestamptz
);

ALTER TABLE blobs
    ADD COLUMN IF NOT EXISTS is_remote boolean NOT NULL DEFAULT FALSE,
    ALTER "file" DROP NOT NULL;
