UPDATE game_abilities
SET game_client_ability_id = 0
WHERE label = 'AIRSTRIKE';

UPDATE game_abilities
SET game_client_ability_id = 1
WHERE label = 'NUKE';

UPDATE game_abilities
SET game_client_ability_id = 2
WHERE label = 'REPAIR';

UPDATE game_abilities
SET game_client_ability_id = 3
WHERE label = 'ROBOT DOGS';

UPDATE game_abilities
SET game_client_ability_id = 4
WHERE label = 'REINFORCEMENTS';

UPDATE game_abilities
SET game_client_ability_id = 5
WHERE label = 'OVERCHARGE';

UPDATE game_abilities
SET game_client_ability_id = 7
WHERE label = 'FIREWORKS';
