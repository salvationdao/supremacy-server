CREATE TABLE hair
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hue         TEXT        NOT NULL,
    image_url   TEXT        NOT NULL,

    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE faces
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hue         TEXT        NOT NULL,
    image_url   TEXT        NOT NULL,

    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE profile_custom_avatars
(
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id   UUID        NOT NULL REFERENCES players(id),
    face_id     UUID        NOT NULL REFERENCES players(id),
    hair_id     UUID        NOT NULL REFERENCES players(id),

    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- temp
-- seed hair
insert into hair (hue, image_url) VALUES 
('#000', 'https://user-images.githubusercontent.com/46738862/181250312-2b5ba859-80d7-4150-b6e5-38f378c94540.png'),
('#000', 'https://user-images.githubusercontent.com/46738862/181250334-b8b1d80a-4999-41cf-b52a-b51efb97e051.png'),
('#000', 'https://user-images.githubusercontent.com/46738862/181250348-c8661dce-f003-4fd6-b6c8-5936f9f38b82.png');




