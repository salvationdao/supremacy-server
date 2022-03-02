ALTER TABLE game_abilities
	ADD COLUMN text_colour TEXT;

UPDATE game_abilities SET text_colour = '#173DD1' WHERE label = 'AIRSTRIKE';
UPDATE game_abilities SET text_colour = '#E86621' WHERE label = 'NUKE';
UPDATE game_abilities SET text_colour = '#23AE3C' WHERE label = 'REPAIR';
UPDATE game_abilities SET text_colour = '#428EC1' WHERE label = 'ROBOT DOGS';
UPDATE game_abilities SET text_colour = '#C52A1F' WHERE label = 'REINFORCEMENTS';
UPDATE game_abilities SET text_colour = '#FFFFFF' WHERE label = 'OVERCHARGE';

ALTER TABLE game_abilities
	ALTER COLUMN text_colour SET NOT NULL;
