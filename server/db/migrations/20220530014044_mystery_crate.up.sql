DROP TYPE IF EXISTS CRATE_TYPE;
CREATE TYPE CRATE_TYPE AS ENUM ('MECH', 'WEAPON');

DROP TYPE IF EXISTS MECH_TYPE;
CREATE TYPE MECH_TYPE AS ENUM ('HUMANOID', 'PLATFORM');

ALTER TYPE ITEM_TYPE ADD VALUE 'mystery_crate';

CREATE TABLE storefront_mystery_crates
(
    id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    mystery_crate_type CRATE_TYPE  NOT NULL,
    price              NUMERIC(28) NOT NULL,
    amount             INT         NOT NULL,
    amount_sold        INT         NOT NULL DEFAULT 0,
    faction_id         UUID        NOT NULL,
    deleted_at         TIMESTAMPTZ,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mystery_crate
(
    id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    type               CRATE_TYPE  NOT NULL,
    faction_id         UUID        NOT NULL REFERENCES factions (id),
    label              TEXT        NOT NULL,
    opened             BOOLEAN     NOT NULL DEFAULT FALSE,
    locked_until       TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '1' YEAR,
    purchased          BOOLEAN     NOT NULL DEFAULT FALSE,
    image_url          TEXT,
    card_animation_url TEXT,
    avatar_url         TEXT,
    large_image_url    TEXT,
    background_color   TEXT,
    animation_url      TEXT,
    youtube_url        TEXT,

    deleted_at         TIMESTAMPTZ,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mystery_crate_blueprints
(
    id               UUID PRIMARY KEY            DEFAULT gen_random_uuid(),
    mystery_crate_id UUID               NOT NULL REFERENCES mystery_crate (id),
    blueprint_type   TEMPLATE_ITEM_TYPE NOT NULL,
    blueprint_id     UUID               NOT NULL,
    deleted_at       TIMESTAMPTZ,
    updated_at       TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    created_at       TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

CREATE TABLE weapon_models
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    faction_id      UUID REFERENCES factions (id),
    brand_id        UUID REFERENCES brands (id),
    label           TEXT        NOT NULL,
    weapon_type     WEAPON_TYPE NOT NULL,
    default_skin_id UUID,
    deleted_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE blueprint_weapon_skin
    ADD COLUMN collection      TEXT NOT NULL DEFAULT 'supremacy-general',
    ADD COLUMN weapon_model_id UUID NOT NULL REFERENCES weapon_models (id),
    ADD COLUMN stat_modifier   NUMERIC(8);

ALTER TABLE weapon_skin
    ADD COLUMN weapon_model_id UUID NOT NULL REFERENCES weapon_models (id);

-- inserting brands
INSERT INTO brands (faction_id, label)
VALUES ((SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Daison Avionics');
INSERT INTO brands (faction_id, label)
VALUES ((SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Archon Miltech');

INSERT INTO brands (faction_id, label)
VALUES ((SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'X3 Wartech');
INSERT INTO brands (faction_id, label)
VALUES ((SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Warsui');

INSERT INTO brands (faction_id, label)
VALUES ((SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'),
        'Unified Martian Corporation');
INSERT INTO brands (faction_id, label)
VALUES ((SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Pyrotronics');

--Insert into weapon models

DROP FUNCTION IF EXISTS insert_weapon_model(weapon_label TEXT, weapon TEXT, faction RECORD);
CREATE FUNCTION insert_weapon_model(weapon_label TEXT, weapon TEXT, faction RECORD) RETURNS VOID
    LANGUAGE plpgsql AS
$$
BEGIN
    INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id)
    VALUES (weapon_label, faction.id, weapon::WEAPON_TYPE,
            CASE
                WHEN faction.label = 'Boston Cybernetics' THEN
                        (SELECT id FROM brands WHERE label = 'Archon Miltech')
                WHEN faction.label = 'Zaibatsu Heavy Industries' THEN
                        (SELECT id FROM brands WHERE label = 'Warsui')
                WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                    THEN
                        (SELECT id FROM brands WHERE label = 'Pyrotronics')
                END);
END;
$$;

DO
$$
    DECLARE
        faction FACTIONS%ROWTYPE;
    BEGIN
        FOR faction IN SELECT * FROM factions
            LOOP
                CASE
                    --BC
                    WHEN faction.label = 'Boston Cybernetics'
                        THEN PERFORM insert_weapon_model('ARCHON SPINSHOT MNG-549', 'Minigun', faction);
                             PERFORM insert_weapon_model('ARCHON STORMLINE PLG-351', 'Plasma Gun', faction);
                             PERFORM insert_weapon_model('ARCHON RUPTURE FSK-745', 'Flak', faction);
                             PERFORM insert_weapon_model('ARCHON DEATH RUSH MCN-777', 'Machine Gun', faction);
                             PERFORM insert_weapon_model('ARCHON REACH MLR-909', 'Missile Launcher', faction);
                             PERFORM insert_weapon_model('ARCHON NEO LSG-636', 'Laser Beam', faction);
                             PERFORM insert_weapon_model('ARCHON PATROL CNO-950', 'Cannon', faction);
                             PERFORM insert_weapon_model('ARCHON AMBUSHER GLR-108', 'Grenade Launcher', faction);
                             PERFORM insert_weapon_model('ARCHON HEAVY BFG-800', 'BFG', faction);

                    --RM
                    WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                        THEN PERFORM insert_weapon_model('M17 SCORPION', 'Minigun', faction);
                             PERFORM insert_weapon_model('P80 PANTHER', 'Plasma Gun', faction);
                             PERFORM insert_weapon_model('K45 PULVERIZER', 'Flak', faction);
                             PERFORM insert_weapon_model('G88 RAVAGER', 'Machine Gun', faction);
                             PERFORM insert_weapon_model('M66 VORTEX', 'Missile Launcher', faction);
                             PERFORM insert_weapon_model('L51 BRUTALIZER', 'Laser Beam', faction);
                             PERFORM insert_weapon_model('C22 CULVERIN', 'Cannon', faction);
                             PERFORM insert_weapon_model('A35 HOMEWRECKER', 'Grenade Launcher', faction);
                             PERFORM insert_weapon_model('F75 HELLFIRE', 'Flamethrower', faction);

                    --ZAI
                    WHEN faction.label = 'Zaibatsu Heavy Industries'
                        THEN PERFORM insert_weapon_model('SM-1400 BRUTE FORCE', 'Minigun', faction);
                             PERFORM insert_weapon_model('SP-750 AVENGER', 'Plasma Gun', faction);
                             PERFORM insert_weapon_model('SF-950 BUSHWHACKER', 'Flak', faction);
                             PERFORM insert_weapon_model('SG-850 FURORE', 'Machine Gun', faction);
                             PERFORM insert_weapon_model('SS-1500 VULCAN', 'Missile Launcher', faction);
                             PERFORM insert_weapon_model('SL-900 SAMURAI', 'Laser Beam', faction);
                             PERFORM insert_weapon_model('CSN-1000 ENDURO', 'Cannon', faction);
                             PERFORM insert_weapon_model('ST-1250 VERSYS', 'Grenade Launcher', faction);
                             PERFORM insert_weapon_model('SL-750 ELIMINATOR', 'Lightning Gun', faction);
                    END CASE;
            END LOOP;
    END;
$$;

-- insert genesis weapons that are not faction specific
INSERT INTO weapon_models (label, weapon_type, brand_id)
VALUES ('Plasma Rifle', 'Rifle', (SELECT id FROM brands WHERE label = 'Boston Cybernetics'));
INSERT INTO weapon_models (label, weapon_type, brand_id)
VALUES ('Auto Cannon', 'Cannon', (SELECT id FROM brands WHERE label = 'Red Mountain Offworld Mining Corporation'));
INSERT INTO weapon_models (label, weapon_type, brand_id)
VALUES ('Sniper Rifle', 'Sniper Rifle', (SELECT id FROM brands WHERE label = 'Zaibatsu Heavy Industries'));
INSERT INTO weapon_models (label, weapon_type)
VALUES ('Rocket Pod', 'Missile Launcher');
INSERT INTO weapon_models (label, weapon_type, brand_id)
VALUES ('Sword', 'Sword', (SELECT id FROM brands WHERE label = 'Boston Cybernetics'));
INSERT INTO weapon_models (label, weapon_type, brand_id)
VALUES ('Laser Sword', 'Sword', (SELECT id FROM brands WHERE label = 'Zaibatsu Heavy Industries'));

-- seed blueprint_weapons_skins
--genesis weapons w/o a faction
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type)
VALUES ('Plasma Rifle', (SELECT id FROM weapon_models WHERE label = 'Plasma Rifle'),
        (SELECT weapon_type FROM weapon_models WHERE label = 'Plasma Rifle'));
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type)
VALUES ('Auto Cannon', (SELECT id FROM weapon_models WHERE label = 'Auto Cannon'),
        (SELECT weapon_type FROM weapon_models WHERE label = 'Auto Cannon'));
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type)
VALUES ('Sniper Rifle', (SELECT id FROM weapon_models WHERE label = 'Sniper Rifle'),
        (SELECT weapon_type FROM weapon_models WHERE label = 'Sniper Rifle'));
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type)
VALUES ('Rocket Pod', (SELECT id FROM weapon_models WHERE label = 'Rocket Pod'),
        (SELECT weapon_type FROM weapon_models WHERE label = 'Rocket Pod'));
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type)
VALUES ('Sword', (SELECT id FROM weapon_models WHERE label = 'Sword'),
        (SELECT weapon_type FROM weapon_models WHERE label = 'Sword'));
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type)
VALUES ('Laser Sword', (SELECT id FROM weapon_models WHERE label = 'Laser Sword'),
        (SELECT weapon_type FROM weapon_models WHERE label = 'Laser Sword'));

-- TODO: change weapon crate skin placeholder names
DO
$$
    DECLARE
        weapon_model WEAPON_MODELS%ROWTYPE;
    BEGIN
        FOR weapon_model IN SELECT *
                            FROM weapon_models
                            WHERE faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics')
            LOOP
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Daison Avionics', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Raptor', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Rexeon Guard', weapon_model.id, weapon_model.weapon_type, 'RARE');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Gold', weapon_model.id, weapon_model.weapon_type, 'LEGENDARY');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Paladin', weapon_model.id, weapon_model.weapon_type, 'EXOTIC');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Hive', weapon_model.id, weapon_model.weapon_type, 'MYTHIC');

                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('BC', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Space Marine', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Nerf Gun', weapon_model.id, weapon_model.weapon_type, 'RARE');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Celtic Knot', weapon_model.id, weapon_model.weapon_type, 'LEGENDARY');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Cybernetics', weapon_model.id, weapon_model.weapon_type, 'EXOTIC');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Doom', weapon_model.id, weapon_model.weapon_type, 'MYTHIC');
            END LOOP;
    END;
$$;

DO
$$
    DECLARE
        weapon_model WEAPON_MODELS%ROWTYPE;
    BEGIN
        FOR weapon_model IN SELECT *
                            FROM weapon_models
                            WHERE faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries')
            LOOP
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('X3W', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('XHANCR', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('2501 - Tachikoma', weapon_model.id, weapon_model.weapon_type, 'RARE');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Gold', weapon_model.id, weapon_model.weapon_type, 'LEGENDARY');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Shinobi', weapon_model.id, weapon_model.weapon_type, 'EXOTIC');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Synth Punk', weapon_model.id, weapon_model.weapon_type, 'MYTHIC');


                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Zaibatsu', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Purple and White', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Sonnō jōi', weapon_model.id, weapon_model.weapon_type, 'RARE');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Logogram - Arrival', weapon_model.id, weapon_model.weapon_type, 'LEGENDARY');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Neko', weapon_model.id, weapon_model.weapon_type, 'EXOTIC');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('BOTW', weapon_model.id, weapon_model.weapon_type, 'MYTHIC');
            END LOOP;
    END;
$$;

DO
$$
    DECLARE
        weapon_model WEAPON_MODELS%ROWTYPE;
    BEGIN
        FOR weapon_model IN SELECT *
                            FROM weapon_models
                            WHERE faction_id =
                                  (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation')
            LOOP
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('UMC', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Military', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Mining', weapon_model.id, weapon_model.weapon_type, 'RARE');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Gold', weapon_model.id, weapon_model.weapon_type, 'LEGENDARY');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Heavy Metal', weapon_model.id, weapon_model.weapon_type, 'EXOTIC');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Molten', weapon_model.id, weapon_model.weapon_type, 'MYTHIC');


                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('RM', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Hazard', weapon_model.id, weapon_model.weapon_type, 'COLOSSAL');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Martian Marine Core', weapon_model.id, weapon_model.weapon_type, 'RARE');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Cassowary', weapon_model.id, weapon_model.weapon_type, 'LEGENDARY');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Damascus', weapon_model.id, weapon_model.weapon_type, 'EXOTIC');
                INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type, tier)
                VALUES ('Dantes Inferno', weapon_model.id, weapon_model.weapon_type, 'MYTHIC');
            END LOOP;
    END;
$$;

-- each weapon model set default skin
DO
$$
    DECLARE
        weapon_model WEAPON_MODELS%ROWTYPE;
    BEGIN
        FOR weapon_model IN SELECT * FROM weapon_models
            LOOP
                CASE
                    WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries')
                        THEN UPDATE weapon_models
                             SET default_skin_id = (SELECT id
                                                    FROM blueprint_weapon_skin
                                                    WHERE weapon_model_id = weapon_model.id
                                                      AND label = 'X3W')
                             WHERE id = weapon_model.id;
                    WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics')
                        THEN UPDATE weapon_models
                             SET default_skin_id = (SELECT id
                                                    FROM blueprint_weapon_skin
                                                    WHERE weapon_model_id = weapon_model.id
                                                      AND label = 'Daison Avionics')
                             WHERE id = weapon_model.id;
                    WHEN weapon_model.faction_id =
                         (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation')
                        THEN UPDATE weapon_models
                             SET default_skin_id = (SELECT id
                                                    FROM blueprint_weapon_skin
                                                    WHERE weapon_model_id = weapon_model.id
                                                      AND label = 'UMC')
                             WHERE id = weapon_model.id;
                    WHEN weapon_model.faction_id IS NULL THEN UPDATE weapon_models
                                                              SET default_skin_id = (SELECT id
                                                                                     FROM blueprint_weapon_skin
                                                                                     WHERE weapon_model_id = weapon_model.id
                                                                                       AND label = weapon_model.label)
                                                              WHERE id = weapon_model.id;
                    END CASE;
            END LOOP;
    END;
$$;

ALTER TABLE weapon_models
    ALTER COLUMN default_skin_id SET NOT NULL;

ALTER TABLE blueprint_weapons
    ADD COLUMN weapon_model_id UUID REFERENCES weapon_models (id);

-- update existing blueprint_weapons (factionless)
DO
$$
    DECLARE
        blueprint_weapon BLUEPRINT_WEAPONS%ROWTYPE;
    BEGIN
        FOR blueprint_weapon IN SELECT * FROM blueprint_weapons
            LOOP
                UPDATE blueprint_weapons
                SET weapon_model_id = (SELECT id
                                       FROM weapon_models
                                       WHERE label = blueprint_weapon.label
                                         AND faction_id IS NULL)
                WHERE label = blueprint_weapon.label;
            END LOOP;
    END;
$$;

ALTER TABLE blueprint_weapons
    ALTER COLUMN weapon_model_id SET NOT NULL;

DO
$$
    DECLARE
        weapon_model WEAPON_MODELS%ROWTYPE;
    BEGIN
        FOR weapon_model IN SELECT * FROM weapon_models WHERE faction_id IS NOT NULL
            LOOP
                INSERT INTO blueprint_weapons (brand_id, label, slug, damage, weapon_type, is_melee, weapon_model_id)
                VALUES (CASE
                            WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics')
                                THEN
                                    (SELECT id FROM brands WHERE label = 'Archon Miltech')
                            WHEN weapon_model.faction_id =
                                 (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries') THEN
                                    (SELECT id FROM brands WHERE label = 'Warsui')
                            WHEN weapon_model.faction_id =
                                 (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation') THEN
                                    (SELECT id FROM brands WHERE label = 'Pyrotronics')
                            END,
                        CONCAT((SELECT label FROM brands WHERE id = weapon_model.brand_id), ' ', weapon_model.label),
                        LOWER(CONCAT(REPLACE((SELECT label FROM brands WHERE id = weapon_model.brand_id), ' ', '_'),
                                     '_', REPLACE(weapon_model.label, ' ', '_'))),
                        0,
                        weapon_model.weapon_type,
                        CASE
                            WHEN weapon_model.weapon_type = 'Sword' THEN TRUE
                            ELSE FALSE
                            END,
                        weapon_model.id);
            END LOOP;
    END;
$$;

ALTER TABLE weapons
    ADD COLUMN weapon_model_id         UUID,
    ADD COLUMN equipped_weapon_skin_id UUID;

--updating existing weapons
UPDATE weapons
SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Plasma Rifle' AND brand_id IS NULL)
WHERE label = 'Plasma Rifle'
  AND brand_id IS NULL;
UPDATE weapons
SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Auto Cannon' AND brand_id IS NULL)
WHERE label = 'Auto Cannon'
  AND brand_id IS NULL;
UPDATE weapons
SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Sniper Rifle' AND brand_id IS NULL)
WHERE label = 'Sniper Rifle'
  AND brand_id IS NULL;
UPDATE weapons
SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Sword' AND brand_id IS NULL)
WHERE label = 'Sword'
  AND brand_id IS NULL;
UPDATE weapons
SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Laser Sword' AND brand_id IS NULL)
WHERE label = 'Laser Sword'
  AND brand_id IS NULL;
UPDATE weapons
SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Rocket Pod' AND brand_id IS NULL)
WHERE label = 'Rocket Pod'
  AND brand_id IS NULL;

ALTER TABLE weapons
    ADD CONSTRAINT fk_weapon_models FOREIGN KEY (weapon_model_id) REFERENCES weapon_models (id);

-- MECHS

ALTER TABLE mech_model
    RENAME TO mech_models;

ALTER TABLE mech_models
    ADD COLUMN brand_id   UUID,
    ADD COLUMN faction_id UUID,
    ADD COLUMN mech_type  MECH_TYPE;

UPDATE mech_models
SET mech_type = 'HUMANOID';
UPDATE mech_models
SET faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics')
WHERE label = 'Law Enforcer X-1000';
UPDATE mech_models
SET faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries')
WHERE label = 'Tenshi Mk1';
UPDATE mech_models
SET faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation')
WHERE label = 'Olympus Mons LY07';

INSERT INTO mech_models (label, mech_type, brand_id, faction_id)
VALUES ('WAR ENFORCER', 'HUMANOID', (SELECT id FROM brands WHERE label = 'Daison Avionics'),
        (SELECT id FROM factions WHERE label = 'Boston Cybernetics'));
INSERT INTO mech_models (label, mech_type, brand_id, faction_id)
VALUES ('ANNIHILATOR', 'PLATFORM', (SELECT id FROM brands WHERE label = 'Daison Avionics'),
        (SELECT id FROM factions WHERE label = 'Boston Cybernetics'));

INSERT INTO mech_models (label, mech_type, brand_id, faction_id)
VALUES ('KENJI', 'HUMANOID', (SELECT id FROM brands WHERE label = 'X3 Wartech'),
        (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'));
INSERT INTO mech_models (label, mech_type, brand_id, faction_id)
VALUES ('SHIROKUMA', 'PLATFORM', (SELECT id FROM brands WHERE label = 'X3 Wartech'),
        (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'));

INSERT INTO mech_models (label, mech_type, brand_id, faction_id)
VALUES ('ARIES', 'HUMANOID', (SELECT id FROM brands WHERE label = 'Unified Martian Corporation'),
        (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'));
INSERT INTO mech_models (label, mech_type, brand_id, faction_id)
VALUES ('VIKING', 'PLATFORM', (SELECT id FROM brands WHERE label = 'Unified Martian Corporation'),
        (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'));

ALTER TABLE mech_models
    ADD CONSTRAINT fk_brands FOREIGN KEY (brand_id) REFERENCES brands (id),
    ADD CONSTRAINT fk_factions FOREIGN KEY (faction_id) REFERENCES factions (id);

ALTER TABLE blueprint_mech_skin
    ADD COLUMN mech_type     MECH_TYPE,
    ADD COLUMN stat_modifier NUMERIC(8);

UPDATE blueprint_mech_skin
SET MECH_TYPE = 'HUMANOID';

-- for each mech_model, insert 6 skins
DO
$$
    DECLARE
        mech_model MECH_MODELS%ROWTYPE;
    BEGIN
        FOR mech_model IN SELECT * FROM mech_models WHERE brand_id IS NOT NULL
            LOOP
                CASE
                    WHEN mech_model.faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics')
                        THEN INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Raptor', mech_model.mech_type, 'COLOSSAL');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Daison Avionics', mech_model.mech_type, 'COLOSSAL');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Rexeon Guard', mech_model.mech_type, 'RARE');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Paladin', mech_model.mech_type, 'EXOTIC');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Hive', mech_model.mech_type, 'MYTHIC');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Gold', mech_model.mech_type, 'LEGENDARY');

                    WHEN mech_model.faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries')
                        THEN INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'XHANCR', mech_model.mech_type, 'COLOSSAL');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'X3W', mech_model.mech_type, 'COLOSSAL');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, '2501 - Tachikoma', mech_model.mech_type, 'RARE');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Shinobi', mech_model.mech_type, 'EXOTIC');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Synth Punk', mech_model.mech_type, 'MYTHIC');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Gold', mech_model.mech_type, 'LEGENDARY');

                    WHEN mech_model.faction_id =
                         (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation')
                        THEN INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Military', mech_model.mech_type, 'COLOSSAL');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'UMC', mech_model.mech_type, 'COLOSSAL');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Mining', mech_model.mech_type, 'RARE');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Heavy Metal', mech_model.mech_type, 'EXOTIC');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Molten', mech_model.mech_type, 'MYTHIC');
                             INSERT INTO blueprint_mech_skin (mech_model, label, mech_type, tier)
                             VALUES (mech_model.id, 'Gold', mech_model.mech_type, 'LEGENDARY');
                    END CASE;
            END LOOP;
    END;
$$;

-- set default skins for mech model

DO
$$
    DECLARE
        model MECH_MODELS%ROWTYPE;
    BEGIN
        FOR model IN SELECT * FROM mech_models WHERE default_chassis_skin_id IS NULL
            LOOP
                CASE
                    WHEN model.faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries')
                        THEN UPDATE mech_models
                             SET default_chassis_skin_id = (SELECT id
                                                            FROM blueprint_mech_skin
                                                            WHERE mech_model = model.id
                                                              AND label = 'X3W')
                             WHERE id = model.id;
                    WHEN model.faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics')
                        THEN UPDATE mech_models
                             SET default_chassis_skin_id = (SELECT id
                                                            FROM blueprint_mech_skin
                                                            WHERE mech_model = model.id
                                                              AND label = 'Daison Avionics')
                             WHERE id = model.id;
                    WHEN model.faction_id =
                         (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation')
                        THEN UPDATE mech_models
                             SET default_chassis_skin_id = (SELECT id
                                                            FROM blueprint_mech_skin
                                                            WHERE mech_model = model.id
                                                              AND label = 'UMC')
                             WHERE id = model.id;
                    WHEN model.faction_id IS NULL THEN UPDATE mech_models
                                                       SET default_chassis_skin_id = (SELECT id
                                                                                      FROM blueprint_mech_skin
                                                                                      WHERE mech_model = model.id
                                                                                        AND label = model.label)
                                                       WHERE id = model.id;
                    END CASE;
            END LOOP;
    END;
$$;

ALTER TABLE mech_models
    ALTER COLUMN default_chassis_skin_id SET NOT NULL;

ALTER TABLE blueprint_mechs
    DROP COLUMN skin;

DO
$$
    DECLARE
        mech_model MECH_MODELS%ROWTYPE;
    BEGIN
        FOR mech_model IN SELECT * FROM mech_models WHERE brand_id IS NOT NULL
            LOOP
                INSERT INTO blueprint_mechs (brand_id, label, slug, weapon_hardpoints, utility_slots, speed,
                                             max_hitpoints, model_id, power_core_size)
                VALUES (mech_model.brand_id,
                        CONCAT((SELECT label FROM brands WHERE id = mech_model.brand_id), ' ', mech_model.label),
                        LOWER(CONCAT(REPLACE((SELECT label FROM brands WHERE id = mech_model.brand_id), ' ', '_'), '_',
                                     REPLACE(mech_model.label, ' ', '_'))),
                        CASE --hardpoints
                            WHEN mech_model.mech_type = 'PLATFORM' THEN 5
                            ELSE 2
                            END,
                        CASE -- utility slots
                            WHEN mech_model.mech_type = 'PLATFORM' THEN 2
                            ELSE 4
                            END,
                        CASE --speed
                            WHEN mech_model.mech_type = 'PLATFORM' THEN 1000
                            ELSE 2000
                            END,
                        CASE --max_hitpoints
                            WHEN mech_model.mech_type = 'PLATFORM' THEN 3000
                            ELSE 1500
                            END,
                        mech_model.id,
                        CASE --max_hitpoints
                            WHEN mech_model.mech_type = 'PLATFORM' THEN 'MEDIUM'
                            ELSE 'SMALL'
                            END);
            END LOOP;
    END;
$$;


INSERT INTO blueprint_power_cores (collection, label, size, capacity, max_draw_rate, recharge_rate, armour,
                                   max_hitpoints)
VALUES ('supremacy-general', 'Medium Energy Core', 'MEDIUM', 1500, 150, 100, 0, 1500);

-- seeding mystery crates
-- looping over each type of mystery crate type for x amount of crates for each faction. can do 1 big loop if all crate types have the same amount
DO
$$
    BEGIN
        FOR COUNT IN 1..5000 --change to 5000
            LOOP
                INSERT INTO mystery_crate (type, faction_id, label)
                VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Red Mountain Offworld Mining Corporation'),
                        'Red Mountain War Machine Mystery Crate');
                INSERT INTO mystery_crate (type, faction_id, label)
                VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Zaibatsu Heavy Industries'),
                        'Zaibatsu Nexus War Machine Mystery Crate');
                INSERT INTO mystery_crate (type, faction_id, label)
                VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Boston Cybernetics'),
                        'Boston Cybernetics Nexus War Machine Mystery Crate');
            END LOOP;
    END;
$$;

DO
$$
    BEGIN
        FOR COUNT IN 1..13000 -- change to 13000
            LOOP
                INSERT INTO mystery_crate (type, faction_id, label)
                VALUES ('WEAPON',
                        (SELECT id FROM factions f WHERE f.label = 'Red Mountain Offworld Mining Corporation'),
                        'Red Mountain Nexus Weapon Mystery Crate');
                INSERT INTO mystery_crate (type, faction_id, label)
                VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Zaibatsu Heavy Industries'),
                        'Zaibatsu Nexus Weapon Mystery Crate');
                INSERT INTO mystery_crate (type, faction_id, label)
                VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Boston Cybernetics'),
                        'Boston Cybernetics Nexus Weapon Mystery Crate');
            END LOOP;
    END;
