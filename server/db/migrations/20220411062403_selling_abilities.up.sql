CREATE TABLE blueprint_player_abilities
(
    id                     UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    game_client_ability_id INT4             NOT NULL,
    label                  TEXT             NOT NULL,
    colour                 TEXT             NOT NULL,
    image_url              TEXT             NOT NULL,
    description            TEXT             NOT NULL,
    text_colour            TEXT             NOT NULL,
    location_select_type                 TEXT NOT NULL CHECK (location_select_type IN ('MECH_SELECT', 'LOCATION_SELECT', 'GLOBAL')),
    created_at             TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE player_abilities
( -- ephemeral, entries are removed on use
    id                     UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    owner_id               UUID             NOT NULL REFERENCES players (id),
    blueprint_id           UUID             NOT NULL REFERENCES blueprint_player_abilities (id),
    game_client_ability_id INT4             NOT NULL,
    label                  TEXT             NOT NULL,
    colour                 TEXT             NOT NULL,
    image_url              TEXT             NOT NULL,
    description            TEXT             NOT NULL,
    text_colour            TEXT             NOT NULL,
    location_select_type                 TEXT NOT NULL CHECK (location_select_type IN ('MECH_SELECT', 'LOCATION_SELECT', 'GLOBAL')),
    purchased_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE sale_player_abilities
(
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    blueprint_id    UUID             NOT NULL REFERENCES blueprint_player_abilities (id),
    current_price   NUMERIC(28)      NOT NULL,
    available_until TIMESTAMPTZ
);

CREATE TABLE consumed_abilities
(
    id                     UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    battle_id              UUID             NOT NULL REFERENCES battles (id),
    consumed_by            UUID             NOT NULL REFERENCES players (id),
    blueprint_id           UUID             NOT NULL REFERENCES blueprint_player_abilities (id),
    game_client_ability_id INT4             NOT NULL,
    label                  TEXT             NOT NULL,
    colour                 TEXT             NOT NULL,
    image_url              TEXT             NOT NULL,
    description            TEXT             NOT NULL,
    text_colour            TEXT             NOT NULL,
    location_select_type                    TEXT CHECK (location_select_type IN ('MECH_SELECT', 'LOCATION_SELECT', 'GLOBAL')),
    consumed_at            TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);
