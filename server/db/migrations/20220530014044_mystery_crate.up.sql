ALTER TYPE ITEM_TYPE ADD VALUE 'mystery_crate';

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

ALTER TABLE weapons
    ADD COLUMN equipped_weapon_skin_id UUID REFERENCES weapon_skin (id);

--
-- DO
-- $$
--     DECLARE
--         weapon_model WEAPON_MODELS%ROWTYPE;
--     BEGIN
--         FOR weapon_model IN SELECT * FROM weapon_models
--             LOOP
--                 INSERT INTO blueprint_weapons (brand_id, label, slug, damage, weapon_type, is_melee)
--                 VALUES ((SELECT id FROM brands WHERE id = weapon_model.brand_id),
--                         CONCAT((SELECT label FROM brands WHERE id = weapon_model.brand_id), ' ', weapon_model.label),
--                         LOWER(CONCAT(REPLACE((SELECT label FROM brands WHERE id = weapon_model.brand_id), ' ', '_'),
--                                      '_', REPLACE(weapon_model.label, ' ', '_'))),
--                         0,
--                         weapon_model.weapon_type,
--                         CASE
--                             WHEN weapon_model.weapon_type = 'Sword' THEN TRUE
--                             ELSE FALSE
--                             END);
--             END LOOP;
--     END;
-- $$;
--
-- ALTER TABLE blueprint_mechs_old
--     DROP COLUMN skin;
-- --
-- DO
-- $$
--     DECLARE
--         mech_model MECH_MODELS%ROWTYPE;
--     BEGIN
--         FOR mech_model IN SELECT * FROM blueprint_mechs WHERE brand_id IS NOT NULL
--             LOOP
--                 INSERT INTO blueprint_mechs_old (brand_id, label, slug, weapon_hardpoints, utility_slots, speed,
--                                              max_hitpoints, model_id, power_core_size)
--                 VALUES (mech_model.brand_id,
--                         CONCAT((SELECT label FROM brands WHERE id = mech_model.brand_id), ' ', mech_model.label),
--                         LOWER(CONCAT(REPLACE((SELECT label FROM brands WHERE id = mech_model.brand_id), ' ', '_'), '_',
--                                      REPLACE(mech_model.label, ' ', '_'))),
--                         CASE --hardpoints
--                             WHEN mech_model.mech_type = 'PLATFORM' THEN 5
--                             ELSE 2
--                             END,
--                         CASE -- utility slots
--                             WHEN mech_model.mech_type = 'PLATFORM' THEN 2
--                             ELSE 4
--                             END,
--                         CASE --speed
--                             WHEN mech_model.mech_type = 'PLATFORM' THEN 1000
--                             ELSE 2000
--                             END,
--                         CASE --max_hitpoints
--                             WHEN mech_model.mech_type = 'PLATFORM' THEN 3000
--                             ELSE 1500
--                             END,
--                         mech_model.id,
--                         CASE --max_hitpoints
--                             WHEN mech_model.mech_type = 'PLATFORM' THEN 'MEDIUM'
--                             ELSE 'SMALL'
--                             END);
--             END LOOP;
--     END;
-- $$;

-- seeding mystery crates
-- looping over each type of mystery crate type for x amount of crates for each faction. can do 1 big loop if all crate types have the same amount
DO
$$
    BEGIN
        FOR count IN 1..300 --change from 5000, -- dev changed to not take years
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
        FOR count IN 1..1000 -- changed from 13000, -- dev changed to not take years
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
CREATE FUNCTION insert_mech_into_crate(core_size TEXT, mechcrate_id UUID, faction RECORD) RETURNS VOID
    LANGUAGE plpgsql AS
