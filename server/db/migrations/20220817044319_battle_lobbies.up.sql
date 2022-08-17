-- drop battle unique key to support multi arena
-- DROP INDEX IF EXISTS battles_battle_number_key;
ALTER TABLE battles
    ALTER COLUMN battle_number DROP DEFAULT,
    ALTER COLUMN battle_number TYPE INTEGER USING battle_number::INTEGER,
    DROP CONSTRAINT IF EXISTS battles_battle_number_key CASCADE, -- drop constraint and any related foreign keys
    DROP CONSTRAINT IF EXISTS battles_ended_battle_seconds_key,
    DROP CONSTRAINT IF EXISTS battles_started_battle_seconds_key;

DROP INDEX IF EXISTS battles_ended_battle_seconds_key;
DROP INDEX IF EXISTS battles_started_battle_seconds_key;

ALTER TABLE battles
    ADD CONSTRAINT battles_arena_id_battle_number_key UNIQUE (arena_id, battle_number),
    ADD CONSTRAINT battles_arena_id_ended_battle_seconds_key UNIQUE (arena_id, ended_battle_seconds),
    ADD CONSTRAINT battles_arena_id_started_battle_seconds_key UNIQUE (arena_id, started_battle_seconds);

CREATE TABLE battle_lobbies
(
    id                       UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    host_by_id               UUID        NOT NULL REFERENCES players (id),
    entry_fee                NUMERIC(28) NOT NULL DEFAULT 0,
    first_faction_cut        DECIMAL     NOT NULL DEFAULT 0,
    second_faction_cut       DECIMAL     NOT NULL DEFAULT 0,
    third_faction_cut        DECIMAL     NOT NULL DEFAULT 0,
    each_faction_mech_amount INT         NOT NULL DEFAULT 3,

    -- battle queue
    ready_at                 TIMESTAMPTZ,                  -- order of the battle lobby get in battle arena
    joined_battle_id         UUID REFERENCES battles (id), -- set battle id, if in battle
    finished_at              TIMESTAMPTZ,                  -- set when battle is completed

    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ
);

CREATE TABLE battle_lobbies_mechs
(
    battle_lobby_id UUID        NOT NULL REFERENCES battle_lobbies (id),
    mech_id         UUID        NOT NULL REFERENCES mechs (id),
    PRIMARY KEY (battle_lobby_id, mech_id),

    owner_id        UUID        NOT NULL REFERENCES players (id),
    faction_id      UUID        NOT NULL REFERENCES factions (id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE battle_lobby_bounties
(
    battle_lobby_id UUID        NOT NULL REFERENCES battle_lobbies (id),
    offered_by_id   UUID        NOT NULL REFERENCES players (id),
    target_mech_id  UUID        NOT NULL REFERENCES mechs (id),
    PRIMARY KEY (battle_lobby_id, offered_by_id, target_mech_id),

    amount          NUMERIC(28) NOT NULL DEFAULT 0,
    payoff_tx_id    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);