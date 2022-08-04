-- disable airstrike/nuke/landmine from sale player ability
UPDATE
    sale_player_abilities spa
SET
    deleted_at = now()
where
    exists(
        SELECT 1 FROM blueprint_player_abilities bpa
        WHERE bpa.id = spa.blueprint_id AND (
            bpa.game_client_ability_id = 0 || bpa.game_client_ability_id = 1 || bpa.game_client_ability_id = 11
        )
);

-- set default maximum commander count to 2
UPDATE
    battle_abilities
SET
    maximum_commander_count = 2;

-- set maximum commander count of NUKE to 1
UPDATE
    battle_abilities
SET
    maximum_commander_count = 1
WHERE
    label = 'NUKE';

-- set maximum commander count of LANDMINE to 1
UPDATE
    battle_abilities
SET
    maximum_commander_count = 3
WHERE
    label = 'LANDMINE';