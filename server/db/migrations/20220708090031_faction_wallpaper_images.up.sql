-- add wallpaper collumn
ALTER TABLE factions ADD wallpaper_url TEXT;

-- update factions with new image urls
UPDATE factions 
SET wallpaper_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/rm-wall.png'
WHERE label = 'Red Mountain Offworld Mining Corporation';

UPDATE factions 
SET wallpaper_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/bc-wall.png'
WHERE label = 'Boston Cybernetics';

UPDATE factions 
SET wallpaper_url = 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/zai-wall.png'
WHERE label = 'Zaibatsu Heavy Industries';

-- add about me collumn
ALTER TABLE players ADD about_me TEXT;

-- create indexes for active log
CREATE INDEX idx_player_active_log_active_descding ON player_active_logs (player_id, active_at DESC);
CREATE INDEX idx_player_active_log_inactive_descding ON player_active_logs (player_id, inactive_at DESC);