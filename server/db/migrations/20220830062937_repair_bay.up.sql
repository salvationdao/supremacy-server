DROP TYPE IF EXISTS REPAIR_BAY_STATUS;
CREATE TYPE REPAIR_BAY_STATUS AS ENUM ('REPAIRING','PENDING','DONE');

CREATE TABLE player_mech_repair_bays
(
    id               UUID PRIMARY KEY           DEFAULT gen_random_uuid(),
    player_id        UUID              NOT NULL REFERENCES players (id),
    mech_id          UUID              NOT NULL REFERENCES mechs (id),
    repair_case_id   UUID              NOT NULL REFERENCES repair_cases (id),
    status           REPAIR_BAY_STATUS NOT NULL DEFAULT 'PENDING',
    next_repair_time TIMESTAMPTZ,
    created_at       TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_player_mech_repair_bay_repair_search ON player_mech_repair_bays (status, next_repair_time, deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_mech_repair_bay_player_search ON player_mech_repair_bays (player_id, status, deleted_at);
CREATE INDEX IF NOT EXISTS idx_player_mech_repair_bay_mech_status_search ON player_mech_repair_bays (mech_id, status, next_repair_time, deleted_at);

