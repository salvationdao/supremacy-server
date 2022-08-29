ALTER TABLE game_abilities
    ADD COLUMN should_check_team_kill bool not null DEFAULT false,
    ADD COLUMN maximum_team_kill_tolerant_count int not null DEFAULT 0;

