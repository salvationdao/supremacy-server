BEGIN;
ALTER TABLE battle_queue ADD COLUMN queue_fee_tx_id TEXT;
ALTER TABLE battle_queue ADD COLUMN queue_notification_fee_tx_id TEXT;
ALTER TABLE battle_queue ADD COLUMN queue_fee_tx_id_refund TEXT;
ALTER TABLE battle_queue ADD COLUMN queue_notification_fee_tx_id_refund TEXT;

ALTER TABLE battle_queue DROP CONSTRAINT  battle_queue_pkey;
ALTER TABLE battle_queue ADD COLUMN id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid();

ALTER TABLE battle_queue ADD COLUMN deleted_at TIMESTAMPTZ;
ALTER TABLE battle_queue ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE battle_queue ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
COMMIT;
