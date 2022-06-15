BEGIN;

-- General Assets eg Mechs
CREATE TABLE item_sales (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	faction_id UUID NOT NULL REFERENCES factions (id),
	collection_item_id UUID NOT NULL REFERENCES collection_items (id),
	listing_fee_tx_id TEXT NOT NULL,
	owner_id UUID NOT NULL REFERENCES players(id),

	auction BOOL NOT NULL DEFAULT FALSE,
	auction_current_price DECIMAL CHECK (auction_current_price IS NULL OR auction_current_price >= 0),
	auction_reserved_price DECIMAL CHECK (auction_reserved_price IS NULL OR auction_current_price >= 0),

	buyout BOOL NOT NULL DEFAULT FALSE,
	buyout_price DECIMAL CHECK (buyout_price IS NULL OR buyout_price > 0), -- also is used for dutch auction

	dutch_auction BOOL NOT NULL DEFAULT FALSE,
	dutch_auction_drop_rate DECIMAL CHECK (dutch_auction_drop_rate IS NULL OR dutch_auction_drop_rate >= 0),

    end_at TIMESTAMPTZ NOT NULL,

    sold_at TIMESTAMPTZ,
	sold_for DECIMAL CHECK (sold_for IS NULL OR sold_for > 0),
	sold_by UUID REFERENCES players(id),
	sold_tx_id TEXT,
	sold_fee_tx_id TEXT,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- For Auctions
CREATE TABLE item_sales_bid_history (
	item_sale_id UUID NOT NULL REFERENCES item_sales (id),
	bid_tx_id TEXT NOT NULL,
	refund_bid_tx_id TEXT, 
	bidder_id UUID NOT NULL REFERENCES players (id),
    bid_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	bid_price DECIMAL NOT NULL CHECK (bid_price > 0),
    cancelled_at TIMESTAMPTZ,
	cancelled_reason TEXT,
	PRIMARY KEY (item_sale_id, bidder_id, bid_at)
);

/****************
*  1155 Assets  *
****************/

CREATE TABLE item_keycard_sales (
	id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	faction_id UUID NOT NULL REFERENCES factions (id),
	item_id UUID NOT NULL REFERENCES player_keycards (id),
	listing_fee_tx_id TEXT NOT NULL,
	owner_id UUID NOT NULL REFERENCES players(id),

	buyout_price DECIMAL NOT NULL CHECK (buyout_price > 0),

    end_at TIMESTAMPTZ NOT NULL,

    sold_at TIMESTAMPTZ,
	sold_for DECIMAL CHECK (sold_for IS NULL OR sold_for > 0),
	sold_by UUID REFERENCES players(id),
	sold_tx_id TEXT,
	sold_fee_tx_id TEXT,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
