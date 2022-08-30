ALTER TABLE blueprint_weapon_skin
    ADD COLUMN IF NOT EXISTS image_url          text,
    ADD COLUMN IF NOT EXISTS animation_url      text,
    ADD COLUMN IF NOT EXISTS card_animation_url text,
    ADD COLUMN IF NOT EXISTS large_image_url    text,
    ADD COLUMN IF NOT EXISTS avatar_url         text,
    ADD COLUMN IF NOT EXISTS background_color   text,
    ADD COLUMN IF NOT EXISTS youtube_url        text;

ALTER TABLE blueprint_mech_skin
    ADD COLUMN image_url          text,
    ADD COLUMN animation_url      text,
    ADD COLUMN card_animation_url text,
    ADD COLUMN large_image_url    text,
    ADD COLUMN avatar_url         text,
    ADD COLUMN background_color   text,
    ADD COLUMN youtube_url        text;