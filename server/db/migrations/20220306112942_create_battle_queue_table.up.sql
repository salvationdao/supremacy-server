CREATE TABLE battle_queue (
    mech_id UUID NOT NULL references mechs (id) PRIMARY KEY,
    queued_at TIMESTAMPTZ NOT NULL default NOW(),
    faction_id UUID NOT NULL references factions (id),
    owner_id UUID NOT NULL references players (id),
    battle_id UUID NULL references battles (id)
);

DROP TYPE IF EXISTS BATTLE_EVENT;
CREATE TYPE BATTLE_EVENT AS ENUM ('killed', 'spawned_ai', 'kill','ability_triggered');

ALTER TABLE battles
    DROP COLUMN identifier,
    DROP COLUMN winning_condition,
    ADD COLUMN battle_number SERIAL UNIQUE;

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