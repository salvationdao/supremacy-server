CREATE TABLE blueprint_player_abilities
(
    id                     UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    game_client_ability_id INT4             NOT NULL,
    label                  TEXT             NOT NULL,
    colour                 TEXT             NOT NULL,
    image_url              TEXT             NOT NULL,
    description            TEXT             NOT NULL,
    text_colour            TEXT             NOT NULL,
    "type"                 TEXT CHECK ("type" IN ('MECH_SELECT', 'LOCATION_SELECT', 'GLOBAL'))
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
    "type"                 TEXT CHECK ("type" IN ('MECH_SELECT', 'LOCATION_SELECT', 'GLOBAL')),
    purchased_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE sale_player_abilities
(
    blueprint_id    UUID        NOT NULL PRIMARY KEY REFERENCES blueprint_player_abilities (id),
    current_price   NUMERIC(28) NOT NULL,
    available_until TIMESTAMPTZ
);

CREATE TABLE consumed_abilities
(
    battle_id         UUID        NOT NULL PRIMARY KEY REFERENCES battles (id),
    player_ability_id UUID        NOT NULL REFERENCES player_abilities (id),
    consumed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
