UPDATE  game_abilities SET level = 'PLAYER' WHERE  game_client_ability_id = 5 or game_client_ability_id = 2;

INSERT INTO game_abilities (game_client_ability_id, faction_id, battle_ability_id, label, colour, image_url, description, text_colour, level, location_select_type)
VALUES  (2, '7c6dde21-b067-46cf-9e56-155c88a520e2', null, 'REPAIR', '#23AE3C', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-repair.jpg', 'Support your Syndicate with a well-timed repair.', '#FFFFFF', 'PLAYER', 'LOCATION_SELECT'),
        (2, '98bf7bb3-1a7c-4f21-8843-458d62884060', null, 'REPAIR', '#23AE3C', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-repair.jpg', 'Support your Syndicate with a well-timed repair.', '#FFFFFF', 'PLAYER', 'LOCATION_SELECT'),
        (2, '880db344-e405-428d-84e5-6ebebab1fe6d', null, 'REPAIR', '#23AE3C', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-repair.jpg', 'Support your Syndicate with a well-timed repair.', '#FFFFFF', 'PLAYER', 'LOCATION_SELECT');
