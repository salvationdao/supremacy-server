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