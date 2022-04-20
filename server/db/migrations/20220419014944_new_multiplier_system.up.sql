ALTER TABLE battle_contributions ADD multi_amount DECIMAL NOT NULL default 0;

BEGIN;
ALTER TYPE multiplier_type_enum ADD VALUE 'contribute';
COMMIT;

UPDATE multipliers SET description = 'For a player who contributed towards a won battle. Multipliers adds up based off current contributor amount.' WHERE key = 'contributor';
UPDATE multipliers SET description = 'For a player who contributed towards the battle.' WHERE key = 'citizen';

UPDATE multipliers SET description = 'For a player who contributed towards a successful repair.', multiplier_type = 'contribute', value = 50 WHERE key = 'grease monkey';
UPDATE multipliers SET description = 'For a player who successfully repairs their syndicate mech.', value = 100 WHERE key = 'field mechanic';
UPDATE multipliers SET description = 'For a player who contributed towards a positive nuke.', value = 50, multiplier_type = 'contribute' WHERE key = 'now i am become death';
UPDATE multipliers SET description = 'For a player who got positive kills from a nuke. Multiplier can stack up to 3x with multi kills.', value = 100 WHERE key = 'destroyer of worlds';
UPDATE multipliers SET description = 'For a player who contributed towards a positive airstrike.', value = 50, multiplier_type = 'contribute' WHERE key = 'air support';
UPDATE multipliers SET description = 'For a player who got positive kills from a nuke. Multiplier can stack up to 3x with multi kills.', value = 100 WHERE key = 'air marshal';


ALTER TYPE battle_event ADD VALUE 'pickup';
ALTER TABLE user_multipliers DROP CONSTRAINT IF EXISTS user_multipliers_pkey;
ALTER TABLE user_multipliers ADD id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid();
