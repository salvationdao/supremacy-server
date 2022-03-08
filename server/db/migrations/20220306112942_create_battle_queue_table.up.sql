CREATE TABLE battle_queue (
    mech_id UUID NOT NULL references mechs (id) PRIMARY KEY,
    queued_at TIMESTAMPTZ NOT NULL default NOW(),
    faction_id UUID NOT NULL references factions (id),
    owner_id UUID NOT NULL references players (id),
    battle_id UUID NULL references battles (id)
);

ALTER TABLE factions
    ADD COLUMN primary_color TEXT NOT NULL default '#000000',
    ADD COLUMN secondary_color TEXT NOT NULL default '#ffffff',
    ADD COLUMN background_color TEXT NOT NULL default '#0D0D0D';

UPDATE factions SET primary_color = '#C24242', secondary_color = '#FFFFFF', background_color = '#120E0E' WHERE id = '98bf7bb3-1a7c-4f21-8843-458d62884060';
UPDATE factions SET primary_color = '#428EC1', secondary_color = '#FFFFFF', background_color = '#080C12' WHERE id = '7c6dde21-b067-46cf-9e56-155c88a520e2';
UPDATE factions SET primary_color = '#FFFFFF', secondary_color = '#000000', background_color = '#0D0D0D' WHERE id = '880db344-e405-428d-84e5-6ebebab1fe6d';

DROP TYPE IF EXISTS BATTLE_EVENT;
CREATE TYPE BATTLE_EVENT AS ENUM ('killed', 'spawned_ai', 'kill','ability_triggered');

ALTER TABLE battles
    DROP COLUMN identifier,
    DROP COLUMN winning_condition,
    ADD COLUMN battle_number SERIAL;

CREATE TABLE battle_mechs (
    battle_id UUID NOT NULL references battles(id),
    mech_id UUID NOT NULL references mechs(id),
    owner_id UUID NOT NULL references players(id),
    faction_id UUID NOT NULL references factions(id),
    killed TIMESTAMPTZ NULL,
    killed_by_id UUID NULL references mechs(id),
    kills int NOT NULL default 0,
    damage_taken int NOT NULL default 0,
    updated_at TIMESTAMPTZ NOT NULL default NOW(),
    created_at TIMESTAMPTZ NOT NULL default NOW(),
    PRIMARY KEY(battle_id, mech_id)
);

CREATE TABLE battle_wins (
    battle_id UUID NOT NULL references battles(id),
    mech_id UUID NOT NULL references mechs(id),
    owner_id UUID NOT NULL references players(id),
    faction_id UUID NOT NULL references factions(id),
    win_condition TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL default NOW(),
    PRIMARY KEY(battle_id, mech_id)
);

CREATE TABLE battle_kills (
    battle_id UUID NOT NULL references battles(id),
    mech_id UUID NOT NULL references mechs(id),
    killed_id UUID NOT NULL references mechs(id),
    created_at TIMESTAMPTZ NOT NULL default NOW(),
    PRIMARY KEY(battle_id, killed_id)
);

CREATE TABLE battle_history (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    battle_id UUID NOT NULL references battles(id),
    related_id UUID NULL references battle_history(id),
    war_machine_one_id UUID NOT NULL references mechs(id),
    war_machine_two_id UUID NULL references mechs(id),

    event_type BATTLE_EVENT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL default NOW()
);