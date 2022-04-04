ALTER TABLE telegram_notifications ADD COLUMN shortcode TEXT NOT NULL;
ALTER TABLE telegram_notifications ADD COLUMN registered BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE telegram_notifications ADD COLUMN telegram_id INT;
