ALTER TABLE battle_abilities
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ; -- fix dev data sync
