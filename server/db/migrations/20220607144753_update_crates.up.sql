ALTER TABLE mystery_crate
    ADD COLUMN description TEXT NOT NULL DEFAULT '',
    DROP COLUMN image_url,
    DROP COLUMN card_animation_url,
    DROP COLUMN avatar_url,
    DROP COLUMN large_image_url,
    DROP COLUMN background_color,
    DROP COLUMN animation_url,
    DROP COLUMN youtube_url;

