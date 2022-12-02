ALTER TABLE battle_lobbies
    RENAME password TO access_code;

ALTER TABLE battle_lobbies
    ADD COLUMN IF NOT EXISTS max_deploy_per_player INT NOT NULL CHECK ( max_deploy_per_player >= 1 AND max_deploy_per_player <= 3 ) DEFAULT 3;

CREATE TABLE battle_lobby_extra_sups_rewards(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    battle_lobby_id UUID NOT NULL REFERENCES battle_lobbies (id),
    offered_by_id UUID NOT NULL REFERENCES players (id),
    amount NUMERIC(28) NOT NULL DEFAULT 0,

    paid_tx_id TEXT NOT NULL,
    refunded_tx_id TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);