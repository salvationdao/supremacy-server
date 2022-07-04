-- Update player ability tables to include the rarity weight of the ability
ALTER TABLE
    blueprint_player_abilities
ADD
    COLUMN rarity_weight INT NOT NULL DEFAULT -1;

ALTER TABLE
    consumed_abilities
ADD
    COLUMN rarity_weight INT NOT NULL DEFAULT -1;

ALTER TABLE
    sale_player_abilities
ADD
    COLUMN rarity_weight INT NOT NULL DEFAULT -1;

-- New trigger, t_sale_player_abilities_insert for automatically setting the rarity weight of newly-created
-- sale_player_abilities entries based on the associated blueprint_player_ability
CREATE
OR REPLACE FUNCTION set_rarity_weight() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN NEW.rarity_weight := (
    SELECT
        rarity_weight
    from
        blueprint_player_abilities
    where
        id = NEW.blueprint_id
);

RETURN NEW;

END $$;

CREATE TRIGGER "t_sale_player_abilities_insert" BEFORE
INSERT
    ON "sale_player_abilities" FOR EACH ROW EXECUTE PROCEDURE set_rarity_weight();

-- Update rarities of all player abilities (except for landmines, 11; is a rarer ability)
UPDATE
    blueprint_player_abilities
SET
    rarity_weight = 30
WHERE
    game_client_ability_id IN (10, 12, 13, 14, 15, 16);

UPDATE
    consumed_abilities ca
SET
    rarity_weight = 30
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

-- Update rarities of Airstrike and Landmine player abilities
UPDATE
    blueprint_player_abilities
SET
    rarity_weight = 10
WHERE
    game_client_ability_id IN (0, 11);

UPDATE
    consumed_abilities
SET
    rarity_weight = 30
WHERE
    game_client_ability_id IN (0, 11);

-- Update rarity of Nuke ability
UPDATE
    blueprint_player_abilities
SET
    rarity_weight = 5
WHERE
    game_client_ability_id IN (1);

UPDATE
    consumed_abilities
SET
    rarity_weight = 5
WHERE
    game_client_ability_id IN (1);

-- Update all sale abilities rarity weights
UPDATE
    sale_player_abilities spa
SET
    rarity_weight = (
        SELECT
            rarity_weight
        FROM
            blueprint_player_abilities
        WHERE
            id = spa.blueprint_id
    );