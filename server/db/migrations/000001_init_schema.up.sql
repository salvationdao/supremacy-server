BEGIN;

-- game_map
CREATE TABLE game_maps (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    name text NOT NULL UNIQUE,
    image_url text NOT NULL,
    width int NOT NULL,
    height int NOT NULL,
    cells_x int NOT NULL,
    cells_y int NOT NULL,
    top_pixels int NOT NULL,
    left_pixels int NOT NULL,
    scale float NOT NULL,
    disabled_cells int[] NOT NULL
);

-- battles
CREATE TABLE battles (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    game_map_id uuid NOT NULL REFERENCES game_maps (id),
    winning_condition text,
    started_at timestamptz NOT NULL DEFAULT NOW(),
    ended_at timestamptz
);

-- battle events log all the events that happen in the battle
CREATE TABLE battle_events (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    battle_id uuid REFERENCES battles (id),
    created_at timestamptz NOT NULL DEFAULT NOW()
);

-- battles_war_machines store the war machines attend in the battle
CREATE TABLE battles_war_machines (
    battle_id uuid NOT NULL REFERENCES battles (id),
    war_machine_id numeric(78, 0) NOT NULL,
    war_machine_stat jsonb NOT NULL,
    join_as_faction_id uuid NOT NULL,
    is_winner bool NOT NULL DEFAULT FALSE,
    PRIMARY KEY (battle_id, war_machine_id)
);

-- faction_abilities
CREATE TABLE faction_abilities (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    faction_id uuid NOT NULL,
    label text NOT NULL,
    type text NOT NULL,
    colour text NOT NULL,
    sups_cost int NOT NULL,
    image_url text NOT NULL,
    cooldown_duration_second int NOT NULL
);

-- war_machine_destroyed_events log war machine is destroyed
CREATE TABLE war_machine_destroyed_events (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    event_id uuid NOT NULL REFERENCES battle_events (id),
    destroyed_war_machine_id numeric(78, 0) NOT NULL,
    kill_by_war_machine_id numeric(78, 0),
    kill_by_faction_ability_id uuid REFERENCES faction_abilities (id)
);

CREATE TABLE war_machine_destroyed_events_assisted_war_machines (
    war_machine_destroyed_event_id uuid NOT NULL REFERENCES war_machine_destroyed_events (id),
    war_machine_id numeric(78, 0) NOT NULL,
    PRIMARY KEY (war_machine_destroyed_event_id, war_machine_id)
);

-- faction_ability_events
CREATE TABLE faction_ability_events (
    id uuid PRIMARY KEY NOT NULL DEFAULT gen_random_uuid (),
    event_id uuid NOT NULL REFERENCES battle_events (id),
    faction_ability_id uuid NOT NULL REFERENCES faction_abilities (id),
    is_triggered bool NOT NULL DEFAULT FALSE,
    triggered_by_user text,
    triggered_on_cell_x int,
    triggered_on_cell_y int
);

COMMIT;

