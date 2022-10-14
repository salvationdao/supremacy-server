UPDATE power_cores SET blueprint_id = '62e197a4-f45e-4034-ac0a-3e625a6770d7' WHERE blueprint_id IS NULL AND size_dont_use = 'SMALL';
UPDATE power_cores SET blueprint_id = '6921c16c-44f0-46af-b61b-f8106e40530f' WHERE blueprint_id IS NULL AND size_dont_use = 'MEDIUM';

ALTER TABLE power_cores
    ALTER COLUMN blueprint_id SET NOT NULL;
