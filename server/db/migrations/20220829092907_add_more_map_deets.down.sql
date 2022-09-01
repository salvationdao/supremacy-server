
ALTER TABLE game_maps
    DROP COLUMN IF EXISTS background_url,
    DROP COLUMN IF EXISTS logo_url;

DROP TABLE IF EXISTS battle_map_queue;
