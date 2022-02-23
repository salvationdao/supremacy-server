DROP TABLE blobs;

UPDATE game_abilities
SET image_url = 'https://i.pinimg.com/originals/b1/92/4d/b1924dce177345b5485bb5490ab3441f.jpg'
WHERE label = 'AIRSTRIKE';

UPDATE game_abilities
SET image_url = 'https://images2.minutemediacdn.com/image/upload/c_crop,h_1126,w_2000,x_0,y_83/f_auto,q_auto,w_1100/v1555949079/shape/mentalfloss/581049-mesut_zengin-istock-1138195821.jpg'
WHERE label = 'NUKE';

UPDATE game_abilities
SET image_url = 'https://images2.minutemediacdn.com/image/upload/c_crop,h_1126,w_2000,x_0,y_83/f_auto,q_auto,w_1100/v1555949079/shape/mentalfloss/581049-mesut_zengin-istock-1138195821.jpg'
WHERE label = 'REINFORCEMENTS';

UPDATE game_abilities
SET image_url = 'https://i.pinimg.com/originals/ed/2f/9b/ed2f9b6e66b9efefa84d1ee423c718f0.png'
WHERE label = 'REPAIR';

UPDATE game_abilities
SET image_url = 'https://i.pinimg.com/originals/b1/92/4d/b1924dce177345b5485bb5490ab3441f.jpg'
WHERE label = 'ROBOT DOGS';

ALTER TABLE game_abilities
	DROP COLUMN description;
