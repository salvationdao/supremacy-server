ALTER TABLE
    blueprint_player_abilities
ADD
    COLUMN rarity_weight INT;

ALTER TABLE
    consumed_abilities
ADD
    COLUMN rarity_weight INT;