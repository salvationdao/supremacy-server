/*
  WEAPON SKINS
 */

CREATE TABLE blueprint_weapon_skin
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    label       TEXT        NOT NULL,
    weapon_type WEAPON_TYPE NOT NULL,
    tier        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE weapon_skin
(
    id           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_id UUID        NOT NULL REFERENCES blueprint_weapon_skin (id),
    owner_id     UUID        NOT NULL REFERENCES players (id),
    label        TEXT        NOT NULL,
    weapon_type  WEAPON_TYPE NOT NULL,
    equipped_on  UUID REFERENCES chassis (id),
    tier         TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


/*
  WEAPONS
 */

ALTER TABLE blueprint_weapons
    DROP COLUMN IF EXISTS weapon_type,
    ADD COLUMN game_client_weapon_id   UUID,
    ADD COLUMN weapon_type             WEAPON_TYPE,
    ADD COLUMN collection              COLLECTION  NOT NULL DEFAULT 'supremacy-general',
    ADD COLUMN default_damage_type     DAMAGE_TYPE NOT NULL DEFAULT 'Kinetic',
    ADD COLUMN damage_falloff          INT     DEFAULT 0,
    ADD COLUMN damage_falloff_rate     INT     DEFAULT 0,
    ADD COLUMN spread                  NUMERIC DEFAULT 0,
    ADD COLUMN rate_of_fire            NUMERIC DEFAULT 0,
    ADD COLUMN radius                  INT     DEFAULT 0,
    ADD COLUMN radial_does_full_damage BOOL    DEFAULT TRUE,
    ADD COLUMN projectile_speed        NUMERIC DEFAULT 0,
    ADD COLUMN max_ammo                INT     DEFAULT 0,
    ADD COLUMN energy_cost             NUMERIC DEFAULT 0;

UPDATE blueprint_weapons
SET weapon_type           = 'Sniper Rifle',
    game_client_weapon_id = 'a155bef8-f0e1-4d11-8a23-a93b0bb74d10'
WHERE label = 'Sniper Rifle';

UPDATE blueprint_weapons
SET weapon_type           = 'Sword',
    game_client_weapon_id = '6109e547-5a48-4a76-a3f2-e73ef41505b3'
WHERE label = 'Laser Sword';

UPDATE blueprint_weapons
SET weapon_type           = 'Missile Launcher',
    game_client_weapon_id = '7c082a33-ff87-454f-bf8c-925945dd0ff4'
WHERE label = 'Rocket Pod';

UPDATE blueprint_weapons
SET weapon_type           = 'Cannon',
    game_client_weapon_id = 'a009fbf9-4fe3-48b0-8f34-e207c2b355dc'
WHERE label = 'Auto Cannon';

UPDATE blueprint_weapons
SET weapon_type           = 'Plasma Gun',
    game_client_weapon_id = '26f37473-ccd6-47d0-993e-2b82d725617d'
WHERE label = 'Plasma Rifle';

UPDATE blueprint_weapons
SET weapon_type           = 'Sword',
    game_client_weapon_id = '02c27475-c0ea-4825-8739-9a0b2cdc4201'
WHERE label = 'Sword';

ALTER TABLE blueprint_weapons
    ALTER COLUMN weapon_type SET NOT NULL;

ALTER TABLE weapons
    DROP COLUMN IF EXISTS weapon_type,
    ADD COLUMN blueprint_id             UUID REFERENCES blueprint_weapons,
    ADD COLUMN default_damage_type      DAMAGE_TYPE NOT NULL DEFAULT 'Kinetic',
    ADD COLUMN genesis_token_id         NUMERIC,
    ADD COLUMN limited_release_token_id NUMERIC,
    ADD COLUMN weapon_type              WEAPON_TYPE,
    ADD COLUMN owner_id                 UUID REFERENCES players (id),
    ADD COLUMN damage_falloff           INT     DEFAULT 0,
    ADD COLUMN damage_falloff_rate      INT     DEFAULT 0,
    ADD COLUMN spread                   NUMERIC DEFAULT 0,
    ADD COLUMN rate_of_fire             NUMERIC DEFAULT 0,
    ADD COLUMN radius                   INT     DEFAULT 0,
    ADD COLUMN radial_does_full_damage  BOOL    DEFAULT TRUE,
    ADD COLUMN projectile_speed         NUMERIC DEFAULT 0,
    ADD COLUMN energy_cost              NUMERIC DEFAULT 0,
    ADD COLUMN max_ammo                 INT     DEFAULT 0;


UPDATE weapons
SET weapon_type = 'Sniper Rifle',
    label       = 'Sniper Rifle'
WHERE label = 'Sniper Rifle'
   OR label = 'Zaibatsu Heavy Industries Sniper Rifle';

UPDATE weapons
SET weapon_type = 'Sword',
    label       = 'Laser Sword'
WHERE label = 'Laser Sword'
   OR label = 'Zaibatsu Heavy Industries Laser Sword';

UPDATE weapons
SET weapon_type = 'Missile Launcher',
    label       = 'Rocket Pod'
WHERE label = 'Rocket Pod'
   OR label = 'Red Mountain Offworld Mining Corporation Rocket Pod'
   OR label = 'Zaibatsu Heavy Industries Rocket Pod';

UPDATE weapons
SET weapon_type = 'Cannon',
    label       = 'Auto Cannon'
WHERE label = 'Auto Cannon'
   OR label = 'Red Mountain Offworld Mining Corporation Auto Cannon';

UPDATE weapons
SET weapon_type = 'Plasma Gun',
    label       = 'Plasma Rifle'
WHERE label = 'Plasma Rifle'
   OR label = 'Boston Cybernetics Plasma Rifle';

UPDATE weapons
SET weapon_type = 'Sword',
    label       = 'Sword'
WHERE label = 'Sword'
   OR label = 'Boston Cybernetics Sword';

WITH weapon_owners AS (SELECT m.owner_id, cw.weapon_id
                       FROM chassis_weapons cw
                                INNER JOIN mechs m ON cw.chassis_id = m.chassis_id)
UPDATE weapons w
SET owner_id = weapon_owners.owner_id
FROM weapon_owners
WHERE w.id = weapon_owners.weapon_id;

-- This inserts a new collection_items entry for each weapons and updates the weapons table with token id

WITH weapon AS (SELECT 'weapon' AS item_type, id FROM weapons)
INSERT
INTO collection_items (token_id, item_type, item_id)
SELECT NEXTVAL('collection_general'), weapon.item_type::ITEM_TYPE, weapon.id
FROM weapon;

-- this updates all genesis_token_id for weapons that are in genesis
WITH genesis AS (SELECT external_token_id, m.collection_slug, m.chassis_id, _cw.weapon_id
                 FROM chassis_weapons _cw
                          INNER JOIN mechs m ON m.chassis_id = _cw.chassis_id
                 WHERE m.collection_slug = 'supremacy-genesis')
UPDATE weapons w
SET genesis_token_id = genesis.external_token_id
FROM genesis
WHERE w.id = genesis.weapon_id;

-- this updates all limited release for weapons that are in genesis
WITH limited_release AS (SELECT external_token_id, m.collection_slug, m.chassis_id, _cw.weapon_id
                         FROM chassis_weapons _cw
                                  INNER JOIN mechs m ON m.chassis_id = _cw.chassis_id
                         WHERE m.collection_slug = 'supremacy-limited-release')
UPDATE weapons w
SET limited_release_token_id = limited_release.external_token_id
FROM limited_release
WHERE w.id = limited_release.weapon_id;

ALTER TABLE weapons
    ALTER COLUMN owner_id SET NOT NULL,
    ALTER COLUMN weapon_type SET NOT NULL;


-- update weapon stats
UPDATE weapons
SET damage                  = 20,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 3,
    rate_of_fire            = 0,
    radius                  = 100,
    projectile_speed        = 48000,
    radial_does_full_damage = TRUE,
    energy_cost             = 10,
    default_damage_type     = 'Energy'
WHERE label ILIKE 'Plasma Rifle'
   OR label ILIKE 'Boston Cybernetics Plasma Rifle';

UPDATE weapons
SET damage                  = 12,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 4,
    rate_of_fire            = 270,
    radius                  = 100,
    projectile_speed        = 36000,
    radial_does_full_damage = TRUE,
    energy_cost             = 10,
    default_damage_type     = 'Kinetic'
WHERE label ILIKE 'Auto Cannon'
   OR label ILIKE 'Red Mountain Offworld Mining Corporation Auto Cannon';

UPDATE weapons
SET damage                  = 130,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 3,
    rate_of_fire            = 48,
    radius                  = 100,
    projectile_speed        = 80000,
    radial_does_full_damage = TRUE,
    energy_cost             = 15,
    default_damage_type     = 'Kinetic'
WHERE label ILIKE 'Sniper Rifle'
   OR label ILIKE 'Zaibatsu Heavy Industries Sniper Rifle';

UPDATE weapons
SET damage                  = 70,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 3,
    rate_of_fire            = 0,
    radius                  = 850,
    projectile_speed        = 0,
    radial_does_full_damage = TRUE,
    energy_cost             = 15,
    default_damage_type     = 'Explosive'
WHERE label ILIKE 'Rocket Pod'
   OR label ILIKE 'Zaibatsu Heavy Industries Rocket Pod'
   OR label ILIKE 'Red Mountain Offworld Mining Corporation Rocket Pod';

UPDATE weapons
SET damage                  = 80,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 0,
    rate_of_fire            = 0,
    radius                  = 0,
    projectile_speed        = 0,
    radial_does_full_damage = TRUE,
    energy_cost             = 15,
    default_damage_type     = 'Kinetic'
WHERE label ILIKE 'Sword'
   OR label ILIKE 'Boston Cybernetics Sword';

UPDATE weapons
SET damage                  = 120,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 0,
    rate_of_fire            = 0,
    radius                  = 0,
    projectile_speed        = 0,
    radial_does_full_damage = TRUE,
    energy_cost             = 15,
    default_damage_type     = 'Energy'
WHERE label ILIKE 'Laser Sword'
   OR label ILIKE 'Zaibatsu Heavy Industries Laser Sword';

--  blueprint weapons
-- update weapon stats
UPDATE blueprint_weapons
SET damage                  = 20,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 3,
    rate_of_fire            = 250,
    radius                  = 100,
    projectile_speed        = 48000,
    radial_does_full_damage = TRUE,
    energy_cost             = 10,
    default_damage_type     = 'Energy'
WHERE label ILIKE 'Plasma Rifle'
   OR label ILIKE 'Boston Cybernetics Plasma Rifle';

UPDATE blueprint_weapons
SET damage                  = 12,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 4,
    rate_of_fire            = 270,
    radius                  = 100,
    projectile_speed        = 36000,
    radial_does_full_damage = TRUE,
    energy_cost             = 10,
    default_damage_type     = 'Kinetic'
WHERE label ILIKE 'Auto Cannon'
   OR label ILIKE 'Red Mountain Offworld Mining Corporation Auto Cannon';

UPDATE blueprint_weapons
SET damage                  = 130,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 3,
    rate_of_fire            = 48,
    radius                  = 100,
    projectile_speed        = 80000,
    radial_does_full_damage = TRUE,
    energy_cost             = 15,
    default_damage_type     = 'Kinetic'
WHERE label ILIKE 'Sniper Rifle'
   OR label ILIKE 'Zaibatsu Heavy Industries Sniper Rifle';

UPDATE blueprint_weapons
SET damage                  = 70,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 3,
    rate_of_fire            = 0,
    radius                  = 850,
    projectile_speed        = 0,
    radial_does_full_damage = TRUE,
    energy_cost             = 15,
    default_damage_type     = 'Explosive'
WHERE label ILIKE 'Rocket Pod'
   OR label ILIKE 'Zaibatsu Heavy Industries Rocket Pod'
   OR label ILIKE 'Red Mountain Offworld Mining Corporation Rocket Pod';

UPDATE blueprint_weapons
SET damage                  = 80,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 0,
    rate_of_fire            = 0,
    radius                  = 0,
    projectile_speed        = 0,
    radial_does_full_damage = TRUE,
    energy_cost             = 15,
    default_damage_type     = 'Kinetic'
WHERE label ILIKE 'Sword'
   OR label ILIKE 'Boston Cybernetics Sword';


UPDATE blueprint_weapons
SET damage                  = 120,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 0,
    rate_of_fire            = 0,
    radius                  = 0,
    projectile_speed        = 0,
    radial_does_full_damage = TRUE,
    energy_cost             = 15,
    default_damage_type     = 'Energy'
WHERE label ILIKE 'Laser Sword'
   OR label ILIKE 'Zaibatsu Heavy Industries Laser Sword';


ALTER TABLE chassis_weapons
    DROP CONSTRAINT chassis_weapons_chassis_id_slot_number_mount_location_key;

-- on the mech_weapons join, we need to update slot numbers and then remove mount location column

UPDATE chassis_weapons mw
SET slot_number = 2
WHERE mw.mount_location = 'TURRET'
  AND mw.slot_number = 0;
UPDATE chassis_weapons mw
SET slot_number = 3
WHERE mw.mount_location = 'TURRET'
  AND mw.slot_number = 1;

ALTER TABLE chassis_weapons
    ADD UNIQUE (chassis_id, slot_number),
    DROP COLUMN mount_location;

--  update mech weapoon hardpoints
UPDATE chassis c
SET weapon_hardpoints = (SELECT COUNT(*) FROM chassis_weapons cw WHERE cw.chassis_id = c.id);

--  update blueprint mech weapoon hardpoints
UPDATE blueprint_chassis bc
SET weapon_hardpoints = (SELECT COUNT(*)
                         FROM blueprint_chassis_blueprint_weapons bcbw
                         WHERE bcbw.blueprint_chassis_id = bc.id);


-- below adds the blueprint ids for the weapons
UPDATE weapons w
SET blueprint_id = (SELECT id FROM blueprint_weapons bw WHERE bw.label = w.label);

ALTER TABLE weapons
    ALTER COLUMN blueprint_id SET NOT NULL;

