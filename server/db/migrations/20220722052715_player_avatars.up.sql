-- stores all avatars (faction logos, mech avatars)
CREATE TABLE profile_avatars
(
    id                          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    avatar_url                  TEXT      NOT NULL,
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


ALTER TABLE blueprint_mech_skin
    ADD COLUMN profile_avatar_id UUID REFERENCES profile_avatars (id);

ALTER TABLE players
    ADD COLUMN profile_avatar_id UUID REFERENCES profile_avatars (id);