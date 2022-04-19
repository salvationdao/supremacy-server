CREATE TABLE blueprint_energy_cores
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    size          TEXT        NOT NULL DEFAULT 'MEDIUM' CHECK ( size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    capacity      NUMERIC     NOT NULL DEFAULT 0,
    max_draw_rate NUMERIC     NOT NULL DEFAULT 0,
    recharge_rate NUMERIC     NOT NULL DEFAULT 0,
    armour        NUMERIC     NOT NULL DEFAULT 0,
    max_hitpoints NUMERIC     NOT NULL DEFAULT 0,
    tier          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE energy_cores
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    owner_id      UUID        NOT NULL REFERENCES players (id),
    size          TEXT        NOT NULL DEFAULT 'MEDIUM' CHECK ( size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    capacity      NUMERIC     NOT NULL DEFAULT 0,
    max_draw_rate NUMERIC     NOT NULL DEFAULT 0,
    recharge_rate NUMERIC     NOT NULL DEFAULT 0,
    armour        NUMERIC     NOT NULL DEFAULT 0,
    max_hitpoints NUMERIC     NOT NULL DEFAULT 0,
    tier          TEXT,
    equipped_on   UUID REFERENCES chassis (id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE blueprint_chassis_skin
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    chassis_model TEXT        NOT NULL,
    tier          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chassis_skin
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    owner_id      UUID        NOT NULL REFERENCES players (id),

    chassis_model TEXT        NOT NULL,
    equipped_on   UUID REFERENCES chassis (id),
    tier          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chassis_animation
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    owner_id        UUID        NOT NULL REFERENCES players (id),
    chassis_model   TEXT        NOT NULL,
    equipped_on     UUID REFERENCES chassis (id),
    tier            TEXT,
    intro_animation BOOL                 DEFAULT TRUE,
    outro_animation BOOL                 DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE blueprint_chassis_animation
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    chassis_model   TEXT        NOT NULL,
    equipped_on     UUID REFERENCES chassis (id),
    tier            TEXT,
    intro_animation BOOL                 DEFAULT TRUE,
    outro_animation BOOL                 DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);



ALTER TABLE chassis
    DROP COLUMN IF EXISTS turret_hardpointsm,
    DROP COLUMN IF EXISTS health_remainingm,
    DROP COLUMN IF EXISTS shield_recharge_ratem,
    DROP COLUMN IF EXISTS max_shieldm,
    DROP COLUMN IF EXISTS turret_hardpointsm,
    ADD COLUMN owner_id         UUID REFERENCES players (id),
    ADD COLUMN energy_core_size TEXT NOT NULL DEFAULT 'MEDIUM' CHECK ( energy_core_size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    ADD COLUMN tier             TEXT;

WITH mech_owners AS (SELECT owner_id, chassis_id
                     FROM mechs)
UPDATE chassis c
SET owner_id = mech_owners.owner_id
FROM mech_owners
WHERE c.id = mech_owners.chassis_id;

ALTER TABLE chassis
    ALTER COLUMN owner_id SET NOT NULL;

ALTER TABLE blueprint_chassis
    DROP COLUMN IF EXISTS turret_hardpoints,
    DROP COLUMN IF EXISTS health_remaining,
    DROP COLUMN IF EXISTS shield_recharge_rate,
    DROP COLUMN IF EXISTS max_shield,
    DROP COLUMN IF EXISTS turret_hardpoints,
    ADD COLUMN energy_core_size TEXT NOT NULL DEFAULT 'MEDIUM' CHECK ( energy_core_size IN ('SMALL', 'MEDIUM', 'LARGE') ),
    ADD COLUMN tier             TEXT;

ALTER TABLE chassis
    ADD COLUMN chassis_skin_id    UUID REFERENCES chassis_skin (id),
    ADD COLUMN energy_core_id     UUID REFERENCES energy_cores (id),
    ADD COLUMN intro_animation_id UUID REFERENCES chassis_animation (id),
    ADD COLUMN outro_animation_id UUID REFERENCES chassis_animation (id);

ALTER TABLE blueprint_weapons
    DROP COLUMN IF EXISTS weapon_type,
    ADD COLUMN is_melee                BOOL DEFAULT FALSE,
    ADD COLUMN damage_falloff          INT  DEFAULT 0,
    ADD COLUMN damage_falloff_rate     INT  DEFAULT 0,
    ADD COLUMN spread                  INT  DEFAULT 0,
    ADD COLUMN rate_of_fire            INT  DEFAULT 0,
    ADD COLUMN magazine_size           INT  DEFAULT 0,
    ADD COLUMN reload_speed            INT  DEFAULT 0,
    ADD COLUMN radius                  INT  DEFAULT 0,
    ADD COLUMN radial_does_full_damage BOOL DEFAULT TRUE,
    ADD COLUMN projectile_speed        INT  DEFAULT 0,
    ADD COLUMN energy_cost             INT  DEFAULT 0,
    ADD COLUMN ammo_type               TEXT CHECK (ammo_type IN
                                                   ('GRENADE', 'BULLET', 'FLAK', 'FUEL CELL', 'MISSILE',
                                                    'ENERGY', 'ENERGY CELL', 'NONE'));

UPDATE blueprint_weapons
SET ammo_type = 'BULLET'
WHERE label = 'Sniper Rifle';
UPDATE blueprint_weapons
SET ammo_type = 'NONE',
    is_melee  = TRUE
WHERE label = 'Laser Sword';
UPDATE blueprint_weapons
SET ammo_type = 'MISSILE'
WHERE label = 'Rocket Pod';
UPDATE blueprint_weapons
SET ammo_type = 'BULLET'
WHERE label = 'Auto Cannon';
UPDATE blueprint_weapons
SET ammo_type = 'ENERGY CELL'
WHERE label = 'Plasma Rifle';
UPDATE blueprint_weapons
SET ammo_type = 'NONE',
    is_melee  = TRUE
WHERE label = 'Sword';

ALTER TABLE blueprint_weapons
    ALTER COLUMN ammo_type SET NOT NULL;

ALTER TABLE weapons
    DROP COLUMN IF EXISTS weapon_type,
    ADD COLUMN owner_id                UUID REFERENCES players (id),
    ADD COLUMN is_melee                BOOL DEFAULT FALSE,
    ADD COLUMN damage_falloff          INT  DEFAULT 0,
    ADD COLUMN damage_falloff_rate     INT  DEFAULT 0,
    ADD COLUMN spread                  INT  DEFAULT 0,
    ADD COLUMN rate_of_fire            INT  DEFAULT 0,
    ADD COLUMN magazine_size           INT  DEFAULT 0,
    ADD COLUMN reload_speed            INT  DEFAULT 0,
    ADD COLUMN radius                  INT  DEFAULT 0,
    ADD COLUMN radial_does_full_damage BOOL DEFAULT TRUE,
    ADD COLUMN projectile_speed        INT  DEFAULT 0,
    ADD COLUMN energy_cost             INT  DEFAULT 0,
    ADD COLUMN ammo_type               TEXT CHECK (ammo_type IN
                                                   ('GRENADE', 'BULLET', 'FLAK', 'FUEL CELL', 'MISSILE',
                                                    'ENERGY', 'ENERGY CELL', 'NONE'));
UPDATE weapons
SET ammo_type = 'BULLET'
WHERE label = 'Sniper Rifle';
UPDATE weapons
SET ammo_type = 'NONE',
    is_melee  = TRUE
WHERE label = 'Laser Sword';
UPDATE weapons
SET ammo_type = 'MISSILE'
WHERE label = 'Rocket Pod';
UPDATE weapons
SET ammo_type = 'BULLET'
WHERE label = 'Auto Cannon';
UPDATE weapons
SET ammo_type = 'ENERGY CELL'
WHERE label = 'Plasma Rifle';
UPDATE weapons
SET ammo_type = 'NONE',
    is_melee  = TRUE
WHERE label = 'Sword';

ALTER TABLE weapons
    ALTER COLUMN ammo_type SET NOT NULL;

WITH weapon_owners AS (SELECT m.owner_id, cw.weapon_id
                       FROM chassis_weapons cw
                                INNER JOIN mechs m ON cw.chassis_id = m.chassis_id)
UPDATE weapons w
SET owner_id = weapon_owners.owner_id
FROM weapon_owners
WHERE w.id = weapon_owners.weapon_id;

ALTER TABLE weapons
    ALTER COLUMN owner_id SET NOT NULL;


CREATE TABLE blueprint_ammo
(
    id                             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    type                           TEXT        NOT NULL CHECK (type IN
                                                               ('GRENADE', 'BULLET', 'FLAK', 'FUEL CELL', 'MISSILE',
                                                                'ENERGY', 'ENERGY CELL')),
    damage_multiplier              NUMERIC              DEFAULT 0,
    damage_falloff_multiplier      NUMERIC              DEFAULT 0,
    damage_falloff_rate_multiplier NUMERIC              DEFAULT 0,
    spread_multiplier              NUMERIC              DEFAULT 0,
    rate_of_fire_multiplier        NUMERIC              DEFAULT 0,
    radius_multiplier              NUMERIC              DEFAULT 0,
    projectile_speed_multiplier    NUMERIC              DEFAULT 0,
    energy_cost_multiplier         NUMERIC              DEFAULT 0,
    created_at                     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ammo
(
    blueprint_id UUID        NOT NULL REFERENCES blueprint_ammo (id),
    owner_id     UUID        NOT NULL REFERENCES players (id),
    count        INT         NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blueprint_id, owner_id)
);

CREATE TABLE weapon_ammo
(
    blueprint_ammo_id UUID        NOT NULL REFERENCES blueprint_ammo (id),
    weapon_id         UUID        NOT NULL REFERENCES weapons (id),
    count             INT         NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blueprint_ammo_id, weapon_id)
);

ALTER TABLE blueprint_chassis_blueprint_modules
    RENAME TO blueprint_chassis_blueprint_utility;
ALTER TABLE blueprint_chassis_blueprint_utility
    RENAME COLUMN blueprint_module_id TO blueprint_utility_id;
ALTER TABLE blueprint_modules
    RENAME TO blueprint_utility;
ALTER TABLE blueprint_utility
    DROP COLUMN hitpoint_modifier,
    DROP COLUMN shield_modifier,
    ADD COLUMN type TEXT CHECK (type IN ('SHIELD', 'ATTACK DRONE', 'REPAIR DRONE', 'ANTI MISSILE',
                                         'ACCELERATOR'));
UPDATE blueprint_utility
SET type = 'SHIELD';
ALTER TABLE blueprint_utility
    ALTER COLUMN type SET NOT NULL;


CREATE TABLE blueprint_utility_shield
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_utility_id UUID        NOT NULL REFERENCES blueprint_utility (id),
    hitpoints            INT         NOT NULL DEFAULT 0,
    recharge_rate        INT         NOT NULL DEFAULT 0,
    recharge_energy_cost INT         NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE blueprint_utility_attack_drone
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_utility_id UUID        NOT NULL REFERENCES blueprint_utility (id),
    damage               INT         NOT NULL,
    rate_of_fire         INT         NOT NULL,
    hitpoints            INT         NOT NULL,
    lifespan_seconds     INT         NOT NULL,
    deploy_energy_cost   INT         NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE blueprint_utility_repair_drone
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_utility_id UUID        NOT NULL REFERENCES blueprint_utility (id),
    repair_type          TEXT CHECK (repair_type IN ('SHIELD', 'STRUCTURE')),
    repair_amount        INT         NOT NULL,
    deploy_energy_cost   INT         NOT NULL,
    lifespan_seconds     INT         NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE blueprint_utility_anti_missile
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_utility_id UUID        NOT NULL REFERENCES blueprint_utility (id),
    rate_of_fire         INT         NOT NULL,
    fire_energy_cost     INT         NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE blueprint_utility_accelerator
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    blueprint_utility_id UUID        NOT NULL REFERENCES blueprint_utility (id),
    energy_cost          INT         NOT NULL,
    boost_seconds        INT         NOT NULL,
    boost_amount         INT         NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE chassis_modules
    RENAME TO chassis_utility;
ALTER TABLE chassis_utility
    RENAME COLUMN module_id TO utility_id;

ALTER TABLE modules
    RENAME TO utility;
ALTER TABLE utility
    DROP COLUMN hitpoint_modifier,
    DROP COLUMN shield_modifier,
    ADD COLUMN owner_id    UUID REFERENCES players (id),
    ADD COLUMN equipped_on UUID REFERENCES chassis (id),
    ADD COLUMN type        TEXT CHECK (type IN ('SHIELD', 'ATTACK DRONE', 'REPAIR DRONE', 'ANTI MISSILE',
                                                'ACCELERATOR'));

WITH utility_owners AS (SELECT m.owner_id, cu.utility_id
                        FROM chassis_utility cu
                                 INNER JOIN mechs m ON cu.chassis_id = m.chassis_id)
UPDATE utility u
SET owner_id = utility_owners.owner_id
FROM utility_owners
WHERE u.id = utility_owners.utility_id;

ALTER TABLE utility
    ALTER COLUMN owner_id SET NOT NULL;

UPDATE utility
SET type = 'SHIELD';
ALTER TABLE blueprint_utility
    ALTER COLUMN type SET NOT NULL;


CREATE TABLE utility_shield
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    utility_id           UUID        NOT NULL REFERENCES utility (id),
    hitpoints            INT         NOT NULL DEFAULT 0,
    recharge_rate        INT         NOT NULL DEFAULT 0,
    recharge_energy_cost INT         NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE utility_attack_drone
(
    id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    utility_id         UUID        NOT NULL REFERENCES utility (id),
    damage             INT         NOT NULL,
    rate_of_fire       INT         NOT NULL,
    hitpoints          INT         NOT NULL,
    lifespan_seconds   INT         NOT NULL,
    deploy_energy_cost INT         NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE utility_repair_drone
(
    id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    utility_id         UUID        NOT NULL REFERENCES utility (id),
    repair_type        TEXT CHECK (repair_type IN ('SHIELD', 'STRUCTURE')),
    repair_amount      INT         NOT NULL,
    deploy_energy_cost INT         NOT NULL,
    lifespan_seconds   INT         NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE utility_anti_missile
(
    id               UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    utility_id       UUID        NOT NULL REFERENCES utility (id),
    rate_of_fire     INT         NOT NULL,
    fire_energy_cost INT         NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE utility_accelerator
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    utility_id    UUID        NOT NULL REFERENCES utility (id),
    energy_cost   INT         NOT NULL,
    boost_seconds INT         NOT NULL,
    boost_amount  INT         NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);