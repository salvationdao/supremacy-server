DROP TYPE IF EXISTS UTILITY_TYPE;
CREATE TYPE UTILITY_TYPE AS ENUM ('SHIELD', 'ATTACK DRONE', 'REPAIR DRONE', 'ANTI MISSILE', 'ACCELERATOR');


/*
  UTILITY
 */

ALTER TABLE blueprint_chassis_blueprint_modules
    RENAME TO blueprint_chassis_blueprint_utility;
ALTER TABLE blueprint_chassis_blueprint_utility
    RENAME COLUMN blueprint_module_id TO blueprint_utility_id;
ALTER TABLE blueprint_modules
    DROP CONSTRAINT blueprint_modules_label_key;
ALTER TABLE blueprint_modules
    RENAME TO blueprint_utility;
ALTER TABLE blueprint_utility
    DROP COLUMN hitpoint_modifier,
    DROP COLUMN shield_modifier,
    ADD COLUMN type       UTILITY_TYPE,
    ADD COLUMN tier       TEXT       NOT NULL DEFAULT 'MEGA',
    ADD COLUMN collection COLLECTION NOT NULL DEFAULT 'supremacy-general';

UPDATE blueprint_utility
SET type = 'SHIELD';

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
ALTER TABLE chassis_utility
    DROP COLUMN IF EXISTS tier,
    DROP COLUMN IF EXISTS owner_id;

ALTER TABLE modules
    RENAME TO utility;
ALTER TABLE utility
    DROP COLUMN hitpoint_modifier,
    DROP COLUMN shield_modifier,
    ADD COLUMN blueprint_id             UUID REFERENCES blueprint_utility (id),
    ADD COLUMN genesis_token_id         NUMERIC,
    ADD COLUMN limited_release_token_id NUMERIC,
    ADD COLUMN owner_id                 UUID REFERENCES players (id),
    ADD COLUMN equipped_on              UUID REFERENCES chassis (id),
    ADD COLUMN tier                     TEXT NOT NULL DEFAULT 'MEGA',
    ADD COLUMN type                     UTILITY_TYPE;

WITH utility_owners AS (SELECT m.owner_id, cu.utility_id
                        FROM chassis_utility cu
                                 INNER JOIN mechs m ON cu.chassis_id = m.chassis_id)
UPDATE utility u
SET owner_id = utility_owners.owner_id
FROM utility_owners
WHERE u.id = utility_owners.utility_id;


-- This inserts a new collection_items entry for each utility and updates the utility table with token id
WITH utily AS (SELECT 'utility' AS item_type, id, tier, owner_id FROM utility)
INSERT
INTO collection_items (token_id, item_type, item_id, tier, owner_id)
SELECT NEXTVAL('collection_general'), utily.item_type::ITEM_TYPE, utily.id, utily.tier, utily.owner_id
FROM utily;


-- this updates all genesis_token_id for weapons that are in genesis
WITH genesis AS (SELECT external_token_id, m.collection_slug, m.chassis_id, _cu.utility_id
                 FROM chassis_utility _cu
                          INNER JOIN mechs m ON m.chassis_id = _cu.chassis_id
                 WHERE m.collection_slug = 'supremacy-genesis')
UPDATE utility u
SET genesis_token_id = genesis.external_token_id
FROM genesis
WHERE u.id = genesis.utility_id;

-- this updates all limited_release for weapons that are in genesis
WITH limited_release AS (SELECT external_token_id, m.collection_slug, m.chassis_id, _cu.utility_id
                         FROM chassis_utility _cu
                                  INNER JOIN mechs m ON m.chassis_id = _cu.chassis_id
                         WHERE m.collection_slug = 'supremacy-limited-release')
UPDATE utility u
SET limited_release_token_id = limited_release.external_token_id
FROM limited_release
WHERE u.id = limited_release.utility_id;

-- ALTER TABLE utility
--     DROP COLUMN slug,
--     ALTER COLUMN owner_id SET NOT NULL;

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
    DROP COLUMN IF EXISTS tier,
    DROP COLUMN IF EXISTS owner_id,
    DROP COLUMN IF EXISTS max_shield;


-- for each of the

-- adding temp columns to make inserting all the new ulti easier
ALTER TABLE blueprint_utility
    ADD COLUMN max_shield           INT,
    ADD COLUMN shield_recharge_rate INT,
    DROP COLUMN IF EXISTS slug;


--  Create all the shield utility modules
WITH insrt AS (
    WITH new_ulti AS (
        SELECT '360 Shield'            AS label,
               'SHIELD'::UTILITY_TYPE  AS type,
               _c.max_shield           AS max_shield,
               _c.shield_recharge_rate AS shield_recharge_rate
        FROM blueprint_chassis_blueprint_utility _cu
                 INNER JOIN blueprint_chassis _c ON _c.id = _cu.blueprint_chassis_id
        GROUP BY _c.max_shield, _c.shield_recharge_rate )
        INSERT INTO blueprint_utility (label, type, max_shield, shield_recharge_rate)
            SELECT new_ulti.label,
                   new_ulti.type,
                   new_ulti.max_shield,
                   new_ulti.shield_recharge_rate
            FROM new_ulti RETURNING id, max_shield, shield_recharge_rate)
INSERT
INTO blueprint_utility_shield (blueprint_utility_id, hitpoints, recharge_rate, recharge_energy_cost)
SELECT insrt.id, insrt.max_shield, insrt.shield_recharge_rate, 10
FROM insrt;

UPDATE utility
SET label = '360 Shield'
WHERE label = 'Shield';

-- clear old joins
DELETE
FROM blueprint_chassis_blueprint_utility;

-- create new joins between blueprint chassis and blueprint shields
WITH bm AS (SELECT _c.id                   AS chassis_id,
                   _c.max_shield           AS max_shield,
                   _c.shield_recharge_rate AS shield_recharge_rate
            FROM blueprint_chassis _c)
INSERT
INTO blueprint_chassis_blueprint_utility (blueprint_chassis_id, blueprint_utility_id, slot_number)
SELECT bm.chassis_id,
       (SELECT bus.blueprint_utility_id
        FROM blueprint_utility_shield bus
        WHERE bus.recharge_rate = bm.shield_recharge_rate
          AND bus.hitpoints = bm.max_shield),
       0
FROM bm;

-- deleting old shield
DELETE
FROM blueprint_utility
WHERE max_shield IS NULL;

-- removing temp columns
ALTER TABLE blueprint_utility
    DROP COLUMN IF EXISTS max_shield,
    DROP COLUMN IF EXISTS shield_recharge_rate;

ALTER TABLE blueprint_chassis
    DROP COLUMN IF EXISTS shield_recharge_rate,
    DROP COLUMN IF EXISTS max_shield;


--  below adds the blueprint id for the shields
WITH shield AS (SELECT hitpoints, recharge_rate FROM utility_shield)
UPDATE utility
SET blueprint_id = (SELECT blueprint_utility_id
                    FROM blueprint_utility_shield _bus
                    WHERE _bus.recharge_rate = shield.recharge_rate
                      AND _bus.hitpoints = shield.hitpoints)
FROM shield;

ALTER TABLE utility
    DROP COLUMN slug,
    ALTER COLUMN owner_id SET NOT NULL,
    ALTER COLUMN type SET NOT NULL,
    ALTER COLUMN blueprint_id SET NOT NULL;