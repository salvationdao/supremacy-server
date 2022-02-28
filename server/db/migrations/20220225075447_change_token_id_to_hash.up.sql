-- drop faction stats
DROP MATERIALIZED VIEW faction_stats;

-- drop faction stats
DROP MATERIALIZED VIEW user_stats;

ALTER TABLE battle_events_war_machine_destroyed
    DROP COLUMN IF EXISTS destroyed_war_machine_id,
    ADD COLUMN IF NOT EXISTS destroyed_war_machine_hash TEXT NOT NULL DEFAULT '',
    DROP COLUMN IF EXISTS kill_by_war_machine_id,
    ADD COLUMN IF NOT EXISTS kill_by_war_machine_hash TEXT;

ALTER TABLE battle_events_war_machine_destroyed_assisted_war_machines
    DROP COLUMN IF EXISTS war_machine_id,
    ADD COLUMN IF NOT EXISTS war_machine_hash TEXT NOT NULL;

ALTER TABLE battle_events_game_ability
    DROP COLUMN IF EXISTS ability_token_id,
    ADD COLUMN IF NOT EXISTS ability_hash TEXT;

-- create faction materialize view
CREATE MATERIALIZED VIEW faction_stats AS
SELECT *
FROM (
         SELECT f.id
         FROM factions f
     ) f1
         LEFT JOIN LATERAL (
    SELECT COUNT(DISTINCT bwm.battle_id) AS win_count
    FROM battles_war_machines bwm
    WHERE bwm.is_winner = TRUE
      AND bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::TEXT
    GROUP BY bwm.war_machine_stat -> 'faction' ->> 'id'
    ) f2 ON TRUE
         LEFT JOIN LATERAL (
    SELECT ((SELECT COUNT(b.id) FROM battles b) - COUNT(DISTINCT battle_id)) AS loss_count
    FROM battles_war_machines bwm
    WHERE bwm.is_winner = TRUE
      AND bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::TEXT
    GROUP BY bwm.war_machine_stat -> 'faction' ->> 'id'
    ) f3 ON TRUE
         LEFT JOIN LATERAL (
    SELECT COUNT(bewmd.id) AS kill_count
    FROM battle_events_war_machine_destroyed bewmd
             INNER JOIN battle_events be ON be.id = bewmd.event_id
             INNER JOIN battles_war_machines bwm ON be.battle_id = bwm.battle_id AND
                                                    bewmd.kill_by_war_machine_hash::TEXT =
                                                    bwm.war_machine_stat ->> 'hash' AND
                                                    bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::TEXT
    GROUP BY bwm.war_machine_stat -> 'faction' ->> 'id'
    ) f4 ON TRUE
         LEFT JOIN LATERAL (
    SELECT COUNT(bewmd.id) AS death_count
    FROM battle_events_war_machine_destroyed bewmd
             INNER JOIN battle_events be ON be.id = bewmd.event_id
             INNER JOIN battles_war_machines bwm ON be.battle_id = bwm.battle_id AND
                                                    bewmd.destroyed_war_machine_hash::TEXT =
                                                    bwm.war_machine_stat ->> 'hash' AND
                                                    bwm.war_machine_stat -> 'faction' ->> 'id' = f1.id::TEXT
    GROUP BY bwm.war_machine_stat -> 'faction' ->> 'id'
    ) f5 ON TRUE;

-- create unique index
CREATE UNIQUE INDEX faction_id ON faction_stats (id);


CREATE MATERIALIZED VIEW user_stats AS
SELECT *
FROM (
         SELECT u.id, u.view_battle_count
         FROM users u
     ) u1
         LEFT JOIN LATERAL (
    SELECT SUM(buv.vote_count) AS total_vote_count
    FROM battles_user_votes buv
    WHERE buv.user_id = u1.id
    GROUP BY buv.user_id
    ) u2 ON TRUE
         LEFT JOIN LATERAL (
    SELECT COUNT(bega.id) AS total_ability_triggered
    FROM battle_events_game_ability bega
    WHERE bega.triggered_by_user_id = u1.id
    ) u3 ON TRUE
         LEFT JOIN LATERAL (
    SELECT COUNT(bewmd.id) AS kill_count
    FROM battle_events_war_machine_destroyed bewmd
             INNER JOIN battle_events be ON be.id = bewmd.event_id
    WHERE EXISTS(
                  SELECT 1
                  FROM battles_war_machines bwm
                  WHERE bwm.battle_id = be.battle_id
                    AND bwm.war_machine_stat ->> 'hash' = bewmd.kill_by_war_machine_hash::TEXT
                    AND bwm.war_machine_stat ->> 'OwnedByID' = u1.id ::TEXT
              )
    ) u4 ON TRUE;

CREATE UNIQUE INDEX user_id ON user_stats (id);