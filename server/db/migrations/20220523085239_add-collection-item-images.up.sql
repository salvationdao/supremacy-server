ALTER TABLE collection_items
    ADD COLUMN image_url          TEXT,
    ADD COLUMN card_animation_url TEXT,
    ADD COLUMN avatar_url         TEXT,
    ADD COLUMN large_image_url    TEXT,
    ADD COLUMN background_color   TEXT,
    ADD COLUMN animation_url      TEXT,
    ADD COLUMN youtube_url        TEXT;

ALTER TABLE blueprint_mech_animation
    ADD COLUMN image_url          TEXT,
    ADD COLUMN card_animation_url TEXT,
    ADD COLUMN avatar_url         TEXT,
    ADD COLUMN large_image_url    TEXT,
    ADD COLUMN background_color   TEXT,
    ADD COLUMN animation_url      TEXT,
    ADD COLUMN youtube_url        TEXT;

ALTER TABLE blueprint_utility
    ADD COLUMN image_url          TEXT,
    ADD COLUMN card_animation_url TEXT,
    ADD COLUMN avatar_url         TEXT,
    ADD COLUMN large_image_url    TEXT,
    ADD COLUMN background_color   TEXT,
    ADD COLUMN animation_url      TEXT,
    ADD COLUMN youtube_url        TEXT;

ALTER TABLE blueprint_power_cores
    ADD COLUMN image_url          TEXT,
    ADD COLUMN card_animation_url TEXT,
    ADD COLUMN avatar_url         TEXT,
    ADD COLUMN large_image_url    TEXT,
    ADD COLUMN background_color   TEXT,
    ADD COLUMN animation_url      TEXT,
    ADD COLUMN youtube_url        TEXT;

ALTER TABLE blueprint_mech_skin
    ADD COLUMN background_color TEXT,
    ADD COLUMN youtube_url      TEXT;

ALTER TABLE blueprint_ammo
    ADD COLUMN image_url          TEXT,
    ADD COLUMN card_animation_url TEXT,
    ADD COLUMN avatar_url         TEXT,
    ADD COLUMN large_image_url    TEXT,
    ADD COLUMN background_color   TEXT,
    ADD COLUMN animation_url      TEXT,
    ADD COLUMN youtube_url        TEXT;

ALTER TABLE blueprint_weapon_skin
    ADD COLUMN image_url          TEXT,
    ADD COLUMN card_animation_url TEXT,
    ADD COLUMN avatar_url         TEXT,
    ADD COLUMN large_image_url    TEXT,
    ADD COLUMN background_color   TEXT,
    ADD COLUMN animation_url      TEXT,
    ADD COLUMN youtube_url        TEXT;