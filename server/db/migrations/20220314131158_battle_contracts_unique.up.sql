ALTER TABLE battle_contracts
    DROP CONSTRAINT IF EXISTS bc_unique_mech_battle;
ALTER TABLE battle_contracts ADD CONSTRAINT bc_unique_mech_battle UNIQUE (mech_id, battle_id);