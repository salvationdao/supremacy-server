-- updating bws names to weapon manufacturer's skins in rarities- weapons manufacturer skin is accounted for in rarity
UPDATE blueprint_weapon_skin
SET label = 'Archon Gunmetal'
WHERE label = 'BC';

UPDATE blueprint_weapon_skin
SET label = 'Pyro Crimson'
WHERE label = 'RM';

UPDATE blueprint_weapon_skin
SET label = 'Verdant Warsui'
WHERE label = 'Zaibatsu';

--updating bms manufacturer's names to faction- mech manufacturer's skin is OUT of raririties and faction skin is IN and accounted for in rarities
UPDATE blueprint_mech_skin
SET label = 'Spot Yellow'
WHERE label = 'Daison Avionics';
UPDATE blueprint_weapon_skin
SET label = 'Spot Yellow'
WHERE label = 'Daison Avionics';

UPDATE blueprint_mech_skin
SET label = 'Heavy White'
WHERE label = 'X3W';
UPDATE blueprint_weapon_skin
SET label = 'Heavy White'
WHERE label = 'X3W';

UPDATE blueprint_mech_skin
SET label = 'Pilbara Dust'
WHERE label = 'UMC';
UPDATE blueprint_weapon_skin
SET label = 'Pilbara Dust'
WHERE label = 'UMC';

--Creating new skin that is manufacturer's "default", all mech crates will receive this mech skin
DO
$$
    DECLARE
        mech_modelr MECH_MODELS%ROWTYPE;
    BEGIN
        FOR mech_modelr IN SELECT * FROM mech_models
            LOOP
                CASE
                    -- BC
                    WHEN mech_modelr.brand_id = (SELECT id FROM brands WHERE label = 'Daison Avionics')
                        THEN INSERT INTO blueprint_mech_skin (label, tier, mech_type, mech_model)
                             VALUES ('Daison Sleek', 'COLOSSAL', mech_modelr.mech_type, mech_modelr.id);
                             UPDATE blueprint_mech_skin
                             SET label = 'Bullion'
                             WHERE label = 'Gold'
                               AND mech_model = mech_modelr.id;
                    -- ZAI
                    WHEN mech_modelr.brand_id = (SELECT id FROM brands WHERE label = 'X3 Wartech')
                        THEN INSERT INTO blueprint_mech_skin (label, tier, mech_type, mech_model)
                             VALUES ('X3 Kuro', 'COLOSSAL', mech_modelr.mech_type, mech_modelr.id);
                             UPDATE blueprint_mech_skin
                             SET label = 'Mine God'
                             WHERE label = 'Gold'
                               AND mech_model = mech_modelr.id;
                    -- RM
                    WHEN mech_modelr.brand_id = (SELECT id FROM brands WHERE label = 'Unified Martian Corporation')
                        THEN INSERT INTO blueprint_mech_skin (label, tier, mech_type, mech_model)
                             VALUES ('Martian Soil', 'COLOSSAL', mech_modelr.mech_type, mech_modelr.id);
                             UPDATE blueprint_mech_skin
                             SET label = 'Sovereign Hill'
                             WHERE label = 'Gold'
                               AND mech_model = mech_modelr.id;
                    ELSE CONTINUE;
                    END CASE;
            END LOOP;
    END;
$$;

DO
$$
    DECLARE
        weapon_model WEAPON_MODELS%ROWTYPE;
    BEGIN
        FOR weapon_model IN SELECT * FROM weapon_models
            LOOP
                CASE
                    -- BC
                    WHEN weapon_model.brand_id = (SELECT id FROM brands WHERE label = 'Archon Miltech')
                        THEN INSERT INTO blueprint_weapon_skin (label, tier, weapon_type, weapon_model_id)
                             VALUES ('Daison Sleek', 'COLOSSAL', weapon_model.weapon_type, weapon_model.id);
                             UPDATE blueprint_weapon_skin
                             SET label = 'Bullion'
                             WHERE label = 'Gold'
                               AND weapon_model_id = weapon_model.id;
                    -- ZAI
                    WHEN weapon_model.brand_id = (SELECT id FROM brands WHERE label = 'Warsui')
                        THEN INSERT INTO blueprint_weapon_skin (label, tier, weapon_type, weapon_model_id)
                             VALUES ('X3 Kuro', 'COLOSSAL', weapon_model.weapon_type, weapon_model.id);
                             UPDATE blueprint_weapon_skin
                             SET label = 'Mine God'
                             WHERE label = 'Gold'
                               AND weapon_model_id = weapon_model.id;
                    -- RM
                    WHEN weapon_model.brand_id = (SELECT id FROM brands WHERE label = 'Pyrotronics')
                        THEN INSERT INTO blueprint_weapon_skin (label, tier, weapon_type, weapon_model_id)
                             VALUES ('Martian Soil', 'COLOSSAL', weapon_model.weapon_type, weapon_model.id);
                             UPDATE blueprint_weapon_skin
                             SET label = 'Sovereign Hill'
                             WHERE label = 'Gold'
                               AND weapon_model_id = weapon_model.id;
                    ELSE CONTINUE;
                    END CASE;
            END LOOP;
    END;
