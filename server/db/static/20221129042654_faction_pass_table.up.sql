CREATE TABLE faction_passes(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label TEXT NOT NULL,
    last_for_days INT NOT NULL,

    eth_price_wei DECIMAL NOT NULL DEFAULT 0,
    discount_percentage DECIMAL NOT NULL DEFAULT 0 CHECK ( discount_percentage < 100 AND discount_percentage >=0 ),

    sups_price DECIMAL NOT NULL DEFAULT 0, -- updated in runtime
    usd_price DECIMAL NOT NULL  DEFAULT 0, -- updated in runtime

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
