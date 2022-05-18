INSERT INTO blueprint_player_abilities (game_client_ability_id, "label", colour, image_url, description, text_colour,
                                        location_select_type)
VALUES (11, 'Landmine', '#505944', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-landmine.png',
        'A device that explodes when a War Machine is detected within its proximity.', '#505944', 'LOCATION_SELECT'),
       (12, 'EMP', '#2e8cff',
        'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-emp.png',
        'A short burst of electromagnetic energy that will disrupt War Machine operations in an area for a brief period of time.',
        '#2e8cff', 'LOCATION_SELECT'),
       (10, 'Shield Overdrive', '#30ddff',
        'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-shield-buff.png',
        'An airdropped module that can be picked up by War Machines to boost put their shields into overdrive.',
        '#30ddff',
        'LOCATION_SELECT'),
       (16, 'Blackout', '#112124', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-blackout.png',
        'Drops a cloud of nanorobotics, disabling War Machines locational and navigation.', '#112124',
        'LOCATION_SELECT'),
       (13, 'Hacker Drone', '#2da850',
        'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-hacker-drone.png',
        'Hacks into the nearest War Machine disrupting their targeting systems.', '#2da850', 'MECH_SELECT'),
       (14, 'Drone Camera', '#5f6b63',
        'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-drone-camera.png',
        'Overrides GABs visual broadcasting for a short period.', '#5f6b63', 'MECH_SELECT'),
       (15, 'Incognito', '#30303b',
        'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-incognito.png',
        'Blocks GABs radar technology from finding its position, hiding it from the minimap.', '#30303b',
        'MECH_SELECT');

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


CREATE INDEX IF NOT EXISTS user_multi_index ON user_multipliers (player_id, from_battle_number, until_battle_number)
