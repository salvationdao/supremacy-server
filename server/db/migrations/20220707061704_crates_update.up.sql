-- updating bws names to weapon manufacturer's skins in rarities
UPDATE blueprint_weapon_skin
SET label = 'Archon Miltech'
WHERE label = 'BC';

UPDATE blueprint_weapon_skin
SET label = 'Pyrotronics'
WHERE label = 'RM';

UPDATE blueprint_weapon_skin
SET label = 'Warsui'
WHERE label = 'Zaibatsu';

--updating bms manufacturer's names to faction
UPDATE blueprint_mech_skin
SET label = 'BC'
WHERE label = 'Daison Avionics';
UPDATE blueprint_weapon_skin
SET label = 'BC'
WHERE label = 'Daison Avionics';

UPDATE blueprint_mech_skin
SET label = 'ZAI'
WHERE label = 'X3W';
UPDATE blueprint_weapon_skin
SET label = 'ZAI'
WHERE label = 'X3W';

UPDATE blueprint_mech_skin
SET label = 'RM'
WHERE label = 'UMC';
UPDATE blueprint_weapon_skin
SET label = 'RM'
WHERE label = 'RM';


--Creating new skin that is manufacturer's "default", all mech crates will receive this mech skin
DO
$$
    DECLARE
        mech_modelr MECH_MODELS%ROWTYPE;
    BEGIN
        FOR mech_modelr IN SELECT *
                           FROM mech_models
            LOOP
                CASE
                    -- BC
                    WHEN mech_modelr.brand_id = (SELECT id FROM brands WHERE label = 'Archon Miltech')
                        THEN INSERT INTO blueprint_mech_skin (label, tier, mech_type, mech_model)
                             VALUES ('Daison Avionics', 'COLOSSAL', mech_modelr.mech_type, mech_modelr.id);
                    -- ZAI
                    WHEN mech_modelr.brand_id = (SELECT id FROM brands WHERE label = 'Warsui')
                        THEN INSERT INTO blueprint_mech_skin (label, tier, mech_type, mech_model)
                             VALUES ('X3 Wartech', 'COLOSSAL', mech_modelr.mech_type, mech_modelr.id);
                    -- RM
                    WHEN mech_modelr.brand_id = (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        THEN INSERT INTO blueprint_mech_skin (label, tier, mech_type, mech_model)
                             VALUES ('Unified Martian Corps', 'COLOSSAL', mech_modelr.mech_type, mech_modelr.id);
                    END CASE;
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
            LOOP
                CASE
                    -- BC
                    WHEN weapon_model.brand_id = (SELECT id FROM brands WHERE label = 'Archon Miltech')
                        THEN INSERT INTO blueprint_weapon_skin (label, tier, weapon_type, weapon_model_id)
                             VALUES ('Daison Avionics', 'COLOSSAL', weapon_model.weapon_type, weapon_model.id);
                    -- ZAI
                    WHEN weapon_model.brand_id = (SELECT id FROM brands WHERE label = 'Warsui')
                        THEN INSERT INTO blueprint_weapon_skin (label, tier, weapon_type, weapon_model_id)
                             VALUES ('X3 Wartech', 'COLOSSAL', weapon_model.weapon_type, weapon_model.id);
                    -- RM
                    WHEN weapon_model.brand_id = (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        THEN INSERT INTO blueprint_weapon_skin (label, tier, weapon_type, weapon_model_id)
                             VALUES ('Unified Martian Corps', 'COLOSSAL', weapon_model.weapon_type, weapon_model.id);
                    END CASE;
            END LOOP;
    END;
$$;

--for each mech crate, insert another mech skin(BC, ZAI, RM) and 2x weapon manufacturer's skins

--delete crates where # is wrong

-- change weapon utility slots for platform and humanoid mechs
