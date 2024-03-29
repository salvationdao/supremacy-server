DROP TYPE IF EXISTS MINI_MAP_DISPLAY_EFFECT_TYPE;
CREATE TYPE MINI_MAP_DISPLAY_EFFECT_TYPE AS ENUM ( 'NONE','RANGE','MECH_PULSE','MECH_BORDER');

ALTER TABLE game_abilities
    ADD COLUMN IF NOT EXISTS display_on_mini_map   BOOL                         NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS mini_map_display_effect_type MINI_MAP_DISPLAY_EFFECT_TYPE NOT NULL DEFAULT 'NONE';