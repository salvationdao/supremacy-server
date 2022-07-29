

CREATE TABLE layers
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hue         TEXT        NOT NULL,
    type        TEXT CHECK (type IN ('HAIR', 'FACE', 'BODY', 'ACCESSORY', 'EYEWEAR', 'HELMET')),
    image_url   TEXT        NOT NULL,

    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE profile_custom_avatars
(
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id   UUID        NOT NULL REFERENCES players(id),

    -- layers
    face_id          UUID        NOT NULL REFERENCES layers(id),
    hair_id          UUID        REFERENCES layers(id),
    body_id          UUID        REFERENCES layers(id),
    accessory_id     UUID        REFERENCES layers(id),
    eye_wear_id      UUID        REFERENCES layers(id),
    helmet_id        UUID        REFERENCES layers(id),

    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- seed hair layers
INSERT INTO layers (type, hue, image_url) VALUES
('HAIR', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair2.png'),
('HAIR', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair3.png'),
('HAIR', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair_2_green.png'),
('HAIR', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair_2_red.png'),
('HAIR', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair_2_yellow.png');

-- seed faces 
INSERT INTO layers (type, hue, image_url) VALUES
('FACE', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/faces/face1.png'),
('FACE', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/faces/face2.png'),
('FACE', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/faces/face3.png');

-- seed bodies
INSERT INTO layers (type, hue, image_url) VALUES
('BODY', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/bodies/body1.png'),
('BODY', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/bodies/body2.png'),
('BODY', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/bodies/body3.png');

-- seed accessories 
INSERT INTO layers (type, hue, image_url) VALUES
('ACCESSORY', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/accesories/accessories1.png'),
('ACCESSORY', '#000', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/accesories/earrings1.png');

