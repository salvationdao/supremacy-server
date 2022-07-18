UPDATE
    sale_player_abilities
SET
    rarity_weight = 1
WHERE
    blueprint_id = (
        select
            id
        from
            blueprint_player_abilities
        where
            game_client_ability_id = 1
    );

UPDATE
    sale_player_abilities
SET
    rarity_weight = 2
WHERE
    blueprint_id IN (
        select
            id
        from
            blueprint_player_abilities
        where
            game_client_ability_id in (0, 11)
    );

UPDATE
    sale_player_abilities
SET
    rarity_weight = 6
WHERE
    blueprint_id IN (
        select
            id
        from
            blueprint_player_abilities
        where
            game_client_ability_id not in (0, 1, 11)
    );

UPDATE
    sale_player_abilities
SET
    rarity_weight = -1
WHERE
    blueprint_id IN (
        select
            id
        from
            blueprint_player_abilities
        where
            game_client_ability_id in (11)
    );