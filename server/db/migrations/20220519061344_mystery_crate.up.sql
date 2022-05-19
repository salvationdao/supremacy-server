CREATE TABLE storefront_mystery_crates
(
    id                 UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    mystery_crate_type TEXT        NOT NULL,
    price              numeric(28) NOT NULL,
    amount             INT         NOT NULL,
    amount_sold        INT         NOT NULL,
    faction_id         UUID        NOT NULL,
    deleted_at         TIMESTAMPTZ,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mystery_crate
(
    id           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    type         TEXT        NOT NULL,
    faction_id   UUID        NOT NULL REFERENCES factions (id),
    label        TEXT        NOT NULL,
    opened       BOOLEAN     NOT NULL DEFAULT false,
    locked_until TIMESTAMPTZ NOT NULL,
    purchased    BOOLEAN     NOT NULL DEFAULT false
);

CREATE TABLE mystery_crate_blueprints
(
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mystery_crate_id UUID               NOT NULL REFERENCES mystery_crate (id),
    blueprint_type   TEMPLATE_ITEM_TYPE NOT NULL,
    blueprint_id     UUID               NOT NULL
);

-- DROP TYPE IF EXISTS ABILITY_TYPE_ENUM;
-- CREATE TYPE ABILITY_TYPE_ENUM AS ENUM ('AIRSTRIKE','NUKE','REPAIR', 'ROB','REINFORCEMENTS','ROBOT DOGS','OVERCHARGE');