$$;

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
                                                                                             WHERE label = 'Archon Gunmetal'
                                                                                               AND weapon_type = 'Flak'));
                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Archon Gunmetal'
                                                                                               AND weapon_type = 'Machine Gun'));

                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'MECH_SKIN',
                                     CASE
                                         WHEN powercoresize = 'SMALL' THEN (SELECT id
                                                                            FROM blueprint_mech_skin
                                                                            WHERE label = 'Daison Sleek'
                                                                              AND mech_type = 'HUMANOID'::MECH_TYPE)
                                         WHEN powercoresize = 'MEDIUM' THEN (SELECT id
                                                                             FROM blueprint_mech_skin
                                                                             WHERE label = 'Daison Sleek'
                                                                               AND mech_type = 'PLATFORM'::MECH_TYPE) END);
                    -- ZAI
                    WHEN factionid = (SELECT id FROM factions WHERE label = 'Zaibatsu Heavy Industries')
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Verdant Warsui'
                                                                                               AND weapon_type = 'Flak'));
                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Verdant Warsui'
                                                                                               AND weapon_type = 'Machine Gun'));


                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'MECH_SKIN',
                                     CASE
                                         WHEN powercoresize = 'SMALL' THEN (SELECT id
                                                                            FROM blueprint_mech_skin
                                                                            WHERE label = 'X3 Kuro'
                                                                              AND mech_type = 'HUMANOID'::MECH_TYPE)
                                         WHEN powercoresize = 'MEDIUM' THEN (SELECT id
                                                                             FROM blueprint_mech_skin
                                                                             WHERE label = 'X3 Kuro'
                                                                               AND mech_type = 'PLATFORM'::MECH_TYPE) END);
                    -- RM
                    WHEN factionid = (SELECT id FROM factions WHERE label = 'Red Mountain Offworld Mining Corporation')
                        THEN INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Pyro Crimson'
                                                                                               AND weapon_type = 'Flak'));
                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'WEAPON_SKIN', (SELECT id
                                                                                             FROM blueprint_weapon_skin
                                                                                             WHERE label = 'Pyro Crimson'
                                                                                               AND weapon_type = 'Machine Gun'));


                             INSERT INTO mystery_crate_blueprints (mystery_crate_id, blueprint_type, blueprint_id)
                             VALUES (vmech_crate_blueprint.mystery_crate_id, 'MECH_SKIN',
                                     CASE
                                         WHEN powercoresize = 'SMALL' THEN (SELECT id
                                                                            FROM blueprint_mech_skin
                                                                            WHERE label = 'Martian Soil'
                                                                              AND mech_type = 'HUMANOID'::MECH_TYPE)
                                         WHEN powercoresize = 'MEDIUM' THEN (SELECT id
                                                                             FROM blueprint_mech_skin
                                                                             WHERE label = 'Martian Soil'
                                                                               AND mech_type = 'PLATFORM'::MECH_TYPE) END); END CASE;
            END LOOP;
    END;
$$;


-- rename skins
----MECH
-- BC
UPDATE blueprint_mech_skin
SET label = 'Sea Hawk'
WHERE label = 'Raptor';
UPDATE blueprint_weapon_skin
SET label = 'Sea Hawk'
WHERE label = 'Raptor';

UPDATE blueprint_mech_skin
SET label = 'Thin Blue Line'
WHERE label = 'Rexeon Guard';
UPDATE blueprint_weapon_skin
SET label = 'Thin Blue Line'
WHERE label = 'Rexeon Guard';

-- Gold updated in loop

