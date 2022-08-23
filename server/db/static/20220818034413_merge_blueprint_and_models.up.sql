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
    ALTER COLUMN stat_modifier SET DEFAULT 0;

UPDATE blueprint_mech_skin SET stat_modifier = 0 WHERE stat_modifier IS NULL;

ALTER TABLE blueprint_mech_skin
    ALTER COLUMN stat_modifier SET NOT NULL,
    ALTER COLUMN stat_modifier TYPE INT;
ALTER TABLE blueprint_mech_skin
    RENAME COLUMN stat_modifier TO default_level;

-- weapons

ALTER TABLE blueprint_weapons
    DROP COLUMN default_damage_type;


ALTER TABLE weapon_models
    ADD COLUMN game_client_weapon_id TEXT,
    ADD COLUMN collection            COLLECTION DEFAULT 'supremacy-general' NOT NULL,
    ADD COLUMN damage                INTEGER                                NOT NULL default 0,
    ADD COLUMN default_damage_type   DAMAGE_TYPE                            NOT NULL DEFAULT 'Kinetic',
    ADD COLUMN damage_falloff        INT        DEFAULT 0,
    ADD COLUMN damage_falloff_rate   INT        DEFAULT 0,
    ADD COLUMN radius                INT        DEFAULT 0,
    ADD COLUMN radius_damage_falloff INT        DEFAULT 0,
    ADD COLUMN spread                NUMERIC    DEFAULT 0,
    ADD COLUMN rate_of_fire          NUMERIC    DEFAULT 0,
    ADD COLUMN projectile_speed      NUMERIC    DEFAULT 0,
    ADD COLUMN power_cost            NUMERIC    DEFAULT 0,
    ADD COLUMN power_instant_drain   bool       DEFAULT false,
    ADD COLUMN is_melee              BOOL                                   NOT NULL DEFAULT FALSE,
    ADD COLUMN max_ammo              INT        DEFAULT 0,
    ADD COLUMN projectile_amount     INT        DEFAULT 0,
    ADD COLUMN dot_tick_damage       NUMERIC    DEFAULT 0,
    ADD COLUMN dot_max_ticks         INT        DEFAULT 0,
    ADD COLUMN is_arced              bool       DEFAULT false,
    ADD COLUMN charge_time_seconds   NUMERIC    DEFAULT 0,
    ADD COLUMN burst_rate_of_fire    NUMERIC    DEFAULT 0;

ALTER TABLE blueprint_weapons
    RENAME TO blueprint_weapons_old;

ALTER TABLE weapon_models
    DROP COLUMN IF EXISTS repair_blocks;

ALTER TABLE weapon_models
    RENAME TO blueprint_weapons;

-- utilities
CREATE TABLE IF NOT EXISTS blueprint_modules
(
    id                UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    brand_id          UUID REFERENCES brands (id),

    slug              TEXT UNIQUE      NOT NULL,
    label             TEXT UNIQUE      NOT NULL,
    hitpoint_modifier INTEGER          NOT NULL,
    shield_modifier   INTEGER          NOT NULL,

    deleted_at        TIMESTAMPTZ,
    updated_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

ALTER TABLE blueprint_utility_shield
    DROP CONSTRAINT IF EXISTS blueprint_utility_shield_pkey,
    ADD CONSTRAINT blueprint_utility_shield_pkey PRIMARY KEY (blueprint_utility_id);
