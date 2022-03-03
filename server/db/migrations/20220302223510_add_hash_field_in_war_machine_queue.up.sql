begin;
DELETE FROM battle_war_machine_queues WHERE deleted_at IS NOT NULL;
ALTER TABLE battle_war_machine_queues ADD COLUMN war_machine_hash VARCHAR(20);
ALTER TABLE battle_war_machine_queues ADD COLUMN faction_id UUID;
UPDATE battle_war_machine_queues bwmq SET faction_id = uuid(bwmq.war_machine_metadata ->> 'factionID');
ALTER TABLE battle_war_machine_queues ADD COLUMN created_at TIMESTAMPTZ default now();
ALTER TABLE battle_war_machine_queues ADD CONSTRAINT unique_mech_hash_in_queue UNIQUE (war_machine_hash);
commit;