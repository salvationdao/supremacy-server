CREATE TABLE pending_transactions (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	from_user_id UUID NOT NULL,
	to_user_id UUID NOT NULL REFERENCES players(id),

    amount NUMERIC(28) NOT NULL CHECK (amount > 0),
	transaction_reference TEXT NOT NULL,
	"group" TEXT NOT NULL, 
	subgroup TEXT NOT NULL,

    processed_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW() 
);

ALTER TABLE battle_contributions ADD COLUMN processed_at TIMESTAMPTZ;

CREATE TABLE config (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    sups_per_tick NUMERIC(28) NOT NULL DEFAULT 3000000000000000000
);

INSERT INTO config (sups_per_tick) VALUES (3000000000000000000);
