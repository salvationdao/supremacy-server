BEGIN;

-- game_map
CREATE TABLE game_maps
(
    id             uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    name           text             NOT NULL UNIQUE,
    image_url      text             NOT NULL,
    width          int              NOT NULL,
    height         int              NOT NULL,
    cells_x        int              NOT NULL,
    cells_y        int              NOT NULL,
    top_pixels     int              NOT NULL,
    left_pixels    int              NOT NULL,
    scale          float            NOT NULL,
    disabled_cells int[] NOT NULL
);

-- battles
CREATE TABLE battles
(
    id                uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    identifier        SERIAL,
    game_map_id       uuid             NOT NULL REFERENCES game_maps (id),
    winning_condition text,
    started_at        timestamptz      NOT NULL DEFAULT NOW(),
    ended_at          timestamptz
);

-- factions
CREATE TABLE factions
(
    id                 UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    vote_price         TEXT             NOT NULL DEFAULT '1000000000000000000'
);

-- users
CREATE TABLE users
(
    id                 UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    view_battle_count  INT NOT NULL DEFAULT 0
);

-- record the total vote of a user per battle
CREATE TABLE battles_user_votes
(
    battle_id   UUID    NOT NULL REFERENCES battles (id),
    user_id     UUID    NOT NULL REFERENCES users (id),
    vote_count  INT     NOT NULL DEFAULT 0,
    PRIMARY KEY (battle_id, user_id)
);

-- battles_war_machines store the war machines attend in the battle
CREATE TABLE battles_war_machines
(
    battle_id          uuid           NOT NULL REFERENCES battles (id),
    war_machine_stat   jsonb          NOT NULL,
    is_winner          bool           NOT NULL DEFAULT FALSE,
    PRIMARY KEY (battle_id, war_machine_stat)
);


-- battle_abilities is for voting system
CREATE TABLE battle_abilities(
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    label text NOT NULL,
    cooldown_duration_second int NOT NULL
);

-- game_abilities
CREATE TABLE game_abilities (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    game_client_ability_id int NOT NULL, -- gameclient uses byte/enum instead of uuid
    faction_id uuid NOT NULL,
    battle_ability_id uuid REFERENCES battle_abilities (id), -- not null if the ability is a battle ability
    label text NOT NULL,
    colour text NOT NULL,
    image_url text NOT NULL,
    sups_cost text NOT NULL DEFAULT '0'
);

-- battle_events log all the events that happen in the battle
CREATE TABLE battle_events
(
    id         uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    battle_id  uuid REFERENCES battles (id),
    event_type TEXT,
    created_at timestamptz      NOT NULL DEFAULT NOW()
);

-- battle_events_state logs a battle state change
CREATE TABLE battle_events_state
(
    id       uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    event_id uuid             NOT NULL REFERENCES battle_events (id),
    state    TEXT CHECK (state IN ('START', 'END'))
);

-- battle_events_war_machine_destroyed log war machine is destroyed
CREATE TABLE battle_events_war_machine_destroyed
(
    id                       uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    event_id                 uuid             NOT NULL REFERENCES battle_events (id),
    destroyed_war_machine_id NUMERIC(78, 0)   NOT NULL,
    kill_by_war_machine_id   NUMERIC(78, 0),
    related_event_id         uuid REFERENCES battle_events (id)
);

CREATE TABLE battle_events_war_machine_destroyed_assisted_war_machines
(
    war_machine_destroyed_event_id uuid           NOT NULL REFERENCES battle_events_war_machine_destroyed (id),
    war_machine_id                 NUMERIC(78, 0) NOT NULL,
    PRIMARY KEY (war_machine_destroyed_event_id, war_machine_id)
);

-- battle_events_game_ability
CREATE TABLE battle_events_game_ability
(
    id                   uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    event_id             uuid             NOT NULL REFERENCES battle_events (id),
    game_ability_id      uuid             REFERENCES game_abilities (id), -- not null if it is a faction abitliy
    ability_token_id     NUMERIC(78, 0),                                     -- non-zero if it is a nft ability
    is_triggered         bool             NOT NULL DEFAULT FALSE,
    triggered_by_user_id UUID,
    triggered_on_cell_x  int,
    triggered_on_cell_y  int
);

CREATE TABLE stream_list
(
    host text PRIMARY KEY NOT NULL,
    name text NOT NULL,
    ws_url text NOT NULL,
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
select * from (
	select f.id from factions f 
)f1 left join lateral(
	select count(distinct bwm.battle_id) as win_count from battles_war_machines bwm
	where bwm.is_winner = true and bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::text 
	group by bwm.war_machine_stat -> 'faction' ->> 'id' 
)f2 on true left join lateral (
	select ((select count(b.id) from battles b) - count(distinct battle_id)) as loss_count from battles_war_machines bwm
	where bwm.is_winner = true and bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::text
	group by bwm.war_machine_stat -> 'faction' ->> 'id' 
)f3 on true left join lateral (
	select count(bewmd.id) as kill_count from battle_events_war_machine_destroyed bewmd
	inner join battle_events be on be.id = bewmd.event_id
	inner join battles_war_machines bwm on be.battle_id = bwm.battle_id and bewmd.kill_by_war_machine_id::text = bwm.war_machine_stat->>'tokenID' and  bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::text
	group by bwm.war_machine_stat -> 'faction' ->> 'id'
)f4 on true left join lateral (
	select count(bewmd.id) as death_count from battle_events_war_machine_destroyed bewmd
	inner join battle_events be on be.id = bewmd.event_id
	inner join battles_war_machines bwm on be.battle_id = bwm.battle_id and bewmd.destroyed_war_machine_id::text = bwm.war_machine_stat->>'tokenID' and bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::text
	group by bwm.war_machine_stat -> 'faction' ->> 'id'
)f5 on true;

-- create unique index
CREATE UNIQUE INDEX faction_id ON faction_stats (id);

/**************************************************
 *          user stats materialize view           *
 *************************************************/

CREATE MATERIALIZED VIEW user_stats AS 
select * from (
	select u.id, u.view_battle_count  from users u 
)u1 left join lateral(
	select sum(buv.vote_count) as total_vote_count  from battles_user_votes buv
	where buv.user_id = u1.id 
	group by buv.user_id 
)u2 on true left join lateral(
	select count(bega.id) as total_ability_triggered from battle_events_game_ability bega 
	where bega.triggered_by_user_id = u1.id
)u3 on true left join lateral(
	select count(bewmd.id) as kill_count from battle_events_war_machine_destroyed bewmd 
	inner join battle_events be on be.id = bewmd.event_id 
	where exists(
                    select 1 from battles_war_machines bwm 
                        where   bwm.battle_id = be.battle_id and 
							    bwm.war_machine_stat ->>'tokenID' = bewmd.kill_by_war_machine_id::text and 
							    bwm.war_machine_stat ->>'OwnedByID' = u1.id ::text
                )
)u4 on true;

CREATE UNIQUE INDEX user_id ON user_stats (id);

COMMIT;

