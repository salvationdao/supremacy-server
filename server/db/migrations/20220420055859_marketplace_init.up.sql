BEGIN;

CREATE TABLE item_sales (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	item_type TEXT NOT NULL CHECK (item_type IN ('MECH')),
	item_id UUID NOT NULL,
	listing_fee_tx_id UUID NOT NULL,
	owner_id UUID NOT NULL REFERENCES players(id),

	auction BOOL NOT NULL DEFAULT FALSE,
	auction_current_price TEXT,
	auction_reverse_price TEXT,

	buyout BOOL NOT NULL DEFAULT FALSE,
	buyout_price TEXT,

	dutch_auction BOOL NOT NULL DEFAULT FALSE,
	dutch_action_rate INT,
	dutch_action_next_price_drop INT,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- for checking if item exists and belongs to the owner
CREATE OR REPLACE FUNCTION checkItemOwnerConstraint()
    RETURNS TRIGGER
AS
$checkItemOwnerConstraint$
DECLARE
    record_found BOOL;
BEGIN
	record_found := false;	

	CASE NEW.item_type 
	WHEN 'MECH' THEN 
		record_found := EXISTS (
			SELECT id
			FROM mechs
			WHERE id = NEW.item_id
				AND owner_id = NEW.owner_id
		);
	END CASE;

	IF record_found = FALSE THEN
		RAISE EXCEPTION '% not found, item_id=%, owner_id=%', NEW.item_type, NEW.item_id, NEW.owner_id;
	END IF;

    RETURN NULL;
END;
$checkItemOwnerConstraint$
    LANGUAGE plpgsql;

CREATE TRIGGER checkItemOwnerConstraint
    BEFORE INSERT OR UPDATE
    ON item_sales
EXECUTE PROCEDURE checkItemOwnerConstraint();

-- TODO: do these tables
-- CREATE TABLE item_sales_buyout_price_history (
--     id UUID NOT NULL DEFAULT gen_random_uuid(),
-- 	item_sale_id UUID NOT NULL,
-- 	buyout_price TEXT NOT NULL,
-- 	created_by UUID REFERENCES players (id),
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
-- 	PRIMARY KEY (id, item_sale_id)
-- );

-- CREATE TABLE item_sales_bid_history (
-- 	item_sale_id UUID NOT NULL REFERENCES item_sales (id),
-- 	buyout_price TEXT NOT NULL,
-- 	created_by UUID REFERENCES players (id),
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
-- 	PRIMARY KEY (item_sale_id)
-- );

-- CREATE TABLE item_sales_completed (
-- 	item_sale_id UUID PRIMARY KEY NOT NULL REFERENCES item_sales (id),
-- 	tx_id UUID NOT NULL,
--     sold_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
-- 	sold_for TEXT NOT NULL,
-- 	sold_method TEXT NOT NULL CHECK (sold_method IN ('AUCTION', 'BUY_OUT')),
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );

COMMIT;
