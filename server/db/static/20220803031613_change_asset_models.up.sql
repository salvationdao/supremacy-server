ALTER TABLE mech_models
    ADD COLUMN IF NOT EXISTS repair_blocks INT NOT NULL DEFAULT 20;

ALTER TABLE weapon_models
    ADD COLUMN IF NOT EXISTS repair_blocks INT NOT NULL DEFAULT 20;