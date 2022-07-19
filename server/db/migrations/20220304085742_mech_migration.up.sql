CREATE TABLE players
(
    id             UUID PRIMARY KEY NOT NULL,
    faction_id     UUID REFERENCES factions (id),
    username       TEXT UNIQUE,
    public_address TEXT UNIQUE,
    is_ai          BOOL             NOT NULL DEFAULT FALSE,
    deleted_at     TIMESTAMPTZ,
    updated_at     TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at     TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

ALTER TABLE users
    ADD COLUMN player_id UUID REFERENCES players (id);

CREATE TABLE blueprint_chassis
(
    id                   UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    brand_id             UUID             NOT NULL REFERENCES brands (id),
    label                TEXT UNIQUE      NOT NULL,
    slug                 TEXT UNIQUE      NOT NULL,
    model                TEXT             NOT NULL,
    skin                 TEXT             NOT NULL,
    shield_recharge_rate INTEGER          NOT NULL,
    weapon_hardpoints    INTEGER          NOT NULL,
    turret_hardpoints    INTEGER          NOT NULL,
    utility_slots        INTEGER          NOT NULL,
    speed                INTEGER          NOT NULL,
    max_hitpoints        INTEGER          NOT NULL,
    max_shield           INTEGER          NOT NULL,

    deleted_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE templates
(
    id                   UUID PRIMARY KEY NOT NULL,
    blueprint_chassis_id UUID UNIQUE      NOT NULL REFERENCES blueprint_chassis (id),
    faction_id           UUID             NOT NULL REFERENCES factions (id),
    tier                 TEXT             NOT NULL CHECK (tier IN
                                                          ('MEGA', 'COLOSSAL', 'RARE', 'LEGENDARY', 'ELITE_LEGENDARY',
                                                           'ULTRA_RARE', 'EXOTIC', 'GUARDIAN', 'MYTHIC', 'DEUS_EX')),
    label                TEXT UNIQUE      NOT NULL,
    slug                 TEXT UNIQUE      NOT NULL,
    is_default           BOOLEAN          NOT NULL DEFAULT FALSE,
    image_url            TEXT             NOT NULL CHECK (image_url != ''),
    animation_url        TEXT             NOT NULL CHECK (animation_url != ''),
    card_animation_url   TEXT             NOT NULL CHECK (card_animation_url != ''),
    avatar_url           TEXT             NOT NULL CHECK (avatar_url != ''),
    asset_type           TEXT             NOT NULL CHECK (asset_type != ''),

    deleted_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

-- CREATE TABLE blueprint_weapons
-- (
--     id          UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
--     brand_id    UUID REFERENCES brands (id),
--
--     label       TEXT UNIQUE      NOT NULL,
--     slug        TEXT UNIQUE      NOT NULL,
--     damage      INTEGER          NOT NULL,
--     weapon_type TEXT             NOT NULL CHECK (weapon_type IN ('TURRET', 'ARM')),
--
--     deleted_at  TIMESTAMPTZ,
--     updated_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
--     created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW()
-- );

CREATE TABLE blueprint_modules
(
    id                UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    brand_id          UUID REFERENCES brands (id),

    slug              TEXT UNIQUE      NOT NULL,
    label             TEXT UNIQUE      NOT NULL,
    hitpoint_modifier INTEGER          NOT NULL,
    shield_modifier   INTEGER          NOT NULL,

    deleted_at        TIMESTAMPTZ,
    updated_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE blueprint_chassis_blueprint_weapons
(
    id                   UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    blueprint_weapon_id  UUID             NOT NULL REFERENCES blueprint_weapons (id),
    blueprint_chassis_id UUID             NOT NULL REFERENCES blueprint_chassis (id),
    slot_number          INTEGER          NOT NULL,
    mount_location       TEXT             NOT NULL CHECK (mount_location IN ('ARM', 'TURRET')),
    UNIQUE (blueprint_chassis_id, slot_number, mount_location),

    deleted_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE blueprint_chassis_blueprint_modules
(
    id                   UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),

    blueprint_module_id  UUID             NOT NULL REFERENCES blueprint_modules (id),
    blueprint_chassis_id UUID             NOT NULL REFERENCES blueprint_chassis (id),
    slot_number          INTEGER          NOT NULL,

    UNIQUE (blueprint_chassis_id, slot_number),

    deleted_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE chassis
(
    id                   UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    brand_id             UUID             NOT NULL REFERENCES brands (id),

    label                TEXT             NOT NULL,
    model                TEXT             NOT NULL,
    skin                 TEXT             NOT NULL,
    slug                 TEXT             NOT NULL,
    shield_recharge_rate INTEGER          NOT NULL,
    health_remaining     INTEGER          NOT NULL,
    weapon_hardpoints    INTEGER          NOT NULL,
    turret_hardpoints    INTEGER          NOT NULL,
    utility_slots        INTEGER          NOT NULL,
    speed                INTEGER          NOT NULL,
    max_hitpoints        INTEGER          NOT NULL,
    max_shield           INTEGER          NOT NULL,

    deleted_at           TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);


CREATE TABLE mechs
(
    id                 UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    owner_id           UUID             NOT NULL REFERENCES players (id),
    template_id        UUID             NOT NULL REFERENCES templates (id),
    chassis_id         UUID UNIQUE      NOT NULL REFERENCES chassis (id),
    external_token_id  INTEGER          NOT NULL,
    tier               TEXT             NOT NULL CHECK (tier IN
                                                        ('MEGA', 'COLOSSAL', 'RARE', 'LEGENDARY', 'ELITE_LEGENDARY',
                                                         'ULTRA_RARE', 'EXOTIC', 'GUARDIAN', 'MYTHIC', 'DEUS_EX')),
    is_default         BOOLEAN          NOT NULL DEFAULT FALSE,
    image_url          TEXT             NOT NULL CHECK (image_url != ''),
    animation_url      TEXT             NOT NULL CHECK (animation_url != ''),
    card_animation_url TEXT             NOT NULL CHECK (card_animation_url != ''),
    avatar_url         TEXT             NOT NULL CHECK (avatar_url != ''),
    hash               TEXT UNIQUE      NOT NULL,
    name               TEXT             NOT NULL,
    label              TEXT             NOT NULL,
    slug               TEXT             NOT NULL,
    asset_type         TEXT             NOT NULL CHECK (asset_type != ''),

    deleted_at         TIMESTAMPTZ,
    updated_at         TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at         TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE weapons
(
    id          UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    brand_id    UUID REFERENCES brands (id),

    label       TEXT             NOT NULL,
    slug        TEXT             NOT NULL,
    damage      INTEGER          NOT NULL,
    weapon_type TEXT             NOT NULL CHECK (weapon_type IN ('TURRET', 'ARM')),

    deleted_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);


CREATE TABLE modules
(
    id                UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    brand_id          UUID REFERENCES brands (id),

    slug              TEXT             NOT NULL,
    label             TEXT             NOT NULL,
    hitpoint_modifier INTEGER          NOT NULL,
    shield_modifier   INTEGER          NOT NULL,

    deleted_at        TIMESTAMPTZ,
    updated_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE chassis_weapons
(
    id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    chassis_id     UUID             NOT NULL REFERENCES chassis (id),
    weapon_id      UUID             NOT NULL REFERENCES weapons (id),
    slot_number    INTEGER          NOT NULL,
    mount_location TEXT             NOT NULL CHECK (mount_location IN ('ARM', 'TURRET')),

    UNIQUE (chassis_id, slot_number, mount_location),
    UNIQUE (weapon_id),

    deleted_at     TIMESTAMPTZ,
    updated_at     TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at     TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE chassis_modules
(
    id          UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    chassis_id  UUID             NOT NULL REFERENCES chassis (id),
    module_id   UUID             NOT NULL REFERENCES modules (id),
    slot_number INTEGER          NOT NULL,

    UNIQUE (chassis_id, slot_number),
    UNIQUE (module_id),

    deleted_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);


-- Players
-- 1 for each faction
INSERT INTO players (id, faction_id, username, public_address, is_ai)
VALUES ('1a657a32-778e-4612-8cc1-14e360665f2b', '880db344-e405-428d-84e5-6ebebab1fe6d', 'Zaibatsu', NULL, TRUE);
INSERT INTO players (id, faction_id, username, public_address, is_ai)
VALUES ('305da475-53dc-4973-8d78-a30d390d3de5', '98bf7bb3-1a7c-4f21-8843-458d62884060', 'RedMountain', NULL, TRUE);
INSERT INTO players (id, faction_id, username, public_address, is_ai)
VALUES ('15f29ee9-e834-4f76-aff8-31e39faabe2d', '7c6dde21-b067-46cf-9e56-155c88a520e2', 'BostonCybernetics', NULL,
        TRUE);

-- Default Chassis
-- 1 for ZHI
-- 2 for RM
-- 2 for BC

INSERT INTO blueprint_chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('17a1426a-b17e-4734-b845-588a87b6d8cd', '2b203c87-ad8c-4ce2-af17-e079835fdbcb', 'Zaibatsu WREX Black Chassis',
        'zaibatsu_wrex_black', 'WREX', 'Black', 80, 2, 2, 1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('9bc4611a-0da2-4b07-bab6-7f15b1c97f90', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain BXSD Pink Chassis', 'red_mountain_bxsd_pink', 'BXSD', 'Pink', 80, 2, 2, 1, 1750, 1500, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('a5806303-f397-44d5-b9d8-f435500362e6', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain BXSD Red_Steel Chassis', 'red_mountain_bxsd_red_steel', 'BXSD', 'Red_Steel', 80, 2, 2, 1, 1750,
        1500, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('cfb96c42-873d-43e8-9060-ae107f18037a', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics XFVS BlueWhite Chassis', 'boston_cybernetics_xfvs_blue_white', 'XFVS', 'BlueWhite', 80, 2,
        0, 1, 2750, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('ad124964-e062-4ab8-9cce-e0309fd6b31d', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics XFVS Police_DarkBlue Chassis', 'boston_cybernetics_xfvs_police_dark_blue', 'XFVS',
        'Police_DarkBlue', 80, 2, 0, 1, 2750, 1000, 1000);


-- Default weapons
-- 6 Weapons
-- INSERT INTO blueprint_weapons (id, label, slug, damage, weapon_type)
-- VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', 'Sniper Rifle', 'sniper_rifle', -1, 'ARM');
-- INSERT INTO blueprint_weapons (id, label, slug, damage, weapon_type)
-- VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', 'Laser Sword', 'laser_sword', -1, 'ARM');
-- INSERT INTO blueprint_weapons (id, label, slug, damage, weapon_type)
-- VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
-- INSERT INTO blueprint_weapons (id, label, slug, damage, weapon_type)
-- VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'Auto Cannon', 'auto_cannon', -1, 'ARM');
-- INSERT INTO blueprint_weapons (id, label, slug, damage, weapon_type)
-- VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', 'Plasma Rifle', 'plasma_rifle', -1, 'ARM');
-- INSERT INTO blueprint_weapons (id, label, slug, damage, weapon_type)
-- VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', 'Sword', 'sword', -1, 'ARM');

-- Default modules
-- Shield only
INSERT INTO blueprint_modules (id, slug, label, hitpoint_modifier, shield_modifier)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'shield', 'Shield', '100', '100');

-- Templates
-- 1 for ZHI
-- 2 for RM
-- 2 for BC
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('106f4cd9-6c4b-4967-a799-40f24d9e4ee4', '17a1426a-b17e-4734-b845-588a87b6d8cd',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_wrex_black.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_wrex_black.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_wrex_black.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_static_avatar.png',
        'Zaibatsu WREX Black', 'zaibatsu_wrex_black', TRUE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('5b9acb25-c624-4096-aa6d-c0f4f2869b7c', 'a5806303-f397-44d5-b9d8-f435500362e6',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red_mountain_bxsd_red_steel.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red_mountain_bxsd_red_steel.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red_mountain_bxsd_red_steel.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-hex_avatar.png',
        'Red Mountain BXSD Red_Steel', 'red_mountain_bxsd_red_steel', TRUE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('e0ae2071-2e60-4b7b-854e-2f88a5d80a77', '9bc4611a-0da2-4b07-bab6-7f15b1c97f90',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red_mountain_bxsd_pink.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red_mountain_bxsd_pink.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red_mountain_bxsd_pink.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-hex_avatar.png',
        'Red Mountain BXSD Pink', 'red_mountain_bxsd_pink', TRUE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('19a270bd-14d7-4dc4-84f7-ac128eaea2a8', 'cfb96c42-873d-43e8-9060-ae107f18037a',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston_cybernetics_xfvs_blue_white.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston_cybernetics_xfvs_blue_white.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston_cybernetics_xfvs_blue_white.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dune_avatar.png',
        'Boston Cybernetics XFVS BlueWhite', 'boston_cybernetics_xfvs_blue_white', TRUE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('2c5acb39-9ae3-4755-a689-a8ec1b202c16', 'ad124964-e062-4ab8-9cce-e0309fd6b31d',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston_cybernetics_xfvs_police_dark_blue.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston_cybernetics_xfvs_police_dark_blue.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston_cybernetics_xfvs_police_dark_blue.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dune_avatar.png',
        'Boston Cybernetics XFVS Police_DarkBlue', 'boston_cybernetics_xfvs_police_dark_blue', TRUE, 'War Machine');

-- Weapon Joins
-- ZHI sniper laser rocket rocket
-- RM auto auto rocket rocket
-- BC plasma, sword
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', '17a1426a-b17e-4734-b845-588a87b6d8cd', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', '17a1426a-b17e-4734-b845-588a87b6d8cd', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '17a1426a-b17e-4734-b845-588a87b6d8cd', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '17a1426a-b17e-4734-b845-588a87b6d8cd', 1, 'TURRET');

INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '9bc4611a-0da2-4b07-bab6-7f15b1c97f90', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '9bc4611a-0da2-4b07-bab6-7f15b1c97f90', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '9bc4611a-0da2-4b07-bab6-7f15b1c97f90', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '9bc4611a-0da2-4b07-bab6-7f15b1c97f90', 1, 'TURRET');

INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'a5806303-f397-44d5-b9d8-f435500362e6', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'a5806303-f397-44d5-b9d8-f435500362e6', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'a5806303-f397-44d5-b9d8-f435500362e6', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'a5806303-f397-44d5-b9d8-f435500362e6', 1, 'TURRET');

INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', 'cfb96c42-873d-43e8-9060-ae107f18037a', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', 'cfb96c42-873d-43e8-9060-ae107f18037a', 1, 'ARM');

INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', 'ad124964-e062-4ab8-9cce-e0309fd6b31d', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', 'ad124964-e062-4ab8-9cce-e0309fd6b31d', 1, 'ARM');

-- Module Joins
-- ZHI 1 shield
-- RM#1 1 shield
-- RM#2 1 shield
-- BC#1 1 shield
-- BC#2 1 shield
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '17a1426a-b17e-4734-b845-588a87b6d8cd', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '9bc4611a-0da2-4b07-bab6-7f15b1c97f90', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'a5806303-f397-44d5-b9d8-f435500362e6', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'cfb96c42-873d-43e8-9060-ae107f18037a', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'ad124964-e062-4ab8-9cce-e0309fd6b31d', 0);

-- Default Mechs
INSERT INTO chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints, turret_hardpoints,
                     utility_slots, speed, max_hitpoints, health_remaining, max_shield)
VALUES ('3abd79cb-9e7b-40e4-987f-85d9b82c7f28', '2b203c87-ad8c-4ce2-af17-e079835fdbcb', 'Zaibatsu WREX Black Chassis',
        'zaibatsu_wrex_black', 'WREX', 'Black', 80, 2, 2, 1, 2500, 1000, 1000, 1000);
INSERT INTO chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints, turret_hardpoints,
                     utility_slots, speed, max_hitpoints, health_remaining, max_shield)
VALUES ('4c0aec4f-215e-4dee-889f-ac39fb74c75d', '2b203c87-ad8c-4ce2-af17-e079835fdbcb', 'Zaibatsu WREX Black Chassis',
        'zaibatsu_wrex_black', 'WREX', 'Black', 80, 2, 2, 1, 2500, 1000, 1000, 1000);
INSERT INTO chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints, turret_hardpoints,
                     utility_slots, speed, max_hitpoints, health_remaining, max_shield)
VALUES ('b5892cc1-5fc9-4106-8b00-b9f20e7febed', '2b203c87-ad8c-4ce2-af17-e079835fdbcb', 'Zaibatsu WREX Black Chassis',
        'zaibatsu_wrex_black', 'WREX', 'Black', 80, 2, 2, 1, 2500, 1000, 1000, 1000);
INSERT INTO chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints, turret_hardpoints,
                     utility_slots, speed, max_hitpoints, health_remaining, max_shield)
VALUES ('0511e9e4-7b1f-426d-9d2a-9667be38ff9b', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain BXSD Red_Steel Chassis', 'red_mountain_bxsd_red_steel', 'BXSD', 'Red_Steel', 80, 2, 2, 1, 1750,
        1500, 1500, 1000);
INSERT INTO chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints, turret_hardpoints,
                     utility_slots, speed, max_hitpoints, health_remaining, max_shield)
VALUES ('4b7f8a7b-68c6-45f9-8054-c082c5a5e8c9', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain BXSD Pink Chassis', 'red_mountain_bxsd_pink', 'BXSD', 'Pink', 80, 2, 2, 1, 1750, 1500, 1500,
        1000);
INSERT INTO chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints, turret_hardpoints,
                     utility_slots, speed, max_hitpoints, health_remaining, max_shield)
VALUES ('c638ee34-81ef-410a-8f5d-e10e4a973966', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain BXSD Pink Chassis', 'red_mountain_bxsd_pink', 'BXSD', 'Pink', 80, 2, 2, 1, 1750, 1500, 1500,
        1000);
INSERT INTO chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints, turret_hardpoints,
                     utility_slots, speed, max_hitpoints, health_remaining, max_shield)
