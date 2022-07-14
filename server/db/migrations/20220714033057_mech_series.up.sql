CREATE TABLE availabilities ( 
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	reason TEXT NOT NULL,
	available_at TIMESTAMPTZ NOT NULL
);

INSERT INTO availabilities (id, reason, available_at) VALUES ('518ffb3f-8595-4db0-b9ea-46285f6ccd2f', 'Nexus Release', '2022-07-22 00:00:00');

ALTER TABLE blueprint_mechs 
	ADD COLUMN availability_id UUID REFERENCES availabilities (id);