$$;

-- seeding crate blueprints
DROP FUNCTION IF EXISTS insert_mech_into_crate(core_size TEXT, mechCrate_id UUID, faction RECORD);
CREATE FUNCTION insert_mech_into_crate(core_size TEXT, mechCrate_id UUID, faction RECORD) RETURNS VOID
    LANGUAGE plpgsql AS
$$
BEGIN
    INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
    VALUES (mechCrate_id, 'MECH', (SELECT id
                                   FROM blueprint_mechs
                                   WHERE blueprint_mechs.power_core_size = core_size
                                     AND blueprint_mechs.brand_id =
                                         CASE
                                             WHEN faction.label = 'Boston Cybernetics'
                                                 THEN (SELECT id FROM brands WHERE label = 'Daison Avionics')
                                             WHEN faction.label = 'Zaibatsu Heavy Industries'
                                                 THEN (SELECT id FROM brands WHERE label = 'X3 Wartech')
                                             WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                 THEN (SELECT id FROM brands WHERE label = 'Unified Martian Corporation')
                                             END));
END;
$$;

DROP FUNCTION IF EXISTS insert_mech_skin_into_crate(mechCrate_id UUID, mechType TEXT, skinLabel TEXT, faction RECORD);
CREATE FUNCTION insert_mech_skin_into_crate(mechCrate_id UUID, mechType TEXT, skinLabel TEXT, faction RECORD) RETURNS VOID
    LANGUAGE plpgsql AS
