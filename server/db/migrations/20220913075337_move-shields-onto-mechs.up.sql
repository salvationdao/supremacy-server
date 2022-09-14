UPDATE utility
SET deleted_at = NOW()
WHERE type = 'SHIELD';

ALTER TABLE utility_shield_dont_use
    ADD COLUMN deleted_at TIMESTAMPTZ;

UPDATE utility_shield_dont_use
SET deleted_at = NOW();

ALTER TABLE collection_items
    ADD COLUMN deleted_at TIMESTAMPTZ;

UPDATE collection_items
SET deleted_at = NOW()
WHERE item_id IN (SELECT id
                  FROM utility
                  WHERE type = 'SHIELD');

ALTER TABLE template_blueprints
    ADD COLUMN deleted_at TIMESTAMPTZ;

UPDATE template_blueprints
SET deleted_at = NOW()
WHERE template_blueprints.type = 'UTILITY';

ALTER TABLE mech_utility
    ALTER COLUMN utility_id DROP NOT NULL;

UPDATE mech_utility SET utility_id = NULL; -- we can get this id for the down from the shield.equipped_on

ALTER TABLE blueprint_mechs
    ALTER COLUMN shield_type_id SET NOT NULL;
