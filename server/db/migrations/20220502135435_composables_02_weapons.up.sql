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
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    owner_id    UUID        NOT NULL REFERENCES players (id),
    label       TEXT        NOT NULL,
    weapon_type WEAPON_TYPE NOT NULL,
    equipped_on UUID REFERENCES chassis (id),
    tier        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


/*
  WEAPONS
 */

ALTER TABLE blueprint_weapons
    DROP COLUMN IF EXISTS weapon_type,
    ADD COLUMN game_client_weapon_id   UUID,
    ADD COLUMN weapon_type             WEAPON_TYPE,
    ADD COLUMN default_damage_typ      DAMAGE_TYPE NOT NULL DEFAULT 'Kinetic',
    ADD COLUMN damage_falloff          INT     DEFAULT 0,
    ADD COLUMN damage_falloff_rate     INT     DEFAULT 0,
    ADD COLUMN spread                  NUMERIC DEFAULT 0,
    ADD COLUMN rate_of_fire            NUMERIC DEFAULT 0,
    ADD COLUMN radius                  INT     DEFAULT 0,
    ADD COLUMN radial_does_full_damage BOOL    DEFAULT TRUE,
    ADD COLUMN projectile_speed        INT     DEFAULT 0,
    ADD COLUMN max_ammo                INT     DEFAULT 0,
    ADD COLUMN energy_cost             NUMERIC DEFAULT 0;

UPDATE blueprint_weapons
SET weapon_type           = 'Sniper Rifle',
    game_client_weapon_id = 'a155bef8-f0e1-4d11-8a23-a93b0bb74d10'
WHERE label = 'Sniper Rifle';

UPDATE blueprint_weapons
SET weapon_type = 'Sword'
WHERE label = 'Laser Sword';

UPDATE blueprint_weapons
SET weapon_type = 'Missile Launcher'
WHERE label = 'Rocket Pod';

UPDATE blueprint_weapons
SET weapon_type = 'Cannon'
WHERE label = 'Auto Cannon';

UPDATE blueprint_weapons
SET weapon_type = 'Plasma Gun'
WHERE label = 'Plasma Rifle';

UPDATE blueprint_weapons
SET weapon_type = 'Sword'
WHERE label = 'Sword';

ALTER TABLE blueprint_weapons
    ALTER COLUMN weapon_type SET NOT NULL;

ALTER TABLE weapons
    DROP COLUMN IF EXISTS weapon_type,
    ADD COLUMN default_damage_typ      DAMAGE_TYPE NOT NULL DEFAULT 'Kinetic',
    ADD COLUMN collection_slug         COLLECTION  NOT NULL DEFAULT 'supremacy-general',
    ADD COLUMN token_id                BIGINT,
    ADD COLUMN genesis_token_id        NUMERIC,
    ADD COLUMN weapon_type             WEAPON_TYPE,
    ADD COLUMN owner_id                UUID REFERENCES players (id),
    ADD COLUMN damage_falloff          INT     DEFAULT 0,
    ADD COLUMN damage_falloff_rate     INT     DEFAULT 0,
    ADD COLUMN spread                  INT     DEFAULT 0,
    ADD COLUMN rate_of_fire            NUMERIC DEFAULT 0,
    ADD COLUMN radius                  INT     DEFAULT 0,
    ADD COLUMN radial_does_full_damage BOOL    DEFAULT TRUE,
    ADD COLUMN projectile_speed        NUMERIC DEFAULT 0,
    ADD COLUMN energy_cost             NUMERIC DEFAULT 0,
    ADD COLUMN max_ammo                INT     DEFAULT 0,
    ADD FOREIGN KEY (collection_slug, token_id) REFERENCES collection_items (collection_slug, token_id);


UPDATE weapons
SET weapon_type = 'Sniper Rifle'
WHERE label = 'Sniper Rifle';

UPDATE weapons
SET weapon_type = 'Sword'
WHERE label = 'Laser Sword';

UPDATE weapons
SET weapon_type = 'Missile Launcher'
WHERE label = 'Rocket Pod';

UPDATE weapons
SET weapon_type = 'Cannon'
WHERE label = 'Auto Cannon';

