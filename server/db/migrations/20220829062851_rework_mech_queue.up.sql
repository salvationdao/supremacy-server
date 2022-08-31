CREATE TABLE battle_queue_backlog (
	mech_id uuid NOT NULL,
	faction_id uuid NOT NULL,
	owner_id uuid NOT NULL,
	fee_id uuid NULL,
	queue_fee_tx_id text NULL,
	queue_fee_tx_id_refund text NULL,
	updated_at timestamptz NOT NULL DEFAULT now(),
    queued_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT battle_queue_backlog_pkey PRIMARY KEY (mech_id),
	CONSTRAINT battle_queue_backlog_chassis_id_fkey FOREIGN KEY (mech_id) REFERENCES public.mechs(id),
	CONSTRAINT battle_queue_backlog_faction_id_fkey FOREIGN KEY (faction_id) REFERENCES public.factions(id),
	CONSTRAINT battle_queue_backlog_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES public.players(id),
	CONSTRAINT battle_queue_backlog_fee_id_fkey FOREIGN KEY (fee_id) REFERENCES public.battle_queue_fees(id)
);
