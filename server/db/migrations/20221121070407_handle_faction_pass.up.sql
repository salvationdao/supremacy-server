ALTER TYPE payment_methods ADD VALUE 'eth';
ALTER TYPE payment_methods ADD VALUE 'usd';

CREATE TABLE faction_pass_purchase_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    faction_pass_id UUID NOT NULL REFERENCES faction_passes (id),
    purchased_by_id UUID NOT NULL REFERENCES players (id),

    purchase_method payment_methods NOT NULL,

    sups_paid DECIMAL NOT NULL DEFAULT 0,
    sups_purchase_tx_id TEXT,

    eth_paid DECIMAL NOT NULL DEFAULT 0,

    usd_paid DECIMAL NOT NULL DEFAULT 0,
    stripe_payment_intent_id TEXT,

    expend_faction_pass_days int NOT NULL,

    payment_status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE players
    ADD COLUMN IF NOT EXISTS xsyn_account_id uuid,
    ADD COLUMN IF NOT EXISTS faction_pass_expires_at TIMESTAMPTZ;
