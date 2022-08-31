DROP TYPE IF EXISTS REPAIR_BAY_STATUS;
DROP TYPE IF EXISTS REPAIR_SLOT_STATUS;
CREATE TYPE REPAIR_SLOT_STATUS AS ENUM ('REPAIRING','PENDING','DONE');

CREATE TABLE player_mech_repair_slots
(
    id               UUID PRIMARY KEY           DEFAULT gen_random_uuid(),
    player_id        UUID              NOT NULL REFERENCES players (id),
    mech_id          UUID              NOT NULL REFERENCES mechs (id),
    repair_case_id   UUID              NOT NULL REFERENCES repair_cases (id),
    status           REPAIR_SLOT_STATUS NOT NULL DEFAULT 'PENDING',
    next_repair_time TIMESTAMPTZ,
    created_at       TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_player_mech_repair_slot_repair_search ON player_mech_repair_slots (status, next_repair_time, deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_mech_repair_slot_player_search ON player_mech_repair_slots (player_id, status, deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_mech_repair_slot_mech_status_search ON player_mech_repair_slots (mech_id, status, next_repair_time, deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_mech_repair_slot_created_at ON player_mech_repair_slots (created_at);
