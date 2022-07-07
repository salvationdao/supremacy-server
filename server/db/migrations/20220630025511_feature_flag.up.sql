DROP TYPE IF EXISTS FEATURE_TYPE;
CREATE TYPE FEATURE_TYPE AS ENUM ('MECH_MOVE', 'PLAYER_ABILITY');
CREATE TABLE features
(
    id         UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    type       FEATURE_TYPE PRIMARY KEY,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE players_features
(
    id           UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    player_id    UUID         NOT NULL REFERENCES players (id),
    feature_type FEATURE_TYPE NOT NULL REFERENCES features (type),
    deleted_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO features (type)
VALUES ('MECH_MOVE');
INSERT INTO features (type)
VALUES ('PLAYER_ABILITY');
