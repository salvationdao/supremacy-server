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

CREATE SEQUENCE IF NOT EXISTS collection_genesis AS BIGINT;
ALTER SEQUENCE collection_genesis RESTART WITH 1;

CREATE SEQUENCE IF NOT EXISTS collection_limited_release AS BIGINT;
ALTER SEQUENCE collection_limited_release RESTART WITH 1;

CREATE SEQUENCE IF NOT EXISTS collection_consumables AS BIGINT;
ALTER SEQUENCE collection_consumables RESTART WITH 1;

DROP TYPE IF EXISTS ITEM_TYPE;
CREATE TYPE ITEM_TYPE AS ENUM ('utility', 'weapon', 'mech', 'mech_skin', 'mech_animation', 'power_core');

-- This table is for the look up token ids since the token ids go across tables
CREATE TABLE collection_items
(
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    collection_slug COLLECTION       NOT NULL DEFAULT 'supremacy-general',
    hash            TEXT             NOT NULL DEFAULT random_string(10),
    token_id        BIGINT           NOT NULL,
    item_type       ITEM_TYPE        NOT NULL,
    item_id         UUID             NOT NULL UNIQUE,
    tier            TEXT             NOT NULL DEFAULT 'MEGA',
    owner_id        UUID             NOT NULL REFERENCES players (id),
    market_locked   BOOL             NOT NULL DEFAULT FALSE,
    xsyn_locked     BOOL             NOT NULL DEFAULT FALSE,
    UNIQUE (collection_slug, token_id)
);

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

