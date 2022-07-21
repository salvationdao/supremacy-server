CREATE TYPE system_message_type AS ENUM ('MECH_QUEUE', 'MECH_BATTLE_COMPLETE');

CREATE TABLE system_messages (
    id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id uuid NOT NULL REFERENCES players(id),
    type system_message_type NOT NULL,
    message TEXT,
    data jsonb,
    is_dismissed bool NOT NULL DEFAULT false
);