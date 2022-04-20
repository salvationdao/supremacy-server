BEGIN;

CREATE TABLE item_sales (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	item_type TEXT NOT NULL,
	item_id UUID NOT NULL,
	listing_fee_tx_id UUID NOT NULL,
	owner_id UUID NOT NULL REFERENCES players(id),

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
	WHEN 'mech' THEN 
		record_found := EXISTS (
			SELECT id
			FROM mechs
			WHERE id = NEW.item_id
				AND owner_id = NEW.owner_id
		);
	END CASE;

	IF record_found = FALSE THEN
		RAISE EXCEPTION '% not found', NEW.item_type;
	END IF;

    RETURN NULL;
END;
$checkItemOwnerConstraint$
    LANGUAGE plpgsql;

CREATE TRIGGER checkItemOwnerConstraint
    BEFORE INSERT OR UPDATE
    ON item_sales
EXECUTE PROCEDURE checkItemOwnerConstraint();

CREATE TABLE item_sales_buyout_price_history (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	item_sale_id UUID NOT NULL,
	buyout_price TEXT NOT NULL
);

COMMIT;