VALUES ('088132f5-ee2f-47f7-bd18-9b7dfb466ff5', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics XFVS BlueWhite Chassis', 'boston_cybernetics_xfvs_blue_white', 'XFVS', 'BlueWhite', 80, 2,
        0, 1, 2750, 1000, 1000, 1000);
INSERT INTO chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints, turret_hardpoints,
                     utility_slots, speed, max_hitpoints, health_remaining, max_shield)
VALUES ('67c04307-4f9f-44bb-9bfc-d22f46905e2e', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics XFVS Police_DarkBlue Chassis', 'boston_cybernetics_xfvs_police_dark_blue', 'XFVS',
        'Police_DarkBlue', 80, 2, 0, 1, 2750, 1000, 1000, 1000);
INSERT INTO chassis (id, brand_id, label, slug, model, skin, shield_recharge_rate, weapon_hardpoints, turret_hardpoints,
                     utility_slots, speed, max_hitpoints, health_remaining, max_shield)
VALUES ('4b14cba5-eaaa-4f7b-aa24-34b4072c3589', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics XFVS Police_DarkBlue Chassis', 'boston_cybernetics_xfvs_police_dark_blue', 'XFVS',
        'Police_DarkBlue', 80, 2, 0, 1, 2750, 1000, 1000, 1000);

INSERT INTO mechs (id, owner_id, template_id, chassis_id, external_token_id, tier, image_url, animation_url,
                   card_animation_url, avatar_url, is_default, hash, name, label, slug, asset_type)
VALUES ('fc43fa34-b23f-40f4-afaa-465f4880ef59', '1a657a32-778e-4612-8cc1-14e360665f2b',
        '106f4cd9-6c4b-4967-a799-40f24d9e4ee4', '3abd79cb-9e7b-40e4-987f-85d9b82c7f28', 0, 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_wrex_black.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_wrex_black.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_wrex_black.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_static_avatar.png', TRUE,
        'ZXga92AmGD', 'Alex', 'Zaibatsu WREX Black Alex', 'zaibatsu_wrex_black_alex', 'War Machine');
INSERT INTO mechs (id, owner_id, template_id, chassis_id, external_token_id, tier, image_url, animation_url,
                   card_animation_url, avatar_url, is_default, hash, name, label, slug, asset_type)
VALUES ('7689dff0-cf6d-45dc-8a46-821ff3c9dc68', '1a657a32-778e-4612-8cc1-14e360665f2b',
        '106f4cd9-6c4b-4967-a799-40f24d9e4ee4', '4c0aec4f-215e-4dee-889f-ac39fb74c75d', 1, 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_wrex_black.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_wrex_black.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_wrex_black.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_static_avatar.png', TRUE,
        'dbYaD4a0Zj', 'John', 'Zaibatsu WREX Black John', 'zaibatsu_wrex_black_john', 'War Machine');
