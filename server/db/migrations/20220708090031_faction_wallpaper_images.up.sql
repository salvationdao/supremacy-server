-- add wallpaper collumn
ALTER TABLE factions add wallpaper_url TEXT;

-- update factions with new image urls
update factions 
set wallpaper_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/rm-wall.png'
where label = 'Red Mountain Offworld Mining Corporation';

update factions 
set wallpaper_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/bc-wall.png'
where label = 'Boston Cybernetics';

update factions 
set wallpaper_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/zai-wall.png'
where label = 'Zaibatsu Heavy Industries';

-- add about me collumn
ALTER TABLE factions add about_me TEXT;

-- create indexes for active log
create index idx_player_active_log_active_descding on player_active_logs (player_id, active_at desc);
create index idx_player_active_log_inactive_descding on player_active_logs (player_id, inactive_at desc);