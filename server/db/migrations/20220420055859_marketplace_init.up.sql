BEGIN;

CREATE TABLE item_sales (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	item_type TEXT NOT NULL,
	item_id UUID NOT NULL,
	listing_fee_tx_id UUID NOT NULL,
	owner_id UUID NOT NULL REFERENCES users(id),

	auction BOOL NOT NULL,
	auction_current_price TEXT,
	auction_reverse_price TEXT,

	buyout BOOL NOT NULL,
	buyout_price TEXT,

	dutch_auction BOOL NOT NULL,
	dutch_action_rate INT,
	dutch_action_next_price_drop INT,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
