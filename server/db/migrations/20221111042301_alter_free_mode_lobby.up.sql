ALTER TABLE battle_lobbies
    ADD COLUMN is_free_team_mode BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE battle_lobbies_mechs
    ADD COLUMN team_number INT NOT NULL DEFAULT 0 CHECK(team_number <= 3); -- team number of battle lobby mech