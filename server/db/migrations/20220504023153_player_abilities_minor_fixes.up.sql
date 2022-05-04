BEGIN;

UPDATE consumed_abilities ca
SET location_select_type = (SELECT location_select_type FROM blueprint_player_abilities bpa WHERE bpa.id = ca.blueprint_id)
WHERE location_select_type IS NULL;

ALTER TABLE consumed_abilities
ALTER COLUMN location_select_type SET NOT NULL;

-- Update Landmine
UPDATE blueprint_player_abilities
SET colour = '#d9674c', text_colour = '#d9674c'
WHERE game_client_ability_id = 11;
-- Update Shield Overdrive
UPDATE blueprint_player_abilities
SET description = 'An airdropped module that can be picked up by War Machines to put their shield modules into overdrive, boosting their shields.'
WHERE game_client_ability_id = 10;
-- Update Hacker Drone
UPDATE blueprint_player_abilities
SET colour = '#FF5861', text_colour = '#FF5861'
WHERE game_client_ability_id = 13;
-- Update Blackout
UPDATE blueprint_player_abilities
SET description = 'Drops a cloud of nanorobotics, concealing War Machine locations and disabling their navigations.'
WHERE game_client_ability_id = 16;
-- Update Drone Camera
UPDATE blueprint_player_abilities
SET colour = '#7676F7', text_colour = '#7676F7'
WHERE game_client_ability_id = 14;
-- Update Incognito
UPDATE blueprint_player_abilities
SET colour = '#006600', text_colour = '#006600', description = 'Blocks GABs radar technology from locating a War Machine''s position, hiding it from the minimap.'
WHERE game_client_ability_id = 15;

COMMIT;
