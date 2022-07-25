DROP INDEX IF EXISTS idx_mech_repair_case_search;
DROP INDEX IF EXISTS idx_mech_repair_case_ended_at;
DROP INDEX IF EXISTS idx_mech_repair_case_mech_id;
DROP INDEX IF EXISTS idx_mech_repair_log_created_at_descending;
DROP INDEX IF EXISTS idx_mech_repair_log_mech_id;
DROP INDEX IF EXISTS idx_mech_repair_log_type;
DROP INDEX IF EXISTS idx_mech_repair_log_search;
DROP TABLE IF EXISTS mech_repair_logs;
DROP TABLE IF EXISTS mech_repair_cases;
DROP TYPE IF EXISTS mech_repair_status;
DROP TYPE IF EXISTS MECH_REPAIR_LOG_TYPE;

CREATE TABLE mech_repair_cases(
    id uuid primary key default gen_random_uuid(),
    mech_id uuid not null references mechs(id),

    required_seconds int not null,
    default_instant_repair_fee numeric(28) not null,

    -- set after player click repair, used for recording
    started_at timestamptz,
    default_complete_time timestamptz,

    completed_at timestamptz,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

CREATE TABLE repair_offers(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mech_repair_case_id uuid not null references mech_repair_cases(id),
    repairing_mech_id uuid not null references mechs(id),
    offered_by_id uuid not null references players(id),
    offered_sups_amount numeric(28) not null, -- how much player offer for the entire repair offer
    sups_worth_per_hour numeric(28) not null, -- pre-calculated
    paid_amount numeric(28) not null, -- sups that already paid
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

CREATE INDEX idx_repair_offer_repairing_mech_id ON repair_offers(repairing_mech_id);

DROP TYPE IF EXISTS REPAIR_AGENT_STATUS;
CREATE TYPE REPAIR_AGENT_STATUS AS ENUM ('WIP', 'FAILED', 'SUCCESS');

CREATE TABLE repair_agents(
    id uuid primary key default gen_random_uuid(),
    mech_repair_case_id uuid not null references mech_repair_cases(id),
    repair_offer_id uuid not null references repair_offers(id),
    agent_id uuid not null references players(id),

    status REPAIR_AGENT_STATUS NOT NULL DEFAULT 'WIP',
    repair_code uuid not null default gen_random_uuid(),
    started_at timestamptz not null default now(),
    ended_at timestamptz,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

DROP TYPE IF EXISTS MECH_REPAIR_CASE_LOG_TYPE;
CREATE TYPE MECH_REPAIR_CASE_LOG_TYPE AS ENUM (
    'REGISTER',
    'START_REPAIR_PROCESS',
    'OFFER_REPAIR_CONTRACT',
    'REPAIR_AGENT_COMPLETE',
    'COMPLETE'
);

CREATE TABLE mech_repair_case_logs(
    id uuid primary key default gen_random_uuid(),
    mech_id uuid not null references mechs(id),
    mech_repair_case_id uuid not null references mech_repair_cases(id),
    type MECH_REPAIR_CASE_LOG_TYPE not null,
    repair_offer_id uuid references repair_offers(id),
    repair_agent_id uuid references repair_agents(id),
    created_at timestamptz not null default now()
);

create index idx_mech_repair_case_logs_created_at_descending on mech_repair_case_logs(created_at desc);
create index idx_mech_repair_case_logs_mech_id on mech_repair_case_logs(mech_repair_case_id);
