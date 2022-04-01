ALTER TABLE mechs add collection_slug TEXT;

UPDATE mechs SET collection_slug = 'supremacy-genesis' WHERE is_default = false AND "label" NOT ILIKE '%ukraini%';
UPDATE mechs SET collection_slug = 'supremacy-ai' WHERE is_default = true;

ALTER TABLE templates add collection_slug text;

UPDATE templates SET collection_slug = 'supremacy-genesis' WHERE is_default = false AND "label" NOT ILIKE '%ukraini%';
UPDATE templates SET collection_slug = 'supremacy-ai' WHERE is_default  = true;

INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints, turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield) VALUES ('88f4be7c-20f6-48a9-9d56-8f34ebc4607c', '009f71fc-3594-4d24-a6e2-f05070d66f40', 'Law Enforcer X-1000 Slava Ukraini Chassis', 'boston_cybernetics_law_enforcer_x_1000_slava-ukraini_chassis', 'Slava Ukraini', 'Law Enforcer X-1000', 80, 2, 1, 1, 2860, 1000, 1000);

INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints, turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield) VALUES ('80bf637e-9219-44fc-9907-1f58598d6800', '2b203c87-ad8c-4ce2-af17-e079835fdbcb', 'Zaibatsu Tenshi Mk1 Slava Ukraini Chassis', 'zaibatsu_tenshi_mk1_slava-ukraini_chassis', 'Slava Ukraini', 'Tenshi Mk1', 102, 2, 2, 1, 2500, 1000, 1000);

UPDATE templates SET image_url = 'https://afiles.ninja-cdn.com/passport/limited/img/red-mountain_olympus-mons-ly07_slava-ukraini.png', animation_url = 'https://afiles.ninja-cdn.com/passport/limited/opensea/mp4/red-mountain_olympus-mons-ly07_slava-ukraini.mp4', large_image_url = 'https://afiles.ninja-cdn.com/passport/limited/opensea/img/red-mountain_olympus-mons-ly07_slava-ukraini.png', card_animation_url = 'https://afiles.ninja-cdn.com/passport/limited/webm/red-mountain_olympus-mons-ly07_slava-ukraini.webm', avatar_url = 'https://afiles.ninja-cdn.com/passport/limited/avatar/red-mountain_olympus-mons-ly07_slava-ukraini_avatar.png', collection_slug = 'supremacy-limited-release', tier = 'LEGENDARY' WHERE "label" ILIKE '%ukraini%';

INSERT INTO templates (id, faction_id, tier, image_url, animation_url, card_animation_url, avatar_url, label, slug, is_default, asset_type, large_image_url, collection_slug, blueprint_chassis_id) VALUES ('cb321fe6-8583-4749-9e6a-841bbf6aeb47','7c6dde21-b067-46cf-9e56-155c88a520e2', 'LEGENDARY', 'https://afiles.ninja-cdn.com/passport/limited/img/boston_cybernetics_law_enforcer_x_1000_slava-ukraini.png', 'https://afiles.ninja-cdn.com/passport/limited/opensea/mp4/boston_cybernetics_law_enforcer_x_1000_slava-ukraini.mp4', 'https://afiles.ninja-cdn.com/passport/limited/webm/boston_cybernetics_law_enforcer_x_1000_slava-ukraini.webm', 'https://afiles.ninja-cdn.com/passport/limited/avatar/boston_cybernetics_law_enforcer_x_1000_slava-ukraini_avatar.png', 'Boston Cybernetics Law Enforcer X-1000 Slava Ukraini', 'boston_cybernetics_law_enforcer_x_1000_slava-ukraini', false, 'War Machine', 'https://afiles.ninja-cdn.com/passport/limited/opensea/img/boston_cybernetics_law_enforcer_x_1000_slava-ukraini.png', 'supremacy-limited-release', '88f4be7c-20f6-48a9-9d56-8f34ebc4607c');

INSERT INTO templates (id, faction_id, tier, image_url, animation_url, card_animation_url, avatar_url, label, slug, is_default, asset_type, large_image_url, collection_slug, blueprint_chassis_id) VALUES ('981192c8-8918-4ef2-853f-664586499983','880db344-e405-428d-84e5-6ebebab1fe6d', 'LEGENDARY', 'https://afiles.ninja-cdn.com/passport/limited/img/zaibatsu_tenshi_mk1_slava-ukraini.png', 'https://afiles.ninja-cdn.com/passport/limited/opensea/mp4/zaibatsu_tenshi_mk1_slava-ukraini.mp4', 'https://afiles.ninja-cdn.com/passport/limited/webm/zaibatsu_tenshi_mk1_slava-ukraini.webm', 'https://afiles.ninja-cdn.com/passport/limited/avatar/zaibatsu_tenshi_mk1_slava-ukraini_avatar.png', 'Zaibatsu Tenshi Mk1 Slava Ukraini', 'zaibatsu_tenshi_mk1_slava-ukraini', false, 'War Machine', 'https://afiles.ninja-cdn.com/passport/limited/opensea/img/zaibatsu_tenshi_mk1_slava-ukraini.png', 'supremacy-limited-release', '80bf637e-9219-44fc-9907-1f58598d6800');

INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number) VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '88f4be7c-20f6-48a9-9d56-8f34ebc4607c', 0);
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location) VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '88f4be7c-20f6-48a9-9d56-8f34ebc4607c', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location) VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '88f4be7c-20f6-48a9-9d56-8f34ebc4607c', 1, 'ARM');

INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number) VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '80bf637e-9219-44fc-9907-1f58598d6800', 0);
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location) VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', '80bf637e-9219-44fc-9907-1f58598d6800', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location) VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', '80bf637e-9219-44fc-9907-1f58598d6800', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location) VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '80bf637e-9219-44fc-9907-1f58598d6800', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location) VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '80bf637e-9219-44fc-9907-1f58598d6800', 1, 'TURRET');

