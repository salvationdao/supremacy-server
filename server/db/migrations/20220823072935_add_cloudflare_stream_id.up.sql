DROP TYPE IF EXISTS RECORDING_STATUS;
CREATE TYPE RECORDING_STATUS AS ENUM ('RECORDING', 'STOPPED', 'IDLE');

CREATE TABLE battle_replays
(
    id                 UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    stream_id          TEXT,
    arena_id           UUID             NOT NULL REFERENCES battle_arena (id),
    battle_id          UUID             NOT NULL REFERENCES battles (id),
    is_complete_battle BOOL             NOT NULL DEFAULT false,
    recording_status   RECORDING_STATUS NOT NULL DEFAULT 'IDLE',
    created_at         TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);