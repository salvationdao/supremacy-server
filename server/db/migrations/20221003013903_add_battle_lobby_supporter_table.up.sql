CREATE TABLE battle_lobby_supporters
(
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    supporter_id    UUID             NOT NULL REFERENCES players (id),
    battle_lobby_id UUID             NOT NULL REFERENCES battle_lobbies (id),
    deleted_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);
