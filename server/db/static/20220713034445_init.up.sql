CREATE TABLE IF NOT EXISTS battle_abilities
(
    id                       uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    label                    text                                       NOT NULL,
    cooldown_duration_second integer                                    NOT NULL,
    description              text                                       NOT NULL
);

DROP TYPE IF EXISTS MECH_TYPE;
CREATE TYPE MECH_TYPE AS ENUM ('HUMANOID', 'PLATFORM');

DROP TYPE IF EXISTS COLLECTION;
CREATE TYPE COLLECTION AS ENUM ('supremacy-ai','supremacy-genesis', 'supremacy-limited-release', 'supremacy-general', 'supremacy-consumables');

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

DROP TYPE IF EXISTS WEAPON_TYPE;
CREATE TYPE WEAPON_TYPE AS ENUM ('Grenade Launcher', 'Cannon', 'Minigun', 'Plasma Gun', 'Flak',
    'Machine Gun', 'Flamethrower', 'Missile Launcher', 'Laser Beam',
    'Lightning Gun', 'BFG', 'Rifle', 'Sniper Rifle', 'Sword');


CREATE TABLE IF NOT EXISTS weapon_models
(
    id              uuid PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    brand_id        uuid,
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
    weapon_model_id    uuid                                                 NOT NULL REFERENCES weapon_models (id),
    stat_modifier      numeric(8, 0)
);

CREATE TABLE factions
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


CREATE TABLE brands
(
    id         uuid                     PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    faction_id uuid                                               NOT NULL REFERENCES factions (id),
    label      text                                               NOT NULL,
    deleted_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now()             NOT NULL,
    created_at timestamp with time zone DEFAULT now()             NOT NULL
);


DROP TYPE IF EXISTS ABILITY_LEVEL;
CREATE TYPE ABILITY_LEVEL AS ENUM ('MECH','FACTION','PLAYER');

DROP TYPE IF EXISTS LOCATION_SELECT_TYPE_ENUM;
CREATE TYPE LOCATION_SELECT_TYPE_ENUM AS ENUM (
    'LINE_SELECT',
    'MECH_SELECT',
    'LOCATION_SELECT',
    'GLOBAL'
    );

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

CREATE TABLE IF NOT EXISTS mech_models
(
    id                      uuid PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    label                   text                                               NOT NULL,
    created_at              timestamp with time zone DEFAULT now()             NOT NULL,
    default_chassis_skin_id uuid                                               NOT NULL,
    brand_id                uuid,
    mech_type               MECH_TYPE
);

DROP TYPE IF EXISTS CRATE_TYPE;
CREATE TYPE CRATE_TYPE AS ENUM ('MECH', 'WEAPON');

CREATE TABLE IF NOT EXISTS storefront_mystery_crates
(
    id                 uuid PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    mystery_crate_type CRATE_TYPE                                         NOT NULL,
    price              numeric(28, 0)                                     NOT NULL,
    amount             integer                                            NOT NULL,
    amount_sold        integer                  DEFAULT 0                 NOT NULL,
    faction_id         uuid                                               NOT NULL,
    deleted_at         timestamp with time zone,
    updated_at         timestamp with time zone DEFAULT now()             NOT NULL,
    created_at         timestamp with time zone DEFAULT now()             NOT NULL,
    label              text                     DEFAULT ''::text          NOT NULL,
    description        text                     DEFAULT ''::text          NOT NULL,
    image_url          text,
    card_animation_url text,
    avatar_url         text,
    large_image_url    text,
    background_color   text,
    animation_url      text,
    youtube_url        text
);
