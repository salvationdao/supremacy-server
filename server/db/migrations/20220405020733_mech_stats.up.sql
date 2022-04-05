CREATE TABLE mech_stats (
    mech_id UUID PRIMARY KEY NOT NULL REFERENCES mechs (id),
    total_wins INT4 NOT NULL DEFAULT 0,
    total_deaths INT4 NOT NULL DEFAULT 0,
    total_kills INT4 NOT NULL DEFAULT 0
);
