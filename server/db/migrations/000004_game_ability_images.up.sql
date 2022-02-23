CREATE TABLE blobs (
    id              UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    file_name       TEXT             NOT NULL,
    mime_type       TEXT             NOT NULL,
    file_size_bytes BIGINT           NOT NULL,
    extension       TEXT             NOT NULL,
    file            BYTEA            NOT NULL,
    views           INTEGER          NOT NULL DEFAULT 0,
    hash            TEXT,
    deleted_at      TIMESTAMPTZ,
    updated_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);
