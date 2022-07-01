DROP TYPE IF EXISTS FEATURE_TYPE;
CREATE TYPE FEATURE_TYPE AS ENUM ('MECH_MOVE', 'PLAYER_ABILITY');
CREATE TABLE features
(
    id         UUID PRIMARY KEY             DEFAULT gen_random_uuid(),
    type       FEATURE_TYPE UNIQUE NOT NULL,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE TABLE players_features
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    player_id  UUID        NOT NULL REFERENCES players (id),
    feature_id UUID        NOT NULL REFERENCES features (id),
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO features (type)
VALUES ('MECH_MOVE');
INSERT INTO features (type)
VALUES ('PLAYER_ABILITY');
