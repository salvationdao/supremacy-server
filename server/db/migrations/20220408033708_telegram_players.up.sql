CREATE TABLE player_profile (
    id                              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    player_id                       UUID             NOT NULL REFERENCES players (id),
    shortcode                       TEXT NOT NULL,
    enable_telegram_notifications   BOOLEAN NOT NULL DEFAULT FALSE,
    enable_sms_notifications        BOOLEAN NOT NULL DEFAULT FALSE,
    enable_push_notifications       BOOLEAN NOT NULL DEFAULT FALSE,
    telegram_id                     BIGINT,
    mobile_number                   TEXT
);


