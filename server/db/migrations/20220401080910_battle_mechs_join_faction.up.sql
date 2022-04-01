ALTER TABLE battle_mechs ADD COLUMN faction_won BOOLEAN DEFAULT false;
ALTER TABLE battle_mechs ADD COLUMN mech_survived BOOLEAN DEFAULT false;

UPDATE battle_mechs bm SET faction_won = (EXISTS (SELECT faction_id FROM battle_wins bw WHERE bw.faction_id = bm.faction_id AND bw.battle_id = bm.battle_id));
UPDATE battle_mechs bm SET mech_survived = (EXISTS (SELECT mech_id FROM battle_wins bw WHERE bw.mech_id = bm.mech_id AND bw.battle_id = bm.battle_id));