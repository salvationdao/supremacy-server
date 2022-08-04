ALTER TABLE
    battle_queue
ADD
    COLUMN system_message_notified boolean NOT NULL DEFAULT false;