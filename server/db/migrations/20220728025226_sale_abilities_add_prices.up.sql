ALTER TABLE
    sale_player_abilities
ADD
    COLUMN current_price NUMERIC(28) NOT NULL DEFAULT 100000000000000000000;

ALTER TABLE
    blueprint_player_abilities
ADD
    COLUMN inventory_limit int NOT NULL DEFAULT 1;

-- NUKE
UPDATE
    blueprint_player_abilities
SET
    inventory_limit = 1
WHERE
    game_client_ability_id = 1;

-- AIRSTRIKE
UPDATE
    blueprint_player_abilities
SET
    inventory_limit = 5
WHERE
    game_client_ability_id = 0;

-- SHIELD OVERDRIVE, LANDMINES, EMP, INCOGNITO
UPDATE
    blueprint_player_abilities
SET
    inventory_limit = 8
WHERE
    game_client_ability_id IN (10, 11, 12, 15);

-- HACKER DRONE, DRONE CAMERA, BLACKOUT
UPDATE
    blueprint_player_abilities
SET
    inventory_limit = 10
WHERE
    game_client_ability_id IN (13, 14, 16);