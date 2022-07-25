DROP TYPE IF EXISTS system_message_type CASCADE;

DELETE FROM
    system_messages;

ALTER TABLE
    system_messages DROP COLUMN IF EXISTS "type",
ADD
    COLUMN IF NOT EXISTS title text NOT NULL,
ADD
    COLUMN IF NOT EXISTS data_type text,
ADD
    COLUMN IF NOT EXISTS faction_id uuid REFERENCES factions(id),
ADD
    COLUMN IF NOT EXISTS read_at timestamptz,
    DROP COLUMN IF EXISTS is_dismissed,
ALTER COLUMN
    player_id DROP NOT NULL;