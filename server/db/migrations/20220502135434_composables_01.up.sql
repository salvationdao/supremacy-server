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
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    collection_slug COLLECTION       NOT NULL DEFAULT 'supremacy-general',
    hash            TEXT             NOT NULL DEFAULT random_string(10),
    token_id        BIGINT           NOT NULL,
    item_type       TEXT             NOT NULL CHECK (item_type IN ('utility', 'weapon', 'chassis', 'chassis_skin')),
    item_id         UUID             NOT NULL UNIQUE,
    UNIQUE (collection_slug, token_id)
);


DROP TYPE IF EXISTS WEAPON_TYPE;
CREATE TYPE WEAPON_TYPE AS ENUM ('Grenade Launcher', 'Cannon', 'Minigun', 'Plasma Gun', 'Flak',
    'Machine Gun', 'Flamethrower', 'Missile Launcher', 'Laser Beam',
    'Lightning Gun', 'BFG', 'Rifle', 'Sniper Rifle', 'Sword');



DROP TYPE IF EXISTS DAMAGE_TYPE;
CREATE TYPE DAMAGE_TYPE AS ENUM ('Kinetic', 'Energy', 'Explosive');

DROP TYPE IF EXISTS CHASSIS_MODEL;

-- creating table of war machine chassis modals
CREATE TABLE chassis_model
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    label      TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
    id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    collection_item_id UUID        NOT NULL REFERENCES collection_items (id),
    owner_id           UUID        NOT NULL REFERENCES players (id),
    label              TEXT        NOT NULL,
    size               TEXT        NOT NULL DEFAULT 'MEDIUM' CHECK ( size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    capacity           NUMERIC     NOT NULL DEFAULT 0,
    max_draw_rate      NUMERIC     NOT NULL DEFAULT 0,
    recharge_rate      NUMERIC     NOT NULL DEFAULT 0,
    armour             NUMERIC     NOT NULL DEFAULT 0,
    max_hitpoints      NUMERIC     NOT NULL DEFAULT 0,
    tier               TEXT,
    equipped_on        UUID REFERENCES chassis (id),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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
    blueprint_id       UUID REFERENCES blueprint_chassis_skin (id),
    collection_item_id UUID REFERENCES collection_items (id),
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
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

/*
  CHASSIS ANIMATIONS
 */

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

CREATE TABLE chassis_animation
(
    id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_id       UUID        NOT NULL REFERENCES blueprint_chassis_animation (id),
    collection_item_id UUID        NOT NULL REFERENCES collection_items (id),
    label              TEXT        NOT NULL,
    owner_id           UUID        NOT NULL REFERENCES players (id),
    chassis_model      UUID        NOT NULL REFERENCES chassis_model (id),
    equipped_on        UUID REFERENCES chassis (id),
    tier               TEXT,
    intro_animation    BOOL                 DEFAULT TRUE,
    outro_animation    BOOL                 DEFAULT TRUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


/*
  CHASSIS
 */


ALTER TABLE chassis
--     unused/unneeded columns
    DROP COLUMN IF EXISTS turret_hardpoints,
    DROP COLUMN IF EXISTS health_remaining,
    ADD COLUMN blueprint_id       UUID REFERENCES blueprint_chassis (id),
    ADD COLUMN is_default         BOOL NOT NULL DEFAULT FALSE,
    ADD COLUMN is_insured         BOOL NOT NULL DEFAULT FALSE,
    ADD COLUMN name               TEXT NOT NULL DEFAULT '',
    ADD COLUMN model_id           UUID REFERENCES chassis_model (id),
    ADD COLUMN collection_item_id UUID REFERENCES collection_items (id),
    ADD COLUMN genesis_token_id   INTEGER,
    ADD COLUMN owner_id           UUID REFERENCES players (id),
    ADD COLUMN energy_core_size   TEXT NOT NULL DEFAULT 'MEDIUM' CHECK ( energy_core_size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    ADD COLUMN tier               TEXT,
    ADD COLUMN chassis_skin_id    UUID REFERENCES chassis_skin (id), -- equipped skin
    ADD COLUMN energy_core_id     UUID REFERENCES energy_cores (id),
    ADD COLUMN intro_animation_id UUID REFERENCES chassis_animation (id),
    ADD COLUMN outro_animation_id UUID REFERENCES chassis_animation (id);

UPDATE chassis c
SET model_id = (SELECT id
                FROM chassis_model cm
                WHERE c.model = cm.label);

ALTER TABLE chassis
    DROP COLUMN model,
    ALTER COLUMN model_id SET NOT NULL;

-- This inserts a new collection_items entry for each chassis and updates the chassis table with token id
WITH insrt AS (
    WITH chass AS (SELECT 'chassis' AS item_type, id FROM chassis)
        INSERT INTO collection_items (token_id, item_type, item_id)
            SELECT NEXTVAL('collection_general'), chass.item_type, chass.id
            FROM chass
            RETURNING id, item_id)
UPDATE chassis c
SET collection_item_id = insrt.id
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
            RETURNING id, item_id)
UPDATE chassis_skin cs
SET collection_item_id = insrt.id
FROM insrt
WHERE cs.id = insrt.item_id;


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

WITH mech_owners AS (SELECT owner_id, chassis_id, is_default, is_insured, name, tier, template_id
                     FROM mechs)
UPDATE chassis c
SET owner_id     = mech_owners.owner_id,
    is_insured   = mech_owners.is_insured,
    is_default   = mech_owners.is_default,
    name         = mech_owners.name,
    tier         = mech_owners.tier,
    blueprint_id = (SELECT blueprint_chassis_id FROM templates WHERE id = mech_owners.template_id)
FROM mech_owners
WHERE c.id = mech_owners.chassis_id;

ALTER TABLE chassis
    ALTER COLUMN collection_item_id SET NOT NULL,
    ALTER COLUMN blueprint_id SET NOT NULL,
    ALTER COLUMN owner_id SET NOT NULL;

ALTER TABLE blueprint_chassis
    DROP COLUMN IF EXISTS turret_hardpoints,
    DROP COLUMN IF EXISTS health_remaining,
    ADD COLUMN model_id         UUID REFERENCES chassis_model (id),
    ADD COLUMN energy_core_size TEXT DEFAULT 'MEDIUM' CHECK ( energy_core_size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    ADD COLUMN tier             TEXT,
    ADD COLUMN chassis_skin_id  UUID REFERENCES blueprint_chassis_skin (id); -- this column is used temp and gets removed.

UPDATE blueprint_chassis c
SET model_id = (SELECT id
                FROM chassis_model cm
                WHERE c.model = cm.label);

ALTER TABLE blueprint_chassis
    DROP COLUMN model,
    ALTER COLUMN model_id SET NOT NULL;

UPDATE chassis_model mm
SET default_chassis_skin_id = (SELECT bcs.id
                               FROM blueprint_chassis_skin bcs
                                        INNER JOIN chassis_model cm ON bcs.chassis_model = cm.id
                               WHERE cm.id = bcs.chassis_model
                                 AND bcs.label = 'Blue White')
WHERE mm.label = 'Law Enforcer X-1000';

UPDATE chassis_model mm
SET default_chassis_skin_id = (SELECT bcs.id
                               FROM blueprint_chassis_skin bcs
                                        INNER JOIN chassis_model cm ON bcs.chassis_model = cm.id
                               WHERE cm.id = bcs.chassis_model
                                 AND bcs.label = 'Beetle')
WHERE mm.label = 'Olympus Mons LY07';

UPDATE chassis_model mm
SET default_chassis_skin_id = (SELECT bcs.id
                               FROM blueprint_chassis_skin bcs
                                        INNER JOIN chassis_model cm ON bcs.chassis_model = cm.id
                               WHERE cm.id = bcs.chassis_model
                                 AND bcs.label = 'White Gold')
WHERE mm.label = 'Tenshi Mk1';

-- SET THE CONNECTED SKINS
UPDATE blueprint_chassis bc
SET chassis_skin_id = (SELECT id
                       FROM blueprint_chassis_skin bcs
                       WHERE bcs.label = bc.skin
                         AND bcs.chassis_model = bc.model_id);

-- fix ones we missed somehow
UPDATE chassis_skin
SET label = 'Gumdan'
WHERE label = 'Gundam';

UPDATE chassis_skin ms
SET blueprint_id = (SELECT id
                    FROM blueprint_chassis_skin bms
                    WHERE bms.label = ms.label
                      AND ms.chassis_model = bms.chassis_model);

-- here
ALTER TABLE chassis_skin
    ALTER COLUMN collection_item_id SET NOT NULL,
    ALTER COLUMN blueprint_id SET NOT NULL;

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
