DROP TYPE IF EXISTS CRATE_TYPE_ENUM;
CREATE TYPE CRATE_TYPE_ENUM AS ENUM ('MECH', 'WEAPON', 'UTILITY');

CREATE TABLE storefront_mystery_crates
(
    id                 UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    mystery_crate_type CRATE_TYPE_ENUM NOT NULL,
    price              numeric(28)     NOT NULL,
    amount             INT             NOT NULL,
    amount_sold        INT             NOT NULL,
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

-- seeding mystery crates
-- loop over these for x amount of crates?
do $$
BEGIN
FOR COUNT IN 1..10 loop
INSERT INTO mystery_crate (type, faction_id, label) VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Red Mountain Offworld Mining Corporation'), 'RMOMC Mech Mystery Crate');
INSERT INTO mystery_crate (type, faction_id, label) VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Zaibatsu Heavy Industries'), 'ZHI Mech Mystery Crate');
INSERT INTO mystery_crate (type, faction_id, label) VALUES ('MECH', (SELECT id FROM factions f WHERE f.label = 'Boston Cybernetics'), 'BC Mech Mystery Crate');
END LOOP;
END;
$$;

do $$
BEGIN
FOR COUNT IN 1..10 loop
INSERT INTO mystery_crate (type, faction_id, label) VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Red Mountain Offworld Mining Corporation'), 'RMOMC Weapon Mystery Crate');
INSERT INTO mystery_crate (type, faction_id, label) VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Zaibatsu Heavy Industries'), 'ZHI Weapon Mystery Crate');
INSERT INTO mystery_crate (type, faction_id, label) VALUES ('WEAPON', (SELECT id FROM factions f WHERE f.label = 'Boston Cybernetics'), 'BC Weapon Mystery Crate');
END LOOP;
END;
$$;

do $$
BEGIN
FOR COUNT IN 1..10 loop
INSERT INTO mystery_crate (type, faction_id, label) VALUES ('UTILITY', (SELECT id FROM factions f WHERE f.label = 'Red Mountain Offworld Mining Corporation'), 'RMOMC Utility Mystery Crate');
INSERT INTO mystery_crate (type, faction_id, label) VALUES ('UTILITY', (SELECT id FROM factions f WHERE f.label = 'Zaibatsu Heavy Industries'), 'ZHI Utility Mystery Crate');
INSERT INTO mystery_crate (type, faction_id, label) VALUES ('UTILITY', (SELECT id FROM factions f WHERE f.label = 'Boston Cybernetics'), 'BC Utility Mystery Crate');
END LOOP;
END;
$$;

-- seeding blue prints

--seeding storefront
--
