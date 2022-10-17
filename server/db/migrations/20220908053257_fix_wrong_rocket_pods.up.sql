-- this updates the ukarini template skins id
UPDATE template_blueprints
SET blueprint_id = '72ffdb5c-c003-4734-b1d1-07898f5bcb39'
WHERE blueprint_id IN (
                       'f800d5e6-4654-46d5-957c-ecb3387cf588',
                       '80df03b9-53ec-45cc-843d-2ed921973129'
    );

-- this updates the gold template skins id
UPDATE template_blueprints
SET blueprint_id = '11d6ec72-e75e-433a-9bdf-3d82f4ebec9a'
WHERE blueprint_id IN (
                       'ff1b7d82-40b1-4d1a-b93a-f5da92ec43ed',
                       '377060a2-ec4d-4e01-b5f9-a93657ad409f'
    );

-- this updates the shields
UPDATE template_blueprints
SET blueprint_id = 'd429be75-6f98-4231-8315-a86db8477d05'
WHERE blueprint_id = 'b5c6818f-6e11-43a8-a817-6e9a6f8f3a42';


-- below we select all the template blueprints that are boston and rocket pods and then update them to BC rocket pods
WITH update_templates AS (SELECT tbp.id AS template_blueprint_id, t.label, tbp.type, blueprint_id, bw.label
                          FROM template_blueprints tbp
                                   INNER JOIN templates t ON t.id = tbp.template_id
                                   INNER JOIN blueprint_weapons bw ON tbp.blueprint_id = bw.id
                          WHERE tbp.type = 'WEAPON'
                            AND t.label ILIKE '%Boston Cybernetics%'
                            AND bw.label ILIKE '%rocket pod%')
UPDATE template_blueprints _tbp
SET blueprint_id = 'e9fc2417-6a5b-489d-b82e-42942535af90' -- bc rocket pod
FROM update_templates
WHERE _tbp.id = template_blueprint_id;

-- below we select all the template blueprints that are rm and rocket pods and then update them to rm rocket pods
WITH update_templates AS (SELECT tbp.id AS template_blueprint_id, t.label, tbp.type, blueprint_id, bw.label
                          FROM template_blueprints tbp
                                   INNER JOIN templates t ON t.id = tbp.template_id
                                   INNER JOIN blueprint_weapons bw ON tbp.blueprint_id = bw.id
                          WHERE tbp.type = 'WEAPON'
                            AND t.label ILIKE '%red mou%'
                            AND bw.label ILIKE '%rocket pod%')
UPDATE template_blueprints _tbp
SET blueprint_id = '41099781-8586-4783-9d1c-b515a386fe9f' -- rm rocket pod
FROM update_templates
WHERE _tbp.id = template_blueprint_id;

-- update rm mech rocket pods to have the correct weapon blueprint id
WITH to_update AS (SELECT _w.id AS weapon_id
                   FROM mechs m
                            INNER JOIN mech_weapons mw ON m.id = mw.chassis_id AND slot_number = 2
                            INNER JOIN weapons _w ON _w.id = mw.weapon_id
                            INNER JOIN blueprint_weapons bw ON bw.id = _w.blueprint_id
                   WHERE m.blueprint_id = 'ac27f3b9-753d-4ace-84a9-21c041195344'
                     AND bw.label != 'RMMC Rocket Pod')
UPDATE weapons w
SET blueprint_id = '41099781-8586-4783-9d1c-b515a386fe9f' -- RMMC rocket pod
FROM to_update
WHERE w.id = to_update.weapon_id;

-- update bc mech rocket pods to have the correct weapon blueprint id
WITH to_update AS (SELECT _w.id AS weapon_id
                   FROM mechs m
                            INNER JOIN mech_weapons mw ON m.id = mw.chassis_id AND slot_number = 2
                            INNER JOIN weapons _w ON _w.id = mw.weapon_id
                            INNER JOIN blueprint_weapons bw ON bw.id = _w.blueprint_id
                   WHERE m.blueprint_id = '5d3a973b-c62b-4438-b746-d3de2699d42a'
                     AND bw.label != 'BC Rocket Pod')
UPDATE weapons w
SET blueprint_id = 'e9fc2417-6a5b-489d-b82e-42942535af90' -- BC Rocket Pod
FROM to_update
WHERE w.id = to_update.weapon_id;

-- after the above update all skins on the weapons we changed will be invalid, causing issues, so now we need to update the skins that are equipped on the weapons we changed
-- update all weapons that do not have a valid skin, set as a valid one (rocket pods only have one skin each, so the set sub select should only return one row)
-- 17/10/2022 had to update this query due to static data changes
WITH to_update AS (
-- this returns all weapons that have an invalid skin equipped
-- added `AND bms.blueprint_weapon_skin_id IS NOT NULL` and `SET blueprint_id = to_update.new_id`
    SELECT ws.id          AS skin_id,
           w.blueprint_id AS weapon_blueprint_id,
           bw.label as weapon_label,
           bws.label as weapon_skin_label,
           bms.label as mech_skin_label,
           bms.blueprint_weapon_skin_id as new_id
    FROM weapon_skin ws
             INNER JOIN weapons w ON ws.equipped_on = w.id
             INNER JOIN mech_weapons mw ON mw.weapon_id = w.id
             INNER JOIN blueprint_weapons bw ON bw.id = w.blueprint_id
             INNER JOIN blueprint_weapon_skin bws ON bws.id = ws.blueprint_id
             INNER JOIN mechs m ON m.id = mw.chassis_id
             INNER JOIN mech_skin ms ON ms.equipped_on = m.id
             INNER JOIN blueprint_mech_skin bms ON bms.id = ms.blueprint_id
        -- this will be null if there isn't a entry in weapon_model_skin_compatibilities for this weapon/skin
             LEFT OUTER JOIN weapon_model_skin_compatibilities wmsc
                             ON wmsc.weapon_model_id = w.blueprint_id AND wmsc.blueprint_weapon_skin_id = ws.blueprint_id
    WHERE wmsc.blueprint_weapon_skin_id IS NULL AND bms.blueprint_weapon_skin_id IS NOT NULL) -- only return null entries
UPDATE weapon_skin _ws
SET blueprint_id = to_update.new_id
FROM to_update
WHERE _ws.id = to_update.skin_id;
