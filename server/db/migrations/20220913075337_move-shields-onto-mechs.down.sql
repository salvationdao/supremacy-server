ALTER TABLE blueprint_mechs
    ALTER COLUMN shield_type_id DROP NOT NULL;

WITH shields AS (SELECT _u.id, _u.equipped_on
                 FROM utility _u)
UPDATE mech_utility mu
SET utility_id = shields.id
FROM shields
WHERE mu.chassis_id = shields.equipped_on;

ALTER TABLE mech_utility
    ALTER COLUMN utility_id SET NOT NULL;

UPDATE template_blueprints
SET deleted_at = NULL
WHERE template_blueprints.type = 'UTILITY';

ALTER TABLE template_blueprints
    DROP COLUMN deleted_at;

UPDATE collection_items
SET deleted_at = NULL
WHERE item_id IN (SELECT id
                  FROM utility
                  WHERE type = 'SHIELD');

ALTER TABLE collection_items
    DROP COLUMN deleted_at;

UPDATE utility_shield_dont_use
SET deleted_at = NULL;

ALTER TABLE utility_shield_dont_use
    DROP COLUMN deleted_at;

UPDATE utility
SET deleted_at = NULL
WHERE type = 'SHIELD';