UPDATE blueprint_mech_skin
SET label = 'Code of Chivalry'
WHERE label = 'Paladin';
UPDATE blueprint_weapon_skin
SET label = 'Code of Chivalry'
WHERE label = 'Paladin';

UPDATE blueprint_mech_skin
SET label = 'Telling the Bees'
WHERE label = 'Hive';
UPDATE blueprint_weapon_skin
SET label = 'Telling the Bees'
WHERE label = 'Hive';


-- ZAI
UPDATE blueprint_mech_skin
SET label = 'Nullifier'
WHERE label = 'XHANCR';
UPDATE blueprint_weapon_skin
SET label = 'Nullifier'
WHERE label = 'XHANCR';

UPDATE blueprint_mech_skin
SET label = 'Two Five Zero One'
WHERE label = '2501 - Tachikoma';
UPDATE blueprint_weapon_skin
SET label = 'Two Five Zero One'
WHERE label = '2501 - Tachikoma';

-- Gold updated in loop

UPDATE blueprint_mech_skin
SET label = 'Shadows Steal Away'
WHERE label = 'Shinobi';
UPDATE blueprint_weapon_skin
SET label = 'Shadows Steal Away'
WHERE label = 'Shinobi';
--synth punk has same name


-- RM
UPDATE blueprint_mech_skin
SET label = 'Military'
WHERE label = 'High Caliber';
UPDATE blueprint_weapon_skin
SET label = 'Military'
WHERE label = 'High Caliber';

UPDATE blueprint_mech_skin
SET label = 'Fly In Fly Out'
WHERE label = 'Mining';
UPDATE blueprint_weapon_skin
SET label = 'Fly In Fly Out'
WHERE label = 'Mining';

-- Gold updated in loop

UPDATE blueprint_mech_skin
SET label = 'Osmium Scream'
WHERE label = 'Heavy Metal';
UPDATE blueprint_weapon_skin
SET label = 'Osmium Scream'
WHERE label = 'Heavy Metal';

UPDATE blueprint_mech_skin
SET label = 'Promethean Gold'
WHERE label = 'Molten';
UPDATE blueprint_weapon_skin
SET label = 'Promethean Gold'
WHERE label = 'Molten';


----WEAPON
-- BC
UPDATE blueprint_weapon_skin
SET label = 'Praetor Grunge'
WHERE label = 'Space Marine';
UPDATE blueprint_weapon_skin
SET label = 'Less-Than-Lethal'
WHERE label = 'Nerf Gun';
UPDATE blueprint_weapon_skin
SET label = 'Unbroken Knot'
WHERE label = 'Celtic Knot';
UPDATE blueprint_weapon_skin
SET label = 'Ready To Quench'
WHERE label = 'Cybernetics';
UPDATE blueprint_weapon_skin
SET label = 'Lord of Hell'
WHERE label = 'Doom';

-- ZAI
UPDATE blueprint_weapon_skin
SET label = 'Violet Ice'
WHERE label = 'Purple and White';
UPDATE blueprint_weapon_skin
SET label = 'Rebellion'
WHERE label = 'Sonnō jōi';
UPDATE blueprint_weapon_skin
SET label = 'Cephalopod Ripple'
WHERE label = 'Logogram - Arrival';
UPDATE blueprint_weapon_skin
SET label = 'CATastrophe'
WHERE label = 'Neko';
UPDATE blueprint_weapon_skin
SET label = 'Calm Before the Storm'
WHERE label = 'BOTW';

-- RM
UPDATE blueprint_weapon_skin
SET label = 'Barricade Stripes'
WHERE label = 'Hazard';
UPDATE blueprint_weapon_skin
SET label = 'Shark Skin'
WHERE label = 'Martian Marine Core';
UPDATE blueprint_weapon_skin
SET label = 'Martian Mess Maker'
WHERE label = 'Cassowary';
UPDATE blueprint_weapon_skin
SET label = 'Watered Steel'
WHERE label = 'Damascus';
--Dantes Inferno does not need to be renamed


-- change weapon utility slots for platform and humanoid mechs
UPDATE blueprint_mechs
SET weapon_hardpoints = 4,
    utility_slots     = 4
WHERE power_core_size = 'MEDIUM';
UPDATE blueprint_mechs
SET weapon_hardpoints = 2,
    utility_slots     = 2
WHERE utility_slots = 4;


