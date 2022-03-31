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
    reported_player_id UUID NOT NULL REFERENCES players (id),
    reported_player_username TEXT NOT NULL,
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