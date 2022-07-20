CREATE TABLE IF NOT EXISTS battle_abilities
(
    id                       uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    label                    text                                       NOT NULL,
    cooldown_duration_second integer                                    NOT NULL,
    description              text                                       NOT NULL
);

CREATE TABLE IF NOT EXISTS factions
(
    id               uuid PRIMARY KEY         DEFAULT gen_random_uuid()           NOT NULL,
    vote_price       text                     DEFAULT '1000000000000000000'::text NOT NULL,
    contract_reward  text                     DEFAULT '1000000000000000000'::text NOT NULL,
    label            text                                                         NOT NULL,
    guild_id         uuid,
    deleted_at       timestamp with time zone,
    updated_at       timestamp with time zone DEFAULT now()                       NOT NULL,
    created_at       timestamp with time zone DEFAULT now()                       NOT NULL,
    primary_color    text                     DEFAULT '#000000'::text             NOT NULL,
    secondary_color  text                     DEFAULT '#ffffff'::text             NOT NULL,
    background_color text                     DEFAULT '#0D0D0D'::text             NOT NULL,
    logo_url         text                     DEFAULT ''::text                    NOT NULL,
    background_url   text                     DEFAULT ''::text                    NOT NULL,
    description      text                     DEFAULT ''::text                    NOT NULL
);


CREATE TABLE IF NOT EXISTS brands
(
    id         uuid PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    faction_id uuid                                               NOT NULL REFERENCES factions (id),
    label      text                                               NOT NULL,
    deleted_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now()             NOT NULL,
    created_at timestamp with time zone DEFAULT now()             NOT NULL
);

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'mech_type') THEN
            CREATE TYPE MECH_TYPE AS ENUM ('HUMANOID', 'PLATFORM');
        END IF;
    END
$$;

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'collection') THEN
            CREATE TYPE COLLECTION AS ENUM ('supremacy-ai','supremacy-genesis', 'supremacy-limited-release', 'supremacy-general', 'supremacy-consumables');
        END IF;
    END
$$;


CREATE TABLE IF NOT EXISTS blueprint_mech_skin
(
    id                 uuid PRIMARY KEY         DEFAULT gen_random_uuid()   NOT NULL,
    collection         COLLECTION               DEFAULT 'supremacy-general' NOT NULL,
    mech_model         uuid                                                 NOT NULL,
    label              text                                                 NOT NULL,
    tier               text                     DEFAULT 'MEGA'::text        NOT NULL,
    image_url          text,
    animation_url      text,
    card_animation_url text,
    large_image_url    text,
    avatar_url         text,
    created_at         timestamp with time zone DEFAULT now()               NOT NULL,
    background_color   text,
    youtube_url        text,
    mech_type          MECH_TYPE,
    stat_modifier      numeric(8, 0)
);

CREATE TABLE IF NOT EXISTS mech_models
(
    id                      uuid PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    label                   text                                               NOT NULL,
    created_at              timestamp with time zone DEFAULT now()             NOT NULL,
    default_chassis_skin_id uuid                                               NOT NULL REFERENCES blueprint_mech_skin (id),
    brand_id                uuid,
    mech_type               MECH_TYPE
);



CREATE TABLE IF NOT EXISTS blueprint_power_cores
(
    id                 uuid PRIMARY KEY         DEFAULT gen_random_uuid()   NOT NULL,
    collection         COLLECTION               DEFAULT 'supremacy-general' NOT NULL,
    label              text                                                 NOT NULL,
    size               text                     DEFAULT 'MEDIUM'::text      NOT NULL,
    capacity           numeric                  DEFAULT 0                   NOT NULL,
    max_draw_rate      numeric                  DEFAULT 0                   NOT NULL,
    recharge_rate      numeric                  DEFAULT 0                   NOT NULL,
    armour             numeric                  DEFAULT 0                   NOT NULL,
    max_hitpoints      numeric                  DEFAULT 0                   NOT NULL,
    tier               text                     DEFAULT 'MEGA'::text        NOT NULL,
    created_at         timestamp with time zone DEFAULT now()               NOT NULL,
    image_url          text,
    card_animation_url text,
    avatar_url         text,
    large_image_url    text,
    background_color   text,
    animation_url      text,
    youtube_url        text,
    CONSTRAINT blueprint_power_cores_size_check CHECK ((size = ANY (ARRAY ['SMALL'::text, 'MEDIUM'::text, 'LARGE'::text])))
);

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'weapon_type') THEN
            CREATE TYPE WEAPON_TYPE AS ENUM ('Grenade Launcher', 'Cannon', 'Minigun', 'Plasma Gun', 'Flak',
                'Machine Gun', 'Flamethrower', 'Missile Launcher', 'Laser Beam',
                'Lightning Gun', 'BFG', 'Rifle', 'Sniper Rifle', 'Sword');
        END IF;
    END
$$;



CREATE TABLE IF NOT EXISTS weapon_models
(
    id              uuid PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    brand_id        uuid REFERENCES brands (id),
    label           text                                               NOT NULL,
    weapon_type     WEAPON_TYPE                                        NOT NULL,
    default_skin_id uuid                                               NOT NULL,
    deleted_at      timestamp with time zone,
    updated_at      timestamp with time zone DEFAULT now()             NOT NULL,
    created_at      timestamp with time zone DEFAULT now()             NOT NULL
);


