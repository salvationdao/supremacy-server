-- drop materialise view
drop materialized view player_last_seven_day_ability_kills;

-- create new last seven day kills table
create table player_kill_log(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    battle_id uuid not null references battles (id),
    player_id uuid not null references players (id),
    faction_id uuid not null references factions (id),
    is_team_kill bool not null default false,
    created_at TIMESTAMPTZ NOT NULL default NOW()
);

CREATE INDEX idx_last_seven_days_ability_kills ON player_kill_log(player_id, created_at DESC);

-- insert the table on the very first time

-- get the last seven days ability kill record
INSERT INTO player_kill_log (battle_id, player_id, faction_id, is_team_kill, created_at)
SELECT bh.battle_id, bat.player_id, bat.faction_id, bat.faction_id = bm.faction_id, bh.created_at from battle_history bh
    INNER JOIN battle_ability_triggers bat on bat.ability_offering_id = bh.related_id
    INNER JOIN battle_mechs bm on bm.mech_id = bh.war_machine_one_id AND bm.battle_id = bat.battle_id
where bh.related_id notnull and bh.event_type = 'killed';

delete from user_stats;
alter table user_stats RENAME COLUMN kill_count TO ability_kill_count;
INSERT INTO user_stats (id, view_battle_count, total_ability_triggered, ability_kill_count, mech_kill_count)
select p1.id, coalesce(p2.view_count,0), coalesce(p3.ability_triggered,0), (coalesce(p4.ability_kills,0) - coalesce(p5.team_kills,0)), coalesce(p6.mech_kills,0) from (
select p.id from players p
) p1 left join lateral(
    select count(bv.battle_id) as view_count from battles_viewers bv where bv.player_id = p1.id group by bv.player_id
) p2 on true left join lateral(
    select count(bat.id) as ability_triggered from battle_ability_triggers bat where bat.player_id = p1.id group by bat.player_id
) p3 on true left join lateral(
    select count(distinct bh.id) as ability_kills from battle_history bh
    inner join battle_ability_triggers bat on bat.ability_offering_id = bh.related_id and bat.player_id = p1.id
    inner join battle_mechs bm on bm.mech_id = bh.war_machine_one_id
    where bh.event_type = 'killed' and bh.related_id notnull and bh.war_machine_two_id isnull and bat.faction_id != bm.faction_id
    group by bh.event_type
) p4 on true left join lateral(
    select count( distinct bh.id) as team_kills from battle_history bh
    inner join battle_ability_triggers bat on bat.ability_offering_id = bh.related_id and bat.player_id = p1.id
    inner join battle_mechs bm on bm.mech_id = bh.war_machine_one_id
    where bh.event_type = 'killed' and bh.related_id notnull and bh.war_machine_two_id isnull and bat.faction_id = bm.faction_id
    group by bh.event_type
) p5 on true left join lateral(
    select count(bh.id) as mech_kills from battle_history bh
    where bh.event_type = 'killed' and bh.war_machine_two_id notnull and exists (select 1 from battle_mechs bm where bm.mech_id = bh.war_machine_two_id and bm.owner_id = p1.id)
) p6 on true;