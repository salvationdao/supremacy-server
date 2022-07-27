DROP TYPE IF EXISTS system_message_type CASCADE;

DELETE FROM
    system_messages;

ALTER TABLE
    system_messages DROP COLUMN IF EXISTS "type",
ADD
    COLUMN IF NOT EXISTS sender_id uuid NOT NULL REFERENCES players,
ADD
    COLUMN IF NOT EXISTS title text NOT NULL,
ADD
    COLUMN IF NOT EXISTS data_type text,
ADD
    COLUMN IF NOT EXISTS read_at timestamptz,
    DROP COLUMN IF EXISTS is_dismissed;

BEGIN;

ALTER TYPE FEATURE_NAME
ADD
    VALUE 'SYSTEM_MESSAGES';

COMMIT;

INSERT INTO
    features (name)
VALUES
    ('SYSTEM_MESSAGES');

UPDATE
    players
SET
    username = 'Overseer'
WHERE
    username = 'System Admin';