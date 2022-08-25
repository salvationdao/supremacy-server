ALTER TYPE ABILITY_TYPE_ENUM ADD VALUE 'FIREWORKS';

UPDATE punish_options SET key = 'restrict_location_select', description = 'Restrict player to select location for 24 hours' WHERE key = 'limit_location_select';
UPDATE punish_options SET key = 'restrict_chat', description = 'Restrict player to chat for 24 hours' WHERE key = 'limit_chat';
UPDATE punish_options SET key = 'restrict_sups_contribution', description = 'Restrict player to contribute sups for 24 hours' WHERE key = 'limit_sups_contibution';