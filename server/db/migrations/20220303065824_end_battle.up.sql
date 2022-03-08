CREATE table issued_contract_rewards (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    battle_id UUID NOT NULL REFERENCES battles(id),
    reward NUMERIC(28) NOT NULL,
    war_machine_hash TEXT NOT NULL,
    
    deleted_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ NOT NULL             DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL             DEFAULT NOW()
);