INSERT INTO mechs (id, owner_id, template_id, chassis_id, external_token_id, tier, image_url, animation_url,
                   card_animation_url, avatar_url, is_default, hash, name, label, slug, asset_type)
VALUES ('d7ac6e99-5587-4a4a-a66d-486549ac0ffc', '1a657a32-778e-4612-8cc1-14e360665f2b',
        '106f4cd9-6c4b-4967-a799-40f24d9e4ee4', 'b5892cc1-5fc9-4106-8b00-b9f20e7febed', 2, 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_wrex_black.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_wrex_black.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_wrex_black.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_static_avatar.png', TRUE,
        'l7epj2pPL4', 'Mac', 'Zaibatsu WREX Black Mac', 'zaibatsu_wrex_black_mac', 'War Machine');
INSERT INTO mechs (id, owner_id, template_id, chassis_id, external_token_id, tier, image_url, animation_url,
                   card_animation_url, avatar_url, is_default, hash, name, label, slug, asset_type)
VALUES ('7e752603-21a0-45c1-9502-e8213ee58f8c', '305da475-53dc-4973-8d78-a30d390d3de5',
        '5b9acb25-c624-4096-aa6d-c0f4f2869b7c', '0511e9e4-7b1f-426d-9d2a-9667be38ff9b', 3, 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red_mountain_bxsd_red_steel.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red_mountain_bxsd_red_steel.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red_mountain_bxsd_red_steel.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-hex_avatar.png', TRUE,
        'kN7aVgAenK', 'Vinnie', 'Red Mountain BXSD Red_Steel Vinnie', 'red_mountain_bxsd_red_steel_vinnie',
        'War Machine');
INSERT INTO mechs (id, owner_id, template_id, chassis_id, external_token_id, tier, image_url, animation_url,
                   card_animation_url, avatar_url, is_default, hash, name, label, slug, asset_type)
VALUES ('0c3fc4dc-b4c7-4899-bab3-87abf1e1e1fb', '305da475-53dc-4973-8d78-a30d390d3de5',
        'e0ae2071-2e60-4b7b-854e-2f88a5d80a77', '4b7f8a7b-68c6-45f9-8054-c082c5a5e8c9', 4, 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red_mountain_bxsd_pink.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red_mountain_bxsd_pink.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red_mountain_bxsd_pink.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-hex_avatar.png', TRUE,
        'wdBAN1aeo5', 'Owen', 'Red Mountain BXSD Pink Owen', 'red_mountain_bxsd_pink_owen', 'War Machine');
INSERT INTO mechs (id, owner_id, template_id, chassis_id, external_token_id, tier, image_url, animation_url,
                   card_animation_url, avatar_url, is_default, hash, name, label, slug, asset_type)
VALUES ('e9d202e3-98f4-4197-b3ec-12c97e24923c', '305da475-53dc-4973-8d78-a30d390d3de5',
        'e0ae2071-2e60-4b7b-854e-2f88a5d80a77', 'c638ee34-81ef-410a-8f5d-e10e4a973966', 5, 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red_mountain_bxsd_pink.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red_mountain_bxsd_pink.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red_mountain_bxsd_pink.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-hex_avatar.png', TRUE,
        '018pkXaRWM', 'James', 'Red Mountain BXSD Pink James', 'red_mountain_bxsd_pink_james', 'War Machine');
INSERT INTO mechs (id, owner_id, template_id, chassis_id, external_token_id, tier, image_url, animation_url,
                   card_animation_url, avatar_url, is_default, hash, name, label, slug, asset_type)
VALUES ('51ff0d47-f35a-4d49-8987-d42857d1e345', '15f29ee9-e834-4f76-aff8-31e39faabe2d',
        '19a270bd-14d7-4dc4-84f7-ac128eaea2a8', '088132f5-ee2f-47f7-bd18-9b7dfb466ff5', 6, 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston_cybernetics_xfvs_blue_white.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston_cybernetics_xfvs_blue_white.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston_cybernetics_xfvs_blue_white.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dune_avatar.png',
        TRUE, 'B8x3qdAy6K', 'Darren', 'Boston Cybernetics XFVS BlueWhite Darren',
        'boston_cybernetics_xfvs_blue_white_darren', 'War Machine');
INSERT INTO mechs (id, owner_id, template_id, chassis_id, external_token_id, tier, image_url, animation_url,
                   card_animation_url, avatar_url, is_default, hash, name, label, slug, asset_type)
VALUES ('a797a8c1-157c-4540-a0e6-4b944e07f383', '15f29ee9-e834-4f76-aff8-31e39faabe2d',
        '2c5acb39-9ae3-4755-a689-a8ec1b202c16', '67c04307-4f9f-44bb-9bfc-d22f46905e2e', 7, 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston_cybernetics_xfvs_police_dark_blue.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston_cybernetics_xfvs_police_dark_blue.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston_cybernetics_xfvs_police_dark_blue.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dune_avatar.png',
        TRUE, 'D16aRep0Zo', 'Veracity1', 'Boston Cybernetics XFVS Police_DarkBlue Veracity1',
        'boston_cybernetics_xfvs_police_dark_blue_veracity1', 'War Machine');
INSERT INTO mechs (id, owner_id, template_id, chassis_id, external_token_id, tier, image_url, animation_url,
                   card_animation_url, avatar_url, is_default, hash, name, label, slug, asset_type)
VALUES ('58765b3c-fca0-4759-9357-e6b35823c15f', '15f29ee9-e834-4f76-aff8-31e39faabe2d',
        '2c5acb39-9ae3-4755-a689-a8ec1b202c16', '4b14cba5-eaaa-4f7b-aa24-34b4072c3589', 8, 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston_cybernetics_xfvs_police_dark_blue.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston_cybernetics_xfvs_police_dark_blue.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston_cybernetics_xfvs_police_dark_blue.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dune_avatar.png',
        TRUE, '4Q1p8dpqwX', 'Corey', 'Boston Cybernetics XFVS Police_DarkBlue Corey',
        'boston_cybernetics_xfvs_police_dark_blue_corey', 'War Machine');

