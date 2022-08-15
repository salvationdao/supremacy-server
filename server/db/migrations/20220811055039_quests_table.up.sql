-- for leaderboard and quest
CREATE TABLE rounds
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    type          ROUND_TYPE  NOT NULL,
    name          TEXT        NOT NULL,
    started_at    TIMESTAMPTZ NOT NULL,
    end_at        TIMESTAMPTZ NOT NULL,

    -- regen method
    last_for_days INT         NOT NULL,
    repeatable    BOOL        NOT NULL DEFAULT FALSE,
    next_round_id UUID REFERENCES rounds (id), -- used for recording the season which generated from the current one
    is_init       BOOL        NOT NULL DEFAULT FALSE,
    round_number  INT         NOT NULL DEFAULT 0,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

-- insert an init round
INSERT INTO rounds (id, type, name, started_at, end_at, repeatable, is_init, last_for_days)
VALUES ('21e1c095-3864-499a-a38c-6f7c3e08b4ea', 'daily_quest', 'Daily Quests', NOW(), NOW(), TRUE, TRUE, 3);

CREATE TABLE IF NOT EXISTS blueprint_quests
(
    id             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    round_type     ROUND_TYPE  NOT NULL,
    key            QUEST_KEY   NOT NULL,
    name           TEXT        NOT NULL,
    description    TEXT        NOT NULL,
    request_amount INT         NOT NULL,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_blueprint_quest_round_type ON blueprint_quests (round_type);

INSERT INTO blueprint_quests (id, round_type, key, name, description, request_amount)
VALUES ('c145c789-063c-4131-9aac-5677039a2103', 'daily_quest', 'ability_kill', '3 ability kills', 'Kill three opponent mechs by triggering abilities.', 3),
       ('5f370ca0-ea08-4076-af25-9e91d1be39c6', 'daily_quest', 'mech_kill', '3 mech kills', 'Kill three opponent mechs by your mech.', 3),
       ('764575a3-a342-40b9-9dca-876d60c7288f', 'daily_quest', 'total_battle_used_mech_commander', '3 battles using mech commander', 'Use mech commander in three different battles.', 3),
       ('08bd7912-a444-4e0c-9f76-3d0cae802179', 'daily_quest', 'repair_for_other', '3 blocks repaired for other players', 'Repair three blocks for other players', 3),
       ('bab947a4-00dd-4789-876a-b50297a1fb34', 'daily_quest', 'chat_sent', '20 chat messages', 'Send 20 chat messages', 20),
       ('02f642e0-1d16-45e4-86c4-b58ee4bde9ba', 'daily_quest', 'mech_join_battle', '30 mechs join battle', '30 mechs engaged in battle.', 30);

CREATE TABLE quests
(
    id           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    round_id     UUID        NOT NULL REFERENCES rounds (id),
    blueprint_id UUID        NOT NULL REFERENCES blueprint_quests (id),
    expires_at   TIMESTAMPTZ,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_quest_expired_check ON quests (expires_at);

CREATE TABLE players_obtained_quests
(
    player_id         UUID        NOT NULL REFERENCES players (id),
    obtained_quest_id UUID        NOT NULL REFERENCES quests (id),
    PRIMARY KEY (player_id, obtained_quest_id),
    obtained_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);