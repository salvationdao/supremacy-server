DROP TABLE player_preferences;
ALTER TABLE battle_queue DROP CONSTRAINT IF EXISTS mech_id;
ALTER TABLE battle_queue ADD CONSTRAINT unique_mech_id UNIQUE (mech_id);

CREATE TABLE player_preferences (
    player_id UUID NOT NULL REFERENCES players(id),
    key  TEXT NOT NULL,
    value JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(player_id, key)
);

CREATE TABLE telegram_notifications (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE battle_queue_notifications (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    battle_id UUID REFERENCES battles(id),
    queue_mech_id UUID REFERENCES battle_queue(mech_id),
    mech_id UUID NOT NULL REFERENCES mechs(id),
    mobile_number TEXT,
    push_notifications BOOLEAN NOT NULL DEFAULT FALSE,
    telegram_notification_id UUID REFERENCES telegram_notifications(id),
    sent_at TIMESTAMPTZ,
    message TEXT,
    fee NUMERIC(28) NOT NULL,
    is_refunded BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);