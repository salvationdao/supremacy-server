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
