--  creating a function to auto generate url safe hashes for urls and game client
-- https://stackoverflow.com/questions/3970795/how-do-you-create-a-random-string-thats-suitable-for-a-session-id-in-postgresql
CREATE OR REPLACE FUNCTION random_string(length INTEGER) RETURNS TEXT AS
$$
DECLARE
    chars  TEXT[]  := '{0,1,2,3,4,5,6,7,8,9,A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z,a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z}';
    result TEXT    := '';
    i      INTEGER := 0;
BEGIN
    IF length < 0 THEN
        RAISE EXCEPTION 'Given length cannot be less than 0';
    END IF;
    FOR i IN 1..length
        LOOP
            result := result || chars[CEIL(61 * RANDOM())];
        END LOOP;
    RETURN result;
END;
$$ LANGUAGE plpgsql;

CREATE SEQUENCE IF NOT EXISTS collection_general AS BIGINT;
ALTER SEQUENCE collection_general RESTART WITH 1;

DROP TYPE IF EXISTS COLLECTION;
CREATE TYPE COLLECTION AS ENUM ('supremacy-genesis', 'supremacy-general');

-- This table is for the look up token ids since the token ids go across tables
CREATE TABLE collection_items
(
    collection_slug COLLECTION NOT NULL DEFAULT 'supremacy-general',
    hash            TEXT       NOT NULL DEFAULT random_string(10),
    token_id        BIGINT     NOT NULL,
    item_type       TEXT       NOT NULL CHECK (item_type IN ('utility', 'weapon', 'chassis', 'chassis_skin')),
    item_id         UUID       NOT NULL UNIQUE,
    PRIMARY KEY (collection_slug, token_id)
);


DROP TYPE IF EXISTS WEAPON_TYPE;
CREATE TYPE WEAPON_TYPE AS ENUM ('Grenade Launcher', 'Cannon', 'Minigun', 'Plasma Gun', 'Flak',
    'Machine Gun', 'Flamethrower', 'Missile Launcher', 'Laser Beam',
    'Lightning Gun', 'BFG', 'Rifle', 'Sniper Rifle', 'Sword');

DROP TYPE IF EXISTS UTILITY_TYPE;
CREATE TYPE UTILITY_TYPE AS ENUM ('SHIELD', 'ATTACK DRONE', 'REPAIR DRONE', 'ANTI MISSILE',
    'ACCELERATOR');


DROP TYPE IF EXISTS DAMAGE_TYPE;
CREATE TYPE DAMAGE_TYPE AS ENUM ('Kinetic', 'Energy', 'Explosive');

DROP TYPE IF EXISTS chassis_model;

-- creating table of war machine chassis modals
CREATE TABLE chassis_model
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    label      TEXT        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

INSERT INTO chassis_model (label)
VALUES ('Law Enforcer X-1000'),
       ('Olympus Mons LY07'),
       ('Tenshi Mk1');

/*
  UPDATING DEFAULTS
  For some reason the ai/default mechs had different models, fixing that
 */

UPDATE chassis
SET model = 'Olympus Mons LY07',
    skin  = 'Beetle'
WHERE model = 'BXSD';
UPDATE chassis
SET model = 'Tenshi Mk1',
    skin  = 'Warden'
WHERE model = 'WREX';
UPDATE chassis
SET model = 'Law Enforcer X-1000',
    skin  = 'Blue White'
WHERE model = 'XFVS';

UPDATE blueprint_chassis
SET model = 'Olympus Mons LY07',
    skin  = 'Beetle'
WHERE model = 'BXSD';
UPDATE blueprint_chassis
SET model = 'Tenshi Mk1',
    skin  = 'Warden'
WHERE model = 'WREX';
UPDATE blueprint_chassis
SET model = 'Law Enforcer X-1000',
    skin  = 'Blue White'
WHERE model = 'XFVS';

/*
  ENERGY CORES
 */

