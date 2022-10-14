BEGIN;
ALTER TYPE MINI_MAP_DISPLAY_EFFECT_TYPE ADD VALUE IF NOT EXISTS 'PULSE';
ALTER TYPE MINI_MAP_DISPLAY_EFFECT_TYPE ADD VALUE IF NOT EXISTS 'BORDER';
ALTER TYPE MINI_MAP_DISPLAY_EFFECT_TYPE ADD VALUE IF NOT EXISTS 'DROP';
ALTER TYPE MINI_MAP_DISPLAY_EFFECT_TYPE ADD VALUE IF NOT EXISTS 'SHAKE';
COMMIT;

ALTER TABLE game_abilities
    ADD COLUMN IF NOT EXISTS mech_display_effect_type MINI_MAP_DISPLAY_EFFECT_TYPE NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS animation_duration_seconds int  NOT NULL DEFAULT 0;