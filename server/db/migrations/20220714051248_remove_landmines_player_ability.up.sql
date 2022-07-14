ALTER TABLE
    sale_player_abilities
ADD
    COLUMN deleted_at timestamptz;

UPDATE
    sale_player_abilities
SET
    deleted_at = now()
WHERE
    blueprint_id = (
        SELECT
            id
        from
            blueprint_player_abilities
        WHERE
            game_client_ability_id = 11
    );