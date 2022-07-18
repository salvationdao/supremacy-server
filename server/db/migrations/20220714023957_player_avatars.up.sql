CREATE TABLE profile_avatars
(
    id                          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    avatar_url                  TEXT        NOT NULL,
    tier                        TEXT        NOT NULL,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- stores availible avatars for each player
CREATE TABLE players_profile_avatars
(
    id                          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    player_id                   UUID             NOT NULL REFERENCES players (id),
    profile_avatar_id           UUID             NOT NULL REFERENCES profile_avatars (id),
    updated_at                  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

DO
$$
    DECLARE
        zhi_logo TEXT;
        bc_logo  TEXT;
        rm_logo  TEXT;

        zhi_logo_id UUID;
        bc_logo_id UUID;
        rm_logo_id UUID;

    BEGIN
        -- faction logo urls 
        zhi_logo := 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/zai-logo.svg';
        bc_logo := 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/bc-logo.svg';
        rm_logo := 'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/factions/rm-logo.svg';

        -- seed default avatars (faction logos) 
        INSERT INTO profile_avatars 
        (avatar_url, tier)
        VALUES
        (zhi_logo, 'MEGA'),
        (bc_logo, 'MEGA'),
        (rm_logo, 'MEGA');

        -- player profile logo ids 
        zhi_logo_id := (SELECT id FROM profile_avatars WHERE avatar_url = zhi_logo);
        bc_logo_id := (SELECT id FROM profile_avatars WHERE avatar_url = bc_logo);
        rm_logo_id := (SELECT id FROM profile_avatars WHERE avatar_url = rm_logo);

        -- give ZHI default images 
        INSERT INTO players_profile_avatars 
            (player_id, profile_avatar_id)
            SELECT players.id, zhi_logo_id FROM players
            INNER JOIN factions ON players.faction_id = factions.id 
            WHERE factions.label = 'Zaibatsu Heavy Industries';


        -- give BC default images 
        INSERT INTO players_profile_avatars 
            (player_id, profile_avatar_id)
            SELECT players.id, bc_logo_id from players
            INNER join factions on players.faction_id = factions.id 
            WHERE factions.label = 'Boston Cybernetics';

         -- give RM default images 
        INSERT INTO players_profile_avatars 
            (player_id, profile_avatar_id)
            SELECT players.id, rm_logo_id from players
            INNER join factions on players.faction_id = factions.id 
            WHERE factions.label = 'Red Mountain Offworld Mining Corporation';
        

    END;
$$;

ALTER TABLE blueprint_mech_skin
    ADD COLUMN profile_avatar_id UUID REFERENCES profile_avatars (id);


with inserted_avatars as (
    with bms as (select avatar_url, tier from blueprint_mech_skin)
    insert into profile_avatars(avatar_url, tier) 
    SELECT coalesce(bms.avatar_url, ''), bms.tier from blueprint_mech_skin bms 
    RETURNING id, avatar_url)
update blueprint_mech_skin 
set profile_avatar_id = inserted_avatars.id 
from inserted_avatars
where blueprint_mech_skin.avatar_url = inserted_avatars.avatar_url;

INSERT INTO players_profile_avatars (player_id, profile_avatar_id)
SELECT DISTINCT p.id, bms.profile_avatar_id  FROM players p 
INNER JOIN collection_items ci ON ci.owner_id =  p.id 
inner join mech_skin ms ON ms.id = ci.item_id 
INNER JOIN blueprint_mech_skin bms ON bms.id = ms.blueprint_id;

ALTER TABLE players
    ADD COLUMN profile_avatar_id UUID REFERENCES profile_avatars (id);



-- add to players_profile_avatars
-- - when player buys item 
-- - enlist in a faction 


-- add to players_profile_avatars
-- - sell item 

-- shoiuld mech  skings have unique urls