CREATE TABLE weapon_skin
(
    id           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_id UUID        NOT NULL REFERENCES blueprint_weapon_skin (id),
    owner_id     UUID        NOT NULL REFERENCES players (id),
    label        TEXT        NOT NULL,
    weapon_type  WEAPON_TYPE NOT NULL,
    equipped_on  UUID REFERENCES chassis (id),
    tier         TEXT        NOT NULL DEFAULT 'MEGA',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


/*
  WEAPONS
 */

ALTER TABLE weapons
    DROP COLUMN IF EXISTS weapon_type,
    ADD COLUMN blueprint_id             UUID REFERENCES blueprint_weapons(id),
    ADD COLUMN equipped_on              UUID REFERENCES chassis (id),
    ADD COLUMN default_damage_type      DAMAGE_TYPE NOT NULL DEFAULT 'Kinetic',
    ADD COLUMN genesis_token_id         BIGINT,
    ADD COLUMN limited_release_token_id BIGINT,
    ADD COLUMN weapon_type              WEAPON_TYPE,
    ADD COLUMN owner_id                 UUID REFERENCES players (id),
    ADD COLUMN damage_falloff           INT     DEFAULT 0,
    ADD COLUMN damage_falloff_rate      INT     DEFAULT 0,
    ADD COLUMN radius                   INT     DEFAULT 0,
    ADD COLUMN radius_damage_falloff    INT     DEFAULT 0,
    ADD COLUMN spread                   NUMERIC DEFAULT 0,
    ADD COLUMN rate_of_fire             NUMERIC DEFAULT 0,
    ADD COLUMN projectile_speed         NUMERIC DEFAULT 0,
    ADD COLUMN energy_cost              NUMERIC DEFAULT 0,
    ADD COLUMN is_melee                 BOOL        NOT NULL DEFAULT FALSE,
    ADD COLUMN tier                     TEXT        NOT NULL DEFAULT 'MEGA',
    ADD COLUMN max_ammo                 INT     DEFAULT 0;

UPDATE weapons
SET weapon_type = 'Sniper Rifle',
    label       = 'Sniper Rifle'
WHERE label = 'Sniper Rifle'
   OR label = 'Zaibatsu Heavy Industries Sniper Rifle';

UPDATE weapons
SET weapon_type = 'Sword',
    label       = 'Laser Sword',
    is_melee    = TRUE
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
    label       = 'Sword',
    is_melee    = TRUE
WHERE label = 'Sword'
   OR label = 'Boston Cybernetics Sword';

-- delete rocket pod joins
WITH wep AS (SELECT cw.chassis_id, cw.weapon_id, w.label
             FROM chassis_weapons cw
                      INNER JOIN weapons w ON cw.weapon_id = w.id
             WHERE w.label ILIKE '%Rocket Pod%')
DELETE
FROM chassis_weapons cw
WHERE cw.weapon_id IN (SELECT wep.weapon_id FROM wep);

-- delete rocket pod weapons
DELETE
FROM weapons w
WHERE w.label ILIKE '%Rocket Pod%';

-- temp column
ALTER TABLE weapons
    ADD COLUMN chassis_id UUID;

-- insert weapon and join per mech
WITH wm AS (
    WITH m AS (
        SELECT c.id, 'ZAI Rocket Pod' AS label, 'rocket_pod' AS slug
        FROM mechs
                 INNER JOIN chassis c ON mechs.chassis_id = c.id
        WHERE c.label ILIKE '%Zaibatsu%'
        )
        INSERT INTO weapons (label, slug, chassis_id, damage, weapon_type)
            SELECT m.label, m.slug, m.id, -1, 'Missile Launcher'::WEAPON_TYPE
            FROM m
            RETURNING id, chassis_id)
INSERT
INTO chassis_weapons(chassis_id, weapon_id, slot_number, mount_location)
SELECT wm.chassis_id, wm.id, 2, 'TURRET'
FROM wm;

WITH wm AS (
    WITH m AS (
        SELECT c.id, 'RMMC Rocket Pod' AS label, 'rocket_pod' AS slug
        FROM mechs
                 INNER JOIN chassis c ON mechs.chassis_id = c.id
        WHERE c.label ILIKE '%Mountain%'
        )
        INSERT INTO weapons (label, slug, chassis_id, damage, weapon_type)
            SELECT m.label, m.slug, m.id, -1, 'Missile Launcher'::WEAPON_TYPE
            FROM m
            RETURNING id, chassis_id)
INSERT
INTO chassis_weapons(chassis_id, weapon_id, slot_number, mount_location)
SELECT wm.chassis_id, wm.id, 2, 'TURRET'
FROM wm;

WITH wm AS (
    WITH m AS (
        SELECT c.id, 'BC Rocket Pod' AS label, 'rocket_pod' AS slug
        FROM mechs
                 INNER JOIN chassis c ON mechs.chassis_id = c.id
        WHERE c.label ILIKE '%Boston%'
        )
        INSERT INTO weapons (label, slug, chassis_id, damage, weapon_type)
            SELECT m.label, m.slug, m.id, -1, 'Missile Launcher'::WEAPON_TYPE
            FROM m
            RETURNING id, chassis_id)
INSERT
INTO chassis_weapons(chassis_id, weapon_id, slot_number, mount_location)
SELECT wm.chassis_id, wm.id, 2, 'TURRET'
FROM wm;


ALTER TABLE weapons
    DROP COLUMN chassis_id;



WITH weapon_owners AS (SELECT m.owner_id, cw.weapon_id
                       FROM chassis_weapons cw
                                INNER JOIN mechs m ON cw.chassis_id = m.chassis_id)
UPDATE weapons w
SET owner_id = weapon_owners.owner_id
FROM weapon_owners
WHERE w.id = weapon_owners.weapon_id;

-- This inserts a new collection_items entry for each weapons and updates the weapons table with token id

WITH weapon AS (SELECT 'weapon' AS item_type, id, tier, owner_id FROM weapons)
INSERT
INTO collection_items (token_id, item_type, item_id, tier, owner_id)
SELECT NEXTVAL('collection_general'), weapon.item_type::ITEM_TYPE, weapon.id, weapon.tier, weapon.owner_id
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
    ADD COLUMN allow_melee BOOL NOT NULL DEFAULT TRUE,
    ADD UNIQUE (chassis_id, slot_number),
    DROP COLUMN mount_location;

UPDATE chassis_weapons
SET allow_melee = FALSE
WHERE slot_number = 2;


--  update mech weapoon hardpoints
UPDATE chassis c
SET weapon_hardpoints = 3;

--  update blueprint mech weapoon hardpoints
UPDATE blueprint_chassis bc
SET weapon_hardpoints = 3;

-- set equipped on
WITH wsp AS (SELECT _w.id, mw.chassis_id
             FROM weapons _w
                      INNER JOIN chassis_weapons mw ON _w.id = mw.weapon_id)
UPDATE weapons w
SET equipped_on = wsp.chassis_id
FROM wsp
WHERE wsp.id = w.id;

-- below adds the blueprint ids for the weapons
-- UPDATE weapons w
-- SET blueprint_id = (SELECT id FROM blueprint_weapons bw WHERE bw.label = w.label);
WITH wpns AS (
    SELECT w.id, w.label, w.equipped_on, mm.brand_id
    FROM weapons w
             INNER JOIN chassis m ON m.id = w.equipped_on
             INNER JOIN blueprint_mechs mm ON mm.id = m.blueprint_id
)
UPDATE weapons w SET blueprint_id = (
    SELECT bw.id
    FROM blueprint_weapons bw
    WHERE wpns.label = bw.label and wpns.brand_id = bw.brand_id
)
FROM wpns
WHERE w.id = wpns.id;

ALTER TABLE weapons
    ALTER COLUMN blueprint_id SET NOT NULL;

-- update old blueprint chassis blueprint weapon joins
WITH wep AS (SELECT cbcbw.blueprint_chassis_id, cbcbw.blueprint_weapon_id, bpw.label
             FROM blueprint_chassis_blueprint_weapons cbcbw
                      INNER JOIN blueprint_weapons bpw ON cbcbw.blueprint_weapon_id = bpw.id
             WHERE bpw.label ILIKE '%Rocket Pod%')
DELETE
FROM blueprint_chassis_blueprint_weapons bpcbpw
WHERE bpcbpw.blueprint_weapon_id IN (SELECT wep.blueprint_weapon_id FROM wep);


WITH bpc AS (
    SELECT _bpc.id, mm.brand_id
    FROM blueprint_chassis _bpc
             INNER JOIN blueprint_mechs mm on _bpc.model_id = mm.id
)
INSERT
INTO blueprint_chassis_blueprint_weapons(blueprint_weapon_id, blueprint_chassis_id, slot_number, mount_location)
SELECT (
           SELECT bpw.id
           FROM blueprint_weapons bpw
           WHERE bpw.label ILIKE '%Rocket Pod%' and bpw.brand_id = bpc.brand_id
       ),
       bpc.id,
       2,
       'TURRET'
FROM bpc;

