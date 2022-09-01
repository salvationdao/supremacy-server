ALTER TABLE game_abilities
    DROP COLUMN IF EXISTS should_check_team_kill,
    DROP COLUMN IF EXISTS maximum_team_kill_tolerant_count,
    DROP COLUMN IF EXISTS ignore_self_kill;