CREATE TABLE blueprint_energy_cores
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    collection    COLLECTION  NOT NULL DEFAULT 'supremacy-general',
    label         TEXT        NOT NULL,
    size          TEXT        NOT NULL DEFAULT 'MEDIUM' CHECK ( size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    capacity      NUMERIC     NOT NULL DEFAULT 0,
    max_draw_rate NUMERIC     NOT NULL DEFAULT 0,
    recharge_rate NUMERIC     NOT NULL DEFAULT 0,
    armour        NUMERIC     NOT NULL DEFAULT 0,
    max_hitpoints NUMERIC     NOT NULL DEFAULT 0,
    tier          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE energy_cores
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    collection_slug COLLECTION  NOT NULL DEFAULT 'supremacy-general',
    token_id        BIGINT      NOT NULL,
    owner_id        UUID        NOT NULL REFERENCES players (id),
    label           TEXT        NOT NULL,
    size            TEXT        NOT NULL DEFAULT 'MEDIUM' CHECK ( size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    capacity        NUMERIC     NOT NULL DEFAULT 0,
    max_draw_rate   NUMERIC     NOT NULL DEFAULT 0,
    recharge_rate   NUMERIC     NOT NULL DEFAULT 0,
    armour          NUMERIC     NOT NULL DEFAULT 0,
    max_hitpoints   NUMERIC     NOT NULL DEFAULT 0,
    tier            TEXT,
    equipped_on     UUID REFERENCES chassis (id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (collection_slug, token_id) REFERENCES collection_items (collection_slug, token_id)
);
-- TODO: ADD COLLECTION/TOKEN ID CHECK

/*
  CHASSIS SKINS
 */

CREATE TABLE blueprint_chassis_skin
(
    id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    collection         COLLECTION  NOT NULL DEFAULT 'supremacy-general',
    chassis_model      UUID        NOT NULL REFERENCES chassis_model (id),
    label              TEXT        NOT NULL,
    tier               TEXT,
    image_url          TEXT,
    animation_url      TEXT,
    card_animation_url TEXT,
    large_image_url    TEXT,
    avatar_url         TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (chassis_model, label)
);

ALTER TABLE chassis_model
    ADD COLUMN default_chassis_skin_id UUID REFERENCES blueprint_chassis_skin (id); -- default skin


CREATE TABLE chassis_skin
(
    id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    collection_slug    COLLECTION  NOT NULL DEFAULT 'supremacy-general',
    token_id           BIGINT,
    genesis_token_id   NUMERIC,
    label              TEXT        NOT NULL,
    owner_id           UUID        NOT NULL REFERENCES players (id),
    chassis_model      UUID        NOT NULL REFERENCES chassis_model (id),
    equipped_on        UUID REFERENCES chassis (id),
    tier               TEXT,
    image_url          TEXT,
    animation_url      TEXT,
    card_animation_url TEXT,
    avatar_url         TEXT,
    large_image_url    TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (collection_slug, token_id) REFERENCES collection_items (collection_slug, token_id)
);

-- TODO: CREATE CHECK ON COLLECTION/TOKEN_ID

/*
  CHASSIS ANIMATIONS
 */

CREATE TABLE chassis_animation
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    collection_slug COLLECTION  NOT NULL DEFAULT 'supremacy-general',
    token_id        BIGINT      NOT NULL,
    label           TEXT        NOT NULL,
    owner_id        UUID        NOT NULL REFERENCES players (id),
    chassis_model   UUID        NOT NULL REFERENCES chassis_model (id),
    equipped_on     UUID REFERENCES chassis (id),
    tier            TEXT,
    intro_animation BOOL                 DEFAULT TRUE,
    outro_animation BOOL                 DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (collection_slug, token_id) REFERENCES collection_items (collection_slug, token_id)
);

-- TODO: CREATE CHECK ON COLLECTION/TOKEN ID

CREATE TABLE blueprint_chassis_animation
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    collection      COLLECTION  NOT NULL DEFAULT 'supremacy-general',
    label           TEXT        NOT NULL,
    chassis_model   UUID        NOT NULL REFERENCES chassis_model (id),
    equipped_on     UUID REFERENCES chassis (id),
    tier            TEXT,
    intro_animation BOOL                 DEFAULT TRUE,
    outro_animation BOOL                 DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

/*
  CHASSIS
 */


ALTER TABLE chassis
--     unused/unneeded columns
    DROP COLUMN IF EXISTS turret_hardpoints,
    DROP COLUMN IF EXISTS health_remaining,
    ADD COLUMN is_default              bool       NOT NULL DEFAULT FALSE,
    ADD COLUMN is_insured              bool       NOT NULL DEFAULT FALSE,
    ADD COLUMN model_id                UUID REFERENCES chassis_model (id),
    ADD COLUMN collection_slug         COLLECTION NOT NULL DEFAULT 'supremacy-general',
    ADD COLUMN token_id                BIGINT,
    ADD COLUMN genesis_token_id        INTEGER,
    ADD COLUMN owner_id                UUID REFERENCES players (id),
    ADD COLUMN energy_core_size        TEXT       NOT NULL DEFAULT 'MEDIUM' CHECK ( energy_core_size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    ADD COLUMN default_chassis_skin_id UUID REFERENCES blueprint_chassis_skin (id), -- default skin
    ADD COLUMN tier                    TEXT,
    ADD COLUMN chassis_skin_id         UUID REFERENCES chassis_skin (id), -- equipped skin
    ADD COLUMN energy_core_id          UUID REFERENCES energy_cores (id),
    ADD COLUMN intro_animation_id      UUID REFERENCES chassis_animation (id),
    ADD COLUMN outro_animation_id      UUID REFERENCES chassis_animation (id),
    ADD FOREIGN KEY (collection_slug, token_id) REFERENCES collection_items (collection_slug, token_id);

UPDATE chassis c
SET model_id = (SELECT id
                FROM chassis_model cm
                WHERE c.model = cm.label);

ALTER TABLE chassis
    DROP COLUMN model,
    ALTER COLUMN model_id SET NOT NULL;


-- TODO: ADD CHECK ON COLLECTION/TOKEN ID

-- This inserts a new collection_items entry for each chassis and updates the chassis table with token id
WITH insrt AS (
    WITH chass AS (SELECT 'chassis' AS item_type, id FROM chassis)
        INSERT INTO collection_items (token_id, item_type, item_id)
            SELECT NEXTVAL('collection_general'), chass.item_type, chass.id
            FROM chass
            RETURNING token_id, item_id)
UPDATE chassis c
SET token_id = insrt.token_id
FROM insrt
WHERE c.id = insrt.item_id;

-- this updates all genesis_token_id for chassis that are in genesis, also get is default and is insured
WITH genesis AS (SELECT external_token_id, collection_slug, chassis_id
                 FROM mechs
                 WHERE collection_slug = 'supremacy-genesis')
UPDATE chassis c
SET genesis_token_id = genesis.external_token_id
FROM genesis
WHERE c.id = genesis.chassis_id;


-- extract and insert current skin blueprints
WITH new_skins AS (SELECT DISTINCT c.skin,
                                   c.model,
                                   image_url,
                                   animation_url,
                                   card_animation_url,
                                   avatar_url,
                                   large_image_url,
                                   t.tier
                   FROM templates t
                            INNER JOIN blueprint_chassis c ON t.blueprint_chassis_id = c.id)
INSERT
INTO blueprint_chassis_skin(chassis_model, label, tier, image_url, animation_url, card_animation_url, large_image_url,
                            avatar_url)
SELECT (SELECT id FROM chassis_model WHERE label = new_skins.model),
       new_skins.skin,
       new_skins.tier,
       new_skins.image_url,
       new_skins.animation_url,
       new_skins.card_animation_url,
       new_skins.large_image_url,
       new_skins.avatar_url
FROM new_skins
ON CONFLICT (chassis_model, label) DO NOTHING;

-- extract and insert current equipped skins
WITH new_skins AS (SELECT DISTINCT c.skin,
                                   c.model_id,
                                   image_url,
                                   animation_url,
                                   card_animation_url,
                                   large_image_url,
                                   avatar_url,
                                   m.tier,
                                   m.owner_id,
                                   c.id AS chassis_id
                   FROM mechs m
                            INNER JOIN chassis c ON m.chassis_id = c.id)
INSERT
INTO chassis_skin(owner_id, equipped_on, chassis_model, label, tier, image_url, large_image_url, animation_url,
                  card_animation_url,
                  avatar_url)
SELECT new_skins.owner_id,
       new_skins.chassis_id,
       new_skins.model_id,
       new_skins.skin,
       new_skins.tier,
       new_skins.image_url,
       new_skins.large_image_url,
       new_skins.animation_url,
       new_skins.card_animation_url,
       new_skins.avatar_url
FROM new_skins;


-- This inserts a new collection_items entry for each chassis_skin and updates the chassis_skin table with token id
WITH insrt AS (
    WITH chass_skin AS (SELECT 'chassis_skin' AS item_type, id FROM chassis_skin)
        INSERT INTO collection_items (token_id, item_type, item_id)
            SELECT NEXTVAL('collection_general'), chass_skin.item_type, chass_skin.id
            FROM chass_skin
            RETURNING token_id, item_id)
UPDATE chassis_skin cs
SET token_id = insrt.token_id
FROM insrt
WHERE cs.id = insrt.item_id;

ALTER TABLE chassis_skin
    ALTER COLUMN token_id SET NOT NULL;

-- this updates all genesis_token_id for chassis_skin that are in genesis
WITH genesis AS (SELECT external_token_id, m.collection_slug, chassis_id
                 FROM chassis_skin _cs
                          INNER JOIN mechs m ON m.chassis_id = _cs.equipped_on
                 WHERE m.collection_slug = 'supremacy-genesis')
UPDATE chassis_skin cs
SET genesis_token_id = genesis.external_token_id
FROM genesis
WHERE cs.equipped_on = genesis.chassis_id;

-- update the owners of the newly extracted and inserted skins
UPDATE chassis c
SET chassis_skin_id = (SELECT id FROM chassis_skin cs WHERE cs.equipped_on = c.id);

-- update all the default model skins, picked random mega skins to be the default fallback..
UPDATE chassis c
SET default_chassis_skin_id = (SELECT bcs.id
                               FROM blueprint_chassis_skin bcs
                                        INNER JOIN chassis_model cm ON bcs.chassis_model = cm.id
                               WHERE cm.id = bcs.chassis_model
                                 AND bcs.label = 'Blue White')
WHERE c.model_id = (SELECT id FROM chassis_model WHERE label = 'Law Enforcer X-1000');

UPDATE chassis c
SET default_chassis_skin_id = (SELECT bcs.id
                               FROM blueprint_chassis_skin bcs
                                        INNER JOIN chassis_model cm ON bcs.chassis_model = cm.id
                               WHERE cm.id = bcs.chassis_model
                                 AND bcs.label = 'Beetle')
WHERE c.model_id = (SELECT id FROM chassis_model WHERE label = 'Olympus Mons LY07');

UPDATE chassis c
SET default_chassis_skin_id = (SELECT bcs.id
                               FROM blueprint_chassis_skin bcs
                                        INNER JOIN chassis_model cm ON bcs.chassis_model = cm.id
                               WHERE cm.id = bcs.chassis_model
                                 AND bcs.label = 'Warden')
WHERE c.model_id = (SELECT id FROM chassis_model WHERE label = 'Tenshi Mk1');

ALTER TABLE chassis
    ALTER COLUMN default_chassis_skin_id SET NOT NULL;

WITH mech_owners AS (SELECT owner_id, chassis_id, is_default, is_insured
                     FROM mechs)
UPDATE chassis c
SET owner_id   = mech_owners.owner_id,
    is_insured = mech_owners.is_insured,
    is_default = mech_owners.is_default
FROM mech_owners
WHERE c.id = mech_owners.chassis_id;

ALTER TABLE chassis
    ALTER COLUMN owner_id SET NOT NULL;

ALTER TABLE blueprint_chassis
    DROP COLUMN IF EXISTS turret_hardpoints,
    DROP COLUMN IF EXISTS health_remaining,
    ADD COLUMN model_id                UUID REFERENCES chassis_model (id),
    ADD COLUMN energy_core_size        TEXT DEFAULT 'MEDIUM' CHECK ( energy_core_size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    ADD COLUMN tier                    TEXT,
    ADD COLUMN default_chassis_skin_id UUID REFERENCES blueprint_chassis_skin (id),
    ADD COLUMN chassis_skin_id         UUID REFERENCES blueprint_chassis_skin (id),
    ADD COLUMN energy_core_id          UUID REFERENCES blueprint_energy_cores (id),
    ADD COLUMN intro_animation_id      UUID REFERENCES blueprint_chassis_animation (id),
    ADD COLUMN outro_animation_id      UUID REFERENCES blueprint_chassis_animation (id);

UPDATE blueprint_chassis c
SET model_id = (SELECT id
                FROM chassis_model cm
                WHERE c.model = cm.label);

ALTER TABLE blueprint_chassis
    DROP COLUMN model,
    ALTER COLUMN model_id SET NOT NULL;


UPDATE blueprint_chassis c
SET default_chassis_skin_id = (SELECT bcs.id
                               FROM blueprint_chassis_skin bcs
                                        INNER JOIN chassis_model cm ON bcs.chassis_model = cm.id
                               WHERE cm.id = bcs.chassis_model
                                 AND bcs.label = 'Blue White')
WHERE c.model_id = (SELECT id FROM chassis_model WHERE label = 'Law Enforcer X-1000');

UPDATE blueprint_chassis c
SET default_chassis_skin_id = (SELECT bcs.id
                               FROM blueprint_chassis_skin bcs
                                        INNER JOIN chassis_model cm ON bcs.chassis_model = cm.id
                               WHERE cm.id = bcs.chassis_model
                                 AND bcs.label = 'Beetle')
WHERE c.model_id = (SELECT id FROM chassis_model WHERE label = 'Olympus Mons LY07');

UPDATE blueprint_chassis c
SET default_chassis_skin_id = (SELECT bcs.id
                               FROM blueprint_chassis_skin bcs
                                        INNER JOIN chassis_model cm ON bcs.chassis_model = cm.id
                               WHERE cm.id = bcs.chassis_model
                                 AND bcs.label = 'Warden')
WHERE c.model_id = (SELECT id FROM chassis_model WHERE label = 'Tenshi Mk1');

ALTER TABLE blueprint_chassis
    ALTER COLUMN default_chassis_skin_id SET NOT NULL;

-- SET THE CONNECTED SKINS
UPDATE blueprint_chassis bc
SET chassis_skin_id = (SELECT id
                       FROM blueprint_chassis_skin bcs
                       WHERE bcs.label = bc.skin
                         AND bcs.chassis_model = bc.model_id);


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


/*
  AMMO
 */

CREATE TABLE blueprint_ammo
(
    id                             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    label                          TEXT        NOT NULL,
    weapon_type                    WEAPON_TYPE NOT NULL,
    damage_multiplier              NUMERIC              DEFAULT 0,
    damage_falloff_multiplier      NUMERIC              DEFAULT 0,
    damage_falloff_rate_multiplier NUMERIC              DEFAULT 0,
    spread_multiplier              NUMERIC              DEFAULT 0,
    rate_of_fire_multiplier        NUMERIC              DEFAULT 0,
    radius_multiplier              NUMERIC              DEFAULT 0,
    projectile_speed_multiplier    NUMERIC              DEFAULT 0,
    energy_cost_multiplier         NUMERIC              DEFAULT 0,
    max_ammo_multiplier            NUMERIC              DEFAULT 0,
    created_at                     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ammo
(
    blueprint_id UUID        NOT NULL REFERENCES blueprint_ammo (id),
    owner_id     UUID        NOT NULL REFERENCES players (id),
    count        INT         NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blueprint_id, owner_id)
);

CREATE TABLE weapon_ammo
(
    blueprint_ammo_id UUID        NOT NULL REFERENCES blueprint_ammo (id),
    weapon_id         UUID        NOT NULL REFERENCES weapons (id),
    count             INT         NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blueprint_ammo_id, weapon_id)
);

/*
  UTILITY
 */

ALTER TABLE blueprint_chassis_blueprint_modules
    RENAME TO blueprint_chassis_blueprint_utility;
ALTER TABLE blueprint_chassis_blueprint_utility
    RENAME COLUMN blueprint_module_id TO blueprint_utility_id;
ALTER TABLE blueprint_modules
    RENAME TO blueprint_utility;
ALTER TABLE blueprint_utility
    DROP COLUMN hitpoint_modifier,
    DROP COLUMN shield_modifier,
    ADD COLUMN type UTILITY_TYPE;

UPDATE blueprint_utility
SET type = 'SHIELD';
ALTER TABLE blueprint_utility
    ALTER COLUMN type SET NOT NULL;

CREATE TABLE blueprint_utility_shield
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_utility_id UUID        NOT NULL REFERENCES blueprint_utility (id),
    hitpoints            INT         NOT NULL DEFAULT 0,
    recharge_rate        INT         NOT NULL DEFAULT 0,
    recharge_energy_cost INT         NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE blueprint_utility_attack_drone
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_utility_id UUID        NOT NULL REFERENCES blueprint_utility (id),
    damage               INT         NOT NULL,
    rate_of_fire         INT         NOT NULL,
    hitpoints            INT         NOT NULL,
    lifespan_seconds     INT         NOT NULL,
    deploy_energy_cost   INT         NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE blueprint_utility_repair_drone
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_utility_id UUID        NOT NULL REFERENCES blueprint_utility (id),
    repair_type          TEXT CHECK (repair_type IN ('SHIELD', 'STRUCTURE')),
    repair_amount        INT         NOT NULL,
    deploy_energy_cost   INT         NOT NULL,
    lifespan_seconds     INT         NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE blueprint_utility_anti_missile
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_utility_id UUID        NOT NULL REFERENCES blueprint_utility (id),
    rate_of_fire         INT         NOT NULL,
    fire_energy_cost     INT         NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE blueprint_utility_accelerator
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_utility_id UUID        NOT NULL REFERENCES blueprint_utility (id),
    energy_cost          INT         NOT NULL,
    boost_seconds        INT         NOT NULL,
    boost_amount         INT         NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


ALTER TABLE chassis_modules
    RENAME TO chassis_utility;
ALTER TABLE chassis_utility
    RENAME COLUMN module_id TO utility_id;

ALTER TABLE modules
    RENAME TO utility;
ALTER TABLE utility
    DROP COLUMN hitpoint_modifier,
    DROP COLUMN shield_modifier,
    ADD COLUMN collection_slug  COLLECTION NOT NULL DEFAULT 'supremacy-general',
    ADD COLUMN token_id         BIGINT,
    ADD COLUMN genesis_token_id NUMERIC,
    ADD COLUMN owner_id         UUID REFERENCES players (id),
    ADD COLUMN equipped_on      UUID REFERENCES chassis (id),
    ADD COLUMN type             UTILITY_TYPE,
    ADD FOREIGN KEY (collection_slug, token_id) REFERENCES collection_items (collection_slug, token_id);

WITH utility_owners AS (SELECT m.owner_id, cu.utility_id
                        FROM chassis_utility cu
                                 INNER JOIN mechs m ON cu.chassis_id = m.chassis_id)
UPDATE utility u
SET owner_id = utility_owners.owner_id
FROM utility_owners
WHERE u.id = utility_owners.utility_id;


-- This inserts a new collection_items entry for each utility and updates the utility table with token id
WITH insrt AS (
    WITH utily AS (SELECT 'utility' AS item_type, id FROM utility)
        INSERT INTO collection_items (token_id, item_type, item_id)
            SELECT NEXTVAL('collection_general'), utily.item_type, utily.id
            FROM utily
            RETURNING token_id, item_id)
UPDATE utility u
SET token_id = insrt.token_id
FROM insrt
WHERE u.id = insrt.item_id;

-- this updates all genesis_token_id for weapons that are in genesis
WITH genesis AS (SELECT external_token_id, m.collection_slug, m.chassis_id, _cu.utility_id
                 FROM chassis_utility _cu
                          INNER JOIN mechs m ON m.chassis_id = _cu.chassis_id
                 WHERE m.collection_slug = 'supremacy-genesis')
UPDATE utility u
SET genesis_token_id = genesis.external_token_id
FROM genesis
WHERE u.id = genesis.utility_id;


ALTER TABLE utility
    DROP COLUMN slug,
    ALTER COLUMN token_id SET NOT NULL,
    ALTER COLUMN owner_id SET NOT NULL;

UPDATE utility
SET type = 'SHIELD';
ALTER TABLE blueprint_utility
    ALTER COLUMN type SET NOT NULL;


CREATE TABLE utility_shield
(
    utility_id           UUID PRIMARY KEY REFERENCES utility (id),
    hitpoints            INT NOT NULL DEFAULT 0,
    recharge_rate        INT NOT NULL DEFAULT 0,
    recharge_energy_cost INT NOT NULL DEFAULT 0
);

CREATE TABLE utility_attack_drone
(
    utility_id         UUID PRIMARY KEY REFERENCES utility (id),
    damage             INT NOT NULL,
    rate_of_fire       INT NOT NULL,
    hitpoints          INT NOT NULL,
    lifespan_seconds   INT NOT NULL,
    deploy_energy_cost INT NOT NULL
);

CREATE TABLE utility_repair_drone
(
    utility_id         UUID PRIMARY KEY REFERENCES utility (id),
    repair_type        TEXT CHECK (repair_type IN ('SHIELD', 'STRUCTURE')),
    repair_amount      INT NOT NULL,
    deploy_energy_cost INT NOT NULL,
    lifespan_seconds   INT NOT NULL
);

CREATE TABLE utility_anti_missile
(
    utility_id       UUID PRIMARY KEY REFERENCES utility (id),
    rate_of_fire     INT NOT NULL,
    fire_energy_cost INT NOT NULL
);

CREATE TABLE utility_accelerator
(
    utility_id    UUID PRIMARY KEY REFERENCES utility (id),
    energy_cost   INT NOT NULL,
    boost_seconds INT NOT NULL,
    boost_amount  INT NOT NULL
);


-- for each utility, create the shield utility
WITH umj AS (SELECT _cu.utility_id AS uid, _c.max_shield AS max_shield, _c.shield_recharge_rate AS shield_recharge_rate
             FROM chassis_utility _cu
                      INNER JOIN chassis _c ON _c.id = _cu.chassis_id
                      INNER JOIN mechs _m ON _m.chassis_id = _c.id)
INSERT
INTO utility_shield (utility_id, hitpoints, recharge_rate, recharge_energy_cost)
SELECT umj.uid, umj.max_shield, umj.shield_recharge_rate, 10
FROM umj;

ALTER TABLE chassis
    DROP COLUMN IF EXISTS skin,
    DROP COLUMN IF EXISTS slug,
    DROP COLUMN IF EXISTS shield_recharge_rate,
    DROP COLUMN IF EXISTS max_shield;


-- for each blueprint utility, create the blueprint shield utility
WITH umj AS (SELECT _cu.blueprint_utility_id AS uid,
                    _c.max_shield            AS max_shield,
                    _c.shield_recharge_rate  AS shield_recharge_rate
             FROM blueprint_chassis_blueprint_utility _cu
                      INNER JOIN blueprint_chassis _c ON _c.id = _cu.blueprint_chassis_id
                      INNER JOIN templates _m ON _m.blueprint_chassis_id = _c.id)
INSERT
INTO blueprint_utility_shield (blueprint_utility_id, hitpoints, recharge_rate, recharge_energy_cost)
SELECT umj.uid, umj.max_shield, umj.shield_recharge_rate, 10
FROM umj;

ALTER TABLE blueprint_chassis
    DROP COLUMN IF EXISTS shield_recharge_rate,
    DROP COLUMN IF EXISTS max_shield;

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


--  Here we are trying to remove the unneeded mechs table,
--  basically the mechs table turned out be redundant with the chassis table basically serving the same purpose.
--  1. We are going to swap all the FKs over to use chassis id
--  2. Rename chassis table mechs.
--  3. Yes I know I should have just updated the mechs table to begin with.

-- CREATE TABLE battle_queue (
-- mech_id UUID NOT NULL references mechs (id) PRIMARY KEY,

ALTER TABLE battle_queue
    ADD COLUMN chassis_id UUID REFERENCES chassis (id);

UPDATE battle_queue bq
SET chassis_id = (SELECT c.id
                  FROM mechs m
                           INNER JOIN chassis c ON m.chassis_id = c.id
                  WHERE m.id = bq.mech_id);

ALTER TABLE battle_queue
    DROP CONSTRAINT battle_queue_pkey,
    ADD PRIMARY KEY (chassis_id);

-- deal with table battle_queue_notifications that uses battle_queue.mech_id FK
-- CREATE TABLE battle_queue_notifications (
--     id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
--     battle_id UUID REFERENCES battles(id),
--     queue_mech_id UUID REFERENCES battle_queue(mech_id),
--     mech_id UUID NOT NULL REFERENCES mechs(id),

ALTER TABLE battle_queue_notifications
    ADD COLUMN chassis_id       UUID REFERENCES chassis (id),
    ADD COLUMN queue_chassis_id UUID REFERENCES battle_queue (chassis_id);

UPDATE battle_queue_notifications bqn
SET queue_chassis_id = (SELECT c.id
                        FROM mechs m
                                 INNER JOIN chassis c ON m.chassis_id = c.id
                        WHERE m.id = bqn.queue_mech_id),
    chassis_id       = (SELECT c.id
                        FROM mechs m
                                 INNER JOIN chassis c ON m.chassis_id = c.id
                        WHERE m.id = bqn.mech_id);

ALTER TABLE battle_queue_notifications
    DROP COLUMN queue_mech_id,
    DROP COLUMN mech_id;
ALTER TABLE battle_queue_notifications -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE battle_queue_notifications -- unsure why it wanted me to do a new alter table
    RENAME COLUMN queue_chassis_id TO queue_mech_id;


-- battle_queue_notifications_queue_mech_id_fkey
ALTER TABLE battle_queue
    DROP COLUMN mech_id;
ALTER TABLE battle_queue -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;



-- CREATE TABLE battle_mechs (
--     battle_id UUID NOT NULL references battles(id),
--     mech_id UUID NOT NULL references mechs(id),
--     owner_id UUID NOT NULL references players(id),
--     faction_id UUID NOT NULL references factions(id),
--     killed TIMESTAMPTZ NULL,
--     killed_by_id UUID NULL references mechs(id),

ALTER TABLE battle_mechs
    ADD COLUMN chassis_id           UUID REFERENCES chassis (id),
    ADD COLUMN killed_by_chassis_id UUID REFERENCES chassis (id);

UPDATE battle_mechs bm
SET chassis_id           = (SELECT c.id
                            FROM mechs m
                                     INNER JOIN chassis c ON m.chassis_id = c.id
                            WHERE m.id = bm.mech_id),
    killed_by_chassis_id = (SELECT c.id
                            FROM mechs m
                                     INNER JOIN chassis c ON m.chassis_id = c.id
                            WHERE m.id = bm.killed_by_id);

ALTER TABLE battle_mechs
    DROP CONSTRAINT battle_mechs_pkey,
    DROP COLUMN mech_id,
    DROP COLUMN killed_by_id;
ALTER TABLE battle_mechs -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE battle_mechs -- unsure why it wanted me to do a new alter table
    RENAME COLUMN killed_by_chassis_id TO killed_by_id;
ALTER TABLE battle_mechs -- unsure why it wanted me to do a new alter table
    ADD PRIMARY KEY (battle_id, mech_id);



-- DROP TABLE mechs;

-- CREATE TABLE battle_wins (
--     battle_id UUID NOT NULL references battles(id),
--     mech_id UUID NOT NULL references mechs(id),


ALTER TABLE battle_wins
    ADD COLUMN chassis_id UUID REFERENCES chassis (id);

UPDATE battle_wins bw
SET chassis_id = (SELECT c.id
                  FROM mechs m
                           INNER JOIN chassis c ON m.chassis_id = c.id
                  WHERE m.id = bw.mech_id);

ALTER TABLE battle_wins
    DROP CONSTRAINT battle_wins_pkey,
    DROP COLUMN mech_id;
ALTER TABLE battle_wins -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE battle_wins -- unsure why it wanted me to do a new alter table
    ADD PRIMARY KEY (battle_id, mech_id);


-- CREATE TABLE battle_kills (
--     battle_id UUID NOT NULL references battles(id),
--     mech_id UUID NOT NULL references mechs(id),
--     killed_id UUID NOT NULL references mechs(id),

ALTER TABLE battle_kills
    ADD COLUMN chassis_id        UUID REFERENCES chassis (id),
    ADD COLUMN killed_chassis_id UUID REFERENCES chassis (id);

UPDATE battle_kills bm
SET chassis_id        = (SELECT c.id
                         FROM mechs m
                                  INNER JOIN chassis c ON m.chassis_id = c.id
                         WHERE m.id = bm.mech_id),
    killed_chassis_id = (SELECT c.id
                         FROM mechs m
                                  INNER JOIN chassis c ON m.chassis_id = c.id
                         WHERE m.id = bm.killed_id);

ALTER TABLE battle_kills
    DROP CONSTRAINT battle_kills_pkey,
    DROP COLUMN mech_id,
    DROP COLUMN killed_id;
ALTER TABLE battle_kills -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE battle_kills -- unsure why it wanted me to do a new alter table
    RENAME COLUMN killed_chassis_id TO killed_id;
ALTER TABLE battle_kills -- unsure why it wanted me to do a new alter table
    ADD PRIMARY KEY (battle_id, mech_id);

-- CREATE TABLE battle_history (
--     id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
--     battle_id UUID NOT NULL references battles(id),
--     related_id UUID NULL references battle_history(id),
--     war_machine_one_id UUID NOT NULL references mechs(id),
--     war_machine_two_id UUID NULL references mechs(id),

ALTER TABLE battle_history
    ADD COLUMN war_machine_one_id_chassis UUID REFERENCES chassis (id),
    ADD COLUMN war_machine_two_id_chassis UUID REFERENCES chassis (id);

UPDATE battle_history bk
SET war_machine_one_id_chassis = (SELECT c.id
                                  FROM mechs m
                                           INNER JOIN chassis c ON m.chassis_id = c.id
                                  WHERE m.id = bk.war_machine_one_id),
    war_machine_two_id_chassis = (SELECT c.id
                                  FROM mechs m
                                           INNER JOIN chassis c ON m.chassis_id = c.id
                                  WHERE m.id = bk.war_machine_two_id);

ALTER TABLE battle_history
    DROP CONSTRAINT battle_history_pkey,
    DROP COLUMN war_machine_one_id,
    DROP COLUMN war_machine_two_id;
ALTER TABLE battle_history -- unsure why it wanted me to do a new alter table
    RENAME COLUMN war_machine_one_id_chassis TO war_machine_one_id;
ALTER TABLE battle_history -- unsure why it wanted me to do a new alter table
    RENAME COLUMN war_machine_two_id_chassis TO war_machine_two_id;



-- CREATE TABLE mech_stats (
--     mech_id UUID PRIMARY KEY NOT NULL REFERENCES mechs (id),


ALTER TABLE mech_stats
    ADD COLUMN chassis_id UUID REFERENCES chassis (id);

UPDATE mech_stats ms
SET chassis_id = (SELECT c.id
                  FROM mechs m
                           INNER JOIN chassis c ON m.chassis_id = c.id
                  WHERE m.id = ms.mech_id);

ALTER TABLE mech_stats
    DROP CONSTRAINT mech_stats_pkey,
    DROP COLUMN mech_id;
ALTER TABLE mech_stats -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE mech_stats -- unsure why it wanted me to do a new alter table
    ADD PRIMARY KEY (mech_id);

-- CREATE TABLE asset_repair(
--     id uuid primary key DEFAULT gen_random_uuid(),
--     mech_id UUID NOT NULL REFERENCES mechs (id),

ALTER TABLE asset_repair
    ADD COLUMN chassis_id UUID REFERENCES chassis (id);

UPDATE asset_repair ms
SET chassis_id = (SELECT c.id
                  FROM mechs m
                           INNER JOIN chassis c ON m.chassis_id = c.id
                  WHERE m.id = ms.mech_id);

ALTER TABLE asset_repair
    DROP COLUMN mech_id;
ALTER TABLE asset_repair -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;

-- CREATE TABLE battle_contracts (
--     id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
--     mech_id UUID NOT NULL references mechs (id),

ALTER TABLE battle_contracts
    ADD COLUMN chassis_id UUID REFERENCES chassis (id);

UPDATE battle_contracts bc
SET chassis_id = (SELECT c.id
                  FROM mechs m
                           INNER JOIN chassis c ON m.chassis_id = c.id
                  WHERE m.id = bc.mech_id);

ALTER TABLE battle_contracts
    DROP CONSTRAINT bc_unique_mech_battle,
    DROP COLUMN mech_id;
ALTER TABLE battle_contracts -- unsure why it wanted me to do a new alter table
    RENAME COLUMN chassis_id TO mech_id;
ALTER TABLE battle_contracts
    ADD UNIQUE (mech_id, battle_id);



DROP TABLE mechs;

ALTER TABLE chassis
    RENAME TO mechs;

ALTER TABLE chassis_model
    RENAME TO mech_model;

ALTER TABLE chassis_animation
    RENAME TO mech_animation;

ALTER TABLE chassis_skin
    RENAME TO mech_skin;

ALTER TABLE chassis_utility
    RENAME TO mech_utility;

ALTER TABLE blueprint_chassis
    RENAME TO blueprint_mechs;

ALTER TABLE blueprint_chassis_blueprint_utility
    RENAME TO blueprint_mech_blueprint_utility;

ALTER TABLE blueprint_chassis_blueprint_weapons
    RENAME TO blueprint_mech_blueprint_weapons;

ALTER TABLE blueprint_chassis_skin
    RENAME TO blueprint_mech_skin;
