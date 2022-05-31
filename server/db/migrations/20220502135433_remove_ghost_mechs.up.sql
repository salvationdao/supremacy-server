DELETE
FROM mech_stats
WHERE mech_id IN ('c120db1d-ca1c-4d03-b73c-b16bc5eb08cb', '615e1e0b-6604-4d6f-8b3c-a161bf4e4558');


-- delete weapons
WITH deleted_weapons AS (
    WITH wpns AS (SELECT cw.weapon_id
                  FROM chassis_weapons cw
                  WHERE cw.chassis_id IN (SELECT chassis_id
                                          FROM mechs
                                          WHERE id IN
                                                ('c120db1d-ca1c-4d03-b73c-b16bc5eb08cb',
                                                 '615e1e0b-6604-4d6f-8b3c-a161bf4e4558')))
        DELETE FROM weapons w WHERE w.id = (SELECT weapon_id FROM wpns WHERE weapon_id = w.id)
            RETURNING w.id)
DELETE
FROM chassis_weapons cw
WHERE cw.weapon_id = (SELECT id FROM deleted_weapons WHERE cw.weapon_id = id);

-- delete utility
WITH deleted_mods AS (
    WITH mods AS (SELECT cm.module_id
                  FROM chassis_modules cm
                  WHERE cm.chassis_id IN (SELECT chassis_id
                                          FROM mechs
                                          WHERE id IN
                                                ('c120db1d-ca1c-4d03-b73c-b16bc5eb08cb',
                                                 '615e1e0b-6604-4d6f-8b3c-a161bf4e4558')))
        DELETE FROM modules m WHERE m.id = (SELECT module_id FROM mods WHERE module_id = m.id)
            RETURNING m.id)
DELETE
FROM chassis_modules cm
WHERE cm.module_id = (SELECT id FROM deleted_mods WHERE cm.module_id = id);

-- delete chassis
WITH deleted_mechs AS (
    DELETE
        FROM mechs
            WHERE id IN ('c120db1d-ca1c-4d03-b73c-b16bc5eb08cb', '615e1e0b-6604-4d6f-8b3c-a161bf4e4558')
            RETURNING id, chassis_id)
DELETE
FROM chassis c
WHERE c.id IN (SELECT id FROM deleted_mechs WHERE chassis_id = c.id);
