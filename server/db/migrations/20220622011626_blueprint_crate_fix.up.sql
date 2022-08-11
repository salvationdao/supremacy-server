DELETE
FROM mystery_crate_blueprints
WHERE blueprint_id =
      (SELECT id FROM blueprint_weapon_skin WHERE label = 'Dantes Inferno');
DELETE
FROM mystery_crate_blueprints
WHERE blueprint_id =
      (SELECT id FROM blueprint_weapon_skin WHERE label = 'Watered Steel');

DELETE
FROM mystery_crate_blueprints
WHERE blueprint_id =
      (SELECT id FROM blueprint_weapon_skin WHERE label = 'Calm Before the Store');
DELETE
FROM mystery_crate_blueprints
WHERE blueprint_id =
      (SELECT id FROM blueprint_weapon_skin WHERE label = 'Catastrophe');

DELETE
FROM mystery_crate_blueprints
WHERE blueprint_id = (SELECT id FROM blueprint_weapon_skin WHERE label = 'Ready To Quench');
DELETE
FROM mystery_crate_blueprints
WHERE blueprint_id = (SELECT id FROM blueprint_weapon_skin WHERE label = 'Lord of Hell');

-- get rid of weapon skins in mech crates
DELETE
FROM mystery_crate_blueprints mcb USING mystery_crate mc
WHERE mc.id = mcb.mystery_crate_id
  AND mc.type = 'MECH'
  AND mcb.blueprint_type = 'WEAPON_SKIN';

-- RM
DO
$$
    DECLARE
        mystery_crate_blueprint MYSTERY_CRATE_BLUEPRINTS%ROWTYPE;
        i                       INTEGER;
    BEGIN
        i := 1;
        FOR mystery_crate_blueprint IN SELECT *
                                       FROM mystery_crate_blueprints mcb
                                       WHERE mcb.blueprint_id =
                                             (SELECT id
                                              FROM blueprint_weapons
                                              WHERE label = 'Pyrotronics F75 HELLFIRE')
            LOOP
                CASE
                    WHEN i <= 700
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (mystery_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN',
                                     (SELECT id
                                      FROM blueprint_weapon_skin
                                      WHERE label = 'Watered Steel'));
                             i := i + 1;
                    WHEN i > 700
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (mystery_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN',
                                     (SELECT id
                                      FROM blueprint_weapon_skin
                                      WHERE label = 'Dantes Inferno'));
                             i := i + 1;
                    END CASE;
            END LOOP;
    END;
$$;

-- ZAI
DO
$$
    DECLARE
        mystery_crate_blueprint MYSTERY_CRATE_BLUEPRINTS%ROWTYPE;
        i                       INTEGER;
    BEGIN
        i := 1;
        FOR mystery_crate_blueprint IN SELECT *
                                       FROM mystery_crate_blueprints mcb
                                       WHERE mcb.blueprint_id =
                                             (SELECT id
                                              FROM blueprint_weapons
                                              WHERE label = 'Warsui SL-750 ELIMINATOR')
            LOOP
                CASE
                    WHEN i <= 700
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (mystery_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN',
                                     (SELECT id
                                      FROM blueprint_weapon_skin
                                      WHERE label = 'Catastrophe'));
                             i := i + 1;
                    WHEN i > 700
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (mystery_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN',
                                     (SELECT id
                                      FROM blueprint_weapon_skin
                                      WHERE label = 'Calm Before the Storm'));
                             i := i + 1;
                    END CASE;
            END LOOP;
    END;
$$;

-- BC
DO
$$
    DECLARE
        mystery_crate_blueprint MYSTERY_CRATE_BLUEPRINTS%ROWTYPE;
        i                       INTEGER;
    BEGIN
        i := 1;
        FOR mystery_crate_blueprint IN SELECT *
                                       FROM mystery_crate_blueprints mcb
                                       WHERE mcb.blueprint_id =
                                             (SELECT id
                                              FROM blueprint_weapons
                                              WHERE label = 'Archon Miltech ARCHON HEAVY BFG-800')
            LOOP
                CASE
                    WHEN i <= 700
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (mystery_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN',
                                     (SELECT id
                                      FROM blueprint_weapon_skin
                                      WHERE label = 'Ready To Quench'));
                             i := i + 1;
                    WHEN i > 700
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (mystery_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN',
                                     (SELECT id
                                      FROM blueprint_weapon_skin
                                      WHERE label = 'Lord of Hell'));
                             i := i + 1;
                    END CASE;
            END LOOP;
    END;
$$;

-- updating mystery crate loot box animation URL

UPDATE storefront_mystery_crates smc
SET animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/mp4/mech/nexus_zai_x3_lootbox_opensea_angled_cropped.mp4'
WHERE mystery_crate_type = 'MECH'
  AND faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries');

UPDATE storefront_mystery_crates
SET animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/mp4/mech/nexus_rm_umc_lootbox_loop_opensea_angled_cropped.mp4'
WHERE mystery_crate_type = 'MECH'
  AND faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation');

UPDATE storefront_mystery_crates
SET animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/mp4/mech/nexus_bc_dai_lootbox_loop_angled_opensea_cropped.mp4'
WHERE mystery_crate_type = 'MECH'
  AND faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics');


UPDATE storefront_mystery_crates
SET animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/mp4/weapon/nexus_zai_wars_lootbox_loop_opensea_angled_cropped.mp4'
WHERE mystery_crate_type = 'WEAPON'
  AND faction_id = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries');

UPDATE storefront_mystery_crates
SET animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/mp4/weapon/nexus_rm_pyro_lootbox_loop_opensea_angled_cropped.mp4'
WHERE mystery_crate_type = 'WEAPON'
  AND faction_id = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation');

UPDATE storefront_mystery_crates
SET animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/mp4/weapon/nexus_bc_am_lootbox_loop_opensea_angled_cropped.mp4'
WHERE mystery_crate_type = 'WEAPON'
  AND faction_id = (SELECT id FROM factions WHERE label = 'Boston Cybernetics');
