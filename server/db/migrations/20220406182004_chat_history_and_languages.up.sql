CREATE TABLE languages (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

INSERT INTO languages(name) VALUES ('English');

CREATE TABLE player_languages (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players (id),
    language_id UUID NOT NULL REFERENCES languages (id),
    faction_id UUID NOT NULL REFERENCES factions (id),
    text_identified TEXT NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

DROP TYPE IF EXISTS CHAT_MSG_TYPE_ENUM;
CREATE TYPE CHAT_MSG_TYPE_ENUM AS ENUM ('TEXT','PUNISH_VOTE');

CREATE TABLE chat_history (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    faction_id UUID NOT NULL references factions (id),
    player_id UUID NOT NULL references players (id),
    message_color TEXT NOT NULL,
    text TEXT NOT NULL,
    battle_id UUID NULL references battles (id),
    msg_type CHAT_MSG_TYPE_ENUM NOT NULL DEFAULT 'TEXT',
    chat_stream TEXT NOT NULL DEFAULT 'global',
    user_rank TEXT NOT NULL,
    total_multiplier TEXT NOT NULL,
    kill_count TEXT NOT NULL,
    is_citizen BOOL NOT NULL DEFAULT false,
    lang TEXT NOT NULL DEFAULT 'English',
    created_at TIMESTAMPTZ NOT NULL default NOW()
);

CREATE INDEX idx_chat_history_time ON chat_history(faction_id, created_at DESC);