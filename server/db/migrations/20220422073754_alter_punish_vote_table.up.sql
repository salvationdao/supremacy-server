ALTER TABLE punish_votes
    ADD COLUMN instant_pass_by_id UUID REFERENCES players (id),
    ADD COLUMN instant_pass_fee   DECIMAL NOT NULL DEFAULT 0,
    ADD COLUMN instant_pass_tx_id TEXT;

CREATE INDEX IF NOT EXISTS contribute_index ON battle_contributions (contributed_at);
CREATE INDEX IF NOT EXISTS battle_id_index ON battle_contributions (battle_id);