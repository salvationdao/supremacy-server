-- Update player ability tables to include the rarity weight of the ability


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



UPDATE
    consumed_abilities ca
SET
    rarity_weight = 30
WHERE
    game_client_ability_id IN (10, 12, 13, 14, 15, 16);



UPDATE
    consumed_abilities
SET
    rarity_weight = 10
WHERE
    game_client_ability_id IN (11);

-- Add nuke and airstrike as sale player abilities
INSERT INTO
    sale_player_abilities (
        blueprint_id,
        current_price,
        available_until
    )
VALUES
    (
        (SELECT id from blueprint_player_abilities WHERE game_client_ability_id = 1),
        100000000000000000000,
        now()
    ),
    (
        (SELECT id from blueprint_player_abilities WHERE game_client_ability_id = 0),
        100000000000000000000,
        now()
    );

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