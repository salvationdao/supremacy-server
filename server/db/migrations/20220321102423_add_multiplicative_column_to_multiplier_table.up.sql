ALTER TABLE multipliers
    ADD COLUMN IF NOT EXISTS is_multiplicative BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE 
    multipliers 
SET 
    is_multiplicative = true,
    value = 5
WHERE 
    key = 'won battle';
