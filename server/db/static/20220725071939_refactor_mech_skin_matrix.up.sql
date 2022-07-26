ALTER TABLE blueprint_mech_skin
    DROP COLUMN IF EXISTS mech_model,
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS animation_url,
    DROP COLUMN IF EXISTS card_animation_url,
    DROP COLUMN IF EXISTS large_image_url,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS background_color,
    DROP COLUMN IF EXISTS youtube_url,
    DROP COLUMN IF EXISTS mech_type;

CREATE TABLE mech_model_skin_compatibilities
(
    blueprint_mech_skin_id UUID        NOT NULL REFERENCES blueprint_mech_skin,
    mech_model_id          UUID        NOT NULL REFERENCES mech_models,
    image_url              TEXT,
    card_animation_url     TEXT,
    avatar_url             TEXT,
    large_image_url        TEXT,
    background_color       TEXT,
    animation_url          TEXT,
    youtube_url            TEXT,
    deleted_at             TIMESTAMPTZ,
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blueprint_mech_skin_id, mech_model_id)
);