$$
BEGIN

    INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
    VALUES (mechCrate_id, 'MECH_SKIN', (SELECT id
                                        FROM blueprint_mech_skin
                                        WHERE mech_type = mechType::MECH_TYPE
                                          AND blueprint_mech_skin.mech_model =
                                              CASE
                                                  WHEN faction.label = 'Boston Cybernetics'
                                                      THEN (SELECT id
                                                            FROM mech_models mm
                                                            WHERE mm.mech_type = mechType::MECH_TYPE
                                                              AND mm.brand_id =
                                                                  (SELECT id FROM brands WHERE brands.label = 'Daison Avionics'))
                                                  WHEN faction.label = 'Zaibatsu Heavy Industries'
                                                      THEN (SELECT id
                                                            FROM mech_models mm
                                                            WHERE mm.mech_type = mechType::MECH_TYPE
                                                              AND mm.brand_id =
                                                                  (SELECT id FROM brands WHERE brands.label = 'X3 Wartech'))
                                                  WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                      THEN (SELECT id
                                                            FROM mech_models mm
                                                            WHERE mm.mech_type = mechType::MECH_TYPE
                                                              AND mm.brand_id =
                                                                  (SELECT id
                                                                   FROM brands
                                                                   WHERE brands.label = 'Unified Martian Corporation'))
                                                  END
                                          AND label = skinLabel));
