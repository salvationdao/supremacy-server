ALTER TABLE blueprint_player_abilities
    ADD COLUMN IF NOT EXISTS mech_display_effect_type MINI_MAP_DISPLAY_EFFECT_TYPE NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS animation_duration_seconds int  NOT NULL DEFAULT 0;

UPDATE blueprint_player_abilities
SET mini_map_display_effect_type = 'PULSE',
    animation_duration_seconds = 10
WHERE game_client_ability_id = 12; -- EMP

UPDATE blueprint_player_abilities
SET display_on_mini_map = false,
    mini_map_display_effect_type = 'NONE',
    mech_display_effect_type = 'BORDER'
WHERE game_client_ability_id = 13; -- Hacker drone