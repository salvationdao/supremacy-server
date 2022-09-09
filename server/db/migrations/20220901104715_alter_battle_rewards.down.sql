UPDATE kv SET value = '0.6' WHERE key = 'first_rank_faction_reward_ratio';
UPDATE kv SET value = '0.25' WHERE key = 'second_rank_faction_reward_ratio';
UPDATE kv SET value = '0.15' WHERE key = 'third_rank_faction_reward_ratio';

ALTER TABLE battle_queue_fees
    DROP COLUMN IF EXISTS bonus_sups_tx_id;