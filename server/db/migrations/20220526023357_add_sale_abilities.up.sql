

INSERT INTO sale_player_abilities (blueprint_id, current_price, available_until)
VALUES ((SELECT id FROM blueprint_player_abilities bpa WHERE bpa.game_client_ability_id = 10), 100000000000000000000,
        NOW()),
       ((SELECT id FROM blueprint_player_abilities bpa WHERE bpa.game_client_ability_id = 11), 100000000000000000000,
        NOW()),
       ((SELECT id FROM blueprint_player_abilities bpa WHERE bpa.game_client_ability_id = 12), 100000000000000000000,
        NOW()),
       ((SELECT id FROM blueprint_player_abilities bpa WHERE bpa.game_client_ability_id = 13), 100000000000000000000,
        NOW()),
       ((SELECT id FROM blueprint_player_abilities bpa WHERE bpa.game_client_ability_id = 14), 100000000000000000000,
        NOW()),
       ((SELECT id FROM blueprint_player_abilities bpa WHERE bpa.game_client_ability_id = 15), 100000000000000000000,
        NOW()),
       ((SELECT id FROM blueprint_player_abilities bpa WHERE bpa.game_client_ability_id = 16), 100000000000000000000,
        NOW());
