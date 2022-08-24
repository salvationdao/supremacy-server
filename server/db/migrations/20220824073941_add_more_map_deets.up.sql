-- add logo/ background url columns
ALTER TABLE game_maps
    ADD COLUMN IF NOT EXISTS background_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS logo_url TEXT NOT NULL DEFAULT '';

-- update 
update game_maps set logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/aokigahara.png' where name = 'AokigaharaForest';
update game_maps set logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/arctic_bay.png' where name = 'ArcticBay';
update game_maps set logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/cloudku_9.png' where name = 'CloudKu';
update game_maps set logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/desert_city.png' where name = 'DesertCity';
update game_maps set logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/hive.png' where name = 'TheHive';
update game_maps set logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/iron_dust_5.png' where name = '';
update game_maps set logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/kazuya_city.png' where name = '';
update game_maps set logo_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/maps/logos/mibt.png' where name = '';
