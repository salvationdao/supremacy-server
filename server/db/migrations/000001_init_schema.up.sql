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


-- battles_war_machines store the war machines attend in the battle
CREATE TABLE battles_war_machines
(
    battle_id          uuid           NOT NULL REFERENCES battles (id),
    war_machine_id     numeric(78, 0) NOT NULL,
    war_machine_stat   jsonb          NOT NULL,
    join_as_faction_id uuid           NOT NULL,
    is_winner          bool           NOT NULL DEFAULT FALSE,
    PRIMARY KEY (battle_id, war_machine_id)
);


-- battle_abilities is for voting system
CREATE TABLE battle_abilities(
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    label text NOT NULL,
    cooldown_duration_second int NOT NULL
);

-- faction_abilities
CREATE TABLE faction_abilities (
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
    destroyed_war_machine_id numeric(78, 0)   NOT NULL,
    kill_by_war_machine_id   numeric(78, 0),
    related_event_id         uuid REFERENCES battle_events (id)
);

CREATE TABLE battle_events_war_machine_destroyed_assisted_war_machines
(
    war_machine_destroyed_event_id uuid           NOT NULL REFERENCES battle_events_war_machine_destroyed (id),
    war_machine_id                 numeric(78, 0) NOT NULL,
    PRIMARY KEY (war_machine_destroyed_event_id, war_machine_id)
);

-- battle_events_faction_ability
CREATE TABLE battle_events_faction_ability
(
    id                   uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    event_id             uuid             NOT NULL REFERENCES battle_events (id),
    faction_ability_id   uuid             NOT NULL REFERENCES faction_abilities (id),
    is_triggered         bool             NOT NULL DEFAULT FALSE,
    triggered_by_user_id text,
    triggered_on_cell_x  int,
    triggered_on_cell_y  int
);

COMMIT;

