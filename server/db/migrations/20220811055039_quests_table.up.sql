-- for leaderboard and quest
CREATE TABLE rounds
(
    id            UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
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
INSERT INTO rounds (id, name, started_at, end_at, last_for_days, repeatable, is_init)
VALUES ('21e1c095-3864-499a-a38c-6f7c3e08b4ea', 'QUEST', NOW(), NOW(), 3, TRUE, TRUE);

DROP TYPE IF EXISTS QUEST_KEY;
CREATE TYPE QUEST_KEY AS ENUM (
    'ability_kill',
    'mech_kill',
    'total_battle_used_mech_commander',
    'repair_for_other',
    'chat_sent',
    'mech_join_battle'
    );

CREATE TABLE quests
(
    id             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    round_id       UUID        NOT NULL REFERENCES rounds (id),
    name           TEXT        NOT NULL,
    key            QUEST_KEY   NOT NULL,
    description    TEXT        NOT NULL,
    -- requirement
    request_amount INT         NOT NULL,
    expires_at     TIMESTAMPTZ NOT NULL,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_quest_expired_check ON quests (expires_at);
CREATE INDEX idx_quest_available_check ON quests (key, expires_at DESC);

-- insert the very first quests
INSERT INTO quests (name, key, description, request_amount, expires_at, round_id)
VALUES
    ('3 ability kills', 'ability_kill', 'Kill three opponent mechs by triggering abilities.', 3, NOW(), '21e1c095-3864-499a-a38c-6f7c3e08b4ea'),
    ('3 mech kills', 'mech_kill', 'Kill three opponent mechs by your mech.', 3, NOW(), '21e1c095-3864-499a-a38c-6f7c3e08b4ea'),
    ('3 battles using mech commander', 'total_battle_used_mech_commander', 'Use mech commander in three different battles.', 3, NOW(), '21e1c095-3864-499a-a38c-6f7c3e08b4ea'),
    ('3 blocks repaired for other players', 'repair_for_other', 'Repair three blocks for other players', 3, NOW(), '21e1c095-3864-499a-a38c-6f7c3e08b4ea'),
    ('20 chat messages', 'chat_sent', 'Send 20 chat messages', 20, NOW(), '21e1c095-3864-499a-a38c-6f7c3e08b4ea'),
    ('30 mechs join battle', 'mech_join_battle', '30 mechs engaged in battle.', 30, NOW(), '21e1c095-3864-499a-a38c-6f7c3e08b4ea');

CREATE TABLE players_quests
(
    player_id  UUID        NOT NULL REFERENCES players (id),
    quest_id   UUID        NOT NULL REFERENCES quests (id),
    PRIMARY KEY (player_id, quest_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);