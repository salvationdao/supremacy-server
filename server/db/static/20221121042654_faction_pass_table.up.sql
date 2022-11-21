CREATE TABLE faction_passes(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label TEXT NOT NULL,
    last_for_days INT NOT NULL,

    sups_cost NUMERIC(28) NOT NULL DEFAULT 0,
    sups_discount_percentage DECIMAL NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