-- Default Mech Weapons
-- ZHI sniper laser rocket rocket
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('6a0f0305-7a17-4eb9-a55b-872922904699', 'Sniper Rifle', 'sniper_rifle', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('9f07f188-e189-40fd-a73a-2a6e0a2a160b', 'Laser Sword', 'laser_sword', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('16ae3d3c-4388-439e-97a6-ce1b0877471f', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('9d217c41-c11f-41dc-9392-d7ea9625df14', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('330a8eea-9ffb-40df-9425-18822cd0e915', 'Sniper Rifle', 'sniper_rifle', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('72db6eb9-09a2-41ad-a227-5447f3c62b43', 'Laser Sword', 'laser_sword', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('654db741-b93f-4922-a10d-c2bb66a9afbb', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('72497318-74d3-4ce0-a6e9-10a580f900a9', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('a460a92c-94bd-41da-a210-8403ad9288b2', 'Sniper Rifle', 'sniper_rifle', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('f2fd2be1-1ed1-4f85-8e6a-e11757e940c6', 'Laser Sword', 'laser_sword', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('a5bbf058-6741-468d-8285-3e2b260bac3b', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('f01ad402-1f0a-496a-83c0-ea41beb441d7', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');

INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('c494d83d-33ea-4858-af1c-455afbaf36ce', '6a0f0305-7a17-4eb9-a55b-872922904699',
        '3abd79cb-9e7b-40e4-987f-85d9b82c7f28', 0, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('cce721ff-c756-4488-af7a-4783e2a36f17', '9f07f188-e189-40fd-a73a-2a6e0a2a160b',
        '3abd79cb-9e7b-40e4-987f-85d9b82c7f28', 1, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('fdd501e3-59d8-42f5-a002-a3e809edeb2d', '16ae3d3c-4388-439e-97a6-ce1b0877471f',
        '3abd79cb-9e7b-40e4-987f-85d9b82c7f28', 0, 'TURRET');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('a73f7984-b121-4c7f-8596-b9efa3b7801b', '9d217c41-c11f-41dc-9392-d7ea9625df14',
        '3abd79cb-9e7b-40e4-987f-85d9b82c7f28', 1, 'TURRET');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('c2829bf7-c6e0-4d99-b28e-88d438f1d8f1', '330a8eea-9ffb-40df-9425-18822cd0e915',
        '4c0aec4f-215e-4dee-889f-ac39fb74c75d', 0, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('29fbda66-fe2b-4e72-ae44-28201d564055', '72db6eb9-09a2-41ad-a227-5447f3c62b43',
        '4c0aec4f-215e-4dee-889f-ac39fb74c75d', 1, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('18b22884-1ad5-47d1-a073-e924daa4f671', '654db741-b93f-4922-a10d-c2bb66a9afbb',
        '4c0aec4f-215e-4dee-889f-ac39fb74c75d', 0, 'TURRET');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('cbb37ddc-0fd0-4e72-be5c-99e2165b9a69', '72497318-74d3-4ce0-a6e9-10a580f900a9',
        '4c0aec4f-215e-4dee-889f-ac39fb74c75d', 1, 'TURRET');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('e1208d06-af51-46bc-b080-7a775f15253b', 'a460a92c-94bd-41da-a210-8403ad9288b2',
        'b5892cc1-5fc9-4106-8b00-b9f20e7febed', 0, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('d60be9a4-3d38-4c11-803b-8a54abda3e9e', 'f2fd2be1-1ed1-4f85-8e6a-e11757e940c6',
        'b5892cc1-5fc9-4106-8b00-b9f20e7febed', 1, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('e92af183-ea8b-488e-a509-bbf285e7e592', 'a5bbf058-6741-468d-8285-3e2b260bac3b',
        'b5892cc1-5fc9-4106-8b00-b9f20e7febed', 0, 'TURRET');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('1b51b659-ed29-454f-ae84-2fc5abfe2c7b', 'f01ad402-1f0a-496a-83c0-ea41beb441d7',
        'b5892cc1-5fc9-4106-8b00-b9f20e7febed', 1, 'TURRET');

-- Default Mech Weapons
-- RM auto auto rocket rocket

INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('b02c95db-3b5c-4e63-a8e7-3c87ec1b9869', 'Auto Cannon', 'auto_cannon', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('2f577a62-26ee-4de3-8736-ff963210c71e', 'Auto Cannon', 'auto_cannon', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('c7d522b1-ccfa-4a73-9c78-60ed896ceb96', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('e01b088f-a363-4fa0-828e-7d36b3238a31', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('464eb93f-10f7-4407-be4b-5c80566a5ce6', 'Auto Cannon', 'auto_cannon', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('b064f2e6-7e89-4e48-9a37-638fb99c2b18', 'Auto Cannon', 'auto_cannon', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('e39e3bf0-09ae-4cf1-97b9-0276434dc7f0', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('a52ed921-5804-426d-8af2-c7fb48f40457', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('ef7b6cc8-024d-4683-a8b6-743d836c8e54', 'Auto Cannon', 'auto_cannon', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('f0264a96-adf6-461c-9cbf-ff18c062f2b0', 'Auto Cannon', 'auto_cannon', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('530c43ce-6320-4863-ba24-94a9e974a560', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('7bb3adc8-f23d-47dd-a2ee-1dc8808a5a4b', 'Rocket Pod', 'rocket_pod', -1, 'TURRET');

INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('eaf5c851-d776-4d74-b5df-d09a78058522', 'b02c95db-3b5c-4e63-a8e7-3c87ec1b9869',
        '0511e9e4-7b1f-426d-9d2a-9667be38ff9b', 0, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('d664c0db-2d88-4018-85d6-c9d5136817d6', '2f577a62-26ee-4de3-8736-ff963210c71e',
        '0511e9e4-7b1f-426d-9d2a-9667be38ff9b', 1, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('cb7f6404-a194-4e70-8b6a-1403a9565c69', 'c7d522b1-ccfa-4a73-9c78-60ed896ceb96',
        '0511e9e4-7b1f-426d-9d2a-9667be38ff9b', 0, 'TURRET');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('37cb330a-1eba-4518-af26-72e49cd2f1a4', 'e01b088f-a363-4fa0-828e-7d36b3238a31',
        '0511e9e4-7b1f-426d-9d2a-9667be38ff9b', 1, 'TURRET');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('c85242fe-ff97-4f53-8ea1-8c8f84cbfccc', '464eb93f-10f7-4407-be4b-5c80566a5ce6',
        '4b7f8a7b-68c6-45f9-8054-c082c5a5e8c9', 0, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('9dcf79c2-96d1-49b3-a02d-3f0f175b79a3', 'b064f2e6-7e89-4e48-9a37-638fb99c2b18',
        '4b7f8a7b-68c6-45f9-8054-c082c5a5e8c9', 1, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('31ee56fa-f1f7-4d62-8f95-4e42b5e348c0', 'e39e3bf0-09ae-4cf1-97b9-0276434dc7f0',
        '4b7f8a7b-68c6-45f9-8054-c082c5a5e8c9', 0, 'TURRET');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('b5da5710-0374-461e-9237-83f2c9910daa', 'a52ed921-5804-426d-8af2-c7fb48f40457',
        '4b7f8a7b-68c6-45f9-8054-c082c5a5e8c9', 1, 'TURRET');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('41b98452-29dd-4c4c-a89c-49963a8e4591', 'ef7b6cc8-024d-4683-a8b6-743d836c8e54',
        'c638ee34-81ef-410a-8f5d-e10e4a973966', 0, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('a25637c9-9047-4d56-a02c-1849dfcdd0fc', 'f0264a96-adf6-461c-9cbf-ff18c062f2b0',
        'c638ee34-81ef-410a-8f5d-e10e4a973966', 1, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('a9408cdd-f7ad-41bb-bf5f-52a57d67210e', '530c43ce-6320-4863-ba24-94a9e974a560',
        'c638ee34-81ef-410a-8f5d-e10e4a973966', 0, 'TURRET');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('cd90ab96-c046-4e74-a369-b083a108d0ff', '7bb3adc8-f23d-47dd-a2ee-1dc8808a5a4b',
        'c638ee34-81ef-410a-8f5d-e10e4a973966', 1, 'TURRET');

-- Default Mech Weapons
-- BC plasma, sword

INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('2f2f634b-d7a3-4bca-9877-40dacdf0d672', 'Plasma Rifle', 'plasma_rifle', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('aa6c3c23-fa0f-4457-9460-20e5b6e19300', 'Sword', 'sword', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('60260cc1-177b-43af-a2f0-3f642231ac80', 'Plasma Rifle', 'plasma_rifle', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('2cdd7b1a-b6ae-4c4d-b072-123b64d3b99a', 'Sword', 'sword', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('2425fdea-a86b-4c86-b605-7853e2457fb3', 'Plasma Rifle', 'plasma_rifle', -1, 'ARM');
INSERT INTO weapons (id, label, slug, damage, weapon_type)
VALUES ('fe2d6bca-a7ce-4104-88ff-e1b7abda8ba5', 'Sword', 'sword', -1, 'ARM');

INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('53709691-7ba7-40fb-859f-74d1d0ed837c', '2f2f634b-d7a3-4bca-9877-40dacdf0d672',
        '088132f5-ee2f-47f7-bd18-9b7dfb466ff5', 0, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('156bceda-b9bd-4b09-b4df-beed4c5098db', 'aa6c3c23-fa0f-4457-9460-20e5b6e19300',
        '088132f5-ee2f-47f7-bd18-9b7dfb466ff5', 1, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('10e3a960-02ae-408c-bb17-d75b9a7c0267', '60260cc1-177b-43af-a2f0-3f642231ac80',
        '67c04307-4f9f-44bb-9bfc-d22f46905e2e', 0, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('3a4d0074-d845-4c97-982e-f2248760b309', '2cdd7b1a-b6ae-4c4d-b072-123b64d3b99a',
        '67c04307-4f9f-44bb-9bfc-d22f46905e2e', 1, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('62fce0cd-d5a3-401c-9032-288ada3a60f3', '2425fdea-a86b-4c86-b605-7853e2457fb3',
        '4b14cba5-eaaa-4f7b-aa24-34b4072c3589', 0, 'ARM');
INSERT INTO chassis_weapons (id, weapon_id, chassis_id, slot_number, mount_location)
VALUES ('df57ff2f-aaac-4085-a3c4-664f0361bb43', 'fe2d6bca-a7ce-4104-88ff-e1b7abda8ba5',
        '4b14cba5-eaaa-4f7b-aa24-34b4072c3589', 1, 'ARM');

-- Default Mech Modules
INSERT INTO modules (id, slug, label, hitpoint_modifier, shield_modifier)
VALUES ('6d5b1d4e-f583-4e50-aff6-c70b9498b49a', 'shield', 'Shield', '100', '100');
INSERT INTO modules (id, slug, label, hitpoint_modifier, shield_modifier)
VALUES ('54b1ed24-0260-4f6a-aba8-7e7b82bc1b1a', 'shield', 'Shield', '100', '100');
INSERT INTO modules (id, slug, label, hitpoint_modifier, shield_modifier)
VALUES ('fc684dbb-f0f6-42be-a280-72e273aace8e', 'shield', 'Shield', '100', '100');
INSERT INTO modules (id, slug, label, hitpoint_modifier, shield_modifier)
VALUES ('37364bb1-7f8d-4300-b687-dd9eab4ace52', 'shield', 'Shield', '100', '100');
INSERT INTO modules (id, slug, label, hitpoint_modifier, shield_modifier)
VALUES ('876b8a7c-103c-4bbd-ac35-0a7cfaa69f64', 'shield', 'Shield', '100', '100');
INSERT INTO modules (id, slug, label, hitpoint_modifier, shield_modifier)
VALUES ('528dc130-ab50-417a-8951-95e158bacc07', 'shield', 'Shield', '100', '100');
INSERT INTO modules (id, slug, label, hitpoint_modifier, shield_modifier)
VALUES ('6397a691-dce2-4cae-968b-71574700e189', 'shield', 'Shield', '100', '100');
INSERT INTO modules (id, slug, label, hitpoint_modifier, shield_modifier)
VALUES ('00a206ee-9702-4af7-9b39-37934fc2345c', 'shield', 'Shield', '100', '100');
INSERT INTO modules (id, slug, label, hitpoint_modifier, shield_modifier)
VALUES ('6734f03b-1be2-49c9-b5bf-5191ba4bfb0b', 'shield', 'Shield', '100', '100');

INSERT INTO chassis_modules (id, module_id, chassis_id, slot_number)
VALUES ('5426a7f7-c19b-4757-9d84-460e6ea4e7ee', '6d5b1d4e-f583-4e50-aff6-c70b9498b49a',
        '3abd79cb-9e7b-40e4-987f-85d9b82c7f28', 0);
INSERT INTO chassis_modules (id, module_id, chassis_id, slot_number)
VALUES ('7c7cd32d-88b1-4f33-998f-5af374459bfd', '54b1ed24-0260-4f6a-aba8-7e7b82bc1b1a',
        '4c0aec4f-215e-4dee-889f-ac39fb74c75d', 0);
INSERT INTO chassis_modules (id, module_id, chassis_id, slot_number)
VALUES ('342f2b54-c9c2-4733-81f5-c157084bb819', 'fc684dbb-f0f6-42be-a280-72e273aace8e',
        'b5892cc1-5fc9-4106-8b00-b9f20e7febed', 0);
INSERT INTO chassis_modules (id, module_id, chassis_id, slot_number)
VALUES ('c52143f2-0d34-409d-bb35-2e6bfe92cea4', '37364bb1-7f8d-4300-b687-dd9eab4ace52',
        '0511e9e4-7b1f-426d-9d2a-9667be38ff9b', 0);
INSERT INTO chassis_modules (id, module_id, chassis_id, slot_number)
VALUES ('00f09bcb-d3f2-44cf-86ad-6c821b061d35', '876b8a7c-103c-4bbd-ac35-0a7cfaa69f64',
        '4b7f8a7b-68c6-45f9-8054-c082c5a5e8c9', 0);
INSERT INTO chassis_modules (id, module_id, chassis_id, slot_number)
VALUES ('ca2d4beb-cfaf-4c02-bb27-8295b1b330aa', '528dc130-ab50-417a-8951-95e158bacc07',
        'c638ee34-81ef-410a-8f5d-e10e4a973966', 0);
INSERT INTO chassis_modules (id, module_id, chassis_id, slot_number)
VALUES ('50f72d57-978d-4312-a618-c214370e18ef', '6397a691-dce2-4cae-968b-71574700e189',
        '088132f5-ee2f-47f7-bd18-9b7dfb466ff5', 0);
INSERT INTO chassis_modules (id, module_id, chassis_id, slot_number)
VALUES ('a496756e-9dcc-4098-9366-bff53aa480b6', '00a206ee-9702-4af7-9b39-37934fc2345c',
        '67c04307-4f9f-44bb-9bfc-d22f46905e2e', 0);
INSERT INTO chassis_modules (id, module_id, chassis_id, slot_number)
VALUES ('36fc1965-9b4e-4e15-bbd3-720e862a818f', '6734f03b-1be2-49c9-b5bf-5191ba4bfb0b',
        '4b14cba5-eaaa-4f7b-aa24-34b4072c3589', 0);

-- INSERT INTO factions (id, label) VALUES ('880db344-e405-428d-84e5-6ebebab1fe6d', 'Zaibatsu Heavy Industries');
-- INSERT INTO factions (id, label) VALUES ('98bf7bb3-1a7c-4f21-8843-458d62884060', 'Red Mountain Offworld Mining Corporation');
-- INSERT INTO factions (id, label) VALUES ('7c6dde21-b067-46cf-9e56-155c88a520e2', 'Boston Cybernetics');

-- XSYN Existing BC Items


INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('8dd45a32-c201-41d4-a134-4b5a50419a6e', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 BioHazard Chassis',
        'boston_cybernetics_law_enforcer_x_1000_bio_hazard_chassis', 'BioHazard', 'Law Enforcer X-1000', 80, 2, 1, 1,
        3080, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('2b0df453-a9ac-4a43-967e-ed7a1cdaecf1', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 White Blue Chassis',
        'boston_cybernetics_law_enforcer_x_1000_white_blue_chassis', 'White Blue', 'Law Enforcer X-1000', 80, 2, 1, 1,
        2970, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('cc0de847-24a3-4764-bbbd-78304e79150e', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Crystal Blue Chassis',
        'boston_cybernetics_law_enforcer_x_1000_crystal_blue_chassis', 'Crystal Blue', 'Law Enforcer X-1000', 80, 2, 1,
        1, 3135, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('74e3bd32-4b27-4038-8338-f4f60db4be71', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Cyber Chassis', 'boston_cybernetics_law_enforcer_x_1000_cyber_chassis',
        'Cyber', 'Law Enforcer X-1000', 80, 2, 1, 1, 2805, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('8b98d84c-48ad-481b-bc86-d1d02822f16c', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Dune Chassis', 'boston_cybernetics_law_enforcer_x_1000_dune_chassis',
        'Dune', 'Law Enforcer X-1000', 80, 2, 1, 1, 2750, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('12ab349f-0831-4f3f-9aa1-f3aade13abb3', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Dynamic Yellow Chassis',
        'boston_cybernetics_law_enforcer_x_1000_dynamic_yellow_chassis', 'Dynamic Yellow', 'Law Enforcer X-1000', 80, 2,
        1, 1, 2860, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('014371f7-bca3-4a1c-992e-f55ac5721de1', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Gold Chassis', 'boston_cybernetics_law_enforcer_x_1000_gold_chassis',
        'Gold', 'Law Enforcer X-1000', 80, 2, 1, 1, 2915, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('d041e6dc-0ed5-4294-8fdc-f49707e1a854', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Light Blue Police Chassis',
        'boston_cybernetics_law_enforcer_x_1000_light_blue_police_chassis', 'Light Blue Police', 'Law Enforcer X-1000',
        80, 2, 1, 1, 3025, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('76a60a59-291a-49e2-b7d4-d7cbfa2a3feb', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Molten Chassis',
        'boston_cybernetics_law_enforcer_x_1000_molten_chassis', 'Molten', 'Law Enforcer X-1000', 80, 2, 1, 1, 3190,
        1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('0dd254a6-edd6-481b-b3d7-36c57a9836de', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Mystermech Chassis',
        'boston_cybernetics_law_enforcer_x_1000_mystermech_chassis', 'Mystermech', 'Law Enforcer X-1000', 80, 2, 1, 1,
        2860, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('11d09723-a9e2-4c57-bb01-5291be841cb7', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Nebula Chassis',
        'boston_cybernetics_law_enforcer_x_1000_nebula_chassis', 'Nebula', 'Law Enforcer X-1000', 80, 2, 1, 1, 3300,
        1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('3e710f5f-1290-428d-a6c1-1acc08f98837', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Rust Bucket Chassis',
        'boston_cybernetics_law_enforcer_x_1000_rust_bucket_chassis', 'Rust Bucket', 'Law Enforcer X-1000', 80, 2, 1, 1,
        2750, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('d5c07ee1-8213-43bf-bd65-bbacb442d4ff', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Sleek Chassis', 'boston_cybernetics_law_enforcer_x_1000_sleek_chassis',
        'Sleek', 'Law Enforcer X-1000', 80, 2, 1, 1, 2805, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('3e469099-7fba-4a17-a392-8fd850d35c28', '009f71fc-3594-4d24-a6e2-f05070d66f40',
        'Boston Cybernetics Law Enforcer X-1000 Vintage Chassis',
        'boston_cybernetics_law_enforcer_x_1000_vintage_chassis', 'Vintage', 'Law Enforcer X-1000', 80, 2, 1, 1, 2750,
        1000, 1000);

INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('28b5ab75-7c54-47e0-8365-12e417ff6c10', '8b98d84c-48ad-481b-bc86-d1d02822f16c',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_dune.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_dune.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_dune.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dune_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Dune', 'boston_cybernetics_law_enforcer_x_1000_dune', FALSE,
        'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('1e505563-a0c3-4905-b9c2-c3f3fef04888', '76a60a59-291a-49e2-b7d4-d7cbfa2a3feb',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'MYTHIC',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_molten.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_molten.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_molten.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_molten_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Molten', 'boston_cybernetics_law_enforcer_x_1000_molten', FALSE,
        'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('2bf07868-8213-473a-8697-66c5afc1ab66', '74e3bd32-4b27-4038-8338-f4f60db4be71',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'COLOSSAL',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_cyber.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_cyber.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_cyber.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_cyber_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Cyber', 'boston_cybernetics_law_enforcer_x_1000_cyber', FALSE,
        'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('44a956ec-cd55-401f-8499-0e93cb259056', '8dd45a32-c201-41d4-a134-4b5a50419a6e',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'EXOTIC',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_biohazard.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_biohazard.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_biohazard.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_biohazard_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 BioHazard', 'boston_cybernetics_law_enforcer_x_1000_bio_hazard', FALSE,
        'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('567aeedc-dc49-4870-ac58-5805a693dc45', 'cc0de847-24a3-4764-bbbd-78304e79150e',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'GUARDIAN',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_crystal-blue.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_crystal-blue.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_crystal-blue.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_crystal-blue_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Crystal Blue', 'boston_cybernetics_law_enforcer_x_1000_crystal_blue',
        FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('583e8471-bfe5-401e-a623-f57e50e1ea9d', '014371f7-bca3-4a1c-992e-f55ac5721de1',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'LEGENDARY',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_gold.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_gold.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_gold.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_gold_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Gold', 'boston_cybernetics_law_enforcer_x_1000_gold', FALSE,
        'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('607fb969-ecab-4451-8e38-d54780d2bf6b', '3e469099-7fba-4a17-a392-8fd850d35c28',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_vintage.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_vintage.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_vintage.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_vintage_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Vintage', 'boston_cybernetics_law_enforcer_x_1000_vintage', FALSE,
        'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('60acd281-8bbb-4b30-a9b9-7fe315c3009a', '12ab349f-0831-4f3f-9aa1-f3aade13abb3',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'RARE',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_dynamic-yellow.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_dynamic-yellow.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_dynamic-yellow.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_dynamic-yellow_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Dynamic Yellow',
        'boston_cybernetics_law_enforcer_x_1000_dynamic_yellow', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('629d792a-2993-439d-8560-cda4f7857f84', '2b0df453-a9ac-4a43-967e-ed7a1cdaecf1',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'ELITE_LEGENDARY',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_blue-white.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_blue-white.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_blue-white.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_blue-white_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 White Blue', 'boston_cybernetics_law_enforcer_x_1000_white_blue', FALSE,
        'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('9da32703-e5fb-4510-91fa-9a4d3941369e', '0dd254a6-edd6-481b-b3d7-36c57a9836de',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'RARE',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_mystermech.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_mystermech.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_mystermech.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_mystermech_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Mystermech', 'boston_cybernetics_law_enforcer_x_1000_mystermech', FALSE,
        'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('cb39a103-4116-4fec-8ad7-499bd8271e74', '11d09723-a9e2-4c57-bb01-5291be841cb7',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'DEUS_EX',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_nebula.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_nebula.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_nebula.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_nebula_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Nebula', 'boston_cybernetics_law_enforcer_x_1000_nebula', FALSE,
        'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('cdde2497-272d-481b-8238-b13ef96be3bf', '3e710f5f-1290-428d-a6c1-1acc08f98837',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_rust-bucket.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_rust-bucket.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_rust-bucket.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_rust-bucket_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Rust Bucket', 'boston_cybernetics_law_enforcer_x_1000_rust_bucket',
        FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('d7ff95e3-6a5a-49b2-a0b3-8bdcea58e6b8', 'd041e6dc-0ed5-4294-8fdc-f49707e1a854',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'ULTRA_RARE',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_light-blue-police.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_light-blue-police.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_light-blue-police.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_light-blue-police_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Light Blue Police',
        'boston_cybernetics_law_enforcer_x_1000_light_blue_police', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('f52bceae-afaf-42f0-bc4d-a31f2e39ce61', 'd5c07ee1-8213-43bf-bd65-bbacb442d4ff',
        '7c6dde21-b067-46cf-9e56-155c88a520e2', 'COLOSSAL',
        'https://afiles.ninja-cdn.com/passport/genesis/img/boston-cybernetics_law-enforcer-x-1000_sleek.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_sleek.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/boston-cybernetics_law-enforcer-x-1000_sleek.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/boston-cybernetics_law-enforcer-x-1000_sleek_avatar.png',
        'Boston Cybernetics Law Enforcer X-1000 Sleek', 'boston_cybernetics_law_enforcer_x_1000_sleek', FALSE,
        'War Machine');

INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '76a60a59-291a-49e2-b7d4-d7cbfa2a3feb', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '8b98d84c-48ad-481b-bc86-d1d02822f16c', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '74e3bd32-4b27-4038-8338-f4f60db4be71', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '8dd45a32-c201-41d4-a134-4b5a50419a6e', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'cc0de847-24a3-4764-bbbd-78304e79150e', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '014371f7-bca3-4a1c-992e-f55ac5721de1', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '3e469099-7fba-4a17-a392-8fd850d35c28', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '12ab349f-0831-4f3f-9aa1-f3aade13abb3', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '2b0df453-a9ac-4a43-967e-ed7a1cdaecf1', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '0dd254a6-edd6-481b-b3d7-36c57a9836de', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '11d09723-a9e2-4c57-bb01-5291be841cb7', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '3e710f5f-1290-428d-a6c1-1acc08f98837', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'd041e6dc-0ed5-4294-8fdc-f49707e1a854', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'd5c07ee1-8213-43bf-bd65-bbacb442d4ff', 0);

INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '76a60a59-291a-49e2-b7d4-d7cbfa2a3feb', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '76a60a59-291a-49e2-b7d4-d7cbfa2a3feb', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '8b98d84c-48ad-481b-bc86-d1d02822f16c', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '8b98d84c-48ad-481b-bc86-d1d02822f16c', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '74e3bd32-4b27-4038-8338-f4f60db4be71', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '74e3bd32-4b27-4038-8338-f4f60db4be71', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '8dd45a32-c201-41d4-a134-4b5a50419a6e', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '8dd45a32-c201-41d4-a134-4b5a50419a6e', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', 'cc0de847-24a3-4764-bbbd-78304e79150e', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', 'cc0de847-24a3-4764-bbbd-78304e79150e', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '014371f7-bca3-4a1c-992e-f55ac5721de1', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '014371f7-bca3-4a1c-992e-f55ac5721de1', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '3e469099-7fba-4a17-a392-8fd850d35c28', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '3e469099-7fba-4a17-a392-8fd850d35c28', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '12ab349f-0831-4f3f-9aa1-f3aade13abb3', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '12ab349f-0831-4f3f-9aa1-f3aade13abb3', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '2b0df453-a9ac-4a43-967e-ed7a1cdaecf1', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '2b0df453-a9ac-4a43-967e-ed7a1cdaecf1', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '0dd254a6-edd6-481b-b3d7-36c57a9836de', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '0dd254a6-edd6-481b-b3d7-36c57a9836de', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '11d09723-a9e2-4c57-bb01-5291be841cb7', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '11d09723-a9e2-4c57-bb01-5291be841cb7', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', '3e710f5f-1290-428d-a6c1-1acc08f98837', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', '3e710f5f-1290-428d-a6c1-1acc08f98837', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', 'd041e6dc-0ed5-4294-8fdc-f49707e1a854', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', 'd041e6dc-0ed5-4294-8fdc-f49707e1a854', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('ba29ce67-4738-4a66-81dc-932a2ccf6cd7', 'd5c07ee1-8213-43bf-bd65-bbacb442d4ff', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('26cccb14-5e61-4b3b-a522-b3b82b1ee511', 'd5c07ee1-8213-43bf-bd65-bbacb442d4ff', 1, 'ARM');

-- XSYN Existing RM Items


INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('a5966250-4972-425f-b2d4-433e3820741a', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Red Hex Chassis', 'red_mountain_olympus_mons_ly07_red_hex_chassis', 'Red Hex',
        'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1560, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('24a4ebc1-e785-4b72-83e8-c8cf624801d0', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Desert Chassis', 'red_mountain_olympus_mons_ly07_desert_chassis', 'Desert',
        'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1500, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('93a65506-44aa-4d93-b184-7d72b8cf5d9b', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Military Chassis', 'red_mountain_olympus_mons_ly07_military_chassis',
        'Military', 'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1530, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('30b40880-ffa3-4a17-883d-08de0bf1b479', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Gold Chassis', 'red_mountain_olympus_mons_ly07_gold_chassis', 'Gold',
        'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1590, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('bd283770-0d77-45ea-b4d2-6286351eecc7', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Evo Chassis', 'red_mountain_olympus_mons_ly07_evo_chassis', 'Evo',
        'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1650, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('337b3c82-61ae-4959-bcf0-50f0985f8ed6', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Green Yellow Chassis', 'red_mountain_olympus_mons_ly07_green_yellow_chassis',
        'Green Yellow', 'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1530, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('0a5b3678-f6c0-4f53-8a1c-e3c29f3a1632', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Villain Chassis', 'red_mountain_olympus_mons_ly07_villain_chassis', 'Villain',
        'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1620, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('5d213b5d-5d29-4e3d-bdf8-d9867c574d7f', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Navy Chassis', 'red_mountain_olympus_mons_ly07_navy_chassis', 'Navy',
        'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1500, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('beb40230-580e-4aa0-87c9-0aac27edcbb6', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Irradiated Chassis', 'red_mountain_olympus_mons_ly07_irradiated_chassis',
        'Irradiated', 'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1740, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('a8e507cd-f874-4e6a-a321-a4b04e171ba9', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Beetle Chassis', 'red_mountain_olympus_mons_ly07_beetle_chassis', 'Beetle',
        'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1560, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('500128ca-7544-4250-aad4-aa1ae8a8ad1a', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Red White Chassis', 'red_mountain_olympus_mons_ly07_red_white_chassis',
        'Red White', 'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1800, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('683c4461-a291-4245-9120-56c67f091fba', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Nautical Chassis', 'red_mountain_olympus_mons_ly07_nautical_chassis',
        'Nautical', 'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1650, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('35deaa58-d05d-4a8c-8a84-178358c46647', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Vintage Chassis', 'red_mountain_olympus_mons_ly07_vintage_chassis', 'Vintage',
        'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1500, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('1ca58e88-6280-427c-b6d2-25928e7bd292', '953ad4fc-3aa9-471f-a852-f39e9f36cd04',
        'Red Mountain Olympus Mons LY07 Red Blue Chassis', 'red_mountain_olympus_mons_ly07_red_blue_chassis',
        'Red Blue', 'Olympus Mons LY07', 80, 2, 2, 1, 1750, 1710, 1000);

INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('2115a44e-591b-4cff-a85d-425ec2e5c07d', 'a5966250-4972-425f-b2d4-433e3820741a',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'RARE',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_red-hex.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-hex.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-hex.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-hex_avatar.png',
        'Red Mountain Olympus Mons LY07 Red Hex', 'red_mountain_olympus_mons_ly07_red_hex', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('22116dc6-c9e5-43bc-996d-af11d45b7e1c', '24a4ebc1-e785-4b72-83e8-c8cf624801d0',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_desert.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_desert.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_desert.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_desert_avatar.png',
        'Red Mountain Olympus Mons LY07 Desert', 'red_mountain_olympus_mons_ly07_desert', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('233c7bcd-0759-49ce-907d-c224764dab5e', '93a65506-44aa-4d93-b184-7d72b8cf5d9b',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'COLOSSAL',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_military.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_military.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_military.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_military_avatar.png',
        'Red Mountain Olympus Mons LY07 Military', 'red_mountain_olympus_mons_ly07_military', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('309b1140-ed37-4943-9bc2-373ea7b1052e', '30b40880-ffa3-4a17-883d-08de0bf1b479',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'LEGENDARY',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_gold.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_gold.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_gold.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_gold_avatar.png',
        'Red Mountain Olympus Mons LY07 Gold', 'red_mountain_olympus_mons_ly07_gold', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('31a54aef-7c98-4952-81ca-13b5b7328400', 'bd283770-0d77-45ea-b4d2-6286351eecc7',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'EXOTIC',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_evo.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_evo.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_evo.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_evo_avatar.png',
        'Red Mountain Olympus Mons LY07 Evo', 'red_mountain_olympus_mons_ly07_evo', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('374aeeea-05c9-4e1c-8270-33441ab37d54', '337b3c82-61ae-4959-bcf0-50f0985f8ed6',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'COLOSSAL',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_green-yellow.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_green-yellow.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_green-yellow.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_green-yellow_avatar.png',
        'Red Mountain Olympus Mons LY07 Green Yellow', 'red_mountain_olympus_mons_ly07_green_yellow', FALSE,
        'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('67155142-77a7-47b6-bdb5-578fdf7ed825', '0a5b3678-f6c0-4f53-8a1c-e3c29f3a1632',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'ELITE_LEGENDARY',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_villain.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_villain.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_villain.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_villain_avatar.png',
        'Red Mountain Olympus Mons LY07 Villain', 'red_mountain_olympus_mons_ly07_villain', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('885ecaba-b078-448b-98c9-ce96895492ab', '5d213b5d-5d29-4e3d-bdf8-d9867c574d7f',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_navy.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_navy.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_navy.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_navy_avatar.png',
        'Red Mountain Olympus Mons LY07 Navy', 'red_mountain_olympus_mons_ly07_navy', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('8b2730ba-7340-4e96-b94b-d818681b5598', 'beb40230-580e-4aa0-87c9-0aac27edcbb6',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'MYTHIC',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_irradiated.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_irradiated.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_irradiated.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_irradiated_avatar.png',
        'Red Mountain Olympus Mons LY07 Irradiated', 'red_mountain_olympus_mons_ly07_irradiated', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('a20e2e51-3051-427a-a488-777e6af77768', 'a8e507cd-f874-4e6a-a321-a4b04e171ba9',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'RARE',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_beetle.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_beetle.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_beetle.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_beetle_avatar.png',
        'Red Mountain Olympus Mons LY07 Beetle', 'red_mountain_olympus_mons_ly07_beetle', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('af1d66c8-3188-4e73-acf9-253b9e302631', '500128ca-7544-4250-aad4-aa1ae8a8ad1a',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'DEUS_EX',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_red-white.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-white.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-white.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-white_avatar.png',
        'Red Mountain Olympus Mons LY07 Red White', 'red_mountain_olympus_mons_ly07_red_white', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('cb3f12a2-481f-4289-8416-42b390da91d7', '683c4461-a291-4245-9120-56c67f091fba',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'ULTRA_RARE',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_nautical.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_nautical.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_nautical.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_nautical_avatar.png',
        'Red Mountain Olympus Mons LY07 Nautical', 'red_mountain_olympus_mons_ly07_nautical', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('d9fd6437-82bc-46b8-9922-93ee32200aef', '35deaa58-d05d-4a8c-8a84-178358c46647',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_vintage.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_vintage.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_vintage.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_vintage_avatar.png',
        'Red Mountain Olympus Mons LY07 Vintage', 'red_mountain_olympus_mons_ly07_vintage', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('e7e7b940-0fea-493d-9dce-863abdd45760', '1ca58e88-6280-427c-b6d2-25928e7bd292',
        '98bf7bb3-1a7c-4f21-8843-458d62884060', 'GUARDIAN',
        'https://afiles.ninja-cdn.com/passport/genesis/img/red-mountain_olympus-mons-ly07_red-blue.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-blue.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/red-mountain_olympus-mons-ly07_red-blue.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_red-blue_avatar.png',
        'Red Mountain Olympus Mons LY07 Red Blue', 'red_mountain_olympus_mons_ly07_red_blue', FALSE, 'War Machine');

INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'a5966250-4972-425f-b2d4-433e3820741a', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '24a4ebc1-e785-4b72-83e8-c8cf624801d0', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '93a65506-44aa-4d93-b184-7d72b8cf5d9b', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '30b40880-ffa3-4a17-883d-08de0bf1b479', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'bd283770-0d77-45ea-b4d2-6286351eecc7', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '337b3c82-61ae-4959-bcf0-50f0985f8ed6', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '0a5b3678-f6c0-4f53-8a1c-e3c29f3a1632', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '5d213b5d-5d29-4e3d-bdf8-d9867c574d7f', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'beb40230-580e-4aa0-87c9-0aac27edcbb6', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'a8e507cd-f874-4e6a-a321-a4b04e171ba9', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '500128ca-7544-4250-aad4-aa1ae8a8ad1a', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '683c4461-a291-4245-9120-56c67f091fba', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '35deaa58-d05d-4a8c-8a84-178358c46647', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '1ca58e88-6280-427c-b6d2-25928e7bd292', 0);

INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'a5966250-4972-425f-b2d4-433e3820741a', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'a5966250-4972-425f-b2d4-433e3820741a', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'a5966250-4972-425f-b2d4-433e3820741a', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'a5966250-4972-425f-b2d4-433e3820741a', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '24a4ebc1-e785-4b72-83e8-c8cf624801d0', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '24a4ebc1-e785-4b72-83e8-c8cf624801d0', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '24a4ebc1-e785-4b72-83e8-c8cf624801d0', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '24a4ebc1-e785-4b72-83e8-c8cf624801d0', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '93a65506-44aa-4d93-b184-7d72b8cf5d9b', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '93a65506-44aa-4d93-b184-7d72b8cf5d9b', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '93a65506-44aa-4d93-b184-7d72b8cf5d9b', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '93a65506-44aa-4d93-b184-7d72b8cf5d9b', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '30b40880-ffa3-4a17-883d-08de0bf1b479', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '30b40880-ffa3-4a17-883d-08de0bf1b479', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '30b40880-ffa3-4a17-883d-08de0bf1b479', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '30b40880-ffa3-4a17-883d-08de0bf1b479', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'bd283770-0d77-45ea-b4d2-6286351eecc7', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'bd283770-0d77-45ea-b4d2-6286351eecc7', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'bd283770-0d77-45ea-b4d2-6286351eecc7', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'bd283770-0d77-45ea-b4d2-6286351eecc7', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '337b3c82-61ae-4959-bcf0-50f0985f8ed6', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '337b3c82-61ae-4959-bcf0-50f0985f8ed6', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '337b3c82-61ae-4959-bcf0-50f0985f8ed6', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '337b3c82-61ae-4959-bcf0-50f0985f8ed6', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '0a5b3678-f6c0-4f53-8a1c-e3c29f3a1632', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '0a5b3678-f6c0-4f53-8a1c-e3c29f3a1632', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '0a5b3678-f6c0-4f53-8a1c-e3c29f3a1632', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '0a5b3678-f6c0-4f53-8a1c-e3c29f3a1632', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '5d213b5d-5d29-4e3d-bdf8-d9867c574d7f', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '5d213b5d-5d29-4e3d-bdf8-d9867c574d7f', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '5d213b5d-5d29-4e3d-bdf8-d9867c574d7f', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '5d213b5d-5d29-4e3d-bdf8-d9867c574d7f', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'beb40230-580e-4aa0-87c9-0aac27edcbb6', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'beb40230-580e-4aa0-87c9-0aac27edcbb6', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'beb40230-580e-4aa0-87c9-0aac27edcbb6', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'beb40230-580e-4aa0-87c9-0aac27edcbb6', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'a8e507cd-f874-4e6a-a321-a4b04e171ba9', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', 'a8e507cd-f874-4e6a-a321-a4b04e171ba9', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'a8e507cd-f874-4e6a-a321-a4b04e171ba9', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'a8e507cd-f874-4e6a-a321-a4b04e171ba9', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '500128ca-7544-4250-aad4-aa1ae8a8ad1a', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '500128ca-7544-4250-aad4-aa1ae8a8ad1a', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '500128ca-7544-4250-aad4-aa1ae8a8ad1a', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '500128ca-7544-4250-aad4-aa1ae8a8ad1a', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '683c4461-a291-4245-9120-56c67f091fba', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '683c4461-a291-4245-9120-56c67f091fba', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '683c4461-a291-4245-9120-56c67f091fba', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '683c4461-a291-4245-9120-56c67f091fba', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '35deaa58-d05d-4a8c-8a84-178358c46647', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '35deaa58-d05d-4a8c-8a84-178358c46647', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '35deaa58-d05d-4a8c-8a84-178358c46647', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '35deaa58-d05d-4a8c-8a84-178358c46647', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '1ca58e88-6280-427c-b6d2-25928e7bd292', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('daa6c1b0-e6ae-409a-a544-bfe7212d6f45', '1ca58e88-6280-427c-b6d2-25928e7bd292', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '1ca58e88-6280-427c-b6d2-25928e7bd292', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '1ca58e88-6280-427c-b6d2-25928e7bd292', 1, 'TURRET');

-- XSYN Existing ZHI Items
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('9c4cd43d-35ed-4d3a-860c-d7ff00093dd6', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Static Chassis', 'zaibatsu_tenshi_mk1_static_chassis', 'Static', 'Tenshi Mk1', 100, 2, 2,
        1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('f4da1432-0618-47c2-9337-9cda3e8125ba', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Purple Haze Chassis', 'zaibatsu_tenshi_mk1_purple_haze_chassis', 'Purple Haze',
        'Tenshi Mk1', 116, 2, 2, 1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('bc19d28d-3fa4-45e5-86dd-1cf69bc3daee', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Chalky Neon Chassis', 'zaibatsu_tenshi_mk1_chalky_neon_chassis', 'Chalky Neon',
        'Tenshi Mk1', 120, 2, 2, 1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('0b297a84-0a07-4e09-87ba-ed2a877d7c4f', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Evangelica Chassis', 'zaibatsu_tenshi_mk1_evangelica_chassis', 'Evangelica', 'Tenshi Mk1',
        114, 2, 2, 1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('81f1fd1b-458c-499b-bbec-7560f8bea842', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Destroyer Chassis', 'zaibatsu_tenshi_mk1_destroyer_chassis', 'Destroyer', 'Tenshi Mk1',
        112, 2, 2, 1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('b2c43cd1-86ce-4154-b3d5-22776ed3a727', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Gumdan Chassis', 'zaibatsu_tenshi_mk1_gumdan_chassis', 'Gumdan', 'Tenshi Mk1', 110, 2, 2,
        1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('e1dc2da7-78c4-400e-a302-a40b2d27a3c8', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Gold Pattern Chassis', 'zaibatsu_tenshi_mk1_gold_pattern_chassis', 'Gold Pattern',
        'Tenshi Mk1', 102, 2, 2, 1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('f9cfa25b-02bc-4a3a-9a31-9a8a4a073714', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Vector Chassis', 'zaibatsu_tenshi_mk1_vector_chassis', 'Vector', 'Tenshi Mk1', 102, 2, 2,
        1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('8a49a4ca-4a61-4f51-b296-b421c951c1f5', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Gold Chassis', 'zaibatsu_tenshi_mk1_gold_chassis', 'Gold', 'Tenshi Mk1', 106, 2, 2, 1,
        2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('276f62c9-23f3-443c-8fad-90fc4d074009', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Warden Chassis', 'zaibatsu_tenshi_mk1_warden_chassis', 'Warden', 'Tenshi Mk1', 104, 2, 2,
        1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('4f6180a5-4d0c-44ca-9422-a4d28c3b863a', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Cherry Blossom Chassis', 'zaibatsu_tenshi_mk1_cherry_blossom_chassis', 'Cherry Blossom',
        'Tenshi Mk1', 104, 2, 2, 1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('1d76a559-3c6d-4f26-a108-a3b5549e35ab', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 White Gold Chassis', 'zaibatsu_tenshi_mk1_white_gold_chassis', 'White Gold', 'Tenshi Mk1',
        100, 2, 2, 1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('d0c291de-1edb-4b54-9559-ed3587034051', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 Black Digi Chassis', 'zaibatsu_tenshi_mk1_black_digi_chassis', 'Black Digi', 'Tenshi Mk1',
        100, 2, 2, 1, 2500, 1000, 1000);
INSERT INTO blueprint_chassis (id, brand_id, label, slug, skin, model, shield_recharge_rate, weapon_hardpoints,
                               turret_hardpoints, utility_slots, speed, max_hitpoints, max_shield)
VALUES ('ac408b39-ba91-4a0e-9414-a37af0ea6314', '2b203c87-ad8c-4ce2-af17-e079835fdbcb',
        'Zaibatsu Tenshi Mk1 White Neon Chassis', 'zaibatsu_tenshi_mk1_white_neon_chassis', 'White Neon', 'Tenshi Mk1',
        108, 2, 2, 1, 2500, 1000, 1000);


INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('3c3ac5df-bf8e-4cba-8cee-7f05c003fa36', '9c4cd43d-35ed-4d3a-860c-d7ff00093dd6',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_static.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_static.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_static.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_static_avatar.png',
        'Zaibatsu Tenshi Mk1 Static', 'zaibatsu_tenshi_mk1_static', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('46d96d77-12b8-4bb1-b060-8b576ff628a7', 'f4da1432-0618-47c2-9337-9cda3e8125ba',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'MYTHIC',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_purple-haze.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_purple-haze.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_purple-haze.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_purple-haze_avatar.png',
        'Zaibatsu Tenshi Mk1 Purple Haze', 'zaibatsu_tenshi_mk1_purple_haze', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('4bf3173b-8c63-42f5-b563-99ab3c198da7', 'bc19d28d-3fa4-45e5-86dd-1cf69bc3daee',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'DEUS_EX',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_chalky-neon.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_chalky-neon.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_chalky-neon.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_chalky-neon_avatar.png',
        'Zaibatsu Tenshi Mk1 Chalky Neon', 'zaibatsu_tenshi_mk1_chalky_neon', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('589a1273-261a-4c2b-85cf-a9adb2c34a96', '0b297a84-0a07-4e09-87ba-ed2a877d7c4f',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'GUARDIAN',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_evangelion.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_evangelion.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_evangelion.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/red-mountain_olympus-mons-ly07_evo_avatar.png',
        'Zaibatsu Tenshi Mk1 Evangelica', 'zaibatsu_tenshi_mk1_evangelica', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('5baf5d79-a250-4080-af9a-bb1ae70014d9', '81f1fd1b-458c-499b-bbec-7560f8bea842',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'EXOTIC',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_destroyer.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_destroyer.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_destroyer.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_destroyer_avatar.png',
        'Zaibatsu Tenshi Mk1 Destroyer', 'zaibatsu_tenshi_mk1_destroyer', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('5d13d78a-cec5-478c-9537-45fe50e74af1', 'b2c43cd1-86ce-4154-b3d5-22776ed3a727',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'ULTRA_RARE',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_gundam.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_gundam.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_gundam.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_gumdan_avatar.png',
        'Zaibatsu Tenshi Mk1 Gumdan', 'zaibatsu_tenshi_mk1_gumdan', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('7cedfc8f-553f-4926-82c0-5ad4ac420ece', 'e1dc2da7-78c4-400e-a302-a40b2d27a3c8',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'COLOSSAL',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_white-gold-pattern.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_white-gold-pattern.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_white-gold-pattern.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_white-gold_avatar.png',
        'Zaibatsu Tenshi Mk1 Gold Pattern', 'zaibatsu_tenshi_mk1_gold_pattern', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('885dd343-fc20-47ca-b320-c3fa9be0e2ea', 'f9cfa25b-02bc-4a3a-9a31-9a8a4a073714',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'COLOSSAL',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_vector.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_vector.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_vector.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_vector_avatar.png',
        'Zaibatsu Tenshi Mk1 Vector', 'zaibatsu_tenshi_mk1_vector', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('9e5ca941-904d-4f27-a102-3cf5ffa639e0', '8a49a4ca-4a61-4f51-b296-b421c951c1f5',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'LEGENDARY',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_gold.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_gold.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_gold.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_gold_avatar.png',
        'Zaibatsu Tenshi Mk1 Gold', 'zaibatsu_tenshi_mk1_gold', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('a51d1f0a-ebc9-4ac5-b61e-4bbbc0f9b78a', '276f62c9-23f3-443c-8fad-90fc4d074009',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'RARE',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_warden.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_warden.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_warden.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_warden_avatar.png',
        'Zaibatsu Tenshi Mk1 Warden', 'zaibatsu_tenshi_mk1_warden', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('b61d8538-27c8-4c92-9237-edbeb480fb56', '4f6180a5-4d0c-44ca-9422-a4d28c3b863a',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'RARE',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_cherry-blossom.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_cherry-blossom.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_cherry-blossom.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_cherry-blossom_avatar.png',
        'Zaibatsu Tenshi Mk1 Cherry Blossom', 'zaibatsu_tenshi_mk1_cherry_blossom', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('c41dd683-0a5d-4d04-bb6e-cd084f74b3cf', '1d76a559-3c6d-4f26-a108-a3b5549e35ab',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_white-gold.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_white-gold.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_white-gold.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_white-gold-pattern_avatar.png',
        'Zaibatsu Tenshi Mk1 White Gold', 'zaibatsu_tenshi_mk1_white_gold', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('c71f46fc-6feb-4ccf-8e01-cfbc3067ea6b', 'd0c291de-1edb-4b54-9559-ed3587034051',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'MEGA',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_black-digi.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_black-digi.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_black-digi.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_black-digi_avatar.png',
        'Zaibatsu Tenshi Mk1 Black Digi', 'zaibatsu_tenshi_mk1_black_digi', FALSE, 'War Machine');
INSERT INTO templates (id, blueprint_chassis_id, faction_id, tier, image_url, animation_url, card_animation_url,
                       avatar_url, label, slug, is_default, asset_type)
VALUES ('e0caf423-16b5-42d5-a9dc-c735eabf41fa', 'ac408b39-ba91-4a0e-9414-a37af0ea6314',
        '880db344-e405-428d-84e5-6ebebab1fe6d', 'ELITE_LEGENDARY',
        'https://afiles.ninja-cdn.com/passport/genesis/img/zaibatsu_tenshi-mk1_neon.png',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_neon.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/webm/zaibatsu_tenshi-mk1_neon.webm',
        'https://afiles.ninja-cdn.com/passport/genesis/avatar/zaibatsu_tenshi-mk1_neon_avatar.png',
        'Zaibatsu Tenshi Mk1 White Neon', 'zaibatsu_tenshi_mk1_white_neon', FALSE, 'War Machine');


INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '9c4cd43d-35ed-4d3a-860c-d7ff00093dd6', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'f4da1432-0618-47c2-9337-9cda3e8125ba', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'bc19d28d-3fa4-45e5-86dd-1cf69bc3daee', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '0b297a84-0a07-4e09-87ba-ed2a877d7c4f', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '81f1fd1b-458c-499b-bbec-7560f8bea842', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'b2c43cd1-86ce-4154-b3d5-22776ed3a727', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'e1dc2da7-78c4-400e-a302-a40b2d27a3c8', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'f9cfa25b-02bc-4a3a-9a31-9a8a4a073714', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '8a49a4ca-4a61-4f51-b296-b421c951c1f5', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '276f62c9-23f3-443c-8fad-90fc4d074009', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '4f6180a5-4d0c-44ca-9422-a4d28c3b863a', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', '1d76a559-3c6d-4f26-a108-a3b5549e35ab', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'd0c291de-1edb-4b54-9559-ed3587034051', 0);
INSERT INTO blueprint_chassis_blueprint_modules (blueprint_module_id, blueprint_chassis_id, slot_number)
VALUES ('6537f960-aa85-4e68-80b1-fe6e754a1436', 'ac408b39-ba91-4a0e-9414-a37af0ea6314', 0);

INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', '9c4cd43d-35ed-4d3a-860c-d7ff00093dd6', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', '9c4cd43d-35ed-4d3a-860c-d7ff00093dd6', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '9c4cd43d-35ed-4d3a-860c-d7ff00093dd6', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '9c4cd43d-35ed-4d3a-860c-d7ff00093dd6', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', 'f4da1432-0618-47c2-9337-9cda3e8125ba', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', 'f4da1432-0618-47c2-9337-9cda3e8125ba', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'f4da1432-0618-47c2-9337-9cda3e8125ba', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'f4da1432-0618-47c2-9337-9cda3e8125ba', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', 'bc19d28d-3fa4-45e5-86dd-1cf69bc3daee', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', 'bc19d28d-3fa4-45e5-86dd-1cf69bc3daee', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'bc19d28d-3fa4-45e5-86dd-1cf69bc3daee', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'bc19d28d-3fa4-45e5-86dd-1cf69bc3daee', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', '0b297a84-0a07-4e09-87ba-ed2a877d7c4f', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', '0b297a84-0a07-4e09-87ba-ed2a877d7c4f', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '0b297a84-0a07-4e09-87ba-ed2a877d7c4f', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '0b297a84-0a07-4e09-87ba-ed2a877d7c4f', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', '81f1fd1b-458c-499b-bbec-7560f8bea842', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', '81f1fd1b-458c-499b-bbec-7560f8bea842', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '81f1fd1b-458c-499b-bbec-7560f8bea842', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '81f1fd1b-458c-499b-bbec-7560f8bea842', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', 'b2c43cd1-86ce-4154-b3d5-22776ed3a727', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', 'b2c43cd1-86ce-4154-b3d5-22776ed3a727', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'b2c43cd1-86ce-4154-b3d5-22776ed3a727', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'b2c43cd1-86ce-4154-b3d5-22776ed3a727', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', 'e1dc2da7-78c4-400e-a302-a40b2d27a3c8', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', 'e1dc2da7-78c4-400e-a302-a40b2d27a3c8', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'e1dc2da7-78c4-400e-a302-a40b2d27a3c8', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'e1dc2da7-78c4-400e-a302-a40b2d27a3c8', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', 'f9cfa25b-02bc-4a3a-9a31-9a8a4a073714', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', 'f9cfa25b-02bc-4a3a-9a31-9a8a4a073714', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'f9cfa25b-02bc-4a3a-9a31-9a8a4a073714', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'f9cfa25b-02bc-4a3a-9a31-9a8a4a073714', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', '8a49a4ca-4a61-4f51-b296-b421c951c1f5', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', '8a49a4ca-4a61-4f51-b296-b421c951c1f5', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '8a49a4ca-4a61-4f51-b296-b421c951c1f5', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '8a49a4ca-4a61-4f51-b296-b421c951c1f5', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', '276f62c9-23f3-443c-8fad-90fc4d074009', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', '276f62c9-23f3-443c-8fad-90fc4d074009', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '276f62c9-23f3-443c-8fad-90fc4d074009', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '276f62c9-23f3-443c-8fad-90fc4d074009', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', '4f6180a5-4d0c-44ca-9422-a4d28c3b863a', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', '4f6180a5-4d0c-44ca-9422-a4d28c3b863a', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '4f6180a5-4d0c-44ca-9422-a4d28c3b863a', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '4f6180a5-4d0c-44ca-9422-a4d28c3b863a', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', '1d76a559-3c6d-4f26-a108-a3b5549e35ab', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', '1d76a559-3c6d-4f26-a108-a3b5549e35ab', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '1d76a559-3c6d-4f26-a108-a3b5549e35ab', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', '1d76a559-3c6d-4f26-a108-a3b5549e35ab', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', 'd0c291de-1edb-4b54-9559-ed3587034051', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', 'd0c291de-1edb-4b54-9559-ed3587034051', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'd0c291de-1edb-4b54-9559-ed3587034051', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'd0c291de-1edb-4b54-9559-ed3587034051', 1, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('06216d51-e57f-4f60-adee-24d817a397ab', 'ac408b39-ba91-4a0e-9414-a37af0ea6314', 0, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('1b8a0178-b7ab-4016-b203-6ba557107a97', 'ac408b39-ba91-4a0e-9414-a37af0ea6314', 1, 'ARM');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'ac408b39-ba91-4a0e-9414-a37af0ea6314', 0, 'TURRET');
INSERT INTO blueprint_chassis_blueprint_weapons (blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
VALUES ('347cdf83-a245-4552-94b3-68faa88fbf79', 'ac408b39-ba91-4a0e-9414-a37af0ea6314', 1, 'TURRET');
