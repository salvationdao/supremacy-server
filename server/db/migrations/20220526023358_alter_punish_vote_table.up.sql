ALTER TABLE punish_votes
    ADD COLUMN instant_pass_by_id UUID REFERENCES players (id),
    ADD COLUMN instant_pass_fee decimal not null default 0,
    ADD COLUMN instant_pass_tx_id TEXT;