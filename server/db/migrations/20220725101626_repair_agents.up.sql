CREATE TABLE repair_offers(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    offered_mech_id uuid not null references mechs(id),
    offered_by_id uuid not null references players(id),
    offered_amount numeric(28) not null,
    expire_at timestamptz not null, -- full recover duration
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

CREATE INDEX idx_repair_offer_offered_mech_id ON repair_offers(offered_mech_id);

DROP TYPE IF EXISTS REPAIR_AGENT_STATUS;
CREATE TYPE REPAIR_AGENT_STATUS AS ENUM ('WIP', 'CANCELED', 'FAILED', 'SUCCESS');

CREATE TABLE repair_agents(
    repair_offer_id uuid not null references repair_offers(id),
    agent_id uuid not null references players(id),
    PRIMARY KEY (repair_offer_id, agent_id),
    status REPAIR_AGENT_STATUS NOT NULL DEFAULT 'WIP',
    progress int not null default 0,
    complete_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);



drop index if exists idx_mech_repair_case_search;
drop index if exists idx_mech_repair_case_ended_at;

ALTER TABLE mech_repair_cases
    DROP COLUMN IF EXISTS fee,
    DROP COLUMN IF EXISTS fast_repair_fee,
    DROP COLUMN IF EXISTS fast_repair_tx_id,
    DROP COLUMN IF EXISTS repair_period_minutes,
    DROP COLUMN IF EXISTS max_health,
    DROP COLUMN IF EXISTS remain_health,
    DROP COLUMN IF EXISTS started_at,
    DROP COLUMN IF EXISTS expected_end_at,
    DROP COLUMN IF EXISTS ended_at,
    DROP COLUMN IF EXISTS status,
    ADD COLUMN IF NOT EXISTS fully_recover_at timestamptz not null,
    ADD COLUMN IF NOT EXISTS recovered_at timestamptz,
    ADD COLUMN IF NOT EXISTS repair_offer_id uuid references repair_offers(id);

create index idx_mech_repair_case_recovered_at on mech_repair_cases(recovered_at);
create index idx_mech_repair_case_created_at_descending on mech_repair_cases(created_at DESC);


ALTER TYPE mech_repair_log_type add value 'OFFER_REPAIR_JOB';
ALTER TYPE mech_repair_log_type add value 'INSTANT_REPAIR';
DROP TYPE mech_repair_status;