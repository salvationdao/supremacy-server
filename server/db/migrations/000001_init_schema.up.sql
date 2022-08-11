BEGIN;

ALTER TABLE weapon_models
    ADD FOREIGN KEY (default_skin_id) REFERENCES blueprint_weapon_skin(id);

ALTER TABLE blueprint_mechs
    ADD FOREIGN KEY (model_id) REFERENCES mech_models(id);

ALTER TABLE blueprint_weapons
    ADD FOREIGN KEY (weapon_model_id) REFERENCES weapon_models(id);

-- game_map
CREATE TABLE game_maps
(
    id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    name           TEXT             NOT NULL UNIQUE,
    max_spawns     INT              NOT NULL DEFAULT 0,
    image_url      TEXT             NOT NULL,
    width          INT              NOT NULL,
    height         INT              NOT NULL,
    cells_x        INT              NOT NULL,
    cells_y        INT              NOT NULL,
    top_pixels     INT              NOT NULL,
    left_pixels    INT              NOT NULL,
    scale          FLOAT            NOT NULL,
    disabled_cells INT[]            NOT NULL
);


-- battles
CREATE TABLE battles
(
    id                UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    identifier        SERIAL,
    game_map_id       UUID             NOT NULL REFERENCES game_maps (id),
    winning_condition TEXT,
    started_at        TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    ended_at          TIMESTAMPTZ
);

-- users
CREATE TABLE users
(
    id                UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    view_battle_count INT              NOT NULL DEFAULT 0
);

-- record the total vote of a user per battle
CREATE TABLE battles_user_votes
(
    battle_id  UUID NOT NULL REFERENCES battles (id),
    user_id    UUID NOT NULL REFERENCES users (id),
    vote_count INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (battle_id, user_id)
);

-- battles_war_machines store the war machines attend in the battle
CREATE TABLE battles_war_machines
(
    battle_id        UUID  NOT NULL REFERENCES battles (id),
    war_machine_stat JSONB NOT NULL,
    is_winner        BOOL  NOT NULL DEFAULT FALSE,
    PRIMARY KEY (battle_id, war_machine_stat)
);

