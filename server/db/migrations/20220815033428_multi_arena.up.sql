DROP TYPE IF EXISTS ARENA_TYPE_ENUM;
CREATE TYPE ARENA_TYPE_ENUM AS ENUM ('STORY','EXPEDITION');

CREATE SEQUENCE IF NOT EXISTS battle_arena_gid_seq
    INCREMENT 1
    MINVALUE 0
    MAXVALUE 9223372036854775807
    START 0
    CACHE 1;

CREATE TABLE battle_arena
(
    ID         uuid PRIMARY KEY         DEFAULT gen_random_uuid(),
    type       ARENA_TYPE_ENUM NOT NULL,
    gid        integer                  DEFAULT nextval('battle_arena_gid_seq'),
    created_at timestamptz     NOT NULL DEFAULT NOW(),
    updated_at timestamptz     NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);

-- insert default story and expedition battle arena
INSERT INTO battle_arena (id, type)
VALUES ('95774a8a-6b9c-411c-a298-20824d0f00ba', 'STORY'), ('60008739-348e-4cf0-8cca-663685e30142', 'EXPEDITION');

-- add arena id to battle table
ALTER TABLE battles
    ADD COLUMN IF NOT EXISTS arena_id uuid NOT NULL REFERENCES battle_arena (id) DEFAULT '95774a8a-6b9c-411c-a298-20824d0f00ba';

ALTER TABLE battles
    ALTER COLUMN arena_id DROP DEFAULT;

-- add arena id to mech move command table
ALTER TABLE mech_move_command_logs
    ADD COLUMN IF NOT EXISTS arena_id uuid NOT NULL REFERENCES battle_arena (id) DEFAULT '95774a8a-6b9c-411c-a298-20824d0f00ba';

ALTER TABLE mech_move_command_logs
    ALTER COLUMN arena_id DROP DEFAULT;

DROP INDEX IF EXISTS
    mech_move_command_logs_mech_id,
    mech_move_command_logs_available,
    mech_move_command_logs_record_search;

CREATE INDEX mech_move_command_logs_mech_id ON mech_move_command_logs (arena_id, mech_id);
CREATE INDEX mech_move_command_logs_available ON mech_move_command_logs (arena_id, cancelled_at, reached_at, deleted_at);
CREATE INDEX mech_move_command_logs_record_search ON mech_move_command_logs (arena_id, mech_id, battle_id, cancelled_at, reached_at, deleted_at);

ALTER TABLE chat_history
    ADD COLUMN IF NOT EXISTS arena_id uuid REFERENCES battle_arena (id);

CREATE INDEX chat_history_search ON chat_history (arena_id, faction_id);
CREATE INDEX chat_history_created_at_descending ON chat_history (created_at DESC);