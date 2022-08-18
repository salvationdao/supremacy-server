ALTER TABLE mechs
    DROP COLUMN IF EXISTS weapon_hardpoints,
    DROP COLUMN IF EXISTS utility_slots,
    DROP COLUMN IF EXISTS speed,
    DROP COLUMN IF EXISTS max_hitpoints,
    DROP COLUMN IF EXISTS power_core_size;

ALTER TABLE mech_skin
    ADD COLUMN level INT NOT NULL DEFAULT 0;

UPDATE mech_skin ms SET level = (select default_level from blueprint_mech_skin bms where ms.blueprint_id = bms.id)
