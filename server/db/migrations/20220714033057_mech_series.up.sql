CREATE TABLE availabilities ( 
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	reason TEXT NOT NULL,
	available_at TIMESTAMPTZ NOT NULL
);

INSERT INTO availabilities (id, reason, available_at) VALUES
	('518ffb3f-8595-4db0-b9ea-46285f6ccd2f', 'Nexus Release', '2023-07-22 00:00:00') -- NOTE: this needs to be updated manually to correct date
;

ALTER TABLE blueprint_mechs 
	ADD COLUMN availability_id UUID REFERENCES availabilities (id);

UPDATE blueprint_mechs 
SET availability_id = '518ffb3f-8595-4db0-b9ea-46285f6ccd2f'
WHERE model_id IN (
	'02ba91b7-55dc-450a-9fbd-e7337ae97a2b',
	'7068ab3e-89dc-4ac1-bcbb-1089096a5eda',
	'3dc5888b-f5ff-4d08-a520-26fd3681707f',
	'0639ebde-fbba-498b-88ac-f7122ead9c90',
	'fc9546d0-9682-468e-af1f-24eb1735315b',
	'df1ac803-0a90-4631-b9e0-b62a44bdadff'
);
