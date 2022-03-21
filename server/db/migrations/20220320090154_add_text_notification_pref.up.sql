ALTER TABLE battle_queue ADD COLUMN notified BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE players ADD COLUMN mobile_number TEXT;

CREATE TABLE player_preferences (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL UNIQUE REFERENCES players(id),
    notifications_battle_queue_sms BOOL NOT NULL DEFAULT FALSE,
    notifications_battle_queue_browser BOOL NOT NULL DEFAULT TRUE,
    notifications_battle_queue_push_notifications BOOL NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
