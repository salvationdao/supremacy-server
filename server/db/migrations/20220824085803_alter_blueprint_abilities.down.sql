ALTER TABLE blueprint_player_abilities
    DROP COLUMN IF EXISTS display_on_mini_map,
    DROP COLUMN IF EXISTS launching_delay_seconds;