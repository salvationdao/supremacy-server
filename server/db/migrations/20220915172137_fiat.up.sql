ALTER TABLE players ADD COLUMN stripe_customer_id TEXT UNIQUE;

-----------------
-- Products
-----------------
DROP TYPE IF EXISTS FIAT_PRODUCT_TYPES;
CREATE TYPE FIAT_PRODUCT_TYPES AS ENUM (
	'starter_package',
	'mystery_crate',
	'mech_skin',
	'weapon_skin',
	'mech_animation'
);

CREATE TABLE fiat_products (
	id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
	product_type FIAT_PRODUCT_TYPES NOT NULL,
	faction_id UUID NOT NULL,
	name TEXT NOT NULL,
	description TEXT NOT NULL,
	amount_sold INT NOT NULL DEFAULT 0,

	deleted_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fiat_product_pricings (
	fiat_product_id UUID NOT NULL REFERENCES fiat_products (id),	
	currency_code TEXT NOT NULL,
	amount DECIMAL NOT NULL,

	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	PRIMARY KEY (fiat_product_id, currency_code)
);

DROP TYPE IF EXISTS FIAT_PRODUCT_ITEM_TYPES;
CREATE TYPE FIAT_PRODUCT_ITEM_TYPES AS ENUM ('mech_package', 'weapon_package', 'single_item');

CREATE TABLE fiat_product_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES fiat_products (id),
	name TEXT NOT NULL,
    item_type FIAT_PRODUCT_ITEM_TYPES NOT NULL,

    deleted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fiat_product_item_blueprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	product_item_id UUID REFERENCES fiat_product_items (id),
	mech_blueprint_id UUID REFERENCES blueprint_mechs (id),
	mech_animation_blueprint_id UUID REFERENCES blueprint_mech_animation (id),
	mech_skin_blueprint_id UUID REFERENCES blueprint_mech_skin (id),
	utility_blueprint_id UUID REFERENCES blueprint_utility (id),
	weapon_blueprint_id UUID REFERENCES blueprint_weapons (id),
	weapon_skin_blueprint_id UUID REFERENCES blueprint_weapon_skin (id),
	ammo_blueprint_id UUID REFERENCES blueprint_ammo (id),
	power_core_blueprint_id UUID REFERENCES blueprint_power_cores (id),
	player_ability_blueprint_id UUID REFERENCES blueprint_player_abilities (id)
);

---------------------------------
-- Storefront - Mysytery Crates
---------------------------------

-- RM: Mech Create
INSERT INTO fiat_products (id, product_type, faction_id, name, description)
SELECT 'b8b3f9d6-7b4d-4935-a999-51119999c4a8' AS id, 'mystery_crate' AS product_type, faction_id, label AS name, description
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT 'b8b3f9d6-7b4d-4935-a999-51119999c4a8' AS id, 'SUPS' AS currency_code, price AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT 'b8b3f9d6-7b4d-4935-a999-51119999c4a8' AS id, 'USD' AS currency_code, 20000 AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

-- BC: Mech Create
INSERT INTO fiat_products (id, product_type, faction_id, name, description)
SELECT '838d3f28-5f66-4235-a92e-1ecf8b89edf8' AS id, 'mystery_crate' AS product_type, faction_id, label AS name, description
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT '838d3f28-5f66-4235-a92e-1ecf8b89edf8' AS id, 'SUPS' AS currency_code, price AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT '838d3f28-5f66-4235-a92e-1ecf8b89edf8' AS id, 'USD' AS currency_code, 20000 AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

-- ZHI: Mech Create
INSERT INTO fiat_products (id, product_type, faction_id, name, description)
SELECT '56fbe6d4-8bc0-4fb7-9868-2c9e5d845c5e' AS id, 'mystery_crate' AS product_type, faction_id, label AS name, description
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '880db344-e405-428d-84e5-6ebebab1fe6d';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT '56fbe6d4-8bc0-4fb7-9868-2c9e5d845c5e' AS id, 'SUPS' AS currency_code, price AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '880db344-e405-428d-84e5-6ebebab1fe6d';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT '56fbe6d4-8bc0-4fb7-9868-2c9e5d845c5e' AS id, 'USD' AS currency_code, 20000 AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '880db344-e405-428d-84e5-6ebebab1fe6d';

-- RM: Weapon Create
INSERT INTO fiat_products (id, product_type, faction_id, name, description)
SELECT '52a8bc78-fe1b-4ed3-b03c-f5400da957fe' AS id, 'mystery_crate' AS product_type, faction_id, label AS name, description
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT '52a8bc78-fe1b-4ed3-b03c-f5400da957fe' AS id, 'SUPS' AS currency_code, price AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT '52a8bc78-fe1b-4ed3-b03c-f5400da957fe' AS id, 'USD' AS currency_code, 10000 AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

