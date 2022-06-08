ALTER TABLE game_maps
    ADD COLUMN "type"      TEXT NOT NULL DEFAULT 'STORY' CHECK ("type" IN ('STORY', 'EXHIBITION')),
    ADD COLUMN disabled_at TIMESTAMPTZ,
    DROP COLUMN image_url,
    DROP COLUMN width,
    DROP COLUMN height,
    DROP COLUMN cells_x,
    DROP COLUMN cells_y,
    DROP COLUMN top_pixels,
    DROP COLUMN left_pixels,
    DROP COLUMN scale,
    DROP COLUMN disabled_cells;


INSERT INTO game_maps (name, max_spawns, type, disabled_at)
VALUES ('AokigaharaForest', 9, 'STORY', NOW()),
       ('CloudKu', 9, 'STORY', NOW()),
       ('TheHive', 9, 'STORY', NOW()),
       ('Colosseum', 9, 'STORY', NOW());
