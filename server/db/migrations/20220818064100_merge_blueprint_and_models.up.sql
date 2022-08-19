ALTER TABLE mechs
    DROP COLUMN IF EXISTS weapon_hardpoints,
    DROP COLUMN IF EXISTS utility_slots,
    DROP COLUMN IF EXISTS speed,
    DROP COLUMN IF EXISTS max_hitpoints,
    DROP COLUMN IF EXISTS power_core_size;

ALTER TABLE mech_skin
    ADD COLUMN level INT NOT NULL DEFAULT 0;

UPDATE mech_skin ms SET level = (select default_level from blueprint_mech_skin bms where ms.blueprint_id = bms.id);

ALTER TABLE weapons
    DROP COLUMN IF EXISTS slug,
    DROP COLUMN IF EXISTS damage,
    DROP COLUMN IF EXISTS default_damage_type,
    DROP COLUMN IF EXISTS damage_falloff,
    DROP COLUMN IF EXISTS damage_falloff_rate,
    DROP COLUMN IF EXISTS spread,
    DROP COLUMN IF EXISTS rate_of_fire,
    DROP COLUMN IF EXISTS projectile_speed,
    DROP COLUMN IF EXISTS radius,
    DROP COLUMN IF EXISTS radius_damage_falloff,
    DROP COLUMN IF EXISTS energy_cost,
    DROP COLUMN IF EXISTS is_melee,
    DROP COLUMN IF EXISTS max_ammo;
