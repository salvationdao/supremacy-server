ALTER TABLE game_abilities
    ADD COLUMN sups_cost    TEXT DEFAULT '0' NOT NULL,
    ADD COLUMN current_sups TEXT DEFAULT '0' NOT NULL,
    DROP COLUMN count_per_battle;
