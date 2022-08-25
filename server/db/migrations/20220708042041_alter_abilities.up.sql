ALTER TABLE game_abilities
    ADD COLUMN IF NOT EXISTS location_select_type location_select_type_enum not null default 'LOCATION_SELECT';

ALTER TYPE ABILITY_TYPE_ENUM ADD VALUE 'LANDMINE';