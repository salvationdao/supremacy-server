DROP TABLE battle_lobby_extra_sups_rewards;

ALTER TABLE battle_lobbies
    DROP COLUMN IF EXISTS max_deploy_per_player;

ALTER TABLE battle_lobbies
    RENAME access_code TO password;
