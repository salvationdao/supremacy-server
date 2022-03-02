ALTER TABLE factions
    ADD COLUMN IF NOT EXISTS contract_reward TEXT NOT NULL DEFAULT '1000000000000000000';

ALTER TABLE battle_war_machine_queues
    ADD COLUMN IF NOT EXISTS is_insured BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS contract_reward TEXT NOT NULL DEFAULT '1000000000000000000';

ALTER TABLE battle_war_machine_queues
    RENAME COLUMN released_at TO deleted_at;


CREATE TABLE asset_repair(
    hash text not null,
    expect_completed_at TIMESTAMPTZ NOT NULL,
    repair_mode text NOT NULL,
    is_paid_to_complete bool NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);