DROP TYPE IF EXISTS COUPON_ITEM_TYPE;
CREATE TYPE COUPON_ITEM_TYPE AS ENUM ('SUPS', 'WEAPON_CRATE', 'MECH_CRATE', 'GENESIS_MECH');

CREATE TABLE coupons
(
    id          UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    code        TEXT UNIQUE      NOT NULL DEFAULT random_string(6),
    redeemed    BOOLEAN          NOT NULL DEFAULT false,
    expiry_date TIMESTAMPTZ      NOT NULL DEFAULT NOW() + INTERVAL '30' DAY,
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE coupon_items
(
    id             UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    coupon_id      UUID             NOT NULL REFERENCES coupons (id),
    item_type      COUPON_ITEM_TYPE NOT NULL,
    item_id        UUID,
    claimed        BOOLEAN          NOT NULL DEFAULT false,
    amount         numeric(28),
    transaction_id TEXT,
    created_at     TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

-- seed coupon codes
DO
$$
    DECLARE
        couponId UUID;
    BEGIN
        FOR count IN 1..1500
            LOOP
                INSERT INTO coupons DEFAULT VALUES RETURNING id INTO couponId;
                INSERT INTO coupon_items (coupon_id, item_type, amount)
                VALUES (couponId, 'SUPS', '300000000000000000000');
                INSERT INTO coupon_items (coupon_id, item_type) VALUES (couponId, 'WEAPON_CRATE');
                INSERT INTO coupon_items (coupon_id, item_type) VALUES (couponId, 'MECH_CRATE');

            end loop;
    END;
$$;

UPDATE mystery_crate
SET locked_until = '2022-07-07 21:00:00.000 +0800';
