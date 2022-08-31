ALTER TABLE game_abilities
    ADD COLUMN IF NOT EXISTS should_check_team_kill bool not null DEFAULT false,
    ADD COLUMN IF NOT EXISTS maximum_team_kill_tolerant_count int not null DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ignore_self_kill bool not null DEFAULT false;