UPDATE weapons
SET weapon_type = 'Plasma Gun'
WHERE label = 'Plasma Rifle';

UPDATE weapons
SET weapon_type = 'Sword'
WHERE label = 'Sword';

WITH weapon_owners AS (SELECT m.owner_id, cw.weapon_id
                       FROM chassis_weapons cw
                                INNER JOIN mechs m ON cw.chassis_id = m.chassis_id)
UPDATE weapons w
SET owner_id = weapon_owners.owner_id
FROM weapon_owners
WHERE w.id = weapon_owners.weapon_id;

-- This inserts a new collection_items entry for each weapons and updates the weapons table with token id
WITH insrt AS (
    WITH weapon AS (SELECT 'weapon' AS item_type, id FROM weapons)
        INSERT INTO collection_items (token_id, item_type, item_id)
            SELECT NEXTVAL('collection_general'), weapon.item_type, weapon.id
            FROM weapon
            RETURNING token_id, item_id)
UPDATE weapons w
SET token_id = insrt.token_id
FROM insrt
WHERE w.id = insrt.item_id;

-- this updates all genesis_token_id for weapons that are in genesis
WITH genesis AS (SELECT external_token_id, m.collection_slug, m.chassis_id, _cw.weapon_id
                 FROM chassis_weapons _cw
                          INNER JOIN mechs m ON m.chassis_id = _cw.chassis_id
                 WHERE m.collection_slug = 'supremacy-genesis')
UPDATE weapons w
SET genesis_token_id = genesis.external_token_id
FROM genesis
WHERE w.id = genesis.weapon_id;


ALTER TABLE weapons
    ALTER COLUMN token_id SET NOT NULL,
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
    default_damage_typ      = 'Energy'
WHERE label ILIKE 'Plasma Rifle';

UPDATE weapons
SET damage                  = 12,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 4,
    rate_of_fire            = 0,
    radius                  = 100,
    projectile_speed        = 36000,
    radial_does_full_damage = TRUE,
    energy_cost             = 10,
    default_damage_typ      = 'Kinetic'
WHERE label ILIKE 'Auto Cannon';

UPDATE weapons
SET damage                  = 130,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 3,
    rate_of_fire            = 0,
    radius                  = 100,
    projectile_speed        = 80000,
    radial_does_full_damage = TRUE,
    energy_cost             = 15,
    default_damage_typ      = 'Kinetic'
WHERE label ILIKE 'Sniper Rifle';

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
    default_damage_typ      = 'Explosive'
WHERE label ILIKE 'Rocket Pod';

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
    default_damage_typ      = 'Kinetic'
WHERE label ILIKE 'Sword';

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
    default_damage_typ      = 'Energy'
WHERE label ILIKE 'Laser Sword';

--  blueprint weapons
-- update weapon stats
UPDATE blueprint_weapons
SET damage                  = 20,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 3,
    rate_of_fire            = 0,
    radius                  = 100,
    projectile_speed        = 48000,
    radial_does_full_damage = TRUE,
    energy_cost             = 10,
    default_damage_typ      = 'Energy'
WHERE label ILIKE 'Plasma Rifle';

UPDATE blueprint_weapons
SET damage                  = 12,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 4,
    rate_of_fire            = 0,
    radius                  = 100,
    projectile_speed        = 36000,
    radial_does_full_damage = TRUE,
    energy_cost             = 10,
    default_damage_typ      = 'Kinetic'
WHERE label ILIKE 'Auto Cannon';

UPDATE blueprint_weapons
SET damage                  = 130,
    damage_falloff          = 0,
    damage_falloff_rate     = 0,
    spread                  = 3,
    rate_of_fire            = 0,
    radius                  = 100,
    projectile_speed        = 80000,
    radial_does_full_damage = TRUE,
    energy_cost             = 15,
    default_damage_typ      = 'Kinetic'
WHERE label ILIKE 'Sniper Rifle';

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
    default_damage_typ      = 'Explosive'
WHERE label ILIKE 'Rocket Pod';

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
    default_damage_typ      = 'Kinetic'
WHERE label ILIKE 'Sword';

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
    default_damage_typ      = 'Energy'
WHERE label ILIKE 'Laser Sword';