CREATE TABLE power_cores
(
    id                       UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_id             UUID REFERENCES blueprint_power_cores (id),
    label                    TEXT        NOT NULL,
    size                     TEXT        NOT NULL DEFAULT 'MEDIUM' CHECK ( size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    capacity                 NUMERIC     NOT NULL DEFAULT 0,
    genesis_token_id         BIGINT,
    limited_release_token_id BIGINT,
    max_draw_rate            NUMERIC     NOT NULL DEFAULT 0,
    recharge_rate            NUMERIC     NOT NULL DEFAULT 0,
    armour                   NUMERIC     NOT NULL DEFAULT 0,
    max_hitpoints            NUMERIC     NOT NULL DEFAULT 0,
    equipped_on              UUID REFERENCES chassis (id),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

/*
  CHASSIS SKINS
 */

-- CREATE TABLE blueprint_chassis_skin
-- (
--     id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
--     collection         COLLECTION  NOT NULL DEFAULT 'supremacy-general',
--     mech_model         UUID        NOT NULL REFERENCES blueprint_mechs (id),
--     label              TEXT        NOT NULL,
--     tier               TEXT        NOT NULL DEFAULT 'MEGA',
--     image_url          TEXT,
--     animation_url      TEXT,
--     card_animation_url TEXT,
--     large_image_url    TEXT,
--     avatar_url         TEXT,
--     created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     UNIQUE (mech_model, label)
-- );

CREATE TABLE chassis_skin
(
    id                       UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_id             UUID REFERENCES blueprint_mech_skin (id),
    genesis_token_id         BIGINT,
    limited_release_token_id BIGINT,
    label                    TEXT        NOT NULL,
    mech_model               UUID        NOT NULL REFERENCES blueprint_mechs (id),
    equipped_on              UUID REFERENCES chassis (id),
    image_url                TEXT,
    animation_url            TEXT,
    card_animation_url       TEXT,
    avatar_url               TEXT,
    large_image_url          TEXT,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

/*
  CHASSIS ANIMATIONS
 */

CREATE TABLE blueprint_chassis_animation
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    collection      COLLECTION  NOT NULL DEFAULT 'supremacy-general',
    label           TEXT        NOT NULL,
    mech_model      UUID        NOT NULL REFERENCES blueprint_mechs (id),
    tier            TEXT        NOT NULL DEFAULT 'MEGA',
    intro_animation BOOL                 DEFAULT TRUE,
    outro_animation BOOL                 DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chassis_animation
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_id    UUID        NOT NULL REFERENCES blueprint_chassis_animation (id),
    label           TEXT        NOT NULL,
    mech_model      UUID        NOT NULL REFERENCES blueprint_mechs (id),
    equipped_on     UUID REFERENCES chassis (id),
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
    ADD COLUMN blueprint_id             UUID REFERENCES blueprint_mechs (id),
    ADD COLUMN is_default               BOOL NOT NULL DEFAULT FALSE,
    ADD COLUMN is_insured               BOOL NOT NULL DEFAULT FALSE,
    ADD COLUMN name                     TEXT NOT NULL DEFAULT '',
    ADD COLUMN genesis_token_id         BIGINT,
    ADD COLUMN limited_release_token_id BIGINT,
    ADD COLUMN owner_id                 UUID REFERENCES players (id),
    ADD COLUMN power_core_size          TEXT NOT NULL DEFAULT 'SMALL' CHECK ( power_core_size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    ADD COLUMN tier                     TEXT NOT NULL DEFAULT 'MEGA',
    ADD COLUMN chassis_skin_id          UUID REFERENCES chassis_skin (id), -- equipped skin
    ADD COLUMN power_core_id            UUID REFERENCES power_cores (id),
    ADD COLUMN intro_animation_id       UUID REFERENCES chassis_animation (id),
    ADD COLUMN outro_animation_id       UUID REFERENCES chassis_animation (id);

UPDATE chassis c
SET blueprint_id = (SELECT id
                    FROM blueprint_mechs cm
                    WHERE c.model = cm.label);

ALTER TABLE chassis
    DROP COLUMN model,
    ALTER COLUMN blueprint_id SET NOT NULL;

-- This inserts a new collection_items entry for each chassis and updates the chassis table with token id

WITH chass AS (SELECT 'mech' AS item_type, c.id, m.tier, m.owner_id
               FROM chassis c
                        INNER JOIN mechs m ON c.id = m.chassis_id)
INSERT
INTO collection_items (token_id, item_type, item_id, tier, owner_id)
SELECT NEXTVAL('collection_general'), chass.item_type::ITEM_TYPE, chass.id, chass.tier, chass.owner_id
FROM chass;


-- this updates all genesis_token_id for chassis that are in genesis, also get is default and is insured
WITH genesis AS (SELECT external_token_id, collection_slug, chassis_id
                 FROM mechs
                 WHERE collection_slug = 'supremacy-genesis')
UPDATE chassis c
SET genesis_token_id = genesis.external_token_id
FROM genesis
WHERE c.id = genesis.chassis_id;

-- this updates all limited release for chassis that are in genesis, also get is default and is insured
WITH limited_release AS (SELECT external_token_id, collection_slug, chassis_id
                         FROM mechs
                         WHERE collection_slug = 'supremacy-limited-release')
UPDATE chassis c
SET limited_release_token_id = limited_release.external_token_id
FROM limited_release
WHERE c.id = limited_release.chassis_id;


-- extract and insert current skin blueprints
-- WITH new_skins AS (SELECT DISTINCT c.skin,
--                                    c.model,
--                                    image_url,
--                                    animation_url,
--                                    card_animation_url,
--                                    avatar_url,
--                                    large_image_url,
--                                    t.tier
--                    FROM templates t
--                             INNER JOIN blueprint_chassis c ON t.blueprint_chassis_id = c.id)
-- INSERT
-- INTO blueprint_mech_skin(mech_model, label, tier, image_url, animation_url, card_animation_url, large_image_url,
--                             avatar_url)
-- SELECT (SELECT id FROM blueprint_mechs WHERE label = new_skins.model),
--        new_skins.skin,
--        new_skins.tier,
--        new_skins.image_url,
--        new_skins.animation_url,
--        new_skins.card_animation_url,
--        new_skins.large_image_url,
--        new_skins.avatar_url
-- FROM new_skins
-- ON CONFLICT (mech_model, label) DO NOTHING;

-- extract and insert current equipped skins
WITH new_skins AS (SELECT DISTINCT c.skin,
                                   c.blueprint_id,
                                   image_url,
                                   animation_url,
                                   card_animation_url,
                                   large_image_url,
                                   avatar_url,
                                   c.id AS chassis_id
                   FROM mechs m
                            INNER JOIN chassis c ON m.chassis_id = c.id)
INSERT
INTO chassis_skin(equipped_on, mech_model, label, image_url, large_image_url, animation_url,
                  card_animation_url,
                  avatar_url)
SELECT new_skins.chassis_id,
       new_skins.blueprint_id,
       new_skins.skin,
       new_skins.image_url,
       new_skins.large_image_url,
       new_skins.animation_url,
       new_skins.card_animation_url,
       new_skins.avatar_url
FROM new_skins;


-- This inserts a new collection_items entry for each chassis_skin and updates the chassis_skin table with token id

WITH chass_skin AS (SELECT 'mech_skin' AS item_type, cs.id, m.tier, m.owner_id
                    FROM chassis_skin cs
                             INNER JOIN mechs m ON cs.equipped_on = m.chassis_id)
INSERT
INTO collection_items (token_id, item_type, item_id, tier, owner_id)
SELECT NEXTVAL('collection_general'),
       chass_skin.item_type::ITEM_TYPE,
       chass_skin.id,
       chass_skin.tier,
       chass_skin.owner_id
FROM chass_skin;


-- this updates all genesis_token_id for chassis_skin that are in genesis
WITH genesis AS (SELECT external_token_id, m.collection_slug, chassis_id
                 FROM chassis_skin _cs
                          INNER JOIN mechs m ON m.chassis_id = _cs.equipped_on
                 WHERE m.collection_slug = 'supremacy-genesis')
UPDATE chassis_skin cs
SET genesis_token_id = genesis.external_token_id
FROM genesis
WHERE cs.equipped_on = genesis.chassis_id;

-- this updates all limited release for chassis_skin that are in genesis
WITH limit_release AS (SELECT external_token_id, m.collection_slug, chassis_id
                       FROM chassis_skin _cs
                                INNER JOIN mechs m ON m.chassis_id = _cs.equipped_on
                       WHERE m.collection_slug = 'supremacy-limited-release')
UPDATE chassis_skin cs
SET limited_release_token_id = limit_release.external_token_id
FROM limit_release
WHERE cs.equipped_on = limit_release.chassis_id;

-- update the owners of the newly extracted and inserted skins
UPDATE chassis c
SET chassis_skin_id = (SELECT id FROM chassis_skin cs WHERE cs.equipped_on = c.id);

WITH mech_owners AS (SELECT owner_id, chassis_id, is_default, is_insured, name, tier FROM mechs)
UPDATE chassis c
SET owner_id     = mech_owners.owner_id,
    is_insured   = mech_owners.is_insured,
    is_default   = mech_owners.is_default,
    name         = mech_owners.name,
    tier         = mech_owners.tier
FROM mech_owners
WHERE c.id = mech_owners.chassis_id;

ALTER TABLE chassis
    ALTER COLUMN blueprint_id SET NOT NULL,
    ALTER COLUMN owner_id SET NOT NULL;

ALTER TABLE blueprint_chassis
    DROP COLUMN IF EXISTS turret_hardpoints,
    DROP COLUMN IF EXISTS health_remaining,
    ADD COLUMN model_id        UUID REFERENCES blueprint_mechs (id),
    ADD COLUMN collection      COLLECTION NOT NULL DEFAULT 'supremacy-general',
    ADD COLUMN power_core_size TEXT       NOT NULL DEFAULT 'SMALL' CHECK ( power_core_size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    ADD COLUMN tier            TEXT       NOT NULL DEFAULT 'MEGA',
    ADD COLUMN chassis_skin_id UUID REFERENCES blueprint_mech_skin (id); -- this column is used temp and gets removed.

UPDATE blueprint_chassis bc
SET tier = (SELECT tier FROM templates WHERE blueprint_chassis_id = bc.id);

UPDATE blueprint_chassis c
SET model_id = (SELECT id
                FROM blueprint_mechs cm
                WHERE c.model = cm.label);

ALTER TABLE blueprint_chassis
    DROP COLUMN model,
    ALTER COLUMN model_id SET NOT NULL;

-- SET THE CONNECTED SKINS
UPDATE blueprint_chassis bc
SET chassis_skin_id = (SELECT id
                       FROM blueprint_mech_skin bcs
                       WHERE bcs.label = bc.skin);

-- fix ones we missed somehow
UPDATE chassis_skin
SET label = 'Gumdan'
WHERE label = 'Gundam';

UPDATE chassis_skin SET label = 'White Blue' WHERE label = 'Blue White';

UPDATE chassis_skin ms
SET blueprint_id = (SELECT id
                    FROM blueprint_mech_skin bms
                    WHERE bms.label = ms.label);

-- here
ALTER TABLE chassis_skin
    ALTER COLUMN blueprint_id SET NOT NULL;

/*
  AMMO
 */

CREATE TABLE blueprint_ammo
(
    id                             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    label                          TEXT        NOT NULL,
    weapon_type                    WEAPON_TYPE NOT NULL,
    collection                     COLLECTION  NOT NULL DEFAULT 'supremacy-general',
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


--  insert the energy cores lazily

WITH mechs AS (SELECT c.id, m.owner_id
               FROM chassis c
                        INNER JOIN mechs m ON m.chassis_id = c.id)
INSERT
INTO power_cores(label,
                 size,
                 capacity,
                 max_draw_rate,
                 recharge_rate,
                 armour,
                 max_hitpoints,
                 equipped_on,
                 blueprint_id)
SELECT 'Standard Energy Core',
       'SMALL',
       1000,
       100,
       100,
       0,
       1000,
       mechs.id,
       '62e197a4-f45e-4034-ac0a-3e625a6770d7'
FROM mechs;

WITH pc AS (SELECT _pc.id, _pc.equipped_on
            FROM power_cores _pc)
UPDATE chassis m
SET power_core_id = pc.id
FROM pc
WHERE m.id = pc.equipped_on;

-- this updates all genesis_token_id for chassis that are in genesis, also get is default and is insured
WITH genesis AS (SELECT external_token_id, collection_slug, chassis_id
                 FROM mechs
                 WHERE collection_slug = 'supremacy-genesis')
UPDATE power_cores pc
SET genesis_token_id = genesis.external_token_id
FROM genesis
WHERE pc.equipped_on = genesis.chassis_id;

-- this updates all limited release for chassis that are in genesis, also get is default and is insured
WITH limited_release AS (SELECT external_token_id, collection_slug, chassis_id
                         FROM mechs
                         WHERE collection_slug = 'supremacy-limited-release')
UPDATE power_cores pc
SET limited_release_token_id = limited_release.external_token_id
FROM limited_release
WHERE pc.id = limited_release.chassis_id;

-- This inserts a new collection_items entry for each utility and updates the utility table with token id
WITH power_core AS (SELECT 'power_core' AS item_type, _pc.id, tier, _ci.owner_id
                    FROM power_cores _pc
                             INNER JOIN collection_items _ci ON _ci.item_id = _pc.equipped_on)
INSERT
INTO collection_items (token_id, item_type, item_id, tier, owner_id)
SELECT NEXTVAL('collection_general'),
       power_core.item_type::ITEM_TYPE,
       power_core.id,
       power_core.tier,
       power_core.owner_id
FROM power_core;