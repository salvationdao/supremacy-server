CREATE TABLE kv (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT UNIQUE NOT NULL DEFAULT '',
    value TEXT NOT NULL DEFAULT '',

    deleted_at              TIMESTAMPTZ,
    updated_at              TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at              TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);
