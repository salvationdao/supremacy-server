ALTER TABLE collection_items
	ADD COLUMN locked_to_marketplace BOOL NOT NULL DEFAULT FALSE;
