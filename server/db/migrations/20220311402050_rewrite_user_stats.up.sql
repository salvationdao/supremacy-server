CREATE TABLE battles_viewers
(
    battle_id  UUID NOT NULL REFERENCES battles (id),
    player_id    UUID NOT NULL REFERENCES players (id),
    PRIMARY KEY (battle_id, player_id)
);

-- drop user stats
DROP MATERIALIZED VIEW user_stats;

CREATE MATERIALIZED VIEW user_stats AS
SELECT * 
FROM (
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
    )p4 ON true
       LEFT JOIN lateral (
       SELECT SUM(bc.amount) AS total_vote_count FROM battle_contributions bc WHERE bc.player_id = p1.id
       GROUP BY bc.player_id 
    )p5 ON true;

CREATE UNIQUE INDEX user_id ON user_stats (id);