$$
BEGIN
    INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
    VALUES (mechcrate_id, 'MECH', (SELECT mm.id
                                   FROM  blueprint_mechs mm
                                   WHERE mm.power_core_size = core_size::POWERCORE_SIZE
                                     AND mm.brand_id =
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
CREATE FUNCTION insert_mech_skin_into_crate(mechcrate_id UUID, mechtype TEXT, skinlabel TEXT, faction RECORD) RETURNS VOID
    LANGUAGE plpgsql AS
$$
BEGIN

    INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
    VALUES (mechcrate_id, 'MECH_SKIN', (SELECT id
                                        FROM blueprint_mech_skin
                                        WHERE label = skinlabel));
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
                                   THEN 'Spot Yellow'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Heavy White'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Pilbara Dust'
                               END
                       WHEN type = 'WEAPON' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Archon Gunmetal'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Verdant Warsui'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Pyro Crimson'
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
                                   THEN 'Sea Hawk'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Nullifier'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'High Caliber'
                               END
                       WHEN type = 'WEAPON' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Praetor Grunge'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Violet Ice'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Barricade Stripes'
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
                                   THEN 'Thin Blue Line'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Two Five Zero One'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Fly In Fly Out'
                               END
                       WHEN type = 'WEAPON' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Less-Than-Lethal'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Rebellion'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Shark Skin'
                               END
                       END

        --12.5% legendary
               WHEN i > ((.75 * amount_of_type) + previous_crates) AND
                    i <= ((.875 * amount_of_type) + previous_crates)
                   THEN CASE
                            WHEN type = 'MECH' THEN
                                CASE
                                    WHEN faction.label = 'Boston Cybernetics'
                                        THEN 'Bullion'
                                    WHEN faction.label = 'Zaibatsu Heavy Industries'
                                        THEN 'Mine God'
                                    WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                        THEN 'Sovereign Hill'
                                    END
                            WHEN type = 'WEAPON' THEN
                                CASE
                                    WHEN faction.label = 'Boston Cybernetics'
                                        THEN 'Unbroken Knot'
                                    WHEN faction.label = 'Zaibatsu Heavy Industries'
                                        THEN 'Cephalopod Ripple'
                                    WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                        THEN 'Martian Mess Maker'
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
                                   THEN 'Code of Chivalry'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Shadows Steal Away'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Osmium Scream'
                               END
                       WHEN type = 'WEAPON' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Ready To Quench'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Catastrophe'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Watered Steel'
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
                                   THEN 'Telling the Bees'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Synth Punk'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Promethean Gold'
                               END
--                        this will push into rare to account for no mythics for MOST weapons
                       WHEN type = 'WEAPON' THEN
                           CASE
                               WHEN faction.label = 'Boston Cybernetics'
                                   THEN 'Less-Than-Lethal'
                               WHEN faction.label = 'Zaibatsu Heavy Industries'
                                   THEN 'Rebellion'
                               WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                   THEN 'Shark Skin'
                               END
                       END
        END;
END;
$$;

DROP FUNCTION IF EXISTS insert_weapon_skin_into_crate(i INTEGER, weaponCrate_id UUID, weaponType TEXT,
                                                      amount_of_type NUMERIC,
                                                      previous_crates NUMERIC, type TEXT, skin_label TEXT,
                                                      faction RECORD);
CREATE FUNCTION insert_weapon_skin_into_crate(i INTEGER, weaponcrate_id UUID, weapontype TEXT,
                                              amount_of_type NUMERIC,
                                              previous_crates NUMERIC, type TEXT, skin_label TEXT,
                                              faction RECORD) RETURNS VOID
    LANGUAGE plpgsql AS
$$
BEGIN
    INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
    VALUES (weaponcrate_id, 'WEAPON_SKIN', (SELECT id
                                            FROM blueprint_weapon_skin
                                            WHERE label =
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
CREATE FUNCTION insert_weapon_into_crate(crate_id UUID, weapontype TEXT, faction RECORD) RETURNS VOID
    LANGUAGE plpgsql AS
$$
BEGIN
    INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
    VALUES (crate_id, 'WEAPON', (SELECT bw.id
                                 FROM blueprint_weapons bw
                                 WHERE bw.weapon_type = weapontype::WEAPON_TYPE
                                   AND bw.brand_id =
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
        DECLARE mechcrate       MYSTERY_CRATE%ROWTYPE;
        DECLARE weaponcrate     MYSTERY_CRATE%ROWTYPE;
        DECLARE i               INTEGER;
        DECLARE mechcratelen    INTEGER;
        DECLARE mech_skin_label TEXT;

    BEGIN
        --for each faction loop over the mystery crates of specified faction
        FOR faction IN SELECT * FROM factions
            LOOP
                i := 1;
                mechcratelen :=
                        (SELECT COUNT(*) FROM mystery_crate WHERE faction_id = faction.id AND type = 'MECH');
                -- for half of the Mechs, insert a mech object from the appropriate brand's bipedal mechs and a fitted power core
                FOR mechcrate IN SELECT * FROM mystery_crate WHERE faction_id = faction.id AND type = 'MECH'
                    LOOP

                        mech_skin_label := get_mech_skin_label_rarity(i, 'MECH', (.8 * mechcratelen), 0, faction);
                        CASE
                            WHEN i <= (mechcratelen * .8) -- seed humanoid mechs
                                THEN PERFORM insert_mech_into_crate('SMALL', mechcrate.id, faction);
                                     PERFORM insert_mech_skin_into_crate(mechcrate.id, 'HUMANOID',
                                                                         mech_skin_label, faction);

                                     INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                                     VALUES (mechcrate.id, 'POWER_CORE', (SELECT id
                                                                          FROM blueprint_power_cores c
                                                                          WHERE c.label = 'Standard Power Core A'));

                                     PERFORM insert_weapon_into_crate(mechcrate.id, 'Flak', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, mechcrate.id, 'Flak',
                                                                           (.8 * mechcratelen),
                                                                           0, 'MECH', mech_skin_label,
                                                                           faction);

                                     PERFORM insert_weapon_into_crate(mechcrate.id, 'Machine Gun', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, mechcrate.id, 'Machine Gun',
                                                                           (.8 * mechcratelen),
                                                                           0, 'MECH', mech_skin_label,
                                                                           faction);


                                     i := i + 1;
                            -- for other half of the Mechs, insert a mech object from the appropriate brand's platform mechs and a fitted power core
                            ELSE mech_skin_label :=
                                         get_mech_skin_label_rarity(i, 'MECH', (.2 * mechcratelen), .8 * mechcratelen,
                                                                    faction);

                                 PERFORM insert_mech_into_crate('TURBO', mechcrate.id, faction);
                                 PERFORM insert_mech_skin_into_crate(mechcrate.id, 'PLATFORM',
                                                                     mech_skin_label, faction);

                                 INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                                 VALUES (mechcrate.id, 'POWER_CORE',
                                         (SELECT id FROM blueprint_power_cores c WHERE c.label = 'Turbo Power Core A'));


                                 PERFORM insert_weapon_into_crate(mechcrate.id, 'Flak', faction);
                                 PERFORM insert_weapon_skin_into_crate(i, mechcrate.id, 'Flak',
                                                                       (.2 * mechcratelen),
                                                                       (.8 * mechcratelen), 'MECH', mech_skin_label,
                                                                       faction);

                                 PERFORM insert_weapon_into_crate(mechcrate.id, 'Machine Gun', faction);
                                 PERFORM insert_weapon_skin_into_crate(i, mechcrate.id, 'Machine Gun',
                                                                       (.2 * mechcratelen),
                                                                       (.8 * mechcratelen), 'MECH', mech_skin_label,
                                                                       faction);
                                 i = i + 1;
                            END CASE;
                    END LOOP;

                -- for weapons crates of each faction, insert weapon blueprint.
                i := 1;
                FOR weaponcrate IN SELECT * FROM mystery_crate WHERE faction_id = faction.id AND type = 'WEAPON'
                    LOOP
                        CASE
                            --minigun: all factions
                            WHEN i <= 2000
                                THEN PERFORM insert_weapon_into_crate(weaponcrate.id, 'Minigun', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponcrate.id, 'Minigun',
                                                                           2000,
                                                                           0, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --missile launcher: all factions
                            WHEN i > 2000 AND i <= 4000
                                THEN PERFORM insert_weapon_into_crate(weaponcrate.id, 'Missile Launcher', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponcrate.id,
                                                                           'Missile Launcher',
                                                                           2000,
                                                                           2000, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --plasma gun: all factions
                            WHEN i > 4000 AND i <= 6000
                                THEN PERFORM insert_weapon_into_crate(weaponcrate.id, 'Plasma Gun', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponcrate.id, 'Plasma Gun',
                                                                           2000,
                                                                           4000, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --Laser beam: all factions
                            WHEN i > 6000 AND i <= 8000
                                THEN PERFORM insert_weapon_into_crate(weaponcrate.id, 'Laser Beam', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponcrate.id, 'Laser Beam',
                                                                           2000,
                                                                           6000, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --cannon: all factions
                            WHEN i > 8000 AND i <= 10000
                                THEN PERFORM insert_weapon_into_crate(weaponcrate.id, 'Cannon', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponcrate.id, 'Cannon',
                                                                           2000,
                                                                           8000, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --grenade launcher: all factions
                            WHEN i > 10000 AND i <= 12000
                                THEN PERFORM insert_weapon_into_crate(weaponcrate.id, 'Grenade Launcher', faction);
                                     PERFORM insert_weapon_skin_into_crate(i, weaponcrate.id, 'Grenade Launcher',
                                                                           2000,
                                                                           10000, 'WEAPON', '',
                                                                           faction);
                                     i := i + 1;
                            --BFG, flamethrower or Lightning Gun dependent on faction
                            WHEN i > 12000 AND i <= 13000
                                THEN PERFORM insert_weapon_into_crate(weaponcrate.id,
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
                                             THEN PERFORM insert_weapon_skin_into_crate(i, mechcrate.id,
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
                                                                                                THEN 'Ready To Quench'
                                                                                            WHEN faction.label = 'Zaibatsu Heavy Industries'
                                                                                                THEN 'Catastrophe'
                                                                                            WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                                                                THEN 'Watered Steel'
                                                                                            END,
                                                                                        faction);
                                         WHEN i > 12700 AND i <= 13000
                                             THEN PERFORM insert_weapon_skin_into_crate(i, mechcrate.id,
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
                                                                                                THEN 'Lord of Hell'
                                                                                            WHEN faction.label = 'Zaibatsu Heavy Industries'
                                                                                                THEN 'Calm Before the Storm'
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

--seeding storefront
-- for each faction, seed each type of crate and find how much are for sale

CREATE TABLE IF NOT EXISTS storefront_mystery_crates
(
    id                 UUID PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    mystery_crate_type CRATE_TYPE                                         NOT NULL,
    price              NUMERIC(28, 0)                                     NOT NULL,
    amount             INTEGER                                            NOT NULL,
    amount_sold        INTEGER                  DEFAULT 0                 NOT NULL,
    faction_id         UUID                                               NOT NULL REFERENCES factions (id),
    deleted_at         TIMESTAMP WITH TIME ZONE,
    updated_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()             NOT NULL,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()             NOT NULL,
    label              TEXT                     DEFAULT ''::TEXT          NOT NULL,
    description        TEXT                     DEFAULT ''::TEXT          NOT NULL,
    image_url          TEXT,
    card_animation_url TEXT,
    avatar_url         TEXT,
    large_image_url    TEXT,
    background_color   TEXT,
    animation_url      TEXT,
    youtube_url        TEXT
);

DO
$$
    DECLARE
        faction FACTIONS%ROWTYPE;
    BEGIN
        FOR faction IN SELECT * FROM factions
            LOOP
                INSERT INTO storefront_mystery_crates (mystery_crate_type, amount, faction_id, price, label, description)
                VALUES ('MECH', (SELECT COUNT(*) FROM mystery_crate WHERE type = 'MECH' AND faction_id = faction.id),
                        faction.id, 3000000000000000000000, CASE
                                                                WHEN faction.label = 'Boston Cybernetics' THEN
                                                                    'Boston Cybernetics War Machine Crate'
                                                                WHEN faction.label = 'Zaibatsu Heavy Industries' THEN
                                                                    'Zaibatsu Heavy Industries War Machine Crate'
                                                                WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                                    THEN
                                                                    'Red Mountain War Machine Crate'
                            END,
                        'Contains a battle ready war machine with two weapons.');
                INSERT INTO storefront_mystery_crates (mystery_crate_type, amount, faction_id, price, label, description)
                VALUES ('WEAPON',
                        (SELECT COUNT(*) FROM mystery_crate WHERE type = 'WEAPON' AND faction_id = faction.id),
                        faction.id, 1800000000000000000000, CASE
                                                                WHEN faction.label = 'Boston Cybernetics' THEN
                                                                    'Boston Cybernetics Weapons Crate'
                                                                WHEN faction.label = 'Zaibatsu Heavy Industries' THEN
                                                                    'Zaibatsu Heavy Industries Weapons Crate'
                                                                WHEN faction.label = 'Red Mountain Offworld Mining Corporation'
                                                                    THEN
                                                                    'Red Mountain Weapons Crate'
                            END,
                        'Contains a random weapon and weapon sub model attachment.');
            END LOOP;
    END;
$$;

