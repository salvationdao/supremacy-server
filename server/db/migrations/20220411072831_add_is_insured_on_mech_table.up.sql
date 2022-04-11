ALTER TABLE mechs
    ADD COLUMN IF NOT EXISTS is_insured bool not null default false;

ALTER TABLE asset_repair
    ADD COLUMN IF NOT EXISTS full_repair_fee numeric(78,0) NOT NULL DEFAULT 0;