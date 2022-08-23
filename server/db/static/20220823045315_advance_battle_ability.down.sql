ALTER TABLE battle_abilities
    DROP COLUMN IF EXISTS killing_power_level;

ALTER TABLE game_abilities
    DROP COLUMN IF EXISTS trigger_countdown_seconds;

DROP TYPE IF EXISTS ABILITY_KILLING_POWER_LEVEL;
