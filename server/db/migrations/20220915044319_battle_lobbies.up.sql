-- rename battle columns that are no longer used
ALTER TABLE battles
    RENAME COLUMN started_battle_seconds TO started_battle_seconds_old;
ALTER TABLE battles
    RENAME COLUMN ended_battle_seconds TO ended_battle_seconds_old;

ALTER TABLE repair_cases
    ADD COLUMN IF NOT EXISTS paused_at TIMESTAMPTZ;

ALTER TABLE battle_queue
    RENAME TO battle_queue_old;

ALTER TABLE battle_queue_fees
    RENAME TO battle_queue_fees_old;

ALTER TABLE battle_war_machine_queues
    RENAME TO battle_war_machine_queues_old;

ALTER TABLE battle_map_queue
    RENAME TO battle_map_queue_old;

CREATE INDEX IF NOT EXISTS idx_player_kill_log_offering_id ON player_kill_log (ability_offering_id);

CREATE TABLE battle_lobbies
(
    id                       UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    host_by_id               UUID        NOT NULL REFERENCES players (id),
    number                   SERIAL,
    entry_fee                NUMERIC(28) NOT NULL DEFAULT 0,
    first_faction_cut        DECIMAL     NOT NULL DEFAULT 0,
    second_faction_cut       DECIMAL     NOT NULL DEFAULT 0,
    third_faction_cut        DECIMAL     NOT NULL DEFAULT 0,
    each_faction_mech_amount INT         NOT NULL DEFAULT 3,
    game_map_id              UUID        NOT NULL REFERENCES game_maps (id),
    generated_by_system      BOOL        NOT NULL DEFAULT FALSE,
    password                 TEXT,

    -- battle queue
    ready_at                 TIMESTAMPTZ,                  -- order of the battle lobby get in battle arena
    assigned_to_battle_id    UUID REFERENCES battles (id), -- set battle id, if in battle
    ended_at                 TIMESTAMPTZ,                  -- set when battle is completed

    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at               TIMESTAMPTZ
);

CREATE INDEX idx_battle_lobby_complete_check ON battle_lobbies (ended_at, deleted_at);
CREATE INDEX idx_battle_lobby_queue_available_check ON battle_lobbies (ready_at, deleted_at);
CREATE INDEX idx_battle_lobby_queue_position_check ON battle_lobbies (ready_at, assigned_to_battle_id, ended_at, deleted_at);


CREATE TABLE battle_lobbies_mechs
(
    battle_lobby_id       UUID        NOT NULL REFERENCES battle_lobbies (id),
    mech_id               UUID        NOT NULL REFERENCES mechs (id),
    PRIMARY KEY (battle_lobby_id, mech_id),

    paid_tx_id            TEXT,
    refund_tx_id          TEXT,
    owner_id              UUID        NOT NULL REFERENCES players (id),
    faction_id            UUID        NOT NULL REFERENCES factions (id),

-- notification fee
    is_notified           BOOL        NOT NULL DEFAULT FALSE,
    notified_tx_id        TEXT,

-- payout tx id
    bonus_sups_tx_id      TEXT,
    payout_tx_id          TEXT,
    tax_tx_id             TEXT,
    challenge_fund_tx_id  TEXT,

-- battle related column (reduce the layer of inner join)
    locked_at             TIMESTAMPTZ, -- set when lobby is ready
    ended_at              TIMESTAMPTZ, -- set when battle is ended
    assigned_to_battle_id UUID REFERENCES battles (id),


    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX idx_battle_lobbies_mechs_queue_battle_check ON battle_lobbies_mechs (mech_id, ended_at, assigned_to_battle_id, refund_tx_id, deleted_at);
CREATE INDEX idx_battle_lobbies_mechs_queue_check ON battle_lobbies_mechs (mech_id, ended_at, refund_tx_id, deleted_at);
CREATE INDEX idx_battle_lobbies_mechs_lobby_queue_check ON battle_lobbies_mechs (battle_lobby_id, refund_tx_id, deleted_at);

-- only able to set bounties when the lobby is marked as READY
CREATE TABLE battle_bounties
(
    id               UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    battle_lobby_id  UUID        NOT NULL REFERENCES battle_lobbies (id),
    offered_by_id    UUID        NOT NULL REFERENCES players (id),
    targeted_mech_id UUID        NOT NULL REFERENCES mechs (id),

    amount           NUMERIC(28) NOT NULL DEFAULT 0,
    paid_tx_id       TEXT,
    payout_tx_id     TEXT,
    refund_tx_id     TEXT,
    tax_tx_id TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_battle_bounties_available_check ON battle_bounties (battle_lobby_id, payout_tx_id, refund_tx_id, deleted_at);

-- refactor repair block check
CREATE OR REPLACE FUNCTION check_repair_block() RETURNS TRIGGER AS
$check_repair_block$
DECLARE
    can_write_block BOOLEAN DEFAULT FALSE;
BEGIN

SELECT (
           SELECT rc.completed_at IS NULL AND rc.paused_at ISNULL AND
                  ro.expires_at > NOW() AND ro.closed_at IS NULL AND ro.deleted_at IS NULL AND
                  (SELECT COUNT(*) FROM repair_blocks rb WHERE rb.repair_case_id = rc.id) < rc.blocks_required_repair
           FROM repair_offers ro
           INNER JOIN repair_cases rc ON ro.repair_case_id = rc.id
           WHERE ro.id = NEW.repair_offer_id
       )
INTO can_write_block;
-- update blocks required in repair cases and continue the process
IF can_write_block THEN
    UPDATE repair_cases SET blocks_repaired = blocks_repaired + 1 WHERE id = NEW.repair_case_id;
    UPDATE repair_agents SET finished_at = now(), finished_reason = 'SUCCEEDED' WHERE id = NEW.repair_agent_id;
    RETURN NEW;
ELSE
    RAISE EXCEPTION 'unable to write block';
END IF;
END
$check_repair_block$
    LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_check_repair_block ON repair_blocks;

CREATE TRIGGER trigger_check_repair_block
    BEFORE INSERT
    ON repair_blocks
    FOR EACH ROW
EXECUTE PROCEDURE check_repair_block();
