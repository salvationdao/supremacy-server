ALTER TABLE blueprint_player_abilities
    ADD COLUMN IF NOT EXISTS deleted_at                   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS display_on_mini_map          BOOL                         NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS launching_delay_seconds      INT                          NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS mini_map_display_effect_type MINI_MAP_DISPLAY_EFFECT_TYPE NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS mech_display_effect_type     MINI_MAP_DISPLAY_EFFECT_TYPE NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS animation_duration_seconds   INT                          NOT NULL DEFAULT 0;