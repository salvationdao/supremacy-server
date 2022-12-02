UPDATE blueprint_utility
SET deleted_at = NULL
WHERE type = 'SHIELD';

ALTER TABLE blueprint_utility_shield_old
    RENAME TO blueprint_utility_shield;

UPDATE blueprint_utility_shield
SET deleted_at = NULL;

ALTER TABLE blueprint_mechs
    DROP COLUMN shield_type_id,
    DROP COLUMN shield_max,
    DROP COLUMN shield_recharge_rate,
    DROP COLUMN shield_recharge_power_cost;

DROP TABLE blueprint_shield_types;
