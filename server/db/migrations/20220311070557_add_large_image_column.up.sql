ALTER TABLE mechs ADD COLUMN large_image_url TEXT;
UPDATE mechs SET large_image_url = image_url;
ALTER TABLE mechs ALTER COLUMN large_image_url SET NOT NULL;

ALTER TABLE templates ADD COLUMN large_image_url TEXT;
UPDATE templates SET large_image_url = image_url;
ALTER TABLE templates ALTER COLUMN large_image_url SET NOT NULL;