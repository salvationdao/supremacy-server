ALTER TABLE battle_abilities
    ADD COLUMN IF NOT EXISTS maximum_commander_count int not null default 1;

UPDATE
    battle_abilities
SET
    maximum_commander_count = 2
WHERE
    label != 'NUKE';