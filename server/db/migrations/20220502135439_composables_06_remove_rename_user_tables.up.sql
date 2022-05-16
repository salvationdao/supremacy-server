DROP TABLE battles_user_votes;
DROP TABLE users;
DROP TABLE battles_winner_records;
DROP TABLE config;
DROP TABLE issued_contract_rewards;
DROP TABLE pending_transactions;

ALTER TABLE user_spoils_of_war
    RENAME TO player_spoils_of_war;

ALTER TABLE user_multipliers
    RENAME TO player_multipliers;

ALTER TABLE user_stats
    RENAME TO player_stats;

ALTER TABLE battles_viewers
    RENAME TO battle_viewers;

ALTER TABLE asset_repair
    RENAME TO mech_repair;
