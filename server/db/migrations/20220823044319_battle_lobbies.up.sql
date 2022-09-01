-- rename battle columns that are no longer used
ALTER TABLE battles RENAME COLUMN started_battle_seconds TO started_battle_seconds_old;
ALTER TABLE battles RENAME COLUMN ended_battle_seconds TO ended_battle_seconds_old;

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

-- lobby mech trigger
CREATE OR REPLACE FUNCTION check_lobby_mech() RETURNS TRIGGER AS
$check_lobby_mech$
DECLARE
    already_join_lobby BOOLEAN DEFAULT FALSE;
    no_vacancy      BOOLEAN DEFAULT FALSE;
BEGIN

    -- check already join lobby
    SELECT (COALESCE((SELECT TRUE
                      FROM battle_lobbies_mechs blm
                               INNER JOIN battle_lobbies bl ON bl.id = blm.battle_lobby_id AND bl.finished_at ISNULL
                      WHERE blm.mech_id = new.mech_id), FALSE))
    INTO already_join_lobby;
    IF already_join_lobby THEN RAISE EXCEPTION 'already join another lobby'; END IF;

    -- check if there is vacancy left
    SELECT (SELECT lb.each_faction_mech_amount = COALESCE((SELECT COUNT(*)
                                                           FROM battle_lobbies_mechs blm
                                                           WHERE blm.battle_lobby_id = new.battle_lobby_id
                                                             AND blm.faction_id = new.faction_id), 0)
            FROM battle_lobbies lb
            WHERE lb.id = new.battle_lobby_id)
    INTO no_vacancy;
    IF no_vacancy THEN RAISE EXCEPTION 'no vacancy'; END IF;

    RETURN new;
END;
$check_lobby_mech$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_check_lobby_mech ON battle_lobbies_mechs;

CREATE TRIGGER trigger_check_lobby_mech
    BEFORE INSERT
    ON battle_lobbies_mechs
    FOR EACH ROW
EXECUTE PROCEDURE check_lobby_mech();

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
