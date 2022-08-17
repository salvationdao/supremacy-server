UPDATE
    blueprint_player_abilities
SET
    description = 'Deploy a drone onto the battlefield that hacks into the nearest War Machine and disrupts their targeting systems.'
WHERE
    game_client_ability_id = 13;

DELETE FROM
    blueprint_player_abilities
WHERE
    game_client_ability_id = 18;

ALTER TABLE
    mech_move_command_logs DROP COLUMN is_moving;
