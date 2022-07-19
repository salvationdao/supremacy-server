UPDATE weapons
SET damage = 12
WHERE label ILIKE 'Auto Cannon' OR label ILIKE 'Red Mountain Offworld Mining Corporation Auto Cannon';

UPDATE blueprint_weapons
SET damage = 12
WHERE label ILIKE 'Auto Cannon' OR label ILIKE 'Red Mountain Offworld Mining Corporation Auto Cannon';