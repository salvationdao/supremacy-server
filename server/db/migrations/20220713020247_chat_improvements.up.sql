ALTER TABLE chat_history
    ADD COLUMN battle_number INT,
    ADD COLUMN metadata      jsonb;
