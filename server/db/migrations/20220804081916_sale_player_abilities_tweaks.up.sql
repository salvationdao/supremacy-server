UPDATE
    blueprint_player_abilities
SET
    rarity_weight = 4
WHERE
    game_client_ability_id IN (15);

UPDATE
    blueprint_player_abilities
SET
    rarity_weight = 6
WHERE
    game_client_ability_id IN (12);

UPDATE
    blueprint_player_abilities
SET
    rarity_weight = 8
WHERE
    game_client_ability_id IN (10);

UPDATE
    blueprint_player_abilities
SET
    rarity_weight = 10
WHERE
    game_client_ability_id IN (13, 16, 14);

ALTER TABLE
    blueprint_player_abilities
ADD
    COLUMN cooldown_seconds INT NOT NULL DEFAULT 180;

UPDATE
    blueprint_player_abilities
SET
    cooldown_seconds = 120
WHERE
    game_client_ability_id IN (13);

UPDATE
    blueprint_player_abilities
SET
    cooldown_seconds = 180
WHERE
    game_client_ability_id IN (16);

UPDATE
    blueprint_player_abilities
SET
    cooldown_seconds = 240
WHERE
    game_client_ability_id IN (10, 12);

UPDATE
    blueprint_player_abilities
SET
    cooldown_seconds = 360
WHERE
    game_client_ability_id IN (15);

UPDATE
    blueprint_player_abilities
SET
    cooldown_seconds = 600
WHERE
    game_client_ability_id IN (14);

ALTER TABLE
    player_abilities
ADD
    COLUMN cooldown_expires_on timestamptz NOT NULL DEFAULT now();

INSERT INTO
    game_maps ("name", max_spawns)
VALUES
    ('CityBlockArena', 9),
    ('RedMountainMine', 9);

UPDATE
    game_maps
SET
    disabled_at = NULL
WHERE
    "name" IN ('AokigaharaForest', 'CloudKu', 'TheHive');