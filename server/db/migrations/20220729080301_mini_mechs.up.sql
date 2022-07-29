INSERT INTO
    blueprint_player_abilities (
        game_client_ability_id,
        label,
        colour,
        image_url,
        description,
        text_colour,
        location_select_type,
        rarity_weight,
        inventory_limit
    )
VALUES
    (
        18,
        'Mini Mech',
        '#4f5f61',
        'https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-mini-mech.png',
        'Spawn a tiny mech to fight alongside your allies.',
        '#FFFFFF',
        'LOCATION_SELECT',
        10,
        10
    );

INSERT INTO
    sale_player_abilities (blueprint_id, current_price, rarity_weight)
VALUES
    (
        (
            SELECT
                id
            FROM
                blueprint_player_abilities
            WHERE
                game_client_ability_id = 18
        ),
        100000000000000000000,
        2
    );