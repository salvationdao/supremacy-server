
CREATE TABLE fingerprints (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    visitor_id TEXT UNIQUE NOT NULL,
    os_cpu TEXT,
    platform TEXT,
    timezone TEXT,
    confidence DECIMAL,
    user_agent TEXT,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE player_fingerprints (
    player_id UUID NOT NULL REFERENCES players (id),
    fingerprint_id UUID NOT NULL REFERENCES fingerprints (id),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (player_id, fingerprint_id)
);

CREATE TABLE fingerprint_ips (
    ip TEXT NOT NULL,
    fingerprint_id UUID NOT NULL REFERENCES fingerprints (id),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (ip, fingerprint_id)
);

CREATE TABLE chat_banned_fingerprints (
    fingerprint_id UUID PRIMARY KEY NOT NULL
);

