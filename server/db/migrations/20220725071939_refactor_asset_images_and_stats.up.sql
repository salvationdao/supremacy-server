ALTER TABLE mech_skin
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS animation_url,
    DROP COLUMN IF EXISTS card_animation_url,
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS large_image_url;

ALTER TABLE weapon_skin
    DROP COLUMN IF EXISTS weapon_type,
    DROP COLUMN IF EXISTS tier;


