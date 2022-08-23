CREATE TABLE IF NOT EXISTS battle_abilities
(
    id                       uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    label                    TEXT                                       NOT NULL,
    cooldown_duration_second INTEGER                                    NOT NULL,
    description              TEXT                                       NOT NULL
);

CREATE TABLE IF NOT EXISTS factions
(
    id               uuid PRIMARY KEY         DEFAULT gen_random_uuid()           NOT NULL,
    vote_price       TEXT                     DEFAULT '1000000000000000000'::TEXT NOT NULL,
    contract_reward  TEXT                     DEFAULT '1000000000000000000'::TEXT NOT NULL,
    label            TEXT                                                         NOT NULL,
    guild_id         uuid,
    deleted_at       TIMESTAMP WITH TIME ZONE,
    updated_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()                       NOT NULL,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()                       NOT NULL,
    primary_color    TEXT                     DEFAULT '#000000'::TEXT             NOT NULL,
    secondary_color  TEXT                     DEFAULT '#ffffff'::TEXT             NOT NULL,
    background_color TEXT                     DEFAULT '#0D0D0D'::TEXT             NOT NULL,
    logo_url         TEXT                     DEFAULT ''::TEXT                    NOT NULL,
    background_url   TEXT                     DEFAULT ''::TEXT                    NOT NULL,
    description      TEXT                     DEFAULT ''::TEXT                    NOT NULL
);


CREATE TABLE IF NOT EXISTS brands
(
    id         uuid PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    faction_id uuid                                               NOT NULL REFERENCES factions (id),
    label      TEXT                                               NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()             NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()             NOT NULL
);

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'mech_type') THEN
            CREATE TYPE mech_type AS ENUM ('HUMANOID', 'PLATFORM');
        END IF;
    END
$$;

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'collection') THEN
            CREATE TYPE collection AS ENUM ('supremacy-ai','supremacy-genesis', 'supremacy-limited-release', 'supremacy-general', 'supremacy-consumables','supremacy-achievements');
        END IF;
    END
$$;


CREATE TABLE IF NOT EXISTS blueprint_mech_skin
(
    id                 uuid PRIMARY KEY         DEFAULT gen_random_uuid()   NOT NULL,
    collection         collection               DEFAULT 'supremacy-general' NOT NULL,
    mech_model         uuid                                                 NOT NULL,
    label              TEXT                                                 NOT NULL,
    tier               TEXT                     DEFAULT 'MEGA'::TEXT        NOT NULL,
    image_url          TEXT,
    animation_url      TEXT,
    card_animation_url TEXT,
    large_image_url    TEXT,
    avatar_url         TEXT,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()               NOT NULL,
    background_color   TEXT,
    youtube_url        TEXT,
    mech_type          mech_type,
    stat_modifier      NUMERIC(8, 0)
);

CREATE TABLE IF NOT EXISTS mech_models
(
    id                      uuid PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    label                   TEXT                                               NOT NULL,
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT NOW()             NOT NULL,
    default_chassis_skin_id uuid                                               NOT NULL REFERENCES blueprint_mech_skin (id),
    brand_id                uuid,
    mech_type               mech_type
);



CREATE TABLE IF NOT EXISTS blueprint_power_cores
(
    id                 uuid PRIMARY KEY         DEFAULT gen_random_uuid()   NOT NULL,
    collection         collection               DEFAULT 'supremacy-general' NOT NULL,
    label              TEXT                                                 NOT NULL,
    size               TEXT                     DEFAULT 'MEDIUM'::TEXT      NOT NULL,
    capacity           NUMERIC                  DEFAULT 0                   NOT NULL,
    max_draw_rate      NUMERIC                  DEFAULT 0                   NOT NULL,
    recharge_rate      NUMERIC                  DEFAULT 0                   NOT NULL,
    armour             NUMERIC                  DEFAULT 0                   NOT NULL,
    max_hitpoints      NUMERIC                  DEFAULT 0                   NOT NULL,
    tier               TEXT                     DEFAULT 'MEGA'::TEXT        NOT NULL,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()               NOT NULL,
    image_url          TEXT,
    card_animation_url TEXT,
    avatar_url         TEXT,
    large_image_url    TEXT,
    background_color   TEXT,
    animation_url      TEXT,
    youtube_url        TEXT,
    CONSTRAINT blueprint_power_cores_size_check CHECK ((size = ANY (ARRAY ['SMALL'::TEXT, 'MEDIUM'::TEXT, 'LARGE'::TEXT])))
);

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'weapon_type') THEN
            CREATE TYPE weapon_type AS ENUM ('Grenade Launcher', 'Cannon', 'Minigun', 'Plasma Gun', 'Flak',
                'Machine Gun', 'Flamethrower', 'Missile Launcher', 'Laser Beam',
                'Lightning Gun', 'BFG', 'Rifle', 'Sniper Rifle', 'Sword');
        END IF;
    END
