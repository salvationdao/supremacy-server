ALTER TABLE blueprint_weapon_skin
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS animation_url,
    DROP COLUMN IF EXISTS card_animation_url,
    DROP COLUMN IF EXISTS large_image_url,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS background_color,
    DROP COLUMN IF EXISTS youtube_url;

ALTER TABLE blueprint_mech_skin
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS animation_url,
    DROP COLUMN IF EXISTS card_animation_url,
    DROP COLUMN IF EXISTS large_image_url,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS background_color,
    DROP COLUMN IF EXISTS youtube_url;
