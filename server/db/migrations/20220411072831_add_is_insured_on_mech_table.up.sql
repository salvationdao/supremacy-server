ALTER TABLE mechs
    ADD COLUMN IF NOT EXISTS is_insured bool not null default false;

DROP TABLE asset_repair;

CREATE TABLE asset_repair(
    id uuid primary key DEFAULT gen_random_uuid(),
    mech_id UUID NOT NULL REFERENCES mechs (id),
    repair_mode TEXT NOT NULL DEFAULT 'STANDARD',
    complete_until timestamptz NOT NULL,
    full_repair_fee numeric(78) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW()
);
