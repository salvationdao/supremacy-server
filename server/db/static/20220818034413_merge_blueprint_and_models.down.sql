ALTER TABLE blueprint_utility_shield
    DROP CONSTRAINT IF EXISTS blueprint_utility_shield_pkey,
    ADD CONSTRAINT blueprint_utility_shield_pkey PRIMARY KEY (id);

DROP TABLE IF EXISTS blueprint_modules;

ALTER TABLE blueprint_weapons
    RENAME TO weapon_models;

ALTER TABLE weapon_models
    ADD COLUMN IF NOT EXISTS repair_blocks INT NOT NULL DEFAULT 20;

ALTER TABLE blueprint_weapons_old
    RENAME TO blueprint_weapons;

ALTER TABLE weapon_models
    DROP COLUMN IF EXISTS game_client_weapon_id,
    DROP COLUMN IF EXISTS collection,
    DROP COLUMN IF EXISTS damage,
    DROP COLUMN IF EXISTS default_damage_type,
    DROP COLUMN IF EXISTS damage_falloff,
    DROP COLUMN IF EXISTS damage_falloff_rate,
    DROP COLUMN IF EXISTS radius,
    DROP COLUMN IF EXISTS radius_damage_falloff,
    DROP COLUMN IF EXISTS spread,
    DROP COLUMN IF EXISTS rate_of_fire,
    DROP COLUMN IF EXISTS projectile_speed,
    DROP COLUMN IF EXISTS power_cost,
    DROP COLUMN IF EXISTS power_instant_drain,
    DROP COLUMN IF EXISTS is_melee,
    DROP COLUMN IF EXISTS max_ammo,
    DROP COLUMN IF EXISTS projectile_amount,
    DROP COLUMN IF EXISTS dot_tick_damage,
    DROP COLUMN IF EXISTS dot_max_ticks,
    DROP COLUMN IF EXISTS is_arced,
    DROP COLUMN IF EXISTS charge_time_seconds,
    DROP COLUMN IF EXISTS burst_rate_of_fire;

ALTER TABLE blueprint_weapons
    ADD COLUMN default_damage_type   DAMAGE_TYPE                            NOT NULL DEFAULT 'Kinetic';

ALTER TABLE blueprint_mech_skin
    RENAME COLUMN default_level TO stat_modifier;

ALTER TABLE blueprint_mech_skin
    ALTER COLUMN stat_modifier DROP NOT NULL,
    ALTER COLUMN stat_modifier TYPE NUMERIC(8, 0);

UPDATE blueprint_mech_skin SET stat_modifier = NULL WHERE stat_modifier = 0;

ALTER TABLE blueprint_mech_skin
    ALTER COLUMN stat_modifier DROP DEFAULT;

ALTER TABLE blueprint_mechs
    RENAME TO mech_models;

ALTER TABLE blueprint_mechs_old
    RENAME TO blueprint_mechs;

ALTER TABLE mech_models
    DROP COLUMN IF EXISTS boost_stat,
    DROP COLUMN IF EXISTS weapon_hardpoints,
    DROP COLUMN IF EXISTS power_core_size,
    DROP COLUMN IF EXISTS utility_slots,
    DROP COLUMN IF EXISTS speed,
    DROP COLUMN IF EXISTS max_hitpoints,
    DROP COLUMN IF EXISTS collection,
    DROP COLUMN IF EXISTS availability_id,
    ALTER COLUMN mech_type DROP NOT NULL,
    ALTER COLUMN brand_id DROP NOT NULL;

DROP TABLE IF EXISTS availabilities;
