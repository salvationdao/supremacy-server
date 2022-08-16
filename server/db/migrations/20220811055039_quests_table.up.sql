DROP TYPE IF EXISTS ROUND_DURATION_TYPE;
CREATE TYPE ROUND_DURATION_TYPE AS ENUM ( 'daily', 'weekly', 'monthly', 'custom' );

-- for leaderboard and quest
CREATE TABLE rounds
(
    id                   UUID PRIMARY KEY             DEFAULT gen_random_uuid(),
    type                 ROUND_TYPE          NOT NULL,
    name                 TEXT                NOT NULL,
    started_at           TIMESTAMPTZ         NOT NULL,
    end_at               TIMESTAMPTZ         NOT NULL,

    -- regen method
    duration_type        ROUND_DURATION_TYPE NOT NULL,
    custom_duration_days INT,
    repeatable           BOOL                NOT NULL DEFAULT FALSE,
    next_round_id        UUID REFERENCES rounds (id), -- used for recording the season which generated from the current one
    round_number         INT                 NOT NULL DEFAULT 0,

    created_at           TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ
);


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

CREATE TABLE quests
(
    id           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    round_id     UUID        NOT NULL REFERENCES rounds (id),
    blueprint_id UUID        NOT NULL REFERENCES blueprint_quests (id),
    expired_at   TIMESTAMPTZ,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_quest_expired_check ON quests (expired_at);

CREATE TABLE players_obtained_quests
(
    player_id         UUID        NOT NULL REFERENCES players (id),
    obtained_quest_id UUID        NOT NULL REFERENCES quests (id),
    PRIMARY KEY (player_id, obtained_quest_id),
    obtained_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);