-- battle_events log all the events that happen in the battle
CREATE TABLE battle_events
(
    id         UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    battle_id  UUID REFERENCES battles (id),
    event_type TEXT,
    created_at TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

-- battle_events_state logs a battle state change
CREATE TABLE battle_events_state
(
    id       UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    event_id UUID             NOT NULL REFERENCES battle_events (id),
    state    TEXT CHECK (state IN ('START', 'END')),
    detail   JSONB NOT NULL
);

-- battle_events_war_machine_destroyed log war machine is destroyed
CREATE TABLE battle_events_war_machine_destroyed
(
    id                       UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    event_id                 UUID             NOT NULL REFERENCES battle_events (id),
    destroyed_war_machine_id NUMERIC(78, 0)   NOT NULL,
    kill_by_war_machine_id   NUMERIC(78, 0),
    related_event_id         UUID REFERENCES battle_events (id)
);

CREATE TABLE battle_events_war_machine_destroyed_assisted_war_machines
(
    war_machine_destroyed_event_id UUID           NOT NULL REFERENCES battle_events_war_machine_destroyed (id),
    war_machine_id                 NUMERIC(78, 0) NOT NULL,
    PRIMARY KEY (war_machine_destroyed_event_id, war_machine_id)
);

-- battle_events_game_ability
CREATE TABLE battle_events_game_ability
(
    id                   UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    event_id             UUID             NOT NULL REFERENCES battle_events (id),
    game_ability_id      UUID REFERENCES game_abilities (id), -- not null if it is a faction abitliy
    ability_token_id     NUMERIC(78, 0),                      -- non-zero if it is a nft ability
    is_triggered         BOOL             NOT NULL DEFAULT FALSE,
    triggered_by_user_id UUID,
    triggered_on_cell_x  INT,
    triggered_on_cell_y  INT
);

CREATE TABLE stream_list
(
    host text PRIMARY KEY NOT NULL,
    name text NOT NULL,
    url text NOT NULL,
    stream_id text NOT NULL,
    region text NOT NULL,
    resolution text NOT NULL,
    bit_rates_k_bits int NOT NULL,
    user_max int NOT NULL,
    users_now int NOT NULL,
    active boolean NOT NULL,
    status text NOT NULL,
    latitude decimal NOT NULL,
    longitude decimal NOT NULL
);

/*****************************************************
 *          faction stats materialize view           *
 ****************************************************/

-- create faction materialize view
CREATE MATERIALIZED VIEW faction_stats AS
SELECT *
FROM (
         SELECT f.id
         FROM factions f
     ) f1
         LEFT JOIN LATERAL (
    SELECT COUNT(DISTINCT bwm.battle_id) AS win_count
    FROM battles_war_machines bwm
    WHERE bwm.is_winner = TRUE
      AND bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::TEXT
    GROUP BY bwm.war_machine_stat -> 'faction' ->> 'id'
    ) f2 ON TRUE
         LEFT JOIN LATERAL (
    SELECT ((SELECT COUNT(b.id) FROM battles b) - COUNT(DISTINCT battle_id)) AS loss_count
    FROM battles_war_machines bwm
    WHERE bwm.is_winner = TRUE
      AND bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::TEXT
    GROUP BY bwm.war_machine_stat -> 'faction' ->> 'id'
    ) f3 ON TRUE
         LEFT JOIN LATERAL (
    SELECT COUNT(bewmd.id) AS kill_count
    FROM battle_events_war_machine_destroyed bewmd
             INNER JOIN battle_events be ON be.id = bewmd.event_id
             INNER JOIN battles_war_machines bwm ON be.battle_id = bwm.battle_id AND
                                                    bewmd.kill_by_war_machine_id::TEXT =
                                                    bwm.war_machine_stat ->> 'tokenID' AND
                                                    bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::TEXT
    GROUP BY bwm.war_machine_stat -> 'faction' ->> 'id'
    ) f4 ON TRUE
         LEFT JOIN LATERAL (
    SELECT COUNT(bewmd.id) AS death_count
    FROM battle_events_war_machine_destroyed bewmd
             INNER JOIN battle_events be ON be.id = bewmd.event_id
             INNER JOIN battles_war_machines bwm ON be.battle_id = bwm.battle_id AND
                                                    bewmd.destroyed_war_machine_id::TEXT =
                                                    bwm.war_machine_stat ->> 'tokenID' AND
                                                    bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::TEXT
    GROUP BY bwm.war_machine_stat -> 'faction' ->> 'id'
    ) f5 ON TRUE;

-- create unique index
CREATE UNIQUE INDEX faction_id ON faction_stats (id);

/**************************************************
 *          user stats materialize view           *
 *************************************************/

CREATE MATERIALIZED VIEW user_stats AS
SELECT *
FROM (
         SELECT u.id, u.view_battle_count
         FROM users u
     ) u1
         LEFT JOIN LATERAL (
    SELECT SUM(buv.vote_count) AS total_vote_count
    FROM battles_user_votes buv
    WHERE buv.user_id = u1.id
    GROUP BY buv.user_id
    ) u2 ON TRUE
         LEFT JOIN LATERAL (
    SELECT COUNT(bega.id) AS total_ability_triggered
    FROM battle_events_game_ability bega
    WHERE bega.triggered_by_user_id = u1.id
    ) u3 ON TRUE
         LEFT JOIN LATERAL (
    SELECT COUNT(bewmd.id) AS kill_count
    FROM battle_events_war_machine_destroyed bewmd
             INNER JOIN battle_events be ON be.id = bewmd.event_id
    WHERE EXISTS(
                  SELECT 1
                  FROM battles_war_machines bwm
                  WHERE bwm.battle_id = be.battle_id
                    AND bwm.war_machine_stat ->> 'tokenID' = bewmd.kill_by_war_machine_id::TEXT
                    AND bwm.war_machine_stat ->> 'OwnedByID' = u1.id ::TEXT
              )
    ) u4 ON TRUE;

CREATE UNIQUE INDEX user_id ON user_stats (id);

CREATE TABLE battle_war_machine_queues(
    war_machine_metadata JSONB            NOT NULL,
    queued_at            TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    released_at           TIMESTAMPTZ
);

COMMIT;

