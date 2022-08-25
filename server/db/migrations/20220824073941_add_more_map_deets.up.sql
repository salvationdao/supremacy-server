-- add logo/ background url columns
ALTER TABLE game_maps
    ADD COLUMN IF NOT EXISTS background_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS logo_url TEXT NOT NULL DEFAULT '';

-- update 
UPDATE game_maps SET logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/aokigahara.png' WHERE NAME = 'AokigaharaForest';
UPDATE game_maps SET logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/arctic_bay.png' WHERE NAME = 'ArcticBay';
UPDATE game_maps SET logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/cloudku_9.png' WHERE NAME = 'CloudKu';
UPDATE game_maps SET logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/desert_city.png' WHERE NAME = 'DesertCity';
UPDATE game_maps SET logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/hive.png' WHERE NAME = 'TheHive';
UPDATE game_maps SET logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/iron_dust_5.png' WHERE NAME = '';
UPDATE game_maps SET logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/kazuya_city.png' WHERE NAME = '';
UPDATE game_maps SET logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/mibt.png' WHERE NAME = '';

CREATE TABLE battle_map_queue
(
    id                UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    map_id            UUID NOT NULL REFERENCES game_maps(id),
    created_at        timestamp with time zone DEFAULT now() NOT NULL
);

-- seed battle_map_queue
DO 
$$
DECLARE hive_id UUID;
DECLARE desert_id UUID;
BEGIN
    hive_id := (select id from game_maps where name = 'TheHive');
    insert into battle_map_queue (map_id) VALUES (hive_id);

    desert_id := (select id from game_maps where name = 'DesertCity');
    insert into battle_map_queue (map_id) VALUES (desert_id);
END;
$$;