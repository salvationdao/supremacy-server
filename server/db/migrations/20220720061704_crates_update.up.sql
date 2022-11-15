ALTER TABLE mystery_crate_blueprints
    DROP CONSTRAINT mystery_crate_blueprints_mystery_crate_id_fkey;

ALTER TABLE mystery_crate_blueprints
    ADD CONSTRAINT mystery_crate_blueprints_mystery_crate_id_fkey
        FOREIGN KEY (mystery_crate_id)
            REFERENCES mystery_crate (id)
            ON DELETE CASCADE;

--for each mech crate, insert another mech skin(BC, ZAI, RM) and 2x weapon manufacturer's skins
DO
$$
    DECLARE
        vmech_crate_blueprint MYSTERY_CRATE_BLUEPRINTS%ROWTYPE;
        factionid             UUID;
        powercoresize         TEXT;

    BEGIN
        FOR vmech_crate_blueprint IN SELECT * FROM mystery_crate_blueprints WHERE blueprint_type = 'MECH'
            LOOP

                factionid := (SELECT faction_id FROM mystery_crate WHERE id = vmech_crate_blueprint.mystery_crate_id);
                powercoresize :=
                        (SELECT power_core_size FROM blueprint_mechs WHERE id = vmech_crate_blueprint.blueprint_id);

                CASE
                    -- BC
                    WHEN factionid = (SELECT id FROM factions WHERE label = 'Boston Cybernetics')
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Archon Gunmetal'));
                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Archon Gunmetal'));

                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'MECH_SKIN',
                                     CASE
                                         WHEN powercoresize = 'SMALL' THEN (SELECT id
                                                                            FROM blueprint_mech_skin
                                                                            WHERE label = 'Daison Sleek')
                                         WHEN powercoresize = 'TURBO' THEN (SELECT id
                                                                             FROM blueprint_mech_skin
                                                                             WHERE label = 'Daison Sleek') END);
                    -- ZAI
                    WHEN factionid = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries')
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Verdant Warsui'));
                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Verdant Warsui'));


                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'MECH_SKIN',
                                     CASE
                                         WHEN powercoresize = 'SMALL' THEN (SELECT id
                                                                            FROM blueprint_mech_skin
                                                                            WHERE label = 'X3 Kuro')
                                         WHEN powercoresize = 'TURBO' THEN (SELECT id
                                                                             FROM blueprint_mech_skin
                                                                             WHERE label = 'X3 Kuro') END);
                    -- RM
                    WHEN factionid = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation')
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Pyro Crimson'));
                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Pyro Crimson'));


                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'MECH_SKIN',
                                     CASE
                                         WHEN powercoresize = 'SMALL' THEN (SELECT id
                                                                            FROM blueprint_mech_skin
                                                                            WHERE label = 'Martian Soil')
                                         WHEN powercoresize = 'TURBO' THEN (SELECT id
                                                                             FROM blueprint_mech_skin
                                                                             WHERE label = 'Martian Soil') END); END CASE;
            END LOOP;
    END;
$$;

ALTER TABLE weapon_skin
    DROP COLUMN IF EXISTS weapon_model_id;

-- fix powercore images
UPDATE collection_items SET image_url = avatar_url WHERE item_type = 'power_core';
UPDATE blueprint_power_cores SET image_url = avatar_url;
