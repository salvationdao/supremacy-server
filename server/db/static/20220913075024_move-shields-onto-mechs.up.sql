CREATE TABLE blueprint_shield_types
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid() NOT NULL,
    label       TEXT                                               NOT NULL,
    description TEXT                                               NOT NULL,
    deleted_at  TIMESTAMP WITH TIME ZONE,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()             NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()             NOT NULL
);

ALTER TABLE blueprint_mechs
    ADD COLUMN shield_type_id                UUID REFERENCES blueprint_shield_types (id), -- we cannot add the NOT NULL here
    ADD COLUMN shield_max                 INT  NOT NULL DEFAULT 0,
    ADD COLUMN shield_recharge_rate       INT  NOT NULL DEFAULT 0,
    ADD COLUMN shield_recharge_power_cost INT  NOT NULL DEFAULT 0;

UPDATE blueprint_utility_shield
SET deleted_at = NOW();
ALTER TABLE blueprint_utility_shield
    RENAME TO blueprint_utility_shield_old;

UPDATE blueprint_utility
SET deleted_at = NOW()
WHERE type = 'SHIELD';
