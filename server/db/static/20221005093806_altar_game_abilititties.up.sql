ALTER TABLE game_abilities
    DROP COLUMN IF EXISTS sups_cost,
    DROP COLUMN IF EXISTS current_sups,
    ADD COLUMN count_per_battle INT DEFAULT 0 NOT NULL;
