BEGIN;
ALTER TABLE battle_queue DROP COLUMN queue_fee_tx_id;
ALTER TABLE battle_queue DROP COLUMN queue_notification_fee_tx_id;
ALTER TABLE battle_queue DROP COLUMN queue_fee_tx_id_refund;
ALTER TABLE battle_queue DROP COLUMN queue_notification_fee_tx_id_refund;

ALTER TABLE battle_queue DROP CONSTRAINT  battle_queue_pkey;
ALTER TABLE battle_queue ADD CONSTRAINT battle_queue_pkey PRIMARY KEY (mech_id);

ALTER TABLE battle_queue DROP COLUMN deleted_at;
ALTER TABLE battle_queue DROP COLUMN updated_at;
ALTER TABLE battle_queue DROP COLUMN created_at;
COMMIT;