--update asset images
--MECHS
--BC
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/humanoid/nexus_dai_war-enforcer_daison-sleek.png'
WHERE label = 'Daison Sleek'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/humanoid/nexus_dai_war-enforcer_spot-yellow.png'
WHERE label = 'Spot Yellow'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/humanoid/nexus_dai_war-enforcer_sea-hawk.png'
WHERE label = 'Sea Hawk'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/humanoid/nexus_dai_war-enforcer_thin-blue-line.png'
WHERE label = 'Thin Blue Line'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/humanoid/nexus_dai_war-enforcer_bullion.png'
WHERE label = 'Bullion'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/humanoid/nexus_dai_war-enforcer_code-of-chivalry.png'
WHERE label = 'Code of Chivalry'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/humanoid/nexus_dai_war-enforcer_telling-the-bees.png'
WHERE label = 'Telling the Bees'
  AND mech_type = 'HUMANOID';

UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/platform/nexus_dai_annihilator_daison-sleek.png'
WHERE label = 'Daison Sleek'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/platform/nexus_dai_annihilator_spot-yellow.png'
WHERE label = 'Spot Yellow'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/platform/nexus_dai_annihilator_sea-hawk.png'
WHERE label = 'Sea Hawk'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/platform/nexus_dai_annihilator_thin-blue-line.png'
WHERE label = 'Thin Blue Line'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/platform/nexus_dai_annihilator_bullion.png'
WHERE label = 'Bullion'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/platform/nexus_dai_annihilator_code-of-chivalry.png'
WHERE label = 'Code of Chivalry'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/bc-daison-avionics/platform/nexus_dai_annihilator_telling-the-bees.png'
WHERE label = 'Telling the Bees'
  AND mech_type = 'PLATFORM';

--zai
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_kenji_x3-kuro.png'
WHERE label = 'X3 Kuro'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_kenji_heavy-white.png'
WHERE label = 'Heavy White'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_kenji_nullifier.png'
WHERE label = 'Nullifier'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_kenji_two-five-zero-one.png'
WHERE label = 'Two Five Zero One'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_kenji_mine-god.png'
WHERE label = 'Mine God'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_kenji_shadows-steal-away.png'
WHERE label = 'Shadows Steal Away'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_kenji_synth-punk.png'
WHERE label = 'Synth Punk'
  AND mech_type = 'HUMANOID';

UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_shirokuma_x3-kuro.png'
WHERE label = 'X3 Kuro'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_shirokuma_heavy-white.png'
WHERE label = 'Heavy White'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_shirokuma_nullifier.png'
WHERE label = 'Nullifier'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_shirokuma_two-five-zero-one.png'
WHERE label = 'Two Five Zero One'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_shirokuma_mine-god.png'
WHERE label = 'Mine God'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_shirokuma_shadows-steal-away.png'
WHERE label = 'Shadows Steal Away'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/zai-x3-wartech/humanoid/nexus_x3_kenji_synth-punk.png'
WHERE label = 'Synth Punk'
  AND mech_type = 'PLATFORM';

--rm
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_aries_martian-soil.png'
WHERE label = 'Martian Soil'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_aries_pilbara-dust.png'
WHERE label = 'Pilbara Dust'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_aries_high-caliber.png'
WHERE label = 'High Caliber'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_aries_fifo.png'
WHERE label = 'Fly In Fly Out'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_aries_sovereign-hill.png'
WHERE label = 'Sovereign Hill'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_aries_osmium-scream.png'
WHERE label = 'Osmium Scream'
  AND mech_type = 'HUMANOID';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_aries_promethean-gold.png'
WHERE label = 'Promethean Gold'
  AND mech_type = 'HUMANOID';

UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_viking_martian-soil.png'
WHERE label = 'Martian Soil'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_viking_pilbara-dust.png'
WHERE label = 'Pilbara Dust'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_viking_high-caliber.png'
WHERE label = 'High Caliber'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_viking_fifo.png'
WHERE label = 'Fly In Fly Out'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_viking_sovereign-hill.png'
WHERE label = 'Sovereign Hill'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_viking_osmium-scream.png'
WHERE label = 'Osmium Scream'
  AND mech_type = 'PLATFORM';
UPDATE blueprint_mech_skin
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/mech/rm-umc/humanoid/nexus_umc_viking_promethean-gold.png'
WHERE label = 'Promethean Gold'
  AND mech_type = 'PLATFORM';


--WEAPONS

