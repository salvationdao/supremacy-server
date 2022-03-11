CREATE TABLE battles_user_views
(
    battle_id  UUID NOT NULL REFERENCES battles (id),
    player_id    UUID NOT NULL REFERENCES users (id),
    PRIMARY KEY (battle_id, player_id)
);

-- drop user stats
DROP MATERIALIZED VIEW user_stats;

CREATE MATERIALIZED VIEW user_stats AS
select * 
from (
         SELECT p.id
         FROM players p
     ) p1
         LEFT JOIN LATERAL (
   		select count(*) as view_battle_count from battles_user_views buv
   		where buv.player_id = p1.id
  		group by buv.player_id 
    ) p2 ON true 
    	left join lateral(
    	select count(bat.id ) as total_ability_triggered from battle_ability_triggers bat 
    	where bat.player_id = p1.id
    	group by bat.player_id 
    )p3 on true
    	left join lateral(
    	select count(bh.id) as kill_count from battle_history bh 
    	inner join battle_mechs bm on bm.mech_id = bh.war_machine_one_id and bm.owner_id = p1.id
    	group by bm.owner_id 
    )p4 on true
      left join lateral (
    select sum(bc.amount) as total_vote_count from battle_contributions bc where bc.player_id = p1.id
    group by bc.player_id 
    )p5 on true;

CREATE UNIQUE INDEX user_id ON user_stats (id);

