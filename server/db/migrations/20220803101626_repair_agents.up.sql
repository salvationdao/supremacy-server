DROP INDEX IF EXISTS idx_mech_repair_case_search;
DROP INDEX IF EXISTS idx_mech_repair_case_ended_at;
DROP INDEX IF EXISTS idx_mech_repair_case_mech_id;
DROP INDEX IF EXISTS idx_mech_repair_log_created_at_descending;
DROP INDEX IF EXISTS idx_mech_repair_log_mech_id;
DROP INDEX IF EXISTS idx_mech_repair_log_type;
DROP INDEX IF EXISTS idx_mech_repair_log_search;
DROP TABLE IF EXISTS mech_repair_logs;
DROP TABLE IF EXISTS mech_repair_cases;
DROP TABLE IF EXISTS repair_cases;
DROP TYPE IF EXISTS mech_repair_status;
DROP TYPE IF EXISTS MECH_REPAIR_LOG_TYPE;

ALTER TABLE blueprint_mechs
    ADD COLUMN IF NOT EXISTS repair_blocks INT NOT NULL DEFAULT 20;

ALTER TABLE weapon_models
    ADD COLUMN IF NOT EXISTS repair_blocks INT NOT NULL DEFAULT 20;

CREATE TABLE repair_cases(
    id uuid primary key default gen_random_uuid(),
    mech_id uuid not null references mechs(id),
    -- set after player click repair, used for recording
    blocks_required_repair integer not null,
    blocks_repaired integer not null default 0,
    completed_at timestamptz,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz,
    constraint repair_case_blocks_total_gt_zero check (blocks_required_repair > 0),
    constraint repair_case_blocks_repaired_gte_zero check (blocks_repaired >= 0),
    constraint repair_case_blocks_repaired_lte_required_blocks check (blocks_repaired <= blocks_required_repair)
);

CREATE INDEX idx_repair_cases_mech_id on repair_cases(mech_id);
CREATE INDEX idx_repair_cases_search on repair_cases(mech_id, completed_at);

DROP TYPE IF EXISTS REPAIR_FINISH_REASON;
CREATE TYPE REPAIR_FINISH_REASON AS ENUM ('EXPIRED', 'STOPPED', 'SUCCEEDED');

CREATE TABLE repair_offers(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repair_case_id uuid not null references repair_cases(id),
    offered_by_id uuid references players(id),
    blocks_total integer not null,
    offered_sups_amount numeric(28) not null, -- how much player offer for the entire repair offer
    expires_at timestamptz not null,
    finished_reason REPAIR_FINISH_REASON null,
    closed_at timestamptz,
    paid_tx_id text,
    tax_tx_id text,
    refund_tx_id text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz,
    constraint repair_offer_total_blocks_gt_zero check (blocks_total > 0)
);

CREATE INDEX idx_repair_offers_repair_case_id on repair_offers(repair_case_id);
CREATE INDEX idx_repair_offers_offered_by_id on repair_offers(offered_by_id);
CREATE INDEX idx_repair_offers_expired_search on repair_offers(expires_at desc, closed_at);
CREATE INDEX idx_repair_offers_search on repair_offers(repair_case_id,offered_by_id, closed_at);

DROP TYPE IF EXISTS REPAIR_AGENT_FINISH_REASON;
CREATE TYPE REPAIR_AGENT_FINISH_REASON AS ENUM ('ABANDONED', 'EXPIRED', 'SUCCEEDED');

