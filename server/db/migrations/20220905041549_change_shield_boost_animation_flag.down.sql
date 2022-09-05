UPDATE
    blueprint_player_abilities
SET
    display_on_mini_map = false,
    mini_map_display_effect_type = 'NONE'
WHERE
    game_client_ability_id = 10; -- Shield Overdrive