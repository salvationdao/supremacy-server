DROP TYPE IF EXISTS REPAIR_GAME_BLOCK_TYPE;
CREATE TYPE REPAIR_GAME_BLOCK_TYPE AS ENUM ( 'NORMAL', 'SHRINK', 'FAST', 'BOMB', 'END' );

CREATE TABLE repair_game_blocks (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    type REPAIR_GAME_BLOCK_TYPE NOT NULL DEFAULT 'NORMAL',
    min_size_multiplier DECIMAL NOT NULL DEFAULT 1,
    max_size_multiplier DECIMAL NOT NULL DEFAULT 1,
    min_speed_multiplier DECIMAL NOT NULL DEFAULT 1,
    max_speed_multiplier DECIMAL NOT NULL DEFAULT 1,
    extra_speed_multiplier DECIMAL NOT NULL DEFAULT 1,
    probability DECIMAL NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

INSERT INTO repair_game_blocks (type, min_size_multiplier, max_size_multiplier, min_speed_multiplier, max_speed_multiplier, extra_speed_multiplier, probability)
VALUES
    ('NORMAL', 1, 1, 1, 1.5, 1, 40),
    ('SHRINK', 0.7, 1, 1, 1.5, 1, 15),
    ('FAST', 1, 1, 1, 1.5, 2.5, 25),
    ('BOMB', 1, 1, 1, 1.5, 1, 20);

DROP TYPE IF EXISTS REPAIR_GAME_BLOCK_TRIGGER_KEY;
CREATE TYPE REPAIR_GAME_BLOCK_TRIGGER_KEY AS ENUM ( 'M', 'N', 'SPACEBAR' );

CREATE TABLE repair_game_block_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    repair_agent_id uuid NOT NULL REFERENCES repair_agents,
    repair_game_block_type REPAIR_GAME_BLOCK_TYPE NOT NULL DEFAULT 'NORMAL',

    size_multiplier DECIMAL NOT NULL DEFAULT 1,
    speed_multiplier DECIMAL NOT NULL DEFAULT 1,
    trigger_key REPAIR_GAME_BLOCK_TRIGGER_KEY NOT NULL,

    width DECIMAL NOT NULL DEFAULT 10,
    depth DECIMAL NOT NULL DEFAULT 10,

    stacked_at TIMESTAMPTZ,
    stacked_width DECIMAL,
    stacked_depth DECIMAL,
    is_failed bool NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_repair_game_block_log_repair_agent_id ON repair_game_block_logs (repair_agent_id);

ALTER TABLE repair_agent_logs
    RENAME TO repair_agent_logs_old;
