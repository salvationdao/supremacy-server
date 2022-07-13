ALTER TABLE chat_history
    ADD COLUMN battle_number INT NOT NULL DEFAULT 0,
    ADD COLUMN metadata      jsonb;

ALTER TYPE chat_msg_type_enum ADD VALUE 'SYSTEM_BAN';
ALTER TYPE chat_msg_type_enum ADD VALUE 'NEW_BATTLE';