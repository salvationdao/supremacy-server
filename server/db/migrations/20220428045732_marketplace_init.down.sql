BEGIN;

DROP TRIGGER checkItemOwnerConstraint ON item_sales;
DROP FUNCTION checkItemOwnerConstraint;
DROP TABLE
	item_sales,
	item_sales_bid_history,
	item_sales_buyout_price_history
;

COMMIT;
