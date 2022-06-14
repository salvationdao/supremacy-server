ALTER TABLE weapons
    ADD COLUMN locked_to_mech BOOL NOT NULL DEFAULT FALSE;

ALTER TABLE utility
    ADD COLUMN locked_to_mech BOOL NOT NULL DEFAULT FALSE;

ALTER TABLE mech_skin
    ADD COLUMN locked_to_mech BOOL NOT NULL DEFAULT FALSE;

UPDATE weapons
SET locked_to_mech = TRUE
WHERE weapons.label ILIKE '%rocket%';
UPDATE utility
SET locked_to_mech = TRUE;
UPDATE mech_skin
SET locked_to_mech = TRUE;
