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

UPDATE
    blueprint_player_abilities
SET
    colour = '#8638c9'
WHERE
    game_client_ability_id = 11;

UPDATE
    consumed_abilities
SET
    colour = '#8638c9'
WHERE
    game_client_ability_id = 11;

UPDATE
    game_abilities
SET
    colour = '#8638c9'
WHERE
    game_client_ability_id = 11;