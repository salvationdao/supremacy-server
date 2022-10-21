ALTER TABLE battle_lobbies_mechs
    RENAME COLUMN queued_by_id TO owner_id ;

ALTER TABLE battle_mechs
    RENAME COLUMN piloted_by_id TO owner_id;