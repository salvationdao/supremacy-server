DELETE FROM
    sale_player_abilities
WHERE
    blueprint_id = (
        SELECT
            id
        from
            blueprint_player_abilities
        WHERE
            game_client_ability_id = 11
    );