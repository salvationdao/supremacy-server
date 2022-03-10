ALTER TABLE pending_transactions
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

CREATE TABLE spoils_of_war (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    battle_id UUID NOT NULL REFERENCES battles (id),
    battle_number INT NOT NULL REFERENCES battles (battle_number),
    amount numeric(28) NOT NULL,
    amount_sent numeric(28) NOT NULL default 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);