with su as
         (select ci."item_id", w."label", ci.avatar_url
          from collection_items ci
                   inner join weapons w on w.id = ci.item_id
          where ci."item_type" = 'weapon'
            and w."label" = 'Sniper Rifle')
update collection_items ci
set avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/zai/sniper/genesis_zai_weapon_snp_neon_icon.png'
from su
where su.item_id = ci.item_id;

with lsu as
         (select ci."item_id", w."label", ci.avatar_url
          from collection_items ci
                   inner join weapons w on w.id = ci.item_id
          where ci."item_type" = 'weapon'
            and w."label" = 'Laser Sword')
update collection_items ci
set avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/zai/sword/genesis_zai_weapon_swd_neon_icon.png'
from lsu
where lsu.item_id = ci.item_id;

with swu as
         (select ci."item_id", w."label", ci.avatar_url
          from collection_items ci
                   inner join weapons w on w.id = ci.item_id
          where ci."item_type" = 'weapon'
            and w."label" = 'Sword')
update collection_items ci
set avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/bc/sword/genesis_bc_weapon_swd_blue-white_icon.png'
from swu
where swu.item_id = ci.item_id;

with pr as
         (select ci."item_id", w."label", ci.avatar_url
          from collection_items ci
                   inner join weapons w on w.id = ci.item_id
          where ci."item_type" = 'weapon'
            and w."label" = 'Plasma Rifle')
update collection_items ci
set avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/bc/plasma-rifle/genesis_bc_weapon_plas_blue-white_icon.png'
from pr
where pr.item_id = ci.item_id;

with ac as
         (select ci."item_id", w."label", ci.avatar_url
          from collection_items ci
                   inner join weapons w on w.id = ci.item_id
          where ci."item_type" = 'weapon'
            and w."label" = 'Auto Cannon')
update collection_items ci
set avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/rm/cannon/genesis_rm_weapon_cnn_vintage_icon.png'
from ac
where ac.item_id = ci.item_id;

with ac as
         (select ci."item_id", pc."label", ci.avatar_url
          from collection_items ci
                   inner join power_cores pc on pc.id = ci.item_id
          where ci."item_type" = 'power_core'
            and pc."label" = 'Standard Energy Core')
update collection_items ci
set avatar_url = 'https://afiles.ninja-cdn.com/passport/nexus/utility/utility_power-core.png'
from ac
where ac.item_id = ci.item_id;

with rkt as
         (select ci."item_id", w."label", ci.avatar_url
          from collection_items ci
                   inner join weapons w on w.id = ci.item_id
          where ci."item_type" = 'weapon'
            and w."label" = 'Rocket Pod')
update collection_items ci
set avatar_url = 'https://afiles.ninja-cdn.com/passport/genesis/weapons/png/zai/rocket-pod/genesis_zai_weapon_rktpod_icon.png'
from rkt
where rkt.item_id = ci.item_id;

update collection_items
set avatar_url = 'https://afiles.ninja-cdn.com/passport/nexus/utility/genesis_zai_utility_orb-shield.png'
where "item_type" = 'utility';

update blueprint_utility
set avatar_url = 'https://afiles.ninja-cdn.com/passport/nexus/utility/genesis_zai_utility_orb-shield.png'
where "label"='Orb Shield';

update blueprint_power_cores
set avatar_url = 'https://afiles.ninja-cdn.com/passport/nexus/utility/png/utility_power-core.png';

