BEGIN;

CREATE TABLE item_sales (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	faction_id UUID NOT NULL REFERENCES factions (id),
	item_id UUID NOT NULL REFERENCES collection_items (item_id),
	listing_fee_tx_id TEXT NOT NULL,
	owner_id UUID NOT NULL REFERENCES players(id),

	auction BOOL NOT NULL DEFAULT FALSE,
	auction_current_price TEXT,
	auction_reverse_price TEXT,

	buyout BOOL NOT NULL DEFAULT FALSE,
	buyout_price TEXT, -- also is used for dutch auction

	dutch_auction BOOL NOT NULL DEFAULT FALSE,
	dutch_action_drop_rate TEXT,

    end_at TIMESTAMPTZ NOT NULL,

    sold_at TIMESTAMPTZ,
	sold_for TEXT,
	sold_tx_id TEXT,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- For Dutch Auctions
CREATE TABLE item_sales_buyout_price_history (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
	item_sale_id UUID NOT NULL REFERENCES item_sales (id),
	buyout_price TEXT NOT NULL,
	created_by UUID REFERENCES players (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY (id, item_sale_id)
);

-- For Auctions
CREATE TABLE item_sales_bid_history (
	item_sale_id UUID NOT NULL REFERENCES item_sales (id),
	bidder_id UUID NOT NULL REFERENCES players (id),
    bid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	bid_price TEXT NOT NULL,
    cancelled_at TIMESTAMPTZ,
	PRIMARY KEY (item_sale_id, bidder_id, bid_at)
);

COMMIT;