END;
$$;

DROP FUNCTION IF EXISTS get_mech_skin_label_rarity(i INTEGER, type TEXT, amount_of_type NUMERIC,
                                                   previous_crates NUMERIC,
                                                   faction RECORD);
CREATE FUNCTION get_mech_skin_label_rarity(i INTEGER, type TEXT, amount_of_type NUMERIC, previous_crates NUMERIC,
                                           faction RECORD) RETURNS TEXT
    LANGUAGE plpgsql AS
$$
BEGIN
    RETURN CASE
        --30% colossal
               WHEN i <= ((.30 * amount_of_type) + previous_crates)
                   THEN
                   CASE
                       WHEN type = 'MECH' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Daison Avionics'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'X3W'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'UMC'
                               END
                       WHEN type = 'WEAPON' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'BC'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Zaibatsu'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'RM'
                               END
                       END


        --30% colossal
               WHEN i > ((.30 * amount_of_type) + previous_crates) AND
                    i <= ((.60 * amount_of_type) + previous_crates)
                   THEN
                   CASE
                       WHEN type = 'MECH' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Raptor'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'XHANCR'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Military'
                               END
                       WHEN type = 'WEAPON' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Space Marine'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Purple and White'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Hazard'
                               END
                       END

        --15% rare
               WHEN i > ((.60 * amount_of_type) + previous_crates) AND
                    i <= ((.75 * amount_of_type) + previous_crates)
                   THEN
                   CASE
                       WHEN type = 'MECH' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Rexeon Guard'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN '2501 - Tachikoma'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Mining'
                               END
                       WHEN type = 'WEAPON' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Nerf Gun'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Sonnō jōi'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Martian Marine Core'
                               END
                       END

        --12.5% legendary
               WHEN i > ((.75 * amount_of_type) + previous_crates) AND
                    i <= ((.875 * amount_of_type) + previous_crates)
                   THEN CASE
                            WHEN type = 'MECH' THEN 'Gold'
                            WHEN type = 'WEAPON' THEN
                                CASE
                                    WHEN faction.label = 'Boston Cybernetics'
                                        THEN 'Celtic Knot'
                                    WHEN faction.label = 'Zaibatsu Heavy Industries'
                                        THEN 'Logogram - Arrival'
                                    WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                        THEN 'Cassowary'
                                    END
                   END

        --10% for exotic
               WHEN i > ((.875 * amount_of_type) + previous_crates) AND
                    i <= ((.975 * amount_of_type) + previous_crates)
                   THEN
                   CASE
                       WHEN type = 'MECH' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Paladin'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Shinobi'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Heavy Metal'
                               END
                       WHEN type = 'WEAPON' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Cybernetics'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Neko'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Damascus'
                               END
                       END
        --2.5% mythic
               WHEN i > ((.975 * amount_of_type) + previous_crates) AND
                    i <= ((1 * amount_of_type) + previous_crates)
                   THEN
                   CASE
                       WHEN type = 'MECH' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Hive'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Synth Punk'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Molten'
                               END
