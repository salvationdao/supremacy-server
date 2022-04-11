ALTER TABLE battles
    ADD COLUMN IF NOT EXISTS  started_battle_seconds NUMERIC(78,0) unique,
    ADD COLUMN IF NOT EXISTS  ended_battle_seconds NUMERIC(78,0) unique;

ALTER TABLE multipliers
    ADD COLUMN IF NOT EXISTS remain_seconds INT NOT NULL DEFAULT 600; -- ten minutes

-- set won battle to 3 minutes
UPDATE multipliers SET remain_seconds = 180 WHERE key = 'won battle';

-- update from player multipliers
ALTER TABLE user_multipliers
    ADD COLUMN IF NOT EXISTS obtained_at_battle_seconds NUMERIC(78, 0) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS expires_at_battle_seconds NUMERIC(78, 0) NOT NULL DEFAULT 600;

ALTER TABLE battle_contributions
    ADD COLUMN IF NOT EXISTS refund_transaction_id TEXT UNIQUE;

UPDATE user_multipliers SET expires_at_battle_seconds = 0 WHERE until_battle_number < (SELECT battle_number FROM battles ORDER BY battle_number DESC LIMIT 1);

UPDATE punish_options SET description = 'Restrict player to select location for 6 hours', punish_duration_hours = 6 WHERE key = 'restrict_location_select';
UPDATE punish_options SET description = 'Restrict player to chat for 6 hours', punish_duration_hours = 6 WHERE key = 'restrict_chat';
UPDATE punish_options SET description = 'Restrict player to contribute sups for 6 hours', punish_duration_hours = 6 WHERE key = 'restrict_sups_contribution';

UPDATE spoils_of_war set amount_sent = amount where amount_sent > amount;

ALTER TABLE spoils_of_war ADD CONSTRAINT constraint_amount_sent_lte
    CHECK (amount_sent <= amount);