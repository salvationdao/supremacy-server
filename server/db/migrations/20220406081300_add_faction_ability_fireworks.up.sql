INSERT INTO
    game_abilities (game_client_ability_id, faction_id, label, sups_cost, colour, text_colour, description, image_url)
VALUES
    (15, '880db344-e405-428d-84e5-6ebebab1fe6d', 'FIREWORKS', '100000000000000000000', '#FFC524', '#000000', 'display firework on targeted mech', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-fireworks.jpg'),
    (16, '98bf7bb3-1a7c-4f21-8843-458d62884060', 'FIREWORKS', '100000000000000000000', '#FFC524', '#000000', 'display firework on targeted mech', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-fireworks.jpg'),
    (17, '7c6dde21-b067-46cf-9e56-155c88a520e2', 'FIREWORKS', '100000000000000000000', '#FFC524', '#000000', 'display firework on targeted mech', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-fireworks.jpg');

ALTER TYPE ABILITY_TYPE_ENUM ADD VALUE 'FIREWORKS';

UPDATE punish_options SET key = 'restrict_location_select', description = 'Restrict player to select location for 24 hours' WHERE key = 'limit_location_select';
UPDATE punish_options SET key = 'restrict_chat', description = 'Restrict player to chat for 24 hours' WHERE key = 'limit_chat';
UPDATE punish_options SET key = 'restrict_sups_contribution', description = 'Restrict player to contribute sups for 24 hours' WHERE key = 'limit_sups_contibution';