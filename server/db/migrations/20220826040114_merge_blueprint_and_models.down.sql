ALTER TABLE utility
    RENAME COLUMN label_dont_use TO label;
ALTER TABLE utility
    RENAME COLUMN brand_dont_use TO brand_id;

ALTER TABLE utility_shield_dont_use RENAME TO utility_shield;

UPDATE template_blueprints SET blueprint_id = blueprint_id_old WHERE blueprint_id_old IS NOT NULL;

UPDATE blueprint_utility
SET deleted_at = NULL
WHERE id NOT IN (
                 'd429be75-6f98-4231-8315-a86db8477d05',
                 '1e9a8bd4-b6c3-4a46-86e9-4c68a95f09b8',
                 '0551d044-b8ff-47ac-917e-80c3fce37378'
    );


UPDATE blueprint_utility_shield_old
SET deleted_at = NULL
WHERE blueprint_utility_id NOT IN (
                                   'd429be75-6f98-4231-8315-a86db8477d05',
                                   '1e9a8bd4-b6c3-4a46-86e9-4c68a95f09b8',
                                   '0551d044-b8ff-47ac-917e-80c3fce37378'
    );

UPDATE utility SET blueprint_id = blueprint_id_old WHERE blueprint_id_old IS NOT NULL;

ALTER TABLE utility
    DROP COLUMN blueprint_id_old;

ALTER TABLE weapons
    DROP CONSTRAINT weapons_blueprint_id_fkey;

UPDATE weapons SET blueprint_id = blueprint_id_old WHERE blueprint_id_old IS NOT NULL;

ALTER TABLE weapons
    ADD CONSTRAINT weapons_blueprint_id_fkey FOREIGN KEY (blueprint_id) REFERENCES blueprint_weapons_old (id),
    DROP COLUMN blueprint_id_old;

ALTER TABLE mechs
    DROP CONSTRAINT chassis_blueprint_id_fkey;

UPDATE mechs SET blueprint_id = blueprint_id_old WHERE blueprint_id_old IS NOT NULL;

ALTER TABLE mechs
    ADD CONSTRAINT chassis_blueprint_id_fkey FOREIGN KEY (blueprint_id) REFERENCES blueprint_mechs_old (id),
    DROP COLUMN blueprint_id_old;

UPDATE mystery_crate_blueprints SET blueprint_id = blueprint_id_old WHERE blueprint_id_old IS NOT NULL;

ALTER TABLE mystery_crate_blueprints
    DROP COLUMN blueprint_id_old;

ALTER TABLE template_blueprints
    DROP COLUMN blueprint_id_old;

ALTER TABLE weapons
    RENAME COLUMN slug_dont_use TO slug;
ALTER TABLE weapons
    RENAME COLUMN damage_dont_use TO damage;
ALTER TABLE weapons
    RENAME COLUMN default_damage_type_dont_use TO default_damage_type;
ALTER TABLE weapons
    RENAME COLUMN damage_falloff_dont_use TO damage_falloff;
ALTER TABLE weapons
    RENAME COLUMN damage_falloff_rate_dont_use TO damage_falloff_rate;
ALTER TABLE weapons
    RENAME COLUMN spread_dont_use TO spread;
ALTER TABLE weapons
    RENAME COLUMN rate_of_fire_dont_use TO rate_of_fire;
ALTER TABLE weapons
    RENAME COLUMN projectile_speed_dont_use TO projectile_speed;
ALTER TABLE weapons
    RENAME COLUMN radius_dont_use TO radius;
ALTER TABLE weapons
    RENAME COLUMN radius_damage_falloff_dont_use TO radius_damage_falloff;
ALTER TABLE weapons
    RENAME COLUMN energy_cost_dont_use TO energy_cost;
ALTER TABLE weapons
    RENAME COLUMN is_melee_dont_use TO is_melee;
ALTER TABLE weapons
    RENAME COLUMN max_ammo_dont_use TO max_ammo;

ALTER TABLE mech_skin
    DROP COLUMN level;

ALTER TABLE mechs
    RENAME COLUMN weapon_hardpoints_dont_use TO weapon_hardpoints;
ALTER TABLE mechs
    RENAME COLUMN utility_slots_dont_use TO utility_slots;
ALTER TABLE mechs
    RENAME COLUMN speed_dont_use TO speed;
ALTER TABLE mechs
    RENAME COLUMN max_hitpoints_dont_use TO max_hitpoints;
ALTER TABLE mechs
    RENAME COLUMN power_core_size_dont_use TO power_core_size;
