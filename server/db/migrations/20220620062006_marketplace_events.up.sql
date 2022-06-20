ALTER TABLE item_sales RENAME COLUMN sold_by TO sold_to;
ALTER TABLE item_keycard_sales RENAME COLUMN sold_by TO sold_to;

DROP TYPE IF EXISTS MARKETPLACE_EVENT;
CREATE TYPE MARKETPLACE_EVENT AS ENUM (
	'bid',
	'bid_refund',
	'purchase'
);

CREATE TABLE marketplace_events (
	id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	event_type MARKETPLACE_EVENT NOT NULL,
	amount DECIMAL,
	related_sale_item_id UUID REFERENCES item_sales (id),
	related_sale_item_keycard_id UUID REFERENCES item_keycard_sales (id),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
