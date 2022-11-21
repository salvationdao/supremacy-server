CREATE TABLE faction_passes(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label TEXT NOT NULL,
    last_for_days INT NOT NULL,
    sups_cost NUMERIC(28) NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

INSERT INTO faction_passes (id, label, last_for_days, sups_cost) VALUES
('77dd9d60-c39a-4866-8ed6-0f25cf8b7a51', '24 HOUR', 1, 500000000000000000000),
('76c7e69b-b5f3-43b2-9a46-1e0debd70b1d', 'MONTHLY', 30, 8000000000000000000000),
('09782e79-d417-43b3-a395-4da571220e2c', 'YEARLY', 365, 80000000000000000000000);
