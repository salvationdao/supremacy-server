-- CREATE TABLE hair
-- (
--     id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     hue         TEXT        NOT NULL,
--     image_url   TEXT        NOT NULL,

--     updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     deleted_at  TIMESTAMPTZ,
--     created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );

-- CREATE TABLE faces
-- (
--     id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     hue         TEXT        NOT NULL,
--     image_url   TEXT        NOT NULL,

--     updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     deleted_at  TIMESTAMPTZ,
--     created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );

-- CREATE TABLE profile_custom_avatars
-- (
--     id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
--     player_id   UUID        NOT NULL REFERENCES players(id),
--     face_id     UUID        NOT NULL REFERENCES players(id),
--     hair_id     UUID        NOT NULL REFERENCES players(id),

--     updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     deleted_at  TIMESTAMPTZ,
--     created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );
