ALTER TABLE users
    ADD COLUMN IF NOT EXISTS sups_multipliers jsonb;