BEGIN;
ALTER TYPE location_select_type_enum ADD VALUE IF NOT EXISTS 'MECH_SELECT_ALLIED';
ALTER TYPE location_select_type_enum ADD VALUE IF NOT EXISTS 'MECH_SELECT_OPPONENT';
COMMIT;

ALTER TABLE blueprint_player_abilities
    ADD COLUMN IF NOT EXISTS mini_map_display_effect_type MINI_MAP_DISPLAY_EFFECT_TYPE NOT NULL DEFAULT 'NONE';

UPDATE
    blueprint_player_abilities
SET
    location_select_type = 'MECH_SELECT_OPPONENT',
    mini_map_display_effect_type = 'MECH_BORDER'
WHERE
    game_client_ability_id = 13;

UPDATE
    blueprint_player_abilities
SET
    location_select_type = 'MECH_SELECT_ALLIED'
WHERE
    game_client_ability_id = 14 OR game_client_ability_id = 15;