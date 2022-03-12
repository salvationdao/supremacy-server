ALTER TABLE battle_queue
    ADD COLUMN battle_contract_id UUID NULL;

ALTER TABLE battle_contracts
    ADD COLUMN cancelled BOOL default false;