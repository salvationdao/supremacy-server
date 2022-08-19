CREATE TABLE availabilities
(
    id           UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    reason       TEXT             NOT NULL,
    available_at TIMESTAMPTZ      NOT NULL
);

INSERT INTO availabilities (id, reason, available_at)
VALUES ('518ffb3f-8595-4db0-b9ea-46285f6ccd2f', 'Nexus Release',
        '2023-07-22 00:00:00'); -- TODO: move this to static data csv

-- mechs

DROP TYPE IF EXISTS POWERCORE_SIZE;
CREATE TYPE POWERCORE_SIZE AS ENUM ('SMALL', 'MEDIUM', 'LARGE');

DROP TYPE IF EXISTS BOOST_STAT;
CREATE TYPE BOOST_STAT AS ENUM ('MECH_HEALTH', 'MECH_SPEED', 'SHIELD_REGEN');

ALTER TABLE mech_models
    ADD COLUMN boost_stat        BOOST_STAT,
    ADD COLUMN weapon_hardpoints INTEGER                                NOT NULL DEFAULT 0,
    ADD COLUMN power_core_size   POWERCORE_SIZE                         NOT NULL DEFAULT 'SMALL',
    ADD COLUMN utility_slots     INTEGER                                NOT NULL DEFAULT 0,
    ADD COLUMN speed             INTEGER                                NOT NULL DEFAULT 0,
    ADD COLUMN max_hitpoints     INTEGER                                NOT NULL DEFAULT 0,
    ADD COLUMN collection        COLLECTION DEFAULT 'supremacy-general' NOT NULL,
    ADD COLUMN availability_id   UUID REFERENCES availabilities (id),
    ALTER COLUMN mech_type SET NOT NULL,
    ALTER COLUMN brand_id SET NOT NULL;

ALTER TABLE blueprint_mechs
    RENAME TO blueprint_mechs_old;

ALTER TABLE mech_models
    RENAME TO blueprint_mechs;

ALTER TABLE blueprint_mech_skin
    RENAME COLUMN stat_modifier TO default_level;

-- weapons

ALTER TABLE blueprint_weapons
    DROP COLUMN default_damage_type;

DROP TYPE IF EXISTS DAMAGE_TYPE;
CREATE TYPE DAMAGE_TYPE AS ENUM ('KINETIC', 'ENERGY', 'EXPLOSIVE');

ALTER TABLE weapon_models
    ADD COLUMN game_client_weapon_id TEXT,
    ADD COLUMN collection            COLLECTION DEFAULT 'supremacy-general' NOT NULL,
    ADD COLUMN damage                INTEGER                                NOT NULL default 0,
    ADD COLUMN default_damage_type   DAMAGE_TYPE                            NOT NULL DEFAULT 'KINETIC',
    ADD COLUMN damage_falloff        INT        DEFAULT 0,
    ADD COLUMN damage_falloff_rate   INT        DEFAULT 0,
    ADD COLUMN radius                INT        DEFAULT 0,
    ADD COLUMN radius_damage_falloff INT        DEFAULT 0,
    ADD COLUMN spread                NUMERIC    DEFAULT 0,
    ADD COLUMN rate_of_fire          NUMERIC    DEFAULT 0,
    ADD COLUMN projectile_speed      NUMERIC    DEFAULT 0,
    ADD COLUMN energy_cost           NUMERIC    DEFAULT 0,
    ADD COLUMN is_melee              BOOL                                   NOT NULL DEFAULT FALSE,
    ADD COLUMN max_ammo              INT        DEFAULT 0;

ALTER TABLE blueprint_weapons
    RENAME TO blueprint_weapons_old;

ALTER TABLE weapon_models
    DROP COLUMN IF EXISTS repair_blocks;

ALTER TABLE weapon_models
    RENAME TO blueprint_weapons;