CREATE TABLE repair_agents(
    id uuid primary key default gen_random_uuid(),
    repair_case_id uuid not null references repair_cases(id),
    repair_offer_id uuid not null references repair_offers(id),
    player_id uuid not null references players(id),
    required_stacks int not null,

    started_at timestamptz not null default now(),
    finished_at timestamptz null,
    finished_reason REPAIR_AGENT_FINISH_REASON null,

    payout_tx_id text,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

CREATE INDEX idx_repair_agent_search on repair_agents(player_id, finished_at);
CREATE INDEX idx_repair_agent_offered_by_id on repair_agents(repair_offer_id);

CREATE TABLE repair_blocks(
    id uuid primary key default gen_random_uuid(),
    repair_case_id UUID not null references repair_cases(id),
    repair_offer_id UUID not null references repair_offers(id),
    repair_agent_id UUID not null references repair_agents(id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

CREATE INDEX idx_repair_blocks_repair_case_id on repair_blocks(repair_case_id);
CREATE INDEX idx_repair_blocks_repair_offer_id on repair_blocks(repair_offer_id);
CREATE INDEX idx_repair_blocks_repair_agent_id on repair_blocks(repair_agent_id);

-- repair block trigger
CREATE OR REPLACE FUNCTION check_repair_block() RETURNS TRIGGER AS
$check_repair_block$
DECLARE
    can_write_block BOOLEAN DEFAULT FALSE;
BEGIN

SELECT (
           SELECT ro.expires_at > NOW() AND ro.closed_at IS NULL AND
                  ro.deleted_at IS NULL AND
                  rc.completed_at IS NULL AND
                  (SELECT COUNT(*) FROM repair_blocks rb WHERE rb.repair_case_id = rc.id) < rc.blocks_required_repair
           FROM repair_offers ro
           INNER JOIN repair_cases rc ON ro.repair_case_id = rc.id
           WHERE ro.id = NEW.repair_offer_id
       )
INTO can_write_block;
-- update blocks required in repair cases and continue the process
IF can_write_block THEN
    UPDATE repair_cases SET blocks_repaired = blocks_repaired + 1 WHERE id = NEW.repair_case_id;
    UPDATE repair_agents SET finished_at = now(), finished_reason = 'SUCCEEDED' WHERE id = NEW.repair_agent_id;
    RETURN NEW;
ELSE
    RAISE EXCEPTION 'unable to write block';
END IF;
END
$check_repair_block$
    LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_check_repair_block ON repair_blocks;

CREATE TRIGGER trigger_check_repair_block
    BEFORE INSERT
    ON repair_blocks
    FOR EACH ROW
EXECUTE PROCEDURE check_repair_block();

-- repair agent check
CREATE OR REPLACE FUNCTION check_repair_agent() RETURNS TRIGGER AS
$check_repair_agent$
DECLARE
    can_register BOOLEAN DEFAULT FALSE;
BEGIN

SELECT (
           SELECT ro.expires_at > NOW() AND ro.closed_at IS NULL AND
                  ro.deleted_at IS NULL AND
                  rc.completed_at IS NULL AND
                  (SELECT COUNT(*) FROM repair_blocks rb WHERE rb.repair_case_id = rc.id) < rc.blocks_required_repair
           FROM repair_offers ro
                    INNER JOIN repair_cases rc ON ro.repair_case_id = rc.id
           WHERE ro.id = NEW.repair_offer_id
       )
INTO can_register;
-- update blocks required in repair cases and continue the process
IF can_register THEN
    RETURN NEW;
ELSE
    RAISE EXCEPTION 'unable to register repair agent';
END IF;
END
$check_repair_agent$
    LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_check_repair_agent ON repair_agents;

CREATE TRIGGER trigger_check_repair_agent
    BEFORE INSERT
    ON repair_agents
    FOR EACH ROW
EXECUTE PROCEDURE check_repair_agent();

DROP TYPE IF EXISTS REPAIR_TRIGGER_WITH_TYPE;
CREATE TYPE REPAIR_TRIGGER_WITH_TYPE AS ENUM ('SPACE_BAR', 'LEFT_CLICK', 'TOUCH','NONE');

CREATE TABLE repair_agent_logs(
    id UUID PRIMARY KEY default gen_random_uuid(),
    repair_agent_id uuid not null references repair_agents (id),
    triggered_with REPAIR_TRIGGER_WITH_TYPE NOT NULL,
    score int not null,
    block_width decimal not null,
    block_depth decimal not null,
    is_failed bool not null default false,
    created_at timestamptz not null default now()
);

CREATE INDEX idx_repair_agent_logs_passed_records ON repair_agent_logs(repair_agent_id, is_failed);
CREATE INDEX idx_repair_agent_logs_created_at ON repair_agent_logs(created_at);

CREATE TABLE battle_queue_fees(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mech_id uuid not null references mechs(id),
    paid_by_id uuid not null references players(id),
    amount numeric(28) not null,

    paid_tx_id text,
    payout_tx_id text,
    tax_tx_id text,
    challenge_fund_tx_id text,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

ALTER TABLE battle_queue
    ADD COLUMN IF NOT EXISTS fee_id uuid references battle_queue_fees(id),
    DROP COLUMN IF EXISTS battle_contract_id;

CREATE INDEX idx_player_abilities_owner_id ON player_abilities(owner_id);