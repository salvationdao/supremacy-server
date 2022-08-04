ALTER TABLE mech_ability_trigger_logs
    ADD COLUMN IF NOT EXISTS battle_number integer not null default 0; -- track which battle the ability is triggered