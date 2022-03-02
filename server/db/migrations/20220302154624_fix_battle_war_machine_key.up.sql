create table battles_winner_records(
    battle_id           UUID NOT NULL,
    war_machine_hash    TEXT NOT NULL,     
    is_winner           BOOL  NOT NULL DEFAULT FALSE,
    faction_id          UUID not null,
    owner_id            UUID NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (battle_id, war_machine_hash)
);



-- drop faction stats
DROP MATERIALIZED VIEW faction_stats;

-- drop faction stats
DROP MATERIALIZED VIEW user_stats;

-- create faction materialize view
CREATE MATERIALIZED VIEW faction_stats AS
SELECT *
FROM (
         SELECT f.id
         FROM factions f
     ) f1
         LEFT JOIN LATERAL (
    SELECT COUNT(DISTINCT bwr.battle_id) AS win_count
    FROM battles_winner_records bwr
    WHERE bwr.is_winner = TRUE
      AND bwr.faction_id = f1.id
    GROUP BY bwr.faction_id
    ) f2 ON TRUE
         LEFT JOIN LATERAL (
    SELECT ((SELECT COUNT(b.id) FROM battles b) - COUNT(DISTINCT battle_id)) AS loss_count
    FROM battles_winner_records bwr
    WHERE bwr.is_winner = TRUE
      AND bwr.faction_id = f1.id
    GROUP BY bwr.faction_id
    ) f3 ON TRUE
         LEFT JOIN LATERAL (
    SELECT COUNT(bewmd.id) AS kill_count
    FROM battle_events_war_machine_destroyed bewmd
             INNER JOIN battle_events be ON be.id = bewmd.event_id
             INNER JOIN battles_winner_records bwr ON be.battle_id = bwr.battle_id AND
                                                    bewmd.kill_by_war_machine_hash =
                                                    bwr.war_machine_hash AND
                                                    bwr.faction_id = f1.id
    GROUP BY bwr.faction_id
    ) f4 ON TRUE
         LEFT JOIN LATERAL (
    SELECT COUNT(bewmd.id) AS death_count
    FROM battle_events_war_machine_destroyed bewmd
             INNER JOIN battle_events be ON be.id = bewmd.event_id
             INNER JOIN battles_winner_records bwr ON be.battle_id = bwr.battle_id AND
                                                    bewmd.destroyed_war_machine_hash =
                                                    bwr.war_machine_hash AND
                                                    bwr.faction_id = f1.id
    GROUP BY bwr.faction_id
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
                  FROM battles_winner_records bwr
                  WHERE bwr.battle_id = be.battle_id
                    AND bwr.war_machine_hash = bewmd.kill_by_war_machine_hash
                    AND bwr.owner_id = u1.id 
              )
    ) u4 ON TRUE;

CREATE UNIQUE INDEX user_id ON user_stats (id);

-- drop table
drop table battles_war_machines;