ALTER TABLE mechs
    ADD COLUMN IF NOT EXISTS is_insured bool not null default false;

DROP TABLE asset_repair;

CREATE TABLE asset_repair(
    id uuid primary key DEFAULT gen_random_uuid(),
    mech_id UUID NOT NULL REFERENCES mechs (id),
    repair_complete_at timestamptz NOT NULL,
    full_repair_fee numeric(28) NOT NULL,
    pay_to_repair_tx_id TEXT,
    created_at timestamptz NOT NULL DEFAULT NOW()
);