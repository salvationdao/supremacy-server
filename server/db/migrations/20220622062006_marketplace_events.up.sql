ALTER TABLE item_sales RENAME COLUMN sold_by TO sold_to;
ALTER TABLE item_keycard_sales RENAME COLUMN sold_by TO sold_to;

DROP TYPE IF EXISTS MARKETPLACE_EVENT;
CREATE TYPE MARKETPLACE_EVENT AS ENUM (
	-- Buyer
	'bid',
	'bid_refund',
	'purchase',

	-- Seller
	'created',
	'sold',

	-- Common
	'cancelled'
);

CREATE TABLE marketplace_events (
	id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	user_id uuid NOT NULL REFERENCES players (id),
	event_type MARKETPLACE_EVENT NOT NULL,
	amount DECIMAL,
	related_sale_item_id UUID REFERENCES item_sales (id),
	related_sale_item_keycard_id UUID REFERENCES item_keycard_sales (id),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
