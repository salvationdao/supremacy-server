DROP TYPE IF EXISTS QUEST_KEY;
CREATE TYPE QUEST_KEY AS ENUM (
    'ability_kill',
    'mech_kill',
    'mech_commander_used_in_battle',
    'repair_for_other',
    'chat_sent',
    'mech_join_battle'
);

CREATE TABLE quests (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    key QUEST_KEY not null,
    description text not null,
    -- requirement
    request_amount int not null,
    ended_at timestamptz,
    next_quest_id uuid references quests (id),

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

INSERT INTO quests (name, key, description, request_amount)
VALUES
    ('3 Ability Kills', 'ability_kill', 'Kill three opponent mechs by abilities.', 3),
    ('3 Mech Kills', 'mech_kill', 'Kill three opponent mechs by your mech.', 3),
    ('3 battles using mech commander', 'mech_commander_used_in_battle', 'Use mech commander in three different battle.', 3),
    ('3 blocks repaired for other players', 'repair_for_other', 'Repair three blocks for other players', 3),
    ('20 chat messages', 'chat_sent', 'Send 20 chat messages', 20),
    ('30 mechs join battle', 'mech_join_battle', '30 mechs engaged in battle.', 30);

CREATE TABLE players_quests (
    player_id uuid not null references players(id),
    quest_id uuid not null references quests(id),
    primary key (player_id, quest_id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);