CREATE TABLE battle_contracts (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    mech_id UUID NOT NULL references mechs (id),
    player_id UUID NOT NULL references players (id),
    faction_id UUID NOT NULL references factions (id),
    battle_id UUID NULL references battles (id),
    contract_reward NUMERIC(28) NOT NULL,
    fee NUMERIC(28) NOT NULL,
    did_win BOOL NULL,
    paid_out BOOL NOT NULL DEFAULT false,
    queued_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);