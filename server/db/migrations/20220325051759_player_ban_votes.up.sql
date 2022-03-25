CREATE TABLE ban_votes(
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    type TEXT NOT NULL,
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

CREATE TABLE player_votes(
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    ban_vote_id UUID NOT NULL REFERENCES ban_votes (id),
    player_id UUID NOT NULL REFERENCES players (id),
    is_agreed BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- both cost will reduce half every day 00:00 am
ALTER TABLE players
    ADD COLUMN IF NOT EXISTS issue_ban_fee DECIMAL NOT NULL DEFAULT 10, -- double up if anytime a player issue a ban vote
    ADD COLUMN IF NOT EXISTS reported_cost DECIMAL NOT NULL DEFAULT 10; -- double up if anytime a player is reported by the result failed