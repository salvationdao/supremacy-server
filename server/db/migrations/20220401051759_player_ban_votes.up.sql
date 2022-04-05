CREATE TABLE punish_options (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    description TEXT NOT NULL,
    key TEXT NOT NULL UNIQUE,
    punish_duration_hours INT NOT NULL DEFAULT 24,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

INSERT INTO punish_options (description, key, punish_duration_hours) VALUES
('Limit player to select location for 24 hours', 'limit_location_select', 24),
('Limit player to chat for 24 hours', 'limit_chat', 24),
('Limit player to contibute sups for 24 hours', 'limit_sups_contibution', 24);


CREATE TABLE punish_votes(
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    punish_option_id UUID NOT NULL REFERENCES punish_options (id),
    reason TEXT NOT NULL,
    faction_id UUID NOT NULL REFERENCES factions(id),
    issued_by_id UUID NOT NULL REFERENCES players(id),
    issued_by_username TEXT NOT NULL,
    issued_by_gid INT NOT NULL,
    reported_player_id UUID NOT NULL REFERENCES players (id),
    reported_player_username TEXT NOT NULL,
    reported_player_gid INT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('PASSED','FAILED','PENDING')),
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE players_punish_votes(
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    punish_vote_id UUID NOT NULL REFERENCES punish_votes (id),
    player_id UUID NOT NULL REFERENCES players (id),
    is_agreed BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- both cost will reduce half every day 00:00 am
ALTER TABLE players
    ADD COLUMN IF NOT EXISTS issue_punish_fee DECIMAL NOT NULL DEFAULT 10, -- double up if anytime a player issue a punish vote
    ADD COLUMN IF NOT EXISTS reported_cost DECIMAL NOT NULL DEFAULT 10; -- double up if anytime a player is reported by the result failed

CREATE TABLE punished_players(
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players (id),
    punish_option_id UUID NOT NULL REFERENCES punish_options (id),
    punish_until TIMESTAMPTZ NOT NULL,
    related_punish_vote_id UUID REFERENCES punish_votes(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- active player log
CREATE TABLE player_active_logs(
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players (id),
    faction_id UUID REFERENCES factions (id),
    active_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    inactive_at TIMESTAMPTZ
);

CREATE SEQUENCE IF NOT EXISTS players_gid_seq
    INCREMENT 1
    MINVALUE 1
    MAXVALUE 9223372036854775807
    START 1024
    CACHE 1;


DROP TYPE IF EXISTS PLAYER_RANK_ENUM;
CREATE TYPE PLAYER_RANK_ENUM AS ENUM ('GENERAL','CORPORAL','PRIVATE','NEW_RECRUIT');

ALTER TABLE players
    ADD COLUMN gid integer NOT NULL DEFAULT nextval('players_gid_seq'),
    ADD COLUMN rank PLAYER_RANK_ENUM NOT NULL DEFAULT 'NEW_RECRUIT',
    ADD COLUMN sent_message_count int NOT NULL default 0 ;

CREATE MATERIALIZED VIEW player_last_seven_day_ability_kills AS
SELECT p.id, p.faction_id ,(p1.positive_kills - p2.team_kills) as kill_count from players p
left join lateral (
    -- get positive ability kills
    SELECT count(bh.id) as positive_kills from battle_history bh
        INNER JOIN battle_ability_triggers bat on bat.ability_offering_id = bh.related_id AND bat.player_id = p.id
        INNER JOIN battle_mechs bm on bm.mech_id = bh.war_machine_one_id AND bm.battle_id = bat.battle_id
    where bat.faction_id != bm.faction_id and bh.created_at > NOW() - INTERVAL '7 DAY'
    ) p1 on true left join lateral (
    -- get team kill count
    SELECT count(bh.id) as team_kills from battle_history bh
        INNER JOIN battle_ability_triggers bat on bat.ability_offering_id = bh.related_id AND bat.player_id = p.id
        INNER JOIN battle_mechs bm on bm.mech_id = bh.war_machine_one_id AND bm.battle_id = bat.battle_id
    where bat.faction_id = bm.faction_id and bh.created_at > NOW() - INTERVAL '7 DAY'
    ) p2 on true;

CREATE UNIQUE INDEX ON player_last_seven_day_ability_kills (id);

REFRESH MATERIALIZED VIEW CONCURRENTLY player_last_seven_day_ability_kills;

-- set private rank players (accounts are created over 24 hrs)
UPDATE players SET rank = 'PRIVATE' WHERE created_at > NOW() - INTERVAL '1 DAY';

-- set corporal rank players (any private players who have positive ability kill count)
UPDATE players p SET rank = 'CORPORAL' WHERE p.rank = 'PRIVATE' AND EXISTS (SELECT 1 FROM user_stats us WHERE us.id = p.id AND us.kill_count >0);