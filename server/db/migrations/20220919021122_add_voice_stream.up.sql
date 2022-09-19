DROP TYPE IF EXISTS VOICE_SENDER_TYPE;
CREATE TYPE VOICE_SENDER_TYPE AS ENUM ( 'MECH_OWNER', 'FACTION_COMMANDER');


CREATE TABLE  voice_streams
(
    id UUID PRIMARY KEY default gen_random_uuid(),
    arena_id UUID REFERENCES battle_arena(id) NOT NULL,
    owner_id UUID REFERENCES players(id) NOT NULL,
    faction_id UUID REFERENCES factions(id) NOT NULL,
    listen_stream_url TEXT NOT NULL,
    send_stream_url TEXT NOT NULL,
    is_active BOOL NOT NULL default false,
    sender_type VOICE_SENDER_TYPE NOT NULL,
    session_expire_at TIMESTAMPTZ NOT NULL,
    created_at timestamptz NOT NULL default now()
)