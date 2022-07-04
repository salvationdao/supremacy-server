ALTER TABLE
    blueprint_player_abilities
ADD
    COLUMN rarity_weight INT;

ALTER TABLE
    consumed_abilities
ADD
    COLUMN rarity_weight INT;

-- Update rarities of all player abilities (except for landmines, 11; is a rarer ability)
UPDATE
    blueprint_player_abilities
SET
    rarity_weight = 10
WHERE
    game_client_ability_id IN (10, 12, 13, 14, 15, 16);

UPDATE
    consumed_abilities ca
SET
    rarity_weight = 10
WHERE
    game_client_ability_id IN (10, 12, 13, 14, 15, 16);

-- Add nuke and airstrike as player abilities
INSERT INTO
    blueprint_player_abilities (
        game_client_ability_id,
        label,
        colour,
        image_url,
        description,
        text_colour,
        location_select_type
    )
VALUES
    (
        1,
        'Nuke',
        '#E86621',
        'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-nuke.jpg',
        '#FFFFFF',
        'The show-stopper. A tactical nuke at your fingertips.',
        'LOCATION_SELECT'
    ),
    (
        0,
        'Airstrike',
        '#173DD1',
        'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-airstrike.jpg',
        '#FFFFFF',
        'Rain fury on the arena with a targeted airstrike.',
        'LOCATION_SELECT'
    );

-- Update rarities of Nuke, Airstrike and Landmine player abilities
UPDATE
    blueprint_player_abilities
SET
    rarity_weight = 30
WHERE
    game_client_ability_id IN (0, 1, 11);

UPDATE
    consumed_abilities
SET
    rarity_weight = 30
WHERE
    game_client_ability_id IN (0, 1, 11);