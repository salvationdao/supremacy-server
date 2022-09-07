ALTER TABLE battle_replays
    ADD COLUMN intro_ended_at TIMESTAMPTZ,
    ADD COLUMN disabled_at TIMESTAMPTZ;