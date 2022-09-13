ALTER TABLE blueprint_mechs
    ADD COLUMN shield_max INT NOT NULL DEFAULT 0,
    ADD COLUMN shield_recharge_rate INT NOT NULL DEFAULT 0,
    ADD COLUMN shield_recharge_power_cost INT NOT NULL DEFAULT 0;

UPDATE blueprint_utility_shield SET deleted_at = NOW();
ALTER TABLE blueprint_utility_shield RENAME TO blueprint_utility_shield_old;

UPDATE blueprint_utility SET deleted_at = NOW() WHERE type = 'SHIELD';