--                        this will push into rare to account for no mythics for MOST weapons
                       WHEN type = 'WEAPON' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Nerf Gun'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Sonnō jōi'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Martian Marine Core'
                               END
                       END
        END;
END;
$$;

DROP FUNCTION IF EXISTS insert_weapon_skin_into_crate(i INTEGER, weaponCrate_id UUID, weaponType TEXT,
                                                      amount_of_type NUMERIC,
                                                      previous_crates NUMERIC, type TEXT, skin_label TEXT,
                                                      faction RECORD);
CREATE FUNCTION insert_weapon_skin_into_crate(i INTEGER, weaponCrate_id UUID, weaponType TEXT,
                                              amount_of_type NUMERIC,
                                              previous_crates NUMERIC, type TEXT, skin_label TEXT,
                                              faction RECORD) RETURNS VOID
    LANGUAGE plpgsql AS
$$
BEGIN
    INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
    VALUES (weaponCrate_id, 'WEAPON_SKIN', (SELECT id
                                            FROM blueprint_weapon_skin
                                            WHERE weapon_type = weaponType::WEAPON_TYPE
                                              AND blueprint_weapon_skin.weapon_model_id =
                                                  CASE
                                                      WHEN faction.label = 'Boston Cybernetics'
                                                          THEN (SELECT id
                                                                FROM weapon_models
                                                                WHERE weapon_type = weaponType::WEAPON_TYPE
                                                                  AND brand_id = (SELECT id FROM brands WHERE label = 'Archon Miltech'))
                                                      WHEN faction.label = 'Zaibatsu Heavy Industries'
                                                          THEN (SELECT id
                                                                FROM weapon_models
                                                                WHERE weapon_type = weaponType::WEAPON_TYPE
                                                                  AND brand_id = (SELECT id FROM brands WHERE label = 'Warsui'))
                                                      WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                          THEN (SELECT id
                                                                FROM weapon_models
                                                                WHERE weapon_type = weaponType::WEAPON_TYPE
                                                                  AND brand_id = (SELECT id FROM brands WHERE label = 'Pyrotronics'))
                                                      END
                                              AND label =
                                                  CASE
                                                      WHEN type = 'MECH'
                                                          THEN skin_label
                                                      ELSE get_mech_skin_label_rarity(i, 'WEAPON', amount_of_type,
                                                                                      previous_crates,
                                                                                      faction)
                                                      END
--
    ));