-- BC: Weapon Create
INSERT INTO fiat_products (id, product_type, faction_id, name, description)
SELECT 'b576e6e6-06c3-4230-aff2-757bde0b20cc' AS id, 'mystery_crate' AS product_type, faction_id, label AS name, description
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT 'b576e6e6-06c3-4230-aff2-757bde0b20cc' AS id, 'SUPS' AS currency_code, price AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT 'b576e6e6-06c3-4230-aff2-757bde0b20cc' AS id, 'USD' AS currency_code, 10000 AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

-- ZHI: Weapon Create
INSERT INTO fiat_products (id, product_type, faction_id, name, description)
SELECT '9bb21b1b-df3f-4012-8e47-b8f7ceb5a753' AS id, 'mystery_crate' AS product_type, faction_id, label AS name, description
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '880db344-e405-428d-84e5-6ebebab1fe6d';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT '9bb21b1b-df3f-4012-8e47-b8f7ceb5a753' AS id, 'SUPS' AS currency_code, price AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '880db344-e405-428d-84e5-6ebebab1fe6d';

INSERT INTO fiat_product_pricings (fiat_product_id, currency_code, amount)
SELECT '9bb21b1b-df3f-4012-8e47-b8f7ceb5a753' AS id, 'USD' AS currency_code, 10000 AS amount
FROM storefront_mystery_crates
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '880db344-e405-428d-84e5-6ebebab1fe6d';

-- Map Storefront to Fiat Product
ALTER TABLE storefront_mystery_crates ADD COLUMN fiat_product_id UUID UNIQUE REFERENCES fiat_products (id);

UPDATE storefront_mystery_crates
SET fiat_product_id = 'b8b3f9d6-7b4d-4935-a999-51119999c4a8'
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

UPDATE storefront_mystery_crates
SET fiat_product_id = '838d3f28-5f66-4235-a92e-1ecf8b89edf8'
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

UPDATE storefront_mystery_crates
SET fiat_product_id = '56fbe6d4-8bc0-4fb7-9868-2c9e5d845c5e'
WHERE mystery_crate_type = 'MECH'
	AND faction_id = '880db344-e405-428d-84e5-6ebebab1fe6d';

UPDATE storefront_mystery_crates
SET fiat_product_id = '52a8bc78-fe1b-4ed3-b03c-f5400da957fe'
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

UPDATE storefront_mystery_crates
SET fiat_product_id = 'b576e6e6-06c3-4230-aff2-757bde0b20cc'
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

UPDATE storefront_mystery_crates
SET fiat_product_id = '9bb21b1b-df3f-4012-8e47-b8f7ceb5a753'
WHERE mystery_crate_type = 'WEAPON'
	AND faction_id = '880db344-e405-428d-84e5-6ebebab1fe6d';

ALTER TABLE storefront_mystery_crates ALTER COLUMN fiat_product_id SET NOT NULL;

------------------
-- Shopping Cart
------------------

CREATE TABLE shopping_carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES players (id),
	locked BOOLEAN NOT NULL DEFAULT FALSE,

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE shopping_cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	shopping_cart_id UUID NOT NULL REFERENCES shopping_carts (id) ON DELETE CASCADE,
	product_id UUID NOT NULL REFERENCES fiat_products (id),
	quantity INT NOT NULL DEFAULT 1,

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

---------------------------
-- Orders/Billing History
---------------------------

DROP TYPE IF EXISTS ORDER_STATUSES;
CREATE TYPE ORDER_STATUSES AS ENUM ('pending', 'completed', 'refunded');

DROP TYPE IF EXISTS PAYMENT_METHODS;
CREATE TYPE PAYMENT_METHODS AS ENUM ('sups', 'stripe');

DROP SEQUENCE IF EXISTS order_num_seq;
CREATE SEQUENCE order_num_seq AS BIGINT START WITH 10000 INCREMENT BY 1;

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	order_number BIGINT NOT NULL DEFAULT NEXTVAL('order_num_seq'), 
    user_id UUID NOT NULL REFERENCES players (id),
	order_status ORDER_STATUSES NOT NULL DEFAULT 'pending',
	payment_method PAYMENT_METHODS NOT NULL,
	txn_reference TEXT NOT NULL, -- used for storing receipt references
	currency TEXT NOT NULL DEFAULT 'USD',

	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	order_id UUID NOT NULL REFERENCES orders (id),
	fiat_product_id UUID NOT NULL REFERENCES fiat_products (id),
	name TEXT NOT NULL,
	description TEXT NOT NULL,

	quantity INT NOT NULL DEFAULT 1,
	amount DECIMAL NOT NULL -- in whole numbers for non SUP prices
);
