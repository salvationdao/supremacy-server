CREATE TYPE FEATURE_TYPE AS ENUM ('foo', 'bar');
CREATE TABLE features
(
    id         UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    type       FEATURE_TYPE NOT NULL,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE players_features
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    player_id  UUID        NOT NULL REFERENCES players (id),
    feature_id UUID        NOT NULL REFERENCES features (id),
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
