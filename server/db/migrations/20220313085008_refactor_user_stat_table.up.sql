DROP MATERIALIZED VIEW user_stats;

CREATE TABLE user_stats(
    id uuid PRIMARY KEY NOT NULL REFERENCES players (id),
    view_battle_count int NOT NULL DEFAULT 0,
    kill_count int NOT NULL DEFAULT 0,
    total_ability_triggered int NOT NULL DEFAULT 0
);

-- initialise user stat
INSERT INTO 
    user_stats (id, view_battle_count, kill_count, total_ability_triggered)
SELECT 
    p1.id, COALESCE(p2.view_battle_count,0), COALESCE(p4.kill_count,0), COALESCE(p3.total_ability_triggered,0)
FROM(
        SELECT p.id
        FROM players p
    ) p1
        LEFT JOIN LATERAL (
    	SELECT COUNT(*) AS view_battle_count FROM battles_viewers buv
    	WHERE buv.player_id = p1.id
    	GROUP BY buv.player_id 
    ) p2 ON true 
    	LEFT JOIN lateral(
    	SELECT COUNT(bat.id ) AS total_ability_triggered FROM battle_ability_triggers bat 
    	WHERE bat.player_id = p1.id
    	GROUP by bat.player_id 
    )p3 ON true
    	LEFT JOIN lateral(
    	SELECT COUNT(bh.id) AS kill_count FROM battle_history bh 
    	INNER JOIN battle_mechs bm ON bm.mech_id = bh.war_machine_one_id AND bm.owner_id = p1.id
    	GROUP BY bm.owner_id 
    )p4 ON true;