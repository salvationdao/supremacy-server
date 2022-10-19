-- move rocket pods to slot 3
-- move what ever was in slot 2, to slot 1
WITH updated_slots AS (
    UPDATE mech_weapons mw SET slot_number = 3 WHERE weapon_id IN (SELECT w.id
                                                                   FROM mech_weapons mw
                                                                            INNER JOIN weapons w ON w.id = mw.weapon_id
                                                                            INNER JOIN blueprint_weapons bw ON w.blueprint_id = bw.id
                                                                   WHERE bw.id IN (
                                                                                   'c1c78867-9de7-43d3-97e9-91381800f38e', -- rocket pod static id
                                                                                   '41099781-8586-4783-9d1c-b515a386fe9f', -- rocket pod static id
                                                                                   'e9fc2417-6a5b-489d-b82e-42942535af90' -- rocket pod static id
                                                                       )
                                                                     AND mw.slot_number != 2)
        RETURNING mw.chassis_id AS chassis_id)
UPDATE mech_weapons mw2
SET slot_number = 0
FROM updated_slots
WHERE mw2.chassis_id = updated_slots.chassis_id
  AND mw2.slot_number = 2;

-- move rocket pods back to slot 2
UPDATE mech_weapons mw SET slot_number = 2 WHERE weapon_id IN (SELECT w.id
                                                               FROM mech_weapons mw
                                                                        INNER JOIN weapons w ON w.id = mw.weapon_id
                                                                        INNER JOIN blueprint_weapons bw ON w.blueprint_id = bw.id
                                                               WHERE bw.id IN (
                                                                               'c1c78867-9de7-43d3-97e9-91381800f38e', -- rocket pod static id
                                                                               '41099781-8586-4783-9d1c-b515a386fe9f', -- rocket pod static id
                                                                               'e9fc2417-6a5b-489d-b82e-42942535af90' -- rocket pod static id
                                                                   )
                                                                 AND mw.slot_number = 3);
