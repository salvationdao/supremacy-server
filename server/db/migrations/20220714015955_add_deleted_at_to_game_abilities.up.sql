ALTER TABLE game_abilities
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ; -- fix dev data sync


UPDATE game_abilities set deleted_at = NOW() where label in ('OVERCHARGE', 'ROBOT DOGS', 'REINFORCEMENTS');
