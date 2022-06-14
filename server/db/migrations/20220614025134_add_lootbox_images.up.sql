UPDATE storefront_mystery_crates
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/png/nexus_lootbox_redmountain_front.png',
    card_animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/mov/nexus_lootbox_redmountain_loop.mov',
    animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/webm/nexus_lootbox_redmountain_loop.webm',
    large_image_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/png/nexus_lootbox_redmountain_front_opensea.png'
WHERE "label" = 'Red Mountain War Machine Crate' OR "label" = 'Red Mountain Weapons Crate';

UPDATE storefront_mystery_crates
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/png/nexus_lootbox_zaibatsu_front.png',
    card_animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/mov/nexus_lootbox_zaibatsu_loop.mov',
    animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/webm/nexus_lootbox_zaibatsu_loop.webm',
    large_image_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/png/nexus_lootbox_zaibatsu_front_opensea.png'
WHERE "label" = 'Zaibatsu Heavy Industries Weapons Crate' OR "label" = 'Zaibatsu Heavy Industries War Machine Crate';

UPDATE storefront_mystery_crates
SET image_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/png/nexus_lootbox_boston_front.png',
    card_animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/mov/nexus_lootbox_boston_loop.mov',
    animation_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/webm/nexus_lootbox_boston_loop.webm',
    large_image_url = 'https://afiles.ninja-cdn.com/passport/nexus/lootbox/png/nexus_lootbox_boston_front_opensea.png'
WHERE "label" = 'Boston Cybernetics Weapons Crate' OR "label" = 'Boston Cybernetics War Machine Crate';
