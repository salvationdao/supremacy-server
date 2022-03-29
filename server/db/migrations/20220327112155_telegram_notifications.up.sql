ALTER TABLE player_preferences ADD COLUMN notifications_battle_queue_telegram BOOL NOT NULL DEFAULT FALSE;

CREATE TABLE telegram_notifications
(
    id                  UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    player_id           UUID             NOT NULL REFERENCES players (id),
    mech_id             UUID             NOT NULL REFERENCES mechs (id),
    mech_queue_position INT,
    telegram_id         INT,
    registered          BOOLEAN NOT NULL DEFAULT FALSE,
    shortcode           TEXT UNIQUE NOT NULL,
    created_at          TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    expires_at          TIMESTAMPTZ
);
