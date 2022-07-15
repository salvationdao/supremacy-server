ALTER TABLE game_abilities
    ADD COLUMN deleted_at TIMESTAMPTZ;


UPDATE game_abilities set deleted_at = NOW() where label in ('OVERCHARGE', 'ROBOT DOGS', 'REINFORCEMENTS');
