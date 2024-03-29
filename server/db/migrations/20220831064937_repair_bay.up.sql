DROP TYPE IF EXISTS REPAIR_BAY_STATUS;
DROP TYPE IF EXISTS REPAIR_SLOT_STATUS;
CREATE TYPE REPAIR_SLOT_STATUS AS ENUM ('REPAIRING','PENDING','DONE');

CREATE TABLE player_mech_repair_slots
(
    id               UUID PRIMARY KEY            DEFAULT gen_random_uuid(),
    player_id        UUID               NOT NULL REFERENCES players (id),
    mech_id          UUID               NOT NULL REFERENCES mechs (id),
    repair_case_id   UUID               NOT NULL REFERENCES repair_cases (id),
    status           REPAIR_SLOT_STATUS NOT NULL DEFAULT 'PENDING',
    next_repair_time TIMESTAMPTZ,
    slot_number      int                not null default 0, -- mean done
    created_at       TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_player_mech_repair_slot_repair_search ON player_mech_repair_slots (status, next_repair_time, deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_mech_repair_slot_player_search ON player_mech_repair_slots (player_id, status, deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_mech_repair_slot_mech_status_search ON player_mech_repair_slots (mech_id, status, deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_mech_repair_slot_slot_number ON player_mech_repair_slots (slot_number);

-- insert repair center user
INSERT INTO players (id, username, is_ai )
VALUES ('a988b1e3-5556-4cad-83bd-d61c2b149cb7', 'Repair Centre', true);

