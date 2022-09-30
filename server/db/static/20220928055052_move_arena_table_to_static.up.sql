CREATE SEQUENCE IF NOT EXISTS battle_arena_gid_seq
    INCREMENT 1
    MINVALUE 0
    MAXVALUE 9223372036854775807
    START 0
    CACHE 1;

-- auto-generated definition
CREATE TABLE IF NOT EXISTS battle_arena
(
    ID         uuid PRIMARY KEY         DEFAULT gen_random_uuid(),
    type       TEXT NOT NULL,
    gid        integer                  DEFAULT nextval('battle_arena_gid_seq'),
    created_at timestamptz     NOT NULL DEFAULT NOW(),
    updated_at timestamptz     NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
);

ALTER TABLE battle_arena
    DROP COLUMN IF EXISTS type;

ALTER TABLE battle_arena
    ALTER COLUMN gid SET NOT NULL;

DROP TYPE IF EXISTS ARENA_TYPE_ENUM;
