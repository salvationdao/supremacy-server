CREATE TABLE availabilities (
                                id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
                                reason TEXT NOT NULL,
                                available_at TIMESTAMPTZ NOT NULL
);

INSERT INTO availabilities (id, reason, available_at) VALUES
    ('518ffb3f-8595-4db0-b9ea-46285f6ccd2f', 'Nexus Release', '2023-07-22 00:00:00'); -- TODO: move this to static data csv

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


-- ALTER TABLE chassis
--     DROP CONSTRAINT chassis_blueprint_id_fkey,
--     ADD CONSTRAINT chassis_blueprint_id_fkey FOREIGN KEY (blueprint_id) REFERENCES blueprint_mechs (id);
