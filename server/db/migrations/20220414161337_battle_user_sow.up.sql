CREATE TABLE user_spoils_of_war
(
    id                          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    player_id                   UUID        NOT NULL REFERENCES players (id),
    battle_id                   UUID        NOT NULL REFERENCES battles (id),
    total_multiplier_for_battle INT         NOT NULL,
    total_sow                   NUMERIC(28) NOT NULL,
    paid_sow                    NUMERIC(28) NOT NULL,
    tick_amount                 NUMERIC(28) NOT NULL,
    lost_sow                    NUMERIC(28) NOT NULL,
    related_transaction_ids     TEXT[]               DEFAULT '{}',
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at                  TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE spoils_of_war
    ADD COLUMN current_tick INT NOT NULL DEFAULT 0;
ALTER TABLE spoils_of_war
    ADD COLUMN max_ticks INT NOT NULL DEFAULT 20;
