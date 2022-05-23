DROP TYPE IF EXISTS CRATE_TYPE_ENUM;
CREATE TYPE CRATE_TYPE_ENUM AS ENUM ('MECH', 'WEAPON', 'UTILITY');

CREATE TABLE storefront_mystery_crates
(
    id                 UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    mystery_crate_type CRATE_TYPE_ENUM NOT NULL,
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
    faction_id      UUID        NOT NULL,
    label           TEXT        NOT NULL,
    weapon_type     TEXT        NOT NULL,
    default_skin_id UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE blueprint_weapon_skin
    ADD COLUMN collection TEXT NOT NULL DEFAULT 'supremacy-general',
    ADD COLUMN weapon_model_id UUID NOT NULL REFERENCES weapon_models(id),
    ADD COLUMN image_url TEXT, -- make this not null later
    ADD COLUMN animation_url TEXT,
    ADD COLUMN card_animation_url TEXT,
    ADD COLUMN large_image_url TEXT,
    ADD COLUMN avatar_url TEXT,
    DROP COLUMN weapon_type;

ALTER TABLE weapon_skin
    ADD COLUMN weapon_model_id UUID NOT NULL REFERENCES weapon_models (id),
    ADD COLUMN image_url TEXT, -- make this not null later
    ADD COLUMN animation_url TEXT,
    ADD COLUMN card_animation_url TEXT,
    ADD COLUMN large_image_url TEXT,
    ADD COLUMN avatar_url TEXT;

-- inserting brands
INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Daison Avionics');
INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Archon Miltech');

INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'x3 Wartech');
INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Warsui');

INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Unified Martian Corporation');
INSERT INTO brands (faction_id, label) VALUES ((SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Pyrotronics');

--Insert into weapon models
-- looping through common weapons shared between factions
DO $$
DECLARE faction factions%rowtype;
    BEGIN
        FOR faction in SELECT * FROM factions
        LOOP
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Flak', faction.id, 'Shotgun');
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Machine Gun', faction.id, 'Machine Gun');
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Flamethrower', faction.id, 'Flamethrower');
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Missile Launcher', faction.id, 'Missile Launcher');
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Laser Gun', faction.id, 'Laser Gun');
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Plasma Rifle', faction.id, 'Plasma Rifle');
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Auto Cannon', faction.id, 'Cannon');
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Sniper Rifle', faction.id, 'Sniper Rifle');
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Rocket Pod', faction.id, 'Rocket');
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Sword', faction.id, 'Sword');
            INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Laser Sword', faction.id, 'Sword');
        END LOOP;
    END;
$$;

-- Inserting specific weapons for each faction
INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Plasma Gun', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Plasma Gun');
INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Mini Gun', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Mini Gun');
INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('BFG', (SELECT id FROM factions WHERE label = 'Boston Cybernetics'), 'Gun');


INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Plasma Gun', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Plasma Gun');
INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Cannon', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Cannon');
INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Lightning Gun', (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries'), 'Lightning Gun');


INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Mini Gun', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Mini Gun');
INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Cannon', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Cannon');
INSERT INTO weapon_models (label, faction_id, weapon_type) VALUES ('Grenade Launcher', (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation'), 'Grenade Launcher');

-- seed blueprint_weapons_skins
DO $$
    DECLARE weapon_model weapon_models%rowtype;
    BEGIN
        FOR weapon_model in SELECT * FROM weapon_models WHERE faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics')
        LOOP
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Archon Miltech', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Blue Camo', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Police', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Gold', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Crystal', weapon_model.id);
        END LOOP;
    END;
$$;

DO $$
    DECLARE weapon_model weapon_models%rowtype;
    BEGIN
        FOR weapon_model in SELECT * FROM weapon_models WHERE faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries')
        LOOP
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Warsui', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('White Camo', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Ninja', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Neon', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Gold', weapon_model.id);
        END LOOP;
    END;
$$;

DO $$
    DECLARE weapon_model weapon_models%rowtype;
    BEGIN
        FOR weapon_model in SELECT * FROM weapon_models WHERE faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation')
        LOOP
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Pyrotronics', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Red Camo', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Mining', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Molten', weapon_model.id);
            INSERT INTO blueprint_weapon_skin (label, weapon_model_id) VALUES ('Gold', weapon_model.id);
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
                    UPDATE weapon_models SET default_skin_id = (SELECT id FROM blueprint_weapon_skin WHERE weapon_model_id = weapon_model.id AND label = 'Warsui') WHERE id = weapon_model.id;
                WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics') THEN
                    UPDATE weapon_models SET default_skin_id = (SELECT id FROM blueprint_weapon_skin WHERE weapon_model_id = weapon_model.id AND label = 'Archon Miltech') WHERE id = weapon_model.id;
                WHEN weapon_model.faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation') THEN
                    UPDATE weapon_models SET default_skin_id = (SELECT id FROM blueprint_weapon_skin WHERE weapon_model_id = weapon_model.id AND label = 'Pyrotronics') WHERE id = weapon_model.id;
            END CASE;
        END LOOP;
    END;
