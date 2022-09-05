CREATE TABLE IF NOT EXISTS blueprint_player_abilities
(
    id                           UUID PRIMARY KEY             NOT NULL DEFAULT gen_random_uuid(),
    game_client_ability_id       INT4                         NOT NULL,
    label                        TEXT                         NOT NULL,
    colour                       TEXT                         NOT NULL,
    image_url                    TEXT                         NOT NULL,
    description                  TEXT                         NOT NULL,
    text_colour                  TEXT                         NOT NULL,
    location_select_type         TEXT                         NOT NULL CHECK (location_select_type IN ('MECH_SELECT', 'LOCATION_SELECT', 'GLOBAL')),
    created_at                   TIMESTAMPTZ                  NOT NULL DEFAULT NOW(),
    rarity_weight                INT                          NOT NULL DEFAULT -1,
    inventory_limit              INT                          NOT NULL DEFAULT 1,
    cooldown_seconds             INT                          NOT NULL DEFAULT 180,
    display_on_mini_map          BOOL                         NOT NULL DEFAULT FALSE,
    launching_delay_seconds      INT                          NOT NULL DEFAULT 0,
    mini_map_display_effect_type MINI_MAP_DISPLAY_EFFECT_TYPE NOT NULL DEFAULT 'NONE',
    mech_display_effect_type     MINI_MAP_DISPLAY_EFFECT_TYPE NOT NULL DEFAULT 'NONE',
    animation_duration_seconds   INT                          NOT NULL DEFAULT 0
);