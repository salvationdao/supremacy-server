ALTER TABLE coupons
    ADD COLUMN IF NOT EXISTS redeemed_at timestamptz;

UPDATE coupons
SET redeemed_at = NOW()
WHERE redeemed = TRUE;