$$;

ALTER TABLE weapon_models ALTER COLUMN default_skin_id SET NOT NULL;

-- ALTER TABLE blueprint_weapons
--     ADD COLUMN weapon_model_id UUID NOT NULL REFERENCES weapon_models(id);
--
-- ALTER TABLE weapons
--     ADD COLUMN weapon_model_id UUID NOT NULL REFERENCES weapon_models(id),
--     ADD COLUMN weapon_skin_id UUID NOT NULL REFERENCES weapon_skin(id);


INSERT INTO blueprint_power_cores (collection, label, size, capacity, max_draw_rate, recharge_rate, armour, max_hitpoints) VALUES ('supremacy-general', 'Small Energy Core', 'SMALL', 750, 75, 75, 0, 750);


-- seeding mystery crates
-- looping over each type of mystery crate type for x amount of crates for each faction. can do 1 big loop if all crate types have the same amount
DO $$
    BEGIN
        FOR COUNT IN 1..10 LOOP
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Red Mountain Offworld Mining Corporation'), 'RMOMC Mech Mystery Crate');
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Zaibatsu Heavy Industries'), 'ZHI Mech Mystery Crate');
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Boston Cybernetics'), 'BC Mech Mystery Crate');
        END LOOP;
    END;
$$;

DO $$
    BEGIN
        FOR COUNT IN 1..10 LOOP
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Red Mountain Offworld Mining Corporation'), 'RMOMC Weapon Mystery Crate');
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Zaibatsu Heavy Industries'), 'ZHI Weapon Mystery Crate');
            INSERT INTO mystery_crate (type, faction_id, label) VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Boston Cybernetics'), 'BC Weapon Mystery Crate');
        END LOOP;
    END;
$$;



-- seeding blueprints
-- mechs: blueprint_mechs only have brand_id which joins on brand to factions
-- DO $$
--     BEGIN
--     --for each faction loop over the mystery crates of specified faction
--         FOR faction in SELECT * FROM factions
--             --for crates of type mech, loop
--             FOR row in SELECT FROM mystery_crate WHERE faction_id = faction.id AND type = 'MECH'
--                 LOOP
--                 DECLARE i float8 := 1;
--                     DECLARE brandID uuid := (SELECT id FROM brands WHERE faction_id = faction.id)
--                     -- for half of the Mechs, insert a mech object from the appropriate brand's bipedal mechs and a fitted power core
--                     WHILE i <= 5 LOOP
--                         INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (row.id, 'MECH', (SELECT id FROM blueprint_mechs WHERE blueprint_mechs.brand_id = brandID AND blueprint_mechs.power_core_size = 'SMALL'));
--                         INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (row.id, 'POWER_CORE', (SELECT id FROM blueprint_power_cores c WHERE c.size = 'SMALL'));
--                         i = i + 1
--                     END LOOP;
--                         -- for other half of the Mechs, insert a mech object from the appropriate brand's platform mechs and a fitted power core
--                     WHILE i>5 LOOP
--                         INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (row.id, 'MECH', (SELECT blueprint_mechs.id FROM blueprint_mechs LEFT JOIN brands WHERE blueprint_mechs.brand_id = brands.id AND brands.id = faction.id AND blueprint_mechs.collection != 'supremacy-general' AND blueprint_mechs.power_core_size = 'MEDIUM'));
--                         INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (row.id, 'POWER_CORE', (SELECT id FROM blueprint_power_cores c WHERE c.size = 'MEDIUM'));
--                         i = i + 1
--                     END LOOP;
--             END LOOP;
--
--             --for crates of type weapon, loop
--             FOR row in SELECT FROM mystery_crate WHERE faction_id = faction.id AND type = 'WEAPON'
--                 LOOP
--                 DECLARE i float8 := 1;
--                 -- for half of the Mechs, insert a mech object from the appropriate brand's bipedal mechs and a fitted power core
--                 WHILE i <= 3 LOOP
--                     INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id) VALUES (row.id, 'WEAPON', (SELECT blueprint_weapons.id FROM blueprint_weapons LEFT JOIN brands WHERE blueprint_weapons.brand_id = brands.id AND brands.id = faction.id AND ));
--                         i = i + 1
--                 END LOOP;
--             END LOOP;
--         END LOOP;
--     END;
-- $$;

-- weapons


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

