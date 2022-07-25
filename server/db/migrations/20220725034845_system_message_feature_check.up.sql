DROP TYPE IF EXISTS system_message_type CASCADE;

DELETE FROM
    system_messages;

ALTER TABLE
    system_messages DROP COLUMN IF EXISTS "type",
ADD
    COLUMN IF NOT EXISTS data_type text,
ADD
    COLUMN IF NOT EXISTS faction_id uuid REFERENCES factions(id),
ALTER COLUMN
    player_id DROP NOT NULL;