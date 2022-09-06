ALTER TABLE blueprint_player_abilities
    ADD COLUMN IF NOT EXISTS display_on_mini_map bool not null DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS launching_delay_seconds int NOT NULL DEFAULT 0;

UPDATE blueprint_player_abilities
SET display_on_mini_map = true
WHERE game_client_ability_id = 12 OR game_client_ability_id = 13; -- EMP OR Hacker drone
