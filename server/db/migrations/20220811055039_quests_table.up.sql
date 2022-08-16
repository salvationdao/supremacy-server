DROP TYPE IF EXISTS QUEST_EVENT_DURATION_TYPE;
CREATE TYPE QUEST_EVENT_DURATION_TYPE AS ENUM ( 'daily', 'weekly', 'monthly', 'custom' );

-- for leaderboard and quest
CREATE TABLE quest_events
(
    id                   UUID PRIMARY KEY                   DEFAULT gen_random_uuid(),
    type                 ROUND_TYPE                NOT NULL,
    name                 TEXT                      NOT NULL,
    started_at           TIMESTAMPTZ               NOT NULL,
    end_at               TIMESTAMPTZ               NOT NULL,

    -- regen method
    duration_type        QUEST_EVENT_DURATION_TYPE NOT NULL,
    custom_duration_days INT,
    repeatable           BOOL                      NOT NULL DEFAULT FALSE,
    next_quest_event_id  UUID REFERENCES quest_events (id), -- used for recording the season which generated from the current one
    quest_event_number   INT                       NOT NULL DEFAULT 0,

    created_at           TIMESTAMPTZ               NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ               NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ
);

CREATE TABLE quests
(
    id             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    quest_event_id UUID        NOT NULL REFERENCES quest_events (id),
    blueprint_id   UUID        NOT NULL REFERENCES blueprint_quests (id),
    expired_at     TIMESTAMPTZ,

    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_quest_expired_check ON quests (expired_at);

CREATE TABLE players_obtained_quests
(
    player_id         UUID        NOT NULL REFERENCES players (id),
    obtained_quest_id UUID        NOT NULL REFERENCES quests (id),
    PRIMARY KEY (player_id, obtained_quest_id),
    obtained_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);