$$;



CREATE TABLE IF NOT EXISTS weapon_models
(
    id              uuid PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    brand_id        uuid REFERENCES brands (id),
    label           TEXT                                               NOT NULL,
    weapon_type     weapon_type                                        NOT NULL,
    default_skin_id uuid                                               NOT NULL,
    deleted_at      TIMESTAMP WITH TIME ZONE,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()             NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()             NOT NULL
);


CREATE TABLE IF NOT EXISTS blueprint_weapon_skin
(
    id                 uuid PRIMARY KEY         DEFAULT gen_random_uuid()   NOT NULL,
    label              TEXT                                                 NOT NULL,
    weapon_type        weapon_type                                          NOT NULL,
    tier               TEXT                     DEFAULT 'MEGA'::TEXT        NOT NULL,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()               NOT NULL,
    image_url          TEXT,
    card_animation_url TEXT,
    avatar_url         TEXT,
    large_image_url    TEXT,
    background_color   TEXT,
    animation_url      TEXT,
    youtube_url        TEXT,
    collection         TEXT                     DEFAULT 'supremacy-general' NOT NULL,
    weapon_model_id    uuid                                                 NOT NULL,
    stat_modifier      NUMERIC(8, 0)
);

-- ALTER TABLE weapon_models
--     ADD FOREIGN KEY ( default_skin_id) REFERENCES blueprint_weapon_skin(id);


DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'ability_level') THEN
            CREATE TYPE ability_level AS ENUM ('MECH','FACTION','PLAYER');
        END IF;
    END
$$;

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'location_select_type_enum') THEN
            CREATE TYPE location_select_type_enum AS ENUM (
                'LINE_SELECT',
                'MECH_SELECT',
                'LOCATION_SELECT',
                'GLOBAL'
                );
        END IF;
    END
$$;


CREATE TABLE IF NOT EXISTS game_abilities
(
    id                     uuid PRIMARY KEY          DEFAULT gen_random_uuid() NOT NULL,
    game_client_ability_id INTEGER                                             NOT NULL,
    faction_id             uuid                                                NOT NULL REFERENCES factions (id),
    battle_ability_id      uuid REFERENCES battle_abilities (id),
    label                  TEXT                                                NOT NULL,
    colour                 TEXT                                                NOT NULL,
    image_url              TEXT                                                NOT NULL,
    sups_cost              TEXT                      DEFAULT '0'::TEXT         NOT NULL,
    description            TEXT                                                NOT NULL,
    text_colour            TEXT                                                NOT NULL,
    current_sups           TEXT                      DEFAULT '0'::TEXT         NOT NULL,
    level                  ability_level             DEFAULT 'FACTION'         NOT NULL,
    location_select_type   location_select_type_enum DEFAULT 'LOCATION_SELECT' NOT NULL
);


DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'crate_type') THEN
            CREATE TYPE crate_type AS ENUM ('MECH', 'WEAPON');
        END IF;
    END
$$;

CREATE TABLE IF NOT EXISTS blueprint_mechs
(
    id                uuid PRIMARY KEY         DEFAULT gen_random_uuid()   NOT NULL,
    brand_id          uuid                                                 NOT NULL,
    label             TEXT                                                 NOT NULL,
    slug              TEXT                                                 NOT NULL,
    weapon_hardpoints INTEGER                                              NOT NULL,
    utility_slots     INTEGER                                              NOT NULL,
    speed             INTEGER                                              NOT NULL,
    max_hitpoints     INTEGER                                              NOT NULL,
    deleted_at        TIMESTAMP WITH TIME ZONE,
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW()               NOT NULL,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW()               NOT NULL,
    model_id          uuid                                                 NOT NULL,
    collection        collection               DEFAULT 'supremacy-general' NOT NULL,
    power_core_size   TEXT                     DEFAULT 'SMALL'             NOT NULL,
    tier              TEXT                     DEFAULT 'MEGA'              NOT NULL,
    CONSTRAINT blueprint_chassis_power_core_size_check CHECK ((power_core_size = ANY
                                                               (ARRAY ['SMALL'::TEXT, 'MEDIUM'::TEXT, 'LARGE'::TEXT])))
);

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'damage_type') THEN
            CREATE TYPE damage_type AS ENUM ('Kinetic', 'Energy', 'Explosive');
        END IF;
    END
