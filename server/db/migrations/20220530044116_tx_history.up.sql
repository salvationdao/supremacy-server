CREATE TABLE store_purchase_history
(
    id           UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    player_id    UUID        NOT NULL,
    item_id      UUID        NOT NULL,
    item_type    ITEM_TYPE   NOT NULL,
    amount       NUMERIC(28) NOT NULL,
    description  TEXT        NOT NULL,
    tx_id        TEXT        NOT NULL,
    refund_tx_id TEXT,
    refunded_at  TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
