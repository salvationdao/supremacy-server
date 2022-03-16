DROP MATERIALIZED VIEW faction_stats;

CREATE TABLE faction_stats(
    id uuid primary key not null references factions (id),
    win_count int not null default 0,
    loss_count int not null default 0,
    kill_count int not null default 0,
    death_count int not null default 0,
    sups_contribute NUMERIC(28) not null default 0,
    mvp_player_id uuid references players(id)
);

-- insert faction id
insert into faction_stats (id)
select f.id from factions f;

-----------------------
-- Red Mountain stat --
-----------------------

-- win_count 
update
	faction_stats fs2
set
win_count =
(select count(distinct (bw.battle_id)) as win_count  from battle_wins bw
where bw.faction_id = fs2.id)
where fs2.id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

-- loss_count
update 
	faction_stats fs2 
set
	loss_count = 
(select	(select count (b.id) as actaul_battle_count from battles b where exists (select 1 from battle_wins bw where bw.battle_id = b.id)) -  count(distinct (bw.battle_id)) as loss_count  
from battle_wins bw	where bw.faction_id = fs2.id)
where 
	fs2.id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

-- kill_count 
update 
	faction_stats fs2 
set
	kill_count = 
(select count(bh.id) as kill_count from battle_history bh
inner join battle_mechs bm on bm.battle_id = bh.battle_id and bm.mech_id = bh.war_machine_two_id and bm.faction_id = fs2.id 
where bh.event_type = 'killed')
where 
	fs2.id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

-- death_count
update 
	faction_stats fs2 
set
	death_count = 
(select count(bh.id) as death_count from battle_history bh
inner join battle_mechs bm on bm.battle_id = bh.battle_id and bm.mech_id = bh.war_machine_one_id and bm.faction_id = fs2.id 
where bh.event_type = 'killed')
where 
	fs2.id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

-- sups_contribute
update 
	faction_stats fs2 
set
	sups_contribute = 
(select coalesce(round(sum(amount)/1000000000000000000),0) as sups_contribute from battle_contributions bc
where bc.faction_id = fs2.id)
where 
	fs2.id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

-- mvp_player_id 
update 
	faction_stats fs2 
set
	mvp_player_id = 
(select bc.player_id from battle_contributions bc 
where bc.faction_id = fs2.id 
group by player_id
order by sum(amount) desc 
limit 1)
where 
	fs2.id = '98bf7bb3-1a7c-4f21-8843-458d62884060';

-----------------
-- Boston stat --
-----------------

-- win_count 
update
	faction_stats fs2
set
win_count =
(select count(distinct (bw.battle_id)) as win_count  from battle_wins bw
where bw.faction_id = fs2.id)
where fs2.id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

-- loss_count
update 
	faction_stats fs2 
set
	loss_count = 
(select	(select count (b.id) as actaul_battle_count from battles b where exists (select 1 from battle_wins bw where bw.battle_id = b.id)) -  count(distinct (bw.battle_id)) as loss_count  
from battle_wins bw	where bw.faction_id = fs2.id)
where 
	fs2.id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

-- kill_count 
update 
	faction_stats fs2 
set
	kill_count = 
(select count(bh.id) as kill_count from battle_history bh
inner join battle_mechs bm on bm.battle_id = bh.battle_id and bm.mech_id = bh.war_machine_two_id and bm.faction_id = fs2.id 
where bh.event_type = 'killed')
where 
	fs2.id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

-- death_count
update 
	faction_stats fs2 
set
	death_count = 
(select count(bh.id) as death_count from battle_history bh
inner join battle_mechs bm on bm.battle_id = bh.battle_id and bm.mech_id = bh.war_machine_one_id and bm.faction_id = fs2.id 
where bh.event_type = 'killed')
where 
	fs2.id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

-- sups_contribute
update 
	faction_stats fs2 
set
	sups_contribute = 
(select coalesce(round(sum(amount)/1000000000000000000),0) as sups_contribute from battle_contributions bc
where bc.faction_id = fs2.id)
where 
	fs2.id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

-- mvp_player_id 
update 
	faction_stats fs2 
set
	mvp_player_id = 
(select bc.player_id from battle_contributions bc 
where bc.faction_id = fs2.id 
group by player_id
order by sum(amount) desc 
limit 1)
where 
	fs2.id = '7c6dde21-b067-46cf-9e56-155c88a520e2';

-------------------
-- Zaibatsu stat --
-------------------

-- win_count 
update
	faction_stats fs2
set
win_count =
(select count(distinct (bw.battle_id)) as win_count  from battle_wins bw
where bw.faction_id = fs2.id)
where fs2.id = '880db344-e405-428d-84e5-6ebebab1fe6d';

-- loss_count
update 
	faction_stats fs2 
set
	loss_count = 
(select	(select count (b.id) as actaul_battle_count from battles b where exists (select 1 from battle_wins bw where bw.battle_id = b.id)) -  count(distinct (bw.battle_id)) as loss_count  
from battle_wins bw	where bw.faction_id = fs2.id)
where 
	fs2.id = '880db344-e405-428d-84e5-6ebebab1fe6d';

-- kill_count 
update 
	faction_stats fs2 
set
	kill_count = 
(select count(bh.id) as kill_count from battle_history bh
inner join battle_mechs bm on bm.battle_id = bh.battle_id and bm.mech_id = bh.war_machine_two_id and bm.faction_id = fs2.id 
where bh.event_type = 'killed')
where 
	fs2.id = '880db344-e405-428d-84e5-6ebebab1fe6d';

-- death_count
update 
	faction_stats fs2 
set
	death_count = 
(select count(bh.id) as death_count from battle_history bh
inner join battle_mechs bm on bm.battle_id = bh.battle_id and bm.mech_id = bh.war_machine_one_id and bm.faction_id = fs2.id 
where bh.event_type = 'killed')
where 
	fs2.id = '880db344-e405-428d-84e5-6ebebab1fe6d';

-- sups_contribute
update 
	faction_stats fs2 
set
	sups_contribute = 
(select coalesce(round(sum(amount)/1000000000000000000),0) as sups_contribute from battle_contributions bc
where bc.faction_id = fs2.id)
where 
	fs2.id = '880db344-e405-428d-84e5-6ebebab1fe6d';

-- mvp_player_id 
update 
	faction_stats fs2 
set
	mvp_player_id = 
(select bc.player_id from battle_contributions bc 
where bc.faction_id = fs2.id 
group by player_id
order by sum(amount) desc 
limit 1)
where 
	fs2.id = '880db344-e405-428d-84e5-6ebebab1fe6d';