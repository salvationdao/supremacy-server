BEGIN;

DROP TRIGGER checkItemOwnerConstraint ON item_sales;
DROP FUNCTION checkItemOwnerConstraint;
DROP TABLE item_sales;

COMMIT;
