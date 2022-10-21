ALTER TABLE battle_lobbies_mechs
    RENAME COLUMN owner_id TO queued_by_id;

ALTER TABLE battle_mechs
    RENAME COLUMN owner_id TO piloted_by_id;