ALTER TABLE
    player_abilities
ADD
    COLUMN cooldown_expires_on timestamptz NOT NULL DEFAULT now();

INSERT INTO
    game_maps ("name", max_spawns)
VALUES
    ('CityBlockArena', 9),
    ('RedMountainMine', 9) ON CONFLICT ("name") DO NOTHING;

UPDATE
    game_maps
SET
    disabled_at = NULL
WHERE
    "name" IN ('AokigaharaForest', 'CloudKu', 'TheHive');