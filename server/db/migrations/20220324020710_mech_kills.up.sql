ALTER TABLE user_stats
    ADD mech_kill_count int NOT NULL default 0;

ALTER TABLE faction_stats
    ADD mech_kill_count int NOT NULL default 0 ;