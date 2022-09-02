-- drop repair bay indexes
DROP INDEX IF EXISTS idx_player_mech_repair_slot_repair_search;
DROP INDEX IF EXISTS idx_player_mech_repair_slot_player_search;
DROP INDEX IF EXISTS idx_player_mech_repair_slot_mech_status_search;
DROP INDEX IF EXISTS idx_player_mech_repair_slot_slot_number;

-- drop table
DROP TABLE IF EXISTS player_mech_repair_slots;

-- drop enum
DROP TYPE IF EXISTS REPAIR_SLOT_STATUS;
