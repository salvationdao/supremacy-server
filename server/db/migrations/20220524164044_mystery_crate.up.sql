DROP TYPE IF EXISTS CRATE_TYPE;
CREATE TYPE CRATE_TYPE AS ENUM ('MECH', 'WEAPON', 'UTILITY');

DROP TYPE IF EXISTS MECH_TYPE;
CREATE TYPE MECH_TYPE AS ENUM ('HUMANOID', 'PLATFORM');

CREATE TABLE storefront_mystery_crates
(
    id                 UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    mystery_crate_type CRATE_TYPE NOT NULL,
    price              numeric(28)     NOT NULL,
    amount             INT             NOT NULL,
    amount_sold        INT             NOT NULL DEFAULT 0,
    faction_id         UUID            NOT NULL,
    deleted_at         TIMESTAMPTZ,
    updated_at         TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_at         TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE TABLE mystery_crate
(
    id           UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    type         CRATE_TYPE_ENUM NOT NULL,
    faction_id   UUID            NOT NULL REFERENCES factions (id),
    label        TEXT            NOT NULL,
    opened       BOOLEAN         NOT NULL DEFAULT false,
    locked_until TIMESTAMPTZ     NOT NULL DEFAULT NOW() + INTERVAL '30' DAY,
    purchased    BOOLEAN         NOT NULL DEFAULT false
);

CREATE TABLE mystery_crate_blueprints
(
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mystery_crate_id UUID               NOT NULL REFERENCES mystery_crate (id),
    blueprint_type   TEMPLATE_ITEM_TYPE NOT NULL,
    blueprint_id     UUID               NOT NULL
);

CREATE TABLE weapon_models
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    faction_id      UUID REFERENCES factions(id),
    brand_id UUID REFERENCES brands(id),
    label           TEXT        UNIQUE NOT NULL,
    weapon_type     WEAPON_TYPE NOT NULL,
    default_skin_id UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE blueprint_weapon_skin
    ADD COLUMN collection TEXT NOT NULL DEFAULT 'supremacy-general',
    ADD COLUMN weapon_model_id UUID NOT NULL REFERENCES weapon_models(id); -- make this not null later
--     ADD COLUMN card_animation_url TEXT,
--     ADD COLUMN large_image_url TEXT,
--     ADD COLUMN avatar_url TEXT;

ALTER TABLE weapon_skin
    ADD COLUMN weapon_model_id UUID NOT NULL REFERENCES weapon_models (id);
--     ADD COLUMN image_url TEXT, -- make this not null later
--     ADD COLUMN animation_url TEXT,
--     ADD COLUMN card_animation_url TEXT,
--     ADD COLUMN large_image_url TEXT,
--     ADD COLUMN avatar_url TEXT;

-- ALTER TABLE blueprint_power_cores
--     ADD COLUMN image_url          TEXT,
--     ADD COLUMN card_animation_url TEXT,
--     ADD COLUMN avatar_url         TEXT,
--     ADD COLUMN large_image_url    TEXT,
--     ADD COLUMN description        TEXT,
--     ADD COLUMN background_color   TEXT,
--     ADD COLUMN animation_url      TEXT,
--     ADD COLUMN youtube_url        TEXT;

-- inserting brands
INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Daison Avionics');
INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Archon Miltech');

INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'x3 Wartech');
INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Warsui');

INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Unified Martian Corporation');
INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Pyrotronics');

