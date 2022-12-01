CREATE TABLE faction_palettes (
    faction_id uuid PRIMARY KEY NOT NULL REFERENCES factions (id),
    "primary" text NOT NULL,
    "text" text NOT NULL,
    background text NOT NULL,
    s100 text NOT NULL,
    s200 text NOT NULL,
    s300 text NOT NULL,
    s400 text NOT NULL,
    s500 text NOT NULL,
    s600 text NOT NULL,
    s700 text NOT NULL,
    s800 text NOT NULL,
    s900 text NOT NULL
);

ALTER TABLE factions
    DROP COLUMN primary_color,
    DROP COLUMN secondary_color,
    DROP COLUMN background_color;

