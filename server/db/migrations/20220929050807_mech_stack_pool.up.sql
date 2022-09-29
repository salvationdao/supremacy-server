CREATE TABLE IF NOT EXISTS staked_mechs
(
    mech_id    UUID PRIMARY KEY REFERENCES mechs (id),
    owner_id   UUID        NOT NULL REFERENCES players (id),
    faction_id UUID        NOT NULL REFERENCES factions (id),
    staked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- store the record of the staked mechs joining in a battle
CREATE TABLE stacked_mech_battle_logs
(
    id             UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    battle_id      UUID        NOT NULL REFERENCES battles (id),
    staked_mech_id UUID        NOT NULL REFERENCES mechs (id),
    owner_id       UUID        NOT NULL REFERENCES players (id),
    faction_id     UUID        NOT NULL REFERENCES factions (id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_staked_mech_battle_log_created_at_desc ON stacked_mech_battle_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_staked_mech_battle_log_mech_search ON stacked_mech_battle_logs(staked_mech_id);