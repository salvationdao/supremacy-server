DROP TYPE IF EXISTS ABILITY_KILLING_POWER_LEVEL;
CREATE TYPE ABILITY_KILLING_POWER_LEVEL AS ENUM ( 'DEADLY','NORMAL','NONE' );

ALTER TABLE battle_abilities
    ADD COLUMN IF NOT EXISTS maximum_commander_count int not null default 1, -- fix dev data sync
    ADD COLUMN IF NOT EXISTS deleted_at timestamptz, -- fix dev data sync
    ADD COLUMN IF NOT EXISTS killing_power_level ability_killing_power_level not null DEFAULT 'NORMAL';

ALTER TABLE game_abilities
    ADD COLUMN IF NOT EXISTS deleted_at timestamptz, -- fix dev data sync
    ADD COLUMN IF NOT EXISTS launching_delay_seconds int not null DEFAULT 0;