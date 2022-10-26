ALTER TABLE battle_lobbies
    ADD COLUMN auto_fill_at TIMESTAMPTZ,
    ADD COLUMN expired_at TIMESTAMPTZ;