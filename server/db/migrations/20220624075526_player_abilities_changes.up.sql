DELETE
FROM player_abilities;

ALTER TABLE
    player_abilities
    ADD
        COLUMN IF NOT EXISTS count             INT         NOT NULL DEFAULT 0,
    ADD
        COLUMN IF NOT EXISTS last_purchased_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    DROP COLUMN IF EXISTS purchased_at,
    DROP COLUMN IF EXISTS game_client_ability_id,
    DROP COLUMN IF EXISTS label,
    DROP COLUMN IF EXISTS colour,
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS text_colour,
    DROP COLUMN IF EXISTS location_select_type;

ALTER TABLE
    sale_player_abilities
    ADD
        COLUMN IF NOT EXISTS amount_sold INT NOT NULL DEFAULT 0,
    ADD
        COLUMN IF NOT EXISTS sale_limit  INT NOT NULL DEFAULT 10;

ALTER TABLE
    player_abilities
    ADD
        UNIQUE (owner_id, blueprint_id);
