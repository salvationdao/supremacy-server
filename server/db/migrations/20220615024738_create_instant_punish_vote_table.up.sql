CREATE TABLE punish_vote_instant_pass_records(
    id uuid primary key default gen_random_uuid(),
    punish_vote_id uuid not null references punish_votes(id),
    vote_by_player_id uuid not null references players (id),
    tx_id text,

    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz
);