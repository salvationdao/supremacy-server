ALTER TABLE item_sales RENAME COLUMN sold_to TO sold_by;
ALTER TABLE item_keycard_sales RENAME COLUMN sold_to TO sold_by;

DROP TABLE marketplace_events;

DROP TYPE MARKETPLACE_EVENT;
