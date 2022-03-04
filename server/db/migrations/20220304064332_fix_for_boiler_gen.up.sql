-- adding primary key id column to allow sqlboiler to generate

ALTER TABLE public.battle_events_war_machine_destroyed_assisted_war_machines ADD id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE public.battle_events_war_machine_destroyed_assisted_war_machines ADD CONSTRAINT battle_events_war_machine_destroyed_assisted_war_machines_pk PRIMARY KEY (id);

ALTER TABLE public.battle_war_machine_queues ADD id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE public.battle_war_machine_queues ADD CONSTRAINT battle_war_machine_queues_pk PRIMARY KEY (id);

ALTER TABLE public.asset_repair ADD id uuid NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE public.asset_repair ADD CONSTRAINT asset_repair_pk PRIMARY KEY (id);
