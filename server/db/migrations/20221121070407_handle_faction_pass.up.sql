ALTER TYPE payment_methods ADD VALUE 'eth';
ALTER TYPE payment_methods ADD VALUE 'usd';

CREATE TABLE faction_pass_purchase_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    faction_pass_id UUID NOT NULL REFERENCES faction_passes (id),
    purchased_by_id UUID NOT NULL REFERENCES players (id),

    purchase_method payment_methods NOT NULL,
    price DECIMAL NOT NULL, -- maybe sups, eth or usd
    discount DECIMAL NOT NULL DEFAULT 0,

    purchase_tx_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE players
    ADD COLUMN IF NOT EXISTS faction_pass_expires_at TIMESTAMPTZ;