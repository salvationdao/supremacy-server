ALTER TABLE faction_palettes
    ADD COLUMN contrast_primary text NOT NULL DEFAULT '#FFFFFF',
    ADD COLUMN contrast_background text NOT NULL DEFAULT '#FFFFFF';

