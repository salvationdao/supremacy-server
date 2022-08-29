ALTER TABLE game_abilities
    ADD COLUMN should_check_team_kill bool not null DEFAULT false,
    ADD COLUMN maximum_team_kill_tolerant_count int not null DEFAULT 0;

UPDATE
    game_abilities
SET
    should_check_team_kill = true
WHERE
    game_client_ability_id = 0 OR -- AIRSTRIKE
    game_client_ability_id = 1 OR -- NUKE
    game_client_ability_id = 5;   -- OC

UPDATE
    game_abilities
SET
    maximum_team_kill_tolerant_count = 3
WHERE
    game_client_ability_id = 5;   -- OC