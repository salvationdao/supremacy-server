--copied and pasted from add_weapon_images.up.sql to re-update admin given mechs
WITH su AS
         (SELECT ci."item_id", w."label", ci.avatar_url
          FROM collection_items ci
                   INNER JOIN weapons w ON w.id = ci.item_id
          WHERE ci."item_type" = 'weapon'
            AND w."label" = 'Sniper Rifle')
UPDATE collection_items ci
SET avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/zai/sniper/genesis_zai_weapon_snp_neon_icon.png'
FROM su
WHERE su.item_id = ci.item_id;

WITH lsu AS
         (SELECT ci."item_id", w."label", ci.avatar_url
          FROM collection_items ci
                   INNER JOIN weapons w ON w.id = ci.item_id
          WHERE ci."item_type" = 'weapon'
            AND w."label" = 'Laser Sword')
UPDATE collection_items ci
SET avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/zai/sword/genesis_zai_weapon_swd_neon_icon.png'
FROM lsu
WHERE lsu.item_id = ci.item_id;

WITH swu AS
         (SELECT ci."item_id", w."label", ci.avatar_url
          FROM collection_items ci
                   INNER JOIN weapons w ON w.id = ci.item_id
          WHERE ci."item_type" = 'weapon'
            AND w."label" = 'Sword')
UPDATE collection_items ci
SET avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/bc/sword/genesis_bc_weapon_swd_blue-white_icon.png'
FROM swu
WHERE swu.item_id = ci.item_id;

WITH pr AS
         (SELECT ci."item_id", w."label", ci.avatar_url
          FROM collection_items ci
                   INNER JOIN weapons w ON w.id = ci.item_id
          WHERE ci."item_type" = 'weapon'
            AND w."label" = 'Plasma Rifle')
UPDATE collection_items ci
SET avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/bc/plasma-rifle/genesis_bc_weapon_plas_blue-white_icon.png'
FROM pr
WHERE pr.item_id = ci.item_id;

WITH ac AS
         (SELECT ci."item_id", w."label", ci.avatar_url
          FROM collection_items ci
                   INNER JOIN weapons w ON w.id = ci.item_id
          WHERE ci."item_type" = 'weapon'
            AND w."label" = 'Auto Cannon')
UPDATE collection_items ci
SET avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/rm/cannon/genesis_rm_weapon_cnn_vintage_icon.png'
FROM ac
WHERE ac.item_id = ci.item_id;

WITH ac AS
         (SELECT ci."item_id", pc."label", ci.avatar_url
          FROM collection_items ci
                   INNER JOIN power_cores pc ON pc.id = ci.item_id
          WHERE ci."item_type" = 'power_core'
            AND pc."label" = 'Standard Energy Core')
UPDATE collection_items ci
SET avatar_url = 'https://afiles.ninja-cdn.com/passport/nexus/utility/utility_power-core.png'
FROM ac
WHERE ac.item_id = ci.item_id;

WITH rkt AS
         (SELECT ci."item_id", w."label", ci.avatar_url
          FROM collection_items ci
                   INNER JOIN weapons w ON w.id = ci.item_id
          WHERE ci."item_type" = 'weapon'
            AND w."label" = 'Rocket Pod')
UPDATE collection_items ci
SET avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/zai/rocket-pod/genesis_zai_weapon_rktpod_icon.png'
FROM rkt
WHERE rkt.item_id = ci.item_id;

UPDATE collection_items
SET avatar_url = 'https://afiles.ninja-cdn.com/passport/nexus/utility/genesis_zai_utility_orb-shield.png'
WHERE "item_type" = 'utility';
