UPDATE
    consumed_abilities ca
SET
    location_select_type = (
        SELECT
            location_select_type
        FROM
            blueprint_player_abilities bpa
        WHERE
            bpa.id = ca.blueprint_id
    )
WHERE
    location_select_type IS NULL;

ALTER TABLE
    consumed_abilities
ALTER COLUMN
    location_select_type
SET
    NOT NULL;



ALTER TABLE
    player_abilities DROP CONSTRAINT player_abilities_location_select_type_check;

ALTER TABLE
    player_abilities
ALTER COLUMN
    location_select_type TYPE LOCATION_SELECT_TYPE_ENUM USING location_select_type :: LOCATION_SELECT_TYPE_ENUM;

ALTER TABLE
    consumed_abilities DROP CONSTRAINT consumed_abilities_location_select_type_check;

ALTER TABLE
    consumed_abilities
ALTER COLUMN
    location_select_type TYPE LOCATION_SELECT_TYPE_ENUM USING location_select_type :: LOCATION_SELECT_TYPE_ENUM;

UPDATE
    player_abilities
SET
    colour = '#d9674c',
    text_colour = '#d9674c',
    location_select_type = 'LINE_SELECT',
    description = 'Deploy a line of explosives that detonate when a War Machine is detected within its proximity.'
WHERE
    game_client_ability_id = 11;

UPDATE
    consumed_abilities
SET
    colour = '#d9674c',
    text_colour = '#d9674c',
    location_select_type = 'LINE_SELECT',
    description = 'Deploy a line of explosives that detonate when a War Machine is detected within its proximity.'
WHERE
    game_client_ability_id = 11;

UPDATE
    player_abilities
SET
    description = 'Airdrop a module that can be picked up by War Machines to put their shield modules into overdrive, boosting their shields.'
WHERE
    game_client_ability_id = 10;

UPDATE
    consumed_abilities
SET
    description = 'Airdrop a module that can be picked up by War Machines to put their shield modules into overdrive, boosting their shields.'
WHERE
    game_client_ability_id = 10;



UPDATE
    player_abilities
SET
    colour = '#FF5861',
    text_colour = '#FF5861',
    location_select_type = 'LOCATION_SELECT',
    description = 'Deploy a drone onto the battlefield that hacks into the nearest War Machine and disrupts their targeting systems.'
WHERE
    game_client_ability_id = 13;

UPDATE
    consumed_abilities
SET
    colour = '#FF5861',
    text_colour = '#FF5861',
    location_select_type = 'LOCATION_SELECT',
    description = 'Deploy a drone onto the battlefield that hacks into the nearest War Machine and disrupts their targeting systems.'
WHERE
    game_client_ability_id = 13;



UPDATE
    player_abilities
SET
    description = 'Release a cloud of nanorobotics, concealing War Machine locations and disabling their navigations.'
WHERE
    game_client_ability_id = 16;

UPDATE
    consumed_abilities
SET
    description = 'Release a cloud of nanorobotics, concealing War Machine locations and disabling their navigations.'
WHERE
    game_client_ability_id = 16;



UPDATE
    player_abilities
SET
    colour = '#7676F7',
    text_colour = '#7676F7',
    description = 'Override GABs visual broadcasting for a short period.'
WHERE
    game_client_ability_id = 14;

UPDATE
    consumed_abilities
SET
    colour = '#7676F7',
    text_colour = '#7676F7',
    description = 'Override GABs visual broadcasting for a short period.'
WHERE
    game_client_ability_id = 14;



UPDATE
    player_abilities
SET
    colour = '#006600',
    text_colour = '#006600',
    description = 'Block GABs radar technology from locating a War Machine''s position, hiding it from the minimap.'
WHERE
    game_client_ability_id = 15;

UPDATE
    consumed_abilities
SET
    colour = '#006600',
    text_colour = '#006600',
    description = 'Block GABs radar technology from locating a War Machine''s position, hiding it from the minimap.'
WHERE
    game_client_ability_id = 15;



UPDATE
    player_abilities
SET
    description = 'Create a short burst of electromagnetic energy that will disrupt War Machine operations in an area for a brief period of time.'
WHERE
    game_client_ability_id = 12;

UPDATE
    consumed_abilities
SET
    description = 'Create a short burst of electromagnetic energy that will disrupt War Machine operations in an area for a brief period of time.'
WHERE
    game_client_ability_id = 12;