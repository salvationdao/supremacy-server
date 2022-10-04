ALTER TABLE coupons ADD COLUMN redeemed_by_id UUID REFERENCES players (id);
