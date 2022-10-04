-- since we had a skin for sniper rifle and a skin for laser sword, we can merge them to both use a single skin blueprint
-- replace - 2bde926c-07b9-4ff5-965c-49a0660e6a0f - with - f37fba4a-1eeb-49f5-8334-cacca614c0b5
-- update currently owned skins
UPDATE weapon_skin
SET blueprint_id = 'f37fba4a-1eeb-49f5-8334-cacca614c0b5'
WHERE blueprint_id = '2bde926c-07b9-4ff5-965c-49a0660e6a0f';

-- since we had a skin for plasma rifle and a skin for sword, we can merge them to both use a single skin blueprint
-- replace - 79ffd223-5606-48b5-8b7d-1db6018125ce - with - 0a3ea36f-aa1c-4787-b1fd-e6277b0a860c
-- update currently owned skins
UPDATE weapon_skin
SET blueprint_id = '0a3ea36f-aa1c-4787-b1fd-e6277b0a860c'
WHERE blueprint_id = '79ffd223-5606-48b5-8b7d-1db6018125ce';

-- As above, rocket pods did have their own skins, but they don't need them. They can use the same skin blueprint as their weapons
--replace rocket pod skins;
-- bc new skin = 0a3ea36f-aa1c-4787-b1fd-e6277b0a860c - rocket pod id - e9fc2417-6a5b-489d-b82e-42942535af90
UPDATE weapon_skin
SET blueprint_id = '0a3ea36f-aa1c-4787-b1fd-e6277b0a860c'
WHERE id IN (SELECT ws.id
             FROM weapons w
                      INNER JOIN weapon_skin ws ON ws.id = w.equipped_weapon_skin_id
             WHERE w.blueprint_id = 'e9fc2417-6a5b-489d-b82e-42942535af90');
-- zai new skin = f37fba4a-1eeb-49f5-8334-cacca614c0b5 - rocket pod id - c1c78867-9de7-43d3-97e9-91381800f38e
UPDATE weapon_skin
SET blueprint_id = 'f37fba4a-1eeb-49f5-8334-cacca614c0b5'
WHERE id IN (SELECT ws.id
             FROM weapons w
                      INNER JOIN weapon_skin ws ON ws.id = w.equipped_weapon_skin_id
             WHERE w.blueprint_id = 'c1c78867-9de7-43d3-97e9-91381800f38e');
-- rm new skin = e2f0aaf9-67d1-4f1f-8da2-049321f53bd5 - rocket pod id - 41099781-8586-4783-9d1c-b515a386fe9f
UPDATE weapon_skin
SET blueprint_id = 'e2f0aaf9-67d1-4f1f-8da2-049321f53bd5'
WHERE id IN (SELECT ws.id
             FROM weapons w
                      INNER JOIN weapon_skin ws ON ws.id = w.equipped_weapon_skin_id
             WHERE w.blueprint_id = '41099781-8586-4783-9d1c-b515a386fe9f');


-- here we are setting all genesis weapons to inherit skins as default
UPDATE mech_weapons
SET is_skin_inherited = TRUE
WHERE weapon_id IN (SELECT mw.weapon_id
                    FROM mech_weapons mw
                             INNER JOIN mechs m ON m.id = mw.chassis_id
                    WHERE blueprint_id IN (
                                           '5d3a973b-c62b-4438-b746-d3de2699d42a',
                                           'ac27f3b9-753d-4ace-84a9-21c041195344',
                                           '625cd381-7c66-4e2f-9f69-f81589105730'
                        ));
