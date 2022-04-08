CREATE TABLE telegram_players (
    id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    player_id      UUID             NOT NULL REFERENCES players (id),
    shortcode      TEXT NOT NULL,
    telegram_id    BIGINT  
);

ALTER TABLE battle_queue_notifications ADD COLUMN telegram_player_id UUID REFERENCES telegram_players (id);