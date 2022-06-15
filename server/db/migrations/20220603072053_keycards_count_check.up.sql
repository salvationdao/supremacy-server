ALTER TABLE player_keycards
    DROP CONSTRAINT IF EXISTS amount_check,
    ADD CONSTRAINT amount_check CHECK ( count >= 0 );