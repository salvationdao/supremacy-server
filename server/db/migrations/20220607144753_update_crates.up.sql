-- ALTER TABLE storefront_mystery_crates
--     ADD COLUMN label              TEXT NOT NULL DEFAULT '',
--     ADD COLUMN description        TEXT NOT NULL DEFAULT '',
--     ADD COLUMN image_url          TEXT,
--     ADD COLUMN card_animation_url TEXT,
--     ADD COLUMN avatar_url         TEXT,
--     ADD COLUMN large_image_url    TEXT,
--     ADD COLUMN background_color   TEXT,
--     ADD COLUMN animation_url      TEXT,
--     ADD COLUMN youtube_url        TEXT,
--     ADD CONSTRAINT storefront_mystery_crates_faction_id_fk FOREIGN KEY (faction_id) REFERENCES factions (id);

ALTER TABLE mystery_crate
    ADD COLUMN description TEXT NOT NULL DEFAULT '',
    DROP COLUMN image_url,
    DROP COLUMN card_animation_url,
    DROP COLUMN avatar_url,
    DROP COLUMN large_image_url,
    DROP COLUMN background_color,
    DROP COLUMN animation_url,
    DROP COLUMN youtube_url;


-- UPDATE storefront_mystery_crates
-- SET label       = 'Red Mountain War Machine Crate',
--     description = 'Contains a battle ready war machine with two weapons and matching sub models.'
-- WHERE faction_id = '98bf7bb3-1a7c-4f21-8843-458d62884060'
--   AND mystery_crate_type = 'MECH';
-- UPDATE storefront_mystery_crates
-- SET label       = 'Red Mountain Weapons Crate',
--     description = 'Contains a random weapon and weapon sub model attachment. '
-- WHERE faction_id = '98bf7bb3-1a7c-4f21-8843-458d62884060'
--   AND mystery_crate_type = 'WEAPON';
--
-- UPDATE storefront_mystery_crates
-- SET label       = 'Boston Cybernetics War Machine Crate',
--     description = 'Contains a battle ready war machine with two weapons and matching sub models.'
-- WHERE faction_id = '7c6dde21-b067-46cf-9e56-155c88a520e2'
--   AND mystery_crate_type = 'MECH';
-- UPDATE storefront_mystery_crates
-- SET label       = 'Boston Cybernetics Weapons Crate',
--     description = 'Contains a random weapon and weapon sub model attachment. '
-- WHERE faction_id = '7c6dde21-b067-46cf-9e56-155c88a520e2'
--   AND mystery_crate_type = 'WEAPON';
--
-- UPDATE storefront_mystery_crates
-- SET label       = 'Zaibatsu Heavy Industries War Machine Crate',
--     description = 'Contains a battle ready war machine with two weapons and matching sub models.'
-- WHERE faction_id = '880db344-e405-428d-84e5-6ebebab1fe6d'
--   AND mystery_crate_type = 'MECH';
-- UPDATE storefront_mystery_crates
-- SET label       = 'Zaibatsu Heavy Industries Weapons Crate',
--     description = 'Contains a random weapon and weapon sub model attachment. '
-- WHERE faction_id = '880db344-e405-428d-84e5-6ebebab1fe6d'
--   AND mystery_crate_type = 'WEAPON';
