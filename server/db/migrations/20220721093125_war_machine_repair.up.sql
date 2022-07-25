DROP TABLE IF EXISTS mech_repair;

DROP TYPE IF EXISTS MECH_REPAIR_STATUS;
CREATE TYPE MECH_REPAIR_STATUS AS ENUM ('PENDING', 'STANDARD_REPAIR', 'FAST_REPAIR');

CREATE TABLE mech_repair_cases(
    id uuid primary key default gen_random_uuid(),
    mech_id uuid not null references mechs(id),
    fee decimal not null,
    fast_repair_fee decimal not null,
    fast_repair_tx_id text,

    -- for calculate repair duration
    status MECH_REPAIR_STATUS not null default 'PENDING',
    repair_period_minutes int not null,
    max_health decimal not null,
    remain_health decimal not null,

    started_at timestamptz,
    expected_end_at timestamptz,
    ended_at timestamptz,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_mech_repair_case_search on mech_repair_cases(expected_end_at, ended_at);
create index idx_mech_repair_case_ended_at on mech_repair_cases(ended_at);
create index idx_mech_repair_case_mech_id on mech_repair_cases(mech_id);


DROP TYPE IF EXISTS MECH_REPAIR_LOG_TYPE;
CREATE TYPE MECH_REPAIR_LOG_TYPE AS ENUM ('REGISTER_REPAIR','START_STANDARD_REPAIR','START_FAST_REPAIR', 'SPEED_UP','REPAIR_ENDED');


CREATE TABLE mech_repair_logs(
    id uuid primary key default gen_random_uuid(),
    mech_id uuid not null references mechs(id),
    type MECH_REPAIR_LOG_TYPE not null,
    involved_player_id uuid references players(id),
    created_at timestamptz not null default now()
);

create index idx_mech_repair_log_created_at_descending on mech_repair_logs(created_at desc);
create index idx_mech_repair_log_mech_id on mech_repair_logs(mech_id);
create index idx_mech_repair_log_type on mech_repair_logs(type);
create index idx_mech_repair_log_search on mech_repair_logs(mech_id, type);