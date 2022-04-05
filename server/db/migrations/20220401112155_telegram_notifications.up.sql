ALTER TABLE telegram_notifications
    ADD COLUMN shortcode TEXT NOT NULL,
    ADD COLUMN registered BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN telegram_id INT;
    