BEGIN;

-- battle
CREATE TABLE battle
(
    id                   UUID PRIMARY KEY NOT NULL,
    war_machines         JSONB,
    winning_war_machines JSONB,
    winning_condition    TEXT,
    started_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    ended_at             TIMESTAMPTZ
);


-- battle event
CREATE TABLE battle_event
(
    id         UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    battle_id  UUID REFERENCES battle (id),
    event_type TEXT,
    event      JSONB,
    created_at TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

COMMIT;
