ALTER TABLE battle_history
    DROP COLUMN IF EXISTS player_ability_offering_id;

ALTER TABLE battle_history
    RENAME COLUMN battle_ability_offering_id TO related_id;

DROP INDEX IF EXISTS idx_consumed_ability_offering_id;

ALTER TABLE consumed_abilities
    DROP COLUMN IF EXISTS offering_id;

