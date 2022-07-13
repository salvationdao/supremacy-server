UPDATE
    blueprint_player_abilities
SET
    location_select_type = 'LINE_SELECT'
WHERE
    game_client_ability_id = 0;

UPDATE
    consumed_abilities
SET
    location_select_type = 'LINE_SELECT'
WHERE
    game_client_ability_id = 0;

UPDATE
    game_abilities
SET
    location_select_type = 'LINE_SELECT'
WHERE
    game_client_ability_id = 0;