CREATE TABLE IF NOT EXISTS blueprint_weapon_skin
(
    id                 uuid PRIMARY KEY         DEFAULT gen_random_uuid()   NOT NULL,
    label              text                                                 NOT NULL,
    weapon_type        WEAPON_TYPE                                          NOT NULL,
    tier               text                     DEFAULT 'MEGA'::text        NOT NULL,
    created_at         timestamp with time zone DEFAULT now()               NOT NULL,
    image_url          text,
    card_animation_url text,
    avatar_url         text,
    large_image_url    text,
    background_color   text,
    animation_url      text,
    youtube_url        text,
    collection         text                     DEFAULT 'supremacy-general' NOT NULL,
    weapon_model_id    uuid                                                 NOT NULL,
    stat_modifier      numeric(8, 0)
);

-- ALTER TABLE weapon_models
--     ADD FOREIGN KEY ( default_skin_id) REFERENCES blueprint_weapon_skin(id);


DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'ability_level') THEN
            CREATE TYPE ABILITY_LEVEL AS ENUM ('MECH','FACTION','PLAYER');
        END IF;
    END
$$;

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'location_select_type_enum') THEN
            CREATE TYPE LOCATION_SELECT_TYPE_ENUM AS ENUM (
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
    game_client_ability_id integer                                             NOT NULL,
    faction_id             uuid                                                NOT NULL REFERENCES factions (id),
    battle_ability_id      uuid,
    label                  text                                                NOT NULL,
    colour                 text                                                NOT NULL,
    image_url              text                                                NOT NULL,
    sups_cost              text                      DEFAULT '0'::text         NOT NULL,
    description            text                                                NOT NULL,
    text_colour            text                                                NOT NULL,
    current_sups           text                      DEFAULT '0'::text         NOT NULL,
    level                  ABILITY_LEVEL             DEFAULT 'FACTION'         NOT NULL,
    location_select_type   LOCATION_SELECT_TYPE_ENUM DEFAULT 'LOCATION_SELECT' NOT NULL
);


DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'crate_type') THEN
            CREATE TYPE CRATE_TYPE AS ENUM ('MECH', 'WEAPON');
        END IF;
    END
$$;

CREATE TABLE IF NOT EXISTS blueprint_mechs
(
    id                uuid                     PRIMARY KEY DEFAULT gen_random_uuid()   NOT NULL,
    brand_id          uuid                                                 NOT NULL,
    label             text                                                 NOT NULL,
    slug              text                                                 NOT NULL,
    weapon_hardpoints integer                                              NOT NULL,
    utility_slots     integer                                              NOT NULL,
    speed             integer                                              NOT NULL,
    max_hitpoints     integer                                              NOT NULL,
    deleted_at        timestamp with time zone,
    updated_at        timestamp with time zone DEFAULT now()               NOT NULL,
    created_at        timestamp with time zone DEFAULT now()               NOT NULL,
    model_id          uuid                                                 NOT NULL,
    collection        COLLECTION               DEFAULT 'supremacy-general' NOT NULL,
    power_core_size   text                     DEFAULT 'SMALL'             NOT NULL,
    tier              text                     DEFAULT 'MEGA'              NOT NULL,
    CONSTRAINT blueprint_chassis_power_core_size_check CHECK ((power_core_size = ANY
                                                               (ARRAY ['SMALL'::text, 'MEDIUM'::text, 'LARGE'::text])))
);

DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'damage_type') THEN
            CREATE TYPE DAMAGE_TYPE AS ENUM ('Kinetic', 'Energy', 'Explosive');
        END IF;
    END
$$;


CREATE TABLE IF NOT EXISTS blueprint_weapons
(
    id                    uuid                     PRIMARY KEY DEFAULT gen_random_uuid()   NOT NULL,
    brand_id              uuid REFERENCES brands (id),
    label                 text                                                 NOT NULL,
    slug                  text                                                 NOT NULL,
    damage                integer                                              NOT NULL,
    deleted_at            timestamp with time zone,
    updated_at            timestamp with time zone DEFAULT now()               NOT NULL,
    created_at            timestamp with time zone DEFAULT now()               NOT NULL,
    game_client_weapon_id uuid,
    weapon_type           WEAPON_TYPE                                          NOT NULL,
    collection            COLLECTION               DEFAULT 'supremacy-general' NOT NULL,
    default_damage_type   DAMAGE_TYPE              DEFAULT 'Kinetic'           NOT NULL,
    damage_falloff        integer                  DEFAULT 0,
    damage_falloff_rate   integer                  DEFAULT 0,
    radius                integer                  DEFAULT 0,
    radius_damage_falloff integer                  DEFAULT 0,
    spread                numeric                  DEFAULT 0,
    rate_of_fire          numeric                  DEFAULT 0,
    projectile_speed      numeric                  DEFAULT 0,
    max_ammo              integer                  DEFAULT 0,
    is_melee              boolean                  DEFAULT false               NOT NULL,
    tier                  text                     DEFAULT 'MEGA'              NOT NULL,
    energy_cost           numeric                  DEFAULT 0,
    weapon_model_id       uuid                                                 NOT NULL
);