$$;


CREATE TABLE IF NOT EXISTS blueprint_weapons
(
    id                    uuid PRIMARY KEY         DEFAULT gen_random_uuid()   NOT NULL,
    brand_id              uuid REFERENCES brands (id),
    label                 TEXT                                                 NOT NULL,
    slug                  TEXT                                                 NOT NULL,
    damage                INTEGER                                              NOT NULL,
    deleted_at            TIMESTAMP WITH TIME ZONE,
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW()               NOT NULL,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW()               NOT NULL,
    game_client_weapon_id uuid,
    weapon_type           weapon_type                                          NOT NULL,
    collection            collection               DEFAULT 'supremacy-general' NOT NULL,
    default_damage_type   damage_type              DEFAULT 'Kinetic'           NOT NULL,
    damage_falloff        INTEGER                  DEFAULT 0,
    damage_falloff_rate   INTEGER                  DEFAULT 0,
    radius                INTEGER                  DEFAULT 0,
    radius_damage_falloff INTEGER                  DEFAULT 0,
    spread                NUMERIC                  DEFAULT 0,
    rate_of_fire          NUMERIC                  DEFAULT 0,
    projectile_speed      NUMERIC                  DEFAULT 0,
    max_ammo              INTEGER                  DEFAULT 0,
    is_melee              BOOLEAN                  DEFAULT FALSE               NOT NULL,
    tier                  TEXT                     DEFAULT 'MEGA'              NOT NULL,
    energy_cost           NUMERIC                  DEFAULT 0,
    weapon_model_id       uuid                                                 NOT NULL
);

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'utility_type') THEN
            CREATE TYPE UTILITY_TYPE AS ENUM ('SHIELD', 'ATTACK DRONE', 'REPAIR DRONE', 'ANTI MISSILE', 'ACCELERATOR');
        END IF;
    END
$$;


CREATE TABLE IF NOT EXISTS blueprint_utility
(
    id                 uuid PRIMARY KEY         DEFAULT gen_random_uuid()               NOT NULL,
    brand_id           uuid REFERENCES brands (id),
    label              TEXT                                                             NOT NULL,
    deleted_at         TIMESTAMP WITH TIME ZONE,
    updated_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()                           NOT NULL,
    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()                           NOT NULL,
    type               utility_type                                                     NOT NULL,
    tier               TEXT                     DEFAULT 'MEGA'::TEXT                    NOT NULL,
    collection         collection               DEFAULT 'supremacy-general'::collection NOT NULL,
    image_url          TEXT,
    card_animation_url TEXT,
    avatar_url         TEXT,
    large_image_url    TEXT,
    background_color   TEXT,
    animation_url      TEXT,
    youtube_url        TEXT
);

CREATE TABLE IF NOT EXISTS blueprint_utility_shield
(
    id                   uuid PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    blueprint_utility_id uuid REFERENCES blueprint_utility (id),
    hitpoints            INTEGER                  DEFAULT 0                 NOT NULL,
    recharge_rate        INTEGER                  DEFAULT 0                 NOT NULL,
    recharge_energy_cost INTEGER                  DEFAULT 0                 NOT NULL,
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT NOW()             NOT NULL
);

CREATE TABLE IF NOT EXISTS availabilities
(
    id           UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    reason       TEXT             NOT NULL,
    available_at TIMESTAMPTZ      NOT NULL
);

INSERT INTO availabilities (id, reason, available_at)
VALUES ('518ffb3f-8595-4db0-b9ea-46285f6ccd2f', 'Nexus Release',
        '2023-07-22 00:00:00') on conflict do nothing;
-- TODO: move this to static data csv
