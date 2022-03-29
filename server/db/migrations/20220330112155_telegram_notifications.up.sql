-- -- ALTER TABLE player_preferences ADD COLUMN notifications_battle_queue_telegram BOOL NOT NULL DEFAULT FALSE;

-- -- CREATE TABLE telegram_notifications
-- -- (
-- --     id                  UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
-- --     player_id           UUID             NOT NULL REFERENCES players (id),
-- --     mech_id             UUID             NOT NULL REFERENCES mechs (id),
-- --     mech_queue_position INT,
-- --     telegram_id         INT,
-- --     registered          BOOLEAN NOT NULL DEFAULT FALSE,
-- --     shortcode           TEXT UNIQUE NOT NULL,
-- --     created_at          TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
-- --     expires_at          TIMESTAMPTZ
-- -- );


-- DROP TABLE player_preferences;
-- ALTER TABLE battle_queue DROP CONSTRAINT IF EXISTS mech_id;
-- ALTER TABLE battle_queue ADD CONSTRAINT unique_mech_id UNIQUE (mech_id);

-- CREATE TABLE player_preferences (
--     player_id UUID NOT NULL REFERENCES players(id),
--     key  TEXT NOT NULL,
--     value JSONB NOT NULL,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     PRIMARY KEY(player_id, key)
-- );

-- CREATE TABLE telegram_notifications (
--     id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );

-- CREATE TABLE battle_queue_notifications (
--     id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
--     battle_id UUID REFERENCES battles(id),
--     queue_mech_id UUID REFERENCES battle_queue(mech_id),
--     mech_id UUID NOT NULL REFERENCES mechs(id),
--     mobile_number TEXT,
--     push_notifications BOOLEAN NOT NULL DEFAULT FALSE,
--     telegram_notification_id UUID REFERENCES telegram_notifications(id),
--     sent_at TIMESTAMPTZ,
--     message TEXT,
--     fee NUMERIC(28) NOT NULL,
--     is_refunded BOOLEAN NOT NULL DEFAULT FALSE,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );

ALTER TABLE telegram_notifications ADD COLUMN shortcode TEXT UNIQUE NOT NULL;
ALTER TABLE telegram_notifications ADD COLUMN registered BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE telegram_notifications ADD COLUMN telegram_id INT;


