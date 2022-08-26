UPDATE
    blueprint_player_abilities
SET
    location_select_type = 'LOCATION_SELECT'
WHERE
        game_client_ability_id = 13;

UPDATE
    blueprint_player_abilities
SET
    location_select_type = 'MECH_SELECT'
WHERE
        game_client_ability_id = 14 OR game_client_ability_id = 15;
