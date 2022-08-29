INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('f9ca786e-730d-481e-a6c9-f04e5ed975d2', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Slava Ukraini Chassis', 'red_mountain_olympus_mons_ly07_slava_ukraini_chassis',
        'Slava Ukraini', 'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1560, 1000);

INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('cf90467c-9155-425a-be34-dbf03542fba8', 'f9ca786e-730d-481e-a6c9-f04e5ed975d2',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_slava_ukraini.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_slava_ukraini.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_slava_ukraini.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_slava_ukraini_avatar.png',
        'Red Mountain Olympus Mons LY07 Slava Ukraini', 'red_mountain_olympus_mons_ly07_slava_ukraini', false,
        'War Machine');

INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'f9ca786e-730d-481e-a6c9-f04e5ed975d2', 0);
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('74005f3b-a6e2-4e4b-a59c-e07ff42cb800', 'f9ca786e-730d-481e-a6c9-f04e5ed975d2', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('74005f3b-a6e2-4e4b-a59c-e07ff42cb800', 'f9ca786e-730d-481e-a6c9-f04e5ed975d2', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('c1c78867-9de7-43d3-97e9-91381800f38e', 'f9ca786e-730d-481e-a6c9-f04e5ed975d2', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('c1c78867-9de7-43d3-97e9-91381800f38e', 'f9ca786e-730d-481e-a6c9-f04e5ed975d2', 1, 'TURRET');