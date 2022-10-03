CREATE TABLE battle_lobby_supporters
(
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    supporter_id    UUID             NOT NULL REFERENCES players (id),
    faction_id      UUID             NOT NULL REFERENCES factions (id),
    battle_lobby_id UUID             NOT NULL REFERENCES battle_lobbies (id),
    deleted_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    UNIQUE (supporter_id, battle_lobby_id)
);

CREATE TABLE battle_lobby_supporter_opt_ins
(
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    supporter_id    UUID             NOT NULL REFERENCES players (id),
    faction_id      UUID             NOT NULL REFERENCES factions (id),
    battle_lobby_id UUID             NOT NULL REFERENCES battle_lobbies (id),
    deleted_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    UNIQUE (supporter_id, battle_lobby_id)
);
