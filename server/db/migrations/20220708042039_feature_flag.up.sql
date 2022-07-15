DROP TYPE IF EXISTS FEATURE_NAME;
CREATE TYPE FEATURE_NAME AS ENUM ('MECH_MOVE', 'PLAYER_ABILITY');
CREATE TABLE features
(
    name       FEATURE_NAME PRIMARY KEY,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE players_features
(
    id           UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    player_id    UUID         NOT NULL REFERENCES players (id),
    feature_name FEATURE_NAME NOT NULL REFERENCES features (name),
    deleted_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (player_id, feature_name)
);

INSERT INTO features (name)
VALUES ('MECH_MOVE');
INSERT INTO features (name)
VALUES ('PLAYER_ABILITY');
