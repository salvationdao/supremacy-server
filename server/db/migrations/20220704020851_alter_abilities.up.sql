-- remove REPAIR from battle abilities
UPDATE
    game_abilities
SET
    battle_ability_id = null,
    level = 'MECH'
WHERE
    label = 'REPAIR';

delete from battle_abilities ba where ba.label = 'REPAIR';

INSERT INTO battle_abilities (id, label, cooldown_duration_second, description)
VALUES ('87ea13a8-9065-40be-b51d-aee9dd57c23f', 'LANDMINE', 20, 'Deploy a line of explosives that detonate when a War Machine is detected within its proximity.');

ALTER TABLE game_abilities
    ADD COLUMN IF NOT EXISTS location_select_type location_select_type_enum not null default 'LOCATION_SELECT';

INSERT INTO game_abilities (id, game_client_ability_id, faction_id, battle_ability_id, label, colour, image_url, sups_cost, description, text_colour, level, location_select_type)
VALUES ('2e787662-0ce3-47c7-a9ad-45e6ed20beee', 11, '98bf7bb3-1a7c-4f21-8843-458d62884060', '87ea13a8-9065-40be-b51d-aee9dd57c23f','LANDMINE','#d9674c','https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-landmine.png','100000000000000000000','Deploy a line of explosives that detonate when a War Machine is detected within its proximity.', '#d9674c','FACTION','LINE_SELECT'),
       ('e3602296-9f91-4823-b83f-f30ac79505b7', 11, '7c6dde21-b067-46cf-9e56-155c88a520e2', '87ea13a8-9065-40be-b51d-aee9dd57c23f','LANDMINE','#d9674c','https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-landmine.png','100000000000000000000','Deploy a line of explosives that detonate when a War Machine is detected within its proximity.', '#d9674c','FACTION','LINE_SELECT'),
       ('cd61d32b-8e0b-4e55-970d-944867d7b524', 11, '880db344-e405-428d-84e5-6ebebab1fe6d', '87ea13a8-9065-40be-b51d-aee9dd57c23f','LANDMINE','#d9674c','https://afiles.ninja-cdn.com/supremacy-stream-site/assets/img/ability-landmine.png','100000000000000000000','Deploy a line of explosives that detonate when a War Machine is detected within its proximity.', '#d9674c','FACTION','LINE_SELECT');

ALTER TYPE ABILITY_TYPE_ENUM ADD VALUE 'LANDMINE';