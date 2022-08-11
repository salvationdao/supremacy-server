DROP TYPE IF EXISTS QUEST_KEY;
CREATE TYPE QUEST_KEY AS ENUM (
    'ability_kill',
    'mech_kill',
    'total_battle_used_mech_commander',
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
    expires_at timestamptz not null,
    last_for_days int not null,
    repeatable bool not null default false,
    next_quest_id uuid references quests (id), -- used for recording the quest generated from the current one

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);

CREATE INDEX idx_quest_expired_check ON quests (repeatable, expires_at, next_quest_id);
CREATE INDEX idx_quest_available_check ON quests (key, expires_at DESC);

-- insert the very first quests
INSERT INTO quests (name, key, description, request_amount, expires_at, last_for_days, repeatable)
VALUES
    ('3 ability kills', 'ability_kill', 'Kill three opponent mechs by abilities.', 3, now(), 3, true),
    ('3 mech kills', 'mech_kill', 'Kill three opponent mechs by your mech.', 3, now(), 3, true),
    ('3 battles using mech commander', 'total_battle_used_mech_commander', 'Use mech commander in three different battle.', 3, now(), 3, true),
    ('3 blocks repaired for other players', 'repair_for_other', 'Repair three blocks for other players', 3, now(), 3, true),
    ('20 chat messages', 'chat_sent', 'Send 20 chat messages', 20, now(), 3, true),
    ('30 mechs join battle', 'mech_join_battle', '30 mechs engaged in battle.', 30, now(), 3, true);

CREATE TABLE players_quests (
    player_id uuid not null references players(id),
    quest_id uuid not null references quests(id),
    primary key (player_id, quest_id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);