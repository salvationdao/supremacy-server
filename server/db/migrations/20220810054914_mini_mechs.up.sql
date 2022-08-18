ALTER TABLE
    mech_move_command_logs
ADD
    COLUMN is_moving bool NOT NULL DEFAULT false;

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

UPDATE
    blueprint_player_abilities
SET
    description = 'Deploy a drone onto the battlefield that hacks into the nearest enemy War Machine, overriding both targeting and movement systems and causing them to attack their allies when within range.'
WHERE
    game_client_ability_id = 13;
