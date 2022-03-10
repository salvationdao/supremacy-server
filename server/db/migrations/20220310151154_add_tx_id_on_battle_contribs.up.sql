ALTER TABLE battle_contributions
    ADD COLUMN IF NOT EXISTS transaction_id UUID UNIQUE;