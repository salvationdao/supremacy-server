CREATE TABLE layers
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
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
INSERT INTO layers (type, image_url) VALUES
('HAIR', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair_2.png'),
('HAIR', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair_2_green.png'),
('HAIR', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair_2_red.png'),
('HAIR', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair_2_yellow.png'),
('HAIR', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair_3.png'),
('HAIR', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/hair/hair_4.png');

-- seed faces 
INSERT INTO layers (type, image_url) VALUES
('FACE', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/faces/face_1.png'),
('FACE', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/faces/face_2.png'),
('FACE', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/faces/face_3.png'),
('FACE', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/faces/face_4.png'),
('FACE', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/faces/face_5.png');


-- seed bodies
INSERT INTO layers (type, image_url) VALUES
('BODY', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/bodies/body_1.png'),
('BODY', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/bodies/body_2.png'),
('BODY', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/bodies/body_3.png'),
('BODY', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/bodies/body_4.png'),
('BODY', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/bodies/body_5.png');

-- seed accessories 
INSERT INTO layers (type, image_url) VALUES
('ACCESSORY', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/accesories/accessories_1.png'),
('ACCESSORY', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/accesories/accessories_2.png'),
('ACCESSORY', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/accesories/accessories_3.png'),
('ACCESSORY', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/accesories/accessories_4.png'),
('ACCESSORY', 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/custom_avatars/accesories/earrings_1.png');


-- add custom avatar id to players
ALTER TABLE players 
    ADD column custom_avatar_id UUID REFERENCES profile_custom_avatars(id);