-- inserting common weapons
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Archon Miltech Flak', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Flak', (SELECT id FROM brands WHERE label = 'Archon Miltech'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Archon Miltech Machine Gun', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Machine Gun', (SELECT id FROM brands WHERE label = 'Archon Miltech'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Archon Miltech Flamethrower', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Flamethrower', (SELECT id FROM brands WHERE label = 'Archon Miltech'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Archon Miltech Missile Launcher', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Missile Launcher', (SELECT id FROM brands WHERE label = 'Archon Miltech'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Archon Miltech Laser Beam', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Laser Beam', (SELECT id FROM brands WHERE label = 'Archon Miltech'));

INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Warsui Flak', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Flak', (SELECT id FROM brands WHERE label = 'Warsui'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Warsui Machine Gun', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Machine Gun', (SELECT id FROM brands WHERE label = 'Warsui'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Warsui Flamethrower', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Flamethrower', (SELECT id FROM brands WHERE label = 'Warsui'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Warsui Missile Launcher', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Missile Launcher', (SELECT id FROM brands WHERE label = 'Warsui'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Warsui Laser Beam', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Laser Beam', (SELECT id FROM brands WHERE label = 'Warsui'));

INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Pyrotronics Flak', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Flak', (SELECT id FROM brands WHERE label = 'Pyrotronics'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Pyrotronics Machine Gun', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Machine Gun', (SELECT id FROM brands WHERE label = 'Pyrotronics'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Pyrotronics Flamethrower', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Flamethrower', (SELECT id FROM brands WHERE label = 'Pyrotronics'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Pyrotronics Missile Launcher', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Missile Launcher', (SELECT id FROM brands WHERE label = 'Pyrotronics'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Pyrotronics Laser Beam', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Laser Beam', (SELECT id FROM brands WHERE label = 'Pyrotronics'));

-- Inserting specific weapons for each faction
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Archon Miltech Plasma Gun', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Plasma Gun', (SELECT id FROM brands WHERE label = 'Archon Miltech'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Archon Miltech Minigun', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Minigun', (SELECT id FROM brands WHERE label = 'Archon Miltech'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Archon Miltech BFG', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'BFG', (SELECT id FROM brands WHERE label = 'Archon Miltech'));


INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Warsui Plasma Gun', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Plasma Gun', (SELECT id FROM brands WHERE label = 'Warsui'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Warsui Cannon', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Cannon', (SELECT id FROM brands WHERE label = 'Warsui'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Warsui Lightning Gun', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Lightning Gun', (SELECT id FROM brands WHERE label = 'Warsui'));


INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Pyrotronics Minigun', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Minigun', (SELECT id FROM brands WHERE label = 'Pyrotronics'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Pyrotronics Cannon', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Cannon', (SELECT id FROM brands WHERE label = 'Pyrotronics'));
INSERT INTO weapon_models (label, faction_id, weapon_type, brand_id) VALUES ('Pyrotronics Grenade Launcher', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Grenade Launcher', (SELECT id FROM brands WHERE label = 'Pyrotronics'));

-- insert genesis weapons that are not faction specific
INSERT INTO weapon_models (label, weapon_type, brand_id) VALUES ('Plasma Rifle', 'Rifle', (SELECT id FROM brands WHERE label = 'Boston Cybernetics') );
INSERT INTO weapon_models (label, weapon_type, brand_id) VALUES ('Auto Cannon', 'Cannon', (SELECT id FROM brands WHERE label = 'Red Mountain Offworld Mining Corporation'));
INSERT INTO weapon_models (label, weapon_type, brand_id) VALUES ('Sniper Rifle', 'Sniper Rifle', (SELECT id FROM brands WHERE label = 'Zaibatsu Heavy Industries'));
INSERT INTO weapon_models (label, weapon_type) VALUES ('Rocket Pod', 'Missile Launcher');
INSERT INTO weapon_models (label, weapon_type, brand_id) VALUES ('Sword', 'Sword', (SELECT id FROM brands WHERE label = 'Boston Cybernetics') );
INSERT INTO weapon_models (label, weapon_type, brand_id) VALUES ('Laser Sword', 'Sword', (SELECT id FROM brands WHERE label = 'Zaibatsu Heavy Industries'));

-- seed blueprint_weapons_skins
--genesis weapons w/o a faction
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Plasma Rifle', (SELECT id FROM weapon_models WHERE label = 'Plasma Rifle'), (SELECT weapon_type FROM weapon_models WHERE label = 'Plasma Rifle'));
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Auto Cannon', (SELECT id FROM weapon_models WHERE label = 'Auto Cannon'), (SELECT weapon_type FROM weapon_models WHERE label = 'Auto Cannon'));
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Sniper Rifle', (SELECT id FROM weapon_models WHERE label = 'Sniper Rifle'), (SELECT weapon_type FROM weapon_models WHERE label = 'Sniper Rifle'));
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Rocket Pod', (SELECT id FROM weapon_models WHERE label = 'Rocket Pod'), (SELECT weapon_type FROM weapon_models WHERE label = 'Rocket Pod'));
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Sword', (SELECT id FROM weapon_models WHERE label = 'Sword'), (SELECT weapon_type FROM weapon_models WHERE label = 'Sword'));
INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Laser Sword', (SELECT id FROM weapon_models WHERE label = 'Laser Sword'), (SELECT weapon_type FROM weapon_models WHERE label = 'Laser Sword'));

DO $$
    DECLARE weapon_model weapon_models%rowtype;
    BEGIN
        FOR weapon_model in SELECT * FROM weapon_models WHERE faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics')
        LOOP
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('BC Default', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Archon Miltech', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Blue Camo', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Police', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Gold', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Crystal', weapon_model.id, weapon_model.weapon_type);
        END LOOP;
    END;
$$;

DO $$
    DECLARE weapon_model weapon_models%rowtype;
    BEGIN
        FOR weapon_model in SELECT * FROM weapon_models WHERE faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries')
        LOOP
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('ZHI Default', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Warsui', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('White Camo', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Ninja', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Neon', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Gold', weapon_model.id, weapon_model.weapon_type);
        END LOOP;
    END;
$$;

DO $$
    DECLARE weapon_model weapon_models%rowtype;
    BEGIN
        FOR weapon_model in SELECT * FROM weapon_models WHERE faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation')
        LOOP
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('RMOMC Default', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Pyrotronics', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Red Camo', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Mining', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Molten', weapon_model.id, weapon_model.weapon_type);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id, weapon_type) VALUES ('Gold', weapon_model.id, weapon_model.weapon_type);
        END LOOP;
    END;
$$;

-- each weapon model set default skin
DO $$
    DECLARE weapon_model weapon_models%rowtype;
    BEGIN
        FOR weapon_model in SELECT * FROM weapon_models
        LOOP
            CASE
                WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries') THEN
                    UPDATE weapon_models SET default_skin_id = (SELECT id FROM blueprint_weapon_skin WHERE weapon_model_id = weapon_model.id AND label = 'ZHI Default') WHERE id = weapon_model.id;
                WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics') THEN
                    UPDATE weapon_models SET default_skin_id = (SELECT id FROM blueprint_weapon_skin WHERE weapon_model_id = weapon_model.id AND label = 'BC Default') WHERE id = weapon_model.id;
                WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation') THEN
                    UPDATE weapon_models SET default_skin_id = (SELECT id FROM blueprint_weapon_skin WHERE weapon_model_id = weapon_model.id AND label = 'RMOMC Default') WHERE id = weapon_model.id;
                WHEN weapon_model.faction_id IS NULL THEN
                    UPDATE weapon_models SET default_skin_id = (SELECT id FROM blueprint_weapon_skin WHERE weapon_model_id = weapon_model.id AND label = weapon_model.label) WHERE id = weapon_model.id;
            END CASE;
        END LOOP;
    END;
$$;

ALTER TABLE weapon_models ALTER COLUMN default_skin_id SET NOT NULL;

ALTER TABLE blueprint_weapons
    ADD COLUMN weapon_model_id UUID REFERENCES weapon_models(id);

-- update existing blueprint_weapons (factionless)
DO $$
    DECLARE blueprint_weapon blueprint_weapons%rowtype;
    BEGIN
        FOR blueprint_weapon in SELECT * FROM blueprint_weapons
        LOOP
            UPDATE blueprint_weapons SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = blueprint_weapon.label AND faction_id IS NULL) WHERE label = blueprint_weapon.label;
        END LOOP;
    END;
$$;

ALTER TABLE blueprint_weapons ALTER COLUMN weapon_model_id SET NOT NULL;

DO $$
    DECLARE weapon_model weapon_models%rowtype;
    BEGIN
        FOR weapon_model in SELECT * FROM weapon_models WHERE faction_id IS NOT NULL
        LOOP
            INSERT INTO blueprint_weapons (brand_id, label, slug, damage, weapon_type, is_melee, weapon_model_id) VALUES (
                CASE
                    WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics') THEN
                        (SELECT id FROM brands WHERE label = 'Archon Miltech')
                    WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries') THEN
                        (SELECT id FROM brands WHERE label = 'Warsui')
                    WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation') THEN
                        (SELECT id FROM brands WHERE label = 'Pyrotronics')
                END,
                concat((SELECT label FROM brands WHERE id = weapon_model.brand_id), ' ', weapon_model.label),
                lower(concat(replace((SELECT label FROM brands WHERE id = weapon_model.brand_id), ' ', '_'), '_', replace(weapon_model.label, ' ', '_'))),
                0,
                weapon_model.weapon_type,
                CASE
                    WHEN weapon_model.weapon_type = 'Sword' THEN true
                    ELSE false
                END,
                weapon_model.id
            );
        END LOOP;
    END;
$$;

ALTER TABLE weapons
    ADD COLUMN weapon_model_id UUID,
    ADD COLUMN equipped_weapon_skin_id UUID;

--updating existing weapons
UPDATE weapons SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Plasma Rifle' AND brand_id IS NULL) WHERE label = 'Plasma Rifle' AND brand_id IS NULL;
UPDATE weapons SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Auto Cannon' AND brand_id IS NULL) WHERE label = 'Auto Cannon' AND brand_id IS NULL;
UPDATE weapons SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Sniper Rifle' AND brand_id IS NULL) WHERE label = 'Sniper Rifle' AND brand_id IS NULL;
UPDATE weapons SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Sword' AND brand_id IS NULL) WHERE label = 'Sword' AND brand_id IS NULL;
UPDATE weapons SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Laser Sword' AND brand_id IS NULL) WHERE label = 'Laser Sword' AND brand_id IS NULL;
UPDATE weapons SET weapon_model_id = (SELECT id FROM weapon_models WHERE label = 'Rocket Pod' AND brand_id IS NULL) WHERE label = 'Rocket Pod' AND brand_id IS NULL;

--!!: is adding this constraint afterwards ok?
ALTER TABLE weapons
    ADD CONSTRAINT fk_weapon_models FOREIGN KEY (weapon_model_id) REFERENCES weapon_models(id);

-- MECHS

ALTER TABLE mech_model RENAME TO mech_models;

ALTER TABLE mech_models
    ADD COLUMN brand_id UUID,
    ADD COLUMN faction_id UUID,
    ADD COLUMN mech_type MECH_TYPE;

UPDATE mech_models SET mech_type = 'HUMANOID';
UPDATE mech_models SET faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics') WHERE label = 'Law Enforcer X-1000';
UPDATE mech_models SET faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries') WHERE label = 'Tenshi Mk1';
UPDATE mech_models SET faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation') WHERE label = 'Olympus Mons LY07';

--!!: labels to be renamed
INSERT INTO mech_models (label, mech_type, brand_id, faction_id) VALUES ('BC Humanoid', 'HUMANOID', (SELECT id FROM brands WHERE label = 'Daison Avionics'), (SELECT id FROM factions WHERE label = 'Boston Cybernetics'));
INSERT INTO mech_models (label, mech_type, brand_id, faction_id) VALUES ('BC Platform', 'PLATFORM', (SELECT id FROM brands WHERE label = 'Daison Avionics'), (SELECT id FROM factions WHERE label = 'Boston Cybernetics'));

INSERT INTO mech_models (label, mech_type, brand_id, faction_id) VALUES ('ZHI Humanoid', 'HUMANOID', (SELECT id FROM brands WHERE label = 'x3 Wartech'), (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'));
INSERT INTO mech_models (label, mech_type, brand_id, faction_id) VALUES ('ZHI Platform', 'PLATFORM', (SELECT id FROM brands WHERE label = 'x3 Wartech'), (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'));

INSERT INTO mech_models (label, mech_type, brand_id, faction_id) VALUES ('RMOMC Humanoid', 'HUMANOID', (SELECT id FROM brands WHERE label = 'Unified Martian Corporation'), (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'));
INSERT INTO mech_models (label, mech_type, brand_id, faction_id) VALUES ('RMOMC Platform', 'PLATFORM', (SELECT id FROM brands WHERE label = 'Unified Martian Corporation'), (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'));

ALTER TABLE mech_models
    ADD CONSTRAINT fk_brands FOREIGN KEY (brand_id) REFERENCES brands(id),
    ADD CONSTRAINT fk_factions FOREIGN KEY (faction_id) REFERENCES factions(id);

ALTER TABLE blueprint_mech_skin
    ADD COLUMN mech_type MECH_TYPE;

UPDATE blueprint_mech_skin SET MECH_TYPE = 'HUMANOID';

-- for each mech_model, insert 6 skins
DO $$
    DECLARE mech_model mech_models%rowtype;
    BEGIN
        FOR mech_model in SELECT * FROM mech_models WHERE brand_id IS NOT NULL
            LOOP
            CASE
                WHEN mech_model.faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics') THEN
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'BC Default', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Daison Avionics', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Blue Camo', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Police', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Gold', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Crystal', mech_model.mech_type);
                WHEN mech_model.faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries') THEN
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'ZHI Default', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'x3 Wartech', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'White Camo', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Ninja', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Neon', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Gold', mech_model.mech_type);
                WHEN mech_model.faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation') THEN
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'RMOMC Default', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Unified Martian Corporation', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Red Camo', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Mining', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Molten', mech_model.mech_type);
                    INSERT INTO blueprint_mech_skin (mech_model, label, mech_type) VALUES (mech_model.id, 'Gold', mech_model.mech_type);
            END CASE;
        END LOOP;
    END;
$$;

-- set default skins for mech model

DO $$
    DECLARE model mech_models%rowtype;
    BEGIN
        FOR model in SELECT * FROM mech_models WHERE default_chassis_skin_id IS NULL
        LOOP
            CASE
                WHEN model.faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries') THEN
                    UPDATE mech_models SET default_chassis_skin_id = (SELECT id FROM blueprint_mech_skin WHERE mech_model = model.id AND label = 'ZHI Default') WHERE id = model.id;
                WHEN model.faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics') THEN
                    UPDATE mech_models SET default_chassis_skin_id = (SELECT id FROM blueprint_mech_skin WHERE mech_model = model.id AND label = 'BC Default') WHERE id = model.id;
                WHEN model.faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation') THEN
                    UPDATE mech_models SET default_chassis_skin_id = (SELECT id FROM blueprint_mech_skin WHERE mech_model = model.id AND label = 'RMOMC Default') WHERE id = model.id;
                WHEN model.faction_id IS NULL THEN
                    UPDATE mech_models SET default_chassis_skin_id = (SELECT id FROM blueprint_mech_skin WHERE mech_model = model.id AND label = model.label) WHERE id = model.id;
            END CASE;
        END LOOP;
    END;
$$;

ALTER TABLE mech_models ALTER COLUMN default_chassis_skin_id SET NOT NULL;

ALTER TABLE blueprint_mechs DROP COLUMN skin;

--!!: probably can use a with statement here?
DO $$
    DECLARE mech_model mech_models%rowtype;
    BEGIN
        FOR mech_model in SELECT * FROM mech_models WHERE brand_id IS NOT NULL
        LOOP
            INSERT INTO blueprint_mechs (brand_id, label, slug, weapon_hardpoints, utility_slots, speed, max_hitpoints, model_id, power_core_size) VALUES (
                mech_model.brand_id,
                concat((SELECT label FROM brands WHERE id = mech_model.brand_id), ' ', mech_model.label),
                lower(concat(replace((SELECT label FROM brands WHERE id = mech_model.brand_id), ' ', '_'), '_', replace(mech_model.label, ' ', '_'))),
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
                END
            );
        END LOOP;
    END;
$$;


INSERT INTO blueprint_power_cores (collection, label, size, capacity, max_draw_rate, recharge_rate, armour, max_hitpoints) VALUES ('supremacy-general', 'Medium Energy Core', 'MEDIUM', 1500, 150, 100, 0, 1500);


-- seeding mystery crates
-- looping over each type of mystery crate type for x amount of crates for each faction. can do 1 big loop if all crate types have the same amount
DO $$
    BEGIN
    --should be >= 200 and end in -00 for all math to math plz... (2 mech types, each with own percentage of skins allocations)
        FOR COUNT IN 1..500 LOOP
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Red Mountain Offworld Mining Corporation'), 'Red Mountain War Machine Mystery Crate');
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Zaibatsu Heavy Industries'), 'Zaibatsu Nexus War Machine Mystery Crate');
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Boston Cybernetics'), 'Boston Cybernetics Nexus War Machine Mystery Crate');
        END LOOP;
    END;
$$;

DO $$
    BEGIN
    --should be >= 800 and end in 00 for all math to math plz... (8 types of weapons/faction each with own percentage of skin allocations) assuming all weapons will be of equal rarity and all weapons get all skins- easy to change skin percentages
        FOR COUNT IN 1..800 LOOP
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Red Mountain Offworld Mining Corporation'), 'Red Mountain Nexus Weapon Mystery Crate');
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Zaibatsu Heavy Industries'), 'Zaibatsu Nexus Weapon Mystery Crate');
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Boston Cybernetics'), 'Boston Cybernetics Nexus Weapon Mystery Crate');
        END LOOP;
    END;
$$;



-- seeding blueprints
DO $$
    DECLARE faction factions%rowtype;
    DECLARE mechCrate mystery_crate%rowtype;
    DECLARE weaponCrate mystery_crate%rowtype;
    DECLARE i integer;
    DECLARE mechCrateLen integer;
    DECLARE weaponCrateLen integer;

BEGIN
    --for each faction loop over the mystery crates of specified faction
        FOR faction in SELECT * FROM factions
        LOOP
            i := 1;
            mechCrateLen := (SELECT COUNT(*) FROM mystery_crate WHERE faction_id = faction.id AND type = 'MECH');
            -- for 80% of the Mechs, insert a mech object from the appropriate brand's bipedal mechs and a fitted power core
            FOR mechCrate in SELECT * FROM mystery_crate WHERE faction_id = faction.id AND type = 'MECH'
            LOOP
                CASE
                    WHEN i <=(mechCrateLen*.80) THEN
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (mechCrate.id, 'MECH', (SELECT id FROM blueprint_mechs WHERE blueprint_mechs.power_core_size = 'SMALL' AND
                        blueprint_mechs.brand_id =
                            CASE
                                WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM brands WHERE label = 'Daison Avionics')
                                WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM brands WHERE label = 'x3 Wartech')
                                WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM brands WHERE label = 'Unified Martian Corporation')
                            END
                        ));
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (mechCrate.id, 'POWER_CORE', (SELECT id FROM blueprint_power_cores c WHERE c.label = 'Standard Energy Core'));
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (mechCrate.id, 'MECH_SKIN', (SELECT id FROM blueprint_mech_skin WHERE mech_type = 'HUMANOID' AND
                        blueprint_mech_skin.mech_model =
                            CASE
                                WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM mech_models WHERE label = 'BC Humanoid')
                                WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM mech_models WHERE label = 'ZHI Humanoid')
                                WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM mech_models WHERE label = 'RMOMC Humanoid')
                            END
                        AND label =
                            CASE
                                --30% default skin
                                WHEN i <= (.30*(mechCrateLen*.80)) THEN
                                    CASE
                                        WHEN faction.label = 'Boston Cybernetics' THEN 'BC Default'
                                        WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'ZHI Default'
                                        WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'RMOMC Default'
                                    END
                                --25% for manufacturer's skin
                                WHEN i > (.30*(mechCrateLen*.80)) AND i <= (.55*(mechCrateLen*.80)) THEN
                                    CASE
                                        WHEN faction.label = 'Boston Cybernetics' THEN 'Daison Avionics'
                                        WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'x3 Wartech'
                                        WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Unified Martian Corporation'
                                    END
                                --17% for camo
                                WHEN i > (.55*(mechCrateLen*.80)) AND i <= (.72*(mechCrateLen*.80)) THEN
                                    CASE
                                        WHEN faction.label = 'Boston Cybernetics' THEN 'Blue Camo'
                                        WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'White Camo'
                                        WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Red Camo'
                                    END
                                --17% for theme
                                WHEN i > (.72*(mechCrateLen*.80)) AND i <= (.89*(mechCrateLen*.80)) THEN
                                    CASE
                                        WHEN faction.label = 'Boston Cybernetics' THEN 'Police'
                                        WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'Ninja'
                                        WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Mining'
                                    END
                                --10% for Gold
                                WHEN i > (.89*(mechCrateLen*.80)) AND i <= (.99*(mechCrateLen*.80)) THEN 'Gold'
                                --1% for rare color
                                WHEN i > (.99*(mechCrateLen*.80)) AND i <= (1*(mechCrateLen*.80)) THEN
                                    CASE
                                        WHEN faction.label = 'Boston Cybernetics' THEN 'Crystal'
                                        WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'Neon'
                                        WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Molten'
                                    END
                            END
                        ));
                        i := i + 1;
                    -- for other half of the Mechs, insert a mech object from the appropriate brand's platform mechs and a fitted power core
                    ELSE
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (mechCrate.id, 'MECH', (SELECT id FROM blueprint_mechs WHERE blueprint_mechs.power_core_size = 'MEDIUM' AND
                            blueprint_mechs.brand_id =
                                CASE
                                    WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM brands WHERE label = 'Daison Avionics')
                                    WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM brands WHERE label = 'x3 Wartech')
                                    WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM brands WHERE label = 'Unified Martian Corporation')
                                END
                            ));
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (mechCrate.id, 'POWER_CORE', (SELECT id FROM blueprint_power_cores c WHERE c.size = 'MEDIUM'));
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (mechCrate.id, 'MECH_SKIN', (SELECT id FROM blueprint_mech_skin WHERE mech_type = 'PLATFORM' AND
                        blueprint_mech_skin.mech_model =
                            CASE
                                WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM mech_models WHERE label = 'BC Platform')
                                WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM mech_models WHERE label = 'ZHI Platform')
                                WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM mech_models WHERE label = 'RMOMC Platform')
                            END
                            AND label =
                            CASE
                                --30% default skin
                                WHEN i <= ((.30*(mechCrateLen*.20)) + (mechCrateLen*.80)) THEN
                                    CASE
                                        WHEN faction.label = 'Boston Cybernetics' THEN 'BC Default'
                                        WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'ZHI Default'
                                        WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'RMOMC Default'
                                    END
                                --25% for manufacturer's skin
                                WHEN i > ((.30*(mechCrateLen*.20)) + (mechCrateLen*.80)) AND i <= ((.55*(mechCrateLen*.20)) + (mechCrateLen*.80)) THEN
                                    CASE
                                        WHEN faction.label = 'Boston Cybernetics' THEN 'Daison Avionics'
                                        WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'x3 Wartech'
                                        WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Unified Martian Corporation'
                                    END
                                --17% for camo
                                WHEN i > ((.55*(mechCrateLen*.20)) + (mechCrateLen*.80)) AND i <= ((.72*(mechCrateLen*.20)) + (mechCrateLen*.80)) THEN
                                    CASE
                                        WHEN faction.label = 'Boston Cybernetics' THEN 'Blue Camo'
                                        WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'White Camo'
                                        WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Red Camo'
                                    END
                                --17% for theme
                                WHEN i > ((.72*(mechCrateLen*.20)) + (mechCrateLen*.80)) AND i <= ((.89*(mechCrateLen*.20)) + (mechCrateLen*.80)) THEN
                                    CASE
                                        WHEN faction.label = 'Boston Cybernetics' THEN 'Police'
                                        WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'Ninja'
                                        WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Mining'
                                    END
                                --10% for Gold
                                WHEN i > ((.89*(mechCrateLen*.20)) + (mechCrateLen*.80)) AND i <= ((.99*(mechCrateLen*.20)) + (mechCrateLen*.80)) THEN 'Gold'
                                --1% for rare color
                                WHEN i > ((.99*(mechCrateLen*.20)) + (mechCrateLen*.80)) AND i <= ((1*(mechCrateLen*.20)) + (mechCrateLen*.80)) THEN
                                    CASE
                                        WHEN faction.label = 'Boston Cybernetics' THEN 'Crystal'
                                        WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'Neon'
                                        WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Molten'
                                    END
                                END

                        ));
                            i = i + 1;
                END CASE;
            END LOOP;

            -- for weapons crates of each faction, insert weapon blueprint. ** ALL WEAPONS CRATES ARE EQUAL
            i := 1;
            weaponCrateLen := (SELECT COUNT(*) FROM mystery_crate WHERE faction_id = faction.id AND type = 'WEAPON')/8; --length of crate rows to be allocated to 1 type of weapon
            FOR weaponCrate in SELECT * FROM mystery_crate WHERE faction_id = faction.id AND type = 'WEAPON'
                LOOP
                    --flak: all factions
                    CASE
                    WHEN i <= mechCrateLen THEN
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (weaponCrate.id, 'WEAPON', (SELECT id FROM blueprint_weapons WHERE weapon_type = 'Flak' AND
                        brand_id =
                        CASE
                            WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM brands WHERE label = 'Archon Miltech')
                            WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM brands WHERE label = 'Warsui')
                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        END
                        ));
--                         INSERT INTO mystery_crate_blueprints (id, mystery_crate_id, blueprint_type, blueprint_id) VALUES (weaponCrate.id, 'WEAPON_SKIN', (SELECT id FROM blueprint_weapon_skins WHERE weapon_type = 'Flak' AND
--                             blueprint_weapon_skins.weapon_model_id =
--                             CASE
--                                 WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM weapon_models WHERE label = 'BC Humanoid')
--                                 WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM weapon_models WHERE label = 'ZHI Humanoid')
--                                 WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM weapon_models WHERE label = 'RMOMC Humanoid')
--                                 END
--                         ))
                        i := i + 1;
                    --machine gun: all factions
                    WHEN i > mechCrateLen AND i <= (2 * mechCrateLen) THEN
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (weaponCrate.id, 'WEAPON', (SELECT id FROM blueprint_weapons WHERE weapon_type = 'Machine Gun' AND
                        brand_id =
                        CASE
                            WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM brands WHERE label = 'Archon Miltech')
                            WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM brands WHERE label = 'Warsui')
                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        END
                        ));
                        i := i + 1;
                    --flamethrower: all factions
                    WHEN i > (2 * mechCrateLen) AND i <= (3 * mechCrateLen) THEN
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (weaponCrate.id, 'WEAPON', (SELECT id FROM blueprint_weapons WHERE weapon_type = 'Flamethrower' AND
                        brand_id =
                        CASE
                            WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM brands WHERE label = 'Archon Miltech')
                            WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM brands WHERE label = 'Warsui')
                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        END
                        ));
                        i := i + 1;
                    --missile launcher: all factions
                    WHEN i > (3 * mechCrateLen) AND i <= (4 * mechCrateLen) THEN
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (weaponCrate.id, 'WEAPON', (SELECT id FROM blueprint_weapons WHERE weapon_type = 'Missile Launcher' AND
                        brand_id =
                        CASE
                            WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM brands WHERE label = 'Archon Miltech')
                            WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM brands WHERE label = 'Warsui')
                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        END
                        ));
                        i := i + 1;
                    --Laser beam: all factions
                    WHEN i > (4 * mechCrateLen) AND i <= (5 * mechCrateLen) THEN
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (weaponCrate.id, 'WEAPON', (SELECT id FROM blueprint_weapons WHERE weapon_type = 'Laser Beam' AND
                        brand_id =
                        CASE
                            WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM brands WHERE label = 'Archon Miltech')
                            WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM brands WHERE label = 'Warsui')
                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        END
                        ));
                        i := i + 1;

                    --Minigun: BC and RM OR Plasma Gun for ZHI
                    WHEN i > (5 * mechCrateLen) AND i <= (6 * mechCrateLen) THEN
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (weaponCrate.id, 'WEAPON', (SELECT id FROM blueprint_weapons WHERE
                        weapon_type =
                        CASE
                            WHEN faction.label = 'Boston Cybernetics' OR faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Minigun'::WEAPON_TYPE
                            WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'Plasma Gun'::WEAPON_TYPE
                        END
                        AND
                        brand_id =
                        CASE
                            WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM brands WHERE label = 'Archon Miltech')
                            WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM brands WHERE label = 'Warsui')
                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        END
                        ));
                        i := i + 1;

                    --Cannon: ZHI and RM OR Plasma Gun for BC
                    WHEN i > (6 * mechCrateLen) AND i <= (7 * mechCrateLen) THEN
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (weaponCrate.id, 'WEAPON', (SELECT id FROM blueprint_weapons WHERE
                        weapon_type =
                        CASE
                            WHEN faction.label = 'Zaibatsu Heavy Industries' OR faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Cannon'::WEAPON_TYPE
                            WHEN faction.label = 'Boston Cybernetics' THEN 'Plasma Gun'::WEAPON_TYPE
                        END
                        AND
                        brand_id =
                        CASE
                            WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM brands WHERE label = 'Archon Miltech')
                            WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM brands WHERE label = 'Warsui')
                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        END
                        ));
                        i := i + 1;
                    --BFG, Grenade Launcher or Lightning Gun dependent on faction
                    WHEN i > (7 * mechCrateLen) AND i <= (8 * mechCrateLen) THEN
                        INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (weaponCrate.id, 'WEAPON', (SELECT id FROM blueprint_weapons WHERE
                        weapon_type =
                        CASE
                            WHEN faction.label = 'Boston Cybernetics' THEN 'BFG'::WEAPON_TYPE
                            WHEN faction.label = 'Zaibatsu Heavy Industries' THEN 'Lightning Gun'::WEAPON_TYPE
                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN 'Grenade Launcher'::WEAPON_TYPE
                        END
                        AND
                        brand_id =
                        CASE
                            WHEN faction.label = 'Boston Cybernetics' THEN (SELECT id FROM brands WHERE label = 'Archon Miltech')
                            WHEN faction.label = 'Zaibatsu Heavy Industries' THEN (SELECT id FROM brands WHERE label = 'Warsui')
                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation' THEN (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        END
                        ));
                        i := i + 1;
                    END CASE;
            END LOOP;

        END LOOP;
    END;
$$;

ALTER TABLE weapon_models DROP COLUMN faction_id;
ALTER TABLE mech_models DROP COLUMN faction_id;

-- seeding skins... skins can be variable for weapons
-- mech skins

--seeding storefront
-- for each faction, seed each type of crate and find how much are for sale
DO $$
    DECLARE faction factions%rowtype;
    BEGIN
    FOR faction in SELECT * FROM factions
        LOOP
            INSERT INTO storefront_mystery_crates (mystery_crate_type, amount, faction_id, price) VALUES ('MECH', (SELECT COUNT(*) FROM mystery_crate WHERE type='MECH' AND faction_id=faction.id), faction.id, 500);
            INSERT INTO storefront_mystery_crates (mystery_crate_type, amount, faction_id, price) VALUES ('WEAPON', (SELECT COUNT(*) FROM mystery_crate WHERE type='MECH' AND faction_id=faction.id), faction.id, 500);
        END LOOP;
    END;
$$;

