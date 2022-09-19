CREATE TABLE  voice_streams
(
    id UUID PRIMARY KEY default gen_random_uuid(),
    owner_id UUID references players(id) NOT NULL,
    faction_id UUID references factions(id) NOT NULL,
    listen_stream_url TEXT NOT NULL,
    send_stream_url TEXT NOT NULL,
    is_active BOOL NOT NULL default false,
    is_faction_commander BOOL NOT NULL,
    session_expire_at TIMESTAMPTZ NOT NULL,
    created_at timestamptz NOT NULL default now()
)