END;
$$;

DROP FUNCTION IF EXISTS insert_weapon_into_crate(crate_id UUID, weaponType TEXT, faction RECORD);
CREATE FUNCTION insert_weapon_into_crate(crate_id UUID, weaponType TEXT, faction RECORD) RETURNS VOID
    LANGUAGE plpgsql AS
$$
BEGIN
    INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
    VALUES (crate_id, 'WEAPON', (SELECT id
                                 FROM blueprint_weapons
                                 WHERE weapon_type = weaponType::WEAPON_TYPE
                                   AND brand_id =
                                       CASE
                                           WHEN faction.label = 'Boston Cybernetics'
                                               THEN (SELECT id FROM brands WHERE label = 'Archon Miltech')
                                           WHEN faction.label = 'Zaibatsu Heavy Industries'
                                               THEN (SELECT id FROM brands WHERE label = 'Warsui')
                                           WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                               THEN (SELECT id FROM brands WHERE label = 'Pyrotronics')
                                           END));
END;
$$;


DO
$$
    DECLARE
        faction                 FACTIONS%ROWTYPE;
        DECLARE mechCrate       MYSTERY_CRATE%ROWTYPE;
        DECLARE weaponCrate     MYSTERY_CRATE%ROWTYPE;
        DECLARE i               INTEGER;
        DECLARE mechCrateLen    INTEGER;
        DECLARE mech_skin_label TEXT;

    BEGIN
        --for each faction loop over the mystery crates of specified faction
        FOR faction IN SELECT * FROM factions
            LOOP
                i := 1;
                mechCrateLen :=
                        (SELECT COUNT(*) FROM mystery_crate WHERE faction_id = faction.id AND type = 'MECH');
                -- for half of the Mechs, insert a mech object from the appropriate brand's bipedal mechs and a fitted power core
                FOR mechCrate IN SELECT * FROM mystery_crate WHERE faction_id = faction.id AND type = 'MECH'
                    LOOP

                        mech_skin_label := get_mech_skin_label_rarity(i, 'MECH', (.8 * mechCrateLen), 0, faction);
                        CASE
                            WHEN i <= (mechCrateLen * .8) -- seed humanoid mechs
                                THEN PERFORM insert_mech_into_crate('SMALL', mechCrate.id, faction);
                                     PERFORM insert_mech_skin_into_crate(mechCrate.id, 'HUMANOID',
                                                                         mech_skin_label, faction);

                                     INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                                     VALUES (mechCrate.id, 'POWER_CORE', (SELECT id
                                                                          FROM blueprint_power_cores c
                                                                          WHERE c.label = 'Standard Energy Core'));

                                     PERFORM insert_weapon_into_crate(mechCrate.id, 'Flak', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, mechCrate.id, 'Flak',
                                                                           (.8 * mechCrateLen),
                                                                           0, 'MECH', mech_skin_label,
                                                                           faction);

                                     PERFORM insert_weapon_into_crate(mechCrate.id, 'Machine Gun', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, mechCrate.id, 'Machine Gun',
                                                                           (.8 * mechCrateLen),
                                                                           0, 'MECH', mech_skin_label,
                                                                           faction);


                                     i := i + 1;
                            -- for other half of the Mechs, insert a mech object from the appropriate brand's platform mechs and a fitted power core
                            ELSE mech_skin_label :=
                                         get_mech_skin_label_rarity(i, 'MECH', (.2 * mechCrateLen), .8 * mechCrateLen,
                                                                    faction);

                                 PERFORM insert_mech_into_crate('MEDIUM', mechCrate.id, faction);
                                 PERFORM insert_mech_skin_into_crate(mechCrate.id, 'PLATFORM',
                                                                     mech_skin_label, faction);

                                 INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                                 VALUES (mechCrate.id, 'POWER_CORE',
                                         (SELECT id FROM blueprint_power_cores c WHERE c.size = 'MEDIUM'));


                                 PERFORM insert_weapon_into_crate(mechCrate.id, 'Flak', faction);
                                 PERFORM insert_weapon_skin_into_crate(i, mechCrate.id, 'Flak',
                                                                       (.2 * mechCrateLen),
                                                                       (.8 * mechCrateLen), 'MECH', mech_skin_label,
                                                                       faction);

                                 PERFORM insert_weapon_into_crate(mechCrate.id, 'Machine Gun', faction);
                                 PERFORM insert_weapon_skin_into_crate(i, mechCrate.id, 'Machine Gun',
                                                                       (.2 * mechCrateLen),
                                                                       (.8 * mechCrateLen), 'MECH', mech_skin_label,
                                                                       faction);
                                 i = i + 1;
                            END CASE;
                    END LOOP;

                -- for weapons crates of each faction, insert weapon blueprint.
                i := 1;
                FOR weaponCrate IN SELECT * FROM mystery_crate WHERE faction_id = faction.id AND type = 'WEAPON'
                    LOOP
                        CASE
                            --minigun: all factions
                            WHEN i <= 2000
                                THEN PERFORM insert_weapon_into_crate(weaponCrate.id, 'Minigun', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponCrate.id, 'Minigun',
                                                                           2000,
                                                                           0, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --missile launcher: all factions
                            WHEN i > 2000 AND i <= 4000
                                THEN PERFORM insert_weapon_into_crate(weaponCrate.id, 'Missile Launcher', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponCrate.id,
                                                                           'Missile Launcher',
                                                                           2000,
                                                                           2000, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --plasma gun: all factions
                            WHEN i > 4000 AND i <= 6000
                                THEN PERFORM insert_weapon_into_crate(weaponCrate.id, 'Plasma Gun', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponCrate.id, 'Plasma Gun',
                                                                           2000,
                                                                           4000, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --Laser beam: all factions
                            WHEN i > 6000 AND i <= 8000
                                THEN PERFORM insert_weapon_into_crate(weaponCrate.id, 'Laser Beam', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponCrate.id, 'Laser Beam',
                                                                           2000,
                                                                           6000, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --cannon: all factions
                            WHEN i > 8000 AND i <= 10000
                                THEN PERFORM insert_weapon_into_crate(weaponCrate.id, 'Cannon', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponCrate.id, 'Cannon',
                                                                           2000,
                                                                           8000, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --grenade launcher: all factions
                            WHEN i > 10000 AND i <= 12000
                                THEN PERFORM insert_weapon_into_crate(weaponCrate.id, 'Grenade Launcher', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponCrate.id, 'Grenade Launcher',
                                                                           2000,
                                                                           10000, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --BFG, flamethrower or Lightning Gun dependent on faction
                            WHEN i > 12000 AND i <= 13000
                                THEN PERFORM insert_weapon_into_crate(weaponCrate.id,
                                                                      CASE
                                                                          WHEN faction.label = 'Boston Cybernetics'
                                                                              THEN 'BFG'
                                                                          WHEN faction.label = 'Zaibatsu Heavy Industries'
                                                                              THEN 'Lightning Gun'
                                                                          WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                                              THEN 'Flamethrower'
                                                                          END, faction);
                                     CASE
                                         WHEN i > 12000 AND i <= 12700
                                             THEN PERFORM insert_weapon_skin_into_crate(i, mechCrate.id,
                                                                                        CASE
                                                                                            WHEN faction.label = 'Boston Cybernetics'
                                                                                                THEN 'BFG'
                                                                                            WHEN faction.label = 'Zaibatsu Heavy Industries'
                                                                                                THEN 'Lightning Gun'
                                                                                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                                                                THEN 'Flamethrower'
                                                                                            END,
                                                                                        0,
                                                                                        0,
                                                                                        'MECH',
                                                                                        CASE
                                                                                            WHEN faction.label = 'Boston Cybernetics'
                                                                                                THEN 'Cybernetics'
                                                                                            WHEN faction.label = 'Zaibatsu Heavy Industries'
                                                                                                THEN 'Neko'
                                                                                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                                                                THEN 'Damascus'
                                                                                            END,
                                                                                        faction);
                                         WHEN i > 12700 AND i <= 13000
                                             THEN PERFORM insert_weapon_skin_into_crate(i, mechCrate.id,
                                                                                        CASE
                                                                                            WHEN faction.label = 'Boston Cybernetics'
                                                                                                THEN 'BFG'
                                                                                            WHEN faction.label = 'Zaibatsu Heavy Industries'
                                                                                                THEN 'Lightning Gun'
                                                                                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                                                                THEN 'Flamethrower'
                                                                                            END,
                                                                                        0,
                                                                                        0,
                                                                                        'MECH',
                                                                                        CASE
                                                                                            WHEN faction.label = 'Boston Cybernetics'
                                                                                                THEN 'Doom'
                                                                                            WHEN faction.label = 'Zaibatsu Heavy Industries'
                                                                                                THEN 'BOTW'
                                                                                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                                                                THEN 'Dantes Inferno'
                                                                                            END,
                                                                                        faction);
                                         END CASE;
                                     i := i + 1;
                            END CASE;
                    END LOOP;

            END LOOP;
    END;
$$;

ALTER TABLE weapon_models
    DROP COLUMN faction_id;
ALTER TABLE mech_models
    DROP COLUMN faction_id;

--seeding storefront
-- for each faction, seed each type of crate and find how much are for sale
DO
$$
    DECLARE
        faction FACTIONS%ROWTYPE;
    BEGIN
        FOR faction IN SELECT * FROM factions
            LOOP
                INSERT INTO storefront_mystery_crates (mystery_crate_type, amount, faction_id, price)
                VALUES ('MECH', (SELECT COUNT(*) FROM mystery_crate WHERE type = 'MECH' AND faction_id = faction.id),
                        faction.id, 3000000000000000000000);
                INSERT INTO storefront_mystery_crates (mystery_crate_type, amount, faction_id, price)
                VALUES ('WEAPON',
                        (SELECT COUNT(*) FROM mystery_crate WHERE type = 'WEAPON' AND faction_id = faction.id),
                        faction.id, 1800000000000000000000);
            END LOOP;
    END;
$$;

